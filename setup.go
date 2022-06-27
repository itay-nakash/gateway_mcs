package multicluster_gw

import (
	"context"
	"net"
	"strings"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const pluginName = "multicluster_gw"

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
func ParseStanza(c *caddy.Controller) (*MulticlusterGw, error) {
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

		case "gateway":
			var err error
			multiCluster.gatewayIp4, multiCluster.gatewayIp6, err = parseIp(c)
			if err != nil {
				return nil, plugin.Error(pluginName, err)
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
	ip := net.ParseIP(ip_as_string)
	if ip == nil {
		// The gateway was given as string, so we need to extract from the service name the ip
		ipv4, ipv6, err := getGwIpFromString(ip_as_string)
		return ipv4, ipv6, err
		//return net.IPv4(6, 6, 6, 6), net.IPv4(6, 6, 6, 6).To16(), nil
	} else {
		// The gateway was given as an ip address, just foward it
		return ip.To4(), ip.To16(), nil
	}
}

func getGwIpFromString(gwFqdn string) (net.IP, net.IP, error) {
	var gwIpAsString string
	gwName, gwNs := splitNameAndNs(gwFqdn)
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	services, err := clientset.CoreV1().Services(gwNs).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	for _, service := range services.Items {
		if service.Spec.ExternalName == gwName {
			gwIpAsString = service.Spec.ClusterIP
		}
	}
	// Here it from some reason gets in 'gwIpAsString' something that gives 'ParseError' from ParseIP,
	// I'm currently not sure how to debug it (its from the 'List' function, using the client..)
	print(gwIpAsString)
	return net.ParseIP(gwIpAsString).To4(), net.ParseIP(gwIpAsString).To16(), nil
}

func splitNameAndNs(gwFqdn string) (string, string) {
	nameNS := strings.Split(gwFqdn, ".")
	var ns string
	// Trim the string until the '.' :
	if idx := strings.IndexByte(nameNS[1], '.'); idx >= 0 {
		ns = nameNS[1][:idx]
	} else {
		ns = nameNS[1]
	}
	name := nameNS[0]
	return name, ns
}
