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

type session struct {
    Raddr     string
    TxBytes   uint64
    RxBytes   uint64
    BlockSize int // tcp payload
    StartTime time.Time
}

type ppContext struct {
    WaitCtx context.Context
    Wg      *sync.WaitGroup
}

func pingPong(ctx *ppContext, ssn *session, cnn net.Conn) {

    b := make([]byte, 128*1024)
loop:
    for {

        n, err := cnn.Read(b)
        if err != nil {
            log.Printf("rx err=%v", err)
            break loop
        }
        atomic.AddUint64(&ssn.RxBytes, uint64(n))
        _, err = cnn.Write(b[:n])
        if err != nil {
            log.Printf("tx err=%v", err)
            break loop
        }
        atomic.AddUint64(&ssn.TxBytes, uint64(n))
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
func takeoverCnnClose(ctx *ppContext, cnn io.Closer) chan bool {

    ch := make(chan bool, 1)

    ctx.Wg.Add(1)
    go func() {
        select {
        case <-ctx.WaitCtx.Done():
        case <-ch:
        }
        _ = cnn.Close()
        ctx.Wg.Done()
    }()

    return ch
}

func main() {

    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))

    ctx := new(ppContext)
    var cancel context.CancelFunc
    var err error
    //
    ctx.WaitCtx, cancel = context.WithCancel(context.Background())
    ctx.Wg = new(sync.WaitGroup)
    setupSignal(ctx, cancel)

    ssn := new(session)
    flag.StringVar(&ssn.Raddr, "raddr", "", "TCP remote addr")
    flag.IntVar(&ssn.BlockSize, "blocksize", 1500, "TCP payloadsize")
    flag.Parse()
    if ssn.Raddr == "" {
        flag.PrintDefaults()
        cancel()
        return
    }

    dia := &net.Dialer{}
    cnn, err := dia.DialContext(ctx.WaitCtx, "tcp", ssn.Raddr)
    if err != nil {
        cancel()
        log.Fatal(err)
    }
    cnnCloseCh := takeoverCnnClose(ctx, cnn)
    defer close(cnnCloseCh)

    // for start
    bb := new(bytes.Buffer)
    log.Printf("start pingpong")
    for i := 0; i < ssn.BlockSize; i++ {
        _ = bb.WriteByte(byte(i % 128))
    }
    n, _ := cnn.Write(bb.Bytes())
    atomic.AddUint64(&ssn.TxBytes, uint64(n))
    ssn.StartTime = time.Now()
    ctx.Wg.Add(1)
    go func() {
        pingPong(ctx, ssn, cnn)
        cancel()
        ctx.Wg.Done()
    }()

    //

    //
    statTick := time.NewTicker(time.Second * 5)
statLoop:
    for {

        select {
        case <-ctx.WaitCtx.Done():
            break statLoop
        case <-statTick.C:
            dur := time.Since(ssn.StartTime)
            sec := dur.Seconds()
            if sec > 0 {

                fmt.Printf("time take %v\n", sec)
                fmt.Printf("rx %v/%v=%.3f MiB/s\n", ssn.RxBytes, sec, float64(ssn.RxBytes)/(sec*1024*1024))
                fmt.Printf("tx %v/%v=%.3f MiB/s\n", ssn.TxBytes, sec, float64(ssn.TxBytes)/(sec*1024*1024))
            }

        }
    }

    ctx.Wg.Wait()
    log.Printf("main exit")
}
