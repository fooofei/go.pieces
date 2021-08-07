package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/fooofei/tools/sshttp"
)

func usage(program string) string {
	fmt1 := `
Usage:
	ssh -o "ProxyCommand %v proxy_server_host proxy_server_port %vh %vp" user@host
	proxy_server_host: hostname on which Proxy Server runs on
	proxy_server_port: TCP port number to connect to Proxy Server
	user: SSH user
	host: SSH host

Example:
	ssh -o "ProxyCommand %v 127.0.0.1 8888 %vh %vp" work@192.168.200.128
`
	return fmt.Sprintf(fmt1, program, "%%", "%%", program, "%%", "%%")
}

func pipeConnReadApp(app io.Reader, tun *sshttp.Tunnel) {

	buf := make([]byte, 128*1024)
pipeLoop:
	for {
		nr, er := app.Read(buf)
		if nr > 0 {
			_, ew := tun.Write(buf[:nr])
			if ew != nil {
				log.Printf("err= %v", ew)
				break pipeLoop
			}
		}
		if er != nil {
			log.Printf("err= %v", er)
			break pipeLoop
		}
	}
}
func pipeConnReadTun(app io.Writer, tun *sshttp.Tunnel) {

pipeLoop:
	for {
		nw, ew := tun.WriteTo(app)
		if ew != nil {
			log.Printf("WriteTo err= %v", ew)
			break pipeLoop
		}
		_ = nw
	}
}

func bypassProxy(t *sshttp.Tunnel) {
	var err error
	var req *http.Request

	log.Printf("enter login")
	req, err = sshttp.NewLogin()
	if err != nil {
		log.Fatal(err)
	}
	err = req.WriteProxy(t.W)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("leave login")

	log.Printf("enter Read login")
	req, err = http.ReadRequest(t.R)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("leave Read login")
}

func pipeConn(ctx context.Context, app io.ReadWriter, tun io.ReadWriteCloser,
	sshdAddr string) {
	var req *http.Request
	var err error
	t := &sshttp.Tunnel{
		SndNxt:        0,
		SndUna:        0,
		RcvNxt:        0,
		R:             bufio.NewReader(tun),
		W:             tun,
		Ctx:           ctx,
		AckedSndNxtCh: make(chan int64, 1000),
		CopyBuf:       make([]byte, 128*1024),
	}

	bypassProxy(t)
	req, err = sshttp.NewProxyConnect(sshdAddr)
	if err != nil {
		log.Fatal(err)
	}
	err = req.Write(t.W)
	if err != nil {
		log.Fatal(err)
	}

	waitGrp := &sync.WaitGroup{}
	waitGrp.Add(1)
	once := &sync.Once{}
	onceFunc := func() {
		_ = tun.Close()
	}
	go func() {
		pipeConnReadApp(app, t)
		waitGrp.Done()
		once.Do(onceFunc)
	}()

	waitGrp.Add(1)
	go func() {
		pipeConnReadTun(app, t)
		waitGrp.Done()
		once.Do(onceFunc)
	}()
	waitGrp.Wait()
}

func serve(ctx context.Context) {
	pro := os.Args[0]
	fmt.Printf("%v", usage(pro))

	host := os.Args[1]
	port := os.Args[2]
	sshdHost := os.Args[3]
	sshdPort := os.Args[4]

	addr := net.JoinHostPort(host, port)
	sshdAddr := net.JoinHostPort(sshdHost, sshdPort)

	log.Printf("proxy %v sshdAddr %v", addr, sshdAddr)

	dialCtx, _ := context.WithCancel(ctx)
	d := net.Dialer{}
	tunConn, err := d.DialContext(dialCtx, "tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	pipeConn(ctx, &sshttp.SSHPipeRW{}, tunConn, sshdAddr)
}

func main() {
	var waitCtx context.Context
	var cancel context.CancelFunc

	log.SetFlags(log.Lshortfile | log.LstdFlags)
	waitCtx, cancel = context.WithCancel(context.Background())

	if len(os.Args) < 5 {
		fmt.Printf(usage(os.Args[0]))
		return
	}
	serve(waitCtx)
	log.Printf("main exit")
	cancel()
}

// send request not wait response
// use this for bypass some http proxy auth
// url.HostName()+ net.DefaultPort()
//https://github.com/golang/go/issues/16142
func httpWrite(ctx context.Context, addr string, req *http.Request) error {
	req.Header.Set("User-Agent", sshttp.DefaultUserAgent)
	req.Header.Set("Proxy-Connection", "keep-alive")
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	err = req.Write(conn)
	_ = conn.Close()
	return err
}
