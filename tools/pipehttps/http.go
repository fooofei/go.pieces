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
			// addr is must with port format
			var addr2 = addr
			if host, port, err := net.SplitHostPort(addr); err == nil {
				addr2 = mapToHostPort(host, port, m)
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

// WithCustomDomain2 跟 WithCustomDomain 一样，是另一种用法
func WithCustomDomain2(m map[string]string) func(*http.Request) {
	return func(req *http.Request) {
		var defaultPort = "80"
		if strings.EqualFold(req.URL.Scheme, "https") {
			defaultPort = "443"
		}
		var srcHost = req.URL.Host
		// 与上面的函数的区别是，这里的可能没有端口
		if host, port, err := net.SplitHostPort(srcHost); err != nil {
			req.URL.Host = mapToHostPort(host, defaultPort, m)
		} else {
			req.URL.Host = mapToHostPort(host, port, m)
		}
		// req.Host 决定了 http 头部是什么
		// req.URL.Host 决定了 tcp 连接建立到哪里
	}
}

// mapToHostPort 把 host port 根据 m 映射到另外的  hostPort
// 优先映射端口，同时有 "a.com" -> "1.1.1.1" "a.com:443" -> "1.1.1.1:80"  map a.com:443 will result to 1.1.1.1:80
func mapToHostPort(host, port string, m map[string]string) string {
	var hostPort = net.JoinHostPort(host, port)
	if portHost, ok := m[hostPort]; ok {
		return portHost
	}

	if noPortHost, ok := m[host]; ok {
		return net.JoinHostPort(noPortHost, port) // 只是更换 ip， 不更换端口
	}
	// default with src
	return net.JoinHostPort(host, port)
}

