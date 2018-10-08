package main

import (
    "fmt"
    "net"
    "os"
)

func main(){
    sockType := "unixgram"
    localAddr,_ := net.ResolveUnixAddr(sockType, fmt.Sprintf("/tmp/dpdk-rx.sock"))
    os.Remove(localAddr.String())
    conn,err :=net.ListenUnixgram(sockType, localAddr)
    //conn0,err :=net.Listen("unixgram", localAddrS)
    fmt.Printf("listen %v addr=%v err=%v\n", conn, localAddr, err)
    if err != nil{
        panic(err)
    }
    defer os.Remove(localAddr.String())
    rxBuf := make([]byte, 1024*2)
    cnt := 0


    for{
        rx0, remoteAddr, err := conn.ReadFromUnix(rxBuf)
        cnt += 1
        fmt.Printf("%v: [len=%v return=%v]%v, err=%v\n",
            remoteAddr, len(rxBuf),rx0, rxBuf,err)
    }





    fmt.Printf("rx end\n")
}
