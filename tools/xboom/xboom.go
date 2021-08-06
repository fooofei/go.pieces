package xboom

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

type boomContext struct {
	// for start
	Wg      *sync.WaitGroup
	WaitCtx context.Context
	GoCnt   int64
	RAddr   string

	// for stat
	MaxTakeSeconds int64
	ErrorCh        chan error
	stat           rateStat
}

type rateStat struct {
	boomCnt     int64
	boomOkCnt   int64
	boomFailCnt int64
}

type BoomOp interface {
	// LoadBullet for shoot many times, run in routine
	LoadBullet(waitCtx context.Context, addr string) error
	// return values
	//      @time.Duration take time
	//      @error shoot success or fail
	Shoot(waitCtx context.Context) (time.Duration, error)
	io.Closer
}

// enqueue error or ignored it
func (bc *boomContext) nonBlockEnqErr(err error) {
	select {
	case bc.ErrorCh <- err:
	default:
	}
}

func (rs *rateStat) safeDup() *rateStat {
	r := &rateStat{}
	r.boomOkCnt = atomic.LoadInt64(&rs.boomOkCnt)
	r.boomFailCnt = atomic.LoadInt64(&rs.boomFailCnt)
	r.boomCnt = atomic.LoadInt64(&rs.boomCnt)
	return r
}

func (rs *rateStat) dup() *rateStat {
	r := &rateStat{}
	*r = *rs
	return r
}

func (rs *rateStat) sub(b *rateStat) {
	rs.boomCnt = rs.boomCnt - b.boomCnt
	rs.boomOkCnt = rs.boomOkCnt - b.boomOkCnt
	rs.boomFailCnt = rs.boomFailCnt - b.boomFailCnt
}

func (rs *rateStat) perSecond(interval int64) {
	rs.boomCnt = rs.boomCnt / interval
	rs.boomOkCnt = rs.boomOkCnt / interval
	rs.boomFailCnt = rs.boomFailCnt / interval
}

func trendsString(trends []*rateStat) string {
	w := &bytes.Buffer{}
	for _, v := range trends {
		_, _ = fmt.Fprintf(w, "(%v ", v.boomCnt)
		_, _ = fmt.Fprintf(w, "Ok %v ", v.boomOkCnt)
		_, _ = fmt.Fprintf(w, "Fail %v )", v.boomFailCnt)
	}
	return strings.TrimRight(w.String(), " ")
}

func state(boomCtx *boomContext) {

	tick := time.Tick(time.Second * 3)

	oldTime := time.Now()
	oldStat := &rateStat{}
	trends := make([]*rateStat, 0)
stateLoop:
	for {
		select {
		case <-boomCtx.WaitCtx.Done():
			break stateLoop
		case <-tick:
			nowRateStat := boomCtx.stat.safeDup()
			maxTake := atomic.LoadInt64(&boomCtx.MaxTakeSeconds)

			now := time.Now()
			intervalRateStat := nowRateStat.dup()
			intervalRateStat.sub(oldStat)
			interval := int64(now.Sub(oldTime).Seconds())

			if interval > 0 {
				t := intervalRateStat.dup()
				t.perSecond(interval)
				trimIdx := int(math.Max(0, float64(len(trends)-3)))
				trends = append(trends, t)
				trends = trends[trimIdx:]
			}
			fmt.Printf("maxTake= %v (s) %v\n", maxTake, trendsString(trends))

			oldTime = now
			oldStat = nowRateStat

		case err := <-boomCtx.ErrorCh:
			log.Printf("err= %v", err)
		default:
		}

	}
}

func boom(boomCtx *boomContext, boomOp BoomOp) {
	var err error
	var dur time.Duration

boomLoop:
	for {
		select {
		case <-boomCtx.WaitCtx.Done():
			break boomLoop
		case <-time.After(time.Second):
		default:
		}

		err = boomOp.LoadBullet(boomCtx.WaitCtx, boomCtx.RAddr)
		if err != nil {
			boomCtx.nonBlockEnqErr(errors.Wrapf(err, "fail LoadBullet"))
			continue
		}
	shootLoop:
		for {
			select {
			case <-boomCtx.WaitCtx.Done():
				break shootLoop
			default:
			}

			atomic.AddInt64(&boomCtx.stat.boomCnt, 1)
			dur, err = boomOp.Shoot(boomCtx.WaitCtx)
			takeInterval := int64(dur.Seconds())
			if takeInterval > atomic.LoadInt64(&boomCtx.MaxTakeSeconds) {
				atomic.StoreInt64(&boomCtx.MaxTakeSeconds, takeInterval)
			}
			if err != nil {
				atomic.AddInt64(&boomCtx.stat.boomFailCnt, 1)
				boomCtx.nonBlockEnqErr(errors.Wrapf(err, "fail Shoot"))
				break shootLoop
			}
			atomic.AddInt64(&boomCtx.stat.boomOkCnt, 1)

		}

		_ = boomOp.Close()
	}
}

func setupSignal(waitCtx context.Context, waitGrp *sync.WaitGroup, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)
	waitGrp.Add(1)
	go func() {
		select {
		case <-waitCtx.Done():
		case <-sigCh:
			cancel()
		}
		waitGrp.Done()
	}()
}

// Gatelin is the name of 加特林, a kind of gun
func Gatelin(boomOp BoomOp) {
	boomCtx := &boomContext{}
	var cancel context.CancelFunc
	var i int64
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
	log.SetFlags(log.LstdFlags)
	boomCtx.Wg = &sync.WaitGroup{}

	flag.Int64Var(&boomCtx.GoCnt, "gocnt", 10, "the count of routine")
	flag.StringVar(&boomCtx.RAddr, "addr", "", "the boom addr")
	flag.Parse()

	if boomCtx.RAddr == "" {
		flag.PrintDefaults()
		return
	}
	if boomCtx.GoCnt <= 0 {
		flag.PrintDefaults()
		return
	}
	boomCtx.WaitCtx, cancel = context.WithCancel(context.Background())
	boomCtx.ErrorCh = make(chan error, 5)

	for i = 0; i < boomCtx.GoCnt; i++ {
		boomCtx.Wg.Add(1)
		go func() {
			boom(boomCtx, boomOp)
			boomCtx.Wg.Done()
		}()
	}
	setupSignal(boomCtx.WaitCtx, boomCtx.Wg, cancel)
	state(boomCtx)
	cancel()
	log.Printf("wait to exit")
	boomCtx.Wg.Wait()
	log.Printf("exit")
}
