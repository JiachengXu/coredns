// +build idetcd

package test

import (
	"strconv"
	"testing"

	"github.com/coredns/coredns/plugin/proxy"
	"github.com/coredns/coredns/plugin/test"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

func TestLookupNodesRR(t *testing.T) {
	corefiles := generateCorefiles(5)
	var udpp string

	for i, corefile := range corefiles {
		node, udp, _, err := CoreDNSServerAndPorts(corefile)
		if err != nil {
			t.Fatalf("Could not get CoreDNS serving instance: %s,%d", err, i)
		}
		if i == len(corefiles)-1 {
			udpp = udp
		}
		defer node.Stop()
	}

	p := proxy.NewLookup([]string{udpp}) // use udp port from the server
	state := request.Request{W: &test.ResponseWriter{}, Req: new(dns.Msg)}
	for i := 0; i < len(corefiles); i++ {
		resp, err := p.Lookup(state, "worker"+strconv.Itoa(i+1)+".tf.local.", dns.TypeA)
		if err != nil {
			t.Fatalf("Expected to receive reply, but didn't: %v", err)
		}
		if len(resp.Answer) == 0 {
			t.Fatalf("Expected to at least one RR in the answer section, got none")
		}
		if resp.Answer[0].Header().Rrtype != dns.TypeA {
			t.Errorf("Expected RR to A, got: %d", resp.Answer[0].Header().Rrtype)
		}
		if resp.Answer[0].(*dns.A).A.String() != "192.168.0.73" {
			t.Errorf("Expected 192.168.0.73, got: %s", resp.Answer[0].(*dns.A).A.String())
		}
	}

}

func generateCorefiles(numNode int) []string {
	var corefiles []string
	limit := strconv.Itoa(numNode)
	for i := 0; i < numNode; i++ {
		port := strconv.Itoa(1053 + i)
		corefile := `.:` + port + ` {
			idetcd {
				endpoint http://localhost:2379 
				pattern worker{{.ID}}.tf.local.
				limit ` + limit + `
			}
		}`
		corefiles = append(corefiles, corefile)
	}
	return corefiles
}
