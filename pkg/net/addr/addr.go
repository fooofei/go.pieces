package addr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

// StringToNetwork will parse a ip of string format to network order uint32 format
func StringToNetwork(s string) (uint32, error) {
	var ip = net.ParseIP(s)
	if ip == nil {
		return 0, fmt.Errorf("failed parse ip, invalid '%s'", s)
	}
	return binary.LittleEndian.Uint32(ip.To4()), nil
}

func NetworkToString(n uint32) (string, error) {
	var b = new(bytes.Buffer)
	if err := binary.Write(b, binary.LittleEndian, n); err != nil {
		return "", err
	}
	return net.IP(b.Bytes()).String(), nil
}

func NetworkToHost(n uint32) (uint32, error) {
	var b = new(bytes.Buffer)
	if err := binary.Write(b, binary.BigEndian, n); err != nil {
		return 0, err
	}
	var r uint32
	var err = binary.Read(b, binary.LittleEndian, &r)
	return r, err
}

func HostToNetwork(h uint32) (uint32, error) {
	var b = new(bytes.Buffer)
	if err := binary.Write(b, binary.LittleEndian, h); err != nil {
		return 0, err
	}
	var r uint32
	var err = binary.Read(b, binary.BigEndian, &r)
	return r, err
}
