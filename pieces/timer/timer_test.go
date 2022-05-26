package timer

import (
	"fmt"
	"github.com/shoenig/test"
	"runtime"
	"testing"
	"time"
)

func reuseTimerWithLength(t *time.Timer, d time.Duration) {
	if !t.Stop() && len(t.C) > 0 {
		<-t.C
	}
	t.Reset(d)
}

func reuseTimerWithDefault(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(d)
}

func waitTimerWithDefault() error {
	tmr := time.NewTimer(0)
	reuseTimerWithDefault(tmr, time.Minute)
	select {
	case <-tmr.C:
		return fmt.Errorf("unexpected firing of Timer")
	default:
		return nil
	}
}

func waitTimerWithLength() error {
	tmr := time.NewTimer(0)
	reuseTimerWithLength(tmr, time.Minute)
	select {
	case <-tmr.C:
		return fmt.Errorf("unexpected firing of Timer")
	default:
		return nil
	}
}

// TestWaitTimerWithDefault 证明 select default 也不能正确使用 timer.Reset
func TestWaitTimerWithDefault(t *testing.T) {
	runtime.GOMAXPROCS(2)
	for i := 0; ; i++ {
		if err := waitTimerWithDefault(); err != nil {
			t.Logf("occured error '%v' when loop count %v", err, i)
			break
		}
	}

}

// TestWaitTimerWithLength 证明 len(timer.C) 也不能正确使用 timer.Reset
func TestWaitTimerWithLength(t *testing.T) {
	runtime.GOMAXPROCS(2)
	for i := 0; ; i++ {
		if err := waitTimerWithLength(); err != nil {
			t.Logf("occured error '%v' when loop count %v", err, i)
			break
		}
	}
}

func TestMyTimerReset(tst *testing.T) {
	t := New(time.Second)
	defer t.Stop()

	tst.Logf("wait 1s")
	timerArrive := false
	select {
	case <-t.After(2 * time.Second):
		t.SetUnActive()
		timerArrive = true
	case <-time.After(time.Second):
		timerArrive = false
	}
	test.Eq(tst, timerArrive, false)

	tst.Logf("wait 2s")

	select {
	case <-t.After(2 * time.Second):
		t.SetUnActive()
		timerArrive = true
	case <-time.After(4 * time.Second):
		timerArrive = false
	}
	test.Eq(tst, timerArrive, true)

	tst.Logf("wait 1s")
	timerArrive = false
	select {
	case <-t.After(3 * time.Second):
		t.SetUnActive()
		timerArrive = true
	case <-time.After(time.Second):
		timerArrive = false
	}
	test.Eq(tst, timerArrive, false)
}
