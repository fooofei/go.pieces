package go_pieces

import (
    "fmt"
    "sync"
    "time"
)

// 1 告诉 sub routine close，需要1个 chan ，在main 中close chan 就可以了
// 2 等待多个 sub routine 结束，如果是死等，需要 sync.WaitGroup.Wait
//   如果是超时的等， 需要搭配一个 chan 个数为 routine的个数，routine结束的时候，
//   向chan 中写入数据，观察 chan 中个数够了，然后触发另一个chan，说够了

func produceForBasic(q chan<- int) {
    for i := 0; i < 10; i += 1 {
        q <- i + 1
    }
    q <- -1
}

func consumeForBasic(q <-chan int, finish chan<- int) {
    for {
        v := <-q
        if v == -1 {
            break
        }
        fmt.Printf("consume got value= %v\n", v)
    }
    close(finish)
}

func ExampleBasicChanNoSelect() {
    q := make(chan int, 5)
    finish := make(chan int)

    go produceForBasic(q)
    go consumeForBasic(q, finish)

    // 猜测这里输出顺序会不稳定 consumeForBasic 也在输出 与这里没有先后顺序
    fmt.Println("Wait for consume exit")
    v := <-finish
    fmt.Printf("got %v from consumer, main exit\n", v)
    //output:Wait for consume exit
    //consume got value= 1
    //consume got value= 2
    //consume got value= 3
    //consume got value= 4
    //consume got value= 5
    //consume got value= 6
    //consume got value= 7
    //consume got value= 8
    //consume got value= 9
    //consume got value= 10
    //got 0 from consumer, main exit
}

// ignore this output
func ExampleBasicChanSndRcvOrder() {

    q := make(chan string, 0) // no buffer
    count := 6

    // 无 buffer的chan，预期snd每次都阻塞在对方rcv之后才返回
    // 但是观察到的是
    // 第一次 send 之后就阻塞住了 recv 之后才得到返回
    // 下面的 send 就可以类似 buffer=1 了 也不是每次都是

    // 需要解释这种现象
    // 也有可能是printf 产生了先后顺序
    go func(count int, q chan<- string) {
        for i := 1; i <= count; i++ {
            fmt.Printf("goroutine snd msg in %v\n", i)
            q <- fmt.Sprintf("push msg %v", i)
            fmt.Printf("goroutine snd msg out %v\n", i)
        }
        fmt.Printf("sub routine exit\n")
    }(count, q)
    fmt.Printf("main sleep 1s\n")
    time.Sleep(time.Second * 1)
    for i := 1; i <= count; i++ {
        v := <-q
        fmt.Printf("main rcv msg %v\n", v)
        time.Sleep(time.Microsecond * 100)
    }
    fmt.Printf("main exit\n")
    //one possible output:
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

func ExampleChanAsSignal() {

    var done = make(chan bool)
    // 1个goroutine snd rcv通信
    go func() {
        time.Sleep(time.Second)
        done <- true
    }()
    var startTime = time.Now()
    fmt.Println("main wait done")
    <-done
    e := int(time.Since(startTime).Seconds())
    fmt.Printf("main done return, main exit cost %v sec\n", e)
    //output:
    //main wait done
    //main done return, main exit cost 1 sec
}

func ExampleChanRoutineClose() {
    var done = make(chan bool)
    // 1个goroutine close rcv 通信

    go func() {
        time.Sleep(time.Second)
        close(done)
    }()
    var startTime = time.Now()
    fmt.Println("main wait done")
    <-done
    e := int(time.Since(startTime).Seconds())
    fmt.Printf("main done return, main exit cost %v sec\n", e)
    //output:
    // main wait done
    //main done return, main exit cost 1 sec
}

func ExampleChanBroadcastRoutineClose() {
    var done = make(chan bool)
    // 一个chan给2个goroutine 通信
    wg := sync.WaitGroup{}

    wg.Add(1)
    go func() {
        <-done
        time.Sleep(time.Second)
        fmt.Println("goroutine1 exit")
        wg.Done()
    }()

    wg.Add(1)
    go func() {
        <-done
        time.Sleep(time.Second * 2)
        fmt.Println("goroutine2 exit")
        wg.Done()
    }()

    close(done)
    wg.Wait()
    fmt.Println("main exit")
    //output:
    //goroutine1 exit
    //goroutine2 exit
    //main exit

}

func ExampleSelectNilChan() {

    a := make(chan string, 1)
    a <- "1"
    a = nil

    // if no default, will infinit wait
    select {
    case v := <-a:
        // not output this
        fmt.Printf("got %v from chan\n", v)
    default:
    }
    fmt.Printf("main exit\n")
    //output:main exit
}

// ignore this output
func ExampleChanCloseTwice() {
    var ch = make(chan bool)

    fmt.Println("close chan 1")
    close(ch)
    fmt.Println("close chan 2")
    //close(ch)
    //fmt.Println("main exit")
    // output:
    // panic: close of closed channel [recovered]
    //	panic: close of closed channel
}

func ExampleCloseChanWhenFull() {
    q := make(chan int, 3)
    q <- 1
    q <- 2
    close(q)
    // we can get value from closed chan
    // we can use v,ok := <- q to detect chan close
    for i := 0; i < 5; i++ {
        v,ok := <-q
        fmt.Printf("get %v from q ok= %v\n", v, ok)
    }

    //output:get 1 from q ok= true
    //get 2 from q ok= true
    //get 0 from q ok= false
    //get 0 from q ok= false
    //get 0 from q ok= false
}

func produceWithSleep(q chan string) {

    for i := 0; i < 5; i += 1 {
        m := i % 3
        m += 1
        time.Sleep(time.Second* time.Duration(m))
        q <- fmt.Sprintf("msg %v sleep for= %v", i, m)
    }
    close(q)
}

// ignore output
func ExampleChanNoblock1() {
    q := make(chan string)
    go produceWithSleep(q)

    startTime := time.Now()
    for {
        msg := ""
        ok := true
        select {
        case msg, ok = <-q:
        default:
        }
        if !ok {
            break
        }
        e := int(time.Since(startTime).Seconds())
        fmt.Printf("[%v]main get msg %v\n", e, msg)
        time.Sleep(time.Second)
    }
    fmt.Printf("main exit\n")
    //output:
}

func ExampleChanWithTimeout() {
    q := make(chan string)
    go produceWithSleep(q)

    startTime := time.Now()
    loop:
    for {
        msg := ""
        ok := true
        select {
        case msg, ok = <-q:
            if !ok {
                break loop
            }
        // time.After must recreate every loop
        case <-time.After(time.Second):
            msg="timeout"
        }
        e := int(time.Since(startTime).Seconds())
        fmt.Printf("[%v]main get msg %v\n", e, msg)

    }
    fmt.Printf("main exit\n")
    // output:[1]main get msg timeout
    //[1]main get msg msg 0 sleep for= 1
    //[2]main get msg timeout
    //[3]main get msg msg 1 sleep for= 2
    //[4]main get msg timeout
    //[5]main get msg timeout
    //[6]main get msg msg 2 sleep for= 3
    //[7]main get msg timeout
    //[7]main get msg msg 3 sleep for= 1
    //[8]main get msg timeout
    //[9]main get msg timeout
    //[9]main get msg msg 4 sleep for= 2
}


func ExampleChanWithTick() {
    q := make(chan string)
    go produceWithSleep(q)

    startTime := time.Now()
    tkCh := time.Tick(time.Second)
loop:
    for {
        msg := ""
        ok := true
        select {
        case msg, ok = <-q:
            if !ok {
                break loop
            }
        case <-tkCh:
            msg = "timeout"
        }
        e := int(time.Since(startTime).Seconds())
        fmt.Printf("[%v]main get msg %v\n", e, msg)

    }
    fmt.Printf("main exit\n")
    //output:
}

func ExampleChanRange(){
    // range auto detect chan closed if or not

    q := make(chan string)

    go produceWithSleep(q)

    st := time.Now()

    for v := range q {
        e := int(time.Since(st).Seconds())
        fmt.Printf("[%v] main got %v\n", e, v)
    }

    fmt.Printf("main exit\n")
    // output:[1] main got msg 0 sleep for= 1
    //[3] main got msg 1 sleep for= 2
    //[6] main got msg 2 sleep for= 3
    //[7] main got msg 3 sleep for= 1
    //[9] main got msg 4 sleep for= 2
}