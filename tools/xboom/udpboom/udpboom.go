package main

import (
	"context"
	"encoding/hex"
	"net"
	"time"

	"github.com/fooofei/tools/xboom"
)

type udpBoomOp struct {
	RAddr string
	Pld   []byte
	Conn  net.Conn
}

func (ub *udpBoomOp) LoadBullet(waitCtx context.Context, addr string) error {
	d := net.Dialer{}
	var err error
	ub.Conn, err = d.DialContext(waitCtx, "udp", ub.RAddr)
	if err != nil {
		return err
	}
	return nil
}

func (ub *udpBoomOp) Shoot(waitCtx context.Context) (time.Duration, error) {
	start := time.Now()
	var err error
	_, err = ub.Conn.Write(ub.Pld)
	return time.Since(start), err
}

func (ub *udpBoomOp) Close() error {
	err := ub.Conn.Close()
	ub.Conn = nil
	return err
}

func main() {
	ub := &udpBoomOp{}
	ub.Pld, _ = hex.DecodeString(``)
	ub.RAddr = "114.116.111.96:5683"

	xboom.Gatelin(ub)

}
