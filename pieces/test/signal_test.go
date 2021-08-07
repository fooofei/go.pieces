package test

import (
	"context"
	"gotest.tools/assert"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"testing"
)

func TestSignalContext(t *testing.T) {
	ctx,cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	stx, cancelStx := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	var value int32
	atomic.AddInt32(&value, 1)
	go func() {
		<- stx.Done()
		atomic.AddInt32(&value, -1)
		cancelCtx()
	}()

	// test we call call twice
	cancelStx()
	cancelStx()
	<- ctx.Done()
	assert.Equal(t, atomic.LoadInt32(&value), int32(0))
}