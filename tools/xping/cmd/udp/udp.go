package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/fooofei/go_pieces/tools/ping/pkg/prober"
	"net"
	"time"
)

type udpProbe struct {
	Cnn         net.Conn
	SendContent []byte
	Buf         []byte
}

func (t *udpProbe) Probe(waitCtx context.Context, raddr string) (string, error) {
	var err error
	var readSize int
	var noDeadline = time.Time{}
	if _, err = t.Cnn.Write(t.SendContent); err != nil {
		return "", err
	}
	var deadline, ok = waitCtx.Deadline()
	if !ok {
		deadline = time.Now().Add(time.Millisecond * 950)
	}
	t.Cnn.SetReadDeadline(deadline)
	readSize, err = t.Cnn.Read(t.Buf)
	t.Cnn.SetReadDeadline(noDeadline)
	return fmt.Sprintf("ReadSize:%v", readSize), err
}

func (t *udpProbe) Ready(ctx context.Context, raddr string) error {
	var err error
	var d = &net.Dialer{}
	t.Cnn, err = d.DialContext(ctx, "udp", raddr)
	if err != nil {
		return fmt.Errorf("failed udp dial %s with error %w", raddr, err)
	}
	t.Buf = make([]byte, 100*1024)
	t.SendContent, err = hex.DecodeString("hello")
	return nil
}

func (t *udpProbe) Name() string {
	return "udpProbe"
}

func (t *udpProbe) Example() string {
	return "udping 127.0.0.1:53"
}

func (t *udpProbe) Close() error {
	var err error
	if t.Cnn != nil {
		err = t.Cnn.Close()
		t.Cnn = nil
	}
	return err
}

func main() {
	var probe = new(udpProbe)
	prober.Do(probe)
}
