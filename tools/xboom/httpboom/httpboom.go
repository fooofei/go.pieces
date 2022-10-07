package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"

	"github.com/fooofei/go_pieces/tools/xboom"
)

// 测试一个 HTTP 服务的 Performance

type httpBoom struct {
	Headers map[string]string
	Pld     []byte
	Addr    string
	// bullet
	HttpClient *http.Client
}

func (hb *httpBoom) LoadBullet(waitCtx context.Context, addr string) error {
	hb.HttpClient = &http.Client{}
	hb.Addr = addr
	return nil
}

func (hb *httpBoom) Shoot(waitCtx context.Context) error {
	var req, err = http.NewRequestWithContext(waitCtx, "POST", hb.Addr, bytes.NewReader(hb.Pld))
	if err != nil {
		return err
	}
	for k, v := range hb.Headers {
		req.Header.Set(k, v)
	}
	var resp *http.Response
	if resp, err = hb.HttpClient.Do(req); err != nil {
		return err
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()
	return nil
}

func (hb *httpBoom) Close() error {
	hb.HttpClient = nil
	return nil
}

func main() {
	addr := "http://119.3.204.11:9200/server_info_report/_search?pretty"
	os.Args = append(os.Args, "-addr", addr)
	os.Args = append(os.Args, "-gocnt", "100")

	hb := &httpBoom{}
	hb.Headers = make(map[string]string, 0)
	hb.Headers["Content-Type"] = "application/json"
	hb.Pld = []byte(`{"sort":[{"timestamp":{"order":"desc"}}],"size":300}`)

	xboom.Gatelin(hb)
}
