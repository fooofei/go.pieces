package main

import (
    "fmt"
    "net"
    "os"
    "time"
)

type hitStat struct{
    hitTime time.Time
    rx uint64
    rxOk uint64
    rxFail uint64
}
func subTime(startTime time.Time, n time.Time) string{
    d := n.Sub(startTime)
    return fmt.Sprintf("%v %v(s)",n.Format("2006-01-02 15:04:05"),
        d.Seconds())
}

func main(){
    sockType := "unixgram"
    localAddr,_ := net.ResolveUnixAddr(sockType, fmt.Sprintf("/tmp/dpdk-tx-1-0.sock"))
    os.Remove(localAddr.String())
    conn,err :=net.ListenUnixgram(sockType, localAddr)
    //conn0,err :=net.Listen("unixgram", localAddrS)
    fmt.Printf("listen %v addr=%v err=%v\n", conn, localAddr, err)
    if err != nil{
        panic(err)
    }
    defer os.Remove(localAddr.String())
    rxBuf := make([]byte, 1024*2)
    var hit hitStat
    var stat hitStat
    stat.hitTime=time.Now()
    hit.hitTime = stat.hitTime
    for{
        nbRx, remoteAddr, err := conn.ReadFromUnix(rxBuf)
        stat.rx ++
        if nbRx>0{
            stat.rxOk ++
        }else{
            stat.rxFail ++
        }
        //rxBufS := fmt.Sprintf("%s", rxBuf)
        //fmt.Printf("cnt=%v %v: [len=%v return=%v]%v, err=%v\n",
        //    cnt, remoteAddr, len(rxBuf),rx0, rxBufS,err)
        //time.Sleep(time.Duration(time.Second))
        _ = remoteAddr
        _ = err
        if stat.rx % 1000 ==0{
            now:= time.Now()
            d := now.Sub(hit.hitTime)
            secs := uint64(d.Seconds())
            if secs>0{
                var p hitStat
                p.rx = (stat.rx-hit.rx)/secs
                p.rxOk = (stat.rxOk - hit.rxOk)/secs
                p.rxFail = (stat.rxFail - hit.rxFail)/secs
                hit = stat
                hit.hitTime = now
                fmt.Printf("%v %v(s) rx=%v rxok=%v rxfail=%v\n",subTime(stat.hitTime,now),secs,
                    p.rx,p.rxOk,p.rxFail)
            }

        }

    }

    fmt.Printf("rx end\n")
}
