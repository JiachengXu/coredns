package idetcd

import (
	"bytes"
	"context"
	"net"
	"strconv"
	"strings"
	"text/template"
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
	var namebuf bytes.Buffer
	var name string
	for i <= idetc.limit {
		idetc.ID = i
		idetc.pattern.Execute(&namebuf, idetc)
		name = namebuf.String()
		_, err := idetc.get(name)
		if etcdc.IsKeyNotFound(err) {
			idetc.set(name, localIP.String())
			break
		}
		i++
		namebuf.Reset()
	}

	if i > idetc.limit {
		return plugin.Error("idetcd", c.ArgErr())
	}

	//update the record in the etcd
	renewTicker := time.NewTicker(defaultTTL / 2 * time.Second)
	go func() {
		for {
			select {
			case <-renewTicker.C:
				idetc.set(namebuf.String(), localIP.String())
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
		pattern   = template.New("idetcd")
		limit     = defaultLimit
		err       error
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
				pattern, err = pattern.Parse(args[0])
				if err != nil {
					return &Idetcd{}, c.ArgErr()
				}
			case "limit":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return &Idetcd{}, c.ArgErr()
				}
				limit, err = strconv.Atoi(args[0])
				if err != nil {
					return &Idetcd{}, c.ArgErr()
				}
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
	idetc.limit = limit
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
	defaultTTL      = 10
	defaultLimit    = 10
)
