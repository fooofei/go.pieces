package test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"time"
)

const cacheTTL = 2 * time.Second

// traceType only for test, you can delete this when you use cache
type traceType struct {
	RealRequestCount atomic.Int64
	RequestCount     atomic.Int64
}

// this is a simple cache.
// when request result have a ttl, we can use this cache to save result
type cacheDataTypeAtomic struct {
	Data     string
	UpdateAt time.Time
	Err      error
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

// use global atomic value to save data
var gCacheAtomic = &atomic.Pointer[cacheDataTypeAtomic]{}

// limit concurrent request, only send do real one work
// others will share the work result
var gRequestingOnce = &atomic.Pointer[sync.Once]{}

func getWorkResultWithCacheAtomic(ctx context.Context, stat *traceType) (string, error) {
	var slowFunc = func() {
		for {
			var tmpOnce = gRequestingOnce.Load()
			if tmpOnce == nil {
				tmpOnce = &sync.Once{}
				if !gRequestingOnce.CompareAndSwap(nil, tmpOnce) {
					continue
				}
			}
			// 这里会自动等待
			tmpOnce.Do(func() {
				stat.RealRequestCount.Add(1)
				var t, err = realWork(ctx)
				gCacheAtomic.Store(&cacheDataTypeAtomic{
					Data:     t,
					UpdateAt: time.Now(),
					Err:      err,
				})
				gRequestingOnce.Store(nil)
			})
			break
		}
	}

	var tmpCache = gCacheAtomic.Load()
	var now = time.Now()
	if tmpCache == nil {
		slowFunc()
	} else if tmpCache.Err != nil {
		slowFunc()
	} else if tmpCache.UpdateAt.Add(cacheTTL).Before(now) {
		slowFunc()
	}
	stat.RequestCount.Add(1)
	tmpCache = gCacheAtomic.Load()
	if tmpCache.Err != nil {
		return "", tmpCache.Err
	}
	return tmpCache.Data, nil
}
