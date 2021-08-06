package main

// 抓取懒投资网站公布数据 分析懒投资回款趋势

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func setupSignal(waitCtx context.Context, waitGrp *sync.WaitGroup, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)
	waitGrp.Add(1)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-waitCtx.Done():
		}
		waitGrp.Done()
	}()
}

func main() {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	setupSignal(ctx, wg, cancel)
	crawlLantouzi(ctx)

	cancel()
	wg.Wait()
	log.Printf("main exit")
}
