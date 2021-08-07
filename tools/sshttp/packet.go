package sshttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko)" +
		" Chrome/76.0.3809.100 Safari/537.36"
)

type HttpPath struct {
	Type string
	Seq  int64
	Ack  int64
}

func (hp *HttpPath) String() string {
	w := &bytes.Buffer{}
	_, _ = fmt.Fprintf(w, "/%v/%v/%v", hp.Type, hp.Seq, hp.Ack)
	return w.String()
}

func ParseHTTPPath(path string) (*HttpPath, error) {
	hp := &HttpPath{}
	tsa := strings.Split(path, "/")
	if len(tsa) < 4 {
		return nil, fmt.Errorf("short head http path \"%v\"", path)
	}
	hp.Type = tsa[1]
	v1, err := strconv.ParseFloat(tsa[2], 64)
	if err != nil {
		return nil, err
	}
	hp.Seq = int64(v1)
	v1, err = strconv.ParseFloat(tsa[3], 64)
	if err != nil {
		return nil, err
	}
	hp.Ack = int64(v1)
	return hp, nil
}

func ParseUrlPath(u *url.URL) (*HttpPath, error) {
	return ParseHTTPPath(u.Path)
}

func NewDataRequest(seq int64, ack int64, body []byte) (*http.Request, error) {
	return NewRequest("data", seq, ack, bytes.NewReader(body))
}

func NewRequest(reqType string, seq int64, ack int64, body io.Reader) (*http.Request, error) {
	url1 := fmt.Sprintf("/%v/%v/%v", reqType, seq, ack)
	r, err := http.NewRequest("POST", url1, body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("User-Agent", DefaultUserAgent)
	r.Header.Set("Proxy-Connection", "keep-alive")
	return r, err
}

func NewLogin() (*http.Request, error) {
	return NewRequest("login", 0, 0, nil)
}

func NewProxyConnect(value string) (*http.Request, error) {
	req, err := NewRequest("proxy", 0, 0, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Connect", value)
	return req, nil
}
