package main

import (
    "io"
    "log"
    "net"
    "os"
    "os/signal"
    "sync"
    "sync/atomic"
    "time"
)

/**
 1、学习放置 wg.Add(1) 的正确位置
    放到routine里 会遇到主routine wait 直接失败，因为两个routine 的执行时机不等。
    必须确保 Add 之后才wait
 2、学习 golang 中正确使用 cnn Read Write的方式
    带超时的使用方式是不推荐的，那是一种bad 的使用方式，golang 中 cnn 支持多routine，
    因此为了做安全退出，不阻塞在 Read Write，可以在其他routine 中close cnn，
    Read Write 就会返回，这是非常优雅的。

https://gist.github.com/zupzup/14ea270252fbb24332c5e9ba978a8ade
学习的第一个参考来自这里，但是存在诸多缺点。
*/

type LsnCtx struct {
    lsnAddr      string
    addTunnelCnt uint32
    subTunnelCnt uint32
    grpAddr      string
}

type Context struct {
    raddr        string
    addTunnelCnt uint32
    subTunnelCnt uint32
    hitErrs      chan error
    stopCh       chan struct {}
    stop         bool
    wg           sync.WaitGroup
    lsns         []LsnCtx
}

func ctxEnqErr(ctx *Context, err error) {
    select {
    case ctx.hitErrs <- err:
    default:
    }
}

func listenRoutine(ctx *Context, lsnCtx *LsnCtx) {
    defer ctx.wg.Done()
    defer log.Printf("listener[%v] exit", lsnCtx.lsnAddr)

    laddr, err := net.ResolveTCPAddr("tcp", lsnCtx.lsnAddr)
    if err != nil {
        panic(err)
    }
    tcpLsn, err := net.ListenTCP("tcp", laddr)
    if err != nil {
        panic(err)
    }
    defer func() {
        _ = tcpLsn.Close()
    }()

    log.Printf("listener[%v] working", tcpLsn.Addr())

    lsnCloseCh := make(chan struct{},1)

    ctx.wg.Add(1)
    go func() {
        defer ctx.wg.Done()
        defer close(lsnCloseCh)
        //
        for {
            appConn, err := tcpLsn.Accept()
            if err != nil {
                ctxEnqErr(ctx, err)
            }
            if ctx.stop{
                return
            }
            log.Printf("accept %v-%v",appConn.LocalAddr(), appConn.RemoteAddr())
            ctx.wg.Add(1)
            go tunnelRoutine(ctx, lsnCtx, appConn)
        }
    }()

    select {
    case <- ctx.stopCh:
        _ = tcpLsn.Close()
    case <- lsnCloseCh:
    }
}
/*
func tmoRead(cnn net.Conn, b []byte) (int, error) {
    dta := time.Duration(1) * time.Second
    _ = cnn.SetWriteDeadline(time.Now().Add(dta))
    n, err := cnn.Read(b)
    _ = cnn.SetReadDeadline(time.Time{})
    return n, err
}

func tmoWrite(cnn net.Conn, b []byte) (int, error) {
    dta := time.Duration(1) * time.Second
    _ = cnn.SetWriteDeadline(time.Now().Add(dta))
    n, err := cnn.Write(b)
    _ = cnn.SetReadDeadline(time.Time{})
    return n, err
}
*/

func tunnelRoutine(ctx *Context, lsnCtx *LsnCtx, appConn net.Conn) {
    /**
    试错。read 带 deadline 超时，这个设计很差，太复杂。
    如果数据流是以msg为单位，那么会遇到不到一个msg的时候，read超时返回，
    又需要处理读取偏移。
    read 带超时是为了应对无法退出的情况，golang中，可以在其他goroutine 中
    close 这个 conn达到通知的效果，这样简化设计
    */
    defer ctx.wg.Done()

    atomic.AddUint32(&ctx.addTunnelCnt, 1)
    defer atomic.AddUint32(&ctx.subTunnelCnt, 1)

    atomic.AddUint32(&lsnCtx.addTunnelCnt, 1)
    defer atomic.AddUint32(&lsnCtx.subTunnelCnt, 1)

    d := net.Dialer{Timeout: time.Duration(3) * time.Second}

    defer func() {
        err := appConn.Close()
        if err != nil {
            ctxEnqErr(ctx, err)
        }
    }()

    // TODO use tls
    tunConn, err := d.Dial("tcp", ctx.raddr)
    if err != nil {
        ctxEnqErr(ctx, err)
    }
    if tunConn == nil {
        return
    }
    defer func() {
        err := tunConn.Close()
        if err != nil {
            ctxEnqErr(ctx, err)
        }
    }()


    tunWg := sync.WaitGroup{}


    closeAllCnn := func() {
        _ = tunConn.Close()
        _ = appConn.Close()
    }
    _ = closeAllCnn
    // app->tunnel
    // 如果Add(1) 放在 routine里，就有可能遇到 wait 时还没+1的情况就会退出
    ctx.wg.Add(1)
    tunWg.Add(1)
    go func() {
        defer ctx.wg.Done()
        defer tunWg.Done()
        //
        ibuf := make([]byte, 128*1024)
        for {
            // only io.EOF stop read
            n,err := appConn.Read(ibuf)
            if ctx.stop {
                // stop program
                return
            }

            if n > 0 {
                // TODO add custom protocol
                for idx := 0; idx < n; {
                    // only Timeout continue write
                    writed, err := tunConn.Write(ibuf[idx:n])
                    if ctx.stop {
                        // stop program
                        return
                    }

                    if err != nil {
                        ctxEnqErr(ctx, err)
                        tcpCnn, ok := appConn.(*net.TCPConn)
                        if ok {
                            _ = tcpCnn.CloseRead()
                        }
                        //
                        log.Printf("app->tunnel write err=%v", err)
                        //
                        return
                    }
                    idx += writed
                }
            } else if err == io.EOF {
                tcpCnn, ok := tunConn.(*net.TCPConn)
                if ok {
                    _ = tcpCnn.CloseWrite()
                }
                //
                log.Printf("app->tunnel read err=%v", err)
                //
                return
            } else if err != nil{
                ctxEnqErr(ctx, err)
            }

        }

    }()
    // tunnel ->app
    ctx.wg.Add(1)
    tunWg.Add(1)
    go func() {
        defer ctx.wg.Done()
        defer tunWg.Done()
        //
        buf := make([]byte, 128*1024)
        for{
            n,err := tunConn.Read(buf)
            if ctx.stop{
                // stop program
                return
            }
            if n>0 {
                for idx:=0; idx<n; {
                    writed,err := appConn.Write(buf[idx:n])
                    if ctx.stop{
                        return
                    }
                    if err != nil{
                        ctxEnqErr(ctx, err)
                        tcpCnn, ok := tunConn.(*net.TCPConn)
                        if ok {
                            _ = tcpCnn.CloseRead()
                        }
                        //
                        log.Printf("tunnel->app write err=%v", err)
                        //
                        return
                    }
                    idx += writed
                }

            } else if err == io.EOF {
                tcpCnn,ok := appConn.(*net.TCPConn)
                if ok {
                    _ = tcpCnn.CloseWrite()
                }
                //
                log.Printf("tunnel->app read err=%v", err)
                //
                return
            } else if err != nil{
                ctxEnqErr(ctx, err)
            }


        }
    }()

    tunCh := make(chan struct{},1)
    go func() {
        tunWg.Wait()
        close(tunCh)
    }()


    select {
    case <-ctx.stopCh:
        _ = tunConn.Close()
        _ = appConn.Close()
        log.Printf("tunnel routine rcv stopCh")
    case <- tunCh:
        log.Printf("tunnel routine rcv tunCh")
    }

    log.Printf("tunnel routine exit")
}

func main() {

    ctx := new(Context)
    ctx.raddr = "127.0.0.1:8869"
    ctx.hitErrs = make(chan error, 4)
    ctx.stopCh = make(chan  struct{},1)

    ctx.lsns = make([]LsnCtx, 1)

    ctx.lsns[0].lsnAddr = "127.0.0.1:8879"

    // setup signal
    sigCh := make(chan os.Signal,1)
    signal.Notify(sigCh, os.Interrupt)
    go func() {
        //
        ctx.wg.Add(1)
        defer ctx.wg.Done()
        //
        <- sigCh
        log.Printf("rcv signal, goto close")
        ctx.stop=true
        close(ctx.stopCh)
    }()

    for _,lsn := range ctx.lsns{
        ctx.wg.Add(1)
        go listenRoutine(ctx,&lsn)
    }

    go func() {
        ctx.wg.Add(1)
        defer ctx.wg.Done()

        for{
            select {
            case <- ctx.stopCh:
                return
            case err:= <- ctx.hitErrs:
                log.Printf("stat err=%v",err)
            case <- time.After(time.Second*3):
                log.Printf("tun cnt= add%v sub%v", ctx.addTunnelCnt, ctx.subTunnelCnt)
            }
        }
    }()

    log.Printf("main waiting")
    ctx.wg.Wait()

    log.Printf("exit\n")
}
