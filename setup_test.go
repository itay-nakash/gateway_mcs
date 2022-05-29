package gateway_mcs

import (
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/fall"
)

// TestSetup tests the various things that should be parsed by setup.
// Make sure you also test for parse errors.
func TestSetup(t *testing.T) {

	tests := []struct {
		input               string // Corefile data as string
		shouldErr           bool   // true if test case is expected to produce an error.
		expectedErrContent  string // substring from the expected error. Empty for positive cases.
		expectedZoneCount   int    // expected count of defined zones.
		expectedGatewayIp4  net.IP
		expectedGatewayIp6  net.IP
		expectedFallthrough fall.F
	}{
		// positive
		{
			`multicluster clusterset.local`,
			false,
			"",
			1,
			net.IPv4(1, 2, 3, 4),
			net.IPv4(1, 2, 3, 4).To16(),
			fall.Zero,
		},
		{
			`multicluster coredns.local clusterset.local`,
			false,
			"",
			2,
			net.IPv4(1, 2, 3, 4),
			net.IPv4(1, 2, 3, 4).To16(),
			fall.Zero,
		},
		{
			`kubernetes coredns.local clusterset.local {
    fallthrough
}`,
			false,
			"",
			2,
			net.IPv4(1, 2, 3, 4),
			net.IPv4(1, 2, 3, 4).To16(),
			fall.Root,
		},
		{
			`kubernetes coredns.local clusterset.local {
    gateway_ip 1.1.1.1
}`,
			false,
			"",
			2,
			net.IPv4(1, 1, 1, 1),
			net.IPv4(1, 1, 1, 1).To16(),
			fall.Zero,
		},
	}

	/*
		Test for:
		1. error if should
		2. number of zone count that were recognized
		3. if fallthrough was recognized

		Should add (?) :
		1. len + values of gateway ip

	*/
	for i, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		k8sController, err := ParseStanza(c)

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
		//     Zones
		foundZoneCount := len(k8sController.Zones)
		if foundZoneCount != test.expectedZoneCount {
			t.Errorf("Test %d: Expected kubernetes controller to be initialized with %d zones, instead found %d zones: '%v' for input '%s'", i, test.expectedZoneCount, foundZoneCount, k8sController.Zones, test.input)
		}

		// fallthrough
		if !k8sController.Fall.Equal(test.expectedFallthrough) {
			t.Errorf("Test %d: Expected kubernetes controller to be initialized with fallthrough '%v'. Instead found fallthrough '%v' for input '%s'", i, test.expectedFallthrough, k8sController.Fall, test.input)
		}

		// gateway
		if !reflect.DeepEqual(k8sController.gateway_ip4, test.expectedGatewayIp4) {
			t.Errorf("Test %d: Expected kubernetes controller to be initialized with gateway Ip4 of '%v'. Instead found gateway Ip4 of '%v' for input '%s'", i, test.expectedGatewayIp4, k8sController.gateway_ip4, test.input)
		}
		if !reflect.DeepEqual(k8sController.gateway_ip6, test.expectedGatewayIp6) {
			t.Errorf("Test %d: Expected kubernetes controller to be initialized with gateway Ip6 of '%v'. Instead found gateway Ip6 of '%v' for input '%s'", i, test.expectedGatewayIp6, k8sController.gateway_ip6, test.input)
		}
	}
}
