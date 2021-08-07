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

// 测试连接 TCP-TLS 端口 发送消息

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
	// 从 client 角度，限定使用的加密套件
	// 看我们能否使用 wireshark 成功解密
	// 效果：成功了
	// 当用 xx_RSA_xx 系列的加密套件才能使用 wireshark 解密
	// xx_ECDHE_xxx 系列的加密套件就不能使用 wireshark 解密了
	tc.CipherSuites= []uint16{
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	}
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
