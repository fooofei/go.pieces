package main

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
)

type Server struct {
	Records map[string]*Record // key:domain
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
			log.Printf("Query for %s\n", q.Name)
			var rec, exists = srv.Records[q.Name]
			if !exists {
				continue
			}
			var rr, err = dns.NewRR(fmt.Sprintf("%s A %s", q.Name, rec.Host))
			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
		}
	}
}
