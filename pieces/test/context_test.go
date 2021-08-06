package test

import (
	"context"
	"testing"
)

func TestContextParent(t *testing.T){

	ctx,cancel := context.WithCancel(context.Background())

	subCtx,subCancel := context.WithCancel(ctx)


	cancel()

	select {
	case <-subCtx.Done():
		// will go here
		t.Logf("after cancel() subCtx.Done() return")
	default:
		t.Logf("after cancel() subCtx.Done() not return")
	}
	subCancel()

	select {
	case <-subCtx.Done():
		t.Logf("subCancel() subCtx.Done() return")
	default:
		t.Logf("subCancel() subCtx.Done() not return")
	}
}
