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

	"github.com/pkg/errors"
)

type flowCtx struct {
	RAddr   string
	WaitCtx context.Context
	Wg      *sync.WaitGroup
	//
	MsgCh  chan []byte
	ErrsCh chan error
	MsgEnq int64
	MsgDeq int64
}

type hello struct {
	SdkRAddr string
	RsRAddr  string
}

func (h *hello) Bytes() []byte {
	filter := make(map[string]interface{})
	txM := make(map[string]interface{})
	txM["filter"] = filter

	if h.SdkRAddr != "" {
		sdk := make(map[string]interface{})
		sdk["raddr"] = h.SdkRAddr
		filter["sdk"] = sdk
	}
	if h.RsRAddr != "" {
		rs := make(map[string]interface{})
		rs["raddr"] = h.RsRAddr
		filter["rs"] = rs
	}
	b, _ := json.Marshal(filter)
	return b
}

func (fc *flowCtx) nonBlockEnqErr(err error) {
	select {
	case fc.ErrsCh <- err:
	default:
	}
}

func takeOverCnnClose(waitCtx context.Context, cnn io.Closer) (chan bool, *sync.WaitGroup) {
	noWait := make(chan bool, 1)
	waitGrp := &sync.WaitGroup{}
	waitGrp.Add(1)
	go func() {
		select {
		case <-noWait:
		case <-waitCtx.Done():
		}
		_ = cnn.Close()
		waitGrp.Done()
	}()
	return noWait, waitGrp
}

// recv stream bytes from tcp peer
// convert it to msg
// msg format is [uint32 + msgbytes]
func checkoutMsg(flowCtx1 *flowCtx, cnn net.Conn) {
	for {
		msgHdr := make([]byte, 4)
		n, err := io.ReadFull(cnn, msgHdr)
		if err != nil {
			if err != io.EOF {
				flowCtx1.nonBlockEnqErr(err)
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
				flowCtx1.nonBlockEnqErr(err)
			}
			break
		}
		if n != len(msg) {
			break
		}
		select {
		case <-flowCtx1.WaitCtx.Done():
		case flowCtx1.MsgCh <- msg:
			atomic.AddInt64(&flowCtx1.MsgEnq, 1)
		}
	}
}
func dial(flowCtx1 *flowCtx, helloBytes []byte) {
dialLoop:
	for {
		d := net.Dialer{}
		cnn, err := d.DialContext(flowCtx1.WaitCtx, "tcp", flowCtx1.RAddr)
		if err != nil {
			flowCtx1.nonBlockEnqErr(errors.Wrapf(err, "dial %v err= %v", flowCtx1.RAddr, err))
		}
		// this is what we exit
		select {
		case <-flowCtx1.WaitCtx.Done():
			break dialLoop
		default:
		}
		if cnn == nil {
			// wait for a moment to dial again
			select {
			case <-time.After(time.Second * 3):
			case <-flowCtx1.WaitCtx.Done():
			}
			continue
		}
		//
		log.Printf("dialer got connection =%v-%v", cnn.LocalAddr(), cnn.RemoteAddr())

		noWait, closeWaitGrp := takeOverCnnClose(flowCtx1.WaitCtx, cnn)

		// when the cnn broken, we need redial
		// if move `checkoutMsg` to sub routine
		//   we also need to know
		//   whether the cnn is broken or not when reading
		_, _ = cnn.Write(helloBytes)
		checkoutMsg(flowCtx1, cnn)
		close(noWait)
		closeWaitGrp.Wait()

		// this is what we exit
		select {
		case <-flowCtx1.WaitCtx.Done():
			break dialLoop
		default:
		}
	}

}

func beautyJson(cnt int64, j []byte) []byte {
	m := make(map[string]interface{})
	_ = json.Unmarshal(j, &m)
	//
	now := time.Now()
	m["0idx"] = cnt
	m["0utcNow"] = now.UTC().Format(time.RFC3339)
	m["0localNow"] = now.Format(time.RFC3339)
	//
	hitTime, ok := m["hitTime"].(float64)
	if ok {
		t := time.Unix(int64(hitTime), 0)
		m["hitTimeUtc"] = t.UTC().Format(time.RFC3339)
		m["hitTimeLocal"] = t.Local().Format(time.RFC3339)
		r, err := json.Marshal(m)
		if err == nil {
			return r
		}
	}
	// return origin
	return j
}

func setupSignal(flowCtx1 *flowCtx, cancel context.CancelFunc) {

	sigCh := make(chan os.Signal, 2)

	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)

	flowCtx1.Wg.Add(1)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-flowCtx1.WaitCtx.Done():
		}
		flowCtx1.Wg.Done()
	}()
}

func main() {

	var cancel context.CancelFunc
	var err error
	flowCtx1 := &flowCtx{}
	hello1 := &hello{}
	flowCtx1.ErrsCh = make(chan error, 10)
	flowCtx1.MsgCh = make(chan []byte, 1000*1000)
	flowCtx1.WaitCtx, cancel = context.WithCancel(context.Background())

	flowCtx1.Wg = &sync.WaitGroup{}
	flag.StringVar(&flowCtx1.RAddr, "raddr", "127.0.0.1:5679", "sender addr of flow msg")
	flag.StringVar(&hello1.SdkRAddr, "sdk_raddr", "", "filter sdk raddr")
	flag.StringVar(&hello1.RsRAddr, "rs_raddr", "", "filter rs raddr")
	fBeautyJsn := flag.Bool("beauty_json", false, "use beauty json for msg")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))

	setupSignal(flowCtx1, cancel)
	flowCtx1.Wg.Add(1)
	go func() {
		dial(flowCtx1, hello1.Bytes())
		flowCtx1.Wg.Done()
	}()

loop:
	for {
		select {
		case err = <-flowCtx1.ErrsCh:
			log.Printf("got err =%v", err)
		case msg := <-flowCtx1.MsgCh:
			msg = beautyJson(flowCtx1.MsgDeq, msg)
			if *fBeautyJsn {
				bb := new(bytes.Buffer)
				err = json.Indent(bb, []byte(msg), "", "\t")
				if err == nil {
					msg = bb.Bytes()
				}
			}
			fmt.Printf("%s\n", msg)
			atomic.AddInt64(&flowCtx1.MsgDeq, 1)
		case <-flowCtx1.WaitCtx.Done():
			log.Printf("main thread got exit, break loop")
			cancel()
			break loop
		}
	}

	log.Printf("main wait sub")
	flowCtx1.Wg.Wait()
	log.Printf("main exit MsgEnq=%v MsgDeq=%v", flowCtx1.MsgEnq, flowCtx1.MsgDeq)
}
