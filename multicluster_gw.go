// Package example is a CoreDNS plugin that prints "example" to stdout on every packet received.
//
// It serves as an example CoreDNS plugin with numerous code comments.
package multicluster_gw

import (
	"context"
	"errors"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"

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

	// Wrap.
	pw := NewResponsePrinter(w)

	// Export metric with the server label set to the current server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	// Call next plugin (if any).
	return plugin.NextOrFailure(m.Name(), m.Next, ctx, pw, r)
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
