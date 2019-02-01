package main

import (
    "io"
    "log"
    "net"
    "time"
)

func cltRoutine(stopCh chan bool) {
    raddr := "127.0.0.1:8869"
    cnn, err := net.Dial("tcp", raddr)
    if err != nil {
        panic(err)
    }
    if cnn == nil {
        return
    }
    defer func() {
        _ = cnn.Close()
    }()

    defer func() {
        stopCh <- true
    }()

    defer func() {
        log.Printf("dialer exit")
    }()

    rxCtnt := make([]byte, 5)

    n, err := io.ReadFull(cnn, rxCtnt)
    if err != nil {
        panic(err)
    }
    log.Printf("dialer rx (%s), goto sleep 10s\n", rxCtnt[:n])
    time.Sleep(time.Duration(10) * time.Second)
    log.Printf("dialer wake up, goto write\n")
    _, err = io.WriteString(cnn, "world")
    if err != nil {
        panic(err)
    }
    log.Printf("dialer sleeping for 20s")
    time.Sleep(time.Duration(20) * time.Second)

}

func ExampleSetReadDeadLineReadTwice() {
    /*
    http://blog.sina.com.cn/s/blog_9be3b8f10101lhiq.html

    1  SetReadDeadline(15 sec)
    2  接着一个 time take 10sec 的 Read ，在 Deadline 到达之前返回
    3  第二次接着又来一个 time take >=10sec <15sec 的 read
       这个read 会跟 1中第deadline 有影响吗

    有影响。第2个read受 deadline 影响, 在 time=5sec 时read超时

    这样的性质给我们的编程启发是， 如果我们使用了下面的编程模式
    conn.SetReadDeadline(10)
    conn.Read() // 假如是在 10s,即Deadline 内返回
    如果同时在该 Deadline 内发起第二次 Read，就会不符合预期

    因此我们必需更改为下面的编码模式
    conn.SetReadDeadline(10)
    conn.Read()
    conn.SetReadDeadline(time.Time{}) // 显式的取消 Deadline

    */
    stopCh := make(chan bool, 1)
    lsnAddr := "127.0.0.1:8869"
    lsner, err := net.Listen("tcp", lsnAddr)
    if err != nil {
        panic(err)
    }
    defer func() {
        _ = lsner.Close()
    }()
    log.Printf("accept work on %v", lsner.Addr())
    go cltRoutine(stopCh)
    cnn, err := lsner.Accept()
    if err != nil {
        panic(err)
    }
    defer func() {
        _ = cnn.Close()
    }()
    log.Printf("accept tx")
    _, err = io.WriteString(cnn, "hello")
    if err != nil {
        panic(err)
    }
    rxCtnt := make([]byte, 10)
    log.Printf("accept SetReadDeadLine 15s\n")
    t1 := time.Now()
    deltaTime := time.Duration(15) * time.Second
    err = cnn.SetReadDeadline(time.Now().Add(deltaTime))
    log.Printf("accept goto first read")
    n, err := cnn.Read(rxCtnt)
    if err != nil {
        panic(err)
    }
    t2 := time.Now()
    log.Printf("accept rx (%s) take %v(s)\n", rxCtnt[:n], t2.Sub(t1).Seconds())

    t1 = time.Now()
    log.Printf("accept go to second read\n")
    n, err = cnn.Read(rxCtnt)
    t2 = time.Now()
    log.Printf("second read take %v(sec) n=%v err=%v\n",
        t2.Sub(t1).Seconds(), n, err)

    log.Printf("wait routine stop\n")
    <-stopCh
    log.Printf("exit\n")
    // output:
}

func ExampleSetReadDeadLine() {
    /*
    1 SetReadDeadline 10sec
    2 紧接着一个 time take 20sec 的 read ,如果不做任何变更，这里会read因为超时而失败，
    3 假如在 time take 5sec 时，重新SetDeadline 30sec，了，那么能救回来吗
    能救回来 , 会 Read 成功
    */
    stopCh := make(chan bool, 1)
    lsnAddr := "127.0.0.1:8869"
    lsner, err := net.Listen("tcp", lsnAddr)
    if err != nil {
        panic(err)
    }
    defer func() {
        _ = lsner.Close()
    }()
    log.Printf("accept work on %v", lsner.Addr())
    go cltRoutine(stopCh)
    cnn, err := lsner.Accept()
    if err != nil {
        panic(err)
    }
    defer func() {
        _ = cnn.Close()
    }()
    log.Printf("accept tx")
    _, err = io.WriteString(cnn, "hello")
    if err != nil {
        panic(err)
    }
    rxCtnt := make([]byte, 10)
    log.Printf("accept SetReadDeadLine 5s\n")
    t1 := time.Now()
    deltaTime := time.Duration(5) * time.Second
    err = cnn.SetReadDeadline(time.Now().Add(deltaTime))
    log.Printf("accept goto first read")
    // 发动一个 goroutine 去重新设置 SetReadDeadline 补救回来
    go func() {
        time.Sleep(time.Duration(3) * time.Second)
        d2 := time.Duration(20) * time.Second
        err = cnn.SetReadDeadline(time.Now().Add(d2))
        log.Printf("accept re SetReadDeadline(20s) in routine err=%v", err)
    }()
    n, err := cnn.Read(rxCtnt)
    t2 := time.Now()
    log.Printf("accept rx (%s) take %v(s) n=%v err=%v\n",
        rxCtnt[:n], t2.Sub(t1).Seconds(), n, err)

    log.Printf("wait routine stop\n")
    <-stopCh
    log.Printf("exit\n")

}

func main() {
    //ExampleSetReadDeadLineReadTwice()
    ExampleSetReadDeadLine()
}
