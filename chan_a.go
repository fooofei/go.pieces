package main

import (
	"fmt"
	"time"
)

func foo() (a int){
	a = 3
	fmt.Print("aaa")
	return
}

func test_chan()  {
	var c chan int
	c = make(chan int)


	go func() {
		time.Sleep(2)
		c <- 1
	}()
	time.Sleep(1)
	<-c
}

func myproducer(queue chan<- int){
	for i:=0; i<10; i+= 1{
		queue <- i
	}
	queue<- -1
}

func myconsumer(queue <-chan int, finish chan<- int)  {
	var v int
	for{
		v = <-queue
		if v < 0{
			break
		}
		fmt.Printf("In consumer got value=%d\n", v)
	}
	finish<- -1
}

func test_producer_consumer(){
	queue := make(chan int, 5)
	finish := make(chan  int,1)

	go myproducer(queue)
	go myconsumer(queue, finish)

	fmt.Println("[+] Wait finish in test_producer_consumer")
	v:= <-finish
	fmt.Printf("[+] Finished, the chan got %d\n", v)
}


func main(){
	test_producer_consumer()
}
