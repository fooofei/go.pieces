package go_pieces

import (
	"net"
	"testing"
)

// or use "github.com/miekg/dns"

func TestDns1(t *testing.T) {
	names, err := net.LookupAddr("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Logf("no record")
	}
	for _, name := range names {
		t.Logf("%s\n", name)
	}

	r, err := net.LookupIP("localhost")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("r= %v", r)
}
