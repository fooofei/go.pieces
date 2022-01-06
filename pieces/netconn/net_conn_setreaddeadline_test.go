package netonn

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/stretchr/testify/require"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

func writeToConn(t *testing.T, logger logr.Logger, addr string) {
	cnn, err := net.Dial("tcp", addr)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		sleep := time.Second * 3
		logger.Info("enter sleep", "sleepFor", sleep.String())
		time.Sleep(sleep)
		logger.Info("leave sleep", "sleepFor", sleep.String())
		_, err = fmt.Fprintf(cnn, "hello")
		logger.Info("done write to connection", "err", fmt.Sprintf("%v", err))
	}
	_ = cnn.Close()
}

// TestDeadlineNoRead 测试如果没有在指定的 deadline 时间内读取到
// 测试结果：Read() 就会立刻返回 并且之后的尝试 Read() 都是不行的
// 修复方式：重新 SetReadDeadline 更长时间，或者清除 deadline
func TestDeadlineNoRead(t *testing.T) {
	// [1] SetReadDeadline 2 s
	// [1] Read()
	// [2] after 3s write
	// [1] Read() timeout

	// 这样补救
	// [1] SetReadDeadline 2s
	// [1] Read()
	// [1]' SetReadDeadline more
	// [2] after 3s write
	// [1] Read() something

	rootLogger := stdr.New(log.New(os.Stdout, "", log.LstdFlags))
	logger := rootLogger.WithName("[Read]")
	addr := "127.0.0.1:3389"
	lsnCnn, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	logger.Info("start listen tcp port", "addr", addr)

	wait := &sync.WaitGroup{}
	wait.Add(1)
	go func() {
		writeToConn(t, rootLogger.WithName("[write]"), addr)
		wait.Done()
	}()

	cnn, err := lsnCnn.Accept()
	require.NoError(t, err)

	buf := make([]byte, 10)
	timeOff := time.Second * 2
	logger.Info("set read deadline for", "off", timeOff.String())
	cnn.SetReadDeadline(time.Now().Add(timeOff))

	fixFunc := func() {
		time.Sleep(time.Second) // 必须小于第一次设置的 2s
		timeOff = time.Second * 5
		logger.Info("set read deadline again", "off", timeOff.String())
		cnn.SetDeadline(time.Now().Add(timeOff))
	}
	_ = fixFunc
	// 可以继续设置 补救回来 ，之后就能读到内容了
	// go fixFunc()
	logger.Info("enter for read")
	nr, er := cnn.Read(buf)
	logger.Info("leave read", "err", fmt.Sprintf("%v", er), "size", nr)

	// 再尝试 5 次读取
	for i := 0; i < 5; i++ {
		lv := []interface{}{"index", i}
		logger.Info("enter for read", lv...)
		nr, er = cnn.Read(buf)
		logger.Info("leave read", append(lv, "err", fmt.Sprintf("%v", er), "size", nr)...)
	}

	cnn.Close()
	logger.Info("close connection for not read and write")
	wait.Wait()
	lsnCnn.Close()
}

// TestDeadlineReadSome 测试如果在指定的时间内，读取到一部分，剩下的部分是否还能读取到
// 测试结果：无法再继续读取
// 修复方式: 重新设置新的更长的 deadline 或者清除 deadline
func TestDeadlineReadSome(t *testing.T) {
	rootLogger := stdr.New(log.New(os.Stdout, "", log.LstdFlags))
	logger := rootLogger.WithName("[Read]")
	addr := "127.0.0.1:3389"
	lsnCnn, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	logger.Info("start listen tcp port", "addr", addr)

	wait := &sync.WaitGroup{}
	wait.Add(1)
	go func() {
		writeToConn(t, rootLogger.WithName("[write]"), addr)
		wait.Done()
	}()

	cnn, err := lsnCnn.Accept()
	require.NoError(t, err)

	buf := make([]byte, 10)
	timeOff := time.Second * 5
	logger.Info("set read deadline for", "off", timeOff.String())
	cnn.SetReadDeadline(time.Now().Add(timeOff))

	logger.Info("enter for read")
	nr, er := cnn.Read(buf)
	logger.Info("leave read", "err", fmt.Sprintf("%v", er), "size", nr)

	logger.Info("enter for 2 read")
	nr, er = cnn.Read(buf)
	logger.Info("leave 2 read", "err", fmt.Sprintf("%v", er), "size", nr)

	logger.Info("enter sleep for 3 read")
	time.Sleep(4 * time.Second)
	logger.Info("enter for 3 read")
	fixFunc := func() {
		cnn.SetDeadline(time.Time{})
	}
	_ = fixFunc
	// fixFunc()
	nr, er = cnn.Read(buf)
	logger.Info("leave 3 read", "err", fmt.Sprintf("%v", er), "size", nr)

	cnn.Close()
	logger.Info("close connection for not read and write")
	wait.Wait()
	lsnCnn.Close()
}
