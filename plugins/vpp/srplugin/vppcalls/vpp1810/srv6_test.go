// Copyright (c) 2019 Bell Canada, Pantheon Technologies and/or its affiliates.
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

package vpp1810_test

import (
	"fmt"
	"net"
	"testing"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/logrus"
	srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/sr"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/vppcalls/vpp1810"
	"github.com/ligato/vpp-agent/plugins/vpp/vppcallmock"
	. "github.com/onsi/gomega"
)

const (
	ifaceA                  = "A"
	ifaceB                  = "B"
	ifaceBOutOfidxs         = "B"
	swIndexA         uint32 = 1
	invalidIPAddress        = "XYZ"
	memif1                  = "memif1/1"
	memif2                  = "memif2/2"
)

var (
	sidA        = *sid("A::")
	sidB        = *sid("B::")
	sidC        = *sid("C::")
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
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_BaseEndFunction{
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:    0,
				Localsid: sidA,
				Behavior: vpp1810.BehaviorEnd,
				FibTable: 10,
				EndPsp:   1,
			},
		},
		{
			Name: "addition with endX behaviour (ipv6 next hop address)",
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_X{
					EndFunction_X: &srv6.LocalSID_EndX{
						Psp:               true,
						NextHop:           nextHop.String(),
						OutgoingInterface: ifaceA,
					},
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vpp1810.BehaviorX,
				FibTable:  10,
				EndPsp:    1,
				SwIfIndex: swIndexA,
				NhAddr6:   nextHop,
			},
		},
		{
			Name: "addition with endX behaviour (ipv4 next hop address)",
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_X{
					EndFunction_X: &srv6.LocalSID_EndX{
						Psp:               true,
						NextHop:           nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
					},
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vpp1810.BehaviorX,
				FibTable:  10,
				EndPsp:    1,
				SwIfIndex: swIndexA,
				NhAddr4:   nextHopIPv4,
			},
		},
		{
			Name: "addition with endT behaviour",
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_T{
					EndFunction_T: &srv6.LocalSID_EndT{
						Psp:   true,
						VrfId: 11,
					},
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vpp1810.BehaviorT,
				FibTable:  10,
				SwIfIndex: 11,
				EndPsp:    1,
			},
		},
		{
			Name: "addition with endDX2 behaviour",
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DX2{
					EndFunction_DX2: &srv6.LocalSID_EndDX2{
						VlanTag:           1,
						OutgoingInterface: ifaceA,
					},
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vpp1810.BehaviorDX2,
				FibTable:  10,
				EndPsp:    0,
				VlanIndex: 1,
				SwIfIndex: swIndexA,
			},
		},
		{
			Name: "addition with endDX4 behaviour",
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DX4{
					EndFunction_DX4: &srv6.LocalSID_EndDX4{
						NextHop:           nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
					},
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vpp1810.BehaviorDX4,
				FibTable:  10,
				EndPsp:    0,
				SwIfIndex: swIndexA,
				NhAddr4:   nextHopIPv4,
			},
		},
		{
			Name: "addition with endDX6 behaviour",
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DX6{
					EndFunction_DX6: &srv6.LocalSID_EndDX6{
						NextHop:           nextHop.String(),
						OutgoingInterface: ifaceA,
					},
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vpp1810.BehaviorDX6,
				FibTable:  10,
				EndPsp:    0,
				SwIfIndex: swIndexA,
				NhAddr6:   nextHop,
			},
		},
		{
			Name: "addition with endDT4 behaviour",
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DT4{
					EndFunction_DT4: &srv6.LocalSID_EndDT4{
						VrfId: 5,
					},
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vpp1810.BehaviorDT4,
				FibTable:  10,
				SwIfIndex: 5,
				EndPsp:    0,
			},
		},
		{
			Name: "addition with endDT6 behaviour",
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DT6{
					EndFunction_DT6: &srv6.LocalSID_EndDT6{
						VrfId: 5,
					},
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vpp1810.BehaviorDT6,
				FibTable:  10,
				SwIfIndex: 5,
				EndPsp:    0,
			},
		},
		{
			Name:    "addition with endAD behaviour (+ memif interface name translation)",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceA), InterfaceName: toIFaceByte(memif1)},
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceB), InterfaceName: toIFaceByte(memif2)},
			},
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_AD{
					EndFunction_AD: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
			Expected: &vpe.CliInband{
				Cmd:    []byte(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), memif1, memif2)),
				Length: uint32(len(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), memif1, memif2))),
			},
		},
		{
			Name:    "addition with endAD behaviour for L2 sr-unaware service",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceA), InterfaceName: toIFaceByte(memif1)},
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceB), InterfaceName: toIFaceByte(memif2)},
			},
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_AD{
					EndFunction_AD: &srv6.LocalSID_EndAD{ //missing L3ServiceAddress means it is L2 service
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
			Expected: &vpe.CliInband{
				Cmd:    []byte(fmt.Sprintf("sr localsid address %v behavior end.ad oif %v iif %v", sidToStr(sidA), memif1, memif2)),
				Length: uint32(len(fmt.Sprintf("sr localsid address %v behavior end.ad oif %v iif %v", sidToStr(sidA), memif1, memif2))),
			},
		},
		{
			Name:    "etcd-to-vpp-internal interface name translation for endAD behaviour (local and tap kind of interfaces)",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceA), InterfaceName: toIFaceByte("local0")},
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceB), InterfaceName: toIFaceByte("tap0")},
			},
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_AD{
					EndFunction_AD: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
			Expected: &vpe.CliInband{
				Cmd:    []byte(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "local0", "tap0")),
				Length: uint32(len(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "local0", "tap0"))),
			},
		},
		{
			Name:    "etcd-to-vpp-internal interface name translation for endAD behaviour (host and vxlan kind of interfaces)",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceA), InterfaceName: toIFaceByte("host0")},
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceB), InterfaceName: toIFaceByte("vxlan0")},
			},
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_AD{
					EndFunction_AD: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
			Expected: &vpe.CliInband{
				Cmd:    []byte(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "host0", "vxlan0")),
				Length: uint32(len(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "host0", "vxlan0"))),
			},
		},
		{
			Name:    "etcd-to-vpp-internal interface name translation for endAD behaviour (ipsec and vmxnet3 kind of interfaces)",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceA), InterfaceName: toIFaceByte("ipsec0")},
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceB), InterfaceName: toIFaceByte("vmxnet3-0")},
			},
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_AD{
					EndFunction_AD: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: ifaceB,
					},
				},
			},
			Expected: &vpe.CliInband{
				Cmd:    []byte(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "ipsec0", "vmxnet3-0")),
				Length: uint32(len(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "ipsec0", "vmxnet3-0"))),
			},
		},
		{
			Name:    "etcd-to-vpp-internal interface name translation for endAD behaviour (loop and unknown kind of interfaces)",
			cliMode: true,
			MockInterfaceDump: []govppapi.Message{
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceA), InterfaceName: toIFaceByte("loop0")},
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceB), InterfaceName: toIFaceByte("unknown0")},
			},
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_AD{
					EndFunction_AD: &srv6.LocalSID_EndAD{
						L3ServiceAddress:  nextHopIPv4.String(),
						OutgoingInterface: ifaceA,
						IncomingInterface: "unknown0", // interface name is taken from vpp internal name
					},
				},
			},
			Expected: &vpe.CliInband{
				Cmd:    []byte(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "loop0", "unknown0")),
				Length: uint32(len(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), "loop0", "unknown0"))),
			},
		},
		{
			Name:          "fail due to missing end function",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 0,
			},
		},
		{
			Name:          "failure propagation from VPP (doing main VPP call)",
			FailInVPP:     true,
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 0,
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
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceA), InterfaceName: toIFaceByte(memif1)},
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceB), InterfaceName: toIFaceByte(memif2)},
			},
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_AD{
					EndFunction_AD: &srv6.LocalSID_EndAD{
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
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_AD{
					EndFunction_AD: &srv6.LocalSID_EndAD{
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
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceB), InterfaceName: toIFaceByte(memif2)},
			},
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_AD{
					EndFunction_AD: &srv6.LocalSID_EndAD{
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
				&interfaces.SwInterfaceDetails{Tag: toIFaceByte(ifaceA), InterfaceName: toIFaceByte(memif1)},
			},
			Input: &srv6.LocalSID{
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_AD{
					EndFunction_AD: &srv6.LocalSID_EndAD{
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
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_X{
					EndFunction_X: &srv6.LocalSID_EndX{
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
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_X{
					EndFunction_X: &srv6.LocalSID_EndX{
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
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DX2{
					EndFunction_DX2: &srv6.LocalSID_EndDX2{
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
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DX4{
					EndFunction_DX4: &srv6.LocalSID_EndDX4{
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
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DX4{
					EndFunction_DX4: &srv6.LocalSID_EndDX4{
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
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DX4{
					EndFunction_DX4: &srv6.LocalSID_EndDX4{
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
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DX6{
					EndFunction_DX6: &srv6.LocalSID_EndDX6{
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
				Sid:        sidToStr(sidA),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_EndFunction_DX6{
					EndFunction_DX6: &srv6.LocalSID_EndDX6{
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
					ctx.MockVpp.MockReply(&sr.SrPolicyDelReply{}) //unexpected type of message creates error (swInterfaceDetail doesn't have way how to indicate failure)
				} else {
					ctx.MockVpp.MockReply(td.MockInterfaceDump...)
					ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
				}
			}
			if td.cliMode && !td.FailInVPPDump { // SR-proxy can be set only using VPP CLI (-> using VPE binary API to deliver command to VPP)
				if td.FailInVPP {
					ctx.MockVpp.MockReply(&vpe.CliInbandReply{Retval: 1})
				} else {
					ctx.MockVpp.MockReply(&vpe.CliInbandReply{})
				}
			} else { // normal SR binary API
				if td.FailInVPP {
					ctx.MockVpp.MockReply(&sr.SrLocalsidAddDelReply{Retval: 1})
				} else {
					ctx.MockVpp.MockReply(&sr.SrLocalsidAddDelReply{})
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
		Sid       net.IP
		MockReply govppapi.Message
		Verify    func(error, govppapi.Message)
	}{
		{
			Name:      "simple delete of local sid",
			Sid:       sidA.Addr,
			MockReply: &sr.SrLocalsidAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrLocalsidAddDel{
					IsDel:    1,
					Localsid: sidA,
				}))
			},
		},
		{
			Name:      "failure propagation from VPP",
			Sid:       sidA.Addr,
			MockReply: &sr.SrLocalsidAddDelReply{Retval: 1},
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
			localsid := &srv6.LocalSID{
				Sid:        td.Sid.String(),
				FibTableId: 10,
				EndFunction: &srv6.LocalSID_BaseEndFunction{
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				},
			}
			vppCalls.AddLocalSid(localsid)
			ctx.MockVpp.MockReply(td.MockReply)
			// make the call and verify
			err := vppCalls.DeleteLocalSid(td.Sid)
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
			MockReply: &sr.SrSetEncapSourceReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSetEncapSource{
					EncapsSource: nextHop,
				}))
			},
		},
		{
			Name:      "invalid IP address",
			Address:   invalidIPAddress,
			MockReply: &sr.SrSetEncapSourceReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:      "failure propagation from VPP",
			Address:   nextHop.String(),
			MockReply: &sr.SrSetEncapSourceReply{Retval: 1},
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
			Policy:      policy(sidA.Addr, 10, false, true, policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr)),
			MockReplies: []govppapi.Message{&sr.SrPolicyAddReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsgs).To(HaveLen(1))
				Expect(catchedMsgs[0]).To(Equal(&sr.SrPolicyAdd{
					BsidAddr: sidA.Addr,
					FibTable: 10,
					Type:     boolToUint(false),
					IsEncap:  boolToUint(true),
					Sids: sr.Srv6SidList{
						Weight:  1,
						NumSids: 3,
						Sids:    []sr.Srv6Sid{{Addr: sidA.Addr}, {Addr: sidB.Addr}, {Addr: sidC.Addr}},
					},
				}))
			},
		},
		{
			Name: "adding policy with multiple segment lists",
			Policy: policy(sidA.Addr, 10, false, true,
				policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr), policySegmentList(1, sidB.Addr, sidC.Addr, sidA.Addr)),
			MockReplies: []govppapi.Message{&sr.SrPolicyAddReply{}, &sr.SrPolicyModReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsgs).To(HaveLen(2))
				Expect(catchedMsgs[0]).To(Equal(&sr.SrPolicyAdd{
					BsidAddr: sidA.Addr,
					FibTable: 10,
					Type:     boolToUint(false),
					IsEncap:  boolToUint(true),
					Sids: sr.Srv6SidList{
						Weight:  1,
						NumSids: 3,
						Sids:    []sr.Srv6Sid{{Addr: sidA.Addr}, {Addr: sidB.Addr}, {Addr: sidC.Addr}},
					},
				}))
				Expect(catchedMsgs[1]).To(Equal(&sr.SrPolicyMod{
					BsidAddr:  sidA.Addr,
					Operation: vpp1810.AddSRList,
					FibTable:  10,
					Sids: sr.Srv6SidList{
						Weight:  1,
						NumSids: 3,
						Sids:    []sr.Srv6Sid{{Addr: sidB.Addr}, {Addr: sidC.Addr}, {Addr: sidA.Addr}},
					},
				}))
			},
		},
		{
			Name:        "failing when adding policy with empty segment lists",
			Policy:      policy(sidA.Addr, 10, false, true),
			MockReplies: []govppapi.Message{&sr.SrPolicyAddReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid binding SID in policy",
			Policy: &srv6.Policy{
				Bsid:             invalidIPAddress,
				FibTableId:       10,
				SprayBehaviour:   false,
				SrhEncapsulation: true,
				SegmentLists: []*srv6.Policy_SegmentList{
					&srv6.Policy_SegmentList{
						Weight:   1,
						Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
					},
				},
			},
			MockReplies: []govppapi.Message{&sr.SrPolicyAddReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid SID (not IP address) in first segment list",
			Policy: policy(sidA.Addr, 10, false, true,
				&srv6.Policy_SegmentList{
					Weight:   1,
					Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
				}),
			MockReplies: []govppapi.Message{&sr.SrPolicyAddReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid SID (not IP address) in non-first segment list",
			Policy: policy(sidA.Addr, 10, false, true,
				policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr),
				&srv6.Policy_SegmentList{
					Weight:   1,
					Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
				}),
			MockReplies: []govppapi.Message{&sr.SrPolicyAddReply{}, &sr.SrPolicyModReply{}},
			Verify: func(err error, catchedMsgs []govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:        "failure propagation from VPP",
			Policy:      policy(sidA.Addr, 0, true, true, policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr)),
			MockReplies: []govppapi.Message{&sr.SrPolicyAddReply{Retval: 1}},
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
			BSID:      sidA.Addr,
			MockReply: &sr.SrPolicyDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrPolicyDel{
					BsidAddr: sidA,
				}))
			},
		},
		{
			Name:      "failure propagation from VPP",
			BSID:      sidA.Addr,
			MockReply: &sr.SrPolicyDelReply{Retval: 1},
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
			policy := policy(td.BSID, 0, true, true, policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr))
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
			Policy:            policy(sidA.Addr, 10, false, true),
			PolicySegmentList: policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr),
			MockReply:         &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrPolicyMod{
					BsidAddr:  sidA.Addr,
					Operation: vpp1810.AddSRList,
					FibTable:  10,
					Sids: sr.Srv6SidList{
						Weight:  1,
						NumSids: 3,
						Sids:    []sr.Srv6Sid{{Addr: sidA.Addr}, {Addr: sidB.Addr}, {Addr: sidC.Addr}},
					},
				}))
			},
		},
		{
			Name:   "invalid SID (not IP address) in segment list",
			Policy: policy(sidA.Addr, 10, false, true),
			PolicySegmentList: &srv6.Policy_SegmentList{
				Weight:   1,
				Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
			},
			MockReply: &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid binding SID (not IP address) in policy",
			Policy: &srv6.Policy{
				Bsid:             invalidIPAddress,
				FibTableId:       10,
				SprayBehaviour:   false,
				SrhEncapsulation: true,
			},
			PolicySegmentList: policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr),
			MockReply:         &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:              "failure propagation from VPP",
			Policy:            policy(sidA.Addr, 0, true, true),
			PolicySegmentList: policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr),
			MockReply:         &sr.SrPolicyModReply{Retval: 1},
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
			Policy:            policy(sidA.Addr, 10, false, true, policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr)),
			PolicySegmentList: policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr),
			SegmentIndex:      111,
			MockReply:         &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrPolicyMod{
					BsidAddr:  sidA.Addr,
					Operation: vpp1810.DeleteSRList,
					SlIndex:   111,
					FibTable:  10,
					Sids: sr.Srv6SidList{
						Weight:  1,
						NumSids: 3,
						Sids:    []sr.Srv6Sid{{Addr: sidA.Addr}, {Addr: sidB.Addr}, {Addr: sidC.Addr}},
					},
				}))
			},
		},
		{
			Name: "invalid SID (not IP address) in segment list",
			Policy: policy(sidA.Addr, 10, false, true,
				&srv6.Policy_SegmentList{
					Weight:   1,
					Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
				}),
			PolicySegmentList: &srv6.Policy_SegmentList{
				Weight:   1,
				Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
			},
			SegmentIndex: 111,
			MockReply:    &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:              "failure propagation from VPP",
			Policy:            policy(sidA.Addr, 0, true, true, policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr)),
			PolicySegmentList: policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr),
			SegmentIndex:      111,
			MockReply:         &sr.SrPolicyModReply{Retval: 1},
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
						FibTableId:    10,
						PrefixAddress: "1::/64",
					},
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSteeringAddDel{
					IsDel:         boolToUint(removal),
					BsidAddr:      sidA.Addr,
					SrPolicyIndex: uint32(0),
					TableID:       10,
					TrafficType:   vpp1810.SteerTypeIPv6,
					PrefixAddr:    net.ParseIP("1::").To16(),
					MaskWidth:     64,
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
						FibTableId:    10,
						PrefixAddress: "1.2.3.4/24",
					},
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSteeringAddDel{
					IsDel:         boolToUint(removal),
					BsidAddr:      sidA.Addr,
					SrPolicyIndex: uint32(0),
					TableID:       10,
					TrafficType:   vpp1810.SteerTypeIPv4,
					PrefixAddr:    net.ParseIP("1.2.3.4").To16(),
					MaskWidth:     24,
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
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSteeringAddDel{
					IsDel:         boolToUint(removal),
					BsidAddr:      sidA.Addr,
					SrPolicyIndex: uint32(0),
					TrafficType:   vpp1810.SteerTypeL2,
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
						FibTableId:    10,
						PrefixAddress: "1::/64",
					},
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSteeringAddDel{
					IsDel:         boolToUint(removal),
					BsidAddr:      nil,
					SrPolicyIndex: uint32(20),
					TableID:       10,
					TrafficType:   vpp1810.SteerTypeIPv6,
					PrefixAddr:    net.ParseIP("1::").To16(),
					MaskWidth:     64,
				}))
			},
		},
		{
			Name: "missing policy reference ( " + action + " of IPv6 L3 steering)",
			Steering: &srv6.Steering{
				Traffic: &srv6.Steering_L3Traffic_{
					L3Traffic: &srv6.Steering_L3Traffic{
						FibTableId:    10,
						PrefixAddress: "1::/64",
					},
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
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
			MockReply: &sr.SrSteeringAddDelReply{},
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
						FibTableId:    10,
						PrefixAddress: invalidIPAddress,
					},
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
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
			MockReply: &sr.SrSteeringAddDelReply{},
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
						FibTableId:    10,
						PrefixAddress: "1::/64",
					},
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
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
						FibTableId:    10,
						PrefixAddress: "1::/64",
					},
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{Retval: 1},
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
	segmentListABC := policySegmentList(1, sidA.Addr, sidB.Addr, sidC.Addr)
	segmentListBBC := policySegmentList(1, sidB.Addr, sidB.Addr, sidC.Addr)
	notExistingSegmentListCCC := policySegmentList(1, sidC.Addr, sidC.Addr, sidC.Addr)

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
			Policy: policy(sidA.Addr, 10, false, true, segmentListABC, segmentListBBC),
			MockReply: &vpe.CliInbandReply{
				Reply:  []byte(correctCLIOutput),
				Length: uint32(len([]byte(correctCLIOutput))),
				Retval: 0,
			},
			ExpectedPolicyIndex:        correctPolicyIndex,
			ExpectedSegmentListIndexes: map[*srv6.Policy_SegmentList]uint32{segmentListABC: uint32(2), segmentListBBC: uint32(3)},
		},
		{
			Name:             "failure propagation from VPP",
			Policy:           policy(sidA.Addr, 10, false, true, segmentListABC, segmentListBBC),
			MockReply:        &vpe.CliInbandReply{Retval: 1},
			ExpectingFailure: true,
		},
		{
			Name:   "searching for not existing policy ",
			Policy: policy(sidC.Addr, 10, false, true, segmentListABC, segmentListBBC),
			MockReply: &vpe.CliInbandReply{
				Reply:  []byte(correctCLIOutput),
				Length: uint32(len([]byte(correctCLIOutput))),
				Retval: 0,
			},
			ExpectingFailure: true,
		},
		{
			Name:   "searching for not existing policy segment list",
			Policy: policy(sidA.Addr, 10, false, true, notExistingSegmentListCCC),
			MockReply: &vpe.CliInbandReply{
				Reply:  []byte(correctCLIOutput),
				Length: uint32(len([]byte(correctCLIOutput))),
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
			Expect(ctx.MockChannel.Msg).To(Equal(&vpe.CliInband{
				Cmd:    []byte("sh sr policies"),
				Length: uint32(len("sh sr policies")),
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

func setup(t *testing.T) (*vppcallmock.TestCtx, vppcalls.SRv6VppAPI) {
	ctx := vppcallmock.SetupTestCtx(t)
	log := logrus.NewLogger("test")
	swIfIndex := ifaceidx.NewIfaceIndex(log, "test")
	swIfIndex.Put(ifaceA, &ifaceidx.IfaceMetadata{SwIfIndex: swIndexA})
	vppCalls := vpp1810.NewSRv6VppHandler(ctx.MockChannel, swIfIndex, log)
	return ctx, vppCalls
}

func teardown(ctx *vppcallmock.TestCtx) {
	ctx.TeardownTestCtx()
}

func sid(str string) *sr.Srv6Sid {
	bsid, err := parseIPv6(str)
	if err != nil {
		panic(fmt.Sprintf("can't parse %q into SRv6 BSID (IPv6 address)", str))
	}
	return &sr.Srv6Sid{
		Addr: bsid,
	}
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

func policy(bsid srv6.SID, fibtableID uint32, sprayBehaviour bool, srhEncapsulation bool, segmentLists ...*srv6.Policy_SegmentList) *srv6.Policy {
	return &srv6.Policy{
		Bsid:             bsid.String(),
		FibTableId:       fibtableID,
		SprayBehaviour:   sprayBehaviour,
		SrhEncapsulation: srhEncapsulation,
		SegmentLists:     segmentLists,
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

func sidToStr(sid sr.Srv6Sid) string {
	return srv6.SID(sid.Addr).String()
}

// toIFaceByte converts value to byte representation as returned by VPP binary api in case of interface info (string bytes + 1 zero byte)
func toIFaceByte(val string) []byte {
	return append([]byte(val), 0x00)
}
