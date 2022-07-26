package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fooofei/go_pieces/tools/pipehttps/url"
	"github.com/go-logr/logr"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

type ChainHandler struct {
	logger logr.Logger
	clt    *http.Client
	gtx    context.Context
	seq    *int64
	chain  url.Chain
}

func pipeResponse(resp *http.Response, w http.ResponseWriter) {
	for k := range resp.Header {
		w.Header().Set(k, resp.Header.Get(k))
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ServeHTTP 将会把请求 pipe 出去
func (h *ChainHandler) ServeHTTP(w http.ResponseWriter, req1 *http.Request) {
	var (
		ctx, cancel = context.WithTimeout(h.gtx, time.Minute)
		count       = atomic.AddInt64(h.seq, 1)
		b           = bytes.NewBufferString("")
		req2        *http.Request
		err         error
		rsp         *http.Response
	)
	defer cancel()
	defer func() {
		fmt.Print(b.String())
	}()

	if req2, err = cloneReqWithNewHost(ctx, req1, h.chain.To.URL()); err != nil {
		fmt.Fprintf(b, "failed create request with error: <%T>%v\n", err, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(b, "---req %v----------from %v to %v----------------------------------------\n",
		count, h.chain.From.URL(), h.chain.To.URL())
	WithDumpReq(b)(req2)
	if rsp, err = h.clt.Do(req2); err != nil {
		fmt.Fprintf(b, "error: <%T>%v\n", err, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(b, "---resp %v----------from %v to %v----------------------------------------\n",
		count, h.chain.From.URL(), h.chain.To.URL())
	WithDumpResp(b)(rsp)
	// 这个方法不是 pipe 作用，不符合预期
	// _ = resp.Write(w)
	pipeResponse(rsp, w)
	rsp.Body.Close()
}
