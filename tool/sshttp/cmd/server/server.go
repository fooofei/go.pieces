package main

import (
	"context"
	"net"
)

func main() {

	d := net.ListenConfig{}
	waitCtx, cancel := context.WithCancel(context.Background())

	addr := ":8888"
	lsnConn, err := d.Listen(waitCtx, "tcp", addr)

	_ = err
	_ = lsnConn
	_ = cancel
}
