package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// this is a simple cache.
// when request result have a ttl, we can use this cache to save result

type cacheDataType struct {
	Data     string
	UpdateAt time.Time
	Err      error
}

// use global atomic value to save data
var g_cache = &atomic.Pointer[cacheDataType]{}

// limit concurrent request, only send do real one work
// others will share the work result
var g_requestingOnce = &atomic.Pointer[sync.Once]{}

const cacheTTL = 2 * time.Second

// statType only for test, you can delete this when you use cache
type statType struct {
	RealRequestCount atomic.Int64
	RequestCount     atomic.Int64
}
type everyWorkStatType struct {
	RequestCount    int64
	UpdateDataCount int64
}

func getWorkResultWithCache(ctx context.Context, stat *statType) (string, error) {
	var slowFunc = func() {
		for {
			var tmpOnce = g_requestingOnce.Load()
			if tmpOnce == nil {
				tmpOnce = &sync.Once{}
				if !g_requestingOnce.CompareAndSwap(nil, tmpOnce) {
					continue
				}
			}
			// 这里会自动等待
			tmpOnce.Do(func() {
				stat.RealRequestCount.Add(1)
				var t, err = realWork(ctx)
				g_cache.Store(&cacheDataType{
					Data:     t,
					UpdateAt: time.Now(),
					Err:      err,
				})
				g_requestingOnce.Store(nil)
			})
			break
		}
	}

	var tmpCache = g_cache.Load()
	var now = time.Now()
	if tmpCache == nil {
		slowFunc()
	} else if tmpCache.Err != nil {
		slowFunc()
	} else if tmpCache.UpdateAt.Add(cacheTTL).Before(now) {
		slowFunc()
	}
	stat.RequestCount.Add(1)
	tmpCache = g_cache.Load()
	if tmpCache.Err != nil {
		return "", tmpCache.Err
	}
	return tmpCache.Data, nil
}

func realWork(ctx context.Context) (string, error) {
	select {
	case <-time.After(3 * time.Second):
	case <-ctx.Done():
	}
	var b = make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b), nil
}

func workRoutine(ctx context.Context, everyWorkStatType *everyWorkStatType, stat *statType) {
	var localToken = ""
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
		}
		var t, err = getWorkResultWithCache(ctx, stat)
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

func TestCacheWorkedAsExpect(t *testing.T) {
	var stat = &statType{}
	const RoutineCount = 5
	var everyWorkStat = [RoutineCount]everyWorkStatType{}
	var ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var wg = &sync.WaitGroup{}
	for i := 0; i < RoutineCount; i++ {
		wg.Add(1)
		go func(st *everyWorkStatType) {
			workRoutine(ctx, st, stat)
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
