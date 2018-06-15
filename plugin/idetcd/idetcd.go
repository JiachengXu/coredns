//Package idetcd implements a plugin that allow host confiuration config itself without collison by using etcd.
package idetcd

import (
	"context"
	"fmt"
	"net"
	"text/template"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	etcdc "github.com/coreos/etcd/client"
	"github.com/miekg/dns"
)

//Idetcd is a plugin which can configure the cluster without collison.
type Idetcd struct {
	Next      plugin.Handler
	Ctx       context.Context
	Client    etcdc.KeysAPI
	endpoints []string
	pattern   *template.Template
	ID        int
	limit     int
}

//ServeDNS implements the plugin.Handler interface
func (idetcd *Idetcd) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	a := new(dns.Msg)
	a.SetReply(r)
	a.Authoritative = true
	qname := state.Name()
	fmt.Println(qname)
	resp, _ := idetcd.get(qname)
	ip := resp.Node.Value
	var rr dns.RR
	switch state.QType() {
	case dns.TypeA:
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeA, Class: state.QClass()}
		rr.(*dns.A).A = net.ParseIP(ip).To4()

	}
	a.Answer = []dns.RR{rr}
	w.WriteMsg(a)
	return plugin.NextOrFailure(idetcd.Name(), idetcd.Next, ctx, w, r)
}

//set is a wrapper for client.Set
func (idetcd *Idetcd) set(key string, value string, setOptions etcdc.SetOptions) (*etcdc.Response, error) {

	ctx, cancel := context.WithTimeout(idetcd.Ctx, 5*time.Second)
	defer cancel()
	r, err := idetcd.Client.Set(ctx, key, value, &setOptions)
	if err != nil {
		return r, err
	}
	return r, nil
}

// get is a wrapper for client.Get
func (idetcd *Idetcd) get(key string) (*etcdc.Response, error) {
	ctx, cancel := context.WithTimeout(idetcd.Ctx, 5*time.Second)
	defer cancel()
	r, err := idetcd.Client.Get(ctx, key, nil)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//Name implements the Handler interface.
func (idetcd *Idetcd) Name() string { return "idetcd" }
