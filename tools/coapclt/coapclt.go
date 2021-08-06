package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	coap2 "github.com/dustin/go-coap" // 老版本 "github.com/go-ocf/go-coap"
	coapmsg "github.com/plgd-dev/go-coap/v2/message"
	coapudp "github.com/plgd-dev/go-coap/v2/udp"
	coapclt "github.com/plgd-dev/go-coap/v2/udp/client"
)

// 两个 CoAP package
// github.com/dustin/go-coap 比较原始

func post(cnn *coapclt.ClientConn, path string) {
	resp, err := cnn.Post(context.Background(), path, coapmsg.AppJSON, strings.NewReader(""))
	if err != nil {
		log.Printf("error of post %v", err)
		return
	}
	_ = resp
	//log.Printf("Response Type=%v Code=%v PldLen=%v", resp.Type(),resp.Code(), len(resp.Payload()),)
}

func coap1Client(raddr string, path string) {
	// 发现的缺陷 MID 发送到 65536 之后 再有个 0 2 4 就不能发送了
	clt, err := coapudp.Dial(raddr, coapudp.WithContext(context.Background()))
	if err != nil {
		log.Fatal("failed dail err=%v", err)
	}
	for {
		post(clt, path)
	}
	_ = clt.Close()
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
