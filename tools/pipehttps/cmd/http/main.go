package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	. "github.com/fooofei/go_pieces/tools/pipehttps"
)

func main() {
	var ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	setupServer(ctx)
}

var gclt = &http.Client{}

func getClient() *http.Client {
	return gclt
}

func handlerWithCtx(ctx context.Context) http.HandlerFunc {
	var count = &atomic.Int64{}

	return func(w http.ResponseWriter, r *http.Request) {
		var upstream = ""
		slog.Info("get new request", "count", count.Add(1))
		var clt = getClient()

		var buf = bytes.NewBufferString("")
		var req, err = CloneReqWithNewHost(ctx, r, upstream)
		if err != nil {
			slog.Error("failed clone request", "error", err)
			return
		}

		WithDumpReq(buf)(req)
		var resp *http.Response
		if resp, err = clt.Do(req); err != nil {
			slog.Error("failed forward request to upstream", "error", err)
			fmt.Fprintf(os.Stdout, buf.String())
			return
		}
		WithDumpResp(buf)(resp)
		PipeResponse(resp, w)
		fmt.Fprintf(os.Stdout, buf.String())
	}
}

func setupServer(ctx context.Context) error {
	addr := ":9002"
	// 构建一个局部 http server，不使用 http 包的默认 server
	// 使用全局的会跟其他包注册到同一个路由上，会互相妨碍
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerWithCtx(ctx))
	lc := &net.ListenConfig{}
	// this context will close listener
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:    addr,
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
		slog.Info("shutdown server")
		_ = server.Shutdown(ctx)
	}
	return err
}
