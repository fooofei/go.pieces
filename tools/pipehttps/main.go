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
	"strconv"
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
// http://127.0.0.1:18100 https://example.com:1984
// https://127.0.0.1:18101 https://example.com:1984

// WithDumpReq will dump request as http format
func WithDumpReq(w io.Writer) func(*http.Request) {
	return func(req *http.Request) {
		// dump must before .Do()
		content, err := httputil.DumpRequest(req, true)
		if err != nil {
			_, _ = fmt.Fprintf(w, "error: <%T>%v\n", err, err)
			return
		}
		_, _ = fmt.Fprintf(w, "%s\n", content)
	}
}

// WithDumpResp will dump response as http format
func WithDumpResp(w io.Writer) func(*http.Response) {
	return func(resp *http.Response) {
		content, err := httputil.DumpResponse(resp, true)
		if err != nil {
			_, _ = fmt.Fprintf(w, "error: <%T>%v\n", err, err)
			return
		}
		_, _ = fmt.Fprintf(w, "%s\n", content)
	}
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
	Mapper map[string]servicePair
}

func pipeResponse(resp *http.Response, w http.ResponseWriter) {
	for k := range resp.Header {
		w.Header().Set(k, resp.Header.Get(k))
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// ServeHTTP 将会把请求 pipe 出去
func (h *customizedHandler) ServeHTTP(w http.ResponseWriter, fromReq *http.Request) {
	ctx, cancel := context.WithTimeout(h.gtx, time.Minute)
	defer cancel()
	count := atomic.AddInt64(&h.Count, 1)
	b := bytes.NewBufferString("")
	defer func() {
		fmt.Print(b.String())
	}()
	_, _ = fmt.Fprintf(b, "---req %v---------------------------------------------------------\n", count)
	toReq := fromReq.Clone(ctx)

	toService, ok := h.Mapper[fromReq.Host]
	if !ok {
		err := fmt.Errorf("cannot found which service forward to")
		_, _ = fmt.Fprintf(b, "error: <%T>%v\n", err, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	toReq.URL.Scheme = toService.To.Scheme
	toReq.Host = toService.To.Join()
	toReq.URL.Host = toReq.Host
	toReq.RequestURI = "" // must clear

	WithDumpReq(b)(toReq)
	resp, err := h.clt.Do(toReq)
	if err != nil {
		_, _ = fmt.Fprintf(b, "error: <%T>%v\n", err, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WithDumpResp(b)(resp)
	// 这个方法不是 pipe 作用，不符合预期
	// _ = resp.Write(w)
	pipeResponse(resp, w)
	_ = resp.Body.Close()
}

func parseURL(path string) (*service, error) {
	s := &service{}
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed url.Parse '%v' %w", path, err)
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", u.Host)
	if err != nil {
		return nil, err
	}
	s.Scheme = u.Scheme
	s.Host = tcpAddr.IP.String()
	s.Port = strconv.Itoa(tcpAddr.Port)
	return s, nil
}

// parseMapper 解析配置文件
func parseMapper(filePath string) (map[string]servicePair, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	m := make(map[string]servicePair)
	s := bufio.NewScanner(bytes.NewReader(content))
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

		m[s1.Join()] = servicePair{
			From: s1,
			To:   s2,
		}
	}
	return m, nil
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
	m := http.NewServeMux()
	mapper, err := parseMapper(mapperFilePath)
	if err != nil {
		panic(err)
	}
	ch := &customizedHandler{
		logger: logger.WithName("customizedHandler"),
		gtx:    ctx,
		clt: &http.Client{
			Transport:     tr,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		},
		Mapper: mapper,
	}
	m.Handle("/", ch)

	for _, v := range ch.Mapper {
		value := net.JoinHostPort("", v.From.Port)
		logger.Info("Serve HTTP", "addr", value, "from", v.From.URL(), "to", v.To.URL())
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
