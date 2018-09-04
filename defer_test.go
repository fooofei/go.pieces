package main

import "fmt"

func ExampleDefer1(){
    var i int =1
    defer fmt.Printf("1 i=%v\n",i)
    defer fmt.Printf("2 i=%v\n",i+1)
    defer fmt.Printf("3 i=%v\n",i+2)
    i+=1
    fmt.Printf("main exit i=%v\n",i)
    //output:
    //main exit i=2
    //3 i=3
    //2 i=2
    //1 i=1
}

func ExampleDefer2(){
    var i = 1
    defer fmt.Println("result: ", func() int { return i * 2 }())
    i++
    //output:result:  2
}

func ExampleDefer3(){
    var whatever [5]struct{}

    for i := range whatever {
        fmt.Println(i)
    }

    // confused with example1
    for i := range whatever {
        defer func() { fmt.Printf("1 %v\n",i) }()
    }

    for i := range whatever {
        defer func(n int) { fmt.Println(n) }(i)
    }
    //output:
    // 0
    //1
    //2
    //3
    //4
    //4
    //3
    //2
    //1
    //0
    //1 4
    //1 4
    //1 4
    //1 4
    //1 4
}