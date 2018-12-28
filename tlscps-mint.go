package main

// failed for test TLSv1.3

import (
    "bytes"
    "encoding/binary"
    "encoding/hex"
    "fmt"
    "github.com/bifurcation/mint"
)

type VpnHead struct {
    mark uint32
    type_ uint8
    length uint16
}

func hexView( v interface{}) string {
    var binBuf bytes.Buffer
    binary.Write(&binBuf, binary.LittleEndian, v)

    //return hex.EncodeToString(binBuf.Bytes())
    return hex.Dump(binBuf.Bytes())
}

func main(){

    var raddr string

    raddr = "127.0.0.1:886"

    fmt.Printf("using TLSv1.3 to tcp %v\n" , raddr)
    conn, err := mint.Dial("tcp", raddr, nil)

    if err != nil {
        fmt.Printf("err = %v conn=%v \n", err, conn)
    }
    _,_ = conn,err

    fmt.Printf("exit\n")
}
