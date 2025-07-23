package test

import (
	"sync"
	"sync/atomic"
	"time"
)

// this is a simple cache.
// when request result have a ttl, we can use this cache to save result
type cacheDataTypeAtomic struct {
	Data     any
	UpdateAt time.Time
	Err      error
}

// use global atomic value to save data
var gCacheAtomic = &atomic.Pointer[cacheDataTypeAtomic]{}

// limit concurrent request, only send do real one work
// others will share the work result
var gRequestingOnce = &atomic.Pointer[sync.Once]{}

func getWorkResultWithCacheAtomic(taskFn func() (any, error), cacheTTL time.Duration) (any, error) {
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
				var t, err = taskFn()
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
	tmpCache = gCacheAtomic.Load()
	if tmpCache.Err != nil {
		return "", tmpCache.Err
	}
	return tmpCache.Data, nil
}
