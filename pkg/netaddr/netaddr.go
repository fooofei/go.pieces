package netaddr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

func StringToNetwork(s string) (uint32, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return 0, fmt.Errorf("failed parse ip, invalid '%s'", s)
	}
	return binary.LittleEndian.Uint32(ip.To4()), nil
}

func NetworkToString(n uint32) (string, error) {
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, n)
	if err != nil {
		return "", err
	}
	return net.IP(b.Bytes()).String(), nil
}

func NetworkToHost(n uint32) (uint32, error) {
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.BigEndian, n)
	if err != nil {
		return 0, err
	}
	var r uint32
	err = binary.Read(b, binary.LittleEndian, &r)
	return r, err
}

func HostToNetwork(h uint32) (uint32, error) {
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, h)
	if err != nil {
		return 0, err
	}
	var r uint32
	err = binary.Read(b, binary.BigEndian, &r)
	return r, err
}
