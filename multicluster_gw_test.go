package multicluster_gw

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestMultiClusterGw(t *testing.T) {
	tests := []struct {
		question            string // Corefile data as string
		question_type       uint16 // The givven request type
		shouldErr           bool   // True if test case is expected to produce an error.
		expectedReturnValue int    // The expected return value.
		expectedErrContent  error  // The expected error

	}{
		// positive
		{
			`myservice.test.svc.clusterset.local.`,
			dns.TypeA,
			false,
			dns.RcodeSuccess,
			nil,
		},
		// not for the zone, should foward it:
		{
			`myservice.test.svc.cluster.local.`,
			dns.TypeA,
			false,
			dns.RcodeServerFailure,
			nil,
		},
	}
	requestsZone := "svc.clusterset.local."
	mcs_plugin := Multicluster_gw{Zones: []string{requestsZone}, Next: test.ErrorHandler()}
	ctx := context.TODO()
	r := new(dns.Msg)
	rec := dnstest.NewRecorder((&test.ResponseWriter{}))
	for _, test := range tests {
		r.SetQuestion(test.question, test.question_type)

		// call the plugin and check result:
		return_value, err := mcs_plugin.ServeDNS(ctx, rec, r)

		assert.Equal(t, test.expectedReturnValue, return_value)
		if test.shouldErr {
			assert.NotEmpty(t, err)
			// #TODO check how to check the spesific error type, and which errors I except?
		} else {
			assert.Nil(t, err)
		}

		// #TODO how to check the response message?
	}
}
