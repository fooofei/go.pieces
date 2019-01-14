package go_pieces

import (
    "crypto/tls"
    "fmt"
    "io"
    "net"
    "time"
)

func ExampleNetConnNilTest() {

    var cnn io.ReadWriteCloser = nil
    var err error

    fmt.Printf("cnn= %v  cnn==nil =%v\n", cnn, cnn==nil)

    d := net.Dialer{Timeout: time.Duration(3) * time.Second}

    //cnn,err = d.Dial("tcp", "127.0.0.1:8899")
    tlsCfg := tls.Config{InsecureSkipVerify: true}
    cnn1, err := tls.DialWithDialer(&d, "tcp", "127.0.0.1:8899",
        &tlsCfg)

    fmt.Printf("cnn1= %v cnn1==nil =%v\n", cnn1, cnn1==nil)

    cnn = cnn1
    fmt.Printf("err= %v\n", err)
    fmt.Printf("cnn=%v  cnn==nil =%v cnn==(*tls.Conn)(nil) =%v\n",
        cnn, cnn == nil, cnn==(*tls.Conn)(nil))


    if cnn == nil {
        fmt.Printf("cnn == nil, we return in case of panic\n")
        return
    }

    if cnn == (*tls.Conn)(nil) {
        fmt.Printf("cnn == (*tls.Conn)(nil) , we return in case of panic\n")
        return
    }

    _ = cnn.Close()
    //output:cnn= <nil>  cnn==nil =true
    //cnn1= <nil> cnn1==nil =true
    //err= dial tcp 127.0.0.1:8899: connect: connection refused
    //cnn=<nil>  cnn==nil =false cnn==(*tls.Conn)(nil) =true
    //cnn == (*tls.Conn)(nil) , we return in case of panic

}
