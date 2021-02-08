//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package vpp2101_test

import (
	"testing"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/ip_types"

	. "github.com/onsi/gomega"

	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/interface"
	vpp_vxlan "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/vxlan"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestAddVxlanTunnel(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	swIfIdx, err := ifHandler.AddVxLanTunnel("ifName", 0, 2, &ifs.VxlanLink{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	})
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_vxlan.VxlanAddDelTunnel)
		if ok {
			Expect(vppMsg.SrcAddress.Un.GetIP4()).To(BeEquivalentTo(ip_types.IP4Address{10, 0, 0, 1}))
			Expect(vppMsg.DstAddress.Un.GetIP4()).To(BeEquivalentTo(ip_types.IP4Address{20, 0, 0, 1}))
			Expect(vppMsg.IsAdd).To(BeEquivalentTo(true))
			Expect(vppMsg.EncapVrfID).To(BeEquivalentTo(0))
			Expect(vppMsg.McastSwIfIndex).To(BeEquivalentTo(2))
			Expect(vppMsg.Vni).To(BeEquivalentTo(1))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddVxlanTunnelWithVrf(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	// VxLAN resolution
	ctx.MockVpp.MockReply(&vpp_vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	swIfIdx, err := ifHandler.AddVxLanTunnel("ifName", 1, 1, &ifs.VxlanLink{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	})
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_vxlan.VxlanAddDelTunnel)
		if ok {
			Expect(vppMsg.SrcAddress.Un.GetIP4()).To(BeEquivalentTo(ip_types.IP4Address{10, 0, 0, 1}))
			Expect(vppMsg.DstAddress.Un.GetIP4()).To(BeEquivalentTo(ip_types.IP4Address{20, 0, 0, 1}))
			Expect(vppMsg.IsAdd).To(BeEquivalentTo(true))
			Expect(vppMsg.EncapVrfID).To(BeEquivalentTo(1))
			Expect(vppMsg.McastSwIfIndex).To(BeEquivalentTo(1))
			Expect(vppMsg.Vni).To(BeEquivalentTo(1))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddVxlanTunnelIPv6(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	swIfIdx, err := ifHandler.AddVxLanTunnel("ifName", 0, 0, &ifs.VxlanLink{
		SrcAddress: "2001:db8:0:1:1:1:1:1",
		DstAddress: "2002:db8:0:1:1:1:1:1",
		Vni:        1,
	})
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_vxlan.VxlanAddDelTunnel)
		if ok {
			Expect(vppMsg.SrcAddress).To(BeEquivalentTo(ipToAddr("2001:db8:0:1:1:1:1:1")))
			Expect(vppMsg.DstAddress).To(BeEquivalentTo(ipToAddr("2002:db8:0:1:1:1:1:1")))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddVxlanTunnelIPMismatch(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddVxLanTunnel("ifName", 0, 0, &ifs.VxlanLink{
		SrcAddress: "10.0.0.1",
		DstAddress: "2001:db8:0:1:1:1:1:1",
		Vni:        1,
	})
	Expect(err).ToNot(BeNil())
}

func TestAddVxlanTunnelInvalidIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddVxLanTunnel("ifName", 0, 0, &ifs.VxlanLink{
		SrcAddress: "invalid-ip",
		DstAddress: "2001:db8:0:1:1:1:1:1",
		Vni:        1,
	})
	Expect(err).ToNot(BeNil())
}

func TestAddVxlanTunnelError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vxlan.VxlanAddDelTunnel{})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddVxLanTunnel("ifName", 0, 0, &ifs.VxlanLink{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.2",
		Vni:        1,
	})
	Expect(err).ToNot(BeNil())
}

func TestAddVxlanTunnelRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vxlan.VxlanAddDelTunnelReply{
		Retval: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddVxLanTunnel("ifName", 0, 0, &ifs.VxlanLink{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.2",
		Vni:        1,
	})
	Expect(err).ToNot(BeNil())
}

func TestDeleteVxlanTunnel(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DeleteVxLanTunnel("ifName", 1, 0, &ifs.VxlanLink{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	})
	Expect(err).To(BeNil())
}

func TestDeleteVxlanTunnelError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vxlan.VxlanAddDelTunnel{})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DeleteVxLanTunnel("ifName", 1, 0, &ifs.VxlanLink{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	})
	Expect(err).ToNot(BeNil())
}

func TestDeleteVxlanTunnelRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vxlan.VxlanAddDelTunnelReply{
		Retval: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DeleteVxLanTunnel("ifName", 1, 0, &ifs.VxlanLink{
		SrcAddress: "10.0.0.1",
		DstAddress: "20.0.0.1",
		Vni:        1,
	})
	Expect(err).ToNot(BeNil())
}
