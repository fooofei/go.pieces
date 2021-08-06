package main

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/fooofei/xboom"
)

// 测试一个 HTTP 服务的 Performance

type httpBoomOp struct {
	Headers map[string]string
	Pld     []byte
	Addr    string
	// bullet
	HttpClient *http.Client
}

func (hb *httpBoomOp) LoadBullet(waitCtx context.Context, addr string) error {
	hb.HttpClient = &http.Client{}
	hb.Addr = addr
	return nil
}

func (hb *httpBoomOp) Shoot(waitCtx context.Context) (time.Duration, error) {
	start := time.Now()
	req, err := http.NewRequest("POST", hb.Addr, bytes.NewReader(hb.Pld))
	if err != nil {
		return time.Since(start), err
	}
	for k, v := range hb.Headers {
		req.Header.Set(k, v)
	}
	req = req.WithContext(waitCtx)
	resp, err := hb.HttpClient.Do(req)
	if err != nil {
		return time.Since(start), err
	}
	_ = resp.Body.Close()
	return time.Since(start), nil
}

func (hb *httpBoomOp) Close() error {
	hb.HttpClient = nil
	return nil
}

func main() {
	addr := "http://119.3.204.11:9200/server_info_report/_search?pretty"
	os.Args = append(os.Args, "-addr", addr)
	os.Args = append(os.Args, "-gocnt", "100")

	hb := &httpBoomOp{}
	hb.Headers = make(map[string]string, 0)
	hb.Headers["Content-Type"] = "application/json"
	hb.Pld = []byte(`{"sort":[{"timestamp":{"order":"desc"}}],"size":300}`)

	xboom.Gatelin(hb)
}
