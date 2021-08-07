package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	fnet "github.com/fooofei/pkg/net"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// 如何检查一个服务器或者开放的服务的健康与否
// ICMP TCP UDP 三种途径的示例

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
	waitCnnCose, closeCnn := fnet.CloseWhenContext(waitCtx, c)
	rbs := make([]byte, 32*1024)
	n, err := c.Read(rbs)
	closeCnn()
	<-waitCnnCose.Done()

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

// tcpAlive 检查 tcp 端口是否健康 raddr is server ip:port
func tcpAlive(waitCtx context.Context, raddr string) error {
	d := &net.Dialer{}
	c, err := d.DialContext(waitCtx, "tcp", raddr)
	if err != nil {
		return err
	}
	return c.Close()
}

// udpAlive 检查 udp 端口是否健康 raddr is server's ip:port
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
	wait, stop := fnet.CloseWhenContext(waitCtx, c)
	rbs := make([]byte, 32*1024)
	n, err := c.Read(rbs)
	stop()
	<-wait.Done()
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
