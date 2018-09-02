package main

import (
    "fmt"
    "sync"
    "time"
)

func produceForBasic(q chan <-int){
    for i:=0;i<10;i+=1{
        q <- i+1
    }
    q <- -1
}

func consumeForBasic(q <-chan int, finish chan <- int){
    var v int
    for {
        v = <- q
        if v==-1{
            break
        }
        fmt.Printf("%v got value=%v\n",GetFuncName(consumeForBasic),v)
    }
    finish<- -1
}

func ExampleBasicChan(){
    q := make(chan int,5)
    finish := make(chan  int,1)

    go produceForBasic(q)
    go consumeForBasic(q,finish)

    // 猜测这里输出顺序会不稳定 consumeForBasic 也在输出 与这里没有先后顺序
    fmt.Println("Wait for consume exit")
    v := <- finish
    fmt.Printf("got %v from consumer, main exit\n",v)
    //output:
    // Wait for consume exit
    //consumeForBasic got value=1
    //consumeForBasic got value=2
    //consumeForBasic got value=3
    //consumeForBasic got value=4
    //consumeForBasic got value=5
    //consumeForBasic got value=6
    //consumeForBasic got value=7
    //consumeForBasic got value=8
    //consumeForBasic got value=9
    //consumeForBasic got value=10
    //got -1 from consumer, main exit
}


func ExampleBasicChan2(){

    msgQ := make(chan string,0) // no buffer
    count := 6

    // 第一次 send 之后就阻塞住了 recv 之后才得到返回
    // 下面的 send 就可以类似 buffer=1 了 也不是每次都是

    // 需要解释这种现象
    // 也有可能是printf 产生了先后顺序
    go func() {
        for i := 1; i <= count; i++ {
            fmt.Printf("goroutine snd msg in %v\n",i)
            msgQ <- fmt.Sprintf("push msg %v", i)
            fmt.Printf("goroutine snd msg out %v\n",i)
        }
    }()
    fmt.Println("main sleep")
    time.Sleep(time.Second * 3)
    for i := 1; i <= count; i++ {
        var v = <- msgQ
        fmt.Printf("main rcv msg %v\n", v)
        time.Sleep(time.Second)
    }
    fmt.Println("main exit")
    // 一个可能的输出
    //output:
    // main sleep
    //goroutine snd msg in 1
    //main rcv msg push msg 1
    //goroutine snd msg out 1
    //goroutine snd msg in 2
    //goroutine snd msg out 2
    //main rcv msg push msg 2
    //goroutine snd msg in 3
    //goroutine snd msg out 3
    //goroutine snd msg in 4
    //main rcv msg push msg 3
    //main rcv msg push msg 4
    //goroutine snd msg out 4
    //goroutine snd msg in 5
    //main rcv msg push msg 5
    //goroutine snd msg out 5
    //goroutine snd msg in 6
    //goroutine snd msg out 6
    //main rcv msg push msg 6
    //main exit
}

func ExampleChanAsSignal(){

    var done = make(chan bool)
    // 1个goroutine snd rcv通信
    go func() {
        time.Sleep(time.Second)
        done <- true
    }()
    var startTime = time.Now()
    fmt.Println("main wait done")
    <- done
    var nowTime = time.Now()
    var elapseSeconds = nowTime.Sub(startTime).Seconds()
    var e = int(elapseSeconds)
    fmt.Printf("main done return, main exit cost %v sec\n",e)
    //output:
    //main wait done
    //main done return, main exit cost 1 sec
}

func ExampleChanAsSignal2(){
    var done = make(chan bool)
    // 1个goroutine close rcv 通信

    go func() {
        time.Sleep(time.Second)
        close(done)
    }()
    var startTime = time.Now()
    fmt.Println("main wait done")
    <- done
    var nowTime = time.Now()
    var elapseSeconds = nowTime.Sub(startTime).Seconds()
    var e = int(elapseSeconds)
    fmt.Printf("main done return, main exit cost %v sec\n",e)
    //output:
    // main wait done
    //main done return, main exit cost 1 sec
}


func ExampleChanAsSignal3(){
    var done = make(chan bool)
    // 一个chan给2个goroutine 通信

    go func() {
        <- done
        fmt.Println("goroutine1 exit")
    }()

    go func() {
        <- done
        fmt.Println("goroutine2 exit")
    }()

    close(done)
    time.Sleep(500*time.Millisecond)
    fmt.Println("main exit")
    //output:
    //goroutine1 exit
    //goroutine2 exit
    //main exit

    //maybe
    //goroutine2 exit
    //goroutine1 exit
    //main exit

}

func ExampleChanAsSignal4(){
    var done = make(chan chan bool)
    //1个goroutine 用嵌套 chan
    go func() {
        quit:=<- done
        fmt.Println("goroutine1 exit")
        quit <- true
    }()

    quit := make(chan bool)
    fmt.Println("main push quit to chan")
    done <- quit
    fmt.Println("main wait goroutine exit") //这句执行跟goroutine printf有竞争
    <- quit
    fmt.Println("main goroutine exit, main exit")
   // output:
   // main push quit to chan
    //goroutine1 exit
    //main wait goroutine exit
    //main goroutine exit, main exit
}

func ExampleChanAsSignal5(){
    var done = make(chan bool)
    var wg = sync.WaitGroup{}
    // 通知2个goroutine 退出 并等待2个goroutine完全退出

    wg.Add(1)
    go func() {
        <- done
        fmt.Println("goroutine1 exit")
        wg.Done()
    }()
    wg.Add(1)
    go func() {
        <- done
        fmt.Println("goroutine2 exit")
        wg.Done()
    }()

    fmt.Println("main snd close to goroutine")
    close(done)
    fmt.Println("main wait goroutine exit")
    wg.Wait()
    fmt.Println("main exit")
    //output:
    //main snd close to goroutine
    //main wait goroutine exit
    //goroutine2 exit
    //goroutine1 exit
    //main exit
}