package sshttp

import (
	"bufio"
	"context"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// TODO 持续优化

type Tunnel struct {
	SndNxt        int64 // send seq
	SndUna        int64 // send not acknowledged
	RcvNxt        int64
	W             io.Writer
	R             *bufio.Reader
	C             io.Closer
	Ctx           context.Context
	AckedSndNxtCh chan int64
	CopyBuf       []byte // 128*1024 buffer
}

func (t *Tunnel) Write(p []byte) (int, error) {
retransLoop:
	for {
		// SndNxt only changed here, so no need to use atomic
		ack := atomic.LoadInt64(&t.RcvNxt)
		req, err := NewDataRequest(t.SndNxt, ack, p)
		if err != nil {
			return 0, err
		}
		err = req.Write(t.W)
		if err != nil {
			return 0, err
		}
		select {
		case <-t.Ctx.Done():
			return 0, t.Ctx.Err()
		case acked := <-t.AckedSndNxtCh:
			if acked == t.SndNxt+int64(len(p)) {
				atomic.AddInt64(&t.SndNxt, int64(len(p)))
				break retransLoop
			}
		case <-time.After(time.Second * 120):
			// retrans
		}

	}
	return len(p), nil
}

func (t *Tunnel) dataReceivedAck(seq int64, ack int64) {
	req, err := NewDataRequest(seq, ack, nil)
	if err != nil {
		log.Printf("err= %v", err)
		return
	}
	err = req.Write(t.W)
	if err != nil {
		log.Printf("err= %v", err)
	}
}

func (t *Tunnel) WriteTo(w io.Writer) (int64, error) {
	req, er := http.ReadRequest(t.R)
	if er != nil {
		return 0, er
	}
	httpPath, err := ParseUrlPath(req.URL)
	if err != nil {
		return 0, err
	}
	atomic.StoreInt64(&t.SndUna, httpPath.Ack)
	if httpPath.Ack > atomic.LoadInt64(&t.SndNxt) {
		t.AckedSndNxtCh <- httpPath.Ack
	}

	if t.RcvNxt != httpPath.Seq {
		log.Printf("ERROR tunnel t.RcvNxt != httpPath.Seq %v!=%v", t.RcvNxt, httpPath.Seq)
	}
	if httpPath.Type == "data" && req.ContentLength > 0 {
		t.dataReceivedAck(atomic.LoadInt64(&t.SndNxt),
			t.RcvNxt+req.ContentLength)

		nw, ew := io.CopyBuffer(w, req.Body, t.CopyBuf)
		atomic.AddInt64(&t.RcvNxt, req.ContentLength)
		if nw != req.ContentLength {
			log.Printf("ERROR nw != req.ContentLength %v!= %v", nw, req.ContentLength)
		}
		return nw, ew
	}

	return 0, nil
}

func (t *Tunnel) Close() error {
	return t.C.Close()
}
