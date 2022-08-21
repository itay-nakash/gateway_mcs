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
		serviceName         string
		serviceNs           string
		questionType        uint16 // The given request type
		shouldErr           bool   // True if test case is expected to produce an error.
		expectedReturnValue int    // The expected return value.
		expectedErrContent  error  // The expected error
		addToSet            bool
	}{
		// positive
		{
			`myservice.test.svc.clusterset.local.`,
			"myservice",
			"test",
			dns.TypeA,
			false,
			dns.RcodeSuccess,
			nil,
			true,
		},
		// not for the plugin's zone, should foward it and not handle the request:
		{
			`myservice.test.svc.cluster.local.`,
			"myservice",
			"test",
			dns.TypeA,
			false,
			dns.RcodeServerFailure,
			nil,
			true,
		},

		//not in the set, should return NODATA:
		{
			`myservice.test.svc.clusterset.local.`,
			"myservice",
			"test",
			dns.TypeA,
			false,
			dns.RcodeNameError,
			nil,
			false,
		},
	}
	requestsZone := "svc.clusterset.local."
	mcsPlugin := MulticlusterGw{Zones: []string{requestsZone}, Next: test.ErrorHandler()}
	ctx := context.TODO()
	r := new(dns.Msg)
	rec := dnstest.NewRecorder((&test.ResponseWriter{}))
	for _, test := range tests {
		initalizeSetForTest(test.question, test.serviceName, test.serviceNs, test.addToSet)
		r.SetQuestion(test.question, test.questionType)

		// call the plugin and check result:
		returnValue, err := mcsPlugin.ServeDNS(ctx, rec, r)

		assert.Equal(t, test.expectedReturnValue, returnValue)
		if test.shouldErr {
			assert.NotEmpty(t, err)
			// #TODO check how to check the spesific error type, and which errors I except?
		} else {
			assert.Nil(t, err)
		}
		// #TODO how to check the response message?
	}
}

func initalizeSetForTest(qustion string, svcName string, svcNS string, addToSet bool) {
	// empty the set in each test run:
	SIset.Elements = make(map[string]struct{})

	if addToSet {
		// add the current SI to the set:
		SIset.Add(GenerateNameAsString(svcName, svcNS))
	}
}
