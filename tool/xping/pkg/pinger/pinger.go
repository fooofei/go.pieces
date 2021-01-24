package pinger

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

// The framework of pingable

// WorkContext defines work arg and work stats
type WorkContext struct {
	// from command line args
	RAddr    string
	N        int64
	W        time.Duration // timeout default 950ms
	Interval time.Duration
	//
	WaitCtx context.Context
	//
	Sent     int64
	Received int64
	Durs     []time.Duration
}

// Pinger defines which pinger, httping or tcping or other
type Pinger interface {
	Name() string
	// Before DoPing, setup the global ready
	Ready(ctx context.Context, addr string) error
	// The ping may contains several steps, only timing steps which we want
	// return values
	//     @time.Duration take time in time.Duration
	//     @error indicates this ping success or fail
	Ping(ctx context.Context, addr string) (time.Duration, error)

	Close() error
}

func setupSignal(waitCtx context.Context) context.Context {
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)
	signalContext, cancel := context.WithCancel(waitCtx)
	go func() {
		select {
		case <-waitCtx.Done():
		case <-sigCh:
			cancel()
		}
	}()
	return signalContext
}

func doOp(wkCtx *WorkContext, po Pinger) {
	bb := new(bytes.Buffer)
	const TimeFmt = "15:04:05"
	var count int64
	var err error
	var dur time.Duration

	ticker := time.Tick(wkCtx.Interval)
	err = po.Ready(wkCtx.WaitCtx, wkCtx.RAddr)
	if err != nil {
		log.Fatalf("failed Ready() error= <%T> %v", err, err)
	}
pingLoop:
	for {
		count += 1
		bb.Reset()
		start := time.Now()

		pingOpWaitCtx, cancel := context.WithTimeout(wkCtx.WaitCtx, wkCtx.W)
		dur, err = po.Ping(pingOpWaitCtx, wkCtx.RAddr)
		cancel()
		durNanoSec := dur.Nanoseconds()
		durMillSec := float64(durNanoSec) / 1000 / 1000

		_, _ = fmt.Fprintf(bb, "> [%05v][%v] %v: %.2f ms",
			count, start.Format(TimeFmt), wkCtx.RAddr, durMillSec)
		wkCtx.Sent += 1
		wkCtx.Durs = append(wkCtx.Durs, dur)
		if err != nil {
			_, _ = fmt.Fprintf(bb, " err=<%T> %v", err, err)
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

// DoPing the entrance of ping
func DoPing(po Pinger) {
	var cancel context.CancelFunc
	var infinite bool
	var W int64
	var interval time.Duration
	wkCtx := new(WorkContext)
	wg := new(sync.WaitGroup)
	wkCtx.WaitCtx, cancel = context.WithCancel(context.Background())
	defer cancel()

	flag.BoolVar(&infinite, "t", false, "DoPing until stopped with Ctrl+C")
	flag.Int64Var(&wkCtx.N, "n", 4, "Number of requests to send")
	flag.Int64Var(&W, "w", 1000, "Wait timeout (ms) between two requests >50")
	flag.DurationVar(&interval, "i", time.Second, "Interval between two ping")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Please give addr for ping\n")
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
	wkCtx.Interval = interval

	fmt.Printf("=> %v %v for infinite=`%v` n=`%v`\n", po.Name(), wkCtx.RAddr, infinite, wkCtx.N)

	signalCtx := setupSignal(wkCtx.WaitCtx)
	if infinite {
		wkCtx.N = math.MaxInt64
	}
	wg.Add(1)
	go func() {
		doOp(wkCtx, po)
		wg.Done()
		cancel()
	}()
	select {
	case <-signalCtx.Done():
		cancel()
	case <-wkCtx.WaitCtx.Done():
	}

	wg.Wait()
	text := summary(wkCtx.Sent, wkCtx.Received, wkCtx.Durs)
	fmt.Printf(text)
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
