package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

// 测试一个 HTTP 服务的 Performance

type rateCounter struct {
	Cnt     int64
	HitCnt  int64
	HitTime time.Time
	Trends  []int64
}

func newRateCounter() *rateCounter {
	return &rateCounter{HitTime: time.Now()}
}

func (rc *rateCounter) shot() {
	v := rc.Cnt - rc.HitCnt
	e := int64(time.Since(rc.HitTime).Seconds())
	if e > 0 {
		rc.Trends = append(rc.Trends, v/e)
	}
	idx := 0
	if len(rc.Trends) > 3 {
		idx = len(rc.Trends) - 3
	}
	rc.Trends = rc.Trends[idx:len(rc.Trends)]
	rc.HitTime = time.Now()
	rc.HitCnt = rc.Cnt
}

func (rc *rateCounter) rate() string {
	rc.shot()
	sb := strings.Builder{}
	for _, v := range rc.Trends {
		sb.WriteString(fmt.Sprintf("%v (s) ", v))
	}
	return strings.TrimRight(sb.String(), " ")
}
func (rc *rateCounter) inc() {
	atomic.AddInt64(&rc.Cnt, 1)
}

type boomContext struct {
	// for start
	Wg                sync.WaitGroup
	WaitCtx           context.Context
	BoomPld           []byte
	GoCntLimit        int64
	RAddr             string
	Headers           map[string]string
	GoroutineInterval time.Duration
	SomeOneExitCh     chan bool

	// for runtime state
	ReqCntOk       *rateCounter
	MaxTakeSeconds int64
	ErrorCh        chan error
	GoCnt          int64
}

// enqueue success or ignored
func (bc *boomContext) nonBlockEnqErr(err error) {
	select {
	case bc.ErrorCh <- err:
	default:
	}
}

func boom(ctx *boomContext) {
	atomic.AddInt64(&ctx.GoCnt, 1)

	// clt no need close
	clt := new(http.Client)
	for {

		req, err := http.NewRequest("POST", ctx.RAddr, bytes.NewReader(ctx.BoomPld))
		if err != nil {
			ctx.nonBlockEnqErr(errors.Wrapf(err, "make Request err"))
			break
		}
		for k, v := range ctx.Headers {
			req.Header.Set(k, v)
		}
	doLoop:
		for {
			select {
			case <-ctx.WaitCtx.Done():
				break doLoop
			default:
			}

			req = req.WithContext(ctx.WaitCtx)
			start := time.Now()
			resp, err := clt.Do(req)
			if err != nil {
				ctx.nonBlockEnqErr(errors.Wrapf(err, "clt.Do error"))
				break doLoop
			}
			_ = resp.Body.Close()
			ctx.ReqCntOk.inc()

			takeDur := time.Now().Sub(start)
			takeInterval := int64(takeDur.Seconds())
			if takeInterval > atomic.LoadInt64(&ctx.MaxTakeSeconds) {
				atomic.StoreInt64(&ctx.MaxTakeSeconds, takeInterval)
			}
		}
		break
	}

	atomic.AddInt64(&ctx.GoCnt, -1)
	ctx.Wg.Done()
	ctx.SomeOneExitCh <- true
}

func work(ctx *boomContext) {

workLoop:
	for {
		for atomic.LoadInt64(&ctx.GoCnt) < ctx.GoCntLimit {
			// 连续发起的请求之间增加一点缓冲
			// 缓慢增长
			// 不需要可以删除
			if ctx.GoroutineInterval > 0 {
				select {
				case <-ctx.WaitCtx.Done():
					break workLoop
				case <-time.After(ctx.GoroutineInterval):
				}
			}

			ctx.Wg.Add(1)
			go func() {
				boom(ctx)
				ctx.Wg.Done()
			}()
		}

		select {
		case <-ctx.WaitCtx.Done():
			break workLoop
		case <-ctx.SomeOneExitCh:
		}
	}
}

func main() {

	ctx := new(boomContext)
	var cancel context.CancelFunc
	ctx.BoomPld = []byte(`{"sort":[{"timestamp":{"order":"desc"}}],"size":300}`)
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
	log.SetFlags(log.LstdFlags)
	ctx.GoCntLimit = 900
	ctx.WaitCtx, cancel = context.WithCancel(context.Background())
	ctx.SomeOneExitCh = make(chan bool, 1000)
	_ = cancel // we no need call cancel() by signal
	ctx.RAddr = "http://119.3.204.11:9200/server_info_report/_search?pretty"
	ctx.GoroutineInterval = time.Millisecond * 10
	ctx.ErrorCh = make(chan error, 5)

	ctx.Headers = make(map[string]string, 0)
	ctx.Headers["Content-Type"] = "application/json"
	ctx.ReqCntOk = newRateCounter()

	ctx.Wg.Add(1)
	go func() {
		work(ctx)
		ctx.Wg.Done()
	}()

	// to stat
	tick := time.Tick(time.Second * 3)
statLoop:
	for {
		select {
		case <-ctx.WaitCtx.Done():
			break statLoop
		case <-tick:
			reqRate := ctx.ReqCntOk.rate()
			log.Printf("Go %v/%v max task %v(s), %v rate %v",
				atomic.LoadInt64(&ctx.GoCnt), ctx.GoCntLimit,
				ctx.MaxTakeSeconds,
				ctx.ReqCntOk.Cnt, reqRate)
		case err := <-ctx.ErrorCh:
			log.Printf("err= %v", err)
		}
	}

	log.Printf("wait sub")
	ctx.Wg.Wait()
	log.Printf("main exit")
}
