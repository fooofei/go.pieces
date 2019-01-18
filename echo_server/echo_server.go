package main

import (
    "flag"
    "golang.org/x/net/context"
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
    defer atomic.AddUint32(&ctx.SubCnn, 1)

    defer ctx.Wg.Done()
    defer func() {
        _ = cnn.Close()
    }()

    cnnClosedCh := make(chan struct{}, 1)
    ctx.Wg.Add(1)
    go func() {
        select {
        case <-cnnClosedCh:
        case <-ctx.WaitCtx.Done():
            _ = cnn.Close()
        }
        ctx.Wg.Done()
    }()

    buf := make([]byte, 128*1024)
loop:
    for {
        n, err := cnn.Read(buf)
        if err != nil {
            break loop
        }

        // log.Printf("rx %s", buf[:n])
        _, err = cnn.Write(buf[:n])
        if err != nil {
            break loop
        }
    }

    close(cnnClosedCh)
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

    log.SetFlags(log.LstdFlags | log.Lshortfile)

    laddr := flag.String("laddr", "", "The local listen addr")
    flag.Parse()
    if *laddr == "" {
        flag.PrintDefaults()
        return
    }

    var cancel context.CancelFunc
    var err error
    pid := os.Getpid()
    ctx := new(EchoContext)
    ctx.Laddr = *laddr
    ctx.WaitCtx, cancel = context.WithCancel(context.Background())
    ctx.Wg = new(sync.WaitGroup)

    BrkOpenFilesLimit()
    cnn, err := net.Listen("tcp", ctx.Laddr)
    if err != nil {
        log.Fatal(err)
    }
    defer func() {
        _ = cnn.Close()
    }()

    log.Printf("working on \"%v\"", ctx.Laddr)

    SetupSignal(ctx, cancel)
    // a routine to wake up accept()
    cnnClosedCh := make(chan struct{}, 1)
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
    go func() {
        defer ctx.Wg.Done()
        tick := time.NewTicker(time.Second * 5)
        for {
            select {
            case <-ctx.WaitCtx.Done():
                return
            case <-tick.C:
                log.Printf("pid= %v stat AddCnn= %v SubCnn= %v Add-Sub= %v",
                    pid, ctx.AddCnn, ctx.SubCnn, ctx.AddCnn-ctx.SubCnn)
            }
        }
    }()

loop:
    for {
        cltCnn, err := cnn.Accept()
        if err != nil {
            break loop
        }

        ctx.Wg.Add(1)
        go EchoConn(ctx, cltCnn)

    }

    close(cnnClosedCh)
    ctx.Wg.Wait()
    log.Printf("main exit")
}
