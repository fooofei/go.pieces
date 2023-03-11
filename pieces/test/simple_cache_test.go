package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type everyWorkStatType struct {
	RequestCount    int64
	UpdateDataCount int64
}

func workRoutine(ctx context.Context, fn func(ctx2 context.Context, traceType2 *traceType) (string, error),
	everyWorkStatType *everyWorkStatType, stat *traceType) {
	var localToken = ""
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
		}
		var t, err = fn(ctx, stat)
		everyWorkStatType.RequestCount += 1
		if err != nil {
			continue
		}
		if t != localToken {
			everyWorkStatType.UpdateDataCount += 1
			localToken = t
		}
	}
}

func TestAtomicCacheWorkedAsExpect(t *testing.T) {
	var stat = &traceType{}
	const RoutineCount = 5
	var everyWorkStat = [RoutineCount]everyWorkStatType{}
	var ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var wg = &sync.WaitGroup{}
	for i := 0; i < RoutineCount; i++ {
		wg.Add(1)
		go func(st *everyWorkStatType) {
			workRoutine(ctx, getWorkResultWithCacheAtomic, st, stat)
			wg.Done()
		}(&everyWorkStat[i])
	}
	wg.Wait()

	var total = int64(0)
	for i := 0; i < RoutineCount; i++ {
		total += everyWorkStat[i].RequestCount
		assert.Equal(t, everyWorkStat[i].UpdateDataCount, stat.RealRequestCount.Load())
	}
	assert.Equal(t, total, stat.RequestCount.Load())
}
