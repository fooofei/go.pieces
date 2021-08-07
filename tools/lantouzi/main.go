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


func main() {
	wg := &sync.WaitGroup{}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	crawlLantouzi(ctx)

	cancel()
	wg.Wait()
	log.Printf("main exit")
}
