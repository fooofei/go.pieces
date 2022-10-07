package main

import (
	"context"
	"encoding/hex"
	"net"

	"github.com/fooofei/go_pieces/tools/xboom"
)

type udpBoom struct {
	RAddr string
	Pld   []byte
	Conn  net.Conn
}

func (ub *udpBoom) LoadBullet(waitCtx context.Context, addr string) error {
	var d = net.Dialer{}
	var err error
	if ub.Conn, err = d.DialContext(waitCtx, "udp", ub.RAddr); err != nil {
		return err
	}
	return nil
}

func (ub *udpBoom) Shoot(waitCtx context.Context) error {
	var err error
	_, err = ub.Conn.Write(ub.Pld)
	return err
}

func (ub *udpBoom) Close() error {
	err := ub.Conn.Close()
	ub.Conn = nil
	return err
}

func main() {
	ub := &udpBoom{}
	ub.Pld, _ = hex.DecodeString(``)
	ub.RAddr = "114.116.111.96:5683"

	xboom.Gatelin(ub)

}
