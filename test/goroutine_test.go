package go_pieces

import (
	"sync"
	"testing"
)

func routine(t *testing.T, idx int, v *byte) {
	t.Logf("routine %v got %v", idx, v)
}

func TestGoRoutine(t *testing.T) {
	//    goroutine_test.go:19: ptra = [0xc00008c3d0 0xc00008c3d1 0xc00008c3d2 0xc00008c3d3 0xc00008c3d4]
	//    goroutine_test.go:9: routine 4 got 0xc00008c3d4
	//    goroutine_test.go:9: routine 4 got 0xc00008c3d4
	//    goroutine_test.go:9: routine 4 got 0xc00008c3d4
	//    goroutine_test.go:9: routine 4 got 0xc00008c3d4
	//    goroutine_test.go:9: routine 4 got 0xc00008c3d4
	//    goroutine_test.go:30: second
	//    goroutine_test.go:9: routine 4 got 0xc00008c3d4
	//    goroutine_test.go:9: routine 2 got 0xc00008c3d2
	//    goroutine_test.go:9: routine 3 got 0xc00008c3d3
	//    goroutine_test.go:9: routine 0 got 0xc00008c3d0
	//    goroutine_test.go:9: routine 1 got 0xc00008c3d1

	a := make([]byte, 5)
	ptra := make([]*byte, 5)

	for idx, _ := range a {
		ptra[idx] = &a[idx]
	}
	t.Logf("ptra = %v", ptra)

	waitGrp := &sync.WaitGroup{}
	for idx, _ := range a {
		waitGrp.Add(1)
		go func() {
			routine(t, idx, &a[idx])
			waitGrp.Done()
		}()
	}
	waitGrp.Wait()
	t.Logf("second")
	// compare

	for idx, _ := range a {
		waitGrp.Add(1)
		go func(arg0 *testing.T, arg1 int, arg2 *byte) {
			routine(arg0, arg1, arg2)
			waitGrp.Done()
		}(t, idx, &a[idx])
	}
	waitGrp.Wait()
}
