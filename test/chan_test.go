package go_pieces

import (
	"bytes"
	"context"
	"fmt"
	"gotest.tools/assert"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// 1 告诉 sub routine exit，需要1个 chan ，在 main 中 close chan 就可以
//   通知 sub routine 全部退出, sub routine 知道 chan return 或者 不可读取 就退出
// 2 等待多个 sub routine 结束，如果是死等，需要 Add(1)/Done()/sync.WaitGroup.Wait
//   如果是超时的等， 需要搭配一个 chan 个数为 routine的个数，routine结束的时候，
//   向chan 中写入数据，观察 chan 中个数够了，然后触发另一个chan，说够了

func produce(q chan<- int) {
	for i := 0; i < 10; i += 1 {
		q <- i + 1
	}
	close(q)
}

func consume(q <-chan int, finish chan<- string) {
	w := &bytes.Buffer{}
	_, _ = fmt.Fprintf(w, "{")
	for {
		v, more := <-q
		if !more {
			break
		}
		_, _ = fmt.Fprintf(w, "%v,", v)
	}
	_, _ = fmt.Fprintf(w, "}")
	finish <- w.String()
	close(finish)

}

func TestSimpleChanNoSelect(t *testing.T) {
	q := make(chan int, 5)
	finish := make(chan string, 1)

	go produce(q)
	go consume(q, finish)

	// 猜测这里打印顺序会不稳定，其他 go routine 也在打印的话，
	// 这里也在打印 与这里没有先后顺序
	//fmt.Println("Wait for consume exit")

	v := <-finish
	assert.Equal(t, v, "{1,2,3,4,5,6,7,8,9,10,}")
}

func TestBasicChanSndRcvOrder(t *testing.T) {

	q := make(chan string, 0) // no buffer
	count := 6

	// 无 buffer的chan，预期snd每次都阻塞在对方rcv之后才返回
	// 但是观察到的是
	// 第一次 send 之后就阻塞住了 recv 之后才得到返回
	// 下面的 send 就可以类似 buffer=1 了 也不是每次都是

	// 需要解释这种现象
	// 也有可能是printf 产生了先后顺序

	callSeq := make(chan string, 100)
	go func(count int, q chan<- string) {
		for i := 1; i <= count; i++ {
			callSeq <- fmt.Sprintf("enter tx i=%v", i)
			q <- fmt.Sprintf("push msg %v", i)
			callSeq <- fmt.Sprintf("leave tx i=%v", i)
		}
		close(q)
	}(count, q)
	for i := 1; ; i++ {
		v, more := <-q
		if !more {
			break
		}
		callSeq <- fmt.Sprintf("rx i=%v v=%v", i, v)
	}
	close(callSeq)
	for seq := range callSeq {
		t.Logf("%v", seq)
	}
}

func TestChanBroadcastRoutineClose(t *testing.T) {
	var done = make(chan bool)
	// 一个chan给2个goroutine 通信
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		<-done
		t.Logf("TestChanBroadcastRoutineClose routine1 exit")
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		<-done
		t.Logf("TestChanBroadcastRoutineClose routine2 exit")
		wg.Done()
	}()

	close(done)
	wg.Wait()
	t.Logf("TestChanBroadcastRoutineClose main exit")
}

func TestSelectNilChan(t *testing.T) {
	a := make(chan string, 1)
	a <- "1"
	a = nil

	// if no default, will infinit wait
	// a is nil, select return
	select {
	case v := <-a:
		t.Logf("1got %v from chan\n", v)
	default:
	}

	select {
	case v := <-a:
		t.Logf("2got %v from chan\n", v)
	case <-time.After(time.Second):
		// we also wait a timer
		t.Logf("second select timeout of time.Second")
	}

	t.Logf("main exit\n")
}

func TestCloseChanTwice(t *testing.T) {
	defer func() {
		r := recover()
		t.Logf("We got \"%v\" from recover", r)
	}()
	ch := make(chan bool)
	close(ch)
	close(ch)
}

// 这种技法的使用场景是一个 TCP 长连接，一直在收发数据
// 如何给 读或者写 做 Context 接管结束生命周期呢
// 就是这样使用
func registerCloseCnn(waitCtx context.Context, c io.Closer) chan bool {
	// 甚至不需要 sync.WaitGroup ，因为一定会退出
	noNeedWait := make(chan bool, 1)

	go func() {
		select {
		case <-waitCtx.Done():
		case <-noNeedWait:
		}
		_ = c.Close()
	}()
	return noNeedWait
}

func TestRegisterCloseChanForever(t *testing.T) {
	cnn := &net.TCPConn{}
	waitCtx, cancel := context.WithCancel(context.Background())

	noNeedWait := registerCloseCnn(waitCtx, cnn)

	// do work
	for {
		break
	}
	// this will cnn.Close()
	close(noNeedWait)
	cancel()
}

// 另一个技法是 给一个没有 context 的操作封装一个 context 技法
// 这样也能像使用 DialContext 那样随心所欲结束
func registerWaitChan(waitCtx context.Context, c io.Closer) (chan bool, *sync.WaitGroup) {
	noNeedWait := make(chan bool, 1)
	waitGrp := new(sync.WaitGroup)
	waitGrp.Add(1)
	go func() {
		select {
		case <-noNeedWait:
		case <-waitCtx.Done():
			_ = c.Close()
		}
		waitGrp.Done()
	}()
	return noNeedWait, waitGrp
}
func TestRegisterWaitChan(t *testing.T) {
	c := &net.TCPConn{}
	waitCtx, cancel := context.WithCancel(context.Background())

	// do some thing1
	noNeedWait, waitGrp := registerWaitChan(waitCtx, c)
	close(noNeedWait)
	waitGrp.Wait() // must call wait here, wait

	// do some thing2
	noNeedWait, waitGrp = registerWaitChan(waitCtx, c)
	close(noNeedWait)
	waitGrp.Wait()

	// do end
	_ = c.Close()
	cancel()

}
