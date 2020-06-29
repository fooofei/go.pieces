package xping

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	stats2 "github.com/montanaflynn/stats"
)

// The framework of xping

// WorkContext defines work arg and work stats
type WorkContext struct {
	// from command line args
	RAddr string
	N     int64
	W     time.Duration // timeout default 950ms
	//
	Wg      *sync.WaitGroup
	WaitCtx context.Context
	//
	Sent     int64
	Received int64
	Durs     []time.Duration
}

// PingOp defines which pinger, httping or tcping or other
type PingOp interface {
	Name() string
	// Before Ping, setup the global ready
	Ready(addr string) error
	// The ping may contains several steps, only timing steps which we want
	// return values
	//     @time.Duration take time in time.Duration
	//     @error indicates this ping success or fail
	Ping(waitCtx context.Context, addr string) (time.Duration, error)

	Close() error
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

func doOp(wkCtx *WorkContext, po PingOp) {
	bb := new(bytes.Buffer)
	const TimeFmt = "15:04:05"
	var count int64
	var err error
	var dur time.Duration

	ticker := time.Tick(time.Second)
	err = po.Ready(wkCtx.RAddr)
	if err != nil {
		log.Fatalf("fail Ready() error= %v", err)
	}
pingLoop:
	for {
		count += 1
		bb.Reset()
		start := time.Now()

		pingOpWaitCtx, _ := context.WithTimeout(wkCtx.WaitCtx, wkCtx.W)
		dur, err = po.Ping(pingOpWaitCtx, wkCtx.RAddr)
		durNanoSec := dur.Nanoseconds()
		durMillSec := float64(durNanoSec) / 1000 / 1000

		_, _ = fmt.Fprintf(bb, "> [%04v][%v] %v: %.2f ms",
			count, start.Format(TimeFmt), wkCtx.RAddr, durMillSec)
		wkCtx.Sent += 1
		wkCtx.Durs = append(wkCtx.Durs, dur)
		if err != nil {
			_, _ = fmt.Fprintf(bb, " err=%v", err)
		} else {
			wkCtx.Received += 1
		}
		fmt.Println(bb.String())
		//
		if count >= wkCtx.N {
			break pingLoop
		}

		select {
		case <-ticker:
		case <-wkCtx.WaitCtx.Done():
			break pingLoop
		}

		// use tick instead time.After()
		// 下面是不好的用法 计算剩余需要等待多长时间
		// wait next
		//end = time.Now()
		//// need wait
		//expectTime := start.Add(time.Nanosecond * 999 * 998 * 999)
		//if expectTime.UnixNano() > end.UnixNano() {
		//	select {
		//	case <-time.After(expectTime.Sub(end)):
		//	case <-wkCtx.WaitCtx.Done():
		//		break work
		//	}
		//}

	}
	_ = po.Close()
}

// Ping the entrance
func Ping(po PingOp) {
	var cancel context.CancelFunc
	var infinite bool
	var W int64
	wkCtx := new(WorkContext)
	wkCtx.Wg = new(sync.WaitGroup)
	wkCtx.WaitCtx, cancel = context.WithCancel(context.Background())

	flag.BoolVar(&infinite, "t", false, "Ping until stopped with Ctrl+C")
	flag.Int64Var(&wkCtx.N, "n", 4, "Number of requests to send")
	flag.Int64Var(&W, "w", 1000, "Wait timeout (ms) between two requests >50")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Please give addr\n")
		return
	}
	W -= 50
	if W <= 0 {
		fmt.Printf("Invalid wait timeout = %v\n", W)
		return
	}
	log.Printf("args t=%v n=%v w=%v", infinite, wkCtx.N, W)
	wkCtx.RAddr = args[0]
	wkCtx.W = time.Millisecond * time.Duration(W)

	fmt.Printf("=> %v %v for infinite= %v n= %v\n", po.Name(), wkCtx.RAddr, infinite, wkCtx.N)

	setupSignal(wkCtx.WaitCtx, wkCtx.Wg, cancel)
	if infinite {
		wkCtx.N = math.MaxInt64
	}
	doOp(wkCtx, po)
	cancel()
	wkCtx.Wg.Wait()
	sm := summary(wkCtx.Sent, wkCtx.Received, wkCtx.Durs)
	fmt.Printf(sm)
}

func summary(sent int64, received int64, durs []time.Duration) string {
	w := &bytes.Buffer{}
	statsData := stats2.LoadRawData(durs)

	_, _ = fmt.Fprintf(w, " Sent = %v, ", sent)
	_, _ = fmt.Fprintf(w, " Received = %v ", received)
	_, _ = fmt.Fprintf(w, "(%.1f%s)\n",
		float64(received*100)/float64(sent), "%%")

	min, _ := stats2.Min(statsData)
	_, _ = fmt.Fprintf(w, " Minimum = %.2f ms,", min/1000/1000)
	max, _ := stats2.Max(statsData)
	_, _ = fmt.Fprintf(w, " Maximum = %.2f ms\n", max/1000/1000)
	ave, _ := stats2.Mean(statsData)
	_, _ = fmt.Fprintf(w, " Average = %.2f ms,", ave/1000/1000)
	med, _ := stats2.Median(statsData)
	_, _ = fmt.Fprintf(w, " Median = %.2f ms\n", med/1000/1000)

	percentile90, _ := stats2.Percentile(statsData, float64(90))
	_, _ = fmt.Fprintf(w, " 90%s of Request <= %.2f ms\n",
		"%%", percentile90/1000/1000)
	percentile75, _ := stats2.Percentile(statsData, float64(75))
	_, _ = fmt.Fprintf(w, " 75%s of Request <= %.2f ms\n",
		"%%", percentile75/1000/1000)
	percentile50, _ := stats2.Percentile(statsData, float64(50))
	_, _ = fmt.Fprintf(w, " 50%s of Request <= %.2f ms\n",
		"%%", percentile50/1000/1000)
	return w.String()
}
