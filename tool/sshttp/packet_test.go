package sshttp

import (
	"log"
	"strings"
	"testing"
)

func TestHttpPacket_Decode_NoPayload(t *testing.T) {
	httpHeaders := `POST /proxy/0/0 HTTP/1.1
Content-Length:0
Connect:114.115.186.107:22

`

	hp := HttpPacket{}
	err := hp.Decode(strings.NewReader(httpHeaders))
	if err != nil {
		log.Fatalf("err= %v", err)
	}
	v := hp.Encode()
	t.Logf("encode= [%s]", v)

}

func TestHttpPacket_Encode_Payload(t *testing.T) {
	httpHeaders := `POST /proxy/0/0 HTTP/1.1
Content-Length:5
Connect:114.115.186.107:22

hello`

	hp := HttpPacket{}
	err := hp.Decode(strings.NewReader(httpHeaders))
	if err != nil {
		log.Fatalf("err= %v", err)
	}
	v := hp.Encode()
	t.Logf("encode= [%s]", v)
}
