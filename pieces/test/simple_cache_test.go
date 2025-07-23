package test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type everyWorkStatType struct {
	RequestCount    int64
	UpdateDataCount int64
}

func realWork(ctx context.Context) (string, error) {
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
	}
	var b = make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b), nil
}

// traceType only for test, you can delete this when you use cache
type traceType struct {
	RealRequestCount atomic.Int64
	RequestCount     atomic.Int64
}

func TestAtomicCacheWorkedAsExpect(t *testing.T) {
	var stat = &traceType{}

	const RoutineCount = 6
	var everyWorkStat = [RoutineCount]everyWorkStatType{}
	var ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var wg = &sync.WaitGroup{}
	for i := 0; i < RoutineCount; i++ {
		wg.Add(1)
		go func(st *everyWorkStatType) {
			defer wg.Done()
			var localToken = ""
		loop:
			for {
				select {
				case <-ctx.Done():
					break loop
				default:
				}
				var t, err = getWorkResultWithCacheAtomic(func() (any, error) {
					stat.RealRequestCount.Add(1)
					return realWork(ctx)
				}, 2*time.Second)
				st.RequestCount += 1
				stat.RequestCount.Add(1)
				if err != nil {
					continue
				}
				if t != localToken {
					st.UpdateDataCount += 1
					localToken = t.(string)
				}
			}
		}(&everyWorkStat[i])
	}
	wg.Wait()

	var total = int64(0)
	var expect = make([]int64, 0)
	var actual = make([]int64, 0)
	for i := 0; i < RoutineCount; i++ {
		total += everyWorkStat[i].RequestCount
		// 预期所有协程更新 token 的次数相同，且与真实请求次数相同
		actual = append(actual, everyWorkStat[i].UpdateDataCount)
		expect = append(expect, stat.RealRequestCount.Load())
	}
	assert.Equal(t, stat.RequestCount.Load(), total, "test request count")
	assert.Equal(t, expect, actual, "test real request count")
}
