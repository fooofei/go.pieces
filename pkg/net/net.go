package net

import (
	"context"
	"io"
	"sync"
)

// CloseWhenContext will wait to close `toBeClose` when ctx.Done()
// if `toBeClose` closed, the returned context will Done()
// if you cannot wait, you can close the returned stop func
func CloseWhenContext(ctx context.Context, toBeClose io.Closer) (context.Context, context.CancelFunc) {
	noWait := make(chan bool, 1)
	ctx2,cancel2 := context.WithCancel(ctx)
	go func() {
		select {
		case <-noWait:
		case <-ctx.Done():
		}
		_ = toBeClose.Close()
		cancel2()
	}()
	cancelOnce := &sync.Once{}
	cancelFunc := func() {
		cancelOnce.Do(func() {
			close(noWait)
		})
	}
	return ctx2, cancelFunc
}
