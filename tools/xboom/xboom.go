package xboom

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func processBoom(waitCtx context.Context, stat *Stat, boom Boomable) {
	var err error
loop:
	for {
		select {
		case <-waitCtx.Done():
			break loop
		default:
		}

		stat.BoomCnt.Add(1)
		var start = time.Now()
		err = boom.Shoot(waitCtx)
		var sec = int64(time.Since(start).Seconds())
		if sec > stat.MaxTakeSeconds.Load() {
			stat.MaxTakeSeconds.Store(sec)
		}
		if err != nil {
			stat.BoomFailCnt.Add(1)
			stat.nonBlockEnqErr(fmt.Errorf("fail Shoot with error %w", err))
		} else {
			stat.BoomOkCnt.Add(1)
		}
	}
}

// Gatelin is the name of 加特林, a kind of gun
func Gatelin(boom Boomable) {
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
	log.SetFlags(log.LstdFlags)

	var threadCount int64
	var remoteAddr string

	flag.Int64Var(&threadCount, "go", 10, "the count of routine")
	flag.StringVar(&remoteAddr, "addr", "", "the boom addr")
	flag.Parse()

	if remoteAddr == "" {
		flag.PrintDefaults()
		return
	}
	if threadCount <= 0 {
		flag.PrintDefaults()
		return
	}
	var waitCtx, cancel = signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	var stat = &Stat{}
	stat.ErrorCh = make(chan error, 5)
	var wg = &sync.WaitGroup{}
	if err := boom.LoadBullet(waitCtx, remoteAddr); err != nil {
		panic(err)
	}
	for i := int64(0); i < threadCount; i++ {
		wg.Add(1)
		go func() {
			processBoom(waitCtx, stat, boom)
			wg.Done()
		}()
	}
	processStat(waitCtx, stat)
	cancel()
	log.Printf("wait to exit")
	wg.Wait()
	boom.Close()
	log.Printf("exit")
}
