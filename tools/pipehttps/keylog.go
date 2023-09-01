package main

import (
	"context"
	"io"
	"os"
)

type ioWriterAsFunc func(p []byte) (n int, err error)

func (w ioWriterAsFunc) Write(p []byte) (n int, err error) {
	return w(p)
}

func createKeyLogWriter(ctx context.Context, ch chan []byte) io.Writer {
	var pfn = func(p []byte) (n int, err error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}
		ch <- p
		return len(p), nil
	}
	return ioWriterAsFunc(pfn)
}

func writeKeyLog(ctx context.Context, filePath string, ch chan []byte) {
	var f, err = os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var writed = false
	// 因为 ch 不是我们关闭的，因此我们不能以 ch 关闭为结束，只能以 ctx 为结束
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case value := <-ch:
			f.Write(value)
			writed = true
		}
	}
	if !writed {
		f.Close()
		// remove empty file
		os.Remove(filePath)
	}
}
