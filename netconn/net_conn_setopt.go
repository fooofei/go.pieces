// +build !windows

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"syscall"
)

// 演示 如何给 net.Conn setsockopt
// 最重要的是如何完成在 Dial/Accept 之前设置选项

// convert "127.0.0.1:22" to TOA bytes
func toTOAString(addr string) (string, error) {
	rawAddr := &syscall.RawSockaddrInet4{}
	rawAddr.Family = syscall.AF_INET

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return "", err
	}
	portBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBytes, uint16(tcpAddr.Port))
	rawAddr.Port = binary.BigEndian.Uint16(portBytes)
	_ = copy(rawAddr.Addr[:], tcpAddr.IP.To4())

	w := &bytes.Buffer{}
	// WARNING: not binary.BigEndian
	_ = binary.Write(w, binary.LittleEndian, rawAddr)
	return w.String(), nil
}

func setTOAOpt(fd int, toaOpt string) error {
	const TCP_TOA int = 512
	toaString, err := toTOAString(toaOpt)
	if err != nil {
		return err
	}
	log.Printf("toTOAString %v err= %v", toaString, err)
	err = syscall.SetsockoptString(fd, syscall.IPPROTO_IP, TCP_TOA, toaString)
	log.Printf("setsockopt TOA err= %v", err)
	return err
}

// 留着这个函数 学习如何从 net.Conn 得到 fd
func setTOAOptDeprecated(cnn net.Conn) error {
	// 这个 issue 也不教我们怎么转 https://github.com/golang/go/issues/6966
	const TCP_TOA int = 512
	tcpConn, is := cnn.(*net.TCPConn)
	if !is {
		return fmt.Errorf("not tcp conn")
	}
	toaString, err := toTOAString("100.101.102.103:22")
	if err != nil {
		return err
	}
	file, err := tcpConn.File()
	if err != nil {
		return err
	}
	err = syscall.SetsockoptString(int(file.Fd()), syscall.IPPROTO_IP, TCP_TOA, toaString)

	// the File() will set fd to block mode, we revert it
	// <cannot write after tcpConn.File(), will not work>
	_ = syscall.SetNonblock(int(file.Fd()), true)
	_ = file.Close()
	// 可能要 重新转化为 conn net.FileConn(file)
	return err
}

func TestSetsockoptBeforeDial() {
	// 测试步骤
	// 在运行该程序的主机上抓包 ssh root@x.x.x.x "tcpdump -Z root -i eth0 port 443 -s 0 -U -w -" |  wireshark -k -i -
	// 运行这个程序
	//
	// 教学如何完成设置
	// https://stackoverflow.com/questions/40544096/how-to-set-socket-option-ip-tos-for-http-client-in-go-language
	// 这个 Dialer.Control 是在 Go1.11 后才支持
	rawConnControl := func(fd uintptr) {
		// the fd is <int>(FD.Sysfd),
		// and call uintptr(fd.Sysfd) to this func
		// so we can safely call int(fd) to convert it back
		_ = setTOAOpt(int(fd), "100.101.102.103:22")
	}
	dialerControl := func(network string, address string, cnn syscall.RawConn) error {
		err := cnn.Control(rawConnControl)
		return err
	}
	d := net.Dialer{
		Control: dialerControl,
	}
	addr := "x.x.x.x:443"
	cnn, err := d.Dial("tcp", addr)
	if err != nil {
		_ = cnn.Close()
	}
}
