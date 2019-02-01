package main

import (
    "context"
    "flag"
    "fmt"
    "github.com/fooofei/go_pieces/echo_server/rlimit"
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

// EchoContext defines the echo server context
type echoContext struct {
    Laddr   string
    WaitCtx context.Context
    Wg      *sync.WaitGroup
    StatDur time.Duration
    AddCnn  uint32
    SubCnn  uint32
}

func setupSignal(ctx *echoContext, cancel context.CancelFunc) {

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

func takeoverCnnClose(ctx *echoContext, cnn io.Closer) chan bool {
    cnnClosedCh := make(chan bool, 1)
    ctx.Wg.Add(1)
    go func() {
        select {
        case <-cnnClosedCh:
        case <-ctx.WaitCtx.Done():
        }
        _ = cnn.Close()
        ctx.Wg.Done()
    }()
    return cnnClosedCh
}

func echoConn(ctx *echoContext, cnn net.Conn) {

    atomic.AddUint32(&ctx.AddCnn, 1)
    cnnClosedCh := takeoverCnnClose(ctx, cnn)

    //copy until EOF
    //_, _ = io.Copy(cnn, cnn)
    //
    //ctx.Wg.Add(1)
    //go func() {
    //   select {
    //   case <-ctx.WaitCtx.Done():
    //   case <-time.After(time.Second * 8):
    //       tcp, _ := cnn.(*net.TCPConn)
    //       _ = tcp.CloseWrite()
    //       log.Printf("close write")
    //   }
    //   ctx.Wg.Done()
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
            log.Printf("Write = [%v][%v]", n, c)

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
            log.Printf("Read = [%v][%s]", n, buf[:n])

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
    ctx.Wg.Done()
    atomic.AddUint32(&ctx.SubCnn, 1)
}

func main() {
    // log
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))

    ctx := new(echoContext)
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

    setupSignal(ctx, cancel)
    // a routine to wake up accept()
    cnnClosedCh := takeoverCnnClose(ctx, cnn)

    // stat
    ctx.Wg.Add(1)
    go func(ctx *echoContext) {
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
        go echoConn(ctx, cltCnn)

    }
    _ = cnn.Close()
    close(cnnClosedCh)
    ctx.Wg.Wait()
    log.Printf("main exit")
}
