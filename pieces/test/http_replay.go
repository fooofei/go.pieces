package test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

// 记录如何重放 http 请求

/*

fastHttp
func getFastHTTPRequest() *fasthttp.Request {
	content, err := ioutil.ReadFile(fpath)
	if err != nil {
		panic(err)
	}
	content = bytes.TrimSpace(content)
	req := &fasthttp.Request{}
	err = req.Read(bufio.NewReader(bytes.NewReader(content)))
	if err != nil {
		panic(err)
	}
	return req
}

*/

// requestFromString 从一个字节流里解析为 request 对象
// 字节流的格式为 Wireshark 抓包看到的 HTTP 格式
// 在发送出去一定要干几件事，否则会发送失败
// 如果 request 是从 golang 自己的 handler 里拿到的，不是我们解析得到的，也要干这几件事才能发送出去
func requestFromString(ctx context.Context, content []byte) (*http.Request, error) {
	var err error
	var req *http.Request
	content = append(content, []byte("\r\n\r\n")...)
	reader := bufio.NewReader(bytes.NewReader(content))
	req, err = http.ReadRequest(reader)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.RequestURI = "" // must clear
	req.URL.Scheme = "http"
	req.URL.Host = req.Host // must set
	return req, nil
}

// WithDumpReq will dump request as http format
func WithDumpReq(w io.Writer) func(*http.Request) {
	return func(req *http.Request) {
		// dump must before .Do()
		content, err := httputil.DumpRequest(req, true)
		if err != nil {
			_, _ = fmt.Fprintf(w, "error: <%T>%v\n", err, err)
			return
		}
		_, _ = fmt.Fprintf(w, "%s\n", content)
	}
}

// WithDumpResp will dump response as http format
func WithDumpResp(w io.Writer) func(*http.Response) {
	return func(resp *http.Response) {
		content, err := httputil.DumpResponse(resp, true)
		if err != nil {
			_, _ = fmt.Fprintf(w, "error: <%T>%v\n", err, err)
			return
		}
		_, _ = fmt.Fprintf(w, "%s\n", content)
	}
}
