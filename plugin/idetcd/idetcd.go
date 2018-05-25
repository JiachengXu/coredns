//Package idetcd implements a plugin that allow host confiuration config itself without collison by using etcd.
package idetcd

import (
	"context"
	"fmt"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

//Idetcd is a plugin which can configure the cluster without collison.
type Idetcd struct {
	Next plugin.Handler
}

//ServeDNS implements the plugin.Handler interface
func (idetcd Idetcd) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	fmt.Println("hello world!")
	return plugin.NextOrFailure(idetcd.Name(), idetcd.Next, ctx, w, r)
}

//Name implements the Handler interface.
func (idetcd Idetcd) Name() string { return "idetcd" }
