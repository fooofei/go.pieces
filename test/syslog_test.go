// 演示在 Golang 中如何使用 syslog
package go_pieces

import (
	"fmt"
	"log/syslog"
	"testing"
	"time"
)

// macOS write to $ less +G /var/log/system.log
// linux write to  /var/log/messages
func TestUseSyslog(t *testing.T) {
	l, err := syslog.New(syslog.LOG_USER, "testMySysLOG")

	if err != nil {
		t.Fatal(err)
	}

	tick := time.Tick(time.Second * 1)
	// 做出来一个每秒 500 次 LOG 输出
	for {
		select {
		case <-tick:
			for i := 0; i < 5; i += 1 {
				_ = l.Warning(fmt.Sprintf("%v %v", time.Now().Format(time.RFC3339), i))
			}
		}
		break
	}
	_ = l.Close()
}
