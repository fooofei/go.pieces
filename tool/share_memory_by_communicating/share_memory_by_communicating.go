package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// download from https://golang.org/doc/codewalk/sharemem/

const (
	numPollers     = 5                // number of Poller goroutines to launch
	pollInterval   = 5 * time.Second  // how often to poll each URL
	statusInterval = 3 * time.Second  // how often to log status to stdout
	errTimeout     = 10 * time.Second // back-off timeout on error
)

// State represents the last-known state of a URL.
type State struct {
	url    string
	status string
}

// StateMonitor maintains a map that stores the state of the URLs being
// polled, and prints the current state every updateInterval seconds.
// It returns a chan State to which resource state should be sent.
// The caller can close the returned chan to make StateMonitor exit
func StateMonitor(updateInterval time.Duration) chan<- State {
	updates := make(chan State)
	urlStatus := make(map[string]string)
	ticker := time.NewTicker(updateInterval)
	go func() {
	statusLoop:
		for {
			select {
			case <-ticker.C:
				logState(urlStatus)
			case s, more := <-updates:
				if !more {
					break statusLoop
				}
				urlStatus[s.url] = s.status
			}
		}
		logState(urlStatus)
		log.Printf("StateMonitor exit")
	}()
	return updates
}

// logState prints a state map.
func logState(s map[string]string) {
	log.Println("Current state:")
	for k, v := range s {
		log.Printf(" %s %s", k, v)
	}
}

// Resource represents an HTTP URL to be polled by this program.
type Resource struct {
	url      string
	errCount int
}

// Poll executes an HTTP HEAD request for url
// and returns the HTTP status string or an error string.
func (r *Resource) Poll() string {
	resp, err := http.Head(r.url)
	if err != nil {
		log.Println("Error", r.url, err)
		r.errCount++
		return err.Error()
	}
	r.errCount = 0
	log.Printf("Poll %v %v", r.url, resp.Status)
	return resp.Status
}

func Poller(in <-chan *Resource, out chan<- *Resource, status chan<- State) {
	for r := range in {
		s := r.Poll()
		status <- State{r.url, s}
		out <- r
	}
	log.Printf("Poller exit")
}

func main() {

	var urls = []string{
		"http://www.sohu.com/",
		"https://www.baidu.com/",
		//"http://golang.org/",
		//"http://blog.golang.org/",
	}

	// Create our input and output channels.
	pending := make(chan *Resource)
	complete := make(chan *Resource, 100)

	// Launch th e StateMonitor.
	status := StateMonitor(statusInterval)

	// Launch some Poller goroutines.
	pollerGrp := &sync.WaitGroup{}
	for i := 0; i < numPollers; i++ {
		// pending -> complete
		// pending -> status
		pollerGrp.Add(1)
		go func() {
			Poller(pending, complete, status)
			pollerGrp.Done()
		}()
	}

	// Send some Resources to the pending queue.
	go func() {
		for _, url := range urls {
			pending <- &Resource{url: url}
		}
	}()

	completeToPendingGrp := &sync.WaitGroup{}
	i := 0
	for r := range complete {
		// complete -> pending
		completeToPendingGrp.Add(1)
		go func(r *Resource) {
			time.Sleep(pollInterval + errTimeout*time.Duration(r.errCount))
			pending <- r
			completeToPendingGrp.Done()
		}(r)
		i++
		if i > 10 {
			break
		}
	}

	// 退出失败，因为 complete -> pending 退出了
	// 但是 pending -> complete 还在运行，且 cap(chan)=1
	// 于是它就阻塞在 complete chan 上，放不进去, 因此退出的 要点是 chan 有一定容量
	completeToPendingGrp.Wait()
	log.Printf("close(pending)")
	close(pending)
	pollerGrp.Wait()
	log.Printf("close(complete)")
	close(complete)
	close(status)

	// 因为 pending <-> complete 双向
	// 因此不能 close(pending) 不能 close(complete)
	// for 有了 break ，打破了循环，才能退出
}
