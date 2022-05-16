package multicluster_gw

import (
	"net"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

const pluginName = "multicluster_gw"

var (
	Gateway_ip4 net.IP
	Gateway_ip6 net.IP
)

// init registers this plugin.
func init() { plugin.Register(pluginName, setup) }

// setup is the function that gets called when the config parser see the token pluginName. Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is pluginName.
func setup(c *caddy.Controller) error {

	multiCluster, err := ParseStanza(c)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	// #TODO init the controller here

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return MultiCluster{Next: next}
	})

	// All OK, return a nil error.
	return nil
}

// ParseStanza parses a kubernetes stanza
func ParseStanza(c *caddy.Controller) (*MultiCluster, error) {
	c.Next() // Skip pluginName label

	zones := plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)
	multiCluster := New(zones)
	multiCluster.Gateway_ip4 = net.IPv4(1, 2, 3, 4)
	multiCluster.Gateway_ip6 = net.IPv4(1, 2, 3, 4) // #TODO find how to define ip6

	for c.NextBlock() {
		switch c.Val() {
		case "kubeconfig":
			//#TODO  update here to get the kubeconfig - check if needed (think so)
			//#TODO q  check if here you get the global var for the gateway ip
		case "fallthrough":
			//#TODO q  check if needed
		case "noendpoints":
			//#TODO q check if needed
		default:
			return nil, c.Errf("unknown property '%s'", c.Val())
		}
	}

	return multiCluster, nil
}
