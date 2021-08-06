package test

import (
	"sync"
	"testing"
)

// 源自于我写的一个 BUG。
// 不是 defer 引起的延迟，而是 goroutine。
// 在 sub routine 操作了 A 资源，在此 routine 也操作了 A 资源，
// 应该显式的等 sub routine 退出，才能安全的继续操作 A 资源.
// 从下面的例子中能看出延迟。

func doingWork(t *testing.T, loopCnt int) {
	defer func() {
		t.Logf("called defer loopCnt= %v", loopCnt)
	}()
	t.Logf("doing work loopCnt= %v", loopCnt)

}

func TestNoDelayedDefer(t *testing.T) {
	for i := 0; i < 100; i += 1 {
		doingWork(t, i)
	}
}

func doingWorkWithRoutine(t *testing.T, loopCnt int, waitGrp *sync.WaitGroup) {
	noNeedWait := make(chan bool, 1)
	waitGrp.Add(1)
	go func() {
		select {
		case <-noNeedWait:
		}
		t.Logf("called defer loopCnt =%v", loopCnt)
		waitGrp.Done()
	}()
	defer close(noNeedWait)
	t.Logf("doing work loopCnt= %v", loopCnt)
}

func TestDelayedDefer(t *testing.T) {
	// there will have delay
	// delayed by go routine

	// no waitGrp will have panic
	// panic: Log in goroutine after TestDelayedDefer has completed
	waitGrp := new(sync.WaitGroup)
	for i := 0; i < 100; i += 1 {
		doingWorkWithRoutine(t, i, waitGrp)
	}
	waitGrp.Wait()

}
