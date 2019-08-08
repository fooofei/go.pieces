package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// 如何检查一个服务器或者开放的服务的健康与否
// ICMP TCP UDP 三种途径的示例

func takeOverCnnClose(waitCtx context.Context, cnn io.Closer) (chan bool, *sync.WaitGroup) {
	noWait := make(chan bool, 1)
	waitGrp := &sync.WaitGroup{}
	waitGrp.Add(1)
	go func() {
		select {
		case <-noWait:
		case <-waitCtx.Done():
		}
		_ = cnn.Close()
		waitGrp.Done()
	}()
	return noWait, waitGrp
}

// raddr is ip
// if not error, the server alive well
func icmpAlive(waitCtx context.Context, raddr string) error {
	var err error

	d := &net.Dialer{}

	c, err := d.DialContext(waitCtx, "ip4:icmp", raddr)
	if err != nil {
		return err
	}
	msg := &icmp.Message{}
	msg.Type = ipv4.ICMPTypeEcho
	msg.Body = &icmp.Echo{
		ID:   os.Getpid() & 0xFFFF,
		Seq:  1,
		Data: []byte("HELLO"),
	}
	bs, err := msg.Marshal(nil)
	if err != nil {
		_ = c.Close()
		return err
	}
	_, _ = c.Write(bs)
	noWait, waitGrp := takeOverCnnClose(waitCtx, c)
	rbs := make([]byte, 32*1024)
	n, err := c.Read(rbs)
	close(noWait)
	waitGrp.Wait()

	if err != nil {
		return err
	}
	// 其实 protocol 是固定的值, 每个ICMPTypeEchoReply ICMPTypeEcho 等的 Protocol()方法返回一样的值
	msg, err = icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), rbs[:n])
	if err != nil {
		return err
	}
	if msg.Type.Protocol() == ipv4.ICMPTypeEchoReply.Protocol() {
		return nil
	}
	return fmt.Errorf("icmp peer reply not expect type=%v", msg.Type)
}

// raddr is server ip:port
func tcpAlive(waitCtx context.Context, raddr string) error {
	d := &net.Dialer{}
	c, err := d.DialContext(waitCtx, "tcp", raddr)
	if err != nil {
		return err
	}
	return c.Close()
}

// raddr is server's ip:port
func udpAlive(waitCtx context.Context, raddr string) error {
	var err error
	d := &net.Dialer{}
	c, err := d.DialContext(waitCtx, "udp", raddr)
	if err != nil {
		return err
	}
	_, err = c.Write([]byte("HELLO"))
	if err != nil {
		_ = c.Close()
		return err
	}
	noWait, waitGrp := takeOverCnnClose(waitCtx, c)
	rbs := make([]byte, 32*1024)
	n, err := c.Read(rbs)
	close(noWait)
	waitGrp.Wait()
	if err != nil {
		return err
	}
	// keep the n, may check the read content
	_ = n
	return nil
}

func main() {

	log.Printf("main exit")
}
