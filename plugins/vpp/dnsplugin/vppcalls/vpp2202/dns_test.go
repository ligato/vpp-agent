// Copyright (c) 2022 Pantheon.tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vpp2202_test

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/dns"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/vppcalls/vpp2202"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
)

var (
	upstreamDNSServerIPv4 = net.ParseIP("8.8.8.8").To4()        // google dns
	upstreamDNSServerIPv6 = net.ParseIP("2001:4860:4860::8888") // google dns
)

// TestEnableDisableDNS tests all cases for methods EnableDNS and DisableDNS
func TestEnableDisableDNS(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name          string
		TestEnable    bool
		FailInVPP     bool
		ExpectFailure bool
		Expected      govppapi.Message
	}{
		{
			Name:       "successful enabling of DNS",
			TestEnable: true,
			Expected: &dns.DNSEnableDisable{
				Enable: 1,
			},
		},
		{
			Name:          "enable failure",
			TestEnable:    true,
			FailInVPP:     true,
			ExpectFailure: true,
		},
		{
			Name:       "successful disabling of DNS",
			TestEnable: false,
			Expected: &dns.DNSEnableDisable{
				Enable: 0,
			},
		},
		{
			Name:          "disable failure",
			TestEnable:    false,
			FailInVPP:     true,
			ExpectFailure: true,
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)
			// prepare reply
			if td.FailInVPP {
				ctx.MockVpp.MockReply(&dns.DNSEnableDisableReply{Retval: 1})
			} else {
				ctx.MockVpp.MockReply(&dns.DNSEnableDisableReply{})
			}

			// make the call
			var err error
			if td.TestEnable {
				err = vppCalls.EnableDNS()
			} else {
				err = vppCalls.DisableDNS()
			}

			// verify result
			if td.ExpectFailure {
				Expect(err).Should(HaveOccurred())
			} else {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ctx.MockChannel.Msg).To(Equal(td.Expected))
			}
		})
	}
}

// TestAddRemoveUpstreamDNSServer tests all cases for methods AddUpstreamDNSServer and RemoveUpstreamDNSServer
func TestAddRemoveUpstreamDNSServer(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name          string
		TestAdding    bool
		FailInVPP     bool
		ExpectFailure bool
		Input         net.IP
		Expected      govppapi.Message
	}{
		{
			Name:       "successful adding of IPv4 upstream DNS server",
			TestAdding: true,
			Input:      upstreamDNSServerIPv4,
			Expected: &dns.DNSNameServerAddDel{
				IsIP6:         0,
				IsAdd:         1,
				ServerAddress: upstreamDNSServerIPv4,
			},
		},
		{
			Name:       "successful adding of IPv6 upstream DNS server",
			TestAdding: true,
			Input:      upstreamDNSServerIPv6,
			Expected: &dns.DNSNameServerAddDel{
				IsIP6:         1,
				IsAdd:         1,
				ServerAddress: upstreamDNSServerIPv6,
			},
		},
		{
			Name:       "successful removal of IPv4 upstream DNS server",
			TestAdding: false,
			Input:      upstreamDNSServerIPv4,
			Expected: &dns.DNSNameServerAddDel{
				IsIP6:         0,
				IsAdd:         0,
				ServerAddress: upstreamDNSServerIPv4,
			},
		},
		{
			Name:       "successful removal of IPv6 upstream DNS server",
			TestAdding: false,
			Input:      upstreamDNSServerIPv6,
			Expected: &dns.DNSNameServerAddDel{
				IsIP6:         1,
				IsAdd:         0,
				ServerAddress: upstreamDNSServerIPv6,
			},
		},
		{
			Name:          "failure propagation from VPP",
			TestAdding:    false,
			FailInVPP:     true,
			ExpectFailure: true,
			Input:         upstreamDNSServerIPv4,
		},
		{
			Name:          "bad IP address input",
			TestAdding:    false,
			FailInVPP:     true,
			ExpectFailure: true,
			Input:         nil,
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)
			// prepare reply
			if td.FailInVPP {
				ctx.MockVpp.MockReply(&dns.DNSNameServerAddDelReply{Retval: 1})
			} else {
				ctx.MockVpp.MockReply(&dns.DNSNameServerAddDelReply{})
			}

			// make the call
			var err error
			if td.TestAdding {
				err = vppCalls.AddUpstreamDNSServer(td.Input)
			} else {
				err = vppCalls.DeleteUpstreamDNSServer(td.Input)
			}

			// verify result
			if td.ExpectFailure {
				Expect(err).Should(HaveOccurred())
			} else {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ctx.MockChannel.Msg).To(Equal(td.Expected))
			}
		})
	}
}

func setup(t *testing.T) (*vppmock.TestCtx, vppcalls.DNSVppAPI) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test")
	vppCalls := vpp2202.NewDNSVppHandler(ctx.MockVPPClient, log)
	return ctx, vppCalls
}

func teardown(ctx *vppmock.TestCtx) {
	ctx.TeardownTestCtx()
}
