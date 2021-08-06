// +build !windows

package netonn

import (
	"fmt"
	"log"
	"net"
	"time"
)

func dialRoutine(addr string, wait chan bool) {

	prefix := "[2]"
	cnn, err := net.Dial("tcp", addr)
	assertErrNil(prefix, err)

	for i := 0; i < 2; i++ {
		sleep := time.Second * 3
		log.Printf("%v dialRoutine enter sleep %v s", prefix, int64(sleep.Seconds()))
		time.Sleep(sleep)
		log.Printf("%v dialRoutine leave sleep %v s", prefix, int64(sleep.Seconds()))

		_, ew := fmt.Fprintf(cnn, "hello")
		log.Printf("%v write ew= %v", prefix, ew)
	}

	_ = cnn.Close()
	close(wait)
}

func TestSetReadDeadlineWhenRead() {
	// [1] SetReadDeadline 2 s
	// [1] Read()
	// [2] after 3s write
	// [1] Read() timeout

	// 这样补救
	// [1] SetReadDeadline 2s
	// [1] Read()
	// [1]' SetReadDeadline more
	// [2] after 3s write
	// [1] Read() something

	prefix := "[1]"
	addr := "127.0.0.1:3389"
	lsnCnn, err := net.Listen("tcp", addr)
	assertErrNil(prefix, err)
	log.Printf("%v listen on %v", prefix, addr)

	wait := make(chan bool)
	log.Printf("%v call dialRoutine", prefix)
	go dialRoutine(addr, wait)

	cnn, err := lsnCnn.Accept()
	assertErrNil(prefix, err)

	buf := make([]byte, 10)
	timeOff := time.Second * 2
	log.Printf("%v SetReadDeadline %v s", prefix, int64(timeOff.Seconds()))
	_ = cnn.SetReadDeadline(time.Now().Add(timeOff))
	timeOff = time.Second * 5

	// 可以继续设置 补救回来 ，之后就能读到内容了
	//go func() {
	//	time.Sleep(time.Second)
	//	log.Printf("%v SetReadDeadline again ", prefix)
	//	_ = cnn.SetReadDeadline(time.Now().Add(timeOff))
	//}()
	log.Printf("%v enter read", prefix)
	nr, er := cnn.Read(buf)
	log.Printf("%v leave read nr= %v er= %v", prefix, nr, er)
	if nr > 0 {
		log.Printf("%v GOOD we read something", prefix)
	} else {
		log.Printf("%v BAD we read nothing", prefix)
	}
	_ = cnn.Close()
	<-wait
	_ = lsnCnn.Close()
}

func TestSetReadDeadlineMustClear() {
	/*
	   http://blog.sina.com.cn/s/blog_9be3b8f10101lhiq.html

	   1  SetReadDeadline(15 sec)
	   2  接着一个 time take 10sec 的 Read ，在 Deadline 到达之前返回
	   3  第二次接着又来一个 time take >=10sec <15sec 的 read
	      这个read 会跟 1中第deadline 有影响吗

	   有影响。第2个read受 deadline 影响, 在 time=5sec 时read超时

	   这样的性质给我们的编程启发是， 如果我们使用了下面的编程模式
	   conn.SetReadDeadline(10)
	   conn.Read() // 假如是在 10s,即Deadline 内返回
	   如果同时在该 Deadline 内发起第二次 Read，就会不符合预期

	   因此我们必需更改为下面的编码模式
	   conn.SetReadDeadline(10)
	   conn.Read()
	   conn.SetReadDeadline(time.Time{}) // 显式的取消 Deadline

	*/

	prefix := "[1]"
	addr := "127.0.0.1:3399"
	lsnCnn, err := net.Listen("tcp", addr)
	assertErrNil(prefix, err)
	log.Printf("%v listen on %v", prefix, addr)

	wait := make(chan bool)
	log.Printf("%v call dialRoutine", prefix)
	go dialRoutine(addr, wait)

	cnn, err := lsnCnn.Accept()
	assertErrNil(prefix, err)

	// first Read() SetReadDeadline()
	// second Read() not set
	buf := make([]byte, 10)
	timeOff := time.Second * 4
	log.Printf("%v SetReadDeadline %v s", prefix, int64(timeOff.Seconds()))
	_ = cnn.SetReadDeadline(time.Now().Add(timeOff))
	log.Printf("%v enter read", prefix)
	nr, er := cnn.Read(buf)
	log.Printf("%v leave read nr= %v er= %v", prefix, nr, er)
	if nr > 0 {
		log.Printf("%v GOOD we read something", prefix)
	} else {
		log.Printf("%v BAD we read nothing", prefix)
	}
	// MUST clear
	// = cnn.SetReadDeadline(time.Time{})

	log.Printf("%v enter read", prefix)
	nr, er = cnn.Read(buf)
	log.Printf("%v leave read nr= %v er= %v", prefix, nr, er)
	if nr > 0 {
		log.Printf("%v GOOD we read something", prefix)
	} else {
		log.Printf("%v BAD we read nothing", prefix)
	}

	_ = cnn.Close()
	<-wait
	_ = lsnCnn.Close()

}

func main() {
	TestSetReadDeadlineWhenRead()
	TestSetReadDeadlineMustClear()
}
