package go_pieces

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

var (
	ErrTimeout = errors.New("timeout")
)

func waitDeadline(t time.Time, count int) error {
	timeout := -time.Since(t)
	if timeout <= 0 {
		return ErrTimeout
	}
	fmt.Printf("wait timeout %v\n", timeout.String())
	<-time.After(timeout)
	return nil
}

func waitContext(ctx context.Context, count int) error {
	if d, ok := ctx.Deadline(); ok {
		return waitDeadline(d, count)
	}
	return nil
}

// 没有这个等待，测试用例会大概率失败
func untilContextDone(ctx context.Context) {
	if d,ok := ctx.Deadline() ; ok {
		if time.Until(d) <=0 {
			fmt.Printf("wait for ctx.Done\n")
			<- ctx.Done()
		}
	}
}

func TestDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	count := 0
	var err error
	var errs []error
loop:
	for {
		err = waitContext(ctx, count)
		if err != nil {
			errs = append(errs, err)
		}
		_ = err
		count += 1

		untilContextDone(ctx)
		select {
		case <-ctx.Done():
			break loop
		default:
		}
	}

	assert.Equal(t, count == 1, true,
		fmt.Sprintf("%v count = %v, errs size %v", time.Now().Format(time.RFC3339), count, len(errs)))
}
