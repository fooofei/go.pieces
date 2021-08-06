package sshttp

import (
	"io"
	"log"
	"sync"
)

func PipeReaderWriter(app io.ReadWriter, tun io.ReadWriter) {
	var err error

	waitGrp := &sync.WaitGroup{}
	waitGrp.Add(1)
	go func() {
		_,err = io.Copy(app, tun)
		waitGrp.Done()
		log.Printf("err in tun->app err= %v", err)
	}()
	waitGrp.Add(1)
	go func() {
		_,err = io.Copy(tun,app)
		waitGrp.Done()
		log.Printf("err in app->tun err= %v", err)
	}()
	waitGrp.Wait()
}