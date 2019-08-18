package main

import (
	"bufio"
	"context"
	"github.com/pkg/errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fooofei/sshttp"
)

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

func serve(ctx context.Context, conn net.Conn) error {
	t := &sshttp.Tunnel{
		SndNxt:        0,
		SndUna:        0,
		RcvNxt:        0,
		W:             conn,
		R:             bufio.NewReader(conn),
		Ctx:           ctx,
		AckedSndNxtCh: make(chan int64, 1000),
		CopyBuf:       make([]byte, 128*1024),
	}

	var req *http.Request
	var err error
	var httpPath *sshttp.HttpPath

	req, err = http.ReadRequest(t.R)
	if err != nil {
		return errors.Wrapf(err, "fail read first http.ReadRequest")
	}
	httpPath, err = sshttp.ParseUrlPath(req.URL)
	if err != nil {
		return errors.Wrapf(err, "fail path= %v", req.URL)
	}
	log.Printf("first Response = %v", httpPath)

	req, err = http.ReadRequest(t.R)
	if err != nil {
		return errors.Wrapf(err, "fail read second http.ReadRequest")
	}
	httpPath, err = sshttp.ParseUrlPath(req.URL)
	if err != nil {
		return errors.Wrapf(err, "fail path= %v", req.URL)
	}
	if httpPath.Type != "proxy" {
		return errors.Errorf("Unexpected second Reponse = %v ", httpPath)
	}
	sshdAddr := req.Header.Get("Connect")
	if sshdAddr == "" {
		return errors.Errorf("empty sshdAddr in Connect")
	}

	d := net.Dialer{}
	app, err := d.DialContext(ctx, "tcp", sshdAddr)
	if err != nil {
		return errors.Wrapf(err, "fail Dial")
	}

	closeOnceFunc := func() {
		_ = app.Close()
		_ = conn.Close()
	}
	closeOnce := &sync.Once{}
	waitGrp := &sync.WaitGroup{}
	closeCtx, cancel := context.WithCancel(ctx)
	waitGrp.Add(1)
	go func() {
		pipeConnReadApp(app, t)
		waitGrp.Done()
		closeOnce.Do(closeOnceFunc)
	}()
	
	waitGrp.Add(1)
	go func() {
		pipeConnReadTun(app, t)
		waitGrp.Done()
		closeOnce.Do(closeOnceFunc)
	}()

	go func() {
		select {
		case <-closeCtx.Done():
			closeOnce.Do(closeOnceFunc)
		}
	}()
	waitGrp.Wait()
	cancel()
	log.Printf("leave serve")
	return nil
}

func setupSignal(waitCtx context.Context, waitGrp *sync.WaitGroup, cancel context.CancelFunc) {

	sigCh := make(chan os.Signal, 2)

	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)

	waitGrp.Add(1)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-waitCtx.Done():
		}
		waitGrp.Done()
	}()
}

func main() {
	var err error
	var waitCtx context.Context
	var cancel context.CancelFunc

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	addr := ":3389"
	waitCtx, cancel = context.WithCancel(context.Background())
	waitGrp := &sync.WaitGroup{}

	setupSignal(waitCtx, waitGrp, cancel)
	lc := &net.ListenConfig{}
	conn, err := lc.Listen(waitCtx, "tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

acceptLoop:
	for {
		conn, err := conn.Accept()
		if err != nil {
			log.Printf("err= %v", err)
			select {
			case <-waitCtx.Done():
				break acceptLoop
			default:
			}
			continue
		}
		log.Printf("Accept from %v", conn.RemoteAddr())
		waitGrp.Add(1)
		go func(conn net.Conn) {
			err = serve(waitCtx, conn)
			if err != nil {
				_ = conn.Close()
				log.Printf("serve err= %v", err)
			}
			waitGrp.Done()
		}(conn)

	}
	waitGrp.Wait()
	_ = conn.Close()
	log.Printf("main exit")
}
