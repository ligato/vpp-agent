// Copyright (c) 2021 Pantheon.tech
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

package vpp

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"
	idxmap_mem "go.ligato.io/cn-infra/v2/idxmap/mem"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin"
	nat_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
	"google.golang.org/protobuf/proto"
)

const (
	vpp2005 = "20.05"
	vpp2009 = "20.09"
	vpp2101 = "21.01"
	vpp2106 = "21.06"
)

func TestNat44Global(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	unsupportedVPPVersions := []string{vpp2005, vpp2009}
	// in older versions is used VPP startup config
	// exclude test testing feature not supported in currently tested VPP version
	for _, excludedVPPVersion := range unsupportedVPPVersions {
		if ctx.versionInfo.Release() == excludedVPPVersion {
			return
		}
	}

	// nat handler
	swIfIndexes := ifaceidx.NewIfaceIndex(logrus.DefaultLogger(), "test-sw_if_indexes")
	dhcpIndexes := idxmap_mem.NewNamedMapping(logrus.DefaultLogger(), "test-dhcp_indexes", nil)
	natHandler := nat_vppcalls.CompatibleNatVppHandler(ctx.vppClient, swIfIndexes, dhcpIndexes, logrus.NewLogger("test"))

	Expect(natHandler).ShouldNot(BeNil(), "Nat handler should be created.")

	dumpBeforeEnable, err := natHandler.Nat44GlobalConfigDump(false)
	Expect(err).To(Succeed())
	t.Logf("dump before enable: %#v", dumpBeforeEnable)

	if !natHandler.WithLegacyStartupConf() {
		Expect(natHandler.EnableNAT44Plugin(nat_vppcalls.Nat44InitOpts{EndpointDependent: true})).Should(Succeed())
	}
	dumpAfterEnable, err := natHandler.Nat44GlobalConfigDump(false)
	Expect(err).To(Succeed())
	t.Logf("dump after enable: %#v", dumpAfterEnable)

	Expect(natHandler.DisableNAT44Plugin()).Should(Succeed())
	dumpAfterDisable, err := natHandler.Nat44GlobalConfigDump(false)
	Expect(err).To(Succeed())
	Expect(dumpAfterDisable).To(Equal(natHandler.DefaultNat44GlobalConfig()))
	t.Logf("dump after disable: %#v", dumpAfterDisable)

	if !natHandler.WithLegacyStartupConf() {
		Expect(natHandler.EnableNAT44Plugin(nat_vppcalls.Nat44InitOpts{EndpointDependent: true})).Should(Succeed())
	}
	dumpAfterSecondEnable, err := natHandler.Nat44GlobalConfigDump(false)
	Expect(err).To(Succeed())
	t.Logf("dump after second enable: %#v", dumpAfterSecondEnable)
}

// TestNat44StaticMapping tests Create/Read/Delete operations for NAT44 static mappings
func TestNat44StaticMapping(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	// nat handler
	swIfIndexes := ifaceidx.NewIfaceIndex(logrus.DefaultLogger(), "test-sw_if_indexes")
	dhcpIndexes := idxmap_mem.NewNamedMapping(logrus.DefaultLogger(), "test-dhcp_indexes", nil)
	natHandler := nat_vppcalls.CompatibleNatVppHandler(ctx.vppClient, swIfIndexes, dhcpIndexes, logrus.NewLogger("test"))
	Expect(natHandler).ShouldNot(BeNil(), "Handler should be created.")

	// some test constants
	const dnatLabel = "DNAT 1"
	localIP := net.ParseIP("10.0.0.1").To4()
	externalIP := net.ParseIP("10.0.0.2").To4()
	startOfIPPool := net.ParseIP("10.0.0.10").To4()
	endOfIPPool := net.ParseIP("10.0.0.11").To4()

	// setup twice NAT pool
	if !natHandler.WithLegacyStartupConf() {
		Expect(natHandler.EnableNAT44Plugin(nat_vppcalls.Nat44InitOpts{EndpointDependent: true})).Should(Succeed())
	}
	Expect(natHandler.AddNat44AddressPool(0, startOfIPPool.String(), endOfIPPool.String(), true)).Should(Succeed())

	tests := []struct {
		name                          string
		input                         *nat.DNat44_StaticMapping
		expectedDump                  *nat.DNat44_StaticMapping
		excludeUnsupportedVPPVersions []string
	}{
		{
			name: "simple NAT44 static mapping",
			input: &nat.DNat44_StaticMapping{
				Protocol:   nat.DNat44_TCP,
				ExternalIp: externalIP.String(),
				LocalIps: []*nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp: localIP.String(),
					},
				},
			},
		},
		{
			name: "NAT44 static mapping with twice nat",
			input: &nat.DNat44_StaticMapping{
				Protocol:     nat.DNat44_TCP,
				ExternalIp:   externalIP.String(),
				ExternalPort: 80,
				LocalIps: []*nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   localIP.String(),
						LocalPort: 8080,
					},
				},
				TwiceNat: nat.DNat44_StaticMapping_ENABLED,
			},
		},
		{
			name:                          "NAT44 static mapping with twice nat and twice NAT pool IP",
			excludeUnsupportedVPPVersions: []string{vpp2005},
			input: &nat.DNat44_StaticMapping{
				Protocol:     nat.DNat44_TCP,
				ExternalIp:   externalIP.String(),
				ExternalPort: 80,
				LocalIps: []*nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   localIP.String(),
						LocalPort: 8080,
					},
				},
				TwiceNat:       nat.DNat44_StaticMapping_ENABLED,
				TwiceNatPoolIp: endOfIPPool.String(),
			},
			expectedDump: &nat.DNat44_StaticMapping{
				// just missing TwiceNatPoolIp (VPP doesnt dump it)
				// TODO: fix test when dump will dump currently missing information
				Protocol:     nat.DNat44_TCP,
				ExternalIp:   externalIP.String(),
				ExternalPort: 80,
				LocalIps: []*nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   localIP.String(),
						LocalPort: 8080,
					},
				},
				TwiceNat: nat.DNat44_StaticMapping_ENABLED,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// exclude test testing feature not supported in currently tested VPP version
			for _, excludedVPPVersion := range test.excludeUnsupportedVPPVersions {
				if ctx.versionInfo.Release() == excludedVPPVersion {
					return
				}
			}

			// Create
			Expect(test).ShouldNot(BeNil())
			Expect(natHandler.AddNat44StaticMapping(test.input, dnatLabel)).Should(Succeed())

			// Read
			dnatDump, err := natHandler.DNat44Dump()
			t.Logf("received this dnat from dump: %v", dnatDump)
			Expect(err).ShouldNot(HaveOccurred())
			expected := test.input
			if test.expectedDump != nil {
				expected = test.expectedDump
			}
			Expect(dnatDump).To(HaveLen(1))
			Expect(dnatDump[0].StMappings).To(HaveLen(1))
			Expect(proto.Equal(dnatDump[0].StMappings[0], expected)).To(BeTrue())

			// Delete
			Expect(natHandler.DelNat44StaticMapping(test.input, dnatLabel)).Should(Succeed())
		})
	}
}
