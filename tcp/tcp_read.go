package main

import (
    "log"
    "net"
    "reflect"
    "time"
)

func tcpReadTimeout() {
    /**
    make a tcp server
    ncat -l -t -k -v 8869
    not write anything

    error https://liudanking.com/network/go-%E4%B8%AD%E5%A6%82%E4%BD%95%E5%87%86%E7%A1%AE%E5%9C%B0%E5%88%A4%E6%96%AD%E5%92%8C%E8%AF%86%E5%88%AB%E5%90%84%E7%A7%8D%E7%BD%91%E7%BB%9C%E9%94%99%E8%AF%AF/

    2、读一个关闭的tcp
    */

    crlf := []byte("\r\n")
    log.Printf("crlf=%v", crlf)

    raddr := "127.0.0.1:8869"
    cnn, err := net.DialTimeout("tcp", raddr, time.Duration(3)*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    if cnn == nil {
        return
    }
    defer func() {
        _ = cnn.Close()
    }()
    buf := make([]byte, 8*1024)

    dta := time.Duration(3) * time.Second
    _ = cnn.SetReadDeadline(time.Now().Add(dta))
    n, err := cnn.Read(buf)
    _ = cnn.SetReadDeadline(time.Time{})

    opErr,ok := err.(*net.OpError)
    if ok {
        log.Printf("err = opErr and Timeout()=%v", opErr.Timeout())
        log.Printf("err = opErr and Temporary()=%v", opErr.Temporary())
    }
    log.Printf("rx len=%v %v %s err=%v %v",
        n, buf[:n], buf[:n], reflect.TypeOf(err), err)

}

func tcpCloseWhileReading(){
    /**
    make a tcp server
    ncat -l -t -k -v 8869

    2、读一个关闭的tcp (对方一点没发和发送了一点)

    看现象是对方发送过，即使对方close tcp 连接，我们还可以读取，err=nil

     */
    crlf := []byte("\r\n")
    log.Printf("crlf=%v", crlf)

    raddr := "127.0.0.1:8869"
    cnn, err := net.DialTimeout("tcp", raddr, time.Duration(3)*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    if cnn == nil {
        return
    }
    defer func() {
        _ = cnn.Close()
    }()
    buf := make([]byte, 8*1024)

    log.Printf("enter sleep 7s")
    dta := time.Duration(7) * time.Second
    time.Sleep(dta)
    log.Printf("leave sleep")
    // close the tcp server and
    // goto read a closed tcp
    // watch read err
    n, err := cnn.Read(buf)
    opErr,ok := err.(*net.OpError)
    if ok {
        log.Printf("err = opErr and Timeout()=%v", opErr.Timeout())
        log.Printf("err = opErr and Temporary()=%v", opErr.Temporary())
    }
    log.Printf("rx len=%v %v %s err=%v %v",
        n, buf[:n], buf[:n], reflect.TypeOf(err), err)

}

func tcpReadingOtherClose(){
    /**
    make a tcp server
    ncat -l -t -k -v 8869

     正在读，其他routine 关闭了这个 conn

    */
    raddr := "127.0.0.1:8869"
    cnn, err := net.DialTimeout("tcp", raddr, time.Duration(3)*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    if cnn == nil {
        return
    }
    defer func() {
        _ = cnn.Close()
    }()

    go func() {
        log.Printf("other routine enter sleep 5s")
        time.Sleep(time.Duration(5)*time.Second)
        log.Printf("other routine leave sleep")
        log.Printf("other routine close the cnn")
        _ = cnn.Close()
    }()
    buf := make([]byte, 8*1024)
    log.Printf("main routine block on the read")
    n, err := cnn.Read(buf)

    opErr,ok := err.(*net.OpError)
    if ok {
        log.Printf("err = opErr and Timeout()=%v", opErr.Timeout())
        log.Printf("err = opErr and Temporary()=%v", opErr.Temporary())
    }
    log.Printf("rx len=%v %v %s err=%v %v",
        n, buf[:n], buf[:n], reflect.TypeOf(err), err)

}

func tcpBeforeReadClose(){
    /**
    make a tcp server
    ncat -l -t -k -v 8869

    即将读，其他routine 关闭了这个 conn

    */
    raddr := "127.0.0.1:8869"
    cnn, err := net.DialTimeout("tcp", raddr, time.Duration(3)*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    if cnn == nil {
        return
    }
    defer func() {
        _ = cnn.Close()
    }()

    go func() {
        log.Printf("other routine close the cnn")
        _ = cnn.Close()
    }()
    time.Sleep(time.Second)
    buf := make([]byte, 8*1024)
    log.Printf("main routine block on the read")
    n, err := cnn.Read(buf)
    opErr,ok := err.(*net.OpError)
    if ok {
        log.Printf("err = opErr and Timeout()=%v", opErr.Timeout())
        log.Printf("err = opErr and Temporary()=%v", opErr.Temporary())
    }
    log.Printf("rx len=%v %v %s err=%v %v",
        n, buf[:n], buf[:n], reflect.TypeOf(err), err)

    err = cnn.Close()
    log.Printf("multi close err=%v", err)
    // err=close tcp 127.0.0.1:58719->127.0.0.1:8869: use of closed network connection
}

func main(){
    // tcpReadTimeout()
    // tcpCloseWhileReading()
    // tcpReadingOtherClose()
    tcpBeforeReadClose()
}
