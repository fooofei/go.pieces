package caslock

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sync/semaphore"
)

const (
	stateUndefined int32 = iota - 2 // -2
	stateWriteLock                  // -1
	stateNoLock                     // 0
	stateReadLock                   // >= 1
)

type CASMutex struct {
	state        int32
	turnstile    *semaphore.Weighted
	broadcastCh  chan struct{}
	broadcastMut sync.RWMutex
}

func (m *CASMutex) listen() <-chan struct{} {
	m.broadcastMut.RLock()
	defer m.broadcastMut.RUnlock()

	return m.broadcastCh
}

func (m *CASMutex) getState(n int32) int32 {
	switch {
	case n == stateWriteLock:
		return stateWriteLock
	case n == stateNoLock:
		return stateNoLock
	case n >= stateReadLock:
		return stateReadLock
	default:
		return stateUndefined
	}
}

func (m *CASMutex) broadcast() {
	newCh := make(chan struct{})

	m.broadcastMut.Lock()
	ch := m.broadcastCh
	m.broadcastCh = newCh
	m.broadcastMut.Unlock()

	close(ch)
}

func (m *CASMutex) TryLock(ctx context.Context) bool {
	if !m.turnstile.TryAcquire(1) {
		return false
	}
	defer m.turnstile.Release(1)
	return m.tryLock(ctx)
}

func (m *CASMutex) Lock(ctx context.Context) bool {
	if err := m.turnstile.Acquire(ctx, 1); err != nil {
		return false
	}
	defer m.turnstile.Release(1)
	return m.tryLock(ctx)
}

func (m *CASMutex) tryLock(ctx context.Context) bool {
	for {
		broadcastCh := m.listen()
		if atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.state)), stateNoLock, stateWriteLock) {
			return true
		}
		if ctx == nil {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-broadcastCh:
		}
	}
}

func (m *CASMutex) Unlock() {
	if ok := atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.state)), stateWriteLock, stateNoLock); !ok {
		panic("Unlock failed")
	}
	m.broadcast()
}

func (m *CASMutex) RLock(ctx context.Context) bool {
	if err := m.turnstile.Acquire(ctx, 1); err != nil {
		return false
	}
	m.turnstile.Release(1)
	for {
		broadcastCh := m.listen()
		n := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.state)))
		st := m.getState(n)
		switch st {
		case stateNoLock, stateReadLock:
			if atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.state)), n, n+1) {
				return true
			}
		}
		select {
		case <-ctx.Done():
			return false
		default:
			switch st {
			case stateNoLock, stateReadLock:
				runtime.Gosched()
				continue
			}
		}
		select {
		case <-ctx.Done():
			return false
		case <-broadcastCh:
		}
	}
}

func (m *CASMutex) RUnlock() {
	n := atomic.AddInt32((*int32)(unsafe.Pointer(&m.state)), -1)
	switch m.getState(n) {
	case stateUndefined, stateWriteLock:
		panic("RUnlock failed")
	case stateNoLock:
		m.broadcast()
	}
}

// NewCASMutex returns CASMutex lock
func NewCASMutex() *CASMutex {
	return &CASMutex{
		state:       stateNoLock,
		turnstile:   semaphore.NewWeighted(1),
		broadcastCh: make(chan struct{}),
	}
}
