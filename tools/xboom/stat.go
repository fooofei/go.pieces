package xboom

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync/atomic"
	"time"
)

type Stat struct {
	MaxTakeSeconds atomic.Int64
	ErrorCh        chan error
	BoomCnt        atomic.Int64
	BoomOkCnt      atomic.Int64
	BoomFailCnt    atomic.Int64
}

// enqueue error or ignored it
func (bc *Stat) nonBlockEnqErr(err error) {
	select {
	case bc.ErrorCh <- err:
	default:
	}
}

type rateStat struct {
	boomCnt     int64
	boomOkCnt   int64
	boomFailCnt int64
}

func newRateStat(stat *Stat) *rateStat {
	var rs = &rateStat{
		boomCnt:     stat.BoomCnt.Load(),
		boomOkCnt:   stat.BoomOkCnt.Load(),
		boomFailCnt: stat.BoomFailCnt.Load(),
	}
	return rs
}

func (rs *rateStat) dup() *rateStat {
	var r = &rateStat{}
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
	var w = &bytes.Buffer{}
	for _, v := range trends {
		fmt.Fprintf(w, "(%v ", v.boomCnt)
		fmt.Fprintf(w, "Ok %v ", v.boomOkCnt)
		fmt.Fprintf(w, "Fail %v )", v.boomFailCnt)
	}
	return strings.TrimRight(w.String(), " ")
}

func onlyKeep3(array []*rateStat) []*rateStat {
	var trimIdx = int(math.Max(0, float64(len(array)-3)))
	return array[trimIdx:]
}

func processStat(waitCtx context.Context, stat *Stat) {
	var tick = time.Tick(time.Second * 3)

	var oldTime = time.Now()
	var oldStat = &rateStat{}
	var trends = make([]*rateStat, 0)
loop:
	for {
		select {
		case <-waitCtx.Done():
			break loop
		case <-tick:
			var nowRateStat = newRateStat(stat)
			var maxTakeSeconds = stat.MaxTakeSeconds.Load()
			var now = time.Now()
			interval := int64(now.Sub(oldTime).Seconds())
			if interval > 0 {
				var intervalRateStat = nowRateStat.dup()
				intervalRateStat.sub(oldStat)
				intervalRateStat.perSecond(interval)
				trends = append(trends, intervalRateStat)
				trends = onlyKeep3(trends)
			}
			fmt.Printf("maxTake= %v (s) %v\n", maxTakeSeconds, trendsString(trends))
			oldTime = now
			oldStat = nowRateStat
		case err := <-stat.ErrorCh:
			log.Printf("err= %v", err)
		}
	}
}
