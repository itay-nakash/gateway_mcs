// Package example is a CoreDNS plugin that prints "example" to stdout on every packet received.
//
// It serves as an example CoreDNS plugin with numerous code comments.
package multicluster_gw

import (
	"context"
	"errors"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

const (
	// defaultTTL to apply to all answers.
	// #TODO maybe add it to the config?
	defaultTTL = 5
)

var (
	errNoItems        = errors.New("no items found")
	errNsNotExposed   = errors.New("namespace is not exposed")
	errInvalidRequest = errors.New("invalid query name")
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin(pluginName)

//MultiCluster implements a plugin supporting multi-cluster DNS spec using a gateway.
type MultiCluster struct {
	Next  plugin.Handler
	Zones []string
	ttl   uint32
}

func New(zones []string) *MultiCluster {
	m := MultiCluster{
		Zones: zones,
	}

	m.ttl = defaultTTL

	return &m
}

// ServeDNS implements the plugin.Handler interface.
func (m MultiCluster) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	// Debug log that we've have seen the query. This will only be shown when the debug plugin is loaded.
	log.Debug("Received response")

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
	/*
		old example code:

		// Wrap.
		pw := NewResponsePrinter(w)

		// Export metric with the server label set to the current server handling the request.
		requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

		// Call next plugin (if any).
		return plugin.NextOrFailure(m.Name(), m.Next, ctx, pw, r)

	*/

	// #TODO check before if the name is exists (Check serviceImport via controller)
	// #TODO check if the controller can sync

	switch state.QType() {
	case dns.TypeA:
		log.Debug("Got into A type handel")
		// make A req to the gateway (?)

		break

	default:
		//Should I distinguish between NODATA and NXDOMAIN?
		log.Debug("Got into default")
		// #TODO check which error I should return if the req type dosent match
		// #TODO make sure that fallthrough when NXDOMAIN is not a wanted behavior

	}
}

// Name implements the Handler interface.
func (m MultiCluster) Name() string { return pluginName }

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
func (m MultiCluster) IsNameError(err error) bool {
	return err == errNoItems || err == errNsNotExposed || err == errInvalidRequest
}
