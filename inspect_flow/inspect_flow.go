package main

import (
    "bytes"
    "context"
    "encoding/binary"
    "encoding/json"
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

const (
    DateTimeFmt = "2006/01/02 15:04:05"
)

type FlowCtx struct {
    Raddr   string
    MsgCh   chan []byte
    ErrsCh  chan error
    Exit    bool
    ExitCh  chan struct{}
    SigCh   chan os.Signal
    WaitCtx context.Context
    Wg      *sync.WaitGroup
    MsgEnq  uint32
    MsgDeq  uint32
}

// recv stream bytes from tcp peer
// convert it to msg
// msg format is [uint16 + msgbytes]
func Stream2Msg(ctx *FlowCtx, cnn net.Conn) {
    for {
        msgHdr := make([]byte, 2)
        n, err := io.ReadFull(cnn, msgHdr)
        if err != nil {
            if err != io.EOF {
                select {
                case ctx.ErrsCh <- err:
                default:
                }
            }
            break
        }
        if n != len(msgHdr) {
            break
        }
        var msgLen uint16
        err = binary.Read(bytes.NewReader(msgHdr), binary.BigEndian, &msgLen)
        if err != nil {
            break
        }

        msg := make([]byte, msgLen)
        n, err = io.ReadFull(cnn, msg)
        if err != nil {
            if err != io.EOF {
                select {
                case ctx.ErrsCh <- err:
                default:
                }
            }
            break
        }
        if n != len(msg) {
            break
        }

        select {
        case <-ctx.ExitCh:
        case ctx.MsgCh <- msg:
            atomic.AddUint32(&ctx.MsgEnq, 1)
        }
    }
}
func Dial(ctx *FlowCtx, txM map[string]interface{}) {
    defer ctx.Wg.Done()
    for {
        d := net.Dialer{}
        cnn, err := d.DialContext(ctx.WaitCtx, "tcp", ctx.Raddr)
        if err != nil {
            v := fmt.Errorf("dial %v err= %v", ctx.Raddr, err)
            select {
            case ctx.ErrsCh <- v:
            default:
            }
        }
        // this is what we exit
        if ctx.Exit {
            log.Printf("dial got exit")
            break
        }
        if cnn == nil {
            // wait for a moment to redial
            select {
            case <-time.After(time.Second * 3):
            case <-ctx.ExitCh:
            }
            continue
        }
        //
        log.Printf("dialer got cnn=%v-%v", cnn.LocalAddr(), cnn.RemoteAddr())
        ctx.Wg.Add(1)
        cnnClosedCh := make(chan struct{})
        go func(ctx *FlowCtx, cnn net.Conn, closeCh chan struct{}) {
            // when program exit, we need close cnn
            // when cnn close, we need know
            select {
            case <-ctx.ExitCh:
            case <-closeCh:
            }
            ctx.Wg.Done()
            _ = cnn.Close()
            // log.Printf("dialer sub routine close cnn")
        }(ctx, cnn, cnnClosedCh)

        // when the cnn broken, we need redial
        // if move Stream2Msg to sub routine
        //   we also need to know
        //   whether the cnn is broken or not when reading
        tx, err := json.Marshal(txM)
        _, _ = cnn.Write(tx)
        Stream2Msg(ctx, cnn)
        // notify the sub routine exit
        close(cnnClosedCh)
        if ctx.Exit {
            break
        }
    }

}

func UtcNow() string {
    return time.Now().UTC().Format(DateTimeFmt)
}
func LocalNow() string {
    return time.Now().Format(DateTimeFmt)
}

func BeautyJsonTime(j []byte) []byte {
    m := make(map[string]interface{})
    _ = json.Unmarshal(j, &m)

    hitTime, ok := m["hitTime"].(float64)
    if ok {
        t := time.Unix(int64(hitTime), 0)
        m["hitTimeUtc"] = t.UTC().Format(DateTimeFmt)
        m["hitTimeLocal"] = t.Local().Format(DateTimeFmt)
        r, err := json.Marshal(m)
        if err == nil {
            return r
        }
    }
    // return origin
    return j
}

func main() {

    raddr := flag.String("raddr", "", "sender addr of flow msg")
    sdk_raddr := flag.String("sdk_raddr", "", "filter sdk raddr")
    rs_raddr := flag.String("rs_raddr", "", "filter rs raddr")
    beauty_jsn := flag.Bool("beauty_json", true, "use beauty json for msg")
    flag.Parse()
    if *raddr == "" {
        flag.Usage()
        return
    }
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    var cancel context.CancelFunc
    var err error
    filter := make(map[string]interface{})
    txM := make(map[string]interface{})
    txM["filter"] = filter
    flowCtx := new(FlowCtx)
    flowCtx.ExitCh = make(chan struct{})
    flowCtx.Wg = new(sync.WaitGroup)
    flowCtx.ErrsCh = make(chan error, 10)
    flowCtx.Raddr = *raddr
    flowCtx.MsgCh = make(chan []byte, 1000*1000)
    flowCtx.SigCh = make(chan os.Signal, 1)
    flowCtx.WaitCtx, cancel = context.WithCancel(context.Background())

    if *sdk_raddr != "" {
        sdk := make(map[string]interface{})
        sdk["raddr"] = *sdk_raddr
        filter["sdk"] = sdk
    }
    if *rs_raddr != "" {
        rs := make(map[string]interface{})
        rs["raddr"] = *rs_raddr
        filter["rs"] = rs
    }

    flowCtx.Wg.Add(1)
    signal.Notify(flowCtx.SigCh, os.Interrupt)
    signal.Notify(flowCtx.SigCh, syscall.SIGTERM)
    go func(ctx *FlowCtx) {
        // only need to wait signal
        // we run forever
        <-ctx.SigCh
        ctx.Exit = true
        close(ctx.ExitCh)
        flowCtx.Wg.Done()
    }(flowCtx)

    flowCtx.Wg.Add(1)
    go Dial(flowCtx, txM)

loop:
    for {
        select {
        case err = <-flowCtx.ErrsCh:
            log.Printf("got err =%v", err)
        case msg := <-flowCtx.MsgCh:
            msg = BeautyJsonTime(msg)
            if *beauty_jsn {
                bb := new(bytes.Buffer)
                err = json.Indent(bb, []byte(msg), "", "\t")
                if err == nil {
                    msg = bb.Bytes()
                }
            }

            fmt.Printf("[%v]--utc=%v local=%v %s\n", flowCtx.MsgDeq, UtcNow(), LocalNow(), msg)
            atomic.AddUint32(&flowCtx.MsgDeq, 1)
        case <-flowCtx.ExitCh:
            log.Printf("main thread got exit, break loop")
            cancel()
            break loop
        }
    }

    log.Printf("main wait sub")
    flowCtx.Wg.Wait()
    log.Printf("main exit MsgEnq=%v MsgDeq=%v", flowCtx.MsgEnq, flowCtx.MsgDeq)
}
