package go_pieces

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
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

func WithDumpRequest() func(*http.Request) {
	return func(req *http.Request) {
		// dump must before .Do()
		content, err := httputil.DumpRequest(req, true)
		if err == nil {
			fmt.Printf("%s\n\n", string(content))
		}
	}
}

func WithDumpResponse() func(*http.Response) {
	return func(resp *http.Response) {
		content, err := httputil.DumpResponse(resp, true)
		if err == nil {
			fmt.Printf("%s\n\n---------------------------------------------\n", string(content))
		}
	}
}
