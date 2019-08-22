package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	var ctx context.Context
	var cancel context.CancelFunc

	addr := "127.0.0.1:45678"
	d := net.Dialer{}
	ctx, cancel = context.WithCancel(context.Background())
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	tc := &tls.Config{InsecureSkipVerify:true}
	tlsCnn := tls.Client(conn, tc)
	err = tlsCnn.Handshake()
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 1024)
	for i := 0; i < 5; i++ {
		select {
		case <-time.After(time.Second * 1):
			w := &bytes.Buffer{}
			_, _ = fmt.Fprintf(w, "hello")
			_, err = tlsCnn.Write(w.Bytes())
			if err != nil {
				log.Fatal(err)
			}

			nr, _ := tlsCnn.Read(buf)
			if nr > 0 {
				log.Printf("read [%s]", buf[:nr])
			}
		}
	}

	_ = conn.Close()

	cancel()

}
