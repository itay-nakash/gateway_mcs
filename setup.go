package multicluster_gw

import (
	"net"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"k8s.io/client-go/tools/clientcmd"
)

const pluginName = "multicluster_gw"

var (
	gateway_ip4 net.IP
	gateway_ip6 net.IP
)

// init registers this plugin.
func init() { plugin.Register(pluginName, setup) }

// setup is that initialize the plugin givven the core-file settings for it.
// check for the wanted zones, if fallthrough is wanted, and what the wanted gateway-ip
func setup(c *caddy.Controller) error {

	multiCluster, err := ParseStanza(c)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	// ------------#TODO init the controller here------------
	//
	// ------------------------------------------------------

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		multiCluster.Next = next
		return multiCluster
	})

	// All OK, return a nil error.
	return nil
}

// ParseStanza parses a kubernetes stanza
func ParseStanza(c *caddy.Controller) (*Multicluster_gw, error) {
	c.Next() // Skip pluginName label

	zones := plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)
	multiCluster := New(zones)

	for c.NextBlock() {
		switch c.Val() {
		case "kubeconfig":
			args := c.RemainingArgs()
			if len(args) != 1 && len(args) != 2 {
				return nil, c.ArgErr()
			}
			overrides := &clientcmd.ConfigOverrides{}
			if len(args) == 2 {
				overrides.CurrentContext = args[1]
			}
			config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{ExplicitPath: args[0]},
				overrides,
			)
			multiCluster.ClientConfig = config

		case "fallthrough":
			multiCluster.Fall.SetZonesFromArgs(c.RemainingArgs())

		case "gateway_ip":
			multiCluster.gateway_ip4, multiCluster.gateway_ip6 = parseIp(c)

		default:
			return nil, c.Errf("unknown property '%s'", c.Val())
		}
	}

	return multiCluster, nil
}

// parse the Ip given as caddy.Controller arg, as a string, to ipv4 and ipv6 format
func parseIp(c *caddy.Controller) (net.IP, net.IP) {
	ip_as_string := c.RemainingArgs()[0]
	ip := net.ParseIP(ip_as_string)
	if ip == nil {
		//The ip was given as string, not a number
		return nil, nil
	} else {
		return ip.To4(), ip.To16()
	}
}
