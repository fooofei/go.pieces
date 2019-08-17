package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func usage(program string) string {
	fmt1 := `
Usage:
	ssh -o "ProxyCommand %v proxy_server_host proxy_server_port %%h %%p" user@host
	proxy_server_host: hostname on which Proxy Server runs on
	proxy_server_port: TCP port number to connect to Proxy Server
	user: SSH user
	host: SSH host

Example:
	ssh -o "ProxyCommand %v 127.0.0.1 8888 %%h %%p" work@192.168.200.128
`
	return fmt.Sprintf(fmt1, program, program)
}

func proxy() {
	pro := os.Args[0]
	fmt.Printf("%v", usage(pro))

	host := os.Args[1]
	port := os.Args[2]
	sshdHost := os.Args[3]
	sshdPort := os.Args[4]

	addr := net.JoinHostPort(host, port)

	_ = addr
	_ = sshdPort
	_ = sshdHost
	// first login
	//_, _ = conn.Write(sshttp.NewLogin())
	//
	//_, _ = conn.Write(sshttp.NewProxyConnect(net.JoinHostPort(sshdHost, sshdPort)))

}

func main() {
	var waitCtx context.Context
	var cancel context.CancelFunc

	log.SetFlags(log.Lshortfile)
	waitCtx, cancel = context.WithCancel(context.Background())
	d := &websocket.Dialer{}

	reqHeader := http.Header{}
	reqHeader.Add("User-Agent", "Chrome76")
	addr := "ws://:3389/aa"

	conn, resp, err := d.DialContext(waitCtx, addr, reqHeader)
	if err != nil {
		log.Printf("resp= %v", resp)
		if resp != nil {
			log.Printf("code= %v ", resp.Status)
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				log.Printf("body = %s", body)
			}
		}
		log.Fatalf("err= %v conn= %v", err, conn)
	}

	log.Printf("go to loop")
	waitGrp := &sync.WaitGroup{}
	once := &sync.Once{}
	connCloseFunc := func() {
		_ = conn.Close()
	}
	waitGrp.Add(1)
	go func() {
	readLoop:
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				break readLoop
			}
			log.Printf("type= %v msg= len= %v \"%s\"", msgType, len(msg), msg)
		}

		waitGrp.Done()
		once.Do(connCloseFunc)
	}()

	waitGrp.Add(1)
	go func() {
		var cnt int64
		var err error
	writeLoop:
		for {
			select {
			case <-time.After(time.Second * 5):
				w := &bytes.Buffer{}
				_, _ = fmt.Fprintf(w, "hello %v", cnt)
				err = conn.WriteMessage(websocket.BinaryMessage, w.Bytes())
				log.Printf("write \"%s\" err= %v", w.String(), err)
				if err != nil {
					break writeLoop
				}
			}

		}
		waitGrp.Done()
	}()

	waitGrp.Wait()
	log.Printf("main exit")
	_ = cancel
}
