package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly"
)

const (
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"
	// https://lantouzi.com/post/1440
	Pattern1 = "(?s)至(?P<year>.*)年(?P<month>.*)月(?P<day>.*)日.*到期智选服务(?P<count>\\d+)笔，共(?P<money>.*?)元"
	// https://lantouzi.com/post/2212
	Pattern2 = "(?s)至(?P<year>.*)年(?P<month>.*)月(?P<day>.*)日.*智选服务退出(?P<count>\\d+)笔，共(?P<money>.*?)元"
)

type moneyPost struct {
	Year, Month, Day string
	Count            string
	Money            string
}

// Keep the Code
//func newMoneyPost1(year, month, day, count, money string) *moneyPost {
//	m := &moneyPost{}
//	zero := time.Time{}
//	loc := time.Now().Location()
//	zero.In(loc)
//	y, _ := strconv.ParseInt(year, 10, 64)
//	mon, _ := strconv.ParseInt(month, 10, 64)
//	d, _ := strconv.ParseInt(day, 10, 64)
//	m.Day = zero.AddDate(int(y), int(mon), int(d))
//	m.Count, _ = strconv.ParseInt(count, 10, 64)
//	m.Money = floatlify(money)
//	return m
//}

func newMoneyPost(year, month, day, count, money string) *moneyPost {
	m := &moneyPost{}
	m.Year = year
	m.Month = month
	m.Day = day
	m.Count = count
	m.Money = strings.ReplaceAll(money, ",", "")
	return m
}

func (m *moneyPost) ToRecord() []string {
	r := make([]string, 0, 5)
	r = append(r, fmt.Sprintf("%v-%v-%v", m.Year, m.Month, m.Day))
	r = append(r, m.Count)
	r = append(r, m.Money)
	return r
}

func (m *moneyPost) MapKey() string {
	return fmt.Sprintf("%v-%v-%v", m.Year, m.Month, m.Day)
}

type Stat struct {
	PageReqCount       int64
	PageRespValidCount int64
	PostReqCount       int64
	PostRespValidCount int64
	WriterCacheSize    int64
}

func (s *Stat) String() string {
	b := &bytes.Buffer{}
	_, _ = fmt.Fprintf(b, "PageReqCount %v ", s.PageReqCount)
	_, _ = fmt.Fprintf(b, "PageRespValidCount %v ", s.PageRespValidCount)
	_, _ = fmt.Fprintf(b, "PostReqCount %v ", s.PostReqCount)
	_, _ = fmt.Fprintf(b, "PostRespValidCount %v ", s.PostRespValidCount)
	_, _ = fmt.Fprintf(b, "WriterCacheSize %v ", s.WriterCacheSize)
	return b.String()
}

// contextTransport wrapper a context.Context for cancel requests
type contextTransport struct {
	ctx   context.Context
	trans *http.Transport
}

func (t *contextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.WithContext(t.ctx)
	return t.trans.RoundTrip(req)
}

func newCollector(parallelism int) *colly.Collector {
	c := colly.NewCollector()
	c.UserAgent = UserAgent
	c.Async = true
	// no domain, the Parallelism not worked
	limit := &colly.LimitRule{
		Parallelism:  parallelism,
		Delay:        time.Second,
		RandomDelay:  time.Millisecond * 500,
		DomainRegexp: "lantouzi.com"}
	_ = c.Limit(limit)
	return c
}

func collectorWithContext(c *colly.Collector, ctx context.Context) {
	// We can stop all requests at `OnRequest` callback
	// before send request to HTTP client.
	trans := &contextTransport{
		ctx:   ctx,
		trans: &http.Transport{},
	}
	c.WithTransport(trans)
	// Use custome Transport to cancel all pending requests at HTTP client,
	// which not have chance to stop at OnRequest callback.
	c.OnRequest(func(req *colly.Request) {
		select {
		case <-ctx.Done():
			log.Printf("abort request")
			req.Abort()
		default:
		}
	})
}

func collectorRetryOnError(c *colly.Collector, ctx context.Context) {
	c.OnError(func(resp *colly.Response, err error) {
		retry := 0
		for err != nil && retry < 10 {
			select {
			case <-ctx.Done():
				return
			default:
			}
			err = resp.Request.Retry()
			retry += 1
		}
		if err != nil {
			log.Printf("retry %v err= %v", resp.Request.URL, err)
		}
	})
}

func crawlPage(ctx context.Context, postWebs chan string, stat *Stat) {
	c := newCollector(10)
	collectorWithContext(c, ctx)
	collectorRetryOnError(c, ctx)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		//_ = e.Request.Visit(e.Attr("href"))
		if strings.Contains(e.Text, "项目回款") {
			url := e.Attr("href")
			atomic.AddInt64(&stat.PageRespValidCount, 1)
			if len(url) > 0 {
				atomic.AddInt64(&stat.PostReqCount, 1)
				postWebs <- url
			}
		}
	})

	log.Printf("page start enqueue requests")
	for i := 1; i < 4; i++ {
		url := fmt.Sprintf("https://lantouzi.com/post?page=%v", i)
		err := c.Visit(url)
		atomic.AddInt64(&stat.PageReqCount, 1)
		if err != nil {
			log.Printf("page visit err= %v", err)
		}
	}
	log.Printf("page end enqueue requests")
	c.Wait()
	close(postWebs)
}

func crawPosts(ctx context.Context, postWebs chan string, posts chan *moneyPost, stat *Stat) {
	c := newCollector(50)
	collectorWithContext(c, ctx)
	collectorRetryOnError(c, ctx)

	re := make([]*regexp.Regexp, 0, 4)
	re = append(re, regexp.MustCompile(Pattern1))
	re = append(re, regexp.MustCompile(Pattern2))
	// 目标是搜索多个 p[class=MsoNormal] 拼接起来，也可以是搜索 posts rich-text
	c.OnHTML("div[class=\"posts rich-text\"]", func(e *colly.HTMLElement) {
		if len(e.Text) < 1 {
			return
		}
		if !strings.Contains(e.Text, "项目回款") {
			return
		}
		atomic.AddInt64(&stat.PostRespValidCount, 1)
		for _, r := range re {
			// A Regexp is safe for concurrent use by multiple goroutines,
			// except for configuration methods, such as Longest.
			result := r.FindStringSubmatch(e.Text)
			if result != nil && len(result) > 5 {
				v := newMoneyPost(result[1], result[2], result[3], result[4], result[5])
				posts <- v
				break
			}
		}
	})
websLoop:
	for {
		select {
		case <-ctx.Done():
			break websLoop
		case url, more := <-postWebs:
			if !more {
				break websLoop
			}
			_ = c.Visit(url)
		}
	}
	c.Wait()
	close(posts)
}

func crawlLantouzi(ctx context.Context) {
	stat := &Stat{}
	postWebs := make(chan string, 1000)
	posts := make(chan *moneyPost, 1000)
	wg := &sync.WaitGroup{}
	crawlWg := sync.WaitGroup{}
	crawlWg.Add(1)
	wg.Add(1)
	go func() {
		crawlPage(ctx, postWebs, stat)
		crawlWg.Done()
		wg.Done()
	}()
	wg.Add(1)
	crawlWg.Add(1)
	go func() {
		crawPosts(ctx, postWebs, posts, stat)
		crawlWg.Done()
		wg.Done()
	}()

	statCtx, statCancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		statMon(statCtx, stat)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		writePosts(ctx, posts, stat)
		wg.Done()
	}()
	crawlWg.Wait()
	statCancel()
	wg.Wait()
	log.Printf("stat= %v", stat)
	log.Printf("colly exit")
}

func statMon(ctx context.Context, stat *Stat) {
	tick := time.Tick(time.Second * 5)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-tick:
		}
		log.Printf("stat= %v", stat)
	}
	log.Printf("statMon() exit")
}

func floatlify(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Printf("float err= %v", err)
	}
	return f
}

func beginDay(year, month, day int) time.Time {
	d := time.Time{}
	d.In(time.Now().Location())
	d = d.AddDate(year-1, month-1, day-1)
	return d
}

func timeKey(t time.Time) string {
	return fmt.Sprintf("%v-%02d-%02d", t.Year(), int(t.Month()), t.Day())
}

func writePosts(ctx context.Context, posts chan *moneyPost, stat *Stat) {
	path := "lantouzi-back.csv"
	curTime := time.Now()
	curTime = curTime.Add(-time.Hour * 24)
	endDay := beginDay(curTime.Year(), int(curTime.Month()), int(curTime.Day()))
	fw, _ := os.Create(path)
	w := csv.NewWriter(fw)
	cache := make(map[string]*moneyPost, 0)
postLoop:
	for {
		select {
		case <-ctx.Done():
			break postLoop
		case post, more := <-posts:
			if !more {
				break postLoop
			}
			atomic.AddInt64(&stat.WriterCacheSize, 1)
			cache[post.MapKey()] = post
			for {
				k := timeKey(endDay)
				v, exists := cache[k]
				if !exists {
					break
				}
				delete(cache, k)
				atomic.AddInt64(&stat.WriterCacheSize, -1)
				endDay = endDay.Add(-time.Hour * 24)
				_ = w.Write(v.ToRecord())
				w.Flush()
			}
		}
	}
	// dump all
	_ = w.Write([]string{"-", "-", "-", "-", "-"})
	for _, v := range cache {
		_ = w.Write(v.ToRecord())
	}
	w.Flush()
	_ = fw.Close()
	log.Printf("writePosts() exit")
}
