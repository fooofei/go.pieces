package main

import (
    "context"
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
    LsnAddr      string
    AddTunnelCnt uint32
    SubTunnelCnt uint32
    GrpAddr      string
}

type GlbContext struct {
    Raddr        string
    AddTunnelCnt uint32
    SubTunnelCnt uint32
    HitErrs      chan error
    ExitCh       chan struct{}
    Exit         bool
    Wg           sync.WaitGroup
    WaitCtx      context.Context
    Lsns         []LsnCtx
}

func listenRoutine(ctx *GlbContext, lsnCtx *LsnCtx) {
    defer ctx.Wg.Done()
    defer log.Printf("listener[%v] exit", lsnCtx.LsnAddr)

    tcpLsn, err := net.Listen("tcp", lsnCtx.LsnAddr)
    if err != nil {
        panic(err)
    }
    log.Printf("listener[%v] working", tcpLsn.Addr())
    ctx.Wg.Add(1)
    go func(ctx *GlbContext, lsnCtx *LsnCtx, tcpLsn net.Listener) {
        defer ctx.Wg.Done()
        //
        for {
            appCnn, err := tcpLsn.Accept()
            if err != nil {
                select {
                case ctx.HitErrs <- err:
                default:
                }
            }
            if ctx.Exit {
                return
            }
            log.Printf("accept %v-%v", appCnn.LocalAddr(), appCnn.RemoteAddr())
            ctx.Wg.Add(1)
            go tunnelRoutine(ctx, lsnCtx, appCnn)
        }
    }(ctx, lsnCtx, tcpLsn)

    // only program exit will stop listen
    select {
    case <-ctx.ExitCh:
        // wakeup the block of accept()
        _ = tcpLsn.Close()
    }
}

// TCP 是双工，单方向通道坏掉，另一个方向的还可能会好
func app2tunn(appCnn io.ReadCloser, tunCnn io.WriteCloser, ctx *GlbContext, tunWg *sync.WaitGroup) {
    defer ctx.Wg.Done()
    defer tunWg.Done()
    //
    ibuf := make([]byte, 128*1024)
    for {
        // only io.EOF stop read
        n, err := appCnn.Read(ibuf)
        if err != nil {
            tcpCnn, ok := tunCnn.(*net.TCPConn)
            if ok {
                _ = tcpCnn.CloseWrite()
            } else {
                // force close
                // _ = tunCnn.Close()
            }
            return
        }
        if ctx.Exit {
            // stop program
            return
        }
        if n <= 0 {
            continue
        }
        // TODO add custom protocol
        n, err = tunCnn.Write(ibuf[:n])
        if err != nil {
            tcpCnn, ok := appCnn.(*net.TCPConn)
            if ok {
                _ = tcpCnn.CloseRead()
            } else {
                // force
                //_  = appCnn.Close()
            }
            return
        }
    }
}

func tunnelRoutine(ctx *GlbContext, lsnCtx *LsnCtx, appConn net.Conn) {
    // 试错。read 带 deadline 超时，这个设计很差，太复杂。
    // 如果数据流是以msg为单位，那么会遇到不到一个msg的时候，read超时返回，
    // 又需要处理读取偏移。
    // read 带超时是为了应对无法退出的情况，golang中，可以在其他goroutine 中
    // close 这个 conn达到通知的效果，这样简化设计
    defer ctx.Wg.Done()

    atomic.AddUint32(&ctx.AddTunnelCnt, 1)
    defer atomic.AddUint32(&ctx.SubTunnelCnt, 1)

    atomic.AddUint32(&lsnCtx.AddTunnelCnt, 1)
    defer atomic.AddUint32(&lsnCtx.SubTunnelCnt, 1)

    defer func() {
        _ = appConn.Close()
    }()

    // TODO use tls
    d := new(net.Dialer)
    tunConn, err := d.DialContext(ctx.WaitCtx, "tcp", ctx.Raddr)
    if err != nil {
        select {
        case ctx.HitErrs <- err:
        default:
        }
        return
    }
    if ctx.Exit {
        return
    }
    defer func() {
        _ = tunConn.Close()
    }()

    tunCloseCh := make(chan struct{})
    tunWg := new(sync.WaitGroup)
    // app->tunnel
    // 如果Add(1) 放在 routine里，就有可能遇到 wait 时还没+1的情况就会退出
    ctx.Wg.Add(1)
    tunWg.Add(1)
    go app2tunn(appConn, tunConn, ctx, tunWg)
    tunn2app := app2tunn
    ctx.Wg.Add(1)
    tunWg.Add(1)
    go tunn2app(tunConn, appConn, ctx, tunWg)

    ctx.Wg.Add(1)
    go func() {
        tunWg.Wait()
        close(tunCloseCh)
        ctx.Wg.Done()
    }()

    select {
    case <-ctx.ExitCh:
        _ = tunConn.Close()
        _ = appConn.Close()
        log.Printf("tunnel routine rcv stopCh")
    case <-tunCloseCh:
        // app2tun tun2app all closed
        log.Printf("tunnel routine rcv tunCh")
    }

    log.Printf("tunnel routine exit")
}

func main() {
    var cancel context.CancelFunc
    ctx := new(GlbContext)
    ctx.Raddr = "127.0.0.1:8869"
    // can be more, custom
    ctx.HitErrs = make(chan error, 4)
    ctx.ExitCh = make(chan struct{})
    ctx.Lsns = make([]LsnCtx, 1)
    ctx.WaitCtx, cancel = context.WithCancel(context.Background())

    ctx.Lsns[0].LsnAddr = "127.0.0.1:8879"

    // setup signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt)
    signal.Notify(sigCh, syscall.SIGTERM)
    ctx.Wg.Add(1)
    go func() {
        v := <-sigCh
        log.Printf("rcv signal %v, goto close", v)
        ctx.Exit = true
        close(ctx.ExitCh)
        ctx.Wg.Done()
    }()

    for i := 0; i < len(ctx.Lsns); i += 1 {
        ctx.Wg.Add(1)
        go listenRoutine(ctx, &ctx.Lsns[i])
    }

    // stat
loop:
    for {
        select {
        case <-ctx.ExitCh:
            cancel()
            break loop
        case err := <-ctx.HitErrs:
            log.Printf("err= %v", err)
        case <-time.After(time.Second * 3):
            log.Printf("tun cnt (add= %v sub= %v)", ctx.AddTunnelCnt, ctx.SubTunnelCnt)
        }
    }
    log.Printf("main wait sub")
    ctx.Wg.Wait()
    log.Printf("main exit\n")
}
