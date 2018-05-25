//Package idetcd implements a plugin that allow host confiuration config itself without collison by using etcd.
package idetcd

import (
	"context"
	"fmt"
	"time"

	"github.com/coredns/coredns/plugin"
	etcdc "github.com/coreos/etcd/client"
	"github.com/miekg/dns"
)

//Idetcd is a plugin which can configure the cluster without collison.
type Idetcd struct {
	Next      plugin.Handler
	Ctx       context.Context
	Client    etcdc.KeysAPI
	endpoints []string
}

//ServeDNS implements the plugin.Handler interface
func (idetcd *Idetcd) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	//state := request.Request{W: w, Req: r}

	resp, _ := idetcd.get("Test")
	fmt.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
	//a := new(dns.Msg)
	//a.SetReply(r)
	//a.Authoritative = true
	//w.WriteMsg(a)
	return plugin.NextOrFailure(idetcd.Name(), idetcd.Next, ctx, w, r)
}

//set is a wrapper for client.Set
func (idetcd *Idetcd) set(key string, value string) (*etcdc.Response, error) {

	ctx, cancel := context.WithTimeout(idetcd.Ctx, 5*time.Second)
	defer cancel()
	r, err := idetcd.Client.Set(ctx, key, value, nil)
	if err != nil {
		return nil, err
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
