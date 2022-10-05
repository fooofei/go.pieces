package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// simple print current time of local and utc format.
// 通过对比 学习各个时间的区别

func clock() {
	tick := time.Tick(time.Second)
	for {
		select {
		case <-tick:
			now := time.Now()
			log.Printf("local= %v utc= %v unix= %v unixNano= %v",
				now.Local().Format(time.RFC3339),
				now.UTC().Format(time.RFC3339),
				now.Unix(),
				now.UnixNano())
		}
	}
}

func main() {
	f := log.Flags()
	f &= ^log.Ldate
	f &= ^log.Ltime
	log.SetFlags(f)
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
	clock()
}

// example
// pid= 10100 local= 2022-10-05T16:03:17+08:00 utc= 2022-10-05T08:03:17Z unix= 1664956997 unixNano= 1664956997657023800
