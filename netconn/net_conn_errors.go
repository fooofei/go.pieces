package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"time"
)

func assertErrNil(err error) {
	if err != nil {
		log.Fatalf("err= %v", err)
	}
}

func whichNetErr(err error) {
	// https://liudanking.com/network/go-%E4%B8%AD%E5%A6%82%E4%BD%95%E5%87%86%E7%A1%AE%E5%9C%B0%E5%88%A4%E6%96%AD%E5%92%8C%E8%AF%86%E5%88%AB%E5%90%84%E7%A7%8D%E7%BD%91%E7%BB%9C%E9%94%99%E8%AF%AF/
	prefix := "[whichNetErr]"

	if err == io.EOF {
		prefix += "[io.EOF]"
	}

	if netErr, is := err.(net.Error); is {
		prefix += "[net.Error]"
		if netErr.Timeout() {
			prefix += "[Timeout()=true]"
		}
		if netErr.Temporary() {
			prefix += "[Temporary()=true]"
		}

		if opErr, is := netErr.(*net.OpError); is {
			prefix += "[net.OpError]"

			switch t := opErr.Err.(type) {
			case *net.DNSError:
				prefix += "[net.DNSError]"
				if t.Temporary() {
					prefix += "[Temporary()=true]"
				}
				if t.Timeout() {
					prefix += "[Timeout()=true]"
				}
			case *os.SyscallError:
				prefix += "[os.SyscallError]"
				prefix += fmt.Sprintf("[os.SyscallError.Syscall=%v]", t.Syscall)
				if errno, is := t.Err.(syscall.Errno); is {
					prefix += "[syscall.Errno]"
					if errno.Temporary() {
						prefix += "[Temporary()=true]"
					}
					if errno.Timeout() {
						prefix += "[Timeout()=true]"
					}
					prefix += fmt.Sprintf("[errno=%v]", int64(errno))
					switch errno {
					case syscall.ECONNREFUSED:
						prefix += "[errno=syscall.ECONNREFUSED]"
					case syscall.ETIMEDOUT:
						prefix += "[errno=syscall.ETIMEDOUT]"
					case syscall.EPROTOTYPE:
						prefix += "[errno=syscall.EPROTOTYPE]"
					case syscall.EPIPE:
						prefix += "[errno=syscall.EPIPE]"
					}
				}

			default:
				prefix += fmt.Sprintf("[type=%T]", t)
			}
		}

	}
	log.Printf("%v err= %v", prefix, err)

}

func connConnectTimeoutError() {
	addr := "2.2.2.2:3390"
	dialTimeout := time.Second * 3
	cnn, err := net.DialTimeout("tcp", addr, dialTimeout)
	whichNetErr(err)

	// 2019/08/10 11:17:50 [whichNetErr][net.Error][Timeout()=true][Temporary()=true][net.OpError]
	// [type=*poll.TimeoutError] err= dial tcp 2.2.2.2:3390: i/o timeout
	_ = cnn
}

func connConnectReadError() {
	//  make a tcp server
	//	  ncat -l -t -k -v 3390
	//	  not write anything
	crlf := []byte("\r\n")
	log.Printf("crlf = %v", crlf)

	addr := "127.0.0.1:3390"
	dialTimeout := time.Second * 3
	cnn, err := net.DialTimeout("tcp", addr, dialTimeout)
	whichNetErr(err)

	//2019/08/10 11:10:04 [whichNetErr][net.Error][net.OpError][os.SyscallError][os.SyscallError.Syscall=connect]
	// [syscall.Errno][errno=61][errno=syscall.ECONNREFUSED]
	// err= dial tcp 127.0.0.1:3390: connect: connection refused

	assertErrNil(err)

	buf := make([]byte, 1024)

	go func() {
		time.Sleep(time.Second * 2)
		log.Printf("cnn closed by other routine")
		_ = cnn.Close()
	}()

	timeOff := time.Second * 6
	_ = cnn.SetReadDeadline(time.Now().Add(timeOff))
	_, er := cnn.Read(buf)
	_ = cnn.SetReadDeadline(time.Time{})
	whichNetErr(er)
	// 2019/08/10 11:21:02 [whichNetErr][net.Error][Timeout()=true][Temporary()=true]
	// [net.OpError]
	// [type=*poll.TimeoutError] err= read tcp 127.0.0.1:59294->127.0.0.1:3390: i/o timeout

	// 2019/08/10 11:21:36 [whichNetErr][io.EOF] err= EOF

	// 2019/08/10 11:22:13 [whichNetErr][net.Error][net.OpError][type=*errors.errorString]
	// err= read tcp 127.0.0.1:59309->127.0.0.1:3390: use of closed network connection

	// can close many times
	_ = cnn.Close()
	_ = cnn.Close()
}

func connWriteError() {
	// 正在写的时候 关闭 server， 就会有 2 个错误
	addr := "127.0.0.1:3390"
	dialTimeout := time.Second * 3
	cnn, err := net.DialTimeout("tcp", addr, dialTimeout)
	assertErrNil(err)

	for i := 0; i < 1000; i++ {
		buf := make([]byte, 1024*1024)
		nw, ew := cnn.Write(buf)
		whichNetErr(ew)
		log.Printf("nw= %v ew= %v", nw, ew)
		time.Sleep(time.Second)

	}

	//2019/08/10 11:07:08 [whichNetErr][net.Error][net.OpError][os.SyscallError][os.SyscallError.Syscall=write]
	// [syscall.Errno][errno=41][errno=syscall.EPROTOTYPE]
	// err= write tcp 127.0.0.1:59224->127.0.0.1:3390: write: protocol wrong type for socket

	//2019/08/10 11:07:09 [whichNetErr][net.Error][net.OpError][os.SyscallError][os.SyscallError.Syscall=write]
	// [syscall.Errno][errno=32][errno=syscall.EPIPE]
	// err= write tcp 127.0.0.1:59224->127.0.0.1:3390: write: broken pipe

	_ = cnn.Close()
}

func main() {
	connConnectReadError()
}
