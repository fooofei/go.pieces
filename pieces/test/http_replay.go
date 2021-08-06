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


func WithDumpReq() func(io.Writer, *http.Request) {
	return func(w io.Writer, req *http.Request) {
		// dump must before .Do()
		content, err := httputil.DumpRequest(req, true)
		if err != nil {
			_, _ = fmt.Fprintf(w, "error: <%T>%v\n", err, err)
			return
		}
		_, _ = fmt.Fprintf(w, "%s\n", content)
	}
}

func WithDumpResp() func(io.Writer, *http.Response) {
	return func(w io.Writer, resp *http.Response) {
		content, err := httputil.DumpResponse(resp, true)
		if err != nil {
			_, _ = fmt.Fprintf(w, "error: <%T>%v\n", err, err)
			return
		}
		_, _ = fmt.Fprintf(w, "%s\n", content)
	}
}
