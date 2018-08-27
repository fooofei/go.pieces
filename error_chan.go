package main

import (
	"fmt"
	"time"
)

func foo1()  {
	message := make(chan string) // no buffer
	count := 6

	go func() {
		for i := 1; i <= count; i++ {
			fmt.Println("send message")
			message <- fmt.Sprintf("message %d", i)
		}
	}()

	time.Sleep(time.Second * 3)

	for i := 1; i <= count; i++ {
		fmt.Println(<-message)
	}
}


func main9()  {
	foo1()
}
