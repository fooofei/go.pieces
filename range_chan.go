package main

import (
	"fmt"
)

func produce(message chan<- string)  {
	for i := 0; i < 5; i++ {
		message <- fmt.Sprintf("message %d", i)
	}
	close(message)  // must have this close
}

func test_range_chan()  {
	message := make(chan string)
	go produce(message)

	for msg := range message {
		fmt.Println(msg)
	}
}

func main()  {
	test_range_chan()
}
