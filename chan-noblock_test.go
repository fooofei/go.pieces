package main

import (
    "fmt"
    "time"
)

func produceForNoblock(msgQ chan<- string)  {

	for i:=0;i<5;i+=1{
		msgQ <- fmt.Sprintf("push msg %d",i)
		m := i%3
		time.Sleep(time.Second *time.Duration(m))
	}
	close(msgQ)
}

func rcvChanNoblock(ch <- chan string) ( interface{}, bool){
	select {
	case r,ok := <-ch:
		closed := !ok
		return r,closed
	default:
		return nil,false
	}
}

func ExampleChanNoblock1(){
    var msgQ = make(chan string)
    go produceForNoblock(msgQ)

    //每隔1s去channel取内容
    var startTime = time.Now()
    for{
        msg,closed := rcvChanNoblock(msgQ)
        if closed{
            break
        }
        var nowTime = time.Now()
        var elapseSeconds = nowTime.Sub(startTime).Seconds()
        var elapse = int(elapseSeconds)
        fmt.Printf("[%v]main get msg %v\n",elapse,msg)
        time.Sleep(time.Second)
    }
    fmt.Printf("main exit")
    //output:
    // [0]main get msg <nil>
    //[1]main get msg push msg 0
    //[2]main get msg push msg 1
    //[3]main get msg push msg 2
    //[4]main get msg <nil>
    //[5]main get msg push msg 3
    //[6]main get msg push msg 4
    //[7]main get msg <nil>
    //main exit

    // maybe output:
    // [0]main get msg <nil>
    //[1]main get msg push msg 0
    //[2]main get msg push msg 1
    //[3]main get msg <nil>
    //[4]main get msg push msg 2
    //[5]main get msg <nil>
    //[6]main get msg <nil>
    //[7]main get msg push msg 3
    //[8]main get msg push msg 4
    //[9]main get msg <nil>
    //main exit
}


func ExampleChanNoblock2(){
    var msgQ = make(chan string)
    go produceForNoblock(msgQ)

    var startTime = time.Now()
    var timeout = time.After(time.Second)
    // 以不超过 1 s 的超时取内容
    for{
        var msg string
        var closed =false
        select {
            case v,ok:=<-msgQ:
                msg = v
                closed = !ok
            case <-timeout:
                msg = "timeout-msg"
        }
        if closed{
            break
        }
        var nowTime = time.Now()
        var elapseSeconds = nowTime.Sub(startTime).Seconds()
        var elapse = int(elapseSeconds)
        fmt.Printf("[%v]main get msg %v\n",elapse,msg)
    }
    fmt.Printf("main exit")
    //output:
    // [0]main get msg push msg 0
    //[0]main get msg push msg 1
    //[1]main get msg timeout-msg
    //[1]main get msg push msg 2
    //[3]main get msg push msg 3
    //[3]main get msg push msg 4
    //main exit

}


func ExampleChanTimeoutSelect(){
    var c = make(chan int)
    go func() {
        for i:=1;i<8;i+=1{
            time.Sleep(time.Second * time.Duration(i))
            c <- i
        }
        close(c)
    }()

    for{
        closed:=false
        select {
        case v,ok:=<-c:
            closed = !ok
            if closed{
                break
            }
            fmt.Printf("main got %v\n",v)
        case <-time.After(time.Second):
            fmt.Printf("main got timeout\n")
        }
        if closed{
            fmt.Printf("chan closed, main exit\n")
            break
        }
    }
    fmt.Printf("main exit\n")
    //output:
    //main got timeout
    //main got 1
    //main got timeout
    //main got 2
    //main got timeout
    //main got timeout
    //main got 3
    //main got timeout
    //main got timeout
    //main got timeout
    //main got 4
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got 5
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got 6
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got 7
    //chan closed, main exit
    //main exit

    // or
    //main got timeout
    //main got 1
    //main got timeout
    //main got timeout
    //main got 2
    //main got timeout
    //main got timeout
    //main got 3
    //main got timeout
    //main got timeout
    //main got timeout
    //main got 4
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got 5
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got 6
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got timeout
    //main got 7
    //chan closed, main exit
    //main exit
}


func ExampleChanTimeoutConversation() {
    var c= make(chan int)
    //  体会time.After 放在不同位置的区别
    var timeout = time.After(time.Second)
    // 只能使用1次
    go func() {
        for i := 1; i < 8; i += 1 {
            time.Sleep(time.Second * time.Duration(i))
            c <- i
        }
        close(c)
    }()

    for {
        closed := false
        select {
        case v, ok := <-c:
            closed = !ok
            if closed {
                break
            }
            fmt.Printf("main got %v\n", v)
        case <-timeout:
            fmt.Printf("main got timeout\n")
        }
        if closed {
            fmt.Printf("chan closed, main exit\n")
            break
        }
    }
    fmt.Printf("main exit\n")
    //output:
    //main got timeout
    //main got 1
    //main got 2
    //main got 3
    //main got 4
    //main got 5
    //main got 6
    //main got 7
    //chan closed, main exit
    //main exit

}