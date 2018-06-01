package idetcd

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

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
	idetc, err := idetcdParse(c)
	if err != nil {
		return plugin.Error("idetcd", err)
	}
	if c.NextArg() {
		return plugin.Error("idetcd", c.ArgErr())
	}
	localIP := getLocalIPAddress()

	//find id for node
	var i = 1
	for {
		name := idetc.pattern + strconv.Itoa(i)
		_, err := idetc.get(name)
		if etcdc.IsKeyNotFound(err) {
			idetc.id = name
			idetc.set(name, localIP.String())
			break
		}
		i++
	}

	//update the record in the etcd
	renewTicker := time.NewTicker(defaultTTL / 2 * time.Second)
	go func() {
		for {
			select {
			case <-renewTicker.C:
				idetc.set(idetc.id, localIP.String())
			}
		}
	}()

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		idetc.Next = next
		return idetc
	})
	return nil
}

func getLocalIPAddress() net.IP {
	var localIP net.IP
	interfaces, _ := net.Interfaces()
	for _, inter := range interfaces {
		if inter.Name == "eth0" || inter.Name == "en0" {
			addrs, _ := inter.Addrs()
			for _, addr := range addrs {
				localIP = net.ParseIP(strings.Split(addr.String(), "/")[0])
			}
		}
	}
	return localIP
}

func idetcdParse(c *caddy.Controller) (*Idetcd, error) {
	idetc := Idetcd{
		Ctx: context.Background(),
	}
	var (
		endpoints = []string{defaultEndpoint}
		pattern   = defaultPattern
	)
	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "endpoint":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return &Idetcd{}, c.ArgErr()
				}
				endpoints = args
			case "pattern":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return &Idetcd{}, c.ArgErr()
				}
				pattern = args[0]
			}
		}
	}
	client, err := newEtcdClient(endpoints)
	if err != nil {
		return &Idetcd{}, err
	}
	idetc.endpoints = endpoints
	idetc.Client = client
	idetc.pattern = pattern
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

const (
	defaultEndpoint = "http://localhost:2379"
	defaultPattern  = "worker"
	defaultTTL      = 10
)
