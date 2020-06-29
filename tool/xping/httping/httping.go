package main

import (
	"context"
	"net/http"
	"time"

	"github.com/fooofei/xping"
)

type httpingOp struct {
	Clt *http.Client
}

func (t *httpingOp) Ping(waitCtx context.Context, raddr string) (time.Duration, error) {
	req, err := http.NewRequest("GET", raddr, nil)
	if err != nil {
		return time.Duration(0), err
	}
	req = req.WithContext(waitCtx)
	start := time.Now()
	resp, err := t.Clt.Do(req)
	dur := time.Now().Sub(start)
	if err != nil {
		return time.Duration(0), err
	}
	_ = resp.Body.Close()
	return dur, nil
}

func (t *httpingOp) Ready(raddr string) error {
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
	xping.Ping(p)
}
