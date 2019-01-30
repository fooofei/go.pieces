package main

import (
    "context"
    "flag"
    "fmt"
    "github.com/fooofei/go_pieces/echo_server/rlimit"
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

    //copy until EOF
    //_, _ = io.Copy(cnn, cnn)
    //
    //ctx.Wg.Add(1)
    //go func() {
    //  select {
    //  case <-ctx.WaitCtx.Done():
    //  case <-time.After(time.Second * 8):
    //      tcp, _ := cnn.(*net.TCPConn)
    //      _ = tcp.CloseWrite()
    //      log.Printf("close write")
    //  }
    //  ctx.Wg.Done()
    //}()

    // read and write must move to routine
    // each other not influence other
    subWg := new(sync.WaitGroup)
    ctx.Wg.Add(1)
    subWg.Add(1)
    go func() {
        // write
        cnt := 0
    wloop:
        for {
            c := fmt.Sprintf("from server %v", cnt)
            n, err := cnn.Write([]byte(c))
            cnt ++
            if err != nil {
                log.Printf("Write err= %v break", err)
                break wloop
            }
            log.Printf("Write = %v", n)

            select {
            case <-time.After(time.Second):
            case <-ctx.WaitCtx.Done():
            }
        }

        ctx.Wg.Done()
        subWg.Done()
    }()

    ctx.Wg.Add(1)
    subWg.Add(1)
    go func() {

        buf := make([]byte, 128*1024)
    rloop:
        for {

            n, err := cnn.Read(buf)
            if err != nil {
                log.Printf("Read err= %v break", err)
                break rloop
            }
            log.Printf("Read = %s", buf[:n])

            select {
            case <-time.After(time.Second):
            case <-ctx.WaitCtx.Done():
            }

        }

        ctx.Wg.Done()
        subWg.Done()
    }()
    subWg.Wait()
    close(cnnClosedCh)
    _ = cnn.Close()
    ctx.Wg.Done()
    atomic.AddUint32(&ctx.SubCnn, 1)
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

    rlimt.BrkOpenFilesLimit()
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
