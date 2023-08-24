package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/miekg/dns"
)

// setup a dns Server

func serveDnsServer(ctx context.Context, srv *Server) {
	// attach request handler func
	for domain, _ := range srv.Records {
		dns.HandleFunc(domain, srv.handleRequest)
	}

	server := &dns.Server{Addr: srv.Addr, Net: srv.Net}
	log.Printf("dns listen at %v\n", srv.Addr)

	var errCh = make(chan error, 1)
	go func() {
		var err = server.ListenAndServe()
		if err != nil {
			errCh <- err
		}
		close(errCh)
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
	<-errCh
}

func main() {
	var srv = &Server{
		Records: map[string]string{"a.b.com.": "127.0.0.1"},
		Addr:    ":53",
		Net:     "udp",
	}
	var ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	serveDnsServer(ctx, srv)
	log.Printf("main exit")
}
