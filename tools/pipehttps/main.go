package main

import (
	"context"
	"crypto/tls"
	"flag"
	"github.com/fooofei/go_pieces/tools/pipehttps/url"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// 如果一个微服务请求 HTTPS 接口 发生了错误，但是不能抓包定位
// 可以让服务去请求这个文件的 HTTP 接口，然后我们帮他请求 HTTPS

// 对比下这个库 https://github.com/projectdiscovery/proxify
// 代码质量不好，对于协程的等待，关闭，不优雅

// mapper file format:
// http://127.0.0.1:18100 https://example.com:1984+

type serveFunc func(server *http.Server, listener net.Listener) error

func serveHTTP(logWriter io.Writer) serveFunc {
	return func(server *http.Server, listener net.Listener) error {
		// add a tail blank for prefix
		server.ErrorLog = log.New(logWriter, "net/http/server ", log.Lshortfile|log.LstdFlags)
		return server.Serve(listener)
	}
}

func serveHTTPS(certFile, keyFile string, tlsConfig *tls.Config, logWriter io.Writer) serveFunc {
	return func(server *http.Server, listener net.Listener) error {
		server.TLSConfig = tlsConfig
		// add a tail blank for prefix
		server.ErrorLog = log.New(logWriter, "net/http/server ", log.Lshortfile|log.LstdFlags)
		return server.ServeTLS(listener, certFile, keyFile)
	}
}

func listenAndServe(ctx context.Context, addr string, handler http.Handler, fn serveFunc) error {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	serv := &http.Server{Addr: addr, Handler: handler}
	return fn(serv, ln)
}

func partialListenFunc(ctx context.Context, addr string, handler http.Handler, fn serveFunc) func() error {
	return func() error {
		return listenAndServe(ctx, addr, handler, fn)
	}
}

type globalContext struct {
	RequestSequence int64
	ServerCertFile  string
	ServerKeyFile   string
	Transport       *http.Transport // used for http client
	SvrTlsCfg       *tls.Config     // used for http server
}

func listenChainList(ctx context.Context, logger *slog.Logger, logWriter io.Writer, chains []url.Chain, gc *globalContext) {
	var err error
	var errCh = make(chan error, 100)

	for _, v := range chains {
		l := logger.With("from", v.From.URL(), "to", v.To.URL())

		ch := &ChainHandler{
			gtx: ctx,
			clt: &http.Client{
				Transport:     gc.Transport,
				CheckRedirect: nil,
				Jar:           nil,
				Timeout:       0,
			},
			seq:   &gc.RequestSequence,
			chain: v,
		}
		m := http.NewServeMux()
		m.Handle("/", ch)

		l.Info("Serve Handler", "addr", v.From.Join())
		var svFunc serveFunc
		logger.Handler()
		switch v.From.Scheme {
		case "https":
			// https 要本地预置证书文件
			svFunc = serveHTTPS(gc.ServerCertFile, gc.ServerKeyFile, gc.SvrTlsCfg, logWriter)
		case "http":
			svFunc = serveHTTP(logWriter)
		default:
			panic("not support scheme " + v.From.Scheme)
		}

		go func(pfn func() error) {
			if err = pfn(); err != nil {
				errCh <- err
			}
		}(partialListenFunc(ctx, v.From.Join(), m, svFunc))
	}

	select {
	case <-ctx.Done():
	case err = <-errCh:
		panic(err)
	}
}

func createGlobalContext(certsDir string, cltTlsKeyLogWriter io.Writer, svrTlsKeyLogWriter io.Writer) *globalContext {
	var gc = &globalContext{
		ServerCertFile:  filepath.Join(certsDir, "server.cert.pem"),
		ServerKeyFile:   filepath.Join(certsDir, "server.key.pem"),
		RequestSequence: 0,
	}
	// setup client's config
	var trf = http.DefaultTransport.(*http.Transport)
	var tr = trf.Clone()
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}
	tr.Proxy = nil
	tr.TLSClientConfig.InsecureSkipVerify = true // 我们的目的是定位问题 忽略证书校验 我们访问其他地址不校验
	tr.TLSClientConfig.KeyLogWriter = cltTlsKeyLogWriter
	gc.Transport = tr

	// setup server's config
	gc.SvrTlsCfg = &tls.Config{
		InsecureSkipVerify: true,
		KeyLogWriter:       svrTlsKeyLogWriter,
	}
	return gc
}

func main() {
	var mapperFilePath string
	var certsDir string
	flag.StringVar(&mapperFilePath, "mapper", "mapper.txt", "The host:port mapper file path")
	flag.StringVar(&certsDir, "certs", "./certs", "The server.cert.pem and server.key.pem file dir")
	flag.Parse()
	var logWriter = os.Stderr
	var logger = slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{})).With("pid", os.Getpid())
	var ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var chains, err = url.ParseChain(mapperFilePath)
	if err != nil {
		panic(err)
	}

	var keyLogFilePath = generateTmpFile()
	var clt *os.File
	var svr *os.File
	if clt, err = os.Create(keyLogFilePath + "-client.txt"); err != nil {
		panic(err)
	}
	defer clt.Close()
	if svr, err = os.Create(keyLogFilePath + "-server.txt"); err != nil {
		panic(err)
	}
	defer svr.Close()
	var gc = createGlobalContext(certsDir, clt, svr)

	logger.Info("write TLS master secrets", "client", clt.Name(), "server", svr.Name())
	listenChainList(ctx, logger, logWriter, chains, gc)
}
