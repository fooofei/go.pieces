package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

// 如果一个微服务请求 HTTPS 接口 发生了错误，但是不能抓包定位
// 可以让服务去请求这个文件的 HTTP 接口，然后我们帮他请求 HTTPS

// mapper file format:
// http://127.0.0.1:18100 https://example.com:1984+
// WithDumpReq will dump request as http format
func WithDumpReq(w io.Writer) func(*http.Request) {
	return func(req *http.Request) {
		// dump must before .Do()
		content, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Fprintf(w, "error: <%T>%v\n", err, err)
			return
		}
		fmt.Fprintf(w, "%s\n", content)
	}
}

// WithDumpResp will dump response as http format
func WithDumpResp(w io.Writer) func(*http.Response) {
	return func(resp *http.Response) {
		content, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Fprintf(w, "error: <%T>%v\n", err, err)
			return
		}
		fmt.Fprintf(w, "%s\n", content)
	}
}

// cloneReqWithNewHost will clone http request from req, with updated host
// host format is http://1.1.1.1:9090  not tail with '/'
// 这个很重要 找了很久如何只更新连接地址，这个比较理想
func cloneReqWithNewHost(ctx context.Context, req *http.Request, host string) (*http.Request, error) {
	r1 := req.Clone(ctx)
	r2, err := http.NewRequest(req.Method, fmt.Sprintf("%s%s", host, req.URL.Path), nil)
	if err != nil {
		return nil, err
	}
	r1.URL = r2.URL
	r1.Host = r2.Host
	r1.RequestURI = r2.RequestURI
	return r1, nil
}

type serveFunc func(server *http.Server, listener net.Listener) error

func serveHTTP() serveFunc {
	return func(server *http.Server, listener net.Listener) error {
		return server.Serve(listener)
	}
}

func serveHTTPS(certFile, keyFile string) serveFunc {
	return func(server *http.Server, listener net.Listener) error {
		return server.ServeTLS(listener, certFile, keyFile)
	}
}

func listenAndServe(ctx context.Context, addr string, handler http.Handler, opt serveFunc) error {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	serv := &http.Server{Addr: addr, Handler: handler}
	return opt(serv, ln)
}

type service struct {
	Scheme string
	Host   string
	Port   string
}

func (s *service) URL() string {
	return fmt.Sprintf("%v://%v", s.Scheme, s.Join())
}

func (s *service) Join() string {
	if s.Port == "" {
		return s.Host
	}
	return net.JoinHostPort(s.Host, s.Port)
}

type servicePair struct {
	From *service
	To   *service
}

type customizedHandler struct {
	logger logr.Logger
	clt    *http.Client
	gtx    context.Context
	Count  int64
	pair   servicePair
}

func pipeResponse(resp *http.Response, w http.ResponseWriter) {
	for k := range resp.Header {
		w.Header().Set(k, resp.Header.Get(k))
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ServeHTTP 将会把请求 pipe 出去
func (h *customizedHandler) ServeHTTP(w http.ResponseWriter, req1 *http.Request) {
	var (
		ctx, cancel = context.WithTimeout(h.gtx, time.Minute)
		count       = atomic.AddInt64(&h.Count, 1)
		b           = bytes.NewBufferString("")
		req2        *http.Request
		err         error
		rsp         *http.Response
	)
	defer cancel()
	defer func() {
		fmt.Print(b.String())
	}()

	if req2, err = cloneReqWithNewHost(ctx, req1, h.pair.To.URL()); err != nil {
		fmt.Fprintf(b, "failed create request with error: <%T>%v\n", err, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(b, "---req %v----------from %v to %v----------------------------------------\n",
		count, h.pair.From.URL(), h.pair.To.URL())
	WithDumpReq(b)(req2)
	if rsp, err = h.clt.Do(req2); err != nil {
		fmt.Fprintf(b, "error: <%T>%v\n", err, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(b, "---resp %v----------from %v to %v----------------------------------------\n",
		count, h.pair.From.URL(), h.pair.To.URL())
	WithDumpResp(b)(rsp)
	// 这个方法不是 pipe 作用，不符合预期
	// _ = resp.Write(w)
	pipeResponse(rsp, w)
	rsp.Body.Close()
}

func parseURL(path string) (*service, error) {
	s := &service{}
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed url.Parse '%v' %w", path, err)
	}
	s.Scheme = u.Scheme
	host, port, err := net.SplitHostPort(u.Host)
	if err == nil {
		s.Host = host
		s.Port = port
	} else {
		s.Host = u.Host
		s.Port = ""
	}
	return s, nil
}

// parseMapper 解析配置文件
func parseMapper(filePath string) ([]servicePair, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(bytes.NewReader(content))
	pairs := make([]servicePair, 0)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		pair := strings.Split(line, " ")
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid format of '%s'", line)
		}
		var s1 *service
		var s2 *service
		if s1, err = parseURL(pair[0]); err != nil {
			return nil, err
		}
		if s2, err = parseURL(pair[1]); err != nil {
			return nil, err
		}
		pairs = append(pairs, servicePair{
			From: s1,
			To:   s2,
		})
	}
	return pairs, nil
}

func main() {
	var mapperFilePath string
	var certsDir string
	flag.StringVar(&mapperFilePath, "mapper", "mapper.txt", "The host:port mapper file path")
	flag.StringVar(&certsDir, "certs", "", "The server.cert.pem and server.key.pem file dir")
	flag.Parse()
	logger := stdr.New(stdlog.New(os.Stdout, "", stdlog.Lshortfile|stdlog.LstdFlags))
	logger = logger.WithValues("pid", os.Getpid())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var tr *http.Transport
	trf := http.DefaultTransport.(*http.Transport)
	tr = trf.Clone()
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}

	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse("http://127.0.0.1:3128")
	}
	_ = proxy
	// 关闭使用代理
	tr.Proxy = nil
	tr.TLSClientConfig.InsecureSkipVerify = true // 我们的目的是定位问题 忽略证书校验

	pairs, err := parseMapper(mapperFilePath)
	if err != nil {
		panic(err)
	}

	for _, v := range pairs {
		var value string
		if v.From.Port == "" {
			if v.From.Scheme == "https" {
				value = ":443"
			} else {
				value = ":80"
			}
		} else {
			value = net.JoinHostPort("", v.From.Port)
		}
		l := logger.WithValues("from", v.From.URL(), "to", v.To.URL())
		ch := &customizedHandler{
			logger: l.WithName("customizedHandler"),
			gtx:    ctx,
			clt: &http.Client{
				Transport:     tr,
				CheckRedirect: nil,
				Jar:           nil,
				Timeout:       0,
			},
			Count: 0,
			pair:  v,
		}
		m := http.NewServeMux()
		m.Handle("/", ch)
		l.Info("Serve HTTP", "addr", value)
		switch v.From.Scheme {
		case "https":
			// https 要本地预置证书文件
			a := filepath.Join(certsDir, "server.cert.pem")
			b := filepath.Join(certsDir, "server.key.pem")
			go listenAndServe(ctx, value, m, serveHTTPS(a, b))
		case "http":
			go listenAndServe(ctx, value, m, serveHTTP())
		}
	}
	<-ctx.Done()
}
