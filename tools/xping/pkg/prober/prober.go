// Package prober is the framework of prober
package prober

import (
	"context"
	"flag"
	"fmt"
	"github.com/fooofei/go_pieces/tools/ping/pkg/prompt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ProbeContext defines input arguments and stats context
type ProbeContext struct {
	// from command line args
	RAddr      string
	ProbeCount int64
	Timeout    time.Duration // timeout with default 950ms
	Interval   time.Duration
	//
	WaitCtx context.Context
	//
	Sent         int64
	Received     int64
	DurationList []time.Duration
}

func processProbe(ctx *ProbeContext, prober Prober) error {
	var count int64
	var err error
	var dur time.Duration
	var text string

	var ticker = time.Tick(ctx.Interval)
	if err = prober.Ready(ctx.WaitCtx, ctx.RAddr); err != nil {
		return fmt.Errorf("failed ready probe with error %w", err)
	}
loop:
	for {
		count += 1
		var thisWaitCtx, thisCancel = context.WithTimeout(ctx.WaitCtx, ctx.Timeout)
		var start = time.Now()
		text, err = prober.Probe(thisWaitCtx, ctx.RAddr)
		dur = time.Now().Sub(start)
		var p = prompt.With(count, start, ctx.RAddr).WithBlank().WithTakeTime(dur)
		if text != "" {
			p = p.WithBlank().WithText(text)
		}
		if err != nil {
			p = p.WithBlank().WithError(err)
		}
		fmt.Printf("%s\n", p.String())
		thisCancel()
		ctx.DurationList = append(ctx.DurationList, dur)
		ctx.Sent += 1
		if err == nil {
			ctx.Received += 1
		}
		if count >= ctx.ProbeCount {
			break loop
		}

		select {
		case <-ticker:
		case <-ctx.WaitCtx.Done():
			break loop
		}

		// why use tick instead time.After() ?
		// 下面是不好的用法 计算剩余需要等待多长时间
		// wait next
		//end = time.Now()
		//// need wait
		//expectTime := start.Add(time.Nanosecond * 999 * 998 * 999)
		//if expectTime.UnixNano() > end.UnixNano() {
		//	select {
		//	case <-time.After(expectTime.Sub(end)):
		//	case <-ctx.WaitCtx.Done():
		//		break work
		//	}
		//}

	}
	prober.Close()
	return nil
}

// Do is the entry of probe
func Do(prober Prober) {
	var cancel context.CancelFunc
	var infinite bool
	var W int64
	var interval time.Duration
	var probeContext = new(ProbeContext)
	probeContext.WaitCtx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flag.BoolVar(&infinite, "t", false, "DoPing until stopped with Ctrl+C")
	flag.Int64Var(&probeContext.ProbeCount, "n", 4, "Number of requests to send")
	flag.Int64Var(&W, "w", 1000, "Wait timeout (ms) between two requests >50")
	flag.DurationVar(&interval, "i", time.Second, "Interval between two ping")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Exit for no args found.\ne.g. %s\n", prober.Example())
		return
	}
	W -= 50 // decrease 50 for probe loop timeout
	if W <= 0 {
		fmt.Printf("Invalid wait timeout = %v, must > 50\n", W)
		return
	}
	fmt.Printf("=> Print args t=%v n=%v w=%v\n", infinite, probeContext.ProbeCount, W)
	probeContext.RAddr = args[0]
	probeContext.Timeout = time.Millisecond * time.Duration(W)
	probeContext.Interval = interval
	if infinite {
		probeContext.ProbeCount = math.MaxInt64
	}
	fmt.Printf("=> start %v to %v for count %v\n", prober.Name(), probeContext.RAddr, probeContext.ProbeCount)
	var errCh = make(chan error, 1)
	go func() {
		var err2 = processProbe(probeContext, prober)
		cancel()
		errCh <- err2
		close(errCh)
	}()
	var err = <-errCh
	if err != nil {
		panic(err)
	}
	var text = summary(probeContext.Sent, probeContext.Received, probeContext.DurationList)
	fmt.Printf(text)
}
