package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	coap2 "github.com/dustin/go-coap"
	coap "github.com/go-ocf/go-coap"
)

// 两个 CoAP package
// github.com/dustin/go-coap 比较原始

func post(cnn *coap.ClientConn, path string) {
	resp, err := cnn.Post(path, coap.AppJSON, strings.NewReader(""))
	if err != nil {
		log.Printf("error of post %v", err)
		return
	}
	_ = resp
	//log.Printf("Response Type=%v Code=%v PldLen=%v", resp.Type(),resp.Code(), len(resp.Payload()),)
}

func coap1Client(raddr string, path string) {
	// 发现的缺陷 MID 发送到 65536 之后 再有个 0 2 4 就不能发送了
	clt := &coap.Client{
		Net: "udp",
	}
	cnn, err := clt.Dial(raddr)
	if err != nil {
		log.Fatal("error of dial =%v", err)
	}

	for {
		post(cnn, path)
	}
	_ = cnn.Close()
}

func coap2Client(raddr string, path string) {
	cnn, err := coap2.Dial("udp", raddr)
	if err != nil {
		log.Fatal(err)
	}
	req := coap2.Message{}
	req.Type = coap2.Confirmable
	req.Code = coap2.POST
	req.Token = []byte("1")
	req.SetPathString(path)
	req.AddOption(coap2.ContentFormat, coap2.AppJSON)
	req.AddOption(coap2.Accept, coap2.AppJSON)

	for {
		resp, err := cnn.Send(req)
		if err != nil {
			log.Printf("error of post %v", err)
		}
		if resp != nil {
			//log.Printf("Response Type=%v Code=%v PldLen=%v", resp.Type,resp.Code, len(resp.Payload),)
		}
		req.MessageID += 1
	}

}

func main() {

	//
	raddr := ""
	path := ""
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//

	_ = raddr
	_ = path

	for i := 0; i < 4; i += 1 {
		go coap2Client(raddr, path)
	}

	coap2Client(raddr, path)

	log.Printf("main exit")
}
