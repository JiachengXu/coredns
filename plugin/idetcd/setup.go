package idetcd

import (
	"context"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	etcdc "github.com/coreos/etcd/client"

	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("idetcd", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	c.Next() //idetcd
	idetc, err := idetcdParse(c)
	if err != nil {
		return plugin.Error("idetcd", err)
	}
	if c.NextArg() {
		return plugin.Error("idetcd", c.ArgErr())
	}

	idetc.set("Test", "127.0.0.1")
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		idetc.Next = next
		return idetc
	})
	return nil
}

func idetcdParse(c *caddy.Controller) (*Idetcd, error) {
	idetc := Idetcd{
		Ctx: context.Background(),
	}
	var (
		endpoints = []string{defaultEndpoint}
	)
	client, err := newEtcdClient(endpoints)
	if err != nil {
		return &Idetcd{}, err
	}
	idetc.endpoints = endpoints
	idetc.Client = client
	return &idetc, nil

}

func newEtcdClient(endpoints []string) (etcdc.KeysAPI, error) {
	etcdCfg := etcdc.Config{
		Endpoints: endpoints,
	}
	cli, err := etcdc.New(etcdCfg)
	if err != nil {
		return nil, err
	}
	return etcdc.NewKeysAPI(cli), nil
}

const defaultEndpoint = "http://localhost:2379"
