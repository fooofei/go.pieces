package test

import (
	"crypto/tls"
	"io"
	"net"
	"testing"
	"time"

	"gotest.tools/assert"
)

// blog https://www.calhoun.io/when-nil-isnt-equal-to-nil/
func TestNetConnCmpNil(t *testing.T) {
	var cnn io.ReadWriteCloser = nil
	var nilInterface interface{}

	assert.Equal(t, cnn, nil)         //  (<nil>, <nil>) ==  (<nil>, <nil>)
	assert.Equal(t, cnn == nil, true) //  (<nil>, <nil>) ==  (<nil>, <nil>)
	t.Logf("cnn= (%T, %v)", cnn, cnn)

	assert.Equal(t, nilInterface, nil)         //  (<nil>, <nil>) ==  (<nil>, <nil>)
	assert.Equal(t, nilInterface == nil, true) //  (<nil>, <nil>) ==  (<nil>, <nil>)
	t.Logf("nilInterface= (%T, %v)", nilInterface, nilInterface)

	d := &net.Dialer{Timeout: time.Second * 3}

	tlsCfg := &tls.Config{InsecureSkipVerify: true}
	// no need setup server
	tlsAddr := "127.0.0.1:8899"
	// 访问不存在的地址，制造一个 err
	tlsCnn, err := tls.DialWithDialer(d, "tcp", tlsAddr, tlsCfg)
	t.Logf("dial tls err= %v", err)

	assert.Equal(t, tlsCnn == nil, true) // (*tls.Conn, <nil>) == (*tls.Conn, <nil>)
	// false of // assert.Equal(t, cnnTLs, nil) // (*tls.Conn, <nil>) != (<nil>, <nil>)
	assert.Equal(t, tlsCnn, (*tls.Conn)(nil))      // (*tls.Conn, <nil>) == (*tls.Conn, <nil>)
	assert.Equal(t, tlsCnn == nilInterface, false) // (*tls.Conn, <nil>) != (<nil>, <nil>)
	t.Logf("tlsCnn= (%T, %v)", tlsCnn, tlsCnn)

	cnn = tlsCnn
	t.Logf("after cnn=tlsCnn cnn= (%T, %v) tlsCnn= (%T, %v)",
		cnn, cnn, tlsCnn, tlsCnn)
	// false of // assert.Equal(t, cnn, nil) // (*tls.Conn, <nil>) != (<nil>, <nil>)
	assert.Equal(t, cnn == nilInterface, false) // (*tls.Conn, <nil>) != (<nil>, <nil>)
	assert.Equal(t, tlsCnn == nil, true)        // (*tls.Conn, <nil>)  == (*tls.Conn, <nil>)
	assert.Equal(t, cnn == nil, false)          // (*tls.Conn, <nil>) != (<nil>, <nil>)
	assert.Equal(t, cnn == (*tls.Conn)(nil), true)

	if tlsCnn != nil {
		_ = tlsCnn.Close()
	}

	// bad
	//if cnn != nil {
	//	_ = cnn.Close()
	//}

	// Good
	if tc, is := cnn.(*tls.Conn); is && tc != nil {
		_ = tc.Close()
	}
}

type CustomStruct struct {
}

func returnCustomStructPointer() interface{} {
	var nullPointer *CustomStruct
	return nullPointer
}

func TestCustomStructCmpNil(t *testing.T) {
	nullPointer := returnCustomStructPointer()
	assert.Equal(t, nullPointer == nil, false)
	assert.Equal(t, nullPointer, (*CustomStruct)(nil))
}
