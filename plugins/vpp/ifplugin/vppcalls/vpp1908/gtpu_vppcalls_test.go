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

package vpp1908_test

import (
	"net"
	"testing"

	ifs "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	vpp_gtpu "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/gtpu"
	vpp_ifs "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/interfaces"
	. "github.com/onsi/gomega"
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
			Expect(vppMsg.SrcAddress).To(BeEquivalentTo(net.ParseIP("10.0.0.1").To4()))
			Expect(vppMsg.DstAddress).To(BeEquivalentTo(net.ParseIP("20.0.0.1").To4()))
			Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
			Expect(vppMsg.EncapVrfID).To(BeEquivalentTo(10))
			Expect(vppMsg.McastSwIfIndex).To(BeEquivalentTo(2))
			Expect(vppMsg.Teid).To(BeEquivalentTo(100))
			Expect(vppMsg.IsIPv6).To(BeEquivalentTo(0))
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
			Expect(vppMsg.SrcAddress).To(BeEquivalentTo(net.ParseIP("2001:db8:0:1:1:1:1:1").To16()))
			Expect(vppMsg.DstAddress).To(BeEquivalentTo(net.ParseIP("2002:db8:0:1:1:1:1:1").To16()))
			Expect(vppMsg.IsIPv6).To(BeEquivalentTo(1))
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
