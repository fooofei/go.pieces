package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
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
		// body 可能为 gzip 压缩，参数传递 false， 不输出
		content, err := httputil.DumpResponse(resp, false)
		if err != nil {
			fmt.Fprintf(w, "error: <%T>%v\n", err, err)
			return
		}
		fmt.Fprintf(w, "%s", content)
		// 保留原始body，对拷贝份进行 gzip 解压读取
		var b, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(b))
		var bodyReader io.Reader = bytes.NewReader(b)
		if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
			if r, err := gzip.NewReader(bytes.NewReader(b)); err == nil {
				bodyReader = r
				defer func() {
					r.Close()
				}()
			}
		}
		io.Copy(w, bodyReader)
		fmt.Fprintf(w, "\n")
	}
}

// cloneReqWithNewHost will clone http request from req, with updated host
// host format is http://1.1.1.1:9090  not tail with '/'
// 这个很重要 找了很久如何只更新连接地址，这个比较理想
func cloneReqWithNewHost(ctx context.Context, req *http.Request, host string) (*http.Request, error) {
	r := req.Clone(ctx)
	r.URL.Scheme = ""
	r.URL.Host = ""
	reqUrl, err := http.NewRequest(req.Method, fmt.Sprintf("%s%s", host, r.URL.String()), nil)
	if err != nil {
		return nil, err
	}
	r.URL = reqUrl.URL
	r.Host = reqUrl.Host
	r.RequestURI = reqUrl.RequestURI
	return r, nil
}
