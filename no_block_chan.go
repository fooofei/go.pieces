package main

import (
	"fmt"
	"time"
)

func produce1(messages chan<- string)  {

	for i:=0;i<5;i+=1{
		messages<- fmt.Sprintf("%d",i)
		m := i%3
		time.Sleep(time.Second *time.Duration(m))
	}
	close(messages)
}

// 这里对参数如何做到也能传递 int 呢
// 没有办法做范型
func readNoBlockChan(ch <-chan string)(result interface{},closed bool) {
	select {
		case r,ok:= <-ch:
			result=r
			closed = !ok
	default:
		result=nil
		closed=false
	}
	return
}

func testNoBlock1()  {

	messages := make(chan string)
	go produce1(messages)

	// 实现非阻塞的 chan
	for {
		msg,closed:= readNoBlockChan(messages)
		if closed{
			break
		}
		if msg != nil{
			fmt.Printf("We got msg %s\n",msg)
		}else{
			fmt.Printf("No msg\n")
		}
		time.Sleep(time.Millisecond*500)
	}

}

func main3() {
	testNoBlock1()
}
