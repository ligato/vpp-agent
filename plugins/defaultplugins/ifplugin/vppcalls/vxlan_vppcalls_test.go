// Copyright (c) 2018 Cisco and/or its affiliates.
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

package vppcalls

import (
	"net"
	"testing"

	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/vpe"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/vxlan"
	ifModel "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

func TestAddVxlanTunnel(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	swIfIdx, err := AddVxlanTunnel("ifName", &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	}, 0, ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vxlan.VxlanAddDelTunnel)
		if ok {
			Expect(vppMsg.SrcAddress).To(BeEquivalentTo(net.ParseIP("10.0.0.1").To4()))
			Expect(vppMsg.DstAddress).To(BeEquivalentTo(net.ParseIP("20.0.0.1").To4()))
			Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
			Expect(vppMsg.EncapVrfID).To(BeEquivalentTo(0))
			Expect(vppMsg.Vni).To(BeEquivalentTo(1))
			Expect(vppMsg.IsIpv6).To(BeEquivalentTo(0))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddVxlanTunnelWithVrf(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// VRF resolution
	ctx.MockVpp.MockReply(&ip.IPFibDetails{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})
	// VxLAN resolution
	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	swIfIdx, err := AddVxlanTunnel("ifName", &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	}, 1, ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vxlan.VxlanAddDelTunnel)
		if ok {
			Expect(vppMsg.SrcAddress).To(BeEquivalentTo(net.ParseIP("10.0.0.1").To4()))
			Expect(vppMsg.DstAddress).To(BeEquivalentTo(net.ParseIP("20.0.0.1").To4()))
			Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
			Expect(vppMsg.EncapVrfID).To(BeEquivalentTo(1))
			Expect(vppMsg.Vni).To(BeEquivalentTo(1))
			Expect(vppMsg.IsIpv6).To(BeEquivalentTo(0))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddVxlanTunnelIPv6(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	swIfIdx, err := AddVxlanTunnel("ifName", &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "2001:db8:0:1:1:1:1:1",
		DstAddress: "2002:db8:0:1:1:1:1:1",
		Vni:        1,
	}, 0, ctx.MockChannel, nil)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vxlan.VxlanAddDelTunnel)
		if ok {
			Expect(vppMsg.SrcAddress).To(BeEquivalentTo(net.ParseIP("2001:db8:0:1:1:1:1:1").To16()))
			Expect(vppMsg.DstAddress).To(BeEquivalentTo(net.ParseIP("2002:db8:0:1:1:1:1:1").To16()))
			Expect(vppMsg.IsIpv6).To(BeEquivalentTo(1))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddVxlanTunnelIPMismatch(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	_, err := AddVxlanTunnel("ifName", &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "10.0.0.1",
		DstAddress: "2001:db8:0:1:1:1:1:1",
		Vni:        1,
	}, 0, ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}

func TestAddVxlanTunnelInvalidIP(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	_, err := AddVxlanTunnel("ifName", &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "invalid-ip",
		DstAddress: "2001:db8:0:1:1:1:1:1",
		Vni:        1,
	}, 0, ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}

func TestAddVxlanTunnelError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnel{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	_, err := AddVxlanTunnel("ifName", &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.2",
		Vni:        1,
	}, 0, ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}

func TestAddVxlanTunnelWithVrfError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// VRF resolution
	ctx.MockVpp.MockReply(&ip.IPFibDump{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})
	// VxLAN resolution
	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	_, err := AddVxlanTunnel("ifName", &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	}, 1, ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}

func TestAddVxlanTunnelRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnelReply{
		Retval: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	_, err := AddVxlanTunnel("ifName", &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.2",
		Vni:        1,
	}, 0, ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}

func TestDeleteVxlanTunnel(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	err := DeleteVxlanTunnel("ifName", 1, &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	}, ctx.MockChannel, nil)
	Expect(err).To(BeNil())
}

func TestDeleteVxlanTunnelError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnel{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	err := DeleteVxlanTunnel("ifName", 1, &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	}, ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}

func TestDeleteVxlanTunnelRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnelReply{
		Retval: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	err := DeleteVxlanTunnel("ifName", 1, &ifModel.Interfaces_Interface_Vxlan{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	}, ctx.MockChannel, nil)
	Expect(err).ToNot(BeNil())
}
