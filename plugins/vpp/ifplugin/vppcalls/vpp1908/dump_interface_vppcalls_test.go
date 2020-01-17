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

package vpp1908_test

import (
	"net"
	"testing"

	govppapi "git.fd.io/govpp.git/api"

	. "github.com/onsi/gomega"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/gtpu"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/memif"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/tapv2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vxlan"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	interfaces2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

// Test dump of interfaces with vxlan type
func TestDumpInterfacesVxLan(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ipv61Parse := net.ParseIP("dead:beef:feed:face:cafe:babe:baad:c0de").To16()
	ipv62Parse := net.ParseIP("d3ad:beef:feed:face:cafe:babe:baad:c0de").To16()

	ctx.MockReplies([]*vppmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: "vxlan1",
			},
		},
		{
			Name:    (&interfaces.SwInterfaceGetTable{}).GetMessageName(),
			Ping:    false,
			Message: &interfaces.SwInterfaceGetTableReply{},
		},
		{
			Name:    (&ip.IPAddressDump{}).GetMessageName(),
			Ping:    true,
			Message: &ip.IPAddressDetails{},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&memif.MemifDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&tapv2.SwInterfaceTapV2Dump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&vxlan.VxlanTunnelDump{}).GetMessageName(),
			Ping: true,
			Message: &vxlan.VxlanTunnelDetails{
				IsIPv6:     1,
				SwIfIndex:  0,
				SrcAddress: ipv61Parse,
				DstAddress: ipv62Parse,
			},
		},
		{
			Name: (&gtpu.GtpuTunnelDump{}).GetMessageName(),
			Ping: true,
		},
	})

	intfs, err := ifHandler.DumpInterfaces(ctx.Context)
	Expect(err).To(BeNil())
	Expect(intfs).To(HaveLen(1))
	intface := intfs[0].Interface

	// Check vxlan
	Expect(intface.GetVxlan().SrcAddress).To(Equal("dead:beef:feed:face:cafe:babe:baad:c0de"))
	Expect(intface.GetVxlan().DstAddress).To(Equal("d3ad:beef:feed:face:cafe:babe:baad:c0de"))
}

// Test dump of interfaces with host type
func TestDumpInterfacesHost(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockReplies([]*vppmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: "host-localhost",
			},
		},
		{
			Name:    (&interfaces.SwInterfaceGetTable{}).GetMessageName(),
			Ping:    false,
			Message: &interfaces.SwInterfaceGetTableReply{},
		},
		{
			Name:    (&ip.IPAddressDump{}).GetMessageName(),
			Ping:    true,
			Message: &ip.IPAddressDetails{},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&memif.MemifDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&tapv2.SwInterfaceTapV2Dump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&vxlan.VxlanTunnelDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&gtpu.GtpuTunnelDump{}).GetMessageName(),
			Ping: true,
		},
	})

	intfs, err := ifHandler.DumpInterfaces(ctx.Context)
	Expect(err).To(BeNil())
	Expect(intfs).To(HaveLen(1))
	intface := intfs[0].Interface

	// Check interface data
	Expect(intface.GetAfpacket().HostIfName).To(Equal("localhost"))
}

// Test dump of interfaces with memif type
func TestDumpInterfacesMemif(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockReplies([]*vppmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: "memif1",
			},
		},
		{
			Name:    (&interfaces.SwInterfaceGetTable{}).GetMessageName(),
			Ping:    false,
			Message: &interfaces.SwInterfaceGetTableReply{},
		},
		{
			Name:    (&ip.IPAddressDump{}).GetMessageName(),
			Ping:    true,
			Message: &ip.IPAddressDetails{},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
		{
			Name: (&memif.MemifDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifDetails{
				ID:         2,
				SwIfIndex:  0,
				Role:       1, // Slave
				Mode:       1, // IP
				SocketID:   1,
				RingSize:   0,
				BufferSize: 0,
			},
		},
		{
			Name: (&tapv2.SwInterfaceTapV2Dump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&vxlan.VxlanTunnelDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&gtpu.GtpuTunnelDump{}).GetMessageName(),
			Ping: true,
		},
	})

	intfs, err := ifHandler.DumpInterfaces(ctx.Context)
	Expect(err).To(BeNil())
	Expect(intfs).To(HaveLen(1))
	intface := intfs[0].Interface

	// Check memif
	Expect(intface.GetMemif().SocketFilename).To(Equal("test"))
	Expect(intface.GetMemif().Id).To(Equal(uint32(2)))
	Expect(intface.GetMemif().Mode).To(Equal(interfaces2.MemifLink_IP))
	Expect(intface.GetMemif().Master).To(BeFalse())
}

func TestDumpInterfacesTap2(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	hwAddr1Parse, err := net.ParseMAC("01:23:45:67:89:ab")
	Expect(err).To(BeNil())

	ctx.MockReplies([]*vppmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				SwIfIndex:       0,
				InterfaceName:   "tap2",
				Tag:             "mytap2",
				AdminUpDown:     1,
				LinkMtu:         9216, // Default MTU
				L2Address:       hwAddr1Parse,
				L2AddressLength: uint32(len(hwAddr1Parse)),
			},
		},
		{
			Name: (&interfaces.SwInterfaceGetTable{}).GetMessageName(),
			Ping: false,
			Message: &interfaces.SwInterfaceGetTableReply{
				Retval: 0,
				VrfID:  42,
			},
		},
		{
			Name:    (&ip.IPAddressDump{}).GetMessageName(),
			Ping:    true,
			Message: &ip.IPAddressDetails{},
		},
		{
			Name: (&dhcp.DHCPClientDump{}).GetMessageName(),
			Ping: true,
			Message: &dhcp.DHCPClientDetails{
				Client: dhcp.DHCPClient{
					SwIfIndex: 0,
				},
			},
		},
		{
			Name: (&tapv2.SwInterfaceTapV2Dump{}).GetMessageName(),
			Ping: true,
			Message: &tapv2.SwInterfaceTapV2Details{
				SwIfIndex:  0,
				HostIfName: []byte("taptap2"),
			},
		},
		{
			Name: (&vxlan.VxlanTunnelDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&gtpu.GtpuTunnelDump{}).GetMessageName(),
			Ping: true,
		},
	})

	intfs, err := ifHandler.DumpInterfaces(ctx.Context)
	Expect(err).To(BeNil())
	Expect(intfs).To(HaveLen(1))

	intface := intfs[0].Interface
	intMeta := intfs[0].Meta

	// This is last checked type, so it will be equal to that
	Expect(intface.Type).To(Equal(interfaces2.Interface_TAP))
	Expect(intface.PhysAddress).To(Equal("01:23:45:67:89:ab"))
	Expect(intface.Name).To(Equal("mytap2"))
	Expect(intface.Mtu).To(Equal(uint32(0))) // default mtu
	Expect(intface.Enabled).To(BeTrue())
	Expect(intface.Vrf).To(Equal(uint32(42)))
	Expect(intface.SetDhcpClient).To(BeTrue())
	Expect(intface.GetTap().HostIfName).To(Equal("taptap2"))
	Expect(intface.GetTap().Version).To(Equal(uint32(2)))
	Expect(intMeta.VrfIPv4).To(Equal(uint32(42)))
	Expect(intMeta.VrfIPv6).To(Equal(uint32(42)))
}

// Test dump of memif socket details using standard reply mocking
func TestDumpMemifSocketDetails(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&memif.MemifSocketFilenameDetails{
		SocketID:       1,
		SocketFilename: []byte("test"),
	})

	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	result, err := ifHandler.DumpMemifSocketDetails(ctx.Context)
	Expect(err).To(BeNil())
	Expect(result).To(Not(BeEmpty()))

	socketID, ok := result["test"]
	Expect(ok).To(BeTrue())
	Expect(socketID).To(Equal(uint32(1)))
}

func TestDumpInterfacesRxPlacement(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockReplies([]*vppmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: "memif1",
			},
		},
		{
			Name:    (&interfaces.SwInterfaceGetTable{}).GetMessageName(),
			Ping:    false,
			Message: &interfaces.SwInterfaceGetTableReply{},
		},
		{
			Name:    (&ip.IPAddressDump{}).GetMessageName(),
			Ping:    true,
			Message: &ip.IPAddressDetails{},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifSocketFilenameDetails{
				SocketID:       1,
				SocketFilename: []byte("test"),
			},
		},
		{
			Name: (&memif.MemifDump{}).GetMessageName(),
			Ping: true,
			Message: &memif.MemifDetails{
				ID:         2,
				SwIfIndex:  0,
				Role:       1, // Slave
				Mode:       1, // IP
				SocketID:   1,
				RingSize:   0,
				BufferSize: 0,
			},
		},
		{
			Name: (&tapv2.SwInterfaceTapV2Dump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&vxlan.VxlanTunnelDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&gtpu.GtpuTunnelDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&interfaces.SwInterfaceRxPlacementDump{}).GetMessageName(),
			Ping: true,
			Messages: []govppapi.Message{
				&interfaces.SwInterfaceRxPlacementDetails{
					SwIfIndex: 0,
					QueueID:   0,
					WorkerID:  0, // main thread
					Mode:      3, // adaptive
				},
				&interfaces.SwInterfaceRxPlacementDetails{
					SwIfIndex: 0,
					QueueID:   1,
					WorkerID:  1, // worker 0
					Mode:      2, // interrupt
				},
				&interfaces.SwInterfaceRxPlacementDetails{
					SwIfIndex: 0,
					QueueID:   2,
					WorkerID:  2, // worker 1
					Mode:      1, // polling
				},
			},
		},
	})

	intfs, err := ifHandler.DumpInterfaces(ctx.Context)
	Expect(err).To(BeNil())
	Expect(intfs).To(HaveLen(1))
	intface := intfs[0].Interface

	// Check memif
	Expect(intface.GetMemif().SocketFilename).To(Equal("test"))
	Expect(intface.GetMemif().Id).To(Equal(uint32(2)))
	Expect(intface.GetMemif().Mode).To(Equal(interfaces2.MemifLink_IP))
	Expect(intface.GetMemif().Master).To(BeFalse())

	rxMode := intface.GetRxModes()
	Expect(rxMode).To(HaveLen(3))
	Expect(rxMode[0].Queue).To(BeEquivalentTo(0))
	Expect(rxMode[0].Mode).To(BeEquivalentTo(interfaces2.Interface_RxMode_ADAPTIVE))
	Expect(rxMode[1].Queue).To(BeEquivalentTo(1))
	Expect(rxMode[1].Mode).To(BeEquivalentTo(interfaces2.Interface_RxMode_INTERRUPT))
	Expect(rxMode[2].Queue).To(BeEquivalentTo(2))
	Expect(rxMode[2].Mode).To(BeEquivalentTo(interfaces2.Interface_RxMode_POLLING))

	rxPlacement := intface.GetRxPlacements()
	Expect(rxPlacement).To(HaveLen(3))
	Expect(rxPlacement[0].Queue).To(BeEquivalentTo(0))
	Expect(rxPlacement[0].MainThread).To(BeTrue())
	Expect(rxPlacement[0].Worker).To(BeEquivalentTo(0))
	Expect(rxPlacement[1].Queue).To(BeEquivalentTo(1))
	Expect(rxPlacement[1].MainThread).To(BeFalse())
	Expect(rxPlacement[1].Worker).To(BeEquivalentTo(0))
	Expect(rxPlacement[2].Queue).To(BeEquivalentTo(2))
	Expect(rxPlacement[2].MainThread).To(BeFalse())
	Expect(rxPlacement[2].Worker).To(BeEquivalentTo(1))
}

// Test dump of interfaces with gtpu type
func TestDumpInterfacesGtpu(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ipv61Parse := net.ParseIP("dead:beef:feed:face:cafe:babe:baad:c0de").To16()
	ipv62Parse := net.ParseIP("d3ad:beef:feed:face:cafe:babe:baad:c0de").To16()

	ctx.MockReplies([]*vppmock.HandleReplies{
		{
			Name: (&interfaces.SwInterfaceDump{}).GetMessageName(),
			Ping: true,
			Message: &interfaces.SwInterfaceDetails{
				InterfaceName: "gtpu1",
			},
		},
		{
			Name:    (&interfaces.SwInterfaceGetTable{}).GetMessageName(),
			Ping:    false,
			Message: &interfaces.SwInterfaceGetTableReply{},
		},
		{
			Name:    (&ip.IPAddressDump{}).GetMessageName(),
			Ping:    true,
			Message: &ip.IPAddressDetails{},
		},
		{
			Name: (&memif.MemifSocketFilenameDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&memif.MemifDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&tapv2.SwInterfaceTapV2Dump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&vxlan.VxlanTunnelDump{}).GetMessageName(),
			Ping: true,
		},
		{
			Name: (&gtpu.GtpuTunnelDump{}).GetMessageName(),
			Ping: true,
			Message: &gtpu.GtpuTunnelDetails{
				IsIPv6:     1,
				SwIfIndex:  0,
				SrcAddress: ipv61Parse,
				DstAddress: ipv62Parse,
				EncapVrfID: 16,
				Teid:       100,
			},
		},
	})

	intfs, err := ifHandler.DumpInterfaces(ctx.Context)
	Expect(err).To(BeNil())
	Expect(intfs).To(HaveLen(1))
	intface := intfs[0].Interface

	// Check gtpu
	Expect(intface.Type).To(Equal(interfaces2.Interface_GTPU_TUNNEL))
	Expect(intface.GetGtpu().SrcAddr).To(Equal("dead:beef:feed:face:cafe:babe:baad:c0de"))
	Expect(intface.GetGtpu().DstAddr).To(Equal("d3ad:beef:feed:face:cafe:babe:baad:c0de"))
	Expect(intface.GetGtpu().EncapVrfId).To(Equal(uint32(16)))
	Expect(intface.GetGtpu().Teid).To(Equal(uint32(100)))
}
