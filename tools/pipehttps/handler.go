package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

func pipeResponse(resp *http.Response, w http.ResponseWriter) error {
	for k := range resp.Header {
		w.Header().Set(k, resp.Header.Get(k))
	}
	w.WriteHeader(resp.StatusCode)
	var _, err = io.Copy(w, resp.Body)
	return err
}

func getServerHandleFunc(ctxg context.Context, upstreamClient *http.Client, seq *int64, chain Chain) http.HandlerFunc {
	// ServeHTTP 将会把请求 pipe 出去
	return func(w http.ResponseWriter, request *http.Request) {
		var (
			ctx, cancel  = context.WithTimeout(ctxg, time.Minute)
			count        = atomic.AddInt64(seq, 1)
			dumpBuffer   = bytes.NewBufferString("")
			upstreamReq  *http.Request
			err          error
			upstreamResp *http.Response
		)
		defer cancel()
		defer func() {
			fmt.Print(dumpBuffer.String())
		}()

		if upstreamReq, err = cloneReqWithNewHost(ctx, request, chain.To.URL()); err != nil {
			fmt.Fprintf(dumpBuffer, "failed create request with error: <%T>%v\n", err, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(dumpBuffer, "---req %v----------from %v to %v----------------------------------------\n",
			count, chain.From.URL(), chain.To.URL())
		WithDumpReq(dumpBuffer)(upstreamReq)
		if upstreamResp, err = upstreamClient.Do(upstreamReq); err != nil {
			fmt.Fprintf(dumpBuffer, "failed request to upstream error: <%T>%v\n", err, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(dumpBuffer, "---resp %v----------from %v to %v----------------------------------------\n",
			count, chain.From.URL(), chain.To.URL())
		WithDumpResp(dumpBuffer)(upstreamResp)

		// 这个方法不是 pipe 作用，不符合预期
		// _ = resp.Write(w)
		if err = pipeResponse(upstreamResp, w); err != nil {
			fmt.Fprintf(dumpBuffer, "failed write client response from upstream response")
		}
		upstreamResp.Body.Close()
	}
}
