package main

import (
	"context"
	"github.com/fooofei/go_pieces/tools/ping/pkg/prober"
	"io"
	"net/http"
)

type httpProbe struct {
	Clt *http.Client
}

func (t *httpProbe) Probe(waitCtx context.Context, raddr string) (string, error) {
	var req, err = http.NewRequestWithContext(waitCtx, http.MethodGet, raddr, nil)
	if err != nil {
		return "", err
	}
	var resp *http.Response
	if resp, err = t.Clt.Do(req); err != nil {
		return "", err
	}
	// consume all body contents
	io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.Status, nil
}

func (t *httpProbe) Ready(ctx context.Context, raddr string) error {
	t.Clt = http.DefaultClient
	return nil
}

func (t *httpProbe) Name() string {
	return "httpProbe"
}

func (t *httpProbe) Example() string {
	return "httping http://127.0.0.1:8080"
}

func (t *httpProbe) Close() error {
	return nil
}

func main() {
	var p = new(httpProbe)
	prober.Do(p)
}
