package netonn

import (
	"errors"
	"github.com/stretchr/testify/require"
	"net"
	"os"
	"runtime"
	"testing"
	"time"
)

// 这个是老式的了，写这么复杂，业务代码也不容易维护
// https://liudanking.com/network/go-%E4%B8%AD%E5%A6%82%E4%BD%95%E5%87%86%E7%A1%AE%E5%9C%B0%E5%88%A4%E6%96%AD%E5%92%8C%E8%AF%86%E5%88%AB%E5%90%84%E7%A7%8D%E7%BD%91%E7%BB%9C%E9%94%99%E8%AF%AF/


func TestConnDialTimeoutErr(t *testing.T) {
	addr := "2.2.2.2:3390"
	dialTimeout := time.Second * 3
	cnn, err := net.DialTimeout("tcp", addr, dialTimeout)
	var netErr net.Error
	require.Equal(t, true, errors.As(err, &netErr))
	require.Equal(t, true, netErr.Timeout())
	require.Equal(t, true, netErr.Temporary())
	require.Equal(t, true, os.IsTimeout(err))
	require.Equal(t, "dial tcp 2.2.2.2:3390: i/o timeout", err.Error())
	_ = cnn
}

func TestConnRefusedErr(t *testing.T) {
	addr := "127.0.0.1:3390"
	dialTimeout := time.Second * 3
	cnn, err := net.DialTimeout("tcp", addr, dialTimeout)
	var netErr net.Error
	require.Equal(t, true, errors.As(err, &netErr))
	require.Equal(t, false, netErr.Timeout())
	require.Equal(t, false, netErr.Temporary())
	require.Equal(t, false, os.IsTimeout(err))
	if runtime.GOOS == "windows" {
		require.Equal(t, "dial tcp 127.0.0.1:3390: connectex: No connection could be made because the target machine actively refused it.", err.Error())
	} else {
		require.Equal(t, "dial tcp 127.0.0.1:3390: connect: connection refused", err.Error())
	}
	_ = cnn
}
