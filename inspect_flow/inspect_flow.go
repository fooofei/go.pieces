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
        msgHdr := make([]byte, 4)
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
        var msgLen uint32
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
        case <-ctx.WaitCtx.Done():
        case ctx.MsgCh <- msg:
            atomic.AddUint32(&ctx.MsgEnq, 1)
        }
    }
}
func Dial(ctx *FlowCtx, txM map[string]interface{}) {
    defer ctx.Wg.Done()
loop:
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
        select {
        case <-ctx.WaitCtx.Done():
            break loop
        default:
        }
        if cnn == nil {
            // wait for a moment to redial
            select {
            case <-time.After(time.Second * 3):
            case <-ctx.WaitCtx.Done():
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
            case <-ctx.WaitCtx.Done():
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

        // this is what we exit
        select {
        case <-ctx.WaitCtx.Done():
            break loop
        default:
        }
    }

}

func utcNow() string {
    return time.Now().UTC().Format(DateTimeFmt)
}
func localNow() string {
    return time.Now().Format(DateTimeFmt)
}

func beautyJson(cnt uint32, j []byte) []byte {
    m := make(map[string]interface{})
    _ = json.Unmarshal(j, &m)
    //
    m["0idx"]=cnt
    m["0utcNow"] = utcNow()
    m["0localNow"] = localNow()
    //
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

func setupSignal(ctx *FlowCtx, cancel context.CancelFunc) {

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

    raddr := flag.String("raddr", "", "sender addr of flow msg")
    sdkRaddr := flag.String("sdk_raddr", "", "filter sdk raddr")
    rsRaddr := flag.String("rs_raddr", "", "filter rs raddr")
    fBeautyJsn := flag.Bool("beauty_json", false, "use beauty json for msg")
    flag.Parse()
    if *raddr == "" {
        flag.Usage()
        return
    }
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
    var cancel context.CancelFunc
    var err error
    filter := make(map[string]interface{})
    txM := make(map[string]interface{})
    txM["filter"] = filter
    flowCtx := new(FlowCtx)
    flowCtx.Wg = new(sync.WaitGroup)
    flowCtx.ErrsCh = make(chan error, 10)
    flowCtx.Raddr = *raddr
    flowCtx.MsgCh = make(chan []byte, 1000*1000)
    flowCtx.WaitCtx, cancel = context.WithCancel(context.Background())

    if *sdkRaddr != "" {
        sdk := make(map[string]interface{})
        sdk["raddr"] = *sdkRaddr
        filter["sdk"] = sdk
    }
    if *rsRaddr != "" {
        rs := make(map[string]interface{})
        rs["raddr"] = *rsRaddr
        filter["rs"] = rs
    }

    setupSignal(flowCtx, cancel)
    flowCtx.Wg.Add(1)
    go Dial(flowCtx, txM)

loop:
    for {
        select {
        case err = <-flowCtx.ErrsCh:
            log.Printf("got err =%v", err)
        case msg := <-flowCtx.MsgCh:
            msg = beautyJson(flowCtx.MsgDeq, msg)
            if *fBeautyJsn {
                bb := new(bytes.Buffer)
                err = json.Indent(bb, []byte(msg), "", "\t")
                if err == nil {
                    msg = bb.Bytes()
                }
            }
            fmt.Printf("%s\n",msg)
            atomic.AddUint32(&flowCtx.MsgDeq, 1)
        case <-flowCtx.WaitCtx.Done():
            log.Printf("main thread got exit, break loop")
            cancel()
            break loop
        }
    }

    log.Printf("main wait sub")
    flowCtx.Wg.Wait()
    log.Printf("main exit MsgEnq=%v MsgDeq=%v", flowCtx.MsgEnq, flowCtx.MsgDeq)
}
