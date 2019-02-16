package main

import (
    "bytes"
    "context"
    "encoding/binary"
    "encoding/hex"
    "flag"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"
)

type txContext struct {
    Raddr   string
    WaitCtx context.Context
    Wg      *sync.WaitGroup
}

func setupSignal(ctx *txContext, cancel context.CancelFunc) {

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

func takeoverCnnClose(ctx *txContext, cnn io.Closer) chan bool {

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

func rx(ctx *txContext, cnn io.ReadCloser) {

    var err error
    bs := make([]byte, 128*1024)
    var n int
rloop:
    for {
        n, err = cnn.Read(bs)
        if err != nil {
            break rloop
        }

        log.Printf("Read [%v][%s]", n, bs[:n])
    }
}

func tx(ctx *txContext, cnn io.WriteCloser) {
    var cnt int
    var n int
    var err error
wloop:

    for {
        select {
        case <-time.After(time.Second * 2):
        case <-ctx.WaitCtx.Done():
            break wloop
        }

        txContent := fmt.Sprintf("from txer %v", cnt)

        n, err = cnn.Write([]byte(txContent))
        if err != nil {
            break wloop
        }
        log.Printf("Write [%v][%v]", n, txContent)
        cnt ++
    }
}

func setToaOpt(cnn net.Conn){
    const TCP_TOA int=512
    tcpcnn,_ := cnn.(*net.TCPConn)

    file,_ := tcpcnn.File()


    addr := new(syscall.RawSockaddrInet4)
    bport := make([]byte,2)
    binary.LittleEndian.PutUint16(bport,22)
    addr.Port = binary.BigEndian.Uint16(bport)
    addr.Family = syscall.AF_INET

    _ = copy(addr.Addr[:], net.ParseIP("100.101.102.103").To4())
    // convert bytes to string
    b := new(bytes.Buffer)
    _ = binary.Write(b, binary.BigEndian, addr)
    log.Printf("setsockopt TCP_TOA= %v", hex.EncodeToString(b.Bytes()))
    err := syscall.SetsockoptString(int(file.Fd()), syscall.IPPROTO_IP, TCP_TOA,b.String())
    if err != nil {
        log.Printf("setsockopt TCP_TOA err= %v", err)
    }
    // the File() will set fd to block mode, we revert it
    // cannot write after tcpcnn.File(), will not work
    _ = syscall.SetNonblock(int(file.Fd()),true)

}

func main() {

    ctx := new(txContext)
    var cancel context.CancelFunc
    var err error
    var cnn net.Conn
    // log init
    log.SetFlags(log.Lshortfile | log.LstdFlags)
    log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
    //
    flag.StringVar(&ctx.Raddr, "raddr", "", "tcp connect raddr")
    flag.Parse()
    //

    if ctx.Raddr == "" {
        flag.PrintDefaults()
        return
    }
    ctx.Wg = new(sync.WaitGroup)
    ctx.WaitCtx, cancel = context.WithCancel(context.Background())
    dia := &net.Dialer{}
    cnn, err = dia.DialContext(ctx.WaitCtx, "tcp", ctx.Raddr)
    // still don't have a way to setsockopt before connect
    if err != nil {
        log.Fatal(err)
    }
    setupSignal(ctx, cancel)
    cnnCloseCh := takeoverCnnClose(ctx, cnn)
    defer close(cnnCloseCh)

    subWg := new(sync.WaitGroup)
    ctx.Wg.Add(1)
    subWg.Add(1)
    go func() {
        rx(ctx, cnn)
        ctx.Wg.Done()
        subWg.Done()
    }()

    ctx.Wg.Add(1)
    subWg.Add(1)
    go func() {
        tx(ctx, cnn)
        ctx.Wg.Done()
        subWg.Done()
    }()

    subWg.Wait()
    ctx.Wg.Wait()
    log.Printf("main exit")
}
