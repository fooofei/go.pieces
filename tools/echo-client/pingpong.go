package main

import (
	"bytes"
	"context"
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

type ppContext struct {
	WaitCtx context.Context
	Wg      *sync.WaitGroup
	//
	RAddr     string
	TxSize    int64
	RxSize    int64
	BlockSize int64 // tcp payload
	StartTime time.Time
}

func (ppCtx *ppContext) state() string {
	dur := time.Since(ppCtx.StartTime)
	interval := dur.Seconds()
	w := &bytes.Buffer{}
	if interval > 0 {
		unit := interval * 1024 * 1024
		rx := atomic.LoadInt64(&ppCtx.RxSize)
		tx := atomic.LoadInt64(&ppCtx.TxSize)
		_, _ = fmt.Fprintf(w, "time take %v (s)\n", int64(interval))
		_, _ = fmt.Fprintf(w, "rx %v/%v= %.3f MiB/s\n", rx, int64(interval), float64(rx)/unit)
		_, _ = fmt.Fprintf(w, "tx %v/%v= %.3f MiB/s", tx, int64(interval), float64(tx)/unit)
	}
	return w.String()
}

func pingPong(ppCtx *ppContext, cnn net.Conn) {

	b := make([]byte, 128*1024)
	for {
		nr, er := cnn.Read(b)
		if nr > 0 {
			atomic.AddInt64(&ppCtx.RxSize, int64(nr))
			nw, ew := cnn.Write(b[:nr])
			if nw > 0 {
				atomic.AddInt64(&ppCtx.TxSize, int64(nw))
			}
			if ew != nil {
				log.Printf("write err= %v", ew)
				break
			}
		}
		if er != nil {
			log.Printf("read err= %v", er)
			break
		}
	}
}

func setupSignal(ctx *ppContext, cancel context.CancelFunc) {

	sigCh := make(chan os.Signal, 2)

	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)

	ctx.Wg.Add(1)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.WaitCtx.Done():
		}
		ctx.Wg.Done()
	}()
}

func takeOverCnnClose(waitCtx context.Context, cnn io.Closer) (chan bool, *sync.WaitGroup) {
	noWait := make(chan bool, 1)
	waitGrp := &sync.WaitGroup{}
	waitGrp.Add(1)
	go func() {
		select {
		case <-noWait:
		case <-waitCtx.Done():
		}
		_ = cnn.Close()
		waitGrp.Done()
	}()
	return noWait, waitGrp
}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))

	ppCtx := &ppContext{}
	var cancel context.CancelFunc
	var err error
	var i int64
	//
	ppCtx.WaitCtx, cancel = context.WithCancel(context.Background())
	ppCtx.Wg = new(sync.WaitGroup)
	setupSignal(ppCtx, cancel)

	flag.StringVar(&ppCtx.RAddr, "raddr", "127.0.0.1:3389", "TCP remote addr")
	flag.Int64Var(&ppCtx.BlockSize, "blocksize", 1500, "TCP payloadsize")
	flag.Parse()

	dia := &net.Dialer{}
	cnn, err := dia.DialContext(ppCtx.WaitCtx, "tcp", ppCtx.RAddr)
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	noWait, waitGrp := takeOverCnnClose(ppCtx.WaitCtx, cnn)

	// for start
	bb := new(bytes.Buffer)
	log.Printf("start pingpong")
	for i = 0; i < ppCtx.BlockSize; i++ {
		_ = bb.WriteByte(byte(i % 128))
	}
	hello := bb.Bytes()
	log.Printf("Write hello bytes size= %v", len(hello))
	n, _ := cnn.Write(hello)
	atomic.AddInt64(&ppCtx.TxSize, int64(n))
	ppCtx.StartTime = time.Now()
	ppCtx.Wg.Add(1)
	go func() {
		pingPong(ppCtx, cnn)
		cancel()
		ppCtx.Wg.Done()
	}()

	statTick := time.NewTicker(time.Second * 5)
statLoop:
	for {

		select {
		case <-ppCtx.WaitCtx.Done():
			break statLoop
		case <-statTick.C:
			fmt.Printf("%v\n", ppCtx.state())
		}
	}
	close(noWait)
	waitGrp.Wait()
	ppCtx.Wg.Wait()
	log.Printf("main exit")
}
