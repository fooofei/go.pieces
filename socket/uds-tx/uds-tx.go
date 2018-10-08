package main

import (
    "bufio"
    "bytes"
    "fmt"
    "net"
    "os"
    "os/signal"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"
    "runtime"
    "time"
)



type routineStat  struct {
    tx * uint64
    txOk * uint64
    txFail * uint64
    id int
    remoteAddr string
    routineTx uint64
    routineTxOk uint64
    routineTxFail uint64
    stop bool
    stoped bool
    wg * sync.WaitGroup
}

func (stat * routineStat)String() string{
    return fmt.Sprintf("id=%v remoteAddr=%v routineTx=%v ok=%v fail=%v stop=%v stoped=%v",
        stat.id, stat.remoteAddr, stat.routineTx,
        stat.routineTxOk, stat.routineTxFail,
        stat.stop, stat.stoped)
}

func (stat * routineStat)run(){
    defer stat.wg.Done()
    stat.wg.Add(1)

    proto := "uds://"
    remoteAddrS := stat.remoteAddr
    remoteAddrS = strings.TrimPrefix(remoteAddrS,proto)
    pid := os.Getpid()

    localAddrS := fmt.Sprintf("/tmp/go-tx-%v-%v.sock", pid, stat.id)
    if _,err := os.Lstat(localAddrS); !os.IsNotExist(err) {
        os.Remove(localAddrS)
    }
    sockType := "unixgram"


    localAddr,_ := net.ResolveUnixAddr(sockType, localAddrS)
    remoteAddr,_ := net.ResolveUnixAddr(sockType, remoteAddrS)
    conn,err := net.DialUnix(sockType,localAddr,nil)
    //
    fmt.Printf("listen %v addr=%v err=%v\n", conn, localAddr, err)
    //
    if err != nil{
        panic(err)
    }
    defer os.Remove(localAddr.String())
    txBuf := bytes.NewBuffer(make([]byte,0, 1024*2))

    for ;!stat.stop;{
        txBuf.Reset()
        txBuf.WriteString(fmt.Sprintf("msg %v", atomic.LoadUint64(stat.tx)))
        txBytes := txBuf.Bytes()
        //fmt.Printf("tx len=%v\n", len(txBytes))
        rx0,err := conn.WriteToUnix(txBytes,remoteAddr)
        stat.routineTx ++
        atomic.AddUint64(stat.tx, 1)
        if err != nil{
            stat.routineTxFail ++
            atomic.AddUint64(stat.txFail,1)
        }else{
         stat.routineTxOk ++
         atomic.AddUint64(stat.txOk,1)
        }


        _ = rx0
        _ = err
        //fmt.Printf("%v rx from %v %v len(rxBuf)=%v err=%v\n",stat, remoteAddr, rx0, len(txBuf.Bytes()), err)
        //time.Sleep(time.Duration(time.Second*3))
    }
    stat.stoped=true
    fmt.Printf("exit %v\n", stat)
}

type routineStats [] * routineStat
func (stat routineStats) Len() int{
    return len(stat)
}
func (stat routineStats) Swap(i int, j int) {
    stat[i], stat[j] = stat[j], stat[i]
}
func (stat routineStats) Less(i int, j int) bool {
    return stat[i].id < stat[j].id
}
func (stat routineStats) mapId() map[int]*routineStat{
    r := make(map[int]*routineStat)
    for _, s := range stat{
       r[s.id]=s
    }
    return r
}

func mergeStat(oldStat routineStats, newStat routineStats) routineStats {
    oldIdm := oldStat.mapId()
    newIdm := newStat.mapId()

    stat := make(routineStats,0)

    // stop old
    for _, o := range oldStat{
        if _,ok := newIdm[o.id]; !ok{
            fmt.Printf("sub %v\n", o)
            o.stop=true
            for !o.stoped{
                time.Sleep(time.Duration(time.Second))
            }
        }
    }
    // start new
    for _, n := range newStat{
        if _,ok := oldIdm[n.id]; !ok {
            fmt.Printf("add %v\n", n)
            stat = append(stat,n)
            go n.run()
        }else{
            stat = append(stat,oldIdm[n.id])
        }
    }
    return stat
}

func oneTest(){
    sockType := "unixgram"
    pid:=os.Getpid()
    localAddr,_ := net.ResolveUnixAddr(sockType, fmt.Sprintf("/tmp/%d-tx.sock",pid))
    remoteAddr,_ := net.ResolveUnixAddr(sockType, fmt.Sprintf("/tmp/dpdk-rx.sock"))
    os.Remove(localAddr.String())
    //conn,err := net.DialUnix(sockType,localAddr,remoteAddr)
    conn,err := net.DialUnix(sockType,localAddr,nil)
    fmt.Printf("listen %v addr=%v err=%v\n", conn, localAddr, err)
    if err != nil{
        panic(err)
    }
    defer os.Remove(localAddr.String())
    txBuf := bytes.NewBuffer(make([]byte,0, 1024*2))
    cnt := 0

    for{
        txBuf.Reset()
        txBuf.WriteString(fmt.Sprintf("msg %v", cnt))
        txBytes := txBuf.Bytes()
        fmt.Printf("tx len=%v\n", len(txBytes))
        rx0,err := conn.WriteToUnix(txBytes,remoteAddr)
        cnt ++
        
        fmt.Printf("cnt=%v rx from %v %v len(rxBuf)=%v err=%v\n",cnt, remoteAddr, rx0, len(txBuf.Bytes()), err)
        //time.Sleep(time.Duration(time.Second*3))
    }

    fmt.Printf("uds tx exit\n")
}

func getCurrentDir() (executablePath string) {
    _, callerFile, _, _ := runtime.Caller(0)
    executablePath = filepath.Dir(callerFile)
    return executablePath
}

func Map(vs []string, f func(string) string) []string {
    vsm := make([]string, len(vs))
    for i, v := range vs {
        vsm[i] = f(v)
    }
    return vsm
}

func getTasksFromFile(tx * uint64, txOk * uint64, txFail * uint64,
    group * sync.WaitGroup) routineStats {

    stat := make(routineStats, 0)
    curDir := getCurrentDir()
    cfgPath  := filepath.Join(curDir, "tx.cfg")

    file,err := os.Open(cfgPath);
    if err != nil{
        return stat
    }
    defer file.Close()

    sc := bufio.NewScanner(file)
    set := make(map[int]bool)
    for sc.Scan(){
        line := sc.Text()
        if strings.HasPrefix(line, "# uds://"){
            line = strings.TrimPrefix(line, "# ")
            ss := strings.Split(line, ",")
            ss = Map(ss, func(s string) string {
                return strings.TrimSpace(s)
            })
            if len(ss) ==2{
                t := new(routineStat)
                t.remoteAddr = ss[0]
                t.tx = tx
                t.wg = group
                t.txOk = txOk
                t.txFail = txFail
                v,_ := strconv.ParseInt(ss[1],10,64)
                t.id = int(v)
                if _,ok := set[t.id]; !ok{
                    stat = append(stat, t)
                    set[t.id]=true
                }
            }
        }

    }
    sort.Sort(stat)
    return stat
}

func subTime(startTime time.Time, n time.Time) string{
    d := n.Sub(startTime)
    return fmt.Sprintf("%v %v(s)",n.Format("2006-01-02 15:04:05"),
        d.Seconds())
}

func main(){

    cSignal := make(chan os.Signal, 1)
    signal.Notify(cSignal, os.Interrupt, os.Kill)

    var tx uint64
    var txOk uint64
    var txFail uint64
    var wg sync.WaitGroup

    txStat := make(routineStats,0)
    startTime := time.Now()

    serveForever:
    for{
        newStat := getTasksFromFile(&tx,&txOk, &txFail, &wg)
        txStat = mergeStat(txStat,newStat)
        if len(txStat)==0{
            fmt.Printf("- no tasks\n")
        }
        select {
        case c:= <- cSignal:
            fmt.Printf("got signal %v\n", c)
            for _,s := range txStat{
                s.stop=true
            }
            break serveForever
        case <-time.After(time.Duration(time.Second*3)):
        }
    }

    fmt.Printf("sync.WaitGroup wait \n")
    wg.Wait()
    fmt.Printf("main exit tx=%v take %v\n", tx, subTime(startTime,time.Now()))
}
