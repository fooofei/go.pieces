package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/miekg/dns"
)

// setup a dns server

type Record struct {
	Domain string
	Host   string
}

type Conf struct {
	Addr    string
	Net     string
	Records []*Record

	mapRecords map[string]*Record
}

func (c *Conf) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		c.parseQuery(m)
	}

	_ = w.WriteMsg(m)
}

func (c *Conf) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Printf("Query for %s\n", q.Name)
			rec, exists := c.mapRecords[q.Name]
			if exists {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, rec.Host))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}
}

func (c *Conf) ToMap() {
	c.mapRecords = make(map[string]*Record, 0)
	for _, rec := range c.Records {
		c.mapRecords[rec.Domain] = rec
	}
}

func setupServer(conf *Conf) {
	// attach request handler func

	for _, rec := range conf.Records {
		dns.HandleFunc(rec.Domain, conf.handleRequest)
	}

	server := &dns.Server{Addr: conf.Addr, Net: conf.Net}
	log.Printf("dns listen at %v\n", conf.Addr)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}

func main() {

	conf := &Conf{}
	fp, _ := os.Executable()
	fp = filepath.Dir(fp)
	fp = filepath.Join(fp, "conf.toml")
	_, err := toml.DecodeFile(fp, conf)
	if err != nil {
		log.Fatal(err)
	}
	conf.ToMap()

	setupServer(conf)

	log.Printf("main exit")
}
