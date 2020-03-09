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

package vpp2001_test

import (
	"testing"

	. "github.com/onsi/gomega"
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	vpp_ipip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ipip"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestAddIpipTunnel(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipip.IpipAddTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	swIfIdx, err := ifHandler.AddIpipTunnel("ipiptun1", 0, &ifs.IPIPLink{
		SrcAddr: "10.0.0.1",
		DstAddr: "20.0.0.1",
	})
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_ipip.IpipAddTunnel)
		if ok {
			Expect(vppMsg.Tunnel.Src).To(Equal(ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{10, 0, 0, 1}),
			}))
			Expect(vppMsg.Tunnel.Dst).To(Equal(ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{20, 0, 0, 1}),
			}))
			Expect(vppMsg.Tunnel.TableID).To(BeEquivalentTo(0))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddIpipTunnelWithVrf(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipip.IpipAddTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	swIfIdx, err := ifHandler.AddIpipTunnel("ipiptun1", 1, &ifs.IPIPLink{
		SrcAddr: "10.0.0.1",
		DstAddr: "20.0.0.1",
	})
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_ipip.IpipAddTunnel)
		if ok {
			Expect(vppMsg.Tunnel.Src).To(Equal(ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{10, 0, 0, 1}),
			}))
			Expect(vppMsg.Tunnel.Dst).To(Equal(ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{20, 0, 0, 1}),
			}))
			Expect(vppMsg.Tunnel.TableID).To(BeEquivalentTo(1))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddIpipTunnelIPv6(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipip.IpipAddTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	swIfIdx, err := ifHandler.AddIpipTunnel("ipiptun1", 0, &ifs.IPIPLink{
		SrcAddr: "2001:db8:0:1:1:1:1:1",
		DstAddr: "2002:db8:0:1:1:1:1:1",
	})
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_ipip.IpipAddTunnel)
		if ok {
			Expect(vppMsg.Tunnel.Src).To(Equal(ip_types.Address{
				Af: ip_types.ADDRESS_IP6,
				Un: ip_types.AddressUnionIP6(ip_types.IP6Address{
					0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0x01, 0, 0x01, 0, 0x01, 0, 0x01, 0, 0x01,
				}),
			}))
			Expect(vppMsg.Tunnel.Dst).To(Equal(ip_types.Address{
				Af: ip_types.ADDRESS_IP6,
				Un: ip_types.AddressUnionIP6(ip_types.IP6Address{
					0x20, 0x02, 0x0d, 0xb8, 0, 0, 0, 0x01, 0, 0x01, 0, 0x01, 0, 0x01, 0, 0x01,
				}),
			}))
			Expect(vppMsg.Tunnel.TableID).To(BeEquivalentTo(0))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddIpipTunnelIPMismatch(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipip.IpipAddTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddIpipTunnel("ipiptun1", 0, &ifs.IPIPLink{
		SrcAddr: "10.0.0.1",
		DstAddr: "2001:db8:0:1:1:1:1:1",
	})
	Expect(err).ToNot(BeNil())
}

func TestAddIpipTunnelInvalidIP(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipip.IpipAddTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddIpipTunnel("ipiptun1", 0, &ifs.IPIPLink{
		SrcAddr: "invalid-ip",
		DstAddr: "2001:db8:0:1:1:1:1:1",
	})
	Expect(err).ToNot(BeNil())
}

func TestAddIpipTunnelNoIPs(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipip.IpipAddTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddIpipTunnel("ipiptun1", 0, &ifs.IPIPLink{})
	Expect(err).ToNot(BeNil())
}

func TestAddIpipTunnelRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipip.IpipAddTunnelReply{
		Retval: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddIpipTunnel("ipiptun1", 0, &ifs.IPIPLink{
		SrcAddr: "10.0.0.1",
		DstAddr: "20.0.0.2",
	})
	Expect(err).ToNot(BeNil())
}

func TestDeleteIpipTunnel(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipip.IpipDelTunnelReply{})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DelIpipTunnel("ipiptun1", 1)
	Expect(err).To(BeNil())
}

func TestDeleteIpipTunnelRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipip.IpipDelTunnelReply{
		Retval: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DelIpipTunnel("ipiptun1", 1)
	Expect(err).ToNot(BeNil())
}
