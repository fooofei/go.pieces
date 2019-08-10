package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// globals
type perfStat struct {
	TcpCnt     int64
	TcpCntOk   int64
	TcpCntFail int64
	TLsCntOk   int64
	TLSCntFail int64
	AddRtnCnt  int64 // not used
	SubRtnCnt  int64 // not used
}
type perfContext struct {
	StatDur     time.Duration
	RoutinesCnt int64
	RAddr       string
	WaitCtx     context.Context
	TlsConf     *tls.Config
	Wg          *sync.WaitGroup
	//
	ErrCh chan error
	stat  perfStat
}

func (ps *perfStat) safeDup() *perfStat {
	newPs := &perfStat{}
	newPs.TcpCnt = atomic.LoadInt64(&ps.TcpCnt)
	newPs.TcpCntOk = atomic.LoadInt64(&ps.TcpCntOk)
	newPs.TcpCntFail = atomic.LoadInt64(&ps.TcpCntFail)
	newPs.TLsCntOk = atomic.LoadInt64(&ps.TLsCntOk)
	newPs.TLSCntFail = atomic.LoadInt64(&ps.TLSCntFail)
	newPs.AddRtnCnt = atomic.LoadInt64(&ps.AddRtnCnt)
	newPs.SubRtnCnt = atomic.LoadInt64(&ps.SubRtnCnt)
	return newPs
}

func (ps *perfStat) sub(b *perfStat) {
	ps.TcpCnt = ps.TcpCnt - b.TcpCnt
	ps.TcpCntOk = ps.TcpCntOk - b.TcpCntOk
	ps.TcpCntFail = ps.TcpCntFail - b.TcpCntFail
	ps.TLsCntOk = ps.TLsCntOk - b.TLsCntOk
	ps.TLSCntFail = ps.TLSCntFail - b.TLSCntFail
	ps.AddRtnCnt = ps.AddRtnCnt - b.AddRtnCnt
	ps.SubRtnCnt = ps.SubRtnCnt - b.SubRtnCnt
}

func (pc *perfContext) nonBlockEnqErr(err error) {
	select {
	case pc.ErrCh <- err:
	default:
	}
}

func deepCopy(src interface{}, dst interface{}) error {
	bytes_, err := json.Marshal(src)
	if err != nil {
		return errors.Wrapf(err, "fail call json.Marshal")
	}
	err = json.Unmarshal(bytes_, dst)
	if err != nil {
		return errors.Wrapf(err, "fail call json.Unmarshal")
	}
	return nil
}

func toLEBytes(v interface{}) []byte {
	var binBuf bytes.Buffer
	err := binary.Write(&binBuf, binary.LittleEndian, v)
	if err != nil {
		panic(err)
	}
	return binBuf.Bytes()
}

func toBEBytes(v interface{}) []byte {
	var binBuf bytes.Buffer
	err := binary.Write(&binBuf, binary.BigEndian, v)
	if err != nil {
		panic(err)
	}
	return binBuf.Bytes()
}

func takeOverCnnClose(waitCtx context.Context, cnn io.Closer) (chan bool, *sync.WaitGroup) {
	noWait := make(chan bool, 1)
	waitGrp := &sync.WaitGroup{}
	waitGrp.Add(1)
	go func() {
		select {
		case <-noWait:
		case <-waitCtx.Done():
		}
		_ = cnn.Close()
		waitGrp.Done()
	}()
	return noWait, waitGrp
}

func boomTls(perfCtx *perfContext, cnn net.Conn) {
	tlsCnn := tls.Client(cnn, perfCtx.TlsConf)
	err := tlsCnn.Handshake()

	if err != nil {
		perfCtx.nonBlockEnqErr(errors.Wrapf(err, "fail tls handshake"))
		atomic.AddInt64(&perfCtx.stat.TLSCntFail, 1)
	} else {
		atomic.AddInt64(&perfCtx.stat.TLsCntOk, 1)
	}
	_ = tlsCnn.Close()
}

func boomRoutine(perfCtx *perfContext) {

	tmo := time.Duration(time.Second * 3)
boomLoop:
	for {
		d := &net.Dialer{Timeout: tmo}

		atomic.AddInt64(&perfCtx.stat.TcpCnt, 1)
		tcpCnn, err := d.DialContext(perfCtx.WaitCtx, "tcp", perfCtx.RAddr)
		if err != nil {
			atomic.AddInt64(&perfCtx.stat.TcpCntFail, 1)
			perfCtx.nonBlockEnqErr(err)
		} else {
			atomic.AddInt64(&perfCtx.stat.TcpCntOk, 1)
			noWait, waitGrp := takeOverCnnClose(perfCtx.WaitCtx, tcpCnn)
			boomTls(perfCtx, tcpCnn)
			close(noWait)
			waitGrp.Wait()
		}
		select {
		case <-perfCtx.WaitCtx.Done():
			break boomLoop
		default:
		}
	}
}

func statRoutine(perfCtx *perfContext) {

	var err error

	statTick := time.NewTicker(perfCtx.StatDur)
	cnt := 0
	oldTime := time.Now()
	oldStat := &perfStat{}
statLoop:
	for {
		select {
		case <-perfCtx.WaitCtx.Done():
			break statLoop
		case <-statTick.C:
			now := time.Now()
			w := &bytes.Buffer{}
			nowStat := perfCtx.stat.safeDup()
			intervalStat := &perfStat{}
			*intervalStat = *nowStat
			intervalStat.sub(oldStat)
			interval := int64(math.Max(1, now.Sub(oldTime).Seconds()))

			log.Printf("stat cnt= %v raddr= %v", cnt, perfCtx.RAddr)
			_, _ = fmt.Fprintf(w, "  tcpCnt %v-%v/%v= %v (ps)\n",
				nowStat.TcpCnt, oldStat.TcpCnt, interval, intervalStat.TcpCnt/interval)
			_, _ = fmt.Fprintf(w, "  tcpCntOk %v-%v/%v= %v (ps)\n",
				nowStat.TcpCntOk, oldStat.TcpCntOk, interval, intervalStat.TcpCntOk/interval)
			_, _ = fmt.Fprintf(w, "  tcpCntFail %v-%v/%v= %v (ps)\n",
				nowStat.TcpCntFail, oldStat.TcpCntFail, interval, intervalStat.TcpCntFail/interval)
			_, _ = fmt.Fprintf(w, "  tlsCntOk %v-%v/%v= %v (ps)\n",
				nowStat.TLsCntOk, oldStat.TLsCntOk, interval, intervalStat.TLsCntOk/interval)
			_, _ = fmt.Fprintf(w, "  tlsCntFail %v-%v/%v= %v (ps)\n",
				nowStat.TLSCntFail, oldStat.TLSCntFail, interval, intervalStat.TLSCntFail/interval)
			fmt.Printf("%v", w.String())

			oldTime = now
			oldStat = nowStat
			cnt += 1
		case err = <-perfCtx.ErrCh:
			log.Printf("hit err= %v", err)
		}

	}

}

func setupSignal(waitCtx context.Context, waitGrp *sync.WaitGroup, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)
	waitGrp.Add(1)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-waitCtx.Done():
		}
		waitGrp.Done()
	}()
}

func beginPerf(perfCtx *perfContext) {
	var i int64
	for i = 0; i < perfCtx.RoutinesCnt; i += 1 {
		perfCtx.Wg.Add(1)
		go func() {
			boomRoutine(perfCtx)
			perfCtx.Wg.Done()
		}()
	}

	log.Printf("all routines started, go stat")

	statRoutine(perfCtx)
}

func tlsStringToInt(ver string) int {
	switch ver {
	case "tls13":
		return tls.VersionTLS13
	case "tls12":
		return tls.VersionTLS12
	case "tls11":
		return tls.VersionTLS11
	case "tls10":
		return tls.VersionTLS10
	case "ssl30":
		return tls.VersionSSL30
	default:
		return -1
	}
}

func init() {
	// enable tls13 default
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")
}

func main() {

	var cancel context.CancelFunc
	perfCtx := &perfContext{}
	var tlsMinVer string
	var tlsMaxVer string
	var interval int
	tlsVers := "ssl30 tls10 tls11 tls12 tls13"

	flag.Int64Var(&perfCtx.RoutinesCnt, "routines", 1, "count of keep running go routines")
	flag.StringVar(&perfCtx.RAddr, "raddr", "127.0.0.1:886", "to perf tcp-ssl addr")
	flag.IntVar(&interval, "interval", 3, "stat interval (sec)")
	flag.StringVar(&tlsMinVer, "tls_min_ver", "tls12", fmt.Sprintf("the min tls version (%v)", tlsVers))
	flag.StringVar(&tlsMaxVer, "tls_max_ver", "tls13", fmt.Sprintf("the max tls version (%v)", tlsVers))
	flag.Parse()
	perfCtx.StatDur = time.Second * time.Duration(interval)
	perfCtx.ErrCh = make(chan error, 10)
	perfCtx.TlsConf = &tls.Config{}
	perfCtx.TlsConf.InsecureSkipVerify = true

	minv := tlsStringToInt(tlsMinVer)
	if minv > 0 {
		perfCtx.TlsConf.MinVersion = uint16(minv)
		log.Printf("set TlsConf.MinVersion= %v", tlsMinVer)
	}
	maxv := tlsStringToInt(tlsMaxVer)
	if maxv > 0 {
		perfCtx.TlsConf.MaxVersion = uint16(maxv)
		log.Printf("set TlsConf.MaxVersion= %v", tlsMaxVer)
	}

	perfCtx.WaitCtx, cancel = context.WithCancel(context.Background())
	perfCtx.Wg = new(sync.WaitGroup)

	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("use routines=%v to raddr= %v\n", perfCtx.RoutinesCnt, perfCtx.RAddr)

	setupSignal(perfCtx.WaitCtx, perfCtx.Wg, cancel)

	beginPerf(perfCtx)

	log.Printf("wait all routines to exit")
	perfCtx.Wg.Wait()
	log.Printf("exit\n")
}
