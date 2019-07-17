package main

import (
    "container/list"
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "sort"
    "strings"
    "sync"
    "time"
)

type parallelHttpCtx struct {
    ResultCh        chan *workItem
    AllResultDoneCh chan bool
    Results         *list.List
    Wg              sync.WaitGroup
    WaitCtx         context.Context
    WaitTimeout     int
    WaitTimeoutDur  time.Duration
}

// IpGetter defines get ip from url response text
type IpGetter func(r io.Reader) (string, error)

type workItem struct {
    Uri      string
    IpGetter IpGetter
    Result   string
    Take     time.Duration
}

// do http.get with context.Context
// context can be cancel()
// not care about err, only push result to chan
func fetchIpRoutine(wk *workItem, ctx *parallelHttpCtx) {
    defer ctx.Wg.Done()
    defer func() {
        // always push result
        ctx.ResultCh <- wk
    }()
    //
    wk.Take = time.Hour // default is Hour
    start := time.Now()
    clt := &http.Client{}
    req, err := http.NewRequest(http.MethodGet, wk.Uri, nil)
    if err != nil {
        return
    }
    req = req.WithContext(ctx.WaitCtx)

    resp, err := clt.Do(req)
    if err != nil {
        return
    }
    defer func() {
        _ = resp.Body.Close()
    }()
    wk.Take = time.Since(start)
    if resp.StatusCode != 200 {
        return
    }
    v, err := wk.IpGetter(resp.Body)
    if err != nil {
        return
    }
    wk.Result = strings.TrimSpace(string(v))
}

// wait all sub routines result
// when all result done before cancel() then notify a chan
// can be cancel()
func waitAllResultRoutine(ctx *parallelHttpCtx, cnt int) {
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

// Pair help find the most often result
type Pair struct {
    Key   string
    Value int
}
type PairArray []Pair

func (p PairArray) Len() int               { return len(p) }
func (p PairArray) Less(i int, j int) bool { return p[i].Value < p[j].Value }
func (p PairArray) Swap(i int, j int)      { p[i], p[j] = p[j], p[i] }
func getTop(ctx *parallelHttpCtx) string {
    const defaultr = "0.0.0.1"
    if ctx.Results.Len() <= 0 {
        return defaultr
    }
    //
    rm := make(map[string]int, ctx.Results.Len())
    for e := ctx.Results.Front(); e != nil; e = e.Next() {
        v := e.Value.(*workItem)
        if v.Result != "" {
            rm[v.Result] ++
        }
    }
    if len(rm) <= 0 {
        return defaultr
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

// http://ifcfg.cn/echo
func getIpInJsonIP(r io.Reader) (string, error) {
    b, err := ioutil.ReadAll(r)
    if err != nil {
        return "", err
    }
    m := make(map[string]interface{})
    _ = json.Unmarshal(b, &m)
    ip, ok := m["ip"].(string)
    if ok {
        return ip, nil
    }
    return "", fmt.Errorf("not found ip in %v", string(b))
}

// GetIpInJsonIpIp parse result for https://ip.nf/me.json
func getIpInJsonIPIP(r io.Reader) (string, error) {
    b, err := ioutil.ReadAll(r)
    if err != nil {
        return "", err
    }
    m := make(map[string]interface{})
    _ = json.Unmarshal(b, &m)
    ipObj, ok := m["ip"].(map[string]interface{})
    if ok {
        ip, ok := ipObj["ip"].(string)
        if ok {
            return ip, nil
        }
    }
    return "", fmt.Errorf("not found ip in %v", string(b))
}

// GetIpInJsonQuery parse result for http://ip-api.com/json
func getIpInJsonQuery(r io.Reader) (string, error) {
    b, err := ioutil.ReadAll(r)
    if err != nil {
        return "", err
    }
    m := make(map[string]interface{})
    _ = json.Unmarshal(b, &m)
    ip, ok := m["query"].(string)
    if ok {
        return ip, nil
    }
    return "", fmt.Errorf("not found ip in %v", string(b))
}

// GetIpInJsonYourFuck parse result for https://wtfismyip.com/json
func getIpInJsonYourFuck(r io.Reader) (string, error) {
    b, err := ioutil.ReadAll(r)
    if err != nil {
        return "", err
    }
    m := make(map[string]interface{})
    _ = json.Unmarshal(b, &m)
    ip, ok := m["YourFuckingIPAddress"].(string)
    if ok {
        return ip, nil
    }
    return "", fmt.Errorf("not found ip in %v", string(b))
}

// GetIpInPlainText return ip from plain text
// https://api.ipify.org
// https://ip.seeip.org
// https://ifconfig.me/ip
// https://ifconfig.co/ip
func getIpInPlainText(r io.Reader) (string, error) {
    b, err := ioutil.ReadAll(r)
    if err != nil {
        return "", err
    }
    v := strings.TrimSpace(string(b))
    return v, nil
}

// http://ip.taobao.com/service/getIpInfo2.php?ip=myip
func getIpInJsonTaobao(r io.Reader) (string, error) {
    b,err := ioutil.ReadAll(r)
    if err != nil {
        return "",err
    }
    m := make(map[string]interface{})
    _ = json.Unmarshal(b, &m)
    if data,ok := m["data"].(map[string]interface{}) ;ok {
        if ip,ok := data["ip"].(string) ; ok {
            return ip,nil
        }
    }
    return "", fmt.Errorf("not found ip in %v", string(b))
}

// 关于获取自己的 IP 这个需求，有使用 DNS 的方式的，命令为
// dig +short TXT o-o.myaddr.l.google.com @114.114.114.114
// 但是获取到的IP 与 http 获取到的不同
// https://unix.stackexchange.com/questions/22615/how-can-i-get-my-external-ip-address-in-a-shell-script
// https://poplite.xyz/post/2018/05/19/how-to-get-your-public-ip-by-dns-lookup.html

// 1 并行发起http请求
// 2 给定超时，有几个返回结果用几个返回结果
//   且如果在超时时间内全部请求得到返回，这将会是更好的场面，
//   我们就不必要一直死等超时，直接取用结果
// 3 没有必要设置捕获 signal 信号，CTRL +C 可以在任意时刻退出，go 保证
//   我们也没有要优雅退出的需求
// https://api.myip.com/ 美国
// https://myexternalip.com/json 美国
// https://ipapi.co/json 美国
// https://ident.me/.json 英国
// https://get.geojs.io/v1/ip.json 美国
func main() {
    ctx := new(parallelHttpCtx)
    var cancel context.CancelFunc
    //
    flag.IntVar(&ctx.WaitTimeout, "wait", 3, "wait for timeout seconds")
    flag.Parse()
    // no need to use https://api.ipify.org/?format=json
    pubSrvs := &[...]workItem{
        //{Uri: "https://ip.nf/me.json", IpGetter: getIpInJsonIPIP},
        {Uri: "http://ip-api.com/json", IpGetter: getIpInJsonQuery},
        //{Uri: "https://wtfismyip.com/json", IpGetter: getIpInJsonYourFuck},
        {Uri: "https://api.ipify.org", IpGetter: getIpInPlainText},
        {Uri: "https://ip.seeip.org", IpGetter: getIpInPlainText},
        //{Uri: "https://ifconfig.me/ip", IpGetter: getIpInPlainText},
        //{Uri: "https://ifconfig.co/ip", IpGetter: getIpInPlainText},
        // taobao 的服务不稳定
        {Uri: "http://ip.taobao.com/service/getIpInfo2.php?ip=myip",
            IpGetter: getIpInJsonTaobao},
        //{Uri: "http://members.3322.org/dyndns/getip",IpGetter: getIpInPlainText},
        {Uri: "http://ip.cip.cc",IpGetter: getIpInPlainText},
        {Uri: "https://api.ip.sb/ip",IpGetter: getIpInPlainText},
        {Uri: "http://ifcfg.cn/echo",IpGetter: getIpInJsonIP},
        {Uri: "http://eth0.me",IpGetter: getIpInPlainText},
        {Uri: "http://ip.360.cn/IPShare/info",IpGetter: getIpInJsonIP},
    }
    // log init
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
    // ctx init
    ctx.WaitTimeoutDur = time.Second * time.Duration(ctx.WaitTimeout)
    ctx.Results = list.New()
    ctx.ResultCh = make(chan *workItem, len(pubSrvs))
    ctx.AllResultDoneCh = make(chan bool)
    ctx.WaitCtx, cancel = context.WithCancel(context.Background())
    //
    log.Printf("do work")
    for i := 0; i < len(pubSrvs); i += 1 {
        ctx.Wg.Add(1)
        go fetchIpRoutine(&pubSrvs[i], ctx)
    }

    ctx.Wg.Add(1)
    go waitAllResultRoutine(ctx, len(pubSrvs))
    // wait timeout or all done
    select {
    case <-time.After(ctx.WaitTimeoutDur):
        log.Printf("main timeup, cancel it beforehand")
        cancel()
    case <-ctx.AllResultDoneCh:
    }
    //log.Printf("main wait sub routine")
    ctx.Wg.Wait()
    //
    log.Printf("fetch result cnt= %v from %v", ctx.Results.Len(), len(pubSrvs))
    //fmt.Printf("The pub ip= %v\n", getTop(ctx))
    for i := 0; i < len(pubSrvs); i += 1 {
        log.Printf("%v -> %v take %.2f(s)",
            pubSrvs[i].Uri, pubSrvs[i].Result,  pubSrvs[i].Take.Seconds())
    }
    fmt.Printf("%v", getTop(ctx))
    log.Printf("main exit")
}
