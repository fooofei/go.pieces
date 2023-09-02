package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// 如果一个微服务请求 HTTPS 接口 发生了错误，但是不能抓包定位
// 可以让服务去请求这个文件的 HTTP 接口，然后我们帮他请求 HTTPS

// 对比下这个库 https://github.com/projectdiscovery/proxify
// 代码质量不好，对于协程的等待，关闭，不优雅

// mapper file format:
// http://127.0.0.1:18100 https://example.com:1984+

func listenChainList(ctx context.Context, logger *slog.Logger, chains []Chain) {
	var err error
	var requestSequence int64
	var wg = sync.WaitGroup{}

	for _, chain := range chains {
		l := logger.With("from", chain.From.URL(), "to", chain.To.URL())
		wg.Add(1)
		go func(c Chain) {
			defer wg.Done()
			l.Info("create forward pair")
			if err = listenChain(ctx, c, &requestSequence, createUpstreamTransport(c.To.KeyLogWriter)); err != nil {
				l.Error("exit listen for error occur in listenChain", "error", err)
			}
		}(chain)
	}
	wg.Wait()
}

func listenChain(ctx context.Context, chain Chain, requestSequence *int64, upstreamTransport *http.Transport) error {
	var serv = &http.Server{Addr: chain.From.ListenAddr()}
	// create a listen
	var lc = net.ListenConfig{}
	var ln, err = lc.Listen(ctx, "tcp", chain.From.ListenAddr())
	if err != nil {
		return err
	}
	// set handler
	var upstreamClient = &http.Client{
		Transport:     upstreamTransport,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	serv.Handler = getServerHandleFunc(ctx, upstreamClient, requestSequence, chain)

	serv.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
		KeyLogWriter:       chain.From.KeyLogWriter,
	}

	var closeCh = make(chan error, 1)

	go func() {
		var errServe error
		switch chain.From.Scheme {
		case HttpsScheme:
			// https 要本地预置证书文件
			if errServe = testFileExists(chain.CertFilePath); errServe != nil {
				break
			}
			if errServe = testFileExists(chain.KeyFilePath); errServe != nil {
				break
			}
			errServe = serv.ServeTLS(ln, chain.CertFilePath, chain.KeyFilePath)
		case HttpScheme:
			errServe = serv.Serve(ln)
		default:
			errServe = fmt.Errorf("not support scheme %s", chain.From.Scheme)
		}
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

func createUpstreamTransport(tlsKeyLogWriter io.Writer) *http.Transport {
	var trf = http.DefaultTransport.(*http.Transport)
	var tr = trf.Clone()
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}
	tr.Proxy = nil
	tr.TLSClientConfig.InsecureSkipVerify = true // 我们的目的是定位问题 忽略证书校验 我们访问其他地址不校验
	tr.TLSClientConfig.KeyLogWriter = tlsKeyLogWriter
	return tr
}

func main() {
	var mapperFilePath string
	var certsDir string
	flag.StringVar(&mapperFilePath, "mapper", "mapper.txt", "The host:port mapper file path")
	flag.StringVar(&certsDir, "certs", "./certs", "The cert.pem and key.pem file dir")
	flag.Parse()
	var logWriter = os.Stderr
	var logger = slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{})).With("pid", os.Getpid())
	var ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var chainList, err = parseChainList(mapperFilePath)
	if err != nil {
		panic(err)
	}
	if len(chainList) == 0 {
		fmt.Printf("not found any valid chain\n")
		return
	}

	var ts = strconv.FormatInt(time.Now().UnixMilli(), 10)
	var keyLogFilePath = fmt.Sprintf("%s/%s_%s_tls_", getCurDir(), ts, string(randString(16)))
	var clientKeyLogCh = make(chan []byte)
	var upstreamKeyLogCh = make(chan []byte)

	chainList = setChainListKeyLogWriter(chainList, createKeyLogWriter(ctx, clientKeyLogCh), createKeyLogWriter(ctx, upstreamKeyLogCh))
	chainList = setChainListHttpErrLog(chainList, logWriter)
	chainList = setChainListTLSKeyCert(chainList, filepath.Join(certsDir, "key.pem"), filepath.Join(certsDir, "cert.pem"))
	var wg = sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		writeKeyLog(ctx, keyLogFilePath+"client.txt", clientKeyLogCh)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		writeKeyLog(ctx, keyLogFilePath+"upstream.txt", upstreamKeyLogCh)
	}()

	listenChainList(ctx, logger, chainList)
	wg.Wait()
}
