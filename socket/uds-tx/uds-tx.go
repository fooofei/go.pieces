package main

import (
    "bufio"
    "bytes"
    "fmt"
    "net"
    "os"
    "os/signal"
    "path/filepath"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"
    "runtime"
    //"time"
)



type routineStat  struct {
    tx * uint64
    id int
    remoteAddr string
    routineTx uint64
    stop bool
    stoped bool
    wg * sync.WaitGroup
}

func (stat * routineStat)String() string{
    return fmt.Sprintf("id=%v remoteAddr=%v routineTx=%v stop=%v stoped=%v",
        stat.id, stat.remoteAddr, stat.routineTx, stat.stop, stat.stoped)
}

func (stat * routineStat)run(){
    defer stat.wg.Done()
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
        fmt.Printf("tx len=%v\n", len(txBytes))
        rx0,err := conn.WriteToUnix(txBytes,remoteAddr)
        stat.routineTx ++
        atomic.AddUint64(stat.tx, 1)

        fmt.Printf("%v rx from %v %v len(rxBuf)=%v err=%v\n",stat, remoteAddr, rx0, len(txBuf.Bytes()), err)
        //time.Sleep(time.Duration(time.Second*3))
    }

    fmt.Printf("%v exit\n", stat)
    stat.stoped=true
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

func getTasksFromFile() routineStats {

    stat := make(routineStats, 0)
    curDir := getCurrentDir()
    cfgPath  := filepath.Join(curDir, "tx.cfg")

    file,err := os.Open(cfgPath);
    if err != nil{
        return stat
    }
    defer file.Close()

    sc := bufio.NewScanner(file)
    var set map[int] bool
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
                v,_ := strconv.ParseInt(ss[1],10,64)
                t.id = int(v)
                if _,ok := set[t.id]; !ok{
                    stat = append(stat, t)
                    set[t.id]=true
                }
            }
        }

    }
    return stat

}

func main(){

    cSignal := make(chan os.Signal, 1)
    signal.Notify(cSignal, os.Interrupt, os.Kill)

    var tx uint64


    txStat := getTasksFromFile()

    for idx, stat := range txStat{
        fmt.Printf("%v %v\n", idx, stat)
    }
    fmt.Printf("main exit tx=%v\n", tx)
}
