package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

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
