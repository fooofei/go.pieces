package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/miekg/dns"
)

// setup a dns Server

func serveDnsServer(ctx context.Context, srv *Server) {
	// attach request handler func
	for _, rec := range srv.Records {
		dns.HandleFunc(rec.Domain, srv.handleRequest)
	}

	server := &dns.Server{Addr: srv.Addr, Net: srv.Net}
	log.Printf("dns listen at %v\n", srv.Addr)

	var errCh = make(chan error, 1)
	go func() {
		var err = server.ListenAndServe()
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Printf("got error from listenAndServe %v", err)
		}
	case <-ctx.Done():
	}
	var subCtx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	server.ShutdownContext(subCtx)
}

func main() {
	var conf = &Config{}
	fp, _ := os.Executable()
	fp = filepath.Dir(fp)
	fp = filepath.Join(fp, "conf.toml")
	var _, err = toml.DecodeFile(fp, conf)
	if err != nil {
		panic(err)
	}
	var srv = &Server{Records: conf.ToMap(), Addr: conf.Addr, Net: conf.Net}
	var ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	serveDnsServer(ctx, srv)
	log.Printf("main exit")
}
