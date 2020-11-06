package go_pieces

import (
	"bufio"
	"bytes"
	"context"
	"net/http"
)

// 记录如何重放 http 请求


func template(ctx context.Context, body []byte) (*http.Request, error) {
	var err error
	content := bytes.TrimSpace(body)
	content = append(content, []byte("\r\n\r\n")...)
	reader := bufio.NewReader(bytes.NewReader(content))
	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.RequestURI = "" // must clear
	req.URL.Scheme = "http"
	req.URL.Host = req.Host // must set
	return req, nil
}
