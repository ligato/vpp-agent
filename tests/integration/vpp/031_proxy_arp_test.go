//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	netalloc_mock "go.ligato.io/vpp-agent/v3/plugins/netalloc/mock"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	l3plugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
)

func TestProxyARPRange(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	vrfIndexes.Put("vrf1-ipv4", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV4})
	vrfIndexes.Put("vrf1-ipv6", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV6})

	h := l3plugin_vppcalls.CompatibleL3VppHandler(ctx.vppClient, ifIndexes, vrfIndexes,
		netalloc_mock.NewMockNetAlloc(), logrus.NewLogger("test"))

	tests := []struct {
		name          string
		proxyARPRange *vpp_l3.ProxyARP_Range
		addMustFail   bool
	}{
		{
			name: "base test of proxy ARP range",
			proxyARPRange: &vpp_l3.ProxyARP_Range{
				FirstIpAddr: "192.168.1.1",
				LastIpAddr:  "192.168.1.10",
				VrfId:       0,
			},
			addMustFail: false,
		},
		{
			name: "Without first IP address",
			proxyARPRange: &vpp_l3.ProxyARP_Range{
				LastIpAddr: "192.168.1.10",
				VrfId:      0,
			},
			addMustFail: true,
		},
		{
			name: "Without last IP address",
			proxyARPRange: &vpp_l3.ProxyARP_Range{
				FirstIpAddr: "192.168.1.1",
				VrfId:       0,
			},
			addMustFail: true,
		},
		{
			name: "Without both IP addresses",
			proxyARPRange: &vpp_l3.ProxyARP_Range{
				VrfId: 0,
			},
			addMustFail: true,
		},
		{
			name: "No such FIB / VRF",
			proxyARPRange: &vpp_l3.ProxyARP_Range{
				FirstIpAddr: "192.168.1.1",
				LastIpAddr:  "192.168.1.10",
				VrfId:       2,
			},
			addMustFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initialRanges, err := h.DumpProxyArpRanges()
			Expect(err).ToNot(HaveOccurred(), "dumping proxy ARP ranges failed")
			Expect(initialRanges).To(HaveLen(0), "expected no proxy ARP ranges in the beginning")
			err = h.AddProxyArpRange(
				net.ParseIP(test.proxyARPRange.FirstIpAddr).To4(),
				net.ParseIP(test.proxyARPRange.LastIpAddr).To4(),
				test.proxyARPRange.VrfId,
			)
			if test.addMustFail {
				Expect(err).To(HaveOccurred(), "adding proxy ARP range passed successfully but must fail")
				return
			}
			Expect(err).ToNot(HaveOccurred(), "adding proxy ARP range failed")

			afterAddRanges, err := h.DumpProxyArpRanges()
			Expect(err).ToNot(HaveOccurred(), "dumping proxy ARP ranges failed")
			Expect(afterAddRanges).To(HaveLen(1), "expected one proxy ARP range")
			Expect(afterAddRanges[0].Range.FirstIpAddr).To(Equal(test.proxyARPRange.FirstIpAddr),
				"First IP address of proxy ARP range retrieved from dump is not equal to expected",
			)
			Expect(afterAddRanges[0].Range.LastIpAddr).To(Equal(test.proxyARPRange.LastIpAddr),
				"Last IP address of proxy ARP range retrieved from dump is not equal to expected",
			)
			Expect(afterAddRanges[0].Range.VrfId).To(Equal(test.proxyARPRange.VrfId),
				"VRF ID of proxy ARP range retrieved from dump is not equal to expected",
			)

			Expect(
				h.DeleteProxyArpRange(
					net.ParseIP(test.proxyARPRange.FirstIpAddr).To4(),
					net.ParseIP(test.proxyARPRange.LastIpAddr).To4(),
					test.proxyARPRange.VrfId,
				),
			).To(Succeed(), "deleting proxy ARP range failed")

			afterDeleteRanges, err := h.DumpProxyArpRanges()
			Expect(err).ToNot(HaveOccurred(), "dumping proxy ARP ranges failed")
			Expect(afterDeleteRanges).To(HaveLen(0), "expected no proxy ARP ranges after deleting one")
		})
	}
}
