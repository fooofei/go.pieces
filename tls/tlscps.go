package main

import (
    "bytes"
    "crypto/tls"
    "encoding/binary"
    "encoding/json"
    "flag"
    "fmt"
    "golang.org/x/net/context"
    "log"
    "math"
    "net"
    "os"
    "os/signal"
    "sync"
    "sync/atomic"
    "syscall"
    "time"
)

// globals
type MyContext struct {
    StatDur     time.Duration
    RoutinesCnt int
    Raddr       string

    TcpCnt     uint64
    TcpCntOk   uint64
    TcpCntFail uint64
    HitErrs    chan error

    AddRtnCnt int32
    SubRtnCnt int32
    WaitCtx   context.Context
    TlsConf   *tls.Config
    Wg        *sync.WaitGroup
}

func copyContext(src *MyContext, dst *MyContext) {
    dst.TcpCnt = src.TcpCnt
    dst.TcpCntOk = src.TcpCntOk
    dst.TcpCntFail = src.TcpCntFail
    //
    dst.AddRtnCnt = src.AddRtnCnt
    dst.SubRtnCnt = src.SubRtnCnt
}
func subContext(a *MyContext, b *MyContext, c *MyContext) {
    c.TcpCnt = a.TcpCnt - b.TcpCnt
    c.TcpCntOk = a.TcpCntOk - b.TcpCntOk
    c.TcpCntFail = a.TcpCntFail - b.TcpCntFail
    //
    c.AddRtnCnt = a.AddRtnCnt - b.AddRtnCnt
    c.SubRtnCnt = a.SubRtnCnt - b.SubRtnCnt
}

func DeepCopy(src interface{}, dst interface{}) error {
    if dst == nil {
        return fmt.Errorf("dst cannot be nil")
    }
    if src == nil {
        return fmt.Errorf("src cannot be nil")
    }
    bytes_, err := json.Marshal(src)
    if err != nil {
        return fmt.Errorf("Unable to marshal src: %s", err)
    }
    err = json.Unmarshal(bytes_, dst)
    if err != nil {
        return fmt.Errorf("Unable to unmarshal into dst: %s", err)
    }
    return nil
}

func toLEBytes(v interface{}) []byte {
    var binBuf bytes.Buffer
    err := binary.Write(&binBuf, binary.LittleEndian, v)
    if err != nil {
        panic(err)
    }
    return binBuf.Bytes()
}

func toBEBytes(v interface{}) []byte {
    var binBuf bytes.Buffer
    err := binary.Write(&binBuf, binary.BigEndian, v)
    if err != nil {
        panic(err)
    }
    return binBuf.Bytes()
}

func cnnRoutine(ctx *MyContext) {
    defer ctx.Wg.Done()

    tmo := time.Duration(time.Second * 3)
    atomic.AddInt32(&ctx.AddRtnCnt, 1)
    defer atomic.AddInt32(&ctx.SubRtnCnt, 1)
loop:
    for {
        // only tcp connect
        d := &net.Dialer{Timeout: tmo}

        atomic.AddUint64(&ctx.TcpCnt, 1)
        tcpCnn, err := d.DialContext(ctx.WaitCtx, "tcp", ctx.Raddr)
        ok := false
        if err != nil {
            select {
            case ctx.HitErrs <- err:
            default:
            }
        } else {
            cnnClosecCh := make(chan struct{}, 1)
            ctx.Wg.Add(1)
            go func() {
                select {
                case <-ctx.WaitCtx.Done():
                    _ = tcpCnn.Close()
                case <-cnnClosecCh:
                }
                ctx.Wg.Done()
            }()

            tlsCnn := tls.Client(tcpCnn, ctx.TlsConf)
            err = tlsCnn.Handshake()

            if err != nil {
                select {
                case ctx.HitErrs <- err:
                default:
                }
            } else {
                ok = true
            }

            close(cnnClosecCh)
        }

        if !ok {
            atomic.AddUint64(&ctx.TcpCntFail, 1)
        } else {
            atomic.AddUint64(&ctx.TcpCntOk, 1)
        }

        select {
        case <-ctx.WaitCtx.Done():
            break loop
        default:
        }
    }
}

func statRoutine(ctx *MyContext) {

    var hitCtx MyContext
    var nowCtx MyContext
    var itvCtx MyContext
    var hitTime time.Time
    var nowTime time.Time
    var elapse uint64
    var err error

    statTick := time.NewTicker(ctx.StatDur)
    cnt := 0
    hitTime = time.Now()
loop:
    for {
        select {
        case <-ctx.WaitCtx.Done():
            break loop
        case <-statTick.C:
            nowTime = time.Now()
            var buf bytes.Buffer
            copyContext(ctx, &nowCtx)
            // sub value
            log.Printf("hit stat cnt= %v raddr= %v", cnt, ctx.Raddr)
            subContext(&nowCtx, &hitCtx, &itvCtx)
            elapse = uint64(math.Max(float64(1), float64(nowTime.Sub(hitTime).Seconds())))
            // calc value
            buf.WriteString(fmt.Sprintf("  tcpCnt %v-%v/%v=%.3f\n",
                nowCtx.TcpCnt, hitCtx.TcpCnt, elapse, float64(nowCtx.TcpCnt-hitCtx.TcpCnt)/float64(elapse)))
            buf.WriteString(fmt.Sprintf("  tcpCntOk %v-%v/%v=%.3f\n",
                nowCtx.TcpCntOk, hitCtx.TcpCntOk, elapse, float64(nowCtx.TcpCntOk-hitCtx.TcpCntOk)/float64(elapse)))
            buf.WriteString(fmt.Sprintf("  tcpCntFail %v-%v/%v=%.3f\n",
                nowCtx.TcpCntFail, hitCtx.TcpCntFail, elapse, float64(nowCtx.TcpCntFail-hitCtx.TcpCntFail)/float64(elapse)))
            fmt.Printf("%v", buf.String())
            //
            copyContext(&nowCtx, &hitCtx)
            hitTime = nowTime
            cnt += 1

            select {
            case err = <-ctx.HitErrs:
                log.Printf("hit err= %v", err)
            default:
            }
        }

    }

}

func SetupSignal(ctx *MyContext, cancel context.CancelFunc) {

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

func main() {

    //
    ctx := new(MyContext)
    flag.IntVar(&ctx.RoutinesCnt, "routines", 1, "go routines cnt")
    flag.StringVar(&ctx.Raddr, "raddr", "127.0.0.1:886", "tcp-ssl raddr")
    dur := flag.Int("interval", 3, "stat interval")
    flag.Parse()
    ctx.StatDur = time.Second * time.Duration(*dur)
    //
    log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    //
    log.Printf("use routines=%v to raddr= %v\n", ctx.RoutinesCnt, ctx.Raddr)
    //
    var cancel context.CancelFunc
    ctx.HitErrs = make(chan error, 3)
    ctx.TlsConf = new(tls.Config)
    ctx.TlsConf.InsecureSkipVerify = true
    ctx.WaitCtx, cancel = context.WithCancel(context.Background())
    ctx.Wg = new(sync.WaitGroup)

    SetupSignal(ctx, cancel)

    for i := 0; i < ctx.RoutinesCnt; i += 1 {
        ctx.Wg.Add(1)
        go cnnRoutine(ctx)
    }

    log.Printf("all routines started, go stat")

    statRoutine(ctx)
    // wait close
    log.Printf("wait all routines to exit")
    ctx.Wg.Wait()
    log.Printf("exit\n")
}
