package multicluster_gw

import (
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/fall"
)

// TestSetup tests the various things that should be parsed by setup.
func TestSetup(t *testing.T) {

	tests := []struct {
		input               string // Corefile data as string
		shouldErr           bool   // true if test case is expected to produce an error.
		expectedErrContent  string // substring from the expected error. Empty for positive cases.
		expectedZoneCount   int    // expected count of defined zones.
		expectedGatewayIp4  net.IP // expected value of gateway ip in IP4 format.
		expectedGatewayIp6  net.IP // expected value of gateway ip in IP6 format.
		expectedFallthrough fall.F // expected value of fallthourgh setting.
	}{
		{
			`multicluster_gw .svc.clusterset.local.`,
			false,
			"",
			1,
			defaultGwIpv4,
			defaultGwIpv6,
			fall.Zero,
		},
		{
			`multicluster_gw coredns.local .svc.clusterset.local.`,
			false,
			"",
			2,
			defaultGwIpv4,
			defaultGwIpv6,
			fall.Zero,
		},
		{
			`multicluster_gw coredns.local .svc.clusterset.local. {
    fallthrough
}`,
			false,
			"",
			2,
			defaultGwIpv4,
			defaultGwIpv6,
			fall.Root,
		},
		{
			`multicluster_gw coredns.local .svc.clusterset.local. {
    gateway_ip 6.6.6.6
}`,
			false,
			"",
			2,
			net.IPv4(6, 6, 6, 6),
			net.IPv4(6, 6, 6, 6).To16(),
			fall.Zero,
		},
	}

	/*
		Test for those cases:
		1. check if error was raised if needed
		2. check number of zone that were recognized
		3. check if fallthrough was configed correctly
		4. check for the value of the ip of gateway, and if configed correctly
	*/
	for i, test := range tests {
		// init the Mcgw as a new, empty one:
		Mcgw = MulticlusterGw{}
		c := caddy.NewTestController("dns", test.input)
		err := ParseStanza(c, &Mcgw)

		if test.shouldErr && err == nil {
			t.Errorf("Test %d: Expected error, but did not find error for input '%s'. Error was: '%v'", i, test.input, err)
		}

		if err != nil {
			if !test.shouldErr {
				t.Errorf("Test %d: Expected no error but found one for input %s. Error was: %v", i, test.input, err)
				continue
			}

			if test.shouldErr && (len(test.expectedErrContent) < 1) {
				t.Fatalf("Test %d: Test marked as expecting an error, but no expectedErrContent provided for input '%s'. Error was: '%v'", i, test.input, err)
			}

			if test.shouldErr && (test.expectedZoneCount >= 0) {
				t.Errorf("Test %d: Test marked as expecting an error, but provides value for expectedZoneCount!=-1 for input '%s'. Error was: '%v'", i, test.input, err)
			}

			if !strings.Contains(err.Error(), test.expectedErrContent) {
				t.Errorf("Test %d: Expected error to contain: %v, found error: %v, input: %s", i, test.expectedErrContent, err, test.input)
			}
			continue
		}

		// No error was raised, so validate initialization of k8sController
		foundZoneCount := len(Mcgw.Zones)
		if foundZoneCount != test.expectedZoneCount {
			t.Errorf("Test %d: Expected kubernetes controller to be initialized with %d zones, instead found %d zones: '%v' for input '%s'", i, test.expectedZoneCount, foundZoneCount, Mcgw.Zones, test.input)
		}

		// fallthrough
		if !Mcgw.Fall.Equal(test.expectedFallthrough) {
			t.Errorf("Test %d: Expected kubernetes controller to be initialized with fallthrough '%v'. Instead found fallthrough '%v' for input '%s'", i, test.expectedFallthrough, Mcgw.Fall, test.input)
		}

		// gateway
		if !reflect.DeepEqual(Mcgw.gatewayIp4.To16(), test.expectedGatewayIp4) {
			t.Errorf("Test %d: Expected kubernetes controller to be initialized with gateway Ip4 of '%v'. Instead found gateway Ip4 of '%v' for input '%s'", i, test.expectedGatewayIp4, Mcgw.gatewayIp4, test.input)
		}
		if !reflect.DeepEqual(Mcgw.gatewayIp6, test.expectedGatewayIp6) {
			t.Errorf("Test %d: Expected kubernetes controller to be initialized with gateway Ip6 of '%v'. Instead found gateway Ip6 of '%v' for input '%s'", i, test.expectedGatewayIp6, Mcgw.gatewayIp6, test.input)
		}
	}
}
