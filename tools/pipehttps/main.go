package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// 如果一个微服务请求 HTTPS 接口 发生了错误，但是不能抓包定位
// 可以让服务去请求这个文件的 HTTP 接口，然后我们帮他请求 HTTPS

// 对比下这个库 https://github.com/projectdiscovery/proxify
// 代码质量不好，对于协程的等待，关闭，不优雅

func doListenList(ctx context.Context, logger *slog.Logger, urlList []Url,
	fnConfigServer func(serv *http.Server),
	fnConfigUpstreamClient func(trans *http.Transport),
	fnGetListener func(u Url, tcpListener net.Listener) (net.Listener, error),
	proxyServerErrLog *log.Logger) {
	var err error
	var requestSequence atomic.Int64
	var wg = sync.WaitGroup{}

	for _, u := range urlList {
		var l = logger.With("listen", u.URL())
		wg.Add(1)
		go func(u2 Url) {
			defer wg.Done()
			l.Info("enter listen")
			if err = doListen(ctx, u2, &requestSequence,
				fnConfigServer, fnConfigUpstreamClient, fnGetListener,
				proxyServerErrLog); err != nil {
				l.Error("failed do listen", "error", err)
			}
		}(u)
	}
	wg.Wait()
}

func createHttpTransport() *http.Transport {
	var trf = http.DefaultTransport.(*http.Transport)
	var tr = trf.Clone()
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}
	tr.Proxy = nil
	tr.TLSClientConfig.InsecureSkipVerify = true // 我们的目的是定位问题 忽略证书校验 我们访问其他地址不校验
	return tr
}

func doListen(ctx context.Context, u Url, requestSequence *atomic.Int64,
	fnConfigServer func(serv *http.Server),
	fnConfigUpstreamClient func(trans *http.Transport),
	fnGetListener func(u Url, tcpListener net.Listener) (net.Listener, error),
	proxyServerErrLog *log.Logger) error {
	// server and client
	var serv = &http.Server{Addr: u.ListenAddr()}
	var trans = createHttpTransport()
	fnConfigUpstreamClient(trans)

	// create a listen
	var lc = net.ListenConfig{}
	var ln, err = lc.Listen(ctx, "tcp", u.ListenAddr())
	if err != nil {
		return err
	}

	serv.Handler = GetServerHandleFuncV2(ctx, trans, proxyServerErrLog, requestSequence, u)
	serv.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	fnConfigServer(serv)
	if ln, err = fnGetListener(u, ln); err != nil {
		return err
	}

	var closeCh = make(chan error, 1)

	go func() {
		var errServe = serv.Serve(ln) // 可能是 http 可能是 https
		if errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
			closeCh <- errServe
		}
		close(closeCh)
	}()

	select {
	case err = <-closeCh:
		serv.Shutdown(ctx)
		return err
	case <-ctx.Done():
		var shutdownCtx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return serv.Shutdown(shutdownCtx)
	}
}

// 给 net/http 框架使用的 log 打印对象
func createHttpErrLog(w io.Writer) *log.Logger {
	var l = log.New(w, "net/http/server ", log.Lshortfile|log.LstdFlags)
	return l
}

func configServer(httpErrLog *log.Logger) func(serv *http.Server) {
	return func(serv *http.Server) {
		serv.ErrorLog = httpErrLog
	}
}

func getServerListener(tlsCfg *tls.Config) func(u Url, tcpListener net.Listener) (net.Listener, error) {
	return func(u Url, tcpListener net.Listener) (net.Listener, error) {
		switch u.Scheme {
		case HttpsScheme:
			return tls.NewListener(tcpListener, tlsCfg), nil
		case HttpScheme:
			return tcpListener, nil
		default:
			return nil, fmt.Errorf("not support scheme %s", u.Scheme)
		}
	}
}

func configUpstreamClient(upstreamKeyLog io.Writer, dnsMap map[string]string) func(trans *http.Transport) {
	return func(trans *http.Transport) {
		trans.TLSClientConfig.KeyLogWriter = upstreamKeyLog
		WithCustomDomain(dnsMap)(trans)
	}
}

// 不要使用 http.Server.TlsConfig 了，这个在 ServeTLS 会进行复制，这个配置也不优雅
func createServerTlsConfig(keyLogWriter io.Writer, tlsKeyFile string, tlsCertFile string) (*tls.Config, error) {
	// https 要本地预置证书文件
	var err error
	if err = TestFileExists(tlsCertFile); err != nil {
		return nil, err
	}
	if err = TestFileExists(tlsKeyFile); err != nil {
		return nil, err
	}
	var cfg = &tls.Config{}
	cfg.InsecureSkipVerify = true
	cfg.KeyLogWriter = keyLogWriter
	cfg.Certificates = make([]tls.Certificate, 1)
	if cfg.Certificates[0], err = tls.LoadX509KeyPair(tlsCertFile, tlsKeyFile); err != nil {
		return nil, err
	}
	return cfg, nil
}

func main() {
	var dnsFilePath string
	var certsDir string
	var listenList StringSlice
	var version bool
	var versionString = "1.0.20231205"
	flag.StringVar(&dnsFilePath, "dns", "dns.txt", "The dns mapper file path")
	flag.StringVar(&certsDir, "certs", "./certs", "The cert.pem and key.pem file dir")
	flag.Var(&listenList, "listen", "the listeners, format of http://0.0.0.0:443")
	flag.BoolVar(&version, "version", false, fmt.Sprintf("version is %v", versionString))
	flag.Parse()
	if version {
		fmt.Println(versionString)
		return
	}
	var logWriter = os.Stderr
	var logger = slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{})).With("pid", os.Getpid())
	var ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var urlList, err = ParserUrlList(listenList)
	if err != nil {
		panic(err)
	}
	if len(listenList) < 1 {
		fmt.Println("not given of listen")
		return
	}
	var dnsMap map[string]string
	if err = TestFileExists(dnsFilePath); err == nil {
		if dnsMap, err = ParseDnsMapper(dnsFilePath); err != nil {
			panic(err)
		}
	}

	var ts = strconv.FormatInt(time.Now().UnixMilli(), 10)
	var keyLogFilePath = fmt.Sprintf("%s/%s_%s_tls_", GetCurDir(), ts, string(RandString(16)))
	// 要写入文件的内容先写入这两个 ch
	var clientKeyLogCh = make(chan []byte)
	var upstreamKeyLogCh = make(chan []byte)

	var tlsCfg *tls.Config
	if hasHttpsUrl(urlList) {
		var keyPath string
		var certPath string
		if keyPath, err = filepath.Abs(filepath.Join(certsDir, "key.pem")); err != nil {
			panic(err)
		}
		if certPath, err = filepath.Abs(filepath.Join(certsDir, "cert.pem")); err != nil {
			panic(err)
		}
		logger.Info("load certs", "key", keyPath, "cert", certPath)
		if tlsCfg, err = createServerTlsConfig(CreateKeyLogWriter(ctx, clientKeyLogCh), keyPath, certPath); err != nil {
			panic(err)
		}
	}

	var wg = sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		WriteKeyLog(ctx, keyLogFilePath+"client.txt", clientKeyLogCh)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		WriteKeyLog(ctx, keyLogFilePath+"upstream.txt", upstreamKeyLogCh)
	}()

	doListenList(ctx, logger, urlList,
		configServer(createHttpErrLog(logWriter)),
		configUpstreamClient(CreateKeyLogWriter(ctx, upstreamKeyLogCh), dnsMap),
		getServerListener(tlsCfg),
		createHttpErrLog(logWriter))
	wg.Wait()
}
