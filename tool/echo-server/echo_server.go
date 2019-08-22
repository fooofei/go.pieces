package main

import (
	"bytes"
	"context"
	"crypto/tls"
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
	LAddr     string
	WaitCtx   context.Context
	Wg        *sync.WaitGroup
	StatDur   time.Duration
	StartTime time.Time
	TlsCfg    *tls.Config
	//
	AddCnn int64
	SubCnn int64
	RxSize int64
	TxSize int64
}

func (echoCtx *echoContext) state() string {
	dur := time.Since(echoCtx.StartTime)
	interval := dur.Seconds()
	w := &bytes.Buffer{}
	if interval > 0 {
		unit := interval * 1024 * 1024
		rx := atomic.LoadInt64(&echoCtx.RxSize)
		tx := atomic.LoadInt64(&echoCtx.TxSize)
		_, _ = fmt.Fprintf(w, "time take %v (s)\n", int64(interval))
		_, _ = fmt.Fprintf(w, "come connection = %v leave connection= %v\n",
			atomic.LoadInt64(&echoCtx.AddCnn),
			atomic.LoadInt64(&echoCtx.SubCnn))
		_, _ = fmt.Fprintf(w, "rx %v/%v= %.3f MiB/s\n", rx, int64(interval), float64(rx)/unit)
		_, _ = fmt.Fprintf(w, "tx %v/%v= %.3f MiB/s", tx, int64(interval), float64(tx)/unit)
	}
	return w.String()
}

func setupSignal(echoCtx *echoContext, cancel context.CancelFunc) {

	sigCh := make(chan os.Signal, 2)

	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)

	echoCtx.Wg.Add(1)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-echoCtx.WaitCtx.Done():
		}
		echoCtx.Wg.Done()
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

func echoConn(echoCtx *echoContext, rwc io.ReadWriteCloser) {

	atomic.AddInt64(&echoCtx.AddCnn, 1)
	noWait, waitGrp := takeOverCnnClose(echoCtx.WaitCtx, rwc)

	//copy until EOF
	buf := make([]byte, 128*1024)
	for {
		nr, er := rwc.Read(buf)
		if nr > 0 {
			atomic.AddInt64(&echoCtx.RxSize, int64(nr))
			nw, ew := rwc.Write(buf[:nr])
			if nw > 0 {
				atomic.AddInt64(&echoCtx.TxSize, int64(nw))
			}
			if ew != nil {
				break
			}

		}
		if er != nil {
			break
		}
	}
	close(noWait)
	waitGrp.Wait()
	atomic.AddInt64(&echoCtx.SubCnn, 1)
}

func listenAndServe(echoCtx *echoContext) {
	lc := net.ListenConfig{}
	cnn, err := lc.Listen(echoCtx.WaitCtx, "tcp", echoCtx.LAddr)
	if err != nil {
		log.Fatal(err)
	}
	if echoCtx.TlsCfg != nil {
		cnn = tls.NewListener(cnn, echoCtx.TlsCfg)
	}
	// a routine to wake up accept()
	noWait, waitGrp := takeOverCnnClose(echoCtx.WaitCtx, cnn)

loop:
	for {
		cltCnn, err := cnn.Accept()
		if err != nil {
			break loop
		}

		echoCtx.Wg.Add(1)
		go func(arg0 *echoContext, arg1 net.Conn) {
			echoConn(arg0, arg1)
			echoCtx.Wg.Done()
		}(echoCtx, cltCnn)

	}
	close(noWait)
	waitGrp.Wait()
}
func stat(echoCtx *echoContext) {
	tick := time.NewTicker(echoCtx.StatDur)
loop1:
	for {
		select {
		case <-echoCtx.WaitCtx.Done():
			break loop1
		case <-tick.C:
			log.Printf("%v\n", echoCtx.state())
		}
	}
}

func main() {
	// log
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))

	var cancel context.CancelFunc
	echoCtx := new(echoContext)
	echoCtx.WaitCtx, cancel = context.WithCancel(context.Background())
	echoCtx.Wg = new(sync.WaitGroup)
	echoCtx.StatDur = time.Second * 5
	echoCtx.StartTime = time.Now()
	certPath := ""
	privKeyPath := ""

	flag.StringVar(&echoCtx.LAddr, "laddr", ":3389", "The local listen addr")
	flag.StringVar(&certPath, "cert", "", "The cert file, PEM format, null will not use tls")
	flag.StringVar(&privKeyPath, "privkey", "", "The privkey file, PEM format, null will not use tls")
	flag.Parse()
	rlimt.BreakOpenFilesLimit()
	log.Printf("working on \"%v\"", echoCtx.LAddr)
	setupSignal(echoCtx, cancel)
	if certPath != "" && privKeyPath != "" {
		tlsCert, err := tls.LoadX509KeyPair(certPath, privKeyPath)
		if err != nil {
			log.Fatal(err)
		}
		echoCtx.TlsCfg = &tls.Config{}
		echoCtx.TlsCfg.Certificates = []tls.Certificate{tlsCert}
	}

	// stat
	echoCtx.Wg.Add(1)
	go func() {
		stat(echoCtx)
		echoCtx.Wg.Done()
	}()

	listenAndServe(echoCtx)

	echoCtx.Wg.Wait()
	log.Printf("main exit")
}
