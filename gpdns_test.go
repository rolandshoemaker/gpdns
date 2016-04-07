package gpdns

import (
	"net/http"
	"testing"

	"github.com/miekg/dns"
)

func TestGPDNS(t *testing.T) {
	g := Client{"", new(http.Client)}
	m := new(dns.Msg)
	m.SetQuestion("amazon.com", dns.TypeA)
	_, _, err := g.Exchange(m, "")
	if err != nil {
		t.Fatalf("GPDNS.Exchange failed: %s", err)
	}
}
