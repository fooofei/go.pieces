package main

import (
    "bytes"
    "crypto/tls"
    "encoding/binary"
    "encoding/hex"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "math"
    "net"
    "os"
    "sync"
    "sync/atomic"
    "time"
    "os/signal"
    "syscall"
)

const (
    VPN_PKT_MARK = 0xFEFCEFBD
    VPN_PKT_TYPE_AUTH = 1
    VPN_PKT_TYPE_PLD = 2
    VPN_PKT_TYPE_KEEPALIVE = 3
    VPN_PKT_TYPE_CNN_STATUS  = 4

    VPN_APPKEYID_MAX_SIZE = 0x20
    VPN_TERMID_MAX_SIZE = 0x24
    VPN_RS_STATUS_CONNECT = 1
    VPN_RS_STATUS_CLOSE = 2
    VPN_RS_STATUS_CLOSED = 3

    myappkeyid = "\x36\x62\x61\x37\x32\x34\x64\x35\x33\x31\x33\x30\x34\x34\x39" +
        "\x34\x39\x65\x65\x66\x37\x64\x32\x35\x66\x61\x65\x34\x32\x62\x35\x32"
    myappkey="d1406b81635849630c9b2e40884296e2"
    mytermid="\x31\x32\x33\x34\x35\x36\x37\x38\x39\x30\x61\x62\x63\x64\x65" +
        "\x66\x31\x32\x33\x34\x35\x36\x37\x38\x39\x30\x61\x62\x63\x64\x65\x66\x66\x66\x66\x66"

)
// globals
var raddr string
var routines int
var interval int

type VpnHead struct {
    mark uint32
    type_ uint8
    length uint16
}
type VpnAuth struct {
    appkeyid [32]byte
    uuid [36]byte
    iv [16]byte
    code [32]byte
}
type Context struct {
    // cnns
    tcpCnt uint64
    tcpCntOk uint64
    tcpCntFail uint64
    hitErrs chan string
    //
    rtnCnt int32
    rtnStopCnt int32
    sigCh chan os.Signal
    rtnStop bool
    //
    wg sync.WaitGroup
    tlsConf tls.Config
    pid int
}

func copyContext(src * Context, dst * Context){
    dst.tcpCnt = src.tcpCnt
    dst.tcpCntOk = src.tcpCntOk
    dst.tcpCntFail = src.tcpCntFail
    //
    dst.rtnCnt = src.rtnCnt
    dst.rtnStop = src.rtnStop
    dst.rtnStopCnt = src.rtnStopCnt
}
func subContext(a * Context, b * Context, c * Context){
    c.tcpCnt = a.tcpCnt - b.tcpCnt
    c.tcpCntOk = a.tcpCntOk - b.tcpCntOk
    c.tcpCntFail = a.tcpCntFail - b.tcpCntFail
    //
    c.rtnCnt = a.rtnCnt - b.rtnCnt
    //c.rtnStop = a.rtnStop - b.rtnStop
    c.rtnStopCnt = a.rtnStopCnt - b.rtnStopCnt
}

func tlsHexView( v interface{}) string {
    var binBuf bytes.Buffer
    err := binary.Write(&binBuf, binary.LittleEndian, v)
    if err != nil{
        panic(err)
    }


    //return hex.EncodeToString(binBuf.Bytes())
    return hex.Dump(binBuf.Bytes())
}

func DeepCopy(src interface {}, dst interface{}) error {
    if dst == nil {
        return fmt.Errorf("dst cannot be nil")
    }
    if src == nil {
        return fmt.Errorf("src cannot be nil")
    }
    bytes_, err := json.Marshal(src)
    if err != nil {
        return fmt.Errorf("Unable to marshal src: %s", err)
    }
    err = json.Unmarshal(bytes_, dst)
    if err != nil {
        return fmt.Errorf("Unable to unmarshal into dst: %s", err)
    }
    return nil
}

func toLEBytes(v interface {})  []byte {
    var binBuf bytes.Buffer
    err := binary.Write(&binBuf, binary.LittleEndian, v)
    if err != nil{
        panic(err)
    }
    return binBuf.Bytes()
}

func toBEBytes(v interface{}) []byte {
    var binBuf bytes.Buffer
    err := binary.Write(&binBuf, binary.BigEndian, v)
    if err != nil{
        panic(err)
    }
    return binBuf.Bytes()
}


func init(){
    flag.IntVar(&routines, "routines", 1, "go routines cnt")
    flag.StringVar(&raddr, "raddr", "127.0.0.1:886", "tcp-ssl raddr")
    flag.IntVar(&interval, "interval", 3, "stat interval")
}

func cnnRoutine(ctx * Context){
    ctx.wg.Add(1)

    tmo := time.Duration(time.Second * 3)
    atomic.AddInt32(&ctx.rtnCnt, 1)

    authBytes := make([]byte,10)
    _ = authBytes

    hdr := VpnHead{
        mark:VPN_PKT_MARK,
        length:10,
        type_:VPN_PKT_TYPE_AUTH,
    }
    authBytes = append(authBytes, toBEBytes(hdr)...)

    for !ctx.rtnStop {
        // only tcp connect
        d := &net.Dialer{Timeout: tmo}
        conn, err := tls.DialWithDialer(d, "tcp", raddr, &ctx.tlsConf)
        atomic.AddUint64(&ctx.tcpCnt, 1)

        if err != nil {
            atomic.AddUint64(&ctx.tcpCntFail, 1)
            select {
            case ctx.hitErrs <- fmt.Sprintf("%v", err) :
            default:
            }
        }

        if conn != nil{
            // write auth
            _, _ = conn.Write(authBytes)
            //
            _ = conn.CloseWrite()
            _ = conn.Close()
            atomic.AddUint64(&ctx.tcpCntOk, 1)
        }
    }
    atomic.AddInt32(&ctx.rtnStopCnt, 1)
    ctx.wg.Done()
}

func statRoutine(ctx * Context) {

    var hitCtx Context
    var nowCtx Context
    var itvCtx Context
    var hitTime time.Time
    var nowTime time.Time
    var elapse uint64


    itv := interval
    cnt := 0
    hitTime = time.Now()
    ctx.wg.Add(1)

    for !ctx.rtnStop {
        var buf bytes.Buffer
        // get value
        time.Sleep(time.Second * time.Duration(itv))
        nowTime = time.Now()
        copyContext(ctx, &nowCtx)
        // sub value
        log.Printf("hit stat cnt= %v pid= %v raddr= %v", cnt, ctx.pid, raddr)
        subContext(&nowCtx, &hitCtx, &itvCtx)
        elapse = uint64(math.Max(float64(1), float64(nowTime.Sub(hitTime).Seconds())))
        // calc value
        buf.WriteString(fmt.Sprintf("  tcpCnt %v-%v/%v=%.3f\n",
            nowCtx.tcpCnt, hitCtx.tcpCnt, elapse, float64(nowCtx.tcpCnt-hitCtx.tcpCnt)/float64(elapse)) )
        buf.WriteString(fmt.Sprintf("  tcpCntOk %v-%v/%v=%.3f\n",
            nowCtx.tcpCntOk, hitCtx.tcpCntOk, elapse, float64(nowCtx.tcpCntOk-hitCtx.tcpCntOk)/float64(elapse)) )
        buf.WriteString(fmt.Sprintf("  tcpCntFail %v-%v/%v=%.3f\n",
            nowCtx.tcpCntFail, hitCtx.tcpCntFail, elapse, float64(nowCtx.tcpCntFail-hitCtx.tcpCntFail)/float64(elapse)) )
        buf.WriteString(fmt.Sprintf("  rtnCnt %v-%v/%v=%.3f\n",
            nowCtx.rtnCnt, hitCtx.rtnCnt, elapse, float64(nowCtx.rtnCnt-hitCtx.rtnCnt)/float64(elapse)) )
        buf.WriteString(fmt.Sprintf("  rtnStopCnt %v-%v/%v=%.3f\n",
            nowCtx.rtnStopCnt, hitCtx.rtnStopCnt, elapse, float64(nowCtx.rtnStopCnt-hitCtx.rtnStopCnt)/float64(elapse)) )
        buf.WriteString(fmt.Sprintf("  rtnStop=%v\n", ctx.rtnStop))
        select {
        case err,ok := <- ctx.hitErrs:
            if ok {
                buf.WriteString(fmt.Sprintf("  err= %v\n", err))
            }
        default:
        }
        fmt.Printf("%v", buf.String())
        //
        copyContext(&nowCtx, &hitCtx)
        hitTime = nowTime
        cnt += 1
    }

    ctx.wg.Done()
}


func main(){

    flag.Parse()
    //
    log.Printf("use routines=%v to raddr= %v\n" ,routines, raddr)
    //
    


    ibuf := make([]byte, 128*1024)
    _ = ibuf

    hdr := VpnHead{
        mark:VPN_PKT_MARK,
        type_: VPN_PKT_TYPE_AUTH,
        length:0,
    }
    _ = hdr
    ctx := Context{}
    ctx.hitErrs = make(chan string, 3)
    ctx.tlsConf = tls.Config{InsecureSkipVerify:true}
    ctx.sigCh = make(chan os.Signal, 1)
    signal.Notify(ctx.sigCh, os.Interrupt)
    signal.Notify(ctx.sigCh, syscall.SIGTERM)
    ctx.pid = os.Getpid()
    log.Printf("start routines")

    for i :=0; i< routines; i+= 1 {
        go cnnRoutine(&ctx)
    }

    log.Printf("all routines started, go stat")
    
    go func(ctx * Context){
        ctx.wg.Add(1)
        <- ctx.sigCh
        ctx.rtnStop=true
        ctx.wg.Done()
        log.Printf("[!] stop all")
    }(&ctx)
    statRoutine(&ctx)
    
    
    // wait close
    log.Printf("wait all routines to exit")
    ctx.wg.Wait()
    log.Printf("exit\n")
}
