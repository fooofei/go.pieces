package main

import (
    "bytes"
    "encoding/binary"
    "errors"
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

type FlowCtx struct {
    Raddr  string
    MsgCh  chan string
    ErrsCh chan error
    Exit   bool
    ExitCh chan struct{}
    SigCh  chan os.Signal
    Wg     *sync.WaitGroup
    MsgEnq uint32
    MsgDeq uint32
}

func Stream2Msg(ctx *FlowCtx) {
    defer ctx.Wg.Done()
    for {
        d := net.Dialer{Timeout: time.Duration(2) * time.Second}
        cnn, err := d.Dial("tcp", ctx.Raddr)
        if err != nil {
            v := errors.New(fmt.Sprintf("dial %v err= %v", ctx.Raddr, err))
            select {
            case ctx.ErrsCh <- v:
            default:
            }
        }
        if ctx.Exit {
            log.Printf("dial got exit")
            break
        }
        if cnn == nil {
            select {
            case <-time.After(time.Second * 3):
            case <-ctx.ExitCh:
            }
            continue
        }
        //
        // log.Printf("dialer got cnn=%v", cnn)
        ctx.Wg.Add(1)
        closeCh := make(chan struct{})
        go func(ctx *FlowCtx, cnn net.Conn, closeCh chan struct{}) {
            select {
            case <-ctx.ExitCh:
            case <-closeCh:
            }
            ctx.Wg.Done()
            _ = cnn.Close()
            // log.Printf("dialer sub routine close cnn")
        }(ctx, cnn, closeCh)

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
            case ctx.MsgCh <- string(msg):
                atomic.AddUint32(&ctx.MsgEnq, 1)
            }
        }

        // notify the sub routine exit
        close(closeCh)
        if ctx.Exit{
            break
        }
    }

}

func main() {

    raddr := flag.String("raddr", "","receiver of flow msg")
    flag.Parse()
    if *raddr == ""{
        flag.Usage()
        return
    }
    flowCtx := new(FlowCtx)
    flowCtx.ExitCh = make(chan struct{})
    flowCtx.Wg = new(sync.WaitGroup)
    flowCtx.ErrsCh = make(chan error, 10)
    flowCtx.Raddr = *raddr
    flowCtx.MsgCh = make(chan string, 1000*1000)
    flowCtx.SigCh = make(chan os.Signal, 1)

    flowCtx.Wg.Add(1)
    signal.Notify(flowCtx.SigCh, os.Interrupt)
    signal.Notify(flowCtx.SigCh, syscall.SIGTERM)
    go func(ctx *FlowCtx) {
        <-ctx.SigCh
        ctx.Exit = true
        close(ctx.ExitCh)
        flowCtx.Wg.Done()
    }(flowCtx)

    flowCtx.Wg.Add(1)
    go Stream2Msg(flowCtx)

loop:
    for {
        select {
        case err := <-flowCtx.ErrsCh:
            log.Printf("got err =%v", err)
        case msg := <-flowCtx.MsgCh:
            fmt.Printf("[%v]--[%v]\n",flowCtx.MsgDeq, msg)
            atomic.AddUint32(&flowCtx.MsgDeq, 1)
        case <-flowCtx.ExitCh:
            log.Printf("main thread got exit, break loop")
            break loop
        }
    }

    log.Printf("main wait sub")
    flowCtx.Wg.Wait()
    log.Printf("main exit MsgEnq=%v MsgDeq=%v", flowCtx.MsgEnq, flowCtx.MsgDeq)
}
