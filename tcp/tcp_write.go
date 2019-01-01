
package main

import (
    "fmt"
    "log"
    "net"
    "reflect"
    "time"
)

func rxRoutine(stopChan chan bool, setupCh chan bool){
    laddr := "127.0.0.1:8869"

    listener, err := net.Listen("tcp", laddr)
    if err != nil{
        log.Fatal(fmt.Sprintf("listen err =%v", err))
    }

    defer func(){
        _ = listener.Close()
    }()
    setupCh <- true
    cnn, err := listener.Accept()
    if err != nil{
        log.Fatal(fmt.Sprintf("accept err=%v", err))
    }
    if cnn == nil{
        log.Fatal("accept cnn=nil")
    }
    defer func() {
        _ = cnn.Close()
        log.Printf("rx routine cnn closed")
    }()

    defer log.Printf("exit rx routine")

    log.Printf("rx routine laddr=%v raddr=%v",
        cnn.LocalAddr(), cnn.RemoteAddr())
    log.Printf("rx routine enter wait stopChan")

    <- stopChan

}


func tcpWriteTimeout(){
    /**
    1、make a tcp write timeout
    2、这个例子还遇到了 write 一部分数据
    Temporary() error 怎么理解

    */
    // setup listener first
    stopCh := make(chan bool,1)
    setupCh := make(chan bool,1)
    go rxRoutine(stopCh, setupCh)
    defer func() {
        stopCh <- true
    }()
    //
    log.Printf("wait setup listerner")
    <- setupCh
    log.Printf("listener already setuped")
    raddr := "127.0.0.1:8869"
    cnn,err := net.DialTimeout("tcp", raddr,
        time.Duration(3)*time.Second)
    if err != nil{
        log.Fatal(fmt.Sprintf("dial err=%v",err))
    }
    if cnn == nil{
        log.Fatal("dial cnn=nil")
    }
    defer func() {
        _ = cnn.Close()
    }()

    time.Sleep(time.Duration(2)*time.Second)
    buf := make([]byte, 2*1024*1024)

    for {
        dta := time.Duration(2)*time.Second
        err := cnn.SetWriteDeadline(time.Now().Add(dta))
        if err != nil{
            log.Fatal(err)
        }
        log.Printf("enter write")
        n,err := cnn.Write(buf)
        log.Printf("leave write")
        log.Printf("write n=%v err=%v", n, err)
        opErr,ok := err.(*net.OpError)
        if ok{
            log.Printf("err = opErr and Timeout()=%v", opErr.Timeout())
            log.Printf("err = opErr and Temporary()=%v", opErr.Temporary())
            log.Printf("opErr = %v", opErr.Error())
        }
        err = cnn.SetWriteDeadline(time.Time{})
        if err != nil{
            log.Fatal(err)
        }
        time.Sleep(time.Duration(6)*time.Second)
    }


    stopCh <- true

}

func tcpWriteCloseCnn(){
    /**
        1、写一个关闭的 cnn
        可能有两种错误
    */
    // setup listener first
    stopCh := make(chan bool,1)
    setupCh := make(chan bool,1)
    go rxRoutine(stopCh, setupCh)
    defer func() {
        stopCh <- true
    }()
    //
    log.Printf("wait setup listerner")
    <- setupCh
    log.Printf("listener already setuped")
    raddr := "127.0.0.1:8869"
    cnn,err := net.DialTimeout("tcp", raddr,
        time.Duration(3)*time.Second)
    if err != nil{
        log.Fatal(fmt.Sprintf("dial err=%v",err))
    }
    if cnn == nil{
        log.Fatal("dial cnn=nil")
    }
    defer func() {
        _ = cnn.Close()
    }()

    time.Sleep(time.Duration(2)*time.Second)
    log.Printf("close the remote cnn")
    stopCh <- true
    time.Sleep(time.Duration(2)*time.Second)
    log.Printf("then write the cnn")
    cnt := 0
    for {
        //buf := make([]byte, 2)
        //buf := make([]byte, 512*1024)
        // if longth
        buf := make([]byte, 1024*1024)
        // err=*net.OpError write tcp 127.0.0.1:57399->127.0.0.1:8869: write: protocol wrong type for socket

        n,err := cnn.Write(buf)
        log.Printf("write n=%v err=%v %v", n, reflect.TypeOf(err), err)
        opErr,ok := err.(*net.OpError)
        // err=*net.OpError write tcp 127.0.0.1:57383->127.0.0.1:8869: write: broken pipe

        if ok {
            log.Printf("err = opErr and Timeout()=%v", opErr.Timeout())
            log.Printf("err = opErr and Temporary()=%v", opErr.Temporary())
            log.Printf("opErr = %v", opErr.Error())
        }

        if err != nil{
            break
        }
        cnt += 1
    }
    log.Printf("exit at cnt=%v", cnt)
}


func main(){
    // tcpWriteTimeout()
    tcpWriteCloseCnn()
}
