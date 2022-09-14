// Copyright (c) 2022 Bell Canada, Pantheon Technologies and/or its affiliates.
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
	"fmt"
	"net"
	"testing"

	. "github.com/onsi/gomega"
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/memclnt"
	vpp_sr "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/sr"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/vlib"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls"
	vpp2202 "go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls/vpp2202"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"
)

const (
	ifaceA           = "A"
	ifaceB           = "B"
	ifaceBOutOfidxs  = "B"
	swIndexA         = 1
	invalidIPAddress = "XYZ"
	memif1           = "memif1/1"
	memif2           = "memif2/2"
)

var (
	sidA        = sid("A::")
	sidB        = sid("B::")
	sidC        = sid("C::")
	nextHop     = net.ParseIP("B::").To16()
	nextHopIPv4 = net.ParseIP("1.2.3.4").To4()
)

// TODO add tests for new nhAddr4 field in end behaviours
// TestAddLocalSID tests all cases for method AddLocalSID
func TestAddLocalSID(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name              string
		FailInVPP         bool
		FailInVPPDump     bool
		ExpectFailure     bool
		cliMode           bool // sr-proxy can be se only using CLI -> using VPE binary API to send VPP CLI commands
		MockInterfaceDump []govppapi.Message
		Input             *srv6.LocalSID
		Expected          govppapi.Message
	}{
		{
			Name: "addition with end behaviour",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_BaseEndFunction{
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				},
			},
			Expected: &vpp_sr.SrLocalsidAddDel{
				IsDel:    false,
				Localsid: sidA,
				Behavior: vpp2202.BehaviorEnd,
				FibTable: 10, // installationVrfId
				EndPsp:   true,
			},
		},
		{
			Name: "addition with endX behaviour (ipv6 next hop address)",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionX{
					EndFunctionX: &srv6.LocalSID_EndX{
						Psp:               true,
						NextHop:           nextHop.String(),
						OutgoingInterface: ifaceA,
					},
				},
			},
			Expected: &vpp_sr.SrLocalsidAddDel{
				IsDel:     false,
				Localsid:  sidA,
				Behavior:  vpp2202.BehaviorX,
				FibTable:  10, // installationVrfId
				EndPsp:    true,
				SwIfIndex: interface_types.InterfaceIndex(swIndexA),
				NhAddr:    toAddress(nextHop.String()),
			},
		},
		{
			Name: "addition with endX behaviour (ipv4 next hop address)",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionX{
					EndFunctionX: &srv6.LocalSID_EndX{
						Psp:               true,
						NextHop:           nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
					},
				},
			},
			Expected: &vpp_sr.SrLocalsidAddDel{
				IsDel:     false,
				Localsid:  sidA,
				Behavior:  vpp2202.BehaviorX,
				FibTable:  10, // installationVrfId
				EndPsp:    true,
				SwIfIndex: swIndexA,
				NhAddr:    toAddress(nextHopIPv4.String()),
			},
		},
		{
			Name: "addition with endT behaviour",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionT{
					EndFunctionT: &srv6.LocalSID_EndT{
						Psp:   true,
						VrfId: 11,
					},
				},
			},
			Expected: &vpp_sr.SrLocalsidAddDel{
				IsDel:     false,
				Localsid:  sidA,
				Behavior:  vpp2202.BehaviorT,
				FibTable:  10, // installationVrfId
				SwIfIndex: 11,
				EndPsp:    true,
			},
		},
		{
			Name: "addition with endDX2 behaviour",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDx2{
					EndFunctionDx2: &srv6.LocalSID_EndDX2{
						VlanTag:           1,
						OutgoingInterface: ifaceA,
					},
				},
			},
			Expected: &vpp_sr.SrLocalsidAddDel{
				IsDel:     false,
				Localsid:  sidA,
				Behavior:  vpp2202.BehaviorDX2,
				FibTable:  10, // installationVrfId
				EndPsp:    false,
				VlanIndex: 1,
				SwIfIndex: swIndexA,
			},
		},
		{
			Name: "addition with endDX4 behaviour",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDx4{
					EndFunctionDx4: &srv6.LocalSID_EndDX4{
						NextHop:           nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
					},
				},
			},
			Expected: &vpp_sr.SrLocalsidAddDel{
				IsDel:     false,
				Localsid:  sidA,
				Behavior:  vpp2202.BehaviorDX4,
				FibTable:  10, // installationVrfId
				EndPsp:    false,
				SwIfIndex: swIndexA,
				NhAddr:    toAddress(nextHopIPv4.String()),
			},
		},
		{
			Name: "addition with endDX6 behaviour",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDx6{
					EndFunctionDx6: &srv6.LocalSID_EndDX6{
						NextHop:           nextHop.String(),
						OutgoingInterface: ifaceA,
					},
				},
			},
			Expected: &vpp_sr.SrLocalsidAddDel{
				IsDel:     false,
				Localsid:  sidA,
				Behavior:  vpp2202.BehaviorDX6,
				FibTable:  10, // installationVrfId
				EndPsp:    false,
				SwIfIndex: swIndexA,
				NhAddr:    toAddress(nextHop.String()),
			},
		},
		{
			Name: "addition with endDT4 behaviour",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDt4{
					EndFunctionDt4: &srv6.LocalSID_EndDT4{
						VrfId: 5,
					},
				},
			},
			Expected: &vpp_sr.SrLocalsidAddDel{
				IsDel:     false,
				Localsid:  sidA,
				Behavior:  vpp2202.BehaviorDT4,
				FibTable:  10, // installationVrfId
				SwIfIndex: 5,
				EndPsp:    false,
			},
		},
		{
			Name: "addition with endDT6 behaviour",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDt6{
					EndFunctionDt6: &srv6.LocalSID_EndDT6{
						VrfId: 5,
					},
				},
			},
			Expected: &vpp_sr.SrLocalsidAddDel{
				IsDel:     false,
				Localsid:  sidA,
				Behavior:  vpp2202.BehaviorDT6,
				FibTable:  10, // installationVrfId
				SwIfIndex: 5,
				EndPsp:    false,
			},
		},
		{
			Name:    "addition with endAD behaviour (+ memif interface name translation)",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceA, InterfaceName: memif1},
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceB, InterfaceName: memif2},
			},
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionAd{
					EndFunctionAd: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
			Expected: &vlib.CliInband{
				Cmd: fmt.Sprintf("sr localsid address %v fib-table 10 behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), memif1, memif2),
			},
		},
		{
			Name:    "addition with endAD behaviour for L2 sr-unaware service",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceA, InterfaceName: memif1},
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceB, InterfaceName: memif2},
			},
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionAd{
					EndFunctionAd: &srv6.LocalSID_EndAD{ // missing L3ServiceAddress means it is L2 service
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
			Expected: &vlib.CliInband{
				Cmd: fmt.Sprintf("sr localsid address %v fib-table 10 behavior end.ad oif %v iif %v", sidToStr(sidA), memif1, memif2),
			},
		},
		{
			Name:    "etcd-to-vpp-internal interface name translation for endAD behaviour (local and tap kind of interfaces)",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceA, InterfaceName: "local0"},
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceB, InterfaceName: "tap0"},
			},
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionAd{
					EndFunctionAd: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
			Expected: &vlib.CliInband{
				Cmd: fmt.Sprintf("sr localsid address %v fib-table 10 behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "local0", "tap0"),
			},
		},
		{
			Name:    "etcd-to-vpp-internal interface name translation for endAD behaviour (host and vxlan kind of interfaces)",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceA, InterfaceName: "host0"},
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceB, InterfaceName: "vxlan0"},
			},
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionAd{
					EndFunctionAd: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
			Expected: &vlib.CliInband{
				Cmd: fmt.Sprintf("sr localsid address %v fib-table 10 behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "host0", "vxlan0"),
			},
		},
		{
			Name:    "etcd-to-vpp-internal interface name translation for endAD behaviour (ipsec and vmxnet3 kind of interfaces)",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceA, InterfaceName: "ipsec0"},
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceB, InterfaceName: "vmxnet3-0"},
			},
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionAd{
					EndFunctionAd: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
			Expected: &vlib.CliInband{
				Cmd: fmt.Sprintf("sr localsid address %v fib-table 10 behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "ipsec0", "vmxnet3-0"),
			},
		},
		{
			Name:    "etcd-to-vpp-internal interface name translation for endAD behaviour (loop and unknown kind of interfaces)",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceA, InterfaceName: "loop0"},
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceB, InterfaceName: "unknown0"},
			},
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionAd{
					EndFunctionAd: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: "unknown0", // interface name is taken from vpp internal name
					},
				},
			},
			Expected: &vlib.CliInband{
				Cmd: fmt.Sprintf("sr localsid address %v fib-table 10 behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "loop0", "unknown0"),
			},
		},
		{
			Name:          "fail due to missing end function",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 0,
			},
		},
		{
			Name:          "failure propagation from VPP (doing main VPP call)",
			FailInVPP:     true,
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_BaseEndFunction{
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				},
			},
		},
		{
			Name:          "failure propagation from VPP (doing main VPP call) for SR-proxy (CLI using VPE binary API)",
			FailInVPP:     true,
			ExpectFailure: true,
			cliMode:       true,
			MockInterfaceDump: []govppapi.Message{
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceA, InterfaceName: memif1},
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceB, InterfaceName: memif2},
			},
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionAd{
					EndFunctionAd: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
		},
		{
			Name:          "failure propagation from VPP Dump call",
			FailInVPPDump: true,
			ExpectFailure: true,
			cliMode:       true,
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionAd{
					EndFunctionAd: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
		},
		{
			Name:          "missing SR-proxy outgoing interface in VPP interface dump",
			ExpectFailure: true,
			cliMode:       true,
			MockInterfaceDump: []govppapi.Message{
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceB, InterfaceName: memif2},
			},
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionAd{
					EndFunctionAd: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
		},
		{
			Name:          "missing SR-proxy incoming interface in VPP interface dump",
			ExpectFailure: true,
			cliMode:       true,
			MockInterfaceDump: []govppapi.Message{
				&vpp_ifs.SwInterfaceDetails{Tag: ifaceA, InterfaceName: memif1},
			},
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionAd{
					EndFunctionAd: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endX behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionX{
					EndFunctionX: &srv6.LocalSID_EndX{
						Psp:               true,
						NextHop:           nextHop.String(),
						OutgoingInterface: ifaceBOutOfidxs,
					},
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endX behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionX{
					EndFunctionX: &srv6.LocalSID_EndX{
						Psp:               true,
						NextHop:           invalidIPAddress,
						OutgoingInterface: ifaceA,
					},
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endDX2 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDx2{
					EndFunctionDx2: &srv6.LocalSID_EndDX2{
						VlanTag:           1,
						OutgoingInterface: ifaceBOutOfidxs,
					},
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endDX4 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDx4{
					EndFunctionDx4: &srv6.LocalSID_EndDX4{
						NextHop:           nextHopIPv4.String(),
						OutgoingInterface: ifaceBOutOfidxs,
					},
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endDX4 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDx4{
					EndFunctionDx4: &srv6.LocalSID_EndDX4{
						NextHop:           invalidIPAddress,
						OutgoingInterface: ifaceA,
					},
				},
			},
		},
		{
			Name:          "rejection of IPv6 addresses (addition with endDX4 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDx4{
					EndFunctionDx4: &srv6.LocalSID_EndDX4{
						NextHop:           nextHop.String(),
						OutgoingInterface: ifaceA,
					},
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endDX6 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDx6{
					EndFunctionDx6: &srv6.LocalSID_EndDX6{
						NextHop:           nextHop.String(),
						OutgoingInterface: ifaceBOutOfidxs,
					},
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endDX6 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_EndFunctionDx6{
					EndFunctionDx6: &srv6.LocalSID_EndDX6{
						NextHop:           invalidIPAddress,
						OutgoingInterface: ifaceA,
					},
				},
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)
			// prepare reply
			if td.MockInterfaceDump != nil {
				if td.FailInVPPDump {
					ctx.MockVpp.MockReply(&vpp_sr.SrPolicyDelReply{}) // unexpected type of message creates error (swInterfaceDetail doesn't have way how to indicate failure)
				} else {
					ctx.MockVpp.MockReply(td.MockInterfaceDump...)
					ctx.MockVpp.MockReply(&memclnt.ControlPingReply{})
				}
			}
			if td.cliMode && !td.FailInVPPDump { // SR-proxy can be set only using VPP CLI (-> using VPE binary API to deliver command to VPP)
				if td.FailInVPP {
					ctx.MockVpp.MockReply(&vlib.CliInbandReply{Retval: 1})
				} else {
					ctx.MockVpp.MockReply(&vlib.CliInbandReply{})
				}
			} else { // normal SR binary API
				if td.FailInVPP {
					ctx.MockVpp.MockReply(&vpp_sr.SrLocalsidAddDelReply{Retval: 1})
				} else {
					ctx.MockVpp.MockReply(&vpp_sr.SrLocalsidAddDelReply{})
				}
			}
			// make the call
			err := vppCalls.AddLocalSid(td.Input)
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

// TestDeleteLocalSID tests all cases for method DeleteLocalSID
func TestDeleteLocalSID(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name      string
		Fail      bool
		Input     *srv6.LocalSID
		MockReply govppapi.Message
		Verify    func(error, govppapi.Message)
	}{
		{
			Name: "simple delete of local sid (using vrf table with id 0)",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_BaseEndFunction{
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				},
			},
			MockReply: &vpp_sr.SrLocalsidAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&vpp_sr.SrLocalsidAddDel{
					IsDel:    true,
					Localsid: sidA,
					FibTable: 0,
				}))
			},
		},
		{
			Name: "simple delete of local sid (using vrf table with nonzero id)",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 10,
				EndFunction: &srv6.LocalSID_BaseEndFunction{
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				},
			},
			MockReply: &vpp_sr.SrLocalsidAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&vpp_sr.SrLocalsidAddDel{
					IsDel:    true,
					Localsid: sidA,
					FibTable: 10,
				}))
			},
		},
		{
			Name: "failure propagation from VPP",
			Input: &srv6.LocalSID{
				Sid:               sidToStr(sidA),
				InstallationVrfId: 0,
				EndFunction: &srv6.LocalSID_BaseEndFunction{
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				},
			},
			MockReply: &vpp_sr.SrLocalsidAddDelReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)
			// prepare for case
			ctx.MockVpp.MockReply(td.MockReply)
			// make the call and verify
			err := vppCalls.DeleteLocalSid(td.Input)
			td.Verify(err, ctx.MockChannel.Msg)
		})
	}
}

// TestSetEncapsSourceAddress tests all cases for method SetEncapsSourceAddress
func TestSetEncapsSourceAddress(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name      string
		Fail      bool
		Address   string
		MockReply govppapi.Message
		Verify    func(error, govppapi.Message)
	}{
		{
			Name:      "simple SetEncapsSourceAddress",
			Address:   nextHop.String(),
			MockReply: &vpp_sr.SrSetEncapSourceReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&vpp_sr.SrSetEncapSource{
					EncapsSource: sid(nextHop.String()),
				}))
			},
		},
		{
			Name:      "invalid IP address",
			Address:   invalidIPAddress,
			MockReply: &vpp_sr.SrSetEncapSourceReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:      "failure propagation from VPP",
			Address:   nextHop.String(),
			MockReply: &vpp_sr.SrSetEncapSourceReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)

			ctx.MockVpp.MockReply(td.MockReply)
			err := vppCalls.SetEncapsSourceAddress(td.Address)
			td.Verify(err, ctx.MockChannel.Msg)
		})
	}
}

// TestAddPolicy tests all cases for method AddPolicy
func TestAddPolicy(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name        string
		Fail        bool
		Policy      *srv6.Policy
		MockReplies []govppapi.Message
		Verify      func(error, []govppapi.Message)
	}{
		{
			Name:        "simple SetAddPolicy",
			Policy:      policy(sidA[:], 10, false, true, policySegmentList(1, sidA[:], sidB[:], sidC[:])),
			MockReplies: []govppapi.Message{&vpp_sr.SrPolicyAddReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsgs).To(HaveLen(1))
				Expect(catchedMsgs[0]).To(Equal(&vpp_sr.SrPolicyAdd{
					BsidAddr: *(&sidA),
					FibTable: 10, // installationVrfId
					IsSpray:  false,
					IsEncap:  true,
					Sids: vpp_sr.Srv6SidList{
						Weight:  1,
						NumSids: 3,
						Sids:    [16]ip_types.IP6Address{sidA, sidB, sidC},
					},
				}))
			},
		},
		{
			Name: "adding policy with multiple segment lists",
			Policy: policy(sidA[:], 10, false, true,
				policySegmentList(1, sidA[:], sidB[:], sidC[:]), policySegmentList(1, sidB[:], sidC[:], sidA[:])),
			MockReplies: []govppapi.Message{&vpp_sr.SrPolicyAddReply{}, &vpp_sr.SrPolicyModReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsgs).To(HaveLen(2))
				Expect(catchedMsgs[0]).To(Equal(&vpp_sr.SrPolicyAdd{
					BsidAddr: sidA,
					FibTable: 10, // installationVrfId
					IsSpray:  false,
					IsEncap:  true,
					Sids: vpp_sr.Srv6SidList{
						Weight:  1,
						NumSids: 3,
						Sids: [16]ip_types.IP6Address{
							sidA, sidB, sidC,
						},
					},
				}))
				Expect(catchedMsgs[1]).To(Equal(&vpp_sr.SrPolicyMod{
					BsidAddr:  sidA,
					Operation: vpp2202.AddSRList,
					FibTable:  10, // installationVrfId
					Sids: vpp_sr.Srv6SidList{
						Weight:  1,
						NumSids: 3,
						Sids:    [16]ip_types.IP6Address{sidB, sidC, sidA},
					},
				}))
			},
		},
		{
			Name:        "failing when adding policy with empty segment lists",
			Policy:      policy(sidA[:], 10, false, true),
			MockReplies: []govppapi.Message{&vpp_sr.SrPolicyAddReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid binding SID in policy",
			Policy: &srv6.Policy{
				Bsid:              invalidIPAddress,
				InstallationVrfId: 10,
				SprayBehaviour:    false,
				SrhEncapsulation:  true,
				SegmentLists: []*srv6.Policy_SegmentList{
					{
						Weight:   1,
						Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
					},
				},
			},
			MockReplies: []govppapi.Message{&vpp_sr.SrPolicyAddReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid SID (not IP address) in first segment list",
			Policy: policy(sidA[:], 10, false, true,
				&srv6.Policy_SegmentList{
					Weight:   1,
					Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
				}),
			MockReplies: []govppapi.Message{&vpp_sr.SrPolicyAddReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid SID (not IP address) in non-first segment list",
			Policy: policy(sidA[:], 10, false, true,
				policySegmentList(1, sidA[:], sidB[:], sidC[:]),
				&srv6.Policy_SegmentList{
					Weight:   1,
					Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
				}),
			MockReplies: []govppapi.Message{&vpp_sr.SrPolicyAddReply{}, &vpp_sr.SrPolicyModReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:        "failure propagation from VPP",
			Policy:      policy(sidA[:], 0, true, true, policySegmentList(1, sidA[:], sidB[:], sidC[:])),
			MockReplies: []govppapi.Message{&vpp_sr.SrPolicyAddReply{Retval: 1}},
			Verify: func(err error, msgs []govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)
			// prepare reply, make call and verify
			for _, reply := range td.MockReplies {
				ctx.MockVpp.MockReply(reply)
			}
			err := vppCalls.AddPolicy(td.Policy)
			td.Verify(err, ctx.MockChannel.Msgs)
		})
	}
}

// TestDeletePolicy tests all cases for method DeletePolicy
func TestDeletePolicy(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name      string
		BSID      net.IP
		MockReply govppapi.Message
		Verify    func(error, govppapi.Message)
	}{
		{
			Name:      "simple delete of policy",
			BSID:      sidA[:],
			MockReply: &vpp_sr.SrPolicyDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&vpp_sr.SrPolicyDel{
					BsidAddr: sidA,
				}))
			},
		},
		{
			Name:      "failure propagation from VPP",
			BSID:      sidA[:],
			MockReply: &vpp_sr.SrPolicyDelReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)
			// data and prepare case
			policy := policy(td.BSID, 0, true, true, policySegmentList(1, sidA[:], sidB[:], sidC[:]))
			vppCalls.AddPolicy(policy)
			ctx.MockVpp.MockReply(td.MockReply)
			// make the call and verify
			err := vppCalls.DeletePolicy(td.BSID)
			td.Verify(err, ctx.MockChannel.Msg)
		})
	}
}

// TestAddPolicySegmentList tests all cases for method AddPolicySegment
func TestAddPolicySegmentList(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name              string
		Policy            *srv6.Policy
		PolicySegmentList *srv6.Policy_SegmentList
		MockReply         govppapi.Message
		Verify            func(error, govppapi.Message)
	}{
		{
			Name:              "simple addition of policy segment",
			Policy:            policy(sidA[:], 10, false, true),
			PolicySegmentList: policySegmentList(1, sidA[:], sidB[:], sidC[:]),
			MockReply:         &vpp_sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&vpp_sr.SrPolicyMod{
					BsidAddr:  sidA,
					Operation: vpp2202.AddSRList,
					FibTable:  10, // installationVrfId
					Sids: vpp_sr.Srv6SidList{
						Weight:  1,
						NumSids: 3,
						Sids:    [16]ip_types.IP6Address{sidA, sidB, sidC},
					},
				}))
			},
		},
		{
			Name:   "invalid SID (not IP address) in segment list",
			Policy: policy(sidA[:], 10, false, true),
			PolicySegmentList: &srv6.Policy_SegmentList{
				Weight:   1,
				Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
			},
			MockReply: &vpp_sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid binding SID (not IP address) in policy",
			Policy: &srv6.Policy{
				Bsid:              invalidIPAddress,
				InstallationVrfId: 10,
				SprayBehaviour:    false,
				SrhEncapsulation:  true,
			},
			PolicySegmentList: policySegmentList(1, sidA[:], sidB[:], sidC[:]),
			MockReply:         &vpp_sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:              "failure propagation from VPP",
			Policy:            policy(sidA[:], 0, true, true),
			PolicySegmentList: policySegmentList(1, sidA[:], sidB[:], sidC[:]),
			MockReply:         &vpp_sr.SrPolicyModReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)
			// prepare reply, make call and verify
			ctx.MockVpp.MockReply(td.MockReply)
			err := vppCalls.AddPolicySegmentList(td.PolicySegmentList, td.Policy)
			td.Verify(err, ctx.MockChannel.Msg)
		})
	}
}

// TestDeletePolicySegmentList tests all cases for method DeletePolicySegment
func TestDeletePolicySegmentList(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name              string
		Policy            *srv6.Policy
		PolicySegmentList *srv6.Policy_SegmentList
		SegmentIndex      uint32
		MockReply         govppapi.Message
		Verify            func(error, govppapi.Message)
	}{
		{
			Name:              "simple deletion of policy segment",
			Policy:            policy(sidA[:], 10, false, true, policySegmentList(1, sidA[:], sidB[:], sidC[:])),
			PolicySegmentList: policySegmentList(1, sidA[:], sidB[:], sidC[:]),
			SegmentIndex:      111,
			MockReply:         &vpp_sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&vpp_sr.SrPolicyMod{
					BsidAddr:  sidA,
					Operation: vpp2202.DeleteSRList,
					SlIndex:   111,
					FibTable:  10, // installationVrfId
					Sids: vpp_sr.Srv6SidList{
						Weight:  1,
						NumSids: 3,
						Sids:    [16]ip_types.IP6Address{sidA, sidB, sidC},
					},
				}))
			},
		},
		{
			Name: "invalid SID (not IP address) in segment list",
			Policy: policy(sidA[:], 10, false, true,
				&srv6.Policy_SegmentList{
					Weight:   1,
					Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
				}),
			PolicySegmentList: &srv6.Policy_SegmentList{
				Weight:   1,
				Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
			},
			SegmentIndex: 111,
			MockReply:    &vpp_sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:              "failure propagation from VPP",
			Policy:            policy(sidA[:], 0, true, true, policySegmentList(1, sidA[:], sidB[:], sidC[:])),
			PolicySegmentList: policySegmentList(1, sidA[:], sidB[:], sidC[:]),
			SegmentIndex:      111,
			MockReply:         &vpp_sr.SrPolicyModReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)
			// prepare reply, make call and verify
			ctx.MockVpp.MockReply(td.MockReply)
			err := vppCalls.DeletePolicySegmentList(td.PolicySegmentList, td.SegmentIndex, td.Policy)
			td.Verify(err, ctx.MockChannel.Msg)
		})
	}
}

// TestAddSteering tests all cases for method AddSteering
func TestAddSteering(t *testing.T) {
	testAddRemoveSteering(t, false)
}

// TestRemoveSteering tests all cases for method RemoveSteering
func TestRemoveSteering(t *testing.T) {
	testAddRemoveSteering(t, true)
}

func testAddRemoveSteering(t *testing.T, removal bool) {
	action := "addition"
	if removal {
		action = "removal"
	}
	// Prepare different cases
	cases := []struct {
		Name      string
		Steering  *srv6.Steering
		MockReply govppapi.Message
		Verify    func(error, govppapi.Message)
	}{
		{
			Name: action + " of IPv6 L3 steering",
			Steering: &srv6.Steering{
				PolicyRef: &srv6.Steering_PolicyBsid{
					PolicyBsid: sidToStr(sidA),
				},
				Traffic: &srv6.Steering_L3Traffic_{
					L3Traffic: &srv6.Steering_L3Traffic{
						InstallationVrfId: 10,
						PrefixAddress:     "1::/64",
					},
				},
			},
			MockReply: &vpp_sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&vpp_sr.SrSteeringAddDel{
					IsDel:         removal,
					BsidAddr:      sidA,
					SrPolicyIndex: uint32(0),
					TableID:       10,
					TrafficType:   vpp2202.SteerTypeIPv6,
					Prefix:        ip_types.Prefix{Address: toAddress("1::"), Len: 64},
				}))
			},
		},
		{
			Name: action + " of IPv4 L3 steering",
			Steering: &srv6.Steering{
				PolicyRef: &srv6.Steering_PolicyBsid{
					PolicyBsid: sidToStr(sidA),
				},
				Traffic: &srv6.Steering_L3Traffic_{
					L3Traffic: &srv6.Steering_L3Traffic{
						InstallationVrfId: 10,
						PrefixAddress:     "1.2.3.4/24",
					},
				},
			},
			MockReply: &vpp_sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&vpp_sr.SrSteeringAddDel{
					IsDel:         removal,
					BsidAddr:      sidA,
					SrPolicyIndex: uint32(0),
					TableID:       10,
					TrafficType:   vpp2202.SteerTypeIPv4,
					Prefix:        ip_types.Prefix{Address: toAddress("1.2.3.4"), Len: 24},
				}))
			},
		},
		{
			Name: action + " of L2 steering",
			Steering: &srv6.Steering{
				PolicyRef: &srv6.Steering_PolicyBsid{
					PolicyBsid: sidToStr(sidA),
				},
				Traffic: &srv6.Steering_L2Traffic_{
					L2Traffic: &srv6.Steering_L2Traffic{
						InterfaceName: ifaceA,
					},
				},
			},
			MockReply: &vpp_sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&vpp_sr.SrSteeringAddDel{
					IsDel:         removal,
					BsidAddr:      sidA,
					SrPolicyIndex: uint32(0),
					TrafficType:   vpp2202.SteerTypeL2,
					SwIfIndex:     swIndexA,
				}))
			},
		},
		{
			Name: action + " of IPv6 L3 steering with Policy referencing by index",
			Steering: &srv6.Steering{
				PolicyRef: &srv6.Steering_PolicyIndex{
					PolicyIndex: 20,
				},
				Traffic: &srv6.Steering_L3Traffic_{
					L3Traffic: &srv6.Steering_L3Traffic{
						InstallationVrfId: 10,
						PrefixAddress:     "1::/64",
					},
				},
			},
			MockReply: &vpp_sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&vpp_sr.SrSteeringAddDel{
					IsDel:         removal,
					SrPolicyIndex: uint32(20),
					TableID:       10,
					TrafficType:   vpp2202.SteerTypeIPv6,
					Prefix:        ip_types.Prefix{Address: toAddress("1::"), Len: 64},
				}))
			},
		},
		{
			Name: "missing policy reference ( " + action + " of IPv6 L3 steering)",
			Steering: &srv6.Steering{
				Traffic: &srv6.Steering_L3Traffic_{
					L3Traffic: &srv6.Steering_L3Traffic{
						InstallationVrfId: 10,
						PrefixAddress:     "1::/64",
					},
				},
			},
			MockReply: &vpp_sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "missing traffic ( " + action + " of IPv6 L3 steering)",
			Steering: &srv6.Steering{
				PolicyRef: &srv6.Steering_PolicyBsid{
					PolicyBsid: sidToStr(sidA),
				},
			},
			MockReply: &vpp_sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid prefix (" + action + " of IPv4 L3 steering)",
			Steering: &srv6.Steering{
				PolicyRef: &srv6.Steering_PolicyBsid{
					PolicyBsid: sidToStr(sidA),
				},
				Traffic: &srv6.Steering_L3Traffic_{
					L3Traffic: &srv6.Steering_L3Traffic{
						InstallationVrfId: 10,
						PrefixAddress:     invalidIPAddress,
					},
				},
			},
			MockReply: &vpp_sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "interface without index (" + action + " of L2 steering)",
			Steering: &srv6.Steering{
				PolicyRef: &srv6.Steering_PolicyBsid{
					PolicyBsid: sidToStr(sidA),
				},
				Traffic: &srv6.Steering_L2Traffic_{
					L2Traffic: &srv6.Steering_L2Traffic{
						InterfaceName: ifaceBOutOfidxs,
					},
				},
			},
			MockReply: &vpp_sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid BSID (not IP address) as policy reference",
			Steering: &srv6.Steering{
				PolicyRef: &srv6.Steering_PolicyBsid{
					PolicyBsid: invalidIPAddress,
				},
				Traffic: &srv6.Steering_L3Traffic_{
					L3Traffic: &srv6.Steering_L3Traffic{
						InstallationVrfId: 10,
						PrefixAddress:     "1::/64",
					},
				},
			},
			MockReply: &vpp_sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "failure propagation from VPP",
			Steering: &srv6.Steering{
				PolicyRef: &srv6.Steering_PolicyBsid{
					PolicyBsid: sidToStr(sidA),
				},
				Traffic: &srv6.Steering_L3Traffic_{
					L3Traffic: &srv6.Steering_L3Traffic{
						InstallationVrfId: 10,
						PrefixAddress:     "1::/64",
					},
				},
			},
			MockReply: &vpp_sr.SrSteeringAddDelReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)
			// prepare reply, make call and verify
			ctx.MockVpp.MockReply(td.MockReply)
			var err error
			if removal {
				err = vppCalls.RemoveSteering(td.Steering)
			} else {
				err = vppCalls.AddSteering(td.Steering)
			}
			td.Verify(err, ctx.MockChannel.Msg)
		})
	}
}

// RetrievePolicyIndexInfo tests all cases for method RetrievePolicyIndexInfo
func TestRetrievePolicyIndexInfo(t *testing.T) {
	correctCLIOutput := `
[4].-	BSID: a::

	Behavior: SRH insertion

	Type: Spray

	FIB table: 0

	Segment Lists:

  	[2].- < a::, b::, c::,  > weight: 1
  	[3].- < b::, b::, c::,  > weight: 1
  	[4].- < c::, b::, c::,  > weight: 1

-----------
`
	correctPolicyIndex := uint32(4)
	segmentListABC := policySegmentList(1, sidA[:], sidB[:], sidC[:])
	segmentListBBC := policySegmentList(1, sidB[:], sidB[:], sidC[:])
	notExistingSegmentListCCC := policySegmentList(1, sidC[:], sidC[:], sidC[:])

	// Prepare different cases
	cases := []struct {
		Name                       string
		Policy                     *srv6.Policy
		MockReply                  govppapi.Message
		ExpectedPolicyIndex        uint32
		ExpectedSegmentListIndexes map[*srv6.Policy_SegmentList]uint32
		ExpectingFailure           bool
	}{
		{
			Name:   "basic successful index retrieval",
			Policy: policy(sidA[:], 10, false, true, segmentListABC, segmentListBBC),
			MockReply: &vlib.CliInbandReply{
				Reply:  correctCLIOutput,
				Retval: 0,
			},
			ExpectedPolicyIndex:        correctPolicyIndex,
			ExpectedSegmentListIndexes: map[*srv6.Policy_SegmentList]uint32{segmentListABC: uint32(2), segmentListBBC: uint32(3)},
		},
		{
			Name:             "failure propagation from VPP",
			Policy:           policy(sidA[:], 10, false, true, segmentListABC, segmentListBBC),
			MockReply:        &vlib.CliInbandReply{Retval: 1},
			ExpectingFailure: true,
		},
		{
			Name:   "searching for not existing policy ",
			Policy: policy(sidC[:], 10, false, true, segmentListABC, segmentListBBC),
			MockReply: &vlib.CliInbandReply{
				Reply:  correctCLIOutput,
				Retval: 0,
			},
			ExpectingFailure: true,
		},
		{
			Name:   "searching for not existing policy segment list",
			Policy: policy(sidA[:], 10, false, true, notExistingSegmentListCCC),
			MockReply: &vlib.CliInbandReply{
				Reply:  correctCLIOutput,
				Retval: 0,
			},
			ExpectingFailure: true,
		},
	}
	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			ctx, vppCalls := setup(t)
			defer teardown(ctx)
			// prepare reply, make call and verify
			ctx.MockVpp.MockReply(td.MockReply)
			resultPolicyIndex, resultSlIndexes, err := vppCalls.RetrievePolicyIndexInfo(td.Policy)
			Expect(ctx.MockChannel.Msg).To(Equal(&vlib.CliInband{
				Cmd: "sh sr policies",
			}))
			if td.ExpectingFailure {
				Expect(err).Should(HaveOccurred())
			} else {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resultPolicyIndex).To(Equal(td.ExpectedPolicyIndex))
				Expect(resultSlIndexes).To(Equal(td.ExpectedSegmentListIndexes))
			}
		})
	}
}

func setup(t *testing.T) (*vppmock.TestCtx, vppcalls.SRv6VppAPI) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test")
	swIfIndex := ifaceidx.NewIfaceIndex(log, "test")
	swIfIndex.Put(ifaceA, &ifaceidx.IfaceMetadata{SwIfIndex: swIndexA})
	vppCalls := vpp2202.NewSRv6VppHandler(ctx.MockVPPClient, swIfIndex, log)
	return ctx, vppCalls
}

func teardown(ctx *vppmock.TestCtx) {
	ctx.TeardownTestCtx()
}

func sid(str string) ip_types.IP6Address {
	bsid, err := parseIPv6(str)
	if err != nil {
		panic(fmt.Sprintf("can't parse %q into SRv6 BSID (IPv6 address)", str))
	}
	var ip ip_types.IP6Address
	copy(ip[:], bsid)
	return ip
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

func policy(bsid srv6.SID, installationVrfId uint32, sprayBehaviour bool, srhEncapsulation bool, segmentLists ...*srv6.Policy_SegmentList) *srv6.Policy {
	return &srv6.Policy{
		Bsid:              bsid.String(),
		InstallationVrfId: installationVrfId,
		SprayBehaviour:    sprayBehaviour,
		SrhEncapsulation:  srhEncapsulation,
		SegmentLists:      segmentLists,
	}
}

func policySegmentList(weight uint32, sids ...srv6.SID) *srv6.Policy_SegmentList {
	segments := make([]string, len(sids))
	for i, sid := range sids {
		segments[i] = sid.String()
	}

	return &srv6.Policy_SegmentList{
		Weight:   weight,
		Segments: segments,
	}
}

func boolToUint(input bool) uint8 {
	if input {
		return uint8(1)
	}
	return uint8(0)
}

func sidToStr(sid ip_types.IP6Address) string {
	return srv6.SID(sid[:]).String()
}

func toAddress(ip interface{}) (addr ip_types.Address) {
	switch ip := ip.(type) {
	case string:
		addr, _ = vpp2202.IPToAddress(ip)
	case net.IP:
		addr, _ = vpp2202.IPToAddress(ip.String())
	default:
		panic(fmt.Sprintf("cannot convert to ip_types.Address from type %T", ip))
	}
	return
}
