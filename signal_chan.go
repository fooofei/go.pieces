package main

import "fmt"

func test_signal1()  {
	done := make(chan bool)

	go func() {
		fmt.Println("goroutine message")

		// We are only interested in the fact of sending itself,
		// but not in data being sent.
		done <- true
	}()

	fmt.Println("main function message")
	<-done
}

func test_signal2()  {
	done := make(chan struct {})

	go func() {
		fmt.Println("goroutine message")
		close(done)
	}()

	fmt.Println("main function message")
	<-done

}


func main1() {
	test_signal2()
}