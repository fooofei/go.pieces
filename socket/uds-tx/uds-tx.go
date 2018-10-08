package main

import (
    "bytes"
    "fmt"
    "net"
    "os"
    "time"
)

func main(){
    sockType := "unixgram"
    pid:=os.Getpid()
    localAddr,_ := net.ResolveUnixAddr(sockType, fmt.Sprintf("/tmp/%d-tx.sock",pid))
    remoteAddr,_ := net.ResolveUnixAddr(sockType, fmt.Sprintf("/tmp/dpdk-rx.sock"))
    os.Remove(localAddr.String())
    //conn,err := net.DialUnix(sockType,localAddr,remoteAddr)
    conn,err := net.DialUnix(sockType,localAddr,nil)
    fmt.Printf("listen %v addr=%v err=%v\n", conn, localAddr, err)
    if err != nil{
        panic(err)
    }
    defer os.Remove(localAddr.String())
    txBuf := bytes.NewBuffer(make([]byte,0, 1024*2))
    cnt := 0

    for{
        txBuf.Reset()
        txBuf.WriteString(fmt.Sprintf("msg %v", cnt))
        rx0,err := conn.WriteToUnix(txBuf.Bytes(),remoteAddr)
        cnt += 1
        fmt.Printf("rx from %v %v len(rxBuf)=%v err=%v\n",remoteAddr, rx0, len(txBuf.Bytes()), err)
        time.Sleep(time.Duration(time.Second*3))
    }

    fmt.Printf("uds tx exit\n")
}
