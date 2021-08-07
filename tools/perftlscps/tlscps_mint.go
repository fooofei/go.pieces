package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/bifurcation/mint"
)

func mintHexView(v interface{}) string {
	var binBuf bytes.Buffer
	err := binary.Write(&binBuf, binary.LittleEndian, v)
	if err != nil {
		panic(err)
	}

	//return hex.EncodeToString(binBuf.Bytes())
	return hex.Dump(binBuf.Bytes())
}

func main1() {
	var raddr string

	raddr = "127.0.0.1:886"

	fmt.Printf("using TLSv1.3 to tcp %v\n", raddr)
	conn, err := mint.Dial("tcp", raddr, nil)

	if err != nil {
		fmt.Printf("err = %v conn=%v \n", err, conn)
	}
	_, _ = conn, err

	fmt.Printf("exit\n")
}
