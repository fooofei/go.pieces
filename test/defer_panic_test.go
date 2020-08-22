package go_pieces

import (
	"testing"
	"time"
)

// TestDeferPanicMultiRoutine will test defer called in multi routine
// 防止破坏整个测试 先关闭测试
func testDeferPanicMultiRoutine(t *testing.T) {
	t.Logf("enter main routine")

	go func() {
		t.Logf("enter routine 1")
		// will not call this defer
		defer t.Logf("leave routine 1")
		time.Sleep(3 * time.Second)
	}()

	go func() {
		t.Logf("enter routine 2")
		// only call this defer
		// 仅仅调用了 panic routine 中的 defer ，其他 routine 中的 defer 都没调用
		defer t.Logf("leave routine 2")
		panic("panic at routine 2")
	}()

	time.Sleep(5 * time.Second)
	t.Logf("leave main routine")

}
