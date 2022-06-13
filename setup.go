package multicluster_gw

import (
	"net"
	"strconv"
	"strings"

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
			var err error
			multiCluster.gateway_ip4, multiCluster.gateway_ip6, err = parseIp(c)
			if err != nil {
				return nil, c.ArgErr()
			}

		default:
			return nil, c.Errf("unknown property '%s'", c.Val())
		}
	}

	return multiCluster, nil
}

// parse the Ip given as caddy.Controller arg, as a string, to ipv4 and ipv6 format
func parseIp(c *caddy.Controller) (net.IP, net.IP, error) {
	ip_as_string := c.RemainingArgs()[0]
	if ip_as_string == "" {
		return nil, nil, c.ArgErr()
	}
	// split the ip according to the dots:
	ip_ascii := strings.SplitAfter(ip_as_string, ".")
	// #TODO check if it maybe can be 6?
	if len(ip_ascii) != 4 {
		return nil, nil, c.ArgErr()
	}
	// get rid of the dots:
	for i := range ip_ascii {
		ip_ascii[i] = strings.TrimSuffix(ip_ascii[i], ".") // #TODO maybe add error check for ip address?
	}
	// convert to int in order to convert to byte afterwards:
	dig1, err := strconv.Atoi(ip_ascii[0])
	if err != nil {
		return nil, nil, c.ArgErr()
	}
	dig2, err := strconv.Atoi(ip_ascii[1])
	if err != nil {
		return nil, nil, c.ArgErr()
	}
	dig3, err := strconv.Atoi(ip_ascii[2])
	if err != nil {
		return nil, nil, c.ArgErr()
	}
	dig4, err := strconv.Atoi(ip_ascii[3])
	if err != nil {
		return nil, nil, c.ArgErr()
	}
	// convert to []byte and create the ip addresses:
	ip_byte := [4]byte{byte(dig1), byte(dig2), byte(dig3), byte(dig4)}
	ipv4 := net.IPv4(ip_byte[0], ip_byte[1], ip_byte[2], ip_byte[3])
	ipv6 := net.IPv4(ip_byte[0], ip_byte[1], ip_byte[2], ip_byte[3]).To16()
	return ipv4, ipv6, nil

}
