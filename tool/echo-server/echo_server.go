package main

import (
	"context"
	rlimt "echo_server/rlimit"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// echoContext defines the echo server context
type echoContext struct {
	LAddr   string
	WaitCtx context.Context
	Wg      *sync.WaitGroup
	StatDur time.Duration
	AddCnn  int64
	SubCnn  int64
}

func setupSignal(ectx *echoContext, cancel context.CancelFunc) {

	sigCh := make(chan os.Signal, 2)

	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)

	ectx.Wg.Add(1)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ectx.WaitCtx.Done():
		}
		ectx.Wg.Done()
	}()
}

func takeOverCnnClose(ectx *echoContext, cnn io.Closer) chan bool {
	cnnClosedCh := make(chan bool, 1)
	ectx.Wg.Add(1)
	go func() {
		select {
		case <-cnnClosedCh:
		case <-ectx.WaitCtx.Done():
		}
		_ = cnn.Close()
		ectx.Wg.Done()
	}()
	return cnnClosedCh
}

func echoConn(ctx *echoContext, cnn net.Conn) {

	atomic.AddInt64(&ctx.AddCnn, 1)
	cnnClosedCh := takeOverCnnClose(ctx, cnn)

	//copy until EOF
	buf := make([]byte, 128*1024)
	_, _ = io.CopyBuffer(cnn, cnn, buf)
	close(cnnClosedCh)
	atomic.AddInt64(&ctx.SubCnn, 1)
}

func listenAndServe(ectx *echoContext) {
	cnn, err := net.Listen("tcp", ectx.LAddr)
	if err != nil {
		log.Fatal(err)
	}
	// a routine to wake up accept()
	cnnClosedCh := takeOverCnnClose(ectx, cnn)

loop:
	for {
		cltCnn, err := cnn.Accept()
		if err != nil {
			break loop
		}

		ectx.Wg.Add(1)
		go func(arg0 *echoContext, arg1 net.Conn) {
			echoConn(arg0, arg1)
			ectx.Wg.Done()
		}(ectx, cltCnn)

	}
	close(cnnClosedCh)
}
func stat(ectx *echoContext) {
	tick := time.NewTicker(ectx.StatDur)
loop1:
	for {
		select {
		case <-ectx.WaitCtx.Done():
			break loop1
		case <-tick.C:
			log.Printf("stat AddCnn= %v SubCnn= %v Add-Sub= %v",
				ectx.AddCnn, ectx.SubCnn, ectx.AddCnn-ectx.SubCnn)
		}
	}
}

func main() {
	// log
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))

	var cancel context.CancelFunc
	ectx := new(echoContext)
	ectx.WaitCtx, cancel = context.WithCancel(context.Background())
	ectx.Wg = new(sync.WaitGroup)
	ectx.StatDur = time.Second * 5

	flag.StringVar(&ectx.LAddr, "laddr", ":3389", "The local listen addr")
	flag.Parse()
	rlimt.BreakOpenFilesLimit()
	log.Printf("working on \"%v\"", ectx.LAddr)
	setupSignal(ectx, cancel)

	// stat
	ectx.Wg.Add(1)
	go func() {
		stat(ectx)
		ectx.Wg.Done()
	}()

	listenAndServe(ectx)

	ectx.Wg.Wait()
	log.Printf("main exit")
}
