package main

import (
	"context"
	"github.com/fooofei/go_pieces/tools/ping/pkg/prober"
	"net"
)

type tcpProbe struct {
	Dialer *net.Dialer
}

func (t *tcpProbe) Probe(waitCtx context.Context, raddr string) (string, error) {
	var cnn, err = t.Dialer.DialContext(waitCtx, "tcp", raddr)
	if err != nil {
		return "", err
	}
	cnn.Close()
	return "", nil
}

func (t *tcpProbe) Ready(ctx context.Context, raddr string) error {
	t.Dialer = &net.Dialer{}
	return nil
}

func (t *tcpProbe) Name() string {
	return "tcpProbe"
}

func (t *tcpProbe) Example() string {
	return "tcping 127.0.0.1:22"
}

func (t *tcpProbe) Close() error {
	return nil
}

func main() {
	var probe = new(tcpProbe)
	prober.Do(probe)
}
