package main

import (
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

type EchoContext struct {
    Laddr   string
    WaitCtx context.Context
    Wg      *sync.WaitGroup
    StatDur time.Duration
    AddCnn  uint32
    SubCnn  uint32
}

func SetupSignal(ctx *EchoContext, cancel context.CancelFunc) {

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

func EchoConn(ctx *EchoContext, cnn net.Conn) {

    atomic.AddUint32(&ctx.AddCnn, 1)
    cnnClosedCh := make(chan bool, 1)
    ctx.Wg.Add(1)
    go func() {
        select {
        case <-cnnClosedCh:
        case <-ctx.WaitCtx.Done():
            _ = cnn.Close()
        }
        ctx.Wg.Done()
    }()

    // copy until EOF
    _, _ = io.Copy(cnn, cnn)
    close(cnnClosedCh)
    _ = cnn.Close()
    ctx.Wg.Done()
    atomic.AddUint32(&ctx.SubCnn, 1)
}

func BrkOpenFilesLimit() {
    var err error
    var rlim syscall.Rlimit
    var limit uint64 = 1000 * 1000
    err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
    if err != nil {
        log.Fatalf("Getrlimit err= %v", err)
    }
    rlim.Cur = limit + uint64(100)
    rlim.Max = limit + uint64(100)
    err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
    if err != nil {
        log.Fatalf("Setrlimit err= %v", err)
    }
}

func main() {
    // log
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))

    ctx := new(EchoContext)
    flag.StringVar(&ctx.Laddr, "laddr", "", "The local listen addr")
    flag.Parse()
    if ctx.Laddr == "" {
        flag.PrintDefaults()
        return
    }

    var cancel context.CancelFunc
    var err error

    ctx.WaitCtx, cancel = context.WithCancel(context.Background())
    ctx.Wg = new(sync.WaitGroup)
    ctx.StatDur = time.Second * 5

    BrkOpenFilesLimit()
    cnn, err := net.Listen("tcp", ctx.Laddr)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("working on \"%v\"", ctx.Laddr)

    SetupSignal(ctx, cancel)
    // a routine to wake up accept()
    cnnClosedCh := make(chan bool, 1)
    ctx.Wg.Add(1)
    go func() {
        select {
        case <-ctx.WaitCtx.Done():
            _ = cnn.Close()
        case <-cnnClosedCh:
        }
        ctx.Wg.Done()
    }()

    // stat
    ctx.Wg.Add(1)
    go func(ctx *EchoContext) {
        tick := time.NewTicker(ctx.StatDur)
    loop1:
        for {
            select {
            case <-ctx.WaitCtx.Done():
                break loop1
            case <-tick.C:
                log.Printf("stat AddCnn= %v SubCnn= %v Add-Sub= %v",
                    ctx.AddCnn, ctx.SubCnn, ctx.AddCnn-ctx.SubCnn)
            }
        }
        ctx.Wg.Done()
    }(ctx)

loop:
    for {
        cltCnn, err := cnn.Accept()
        if err != nil {
            break loop
        }

        ctx.Wg.Add(1)
        go EchoConn(ctx, cltCnn)

    }
    _ = cnn.Close()
    close(cnnClosedCh)
    ctx.Wg.Wait()
    log.Printf("main exit")
}
