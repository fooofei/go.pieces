package main

import (
	"context"
	"encoding/hex"
	"net"
	"time"

	"github.com/fooofei/ping/pkg/pinger"
)

type udpingOp struct {
	Cnn net.Conn
	Pld []byte
	Buf []byte
}

func (t *udpingOp) Ping(waitCtx context.Context, raddr string) (time.Duration, error) {
	var err error
	noDeadline := time.Time{}
	_ = raddr

	start := time.Now()
	_, err = t.Cnn.Write(t.Pld)
	if err != nil {
		return time.Now().Sub(start), err
	}
	dl, ok := waitCtx.Deadline()
	if !ok {
		dl = time.Now().Add(time.Millisecond * 950)
	}
	_ = t.Cnn.SetReadDeadline(dl)
	_, err = t.Cnn.Read(t.Buf)
	_ = t.Cnn.SetReadDeadline(noDeadline)
	return time.Now().Sub(start), err
}

func (t *udpingOp) Ready(ctx context.Context, raddr string) error {
	var err error
	d := &net.Dialer{}
	t.Cnn, err = d.DialContext(ctx, "udp", raddr)
	if err != nil {
		return err
	}
	t.Buf = make([]byte, 100*1024)
	t.Pld, err = hex.DecodeString("hello")
	_ = err
	return nil
}

func (t *udpingOp) Name() string {
	return "UDPing"
}

func (t *udpingOp) Close() error {
	var err error
	if t.Cnn != nil {
		err = t.Cnn.Close()
		t.Cnn = nil
	}
	return err
}

func main() {
	op := new(udpingOp)
	pinger.DoPing(op)
}
