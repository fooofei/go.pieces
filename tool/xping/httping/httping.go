package main

import (
	"context"
	"net/http"
	"time"

	"github.com/fooofei/ping/pkg/pinger"
)

type httpingOp struct {
	Clt *http.Client
}

func (t *httpingOp) Ping(waitCtx context.Context, raddr string) (time.Duration, error) {
	req, err := http.NewRequestWithContext(waitCtx, http.MethodGet, raddr, nil)
	if err != nil {
		return time.Duration(0), err
	}
	start := time.Now()
	resp, err := t.Clt.Do(req)
	dur := time.Now().Sub(start)
	if err != nil {
		return time.Duration(0), err
	}
	_ = resp.Body.Close()
	return dur, nil
}

func (t *httpingOp) Ready(ctx context.Context, raddr string) error {
	t.Clt = http.DefaultClient
	return nil
}

func (t *httpingOp) Name() string {
	return "HTTPing"
}

func (t *httpingOp) Close() error {
	return nil
}

func main() {
	p := new(httpingOp)
	pinger.DoPing(p)
}
