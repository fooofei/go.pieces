package main

// an Example from https://mzh.io
import (
    "fmt"
)

func b1(g1 chan int, quit chan  bool)  {
    for{
        select {
            case i:= <- g1:
                fmt.Printf("B func got %v\n", i)
            case <- quit:
                fmt.Println("B quit")
                return
        }
    }
}

func ExampleSyncGoroutineBad(){
    var g1 = make(chan int)
    var quit = make(chan bool)

    go b1(g1, quit)

    for i:=0; i<3;i+=1{
        g1 <- i
    }
    quit <- true // tell B exit, but not wait B exit
    // B quit will output as sometime somewhere
    fmt.Println("Main exit")
    //output:
    //B func got 0
    //B func got 1
    //B func got 2
    //Main exit

    // sometimes may got
    //B func got 0
    //B func got 1
    //B func got 2
    //B quit
    //Main exit

}



func b2(g2 chan int, quit chan chan bool){
    for{
        select {
        case i:= <- g2:
            fmt.Printf("B func got %v\n", i)
        case bExit:=<- quit:
            fmt.Println("B quit")
            bExit <- true
            return
        }
    }
}

func ExampleSyncGoroutineGood(){
    var g2 = make(chan int)
    var quit = make(chan chan bool)

    go b2(g2, quit)

    for i:=0; i<3;i+=1{
        g2 <- i
    }
    var waitBExit = make(chan bool)
    quit <- waitBExit // tell B exit, put an chan
    <- waitBExit // wait B exit
    fmt.Println("Main exit")
    //output:
    //B func got 0
    //B func got 1
    //B func got 2
    //B quit
    //Main exit
}

