package main

import (
	"context"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/fooofei/stdr"
	"github.com/go-logr/logr"
	_ "github.com/julienschmidt/httprouter"
)

var logger logr.Logger

func sampleHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "this is a test")
}

func setupServer(ctx context.Context) error {
	metricAddr := ":8888"
	// 构建一个局部 http server，不使用 http 包的默认 server
	// 使用全局的会跟其他包注册到同一个路由上，会互相妨碍
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", sampleHandler)
	lc := &net.ListenConfig{}
	// this context will close listener
	ln, err := lc.Listen(ctx, "tcp", metricAddr)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:    metricAddr,
		Handler: mux,
    // baseCtx 的作用不是关闭 accept，不是关闭正在建立的连接
    // 要 关闭连接还是要用 shutdown
	}
	serverClosedCh := make(chan error, 1)
	go func() {
		err = server.Serve(ln)
		serverClosedCh <- err
		close(serverClosedCh)
	}()

	select {
	case err = <-serverClosedCh:
		return err
	case <-ctx.Done():
		logger.Info("shutdown server")
		_ = server.Shutdown(ctx)
	}
	return err
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	logger = stdr.New(stdlog.New(os.Stdout, "", stdlog.Lshortfile|stdlog.LstdFlags))
	logger = logger.WithValues("pid", os.Getpid())
	logger.Info("enter main")
	ctx, _ = context.WithTimeout(ctx, 6*time.Second)
	err := setupServer(ctx)
	logger.Error(err, "err is")
	logger.Info("main routine exit")
	time.Sleep(time.Minute)
	_ = cancel
	_ = ctx
}
