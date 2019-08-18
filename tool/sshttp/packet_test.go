package sshttp

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestHTTPPathEncodeDecode(t *testing.T) {
	req, err := NewLogin()
	if err != nil {
		t.Fatal(err)
	}
	w := &bytes.Buffer{}
	err = req.Write(w)
	if err != nil {
		t.Fatal(err)
	}
	r := w.Bytes()
	rr := bytes.NewReader(r)
	bior := bufio.NewReader(rr)
	req2, err := http.ReadRequest(bior)
	assert.Equal(t, err == nil, true)
	if err != nil {
		t.Fatal(err)
	}
	w.Reset()
	err = req2.Write(w)
	assert.Equal(t, err == nil, true)
	if err != nil {
		t.Fatal(err)
	}
	r2 := w.Bytes()

	assert.Equal(t, string(r), string(r2))

}

func TestHttpPath_ParseHTTPPath(t *testing.T) {

	req, err := NewLogin()
	assert.Equal(t, err == nil, true)
	if err != nil {
		t.Fatal(err)
	}
	httpPath, err := ParseUrlPath(req.URL)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, httpPath.String(), "/login/0/0")
}

func TestHttpPacket_Decode_NoPayload(t *testing.T) {
	httpHeaders := `POST /proxy/0/0 HTTP/1.1
Content-Length:0
Connect:114.115.186.107:22

`
	r := strings.NewReader(httpHeaders)
	bior := bufio.NewReader(r)
	req, err := http.ReadRequest(bior)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, req.URL.Path, "/proxy/0/0")
}

func TestHttpPacket_Encode_Payload(t *testing.T) {
	httpHeaders := `POST /proxy/0/0 HTTP/1.1
Content-Length:5
Connect:114.115.186.107:22

hello`

	r := strings.NewReader(httpHeaders)
	bior := bufio.NewReader(r)
	req, err := http.ReadRequest(bior)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, req.URL.Path, "/proxy/0/0")
	body, _ := ioutil.ReadAll(req.Body)
	assert.Equal(t, string(body), "hello")
}
