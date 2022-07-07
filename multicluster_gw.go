package multicluster_gw

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/coredns/coredns/plugin/pkg/fall"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// defaultTTL to apply to all answers.
	// #TODO maybe add it to the core-config?
	defaultTTL = 5
)

var (
	errNoItems        = errors.New("no items found")
	errNsNotExposed   = errors.New("namespace is not exposed")
	errInvalidRequest = errors.New("invalid query name")
	defaultGwIpv4     = net.IPv4(1, 2, 3, 4)
	defaultGwIpv6     = net.IPv4(1, 2, 3, 4).To16()
)

// Define log to be a logger with the plugin name in it.
var log = clog.NewWithPlugin(pluginName)

// MulticlusterGw implements a plugin supporting multi-cluster DNS spec using a gateway.
type MulticlusterGw struct {
	Next         plugin.Handler
	Zones        []string
	Fall         fall.F
	ClientConfig clientcmd.ClientConfig
	gatewayIp4   net.IP
	gatewayIp6   net.IP
	svcName      string
	svcNS        string
	ttl          uint32
}

func New(zones []string) *MulticlusterGw {
	m := MulticlusterGw{
		Zones: zones,
	}
	// set default gateway:
	m.gatewayIp4 = defaultGwIpv4
	m.gatewayIp6 = defaultGwIpv6
	m.ttl = defaultTTL

	return &m
}

// ServeDNS implements the plugin.Handler interface.
func (m MulticlusterGw) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Debug log that we've have seen the query.
	log.Debug("gw_mcs received req")

	// parse the req:
	state := request.Request{W: w, Req: r}

	// get the req name:
	qname := state.QName()

	// check if any subdomain of one of the zones
	zone := plugin.Zones(m.Zones).Matches(qname)
	if zone == "" {
		// if not - pass it to the next plugin
		return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
	}

	// get all the request without the zone (the .local..):
	// "maintain case of original query"
	zone = qname[len(qname)-len(zone):]
	state.Zone = zone

	m.svcName, m.svcNS = parseReqNameNs(qname[:len(qname)-len(zone)])

	var (
		records []dns.RR
		// extra   []dns.RR   ---- add in future for SRV records -----
	)

	if SIset.Contains(GenerateNameAsString(m.svcName, m.svcNS)) {

		switch state.QType() {
		case dns.TypeA:
			log.Debug("Handles Type A request")
			records = append(records, NewARecord(qname, m.gatewayIp4))
		case dns.TypeAAAA:
			log.Debug("Handles Type AAAA request")
			records = append(records, NewAAAARecord(qname, m.gatewayIp6))

		default:
			// return NODATA error (?)
			// #TODO q Should I distinguish between NOData and NXDomain?
			// #TODO q check which error I should return if the req type dosent match

			if m.Fall.Through(state.Name()) {
				return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
			}
			return dns.RcodeNameError, nil // return NXDomain
		}

	} else {

		// find what to do here :

		if m.Fall.Through(state.Name()) {
			return plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)
		}
		return dns.RcodeNameError, nil // return NXDomain
	}

	// if the req succeed:
	message := &dns.Msg{}
	message.SetReply(r)
	message.Authoritative = true

	//Add the answer:
	message.Answer = append(message.Answer, records...)
	w.WriteMsg(message)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (m MulticlusterGw) Name() string { return pluginName }

// ResponsePrinter wrap a dns.ResponseWriter and will write example to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter returns ResponseWriter.
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "example" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info(pluginName)
	return r.ResponseWriter.WriteMsg(res)
}

// IsNameError returns true if err indicated a record not found condition
func (m MulticlusterGw) IsNameError(err error) bool {
	return err == errNoItems || err == errNsNotExposed || err == errInvalidRequest
}

// NewA returns a new A record based on the Service.
func NewARecord(name string, ip net.IP) *dns.A {
	return &dns.A{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA,
		Class: dns.ClassINET, Ttl: defaultTTL}, A: ip}
}

// NewAAAA returns a new AAAA record based on the Service.
func NewAAAARecord(name string, ip net.IP) *dns.AAAA {
	return &dns.AAAA{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA,
		Class: dns.ClassINET, Ttl: defaultTTL}, AAAA: ip}
}

func parseReqNameNs(qnameTrimmed string) (string, string) {
	// TODO: might add error checking? or its clear that should contain .
	firstDotIndex := strings.IndexByte(qnameTrimmed, '.')

	name := qnameTrimmed[:firstDotIndex]
	ns := qnameTrimmed[firstDotIndex+1:]
	ns = ns[:len(ns)-1] // remove the dot in the end
	return name, ns
}
