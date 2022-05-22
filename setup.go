package gateway_mcs

import (
	"net"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"k8s.io/client-go/tools/clientcmd"
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

	// ------------#TODO init the controller here------------
	// tmp code until multiCluster will be used:
	if multiCluster == nil {
		print("err")
	}
	// -----------------------------------------------------------

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	// #TODO check why we need caddy controller for the config?..
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
	multiCluster.gateway_ip4 = net.IPv4(1, 2, 3, 4)
	multiCluster.gateway_ip6 = net.IPv4(1, 2, 3, 4).To16()

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
			ip_as_string := c.RemainingArgs()[0]
			ip_address_len := 4
			if len(ip_as_string) != ip_address_len {
				// #TODO check if it maybe can be 6?
				return nil, c.ArgErr()
			}

			input_ip := []byte(ip_as_string)
			// #TODO maybe check if it even possible to squeze the gateway to only 4 bytes? :
			multiCluster.gateway_ip4 = net.IPv4(input_ip[0], input_ip[1], input_ip[2], input_ip[3])
			// #TODO find if it reasnoble way to define ip6
			multiCluster.gateway_ip6 = net.IPv4(input_ip[0], input_ip[1], input_ip[2], input_ip[3]).To16()

		default:
			return nil, c.Errf("unknown property '%s'", c.Val())
		}
	}

	return multiCluster, nil
}
