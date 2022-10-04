package net

import (
	"context"
	"io"
	"sync"
)

// CloseWhenContext will wait to close `toBeClose` when ctx.Done()
// or you can close it by call the returned cancel()
// when `toBeClose` is closed, the returned context will be done
func CloseWhenContext(ctx context.Context, toBeClose io.Closer) (context.Context, context.CancelFunc) {
	var noWait = make(chan bool, 1)
	var ctx2, cancel2 = context.WithCancel(ctx)
	go func() {
		select {
		case <-noWait:
		case <-ctx.Done():
		}
		toBeClose.Close()
		cancel2()
	}()
	var cancelOnce = &sync.Once{}
	var cancelFunc = func() {
		cancelOnce.Do(func() {
			close(noWait)
		})
	}
	return ctx2, cancelFunc
}
