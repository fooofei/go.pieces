package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
)

// WithDumpReq will dump request as http format
func WithDumpReq(w io.Writer) func(*http.Request) {
	return func(req *http.Request) {
		// dump must before .Do()
		content, err := httputil.DumpRequest(req, true)
		if err != nil && !errors.Is(err, http.ErrBodyReadAfterClose) {
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
		if err != nil && !errors.Is(err, http.ErrBodyReadAfterClose) {
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
				defer r.Close()
			}
		}
		io.Copy(w, bodyReader)
		fmt.Fprintf(w, "\n")
	}
}

// CloneReqWithNewHost will clone http request from req, with updated host
// host format is http://1.1.1.1:9090  not tail with '/'
// 这个很重要 找了很久如何只更新连接地址，这个比较理想
func CloneReqWithNewHost(ctx context.Context, req *http.Request, host string) (*http.Request, error) {
	var reqNew = req.Clone(ctx)
	if reqNew.URL.Scheme == "" {
		reqNew.URL.Scheme = "http"
	}
	reqNew.URL.Host = host
	// 不能使用 fmt.Sprintf("host:reqNew.URL.String()") 拼接结果 x.x.x.x:port/path 会导致解析报错
	// first path segment in URL cannot contain colon
	reqUrl, err := http.NewRequest(req.Method, reqNew.URL.String(), nil)
	if err != nil {
		return nil, err
	}
	reqNew.URL = reqUrl.URL
	reqNew.Host = reqUrl.Host
	reqNew.RequestURI = reqUrl.RequestURI
	return reqNew, nil
}

// CloneReqDeep 把 request 深拷贝一份
func CloneReqDeep(req *http.Request, ctx context.Context) *http.Request {
	var reqNew = req.Clone(ctx)
	if req.Body != nil {
		var body, _ = io.ReadAll(req.Body)
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(body))
		reqNew.Body = io.NopCloser(bytes.NewReader(body))
	}
	return reqNew
}

func WithContentLength(l int64) func(*http.Response) {
	return func(resp *http.Response) {
		resp.ContentLength = l
		const k = "Content-Length"
		resp.Header.Del(k)
		resp.Header.Set(k, fmt.Sprintf("%v", l))
	}
}

// WithCustomDomain 可以篡改访问目标，不修改 http 内容，修改的是 tcp 建立连接的目标
func WithCustomDomain(m map[string]string) func(transport *http.Transport) {
	var dialer = &net.Dialer{}
	return func(transport *http.Transport) {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			var addr2 = addr
			if host, port, err := net.SplitHostPort(addr); err == nil {
				if to, ok := m[host]; ok {
					addr2 = net.JoinHostPort(to, port)
				}
			}
			if to, ok := m[addr]; ok {
				addr2 = to
			}
			var conn, err = dialer.DialContext(ctx, network, addr2)
			if err != nil {
				return nil, err
			}
			// 监控使用，回写真实地址
			var uc = getUserCtx(ctx)
			if uc != nil {
				uc.upstreamIpPort = conn.RemoteAddr().String()
			}
			return conn, nil
		}
	}
}
