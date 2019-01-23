// Copyright (c) 2018 Bell Canada, Pantheon Technologies and/or its affiliates.
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

package vppcalls_test

import (
	"fmt"
	"net"
	"testing"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/sr"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/model/srv6"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
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

var swIfIndex = ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "sw_if_indexes", ifaceidx.IndexMetadata))

func init() {
	swIfIndex.RegisterName(ifaceA, swIndexA, nil)
}

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
				FibTableId: 10,
				BaseEndFunction: &srv6.LocalSID_End{
					Psp: true,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:    0,
				Localsid: sidA,
				Behavior: vppcalls.BehaviorEnd,
				FibTable: 10,
				EndPsp:   1,
			},
		},
		{
			Name: "addition with endX behaviour (ipv6 next hop address)",
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_X: &srv6.LocalSID_EndX{
					Psp:               true,
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceA,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vppcalls.BehaviorX,
				FibTable:  10,
				EndPsp:    1,
				SwIfIndex: swIndexA,
				NhAddr6:   nextHop,
			},
		},
		{
			Name: "addition with endX behaviour (ipv4 next hop address)",
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_X: &srv6.LocalSID_EndX{
					Psp:               true,
					NextHop:           nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vppcalls.BehaviorX,
				FibTable:  10,
				EndPsp:    1,
				SwIfIndex: swIndexA,
				NhAddr4:   nextHopIPv4,
			},
		},
		{
			Name: "addition with endT behaviour",
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_T: &srv6.LocalSID_EndT{
					Psp: true,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:    0,
				Localsid: sidA,
				Behavior: vppcalls.BehaviorT,
				FibTable: 10,
				EndPsp:   1,
			},
		},
		{
			Name: "addition with endDX2 behaviour (ipv6 next hop address)",
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX2: &srv6.LocalSID_EndDX2{
					VlanTag:           1,
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceA,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vppcalls.BehaviorDX2,
				FibTable:  10,
				EndPsp:    0,
				VlanIndex: 1,
				SwIfIndex: swIndexA,
				NhAddr6:   nextHop,
			},
		},
		{
			Name: "addition with endDX2 behaviour (ipv4 next hop address)",
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX2: &srv6.LocalSID_EndDX2{
					VlanTag:           1,
					NextHop:           nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vppcalls.BehaviorDX2,
				FibTable:  10,
				EndPsp:    0,
				VlanIndex: 1,
				SwIfIndex: swIndexA,
				NhAddr4:   nextHopIPv4,
			},
		},
		{
			Name: "addition with endDX4 behaviour",
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX4: &srv6.LocalSID_EndDX4{
					NextHop:           nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vppcalls.BehaviorDX4,
				FibTable:  10,
				EndPsp:    0,
				SwIfIndex: swIndexA,
				NhAddr4:   nextHopIPv4,
			},
		},
		{
			Name: "addition with endDX6 behaviour",
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX6: &srv6.LocalSID_EndDX6{
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceA,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:     0,
				Localsid:  sidA,
				Behavior:  vppcalls.BehaviorDX6,
				FibTable:  10,
				EndPsp:    0,
				SwIfIndex: swIndexA,
				NhAddr6:   nextHop,
			},
		},
		// endDT4 and endDT6 are not fully modelled yet -> testing only current implementation
		{
			Name: "addition with endDT4 behaviour",
			Input: &srv6.LocalSID{
				FibTableId:      10,
				EndFunction_DT4: &srv6.LocalSID_EndDT4{},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:    0,
				Localsid: sidA,
				Behavior: vppcalls.BehaviorDT4,
				FibTable: 10,
				EndPsp:   0,
			},
		},
		{
			Name: "addition with endDT6 behaviour",
			Input: &srv6.LocalSID{
				FibTableId:      10,
				EndFunction_DT6: &srv6.LocalSID_EndDT6{},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:    0,
				Localsid: sidA,
				Behavior: vppcalls.BehaviorDT6,
				FibTable: 10,
				EndPsp:   0,
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
				FibTableId: 10,
				EndFunction_AD: &srv6.LocalSID_EndAD{
					ServiceAddress:    nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
					IncomingInterface: ifaceB,
				},
			},
			Expected: &vpe.CliInband{
				Cmd:    []byte(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), memif1, memif2)),
				Length: uint32(len(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidToStr(sidA), nextHopIPv4.String(), memif1, memif2))),
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
				FibTableId: 10,
				EndFunction_AD: &srv6.LocalSID_EndAD{
					ServiceAddress:    nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
					IncomingInterface: ifaceB,
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
				FibTableId: 10,
				EndFunction_AD: &srv6.LocalSID_EndAD{
					ServiceAddress:    nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
					IncomingInterface: ifaceB,
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
				FibTableId: 10,
				EndFunction_AD: &srv6.LocalSID_EndAD{
					ServiceAddress:    nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
					IncomingInterface: ifaceB,
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
				FibTableId: 10,
				EndFunction_AD: &srv6.LocalSID_EndAD{
					ServiceAddress:    nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
					IncomingInterface: "unknown0", // interface name is taken from vpp internal name
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
				FibTableId: 0,
			},
		},
		{
			Name:          "failure propagation from VPP (doing main VPP call)",
			FailInVPP:     true,
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableId: 0,
				BaseEndFunction: &srv6.LocalSID_End{
					Psp: true,
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
				FibTableId: 10,
				EndFunction_AD: &srv6.LocalSID_EndAD{
					ServiceAddress:    nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
					IncomingInterface: ifaceB,
				},
			},
		},
		{
			Name:          "failure propagation from VPP Dump call",
			FailInVPPDump: true,
			ExpectFailure: true,
			cliMode:       true,
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_AD: &srv6.LocalSID_EndAD{
					ServiceAddress:    nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
					IncomingInterface: ifaceB,
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
				FibTableId: 10,
				EndFunction_AD: &srv6.LocalSID_EndAD{
					ServiceAddress:    nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
					IncomingInterface: ifaceB,
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
				FibTableId: 10,
				EndFunction_AD: &srv6.LocalSID_EndAD{
					ServiceAddress:    nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
					IncomingInterface: ifaceB,
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endX behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_X: &srv6.LocalSID_EndX{
					Psp:               true,
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceBOutOfidxs,
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endX behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_X: &srv6.LocalSID_EndX{
					Psp:               true,
					NextHop:           invalidIPAddress,
					OutgoingInterface: ifaceA,
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endDX2 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX2: &srv6.LocalSID_EndDX2{
					VlanTag:           1,
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceBOutOfidxs,
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endDX2 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX2: &srv6.LocalSID_EndDX2{
					VlanTag:           1,
					NextHop:           invalidIPAddress,
					OutgoingInterface: ifaceA,
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endDX4 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX4: &srv6.LocalSID_EndDX4{
					NextHop:           nextHopIPv4.String(),
					OutgoingInterface: ifaceBOutOfidxs,
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endDX4 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX4: &srv6.LocalSID_EndDX4{
					NextHop:           invalidIPAddress,
					OutgoingInterface: ifaceA,
				},
			},
		},
		{
			Name:          "rejection of IPv6 addresses (addition with endDX4 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX4: &srv6.LocalSID_EndDX4{
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceA,
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endDX6 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX6: &srv6.LocalSID_EndDX6{
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceBOutOfidxs,
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endDX6 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableId: 10,
				EndFunction_DX6: &srv6.LocalSID_EndDX6{
					NextHop:           invalidIPAddress,
					OutgoingInterface: ifaceA,
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
			err := vppCalls.AddLocalSid(sidA.Addr, td.Input, swIfIndex)
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
				FibTableId: 10,
				BaseEndFunction: &srv6.LocalSID_End{
					Psp: true,
				},
			}
			vppCalls.AddLocalSid(td.Sid, localsid, swIfIndex)
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
		Name          string
		Fail          bool
		BSID          net.IP
		Policy        *srv6.Policy
		PolicySegment *srv6.PolicySegment
		MockReply     govppapi.Message
		Verify        func(error, govppapi.Message)
	}{
		{
			Name:          "simple SetAddPolicy",
			BSID:          sidA.Addr,
			Policy:        policy(10, false, true),
			PolicySegment: policySegment(1, sidA.Addr, sidB.Addr, sidC.Addr),
			MockReply:     &sr.SrPolicyAddReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrPolicyAdd{
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
			Name:   "invalid SID (not IP address) in segment list",
			BSID:   sidA.Addr,
			Policy: policy(10, false, true),
			PolicySegment: &srv6.PolicySegment{
				Weight:   1,
				Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
			},
			MockReply: &sr.SrPolicyAddReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:          "failure propagation from VPP",
			BSID:          sidA.Addr,
			Policy:        policy(0, true, true),
			PolicySegment: policySegment(1, sidA.Addr, sidB.Addr, sidC.Addr),
			MockReply:     &sr.SrPolicyAddReply{Retval: 1},
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
			err := vppCalls.AddPolicy(td.BSID, td.Policy, td.PolicySegment)
			td.Verify(err, ctx.MockChannel.Msg)
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
			policy := policy(0, true, true)
			segment := policySegment(1, sidA.Addr, sidB.Addr, sidC.Addr)
			vppCalls.AddPolicy(td.BSID, policy, segment)
			ctx.MockVpp.MockReply(td.MockReply)
			// make the call and verify
			err := vppCalls.DeletePolicy(td.BSID)
			td.Verify(err, ctx.MockChannel.Msg)
		})
	}
}

// TestAddPolicySegment tests all cases for method AddPolicySegment
func TestAddPolicySegment(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name          string
		BSID          net.IP
		Policy        *srv6.Policy
		PolicySegment *srv6.PolicySegment
		MockReply     govppapi.Message
		Verify        func(error, govppapi.Message)
	}{
		{
			Name:          "simple addition of policy segment",
			BSID:          sidA.Addr,
			Policy:        policy(10, false, true),
			PolicySegment: policySegment(1, sidA.Addr, sidB.Addr, sidC.Addr),
			MockReply:     &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrPolicyMod{
					BsidAddr:  sidA.Addr,
					Operation: vppcalls.AddSRList,
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
			BSID:   sidA.Addr,
			Policy: policy(10, false, true),
			PolicySegment: &srv6.PolicySegment{
				Weight:   1,
				Segments: []string{sidToStr(sidA), invalidIPAddress, sidToStr(sidC)},
			},
			MockReply: &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:          "failure propagation from VPP",
			BSID:          sidA.Addr,
			Policy:        policy(0, true, true),
			PolicySegment: policySegment(1, sidA.Addr, sidB.Addr, sidC.Addr),
			MockReply:     &sr.SrPolicyModReply{Retval: 1},
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
			err := vppCalls.AddPolicySegment(td.BSID, td.Policy, td.PolicySegment)
			td.Verify(err, ctx.MockChannel.Msg)
		})
	}
}

// TestDeletePolicySegment tests all cases for method DeletePolicySegment
func TestDeletePolicySegment(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name          string
		BSID          net.IP
		Policy        *srv6.Policy
		PolicySegment *srv6.PolicySegment
		SegmentIndex  uint32
		MockReply     govppapi.Message
		Verify        func(error, govppapi.Message)
	}{
		{
			Name:          "simple deletion of policy segment",
			BSID:          sidA.Addr,
			Policy:        policy(10, false, true),
			PolicySegment: policySegment(1, sidA.Addr, sidB.Addr, sidC.Addr),
			SegmentIndex:  111,
			MockReply:     &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrPolicyMod{
					BsidAddr:  sidA.Addr,
					Operation: vppcalls.DeleteSRList,
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
			Name:   "invalid SID (not IP address) in segment list",
			BSID:   sidA.Addr,
			Policy: policy(10, false, true),
			PolicySegment: &srv6.PolicySegment{
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
			Name:          "failure propagation from VPP",
			BSID:          sidA.Addr,
			Policy:        policy(0, true, true),
			PolicySegment: policySegment(1, sidA.Addr, sidB.Addr, sidC.Addr),
			SegmentIndex:  111,
			MockReply:     &sr.SrPolicyModReply{Retval: 1},
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
			err := vppCalls.DeletePolicySegment(td.BSID, td.Policy, td.PolicySegment, td.SegmentIndex)
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
				PolicyBsid: sidToStr(sidA),
				L3Traffic: &srv6.Steering_L3Traffic{
					FibTableId:    10,
					PrefixAddress: "1::/64",
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSteeringAddDel{
					IsDel:       boolToUint(removal),
					BsidAddr:    sidA.Addr,
					TableID:     10,
					TrafficType: vppcalls.SteerTypeIPv6,
					PrefixAddr:  net.ParseIP("1::").To16(),
					MaskWidth:   64,
				}))
			},
		},
		{
			Name: action + " of IPv4 L3 steering",
			Steering: &srv6.Steering{
				PolicyBsid: sidToStr(sidA),
				L3Traffic: &srv6.Steering_L3Traffic{
					FibTableId:    10,
					PrefixAddress: "1.2.3.4/24",
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSteeringAddDel{
					IsDel:       boolToUint(removal),
					BsidAddr:    sidA.Addr,
					TableID:     10,
					TrafficType: vppcalls.SteerTypeIPv4,
					PrefixAddr:  net.ParseIP("1.2.3.4").To16(),
					MaskWidth:   24,
				}))
			},
		},
		{
			Name: action + " of L2 steering",
			Steering: &srv6.Steering{
				PolicyBsid: sidToStr(sidA),
				L2Traffic: &srv6.Steering_L2Traffic{
					InterfaceName: ifaceA,
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSteeringAddDel{
					IsDel:       boolToUint(removal),
					BsidAddr:    sidA.Addr,
					TrafficType: vppcalls.SteerTypeL2,
					SwIfIndex:   swIndexA,
				}))
			},
		},
		{
			Name: "invalid prefix (" + action + " of IPv4 L3 steering)",
			Steering: &srv6.Steering{
				PolicyBsid: sidToStr(sidA),
				L3Traffic: &srv6.Steering_L3Traffic{
					FibTableId:    10,
					PrefixAddress: invalidIPAddress,
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
				PolicyBsid: sidToStr(sidA),
				L2Traffic: &srv6.Steering_L2Traffic{
					InterfaceName: ifaceBOutOfidxs,
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
				PolicyBsid: invalidIPAddress,
				L3Traffic: &srv6.Steering_L3Traffic{
					FibTableId:    10,
					PrefixAddress: "1::/64",
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
				PolicyBsid: sidToStr(sidA),
				L3Traffic: &srv6.Steering_L3Traffic{
					FibTableId:    10,
					PrefixAddress: "1::/64",
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
				err = vppCalls.RemoveSteering(td.Steering, swIfIndex)
			} else {
				err = vppCalls.AddSteering(td.Steering, swIfIndex)
			}
			td.Verify(err, ctx.MockChannel.Msg)
		})
	}
}

func setup(t *testing.T) (*vppcallmock.TestCtx, vppcalls.SRv6VppAPI) {
	ctx := vppcallmock.SetupTestCtx(t)
	vppCalls := vppcalls.NewSRv6VppHandler(ctx.MockChannel, logrus.DefaultLogger())
	return ctx, vppCalls
}

func teardown(ctx *vppcallmock.TestCtx) {
	ctx.TeardownTestCtx()
}

func sid(str string) *sr.Srv6Sid {
	bsid, err := srplugin.ParseIPv6(str)
	if err != nil {
		panic(fmt.Sprintf("can't parse %q into SRv6 BSID (IPv6 address)", str))
	}
	return &sr.Srv6Sid{
		Addr: bsid,
	}
}

func policy(fibtableID uint32, sprayBehaviour bool, srhEncapsulation bool) *srv6.Policy {
	return &srv6.Policy{
		FibTableId:       fibtableID,
		SprayBehaviour:   sprayBehaviour,
		SrhEncapsulation: srhEncapsulation,
	}
}

func policySegment(weight uint32, sids ...srv6.SID) *srv6.PolicySegment {
	segments := make([]string, len(sids))
	for i, sid := range sids {
		segments[i] = sid.String()
	}

	return &srv6.PolicySegment{
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
