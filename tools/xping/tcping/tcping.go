package main

import (
	"context"
	"net"
	"time"

	"github.com/fooofei/tools/ping/pkg/pinger"
)

type tcpingOp struct {
	Dialer *net.Dialer
}

func (t *tcpingOp) Ping(waitCtx context.Context, raddr string) (time.Duration, error) {
	start := time.Now()
	cnn, err := t.Dialer.DialContext(waitCtx, "tcp", raddr)
	dur := time.Now().Sub(start)
	if err != nil {
		return dur, err
	}
	_ = cnn.Close()
	return dur, nil
}
func (t *tcpingOp) Ready(ctx context.Context, raddr string) error {
	t.Dialer = &net.Dialer{}
	return nil
}

func (t *tcpingOp) Name() string {
	return "TCPing"
}

func (t *tcpingOp) Close() error {
	return nil
}

func main() {
	op := new(tcpingOp)
	pinger.DoPing(op)
}
