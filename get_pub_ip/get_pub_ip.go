package main

import (
    "container/list"
    "context"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "sort"
    "strings"
    "sync"
    "time"
)

type ParallelHttpCtx struct {
    ResultCh        chan string
    AllResultDoneCh chan struct{}
    Results         *list.List
    Wg              sync.WaitGroup
    WaitCtx         context.Context
}

// do http.get with context.Context
// context can be cancel()
// not care about err, only push resp.body to chan
func fetchIpRoutine(url string, ctx *ParallelHttpCtx) {
    defer ctx.Wg.Done()
    //
    clt := &http.Client{}
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        return
    }
    req = req.WithContext(ctx.WaitCtx)

    resp, err := clt.Do(req)
    if err != nil {
        return
    }
    if resp.StatusCode != 200 {
        return
    }
    b, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return
    }
    v := strings.TrimSpace(string(b))
    ctx.ResultCh <- v
}

// wait all sub routines result
// when all result done before cancel() then notify a chan
// can be cancel()
func waitAllResultRoutine(ctx *ParallelHttpCtx, cnt int) {
    defer ctx.Wg.Done()
    //
    for i := 0; i < cnt; i += 1 {
        select {
        case <-ctx.WaitCtx.Done():
            return
        case v := <-ctx.ResultCh:
            ctx.Results.PushBack(v)
        }
    }
    close(ctx.AllResultDoneCh)
}

// find the most often result
type Pair struct {
    Key   string
    Value int
}
type PairArray []Pair

func (p PairArray) Len() int               { return len(p) }
func (p PairArray) Less(i int, j int) bool { return p[i].Value < p[j].Value }
func (p PairArray) Swap(i int, j int)      { p[i], p[j] = p[j], p[i] }
func getTop(ctx *ParallelHttpCtx) string {
    if ctx.Results.Len() <= 0 {
        return ""
    }
    //
    rm := make(map[string]int, ctx.Results.Len())
    for e := ctx.Results.Front(); e != nil; e = e.Next() {
        v := e.Value.(string)
        rm[v] ++
    }
    ra := make(PairArray, len(rm))
    i := 0
    for k, v := range rm {
        ra[i] = Pair{k, v}
        i++
    }
    sort.Stable(sort.Reverse(ra))
    return ra[0].Key
}

// 1 请求不必要返回 json 类型，省去我们要 decode json 的步骤
// 2 并行发起http请求
// 3 给定超时，有几个返回结果用几个返回结果
//   且如果在超时时间内全部请求得到返回，这将会是更好的场面，
//   我们就不必要一直死等超时，直接取用结果
// 4 没有必要设置捕获 signal 信号，CTRL +C 可以在任意时刻退出，go 保证
//   我们也没有要优雅退出的需求
func main() {
    // no need to use https://api.ipify.org/?format=json
    pubSrvs := [...]string{
        "https://api.ipify.org",
        "https://ip.seeip.org",
        "https://ifconfig.me/ip",
    }
    log.Printf("pid= %v", os.Getpid())
    ctx := new(ParallelHttpCtx)
    ctx.Results = list.New()
    ctx.ResultCh = make(chan string, len(pubSrvs))
    ctx.AllResultDoneCh = make(chan struct{})
    var cancel context.CancelFunc
    ctx.WaitCtx, cancel = context.WithCancel(context.Background())

    for i := 0; i < len(pubSrvs); i += 1 {
        ctx.Wg.Add(1)
        go fetchIpRoutine(pubSrvs[i], ctx)
    }

    ctx.Wg.Add(1)
    go waitAllResultRoutine(ctx, len(pubSrvs))
    // wait timeout or all done
    select {
    case <-time.After(time.Second * 20):
        log.Printf("main timeup cancel it beforehand")
        cancel()
    case <-ctx.AllResultDoneCh:
    }
    //log.Printf("main wait sub routine")
    ctx.Wg.Wait()
    //
    log.Printf("fetch result cnt=%v from %v", ctx.Results.Len(), len(pubSrvs))
    fmt.Printf("The pub ip= %v\n", getTop(ctx))
    log.Printf("main exit")
}
