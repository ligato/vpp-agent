// Copyright (c) 2019 Pantheon.tech
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
	"fmt"
	"net"
	"testing"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	netalloc_mock "go.ligato.io/vpp-agent/v3/plugins/netalloc/mock"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	l3plugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
	srv6_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin"
)

var (
	sidA         = sid("A::")
	nextHop      = net.ParseIP("B::").To16()
	nextHop2     = net.ParseIP("C::").To16()
	nextHopIPv4  = net.ParseIP("1.2.3.4").To4()
	nextHop2IPv4 = net.ParseIP("1.2.3.5").To4()
	vrfTables    = []*vpp_l3.VrfTable{
		{
			Id:       11,
			Protocol: vpp_l3.VrfTable_IPV6,
			Label:    "testIpv6Table1",
		},
		{
			Id:       12,
			Protocol: vpp_l3.VrfTable_IPV6,
			Label:    "testIpv6Table2",
		},
		{
			Id:       13,
			Protocol: vpp_l3.VrfTable_IPV4,
			Label:    "testIpv4Table1",
		},
		{
			Id:       14,
			Protocol: vpp_l3.VrfTable_IPV4,
			Label:    "testIpv4Table2",
		},
	}
)

//TODO add CRUD tests for SR-proxy (there is no binary API for it -> dump for SR-proxy needs 1. CLI localsid dump
// and 2. CLI fib table dump because CLI localsid dump does not include installation vrf for SR-proxy localsid)

// TestLocalsidCRUD tests CRUD operations for Localsids
func TestLocalsidCRUD(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	// create interfaces (for referencing from localsids)
	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))
	const ifName = "loop1"
	ifIdx, err := ih.AddLoopbackInterface(ifName)
	Expect(err).To(BeNil(), fmt.Sprintf("fixture setup failed in creating of interface %v: %v", ifName, err))
	t.Logf("interface created %v", ifIdx)
	const ifName2 = "loop2"
	ifIdx2, err := ih.AddLoopbackInterface(ifName2)
	Expect(err).To(BeNil(), fmt.Sprintf("fixture setup failed in creating of interface %v: %v", ifName2, err))
	t.Logf("interface created %v", ifIdx2)
	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-idx"), "test-idx")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{SwIfIndex: ifIdx})
	ifIndexes.Put(ifName2, &ifaceidx.IfaceMetadata{SwIfIndex: ifIdx2})

	// create vrf tables (for referencing from localsids)
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	vrfIndexes.Put("vrf1-ipv4", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV4})
	vrfIndexes.Put("vrf1-ipv6", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV6})
	l3h := l3plugin_vppcalls.CompatibleL3VppHandler(ctx.vppClient, ifIndexes, vrfIndexes,
		netalloc_mock.NewMockNetAlloc(), logrus.NewLogger("test-l3"))
	for _, vrfTable := range vrfTables[:4] {
		Expect(l3h.AddVrfTable(vrfTable)).Should(Succeed(), fmt.Sprintf("fixture setup failed "+
			"in creating of vrf table with name %v due to: %v", vrfTable.Label, err))
	}

	// SRv6 handler
	srh := srv6_vppcalls.CompatibleSRv6Handler(ctx.vppClient, ifIndexes, logrus.NewLogger("test"))

	tests := []struct {
		name                string
		input               *srv6.LocalSID
		expectedDump        *srv6.LocalSID
		updatedInput        *srv6.LocalSID
		updatedExpectedDump *srv6.LocalSID
	}{
		{
			name: "base end",
			input: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_BaseEndFunction{
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				},
			},
			updatedInput: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_BaseEndFunction{
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: false,
					},
				},
			},
		},
		{
			name: "end.X",
			input: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionX{
					EndFunctionX: &srv6.LocalSID_EndX{
						Psp:               true,
						NextHop:           nextHop.String(),
						OutgoingInterface: ifName,
					},
				},
			},
			updatedInput: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionX{
					EndFunctionX: &srv6.LocalSID_EndX{
						Psp:               true,
						NextHop:           nextHop2.String(), // updated
						OutgoingInterface: ifName,
					},
				},
			},
		},
		{
			name: "end.T",
			input: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionT{
					EndFunctionT: &srv6.LocalSID_EndT{
						Psp:   true,
						VrfId: vrfTables[0].Id,
					},
				},
			},
			expectedDump: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionT{
					EndFunctionT: &srv6.LocalSID_EndT{
						Psp:   true,
						VrfId: 1, // bug in VPP, it should return client Table ID but it returns vpp-inner Table ID
					},
				},
			},
			updatedInput: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionT{
					EndFunctionT: &srv6.LocalSID_EndT{
						Psp:   true,
						VrfId: vrfTables[1].Id, // updated
					},
				},
			},
			updatedExpectedDump: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionT{
					EndFunctionT: &srv6.LocalSID_EndT{
						Psp:   true,
						VrfId: 2, // bug in VPP, it should return client Table ID but it returns vpp-inner Table ID
					},
				},
			},
		},
		{
			name: "end.DT4",
			input: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDt4{
					EndFunctionDt4: &srv6.LocalSID_EndDT4{
						VrfId: vrfTables[2].Id,
					},
				},
			},
			expectedDump: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDt4{
					EndFunctionDt4: &srv6.LocalSID_EndDT4{
						VrfId: 1, // bug in VPP, it should return client Table ID but it returns vpp-inner Table ID
					},
				},
			},
			updatedInput: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDt4{
					EndFunctionDt4: &srv6.LocalSID_EndDT4{
						VrfId: vrfTables[3].Id, // updated
					},
				},
			},
			updatedExpectedDump: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDt4{
					EndFunctionDt4: &srv6.LocalSID_EndDT4{
						VrfId: 2, // bug in VPP, it should return client Table ID but it returns vpp-inner Table ID
					},
				},
			},
		},
		{
			name: "end.DT6",
			input: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDt6{
					EndFunctionDt6: &srv6.LocalSID_EndDT6{
						VrfId: vrfTables[0].Id,
					},
				},
			},
			expectedDump: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDt6{
					EndFunctionDt6: &srv6.LocalSID_EndDT6{
						VrfId: 1, // bug in VPP, it should return client Table ID but it returns vpp-inner Table ID
					},
				},
			},
			updatedInput: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDt6{
					EndFunctionDt6: &srv6.LocalSID_EndDT6{
						VrfId: vrfTables[1].Id, // updated
					},
				},
			},
			updatedExpectedDump: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDt6{
					EndFunctionDt6: &srv6.LocalSID_EndDT6{
						VrfId: 2, // bug in VPP, it should return client Table ID but it returns vpp-inner Table ID
					},
				},
			},
		},
		{
			name: "end.DX2",
			input: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDx2{
					EndFunctionDx2: &srv6.LocalSID_EndDX2{
						VlanTag:           0,
						OutgoingInterface: ifName,
					},
				},
			},
			updatedInput: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDx2{
					EndFunctionDx2: &srv6.LocalSID_EndDX2{
						VlanTag:           0,
						OutgoingInterface: ifName2,
					},
				},
			},
		},
		{
			name: "end.DX4",
			input: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDx4{
					EndFunctionDx4: &srv6.LocalSID_EndDX4{
						NextHop:           nextHopIPv4.String(),
						OutgoingInterface: ifName,
					},
				},
			},
			updatedInput: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDx4{
					EndFunctionDx4: &srv6.LocalSID_EndDX4{
						NextHop:           nextHop2IPv4.String(), // updated
						OutgoingInterface: ifName,
					},
				},
			},
		},
		{
			name: "end.DX6",
			input: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDx6{
					EndFunctionDx6: &srv6.LocalSID_EndDX6{
						NextHop:           nextHop.String(),
						OutgoingInterface: ifName,
					},
				},
			},
			updatedInput: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_EndFunctionDx6{
					EndFunctionDx6: &srv6.LocalSID_EndDX6{
						NextHop:           nextHop2.String(), // updated
						OutgoingInterface: ifName,
					},
				},
			},
		},
		{
			name: "nondefault installation vrf table",
			input: &srv6.LocalSID{
				Sid:               sidA.String(),
				InstallationVrfId: vrfTables[0].Id,
				EndFunction: &srv6.LocalSID_BaseEndFunction{
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create
			Expect(srh.AddLocalSid(test.input)).Should(Succeed())

			// Read
			localsids, err := srh.DumpLocalSids()
			t.Logf("received this localsids from dump: %v", localsids)
			Expect(err).ShouldNot(HaveOccurred())
			expected := test.input
			if test.expectedDump != nil {
				expected = test.expectedDump
			}
			Expect(localsids).Should(ConsistOf(expected))

			// Update (for localsids it means delete + create)
			if test.updatedInput != nil {
				Expect(srh.DeleteLocalSid(test.input)).Should(Succeed())
				Expect(srh.AddLocalSid(test.updatedInput)).Should(Succeed())
				localsids, err = srh.DumpLocalSids()
				t.Logf("received this localsids from dump: %v", localsids)
				Expect(err).ShouldNot(HaveOccurred())
				expected := test.updatedInput
				if test.updatedExpectedDump != nil {
					expected = test.updatedExpectedDump
				}
				Expect(localsids).Should(ConsistOf(expected))
			}
			// Delete
			if test.updatedInput != nil {
				Expect(srh.DeleteLocalSid(test.updatedInput)).Should(Succeed())
			} else {
				Expect(srh.DeleteLocalSid(test.input)).Should(Succeed())
			}
		})
	}
}

// sid creates segment ID(=net.IP) from string
func sid(str string) net.IP {
	sid, err := parseIPv6(str)
	if err != nil {
		panic(fmt.Sprintf("can't parse %q into SRv6 SID (IPv6 address)", str))
	}
	return sid
}

// parseIPv6 parses string <str> to IPv6 address (including IPv4 address converted to IPv6 address)
func parseIPv6(str string) (net.IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return nil, fmt.Errorf(" %q is not ip address", str)
	}
	ipv6 := ip.To16()
	if ipv6 == nil {
		return nil, fmt.Errorf(" %q is not ipv6 address", str)
	}
	return ipv6, nil
}
