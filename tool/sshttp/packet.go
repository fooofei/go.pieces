package sshttp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	contentLengthName = "Content-Length"
)

const (
	Login = "login"
	Proxy = "proxy"
)

type HttpPacket struct {
	Type   string
	Seq    int64
	Ack    int64
	Header map[string]string
	Body   []byte
}

func (hp *HttpPacket) Encode() []byte {
	w := &bytes.Buffer{}

	_, _ = fmt.Fprintf(w, "POST /%v/%v/%v HTTP/1.1\r\n", hp.Type, hp.Seq, hp.Ack)
	_, _ = fmt.Fprintf(w, "%v:%v\r\n", contentLengthName, len(hp.Body))
	// Cookie ?

	for k, v := range hp.Header {
		_, _ = fmt.Fprintf(w, "%v:%v\r\n", k, v)
	}
	_, _ = fmt.Fprintf(w, "\r\n")
	w.Write(hp.Body)
	return w.Bytes()
}

func (hp *HttpPacket) Decode(r io.Reader) error {

	ioR := bufio.NewReader(r)
	mimeReader := textproto.NewReader(ioR)
	// get first line
	headLine, err := mimeReader.ReadLine()
	if err != nil {
		return err
	}
	if len(headLine) == 0 {
		return errors.Errorf("empty in first line")
	}
	maps, err := mimeReader.ReadMIMEHeader()
	if err != nil {
		return err
	}

	headLineSubs := strings.Split(headLine, " ")
	if len(headLineSubs) < 3 {
		return errors.Errorf("short head line \"%v\"", headLine)
	}
	headLineMiddle := headLineSubs[1]
	tsa := strings.Split(headLineMiddle, "/")
	if len(tsa) < 4 {
		return errors.Errorf("short head line middle \"%v\"", headLineMiddle)
	}
	hp.Type = tsa[1]
	v1, err := strconv.ParseFloat(tsa[2], 64)
	if err != nil {
		return err
	}
	hp.Seq = int64(v1)
	v1, err = strconv.ParseFloat(tsa[3], 64)
	if err != nil {
		return err
	}
	hp.Ack = int64(v1)

	if hp.Header == nil {
		hp.Header = make(map[string]string)
	}
	for k, v := range maps {
		if len(v) > 0 {
			hp.Header[k] = v[0]
		} else {
			hp.Header[k] = ""
		}
	}

	// Read the rest of reader for body bytes
	if sizeStr, exists := hp.Header[contentLengthName]; exists {
		if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
			if body, err := ioutil.ReadAll(ioR); err == nil {
				if len(body) == int(size) {
					hp.Body = body
				}
			}
		}
	}

	delete(hp.Header, contentLengthName)

	return nil
}

func NewLogin() []byte {
	p := &HttpPacket{}
	p.Type = Login
	return p.Encode()
}

func NewProxyConnect(value string) []byte {
	p := &HttpPacket{}
	p.Type = Proxy
	p.Header["Connect"] = value
	return p.Encode()
}
