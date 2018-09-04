package main

// download from https://gist.github.com/zupzup/14ea270252fbb24332c5e9ba978a8ade

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

func init() {

}

func checkErrorExit(err interface{}, rsize int64) bool {
	if err == io.EOF {
		return true
	}
	if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
		return false
	}
	if rsize == 0 {
		return true
	}
	return false
}

func handleConnection(wg *sync.WaitGroup, client net.Conn, target string, childStop chan bool) {

	defer wg.Done()
	defer client.Close()
	defer log.Printf("conn %v-%v closed\n", client.LocalAddr(), client.RemoteAddr())


	connStop := make(chan bool, 2)
	// when this goroutine or the sub goroutine exit, one of them exit, then
	// all of them should exit
	defer func() {
		connStop <- true
	}()


	// dial target
	anotherConn, err := net.DialTimeout("tcp", target, time.Duration(5)*time.Second)
	if err != nil {
		log.Printf("could not connect to target %v", err)
		return
	}
	defer anotherConn.Close()
	defer log.Printf("conn %v->%v closed\n", anotherConn.LocalAddr(), anotherConn.RemoteAddr())
	log.Printf("%v->%v connect\n", anotherConn.LocalAddr(), anotherConn.RemoteAddr())

	direction := fmt.Sprintf("%v->%v", anotherConn.RemoteAddr(), client.RemoteAddr())
	defer log.Printf("%v aged\n", direction)

	wg.Add(1)
	go func() {
		defer wg.Done()

		direction := fmt.Sprintf("<sub> %v->%v", client.RemoteAddr(), anotherConn.RemoteAddr())
		defer log.Printf("%v aged\n", direction)
		defer func() {
			// connStop 如果要用 close 来通知 那么就会有2个goroutine 同时 close 的现象，不允许
			connStop <- true
		}()

		for {
			readTimeOut := time.Now().Add(1e9)
			err := client.SetReadDeadline(readTimeOut)
			if err != nil {
				log.Fatal(err)
				return
			}
			select {
			case <-childStop:
				log.Printf("%v rcv childStop exit\n", direction)
				return
			case <-connStop:
				log.Printf("%v rcv connStop exit\n", direction)
				return
			default:

			}
			rsize, err := io.Copy(anotherConn, client)

			if checkErrorExit(err, rsize) {
				log.Printf("%v read %v err=%T,%v\n", direction, rsize, err, err)
				log.Printf("%v checkErrorExit=true\n", direction)
				return
			}
		}

	}()

	for {
		readTimeOut := time.Now().Add(1e9)
		err := anotherConn.SetReadDeadline(readTimeOut)
		if err != nil {
			log.Fatal(err)
			return
		}
		select {
		case <-childStop:
			log.Printf("%v rcv childStop exit\n", direction)
			return
		case <-connStop:
			log.Printf("%v rcv connStop exit\n", direction)
			return
		default:

		}
		rsize, err := io.Copy(client, anotherConn)

		if checkErrorExit(err, rsize) {
			log.Printf("%v read %v err=%v\n", direction, rsize, err)
			log.Printf("%v checkErrorExit=true\n", direction)
			return
		}
	}
}

func main() {
	var target string
	var port int
	//flag.StringVar(&target, "target", "", "the target (<host>:<port>)")
	//flag.IntVar(&port, "port", 7757, "the tunnelthing port")

	// show file number in log
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//flag.Parse()
	target = "192.145.109.101:2345"
	port = 3306

	var wg sync.WaitGroup

	childStop := make(chan bool)

	// setup signal
	wg.Add(1)
	go func() {
		defer wg.Done()
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		s := <-signals
		log.Printf("rcv signal %v, exit signal goroutine\n", s)
		close(childStop)
	}()

	lsnConn, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("could not start server on %v: %v", port, err)
	}
	defer lsnConn.Close()
	log.Printf("server running on %v\n", lsnConn.Addr())

	for {
		//set timeout
		l := lsnConn.(*net.TCPListener)
		l.SetDeadline(time.Now().Add(time.Second * time.Duration(2)))
		client, err := lsnConn.Accept()
		if err != nil {
			// timeout
			//log.Printf("no accpet err=%v", err)
		} else {
			log.Printf("accept %v->%v\n", client.RemoteAddr(), client.LocalAddr())
			wg.Add(1)
			go handleConnection(&wg, client, target, childStop)
		}
		stopAccept := false
		select {
		case <-childStop:
			stopAccept = true
		default:
		}

		if stopAccept {
			break
		}
	}

	log.Println("main wait goroutine")
	wg.Wait()
	log.Println("main exit")
}
