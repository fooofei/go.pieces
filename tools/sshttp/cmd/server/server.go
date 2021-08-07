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
	"os/signal"
	"sync"
	"syscall"

	fnet "github.com/fooofei/go_pieces/pkg/net"
	"github.com/fooofei/go_pieces/tools/sshttp"
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

func pipeConn(app io.ReadWriteCloser, tun *sshttp.Tunnel) {
	closeBoth := func() {
		_ = app.Close()
		_ = tun.Close()
	}
	once := &sync.Once{}
	waitGrpBoth := &sync.WaitGroup{}
	closeCtx, cancel := context.WithCancel(tun.Ctx)
	waitGrpBoth.Add(1)
	go func() {
		pipeConnReadApp(app, tun)
		waitGrpBoth.Done()
		once.Do(closeBoth)
	}()

	waitGrpBoth.Add(1)
	go func() {
		pipeConnReadTun(app, tun)
		waitGrpBoth.Done()
		once.Do(closeBoth)
	}()

	go func() {
		select {
		case <-closeCtx.Done():
			once.Do(closeBoth)
		}
	}()
	waitGrpBoth.Wait()
	cancel()
}

func serve(ctx context.Context, conn net.Conn) error {
	t := &sshttp.Tunnel{
		SndNxt:        0,
		SndUna:        0,
		RcvNxt:        0,
		W:             conn,
		R:             bufio.NewReader(conn),
		C:             conn,
		Ctx:           ctx,
		AckedSndNxtCh: make(chan int64, 1000),
		CopyBuf:       make([]byte, 128*1024),
	}

	var req *http.Request
	var err error
	var httpPath *sshttp.HttpPath

	_, stop := fnet.CloseWhenContext(ctx, conn)
	defer stop()

	req, err = http.ReadRequest(t.R)
	if err != nil {
		return fmt.Errorf("fail read first http.ReadRequest, err %w", err)
	}
	httpPath, err = sshttp.ParseUrlPath(req.URL)
	if err != nil {
		return fmt.Errorf("fail path= %v, err %w", req.URL, err)
	}
	log.Printf("first Response = %v", httpPath)

	if httpPath.Type != "login" {
		return fmt.Errorf("httpPath.Type %v != login, err %w", httpPath, err)
	}
	// send hello back
	req, err = sshttp.NewLogin()
	if err != nil {
		return fmt.Errorf("failed ssh newLogin err %w", err)
	}
	err = req.Write(t.W)

	req, err = http.ReadRequest(t.R)
	if err != nil {
		return fmt.Errorf("fail read second http.ReadRequest err %w", err)
	}
	httpPath, err = sshttp.ParseUrlPath(req.URL)
	if err != nil {
		return fmt.Errorf("fail path= %v, err %w", req.URL, err)
	}
	if httpPath.Type != "proxy" {
		return fmt.Errorf("Unexpected second Reponse = %v ", httpPath)
	}
	sshdAddr := req.Header.Get("Connect")
	if sshdAddr == "" {
		return fmt.Errorf("empty sshdAddr in Connect")
	}

	d := net.Dialer{}
	app, err := d.DialContext(ctx, "tcp", sshdAddr)
	if err != nil {
		return fmt.Errorf("fail Dial err %w", err)
	}
	pipeConn(app, t)
	log.Printf("leave serve")
	return nil
}

func main() {
	var err error
	var ctx context.Context
	var cancel context.CancelFunc

	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.SetPrefix(fmt.Sprintf("pid =%v ", os.Getpid()))

	addr := ":3389"
	ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	waitGrp := &sync.WaitGroup{}

	lc := &net.ListenConfig{}
	conn, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("serve at %v", addr)
	wait, closerWaitGrp := fnet.CloseWhenContext(ctx, conn)
acceptLoop:
	for {
		subconn, err := conn.Accept()
		if err != nil {
			log.Printf("err= %v", err)
			select {
			case <-ctx.Done():
				break acceptLoop
			default:
			}
			continue
		}
		log.Printf("Accept from %v", subconn.RemoteAddr())
		waitGrp.Add(1)
		go func(ctx context.Context, conn net.Conn) {
			err = serve(ctx, conn)
			if err != nil {
				_ = conn.Close()
				log.Printf("serve err= %v", err)
			}
			waitGrp.Done()
		}(ctx, subconn)

	}

	waitGrp.Wait()
	closerWaitGrp()
	<-wait.Done()
	log.Printf("main exit")
}
