package test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/exp/slog"
)

func sampleHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "this is a test")
}

//
// 这篇文章介绍如何正确 shutdown，就像下面的代码这样 https://dev.to/mokiat/proper-http-shutdown-in-go-3fji
//

// serveHttp 是正确创建一个 http server
//
// h 可以是提前创建的
// 可以学习以下，构建一个局部 http server，不使用 http 包的默认 server
// 使用全局的会跟其他包注册到同一个路由上，会互相妨碍
// mux := http.NewServeMux()
// mux.HandleFunc("/metrics", sampleHandler)
//
// 与语言内建的 http server context 会有这两个 key
// var (
// 	// ServerContextKey is a context key. It can be used in HTTP
// 	// handlers with Context.Value to access the server that
// 	// started the handler. The associated value will be of
// 	// type *http.Server.
// 	ServerContextKey = &contextKey{"http-server"}

//	// LocalAddrContextKey is a context key. It can be used in
//	// HTTP handlers with Context.Value to access the local
//	// address the connection arrived on.
//	// The associated value will be of type net.Addr.
//	LocalAddrContextKey = &contextKey{"local-addr"}
//
// )
// 在 handler 处理中可以从 ctx 取用
func serveHttp(ctx context.Context, logger *slog.Logger, addr string, h http.Handler) error {
	var lc = &net.ListenConfig{}
	// this context will close listener
	var ln, err = lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	var srv = &http.Server{
		Addr:    addr,
		Handler: h,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		// BaseContext 的作用不是关闭 accept，不是关闭正在建立的连接
		// 要 关闭连接还是要用 shutdown
		// 该 context 会传递到 handler 入参 request.Context()
	}
	var serverClosedCh = make(chan error, 1) // 使用有缓存的队列，防止协程入队时阻塞
	go func() {
		err = srv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverClosedCh <- err
		}
		close(serverClosedCh)
	}()

	select {
	case err = <-serverClosedCh:
		return err
	case <-ctx.Done():
		logger.Info("shuting down server")
		// 使用自定义超时的 context，不使用当前的 ctx，使用当前的 ctx 会立刻退出 shutdown
		var exitCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(exitCtx)
		logger.Info("server exit")
		return nil
	}
}

func ExampleHTTPServer() {
	ctx, cancel := context.WithCancel(context.Background())
	logger := slog.New(slog.NewJSONHandler(os.Stdout))
	logger = logger.With("pid", os.Getpid())
	logger.Info("enter main")
	ctx, _ = context.WithTimeout(ctx, 6*time.Second)
	err := setupServer(ctx, logger)
	logger.Error("err is", "error", err)
	logger.Info("main routine exit")
	time.Sleep(time.Minute)
	_ = cancel
	_ = ctx
}

func ExampleHTTPWithProxy() {
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse("http://127.0.0.1:3128")
	}
	transport := &http.Transport{Proxy: proxy}
	_ = &http.Client{Transport: transport}
}

func ExampleHTTPDisableProxy() {
	tr := &http.Transport{
		Proxy: nil,
	}
	_ = &http.Client{Transport: tr}
}

// NewCertPool read ca.cert files to make CertPool.
// ca 证书
func NewCertPool(CAFiles []string) (*x509.CertPool, error) {
	cp := x509.NewCertPool()
	for _, CAFile := range CAFiles {
		pemByte, err := ioutil.ReadFile(CAFile)
		if err != nil {
			return nil, err
		}
		ok := cp.AppendCertsFromPEM(pemByte)
		if !ok {
			return nil, fmt.Errorf("failed AppendCertsFromPEM() for %v", CAFile)
		}
	}
	return cp, nil
}

// 加载 certFile keyFile 示例
//   config.Certificates = make([]tls.Certificate, 1)
//   config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)

// http client tls config
func MakeTlsConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
		MinVersion: tls.VersionTLS12,
	}
}

// 把 context 的时间限制转化为 duration
// timeout 为 0 表示不限制
func contextToTimeout(waitCtx context.Context) time.Duration {
	if deadline, ok := waitCtx.Deadline(); ok {
		// now - deadline < 0
		timeout := -time.Since(deadline)
		if timeout > 0 {
			return timeout
		}
	}
	return 0
}

// CipherSuitesFromName 把 string 类型的转换为 cipher类型数组
func CipherSuitesFromName(names []string) []*tls.CipherSuite {
	m := make(map[string]*tls.CipherSuite, len(tls.CipherSuites()))
	for _, cipher := range tls.CipherSuites() {
		m[cipher.Name] = cipher
	}

	r := make([]*tls.CipherSuite, 0)
	for _, n := range names {
		if _, ok := m[n]; ok {
			r = append(r, m[n])
		}
	}
	return r
}

// WithContext 把只有 doDeadline 的 API 封装为可以使用 context
func WithContext(doDeadline func(t time.Time) error, or func() error) func(context.Context) error {
	return func(ctx context.Context) error {
		if deadline, ok := ctx.Deadline(); ok {
			err := doDeadline(deadline)
			// 因为 deadline 超时后 ctx.Done 不一定返回，
			// 因此在超时时间到达之后，显式等待 ctx.Done 一定返回
			if time.Until(deadline) <= 0 {
				<-ctx.Done()
			}
			return err
		} else {
			return or()
		}
	}
}
