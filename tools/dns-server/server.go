package main

import (
	"fmt"
	"log"

	"github.com/miekg/dns"
)

type Server struct {
	Records map[string]string // key:domain value ip
	Addr    string
	Net     string
}

func (srv *Server) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	var m = new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		srv.parseQuery(m)
	}
	w.WriteMsg(m)
}

func (srv *Server) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:

			var rec, exists = srv.Records[q.Name]
			if !exists {
				log.Printf("failed query for %s not exists\n", q.Name)
				continue
			}
			var rr, err = dns.NewRR(fmt.Sprintf("%s A %s", q.Name, rec))
			log.Printf("query for %v->%v err=%v\n", q.Name, rec, err)
			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
		}
	}
}
