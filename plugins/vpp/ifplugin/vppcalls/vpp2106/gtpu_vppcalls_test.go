//  Copyright (c) 2019 EMnify
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

package vpp2106_test

import (
	"testing"

	. "github.com/onsi/gomega"

	vpp_gtpu "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/gtpu"
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestAddGtpuTunnel(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	swIfIdx, err := ifHandler.AddGtpuTunnel("ifName", &ifs.GtpuLink{
		SrcAddr:    "10.0.0.1",
		DstAddr:    "20.0.0.1",
		EncapVrfId: 10,
		Teid:       100,
	}, 2)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_gtpu.GtpuAddDelTunnel)
		if ok {
			Expect(vppMsg.SrcAddress).To(Equal(ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{10, 0, 0, 1}),
			}))
			Expect(vppMsg.DstAddress).To(Equal(ip_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ip_types.AddressUnionIP4(ip_types.IP4Address{20, 0, 0, 1}),
			}))
			Expect(vppMsg.IsAdd).To(BeTrue())
			Expect(vppMsg.EncapVrfID).To(BeEquivalentTo(10))
			Expect(vppMsg.McastSwIfIndex).To(BeEquivalentTo(2))
			Expect(vppMsg.Teid).To(BeEquivalentTo(100))
			Expect(vppMsg.SrcAddress.Af).To(Equal(ip_types.ADDRESS_IP4))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddGtpuTunnelIPv6(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	swIfIdx, err := ifHandler.AddGtpuTunnel("ifName", &ifs.GtpuLink{
		SrcAddr:    "2001:db8:0:1:1:1:1:1",
		DstAddr:    "2002:db8:0:1:1:1:1:1",
		EncapVrfId: 10,
		Teid:       200,
	}, 0xFFFFFFFF)
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_gtpu.GtpuAddDelTunnel)
		if ok {
			Expect(vppMsg.SrcAddress).To(Equal(ip_types.Address{
				Af: ip_types.ADDRESS_IP6,
				Un: ip_types.AddressUnionIP6(ip_types.IP6Address{
					0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0x01, 0, 0x01, 0, 0x01, 0, 0x01, 0, 0x01,
				}),
			}))
			Expect(vppMsg.DstAddress).To(Equal(ip_types.Address{
				Af: ip_types.ADDRESS_IP6,
				Un: ip_types.AddressUnionIP6(ip_types.IP6Address{
					0x20, 0x02, 0x0d, 0xb8, 0, 0, 0, 0x01, 0, 0x01, 0, 0x01, 0, 0x01, 0, 0x01,
				}),
			}))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddGtpuTunnelIPMismatch(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddGtpuTunnel("ifName", &ifs.GtpuLink{
		SrcAddr:    "10.0.0.1",
		DstAddr:    "2001:db8:0:1:1:1:1:1",
		EncapVrfId: 0,
		Teid:       100,
	}, 0xFFFFFFFF)
	Expect(err).ToNot(BeNil())
}

func TestAddGtpuTunnelInvalidIPv4(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddGtpuTunnel("ifName", &ifs.GtpuLink{
		SrcAddr:    "10.0.0",
		DstAddr:    "20.0.0.1",
		EncapVrfId: 0,
		Teid:       100,
	}, 0xFFFFFFFF)
	Expect(err).ToNot(BeNil())
}

func TestAddGtpuTunnelInvalidIPv6(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddGtpuTunnel("ifName", &ifs.GtpuLink{
		SrcAddr:    "2001:db8:0:1:1:1:1:1",
		DstAddr:    "2002:db8:0:1:1:1:1:1:1",
		EncapVrfId: 0,
		Teid:       100,
	}, 0xFFFFFFFF)
	Expect(err).ToNot(BeNil())
}

func TestAddGtpuTunnelError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnel{})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddGtpuTunnel("ifName", &ifs.GtpuLink{
		SrcAddr:    "10.0.0.1",
		DstAddr:    "20.0.0.2",
		EncapVrfId: 0,
		Teid:       100,
	}, 0xFFFFFFFF)
	Expect(err).ToNot(BeNil())
}

func TestAddGtpuTunnelRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnelReply{
		Retval: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddGtpuTunnel("ifName", &ifs.GtpuLink{
		SrcAddr:    "10.0.0.1",
		DstAddr:    "20.0.0.2",
		EncapVrfId: 0,
		Teid:       100,
	}, 0xFFFFFFFF)
	Expect(err).ToNot(BeNil())
}

func TestDelGtpuTunnel(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DelGtpuTunnel("ifName", &ifs.GtpuLink{
		SrcAddr:    "10.0.0.1",
		DstAddr:    "20.0.0.2",
		EncapVrfId: 0,
		Teid:       100,
	})
	Expect(err).To(BeNil())
}

func TestDelGtpuTunnelError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnel{})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DelGtpuTunnel("ifName", &ifs.GtpuLink{
		SrcAddr:    "10.0.0.1",
		DstAddr:    "20.0.0.2",
		EncapVrfId: 0,
		Teid:       100,
	})
	Expect(err).ToNot(BeNil())
}

func TestDelGtpuTunnelRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnelReply{
		Retval: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DelGtpuTunnel("ifName", &ifs.GtpuLink{
		SrcAddr:    "10.0.0.1",
		DstAddr:    "20.0.0.1",
		EncapVrfId: 0,
		Teid:       100,
	})
	Expect(err).ToNot(BeNil())
}

func TestAddNilGtpuTunnel(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnel{})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	_, err := ifHandler.AddGtpuTunnel("ifName", nil, 0xFFFFFFFF)
	Expect(err).ToNot(BeNil())
}

func TestDelNilGtpuTunnel(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_gtpu.GtpuAddDelTunnel{})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DelGtpuTunnel("ifName", nil)
	Expect(err).ToNot(BeNil())
}
