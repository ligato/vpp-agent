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

package vpp2005_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ip_types"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ipsec_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls/vpp2005"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

func ipToAddr(ip string) vpp_ipsec.Address {
	addr, err := vpp2005.IPToAddress(ip)
	if err != nil {
		panic(fmt.Sprintf("invalid IP: %s", ip))
	}
	return addr
}

func TestVppAddSPD(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSpdAddDelReply{})

	err := ipSecHandler.AddSPD(10)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSpdAddDel{
		IsAdd: true,
		SpdID: 10,
	}))
}

func TestVppDelSPD(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSpdAddDelReply{})

	err := ipSecHandler.DeleteSPD(10)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSpdAddDel{
		IsAdd: false,
		SpdID: 10,
	}))
}

func TestVppAddSP(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSpdEntryAddDelReply{})

	err := ipSecHandler.AddSP(&ipsec.SecurityPolicy{
		SaIndex:    5,
		SpdIndex:   10,
		Priority:   10,
		IsOutbound: true,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSpdEntryAddDel{
		IsAdd: true,
		Entry: vpp_ipsec.IpsecSpdEntry{
			SpdID:              10,
			SaID:               5,
			Priority:           10,
			IsOutbound:         true,
			RemoteAddressStart: ipToAddr("0.0.0.0"),
			RemoteAddressStop:  ipToAddr("255.255.255.255"),
			LocalAddressStart:  ipToAddr("0.0.0.0"),
			LocalAddressStop:   ipToAddr("255.255.255.255"),
			RemotePortStop:     65535,
			LocalPortStop:      65535,
		},
	}))
}

func TestVppDelSP(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSpdEntryAddDelReply{})

	err := ipSecHandler.DeleteSP(&ipsec.SecurityPolicy{
		SpdIndex:   10,
		SaIndex:    2,
		Priority:   5,
		IsOutbound: true,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSpdEntryAddDel{
		IsAdd: false,
		Entry: vpp_ipsec.IpsecSpdEntry{
			SpdID:              10,
			SaID:               2,
			Priority:           5,
			IsOutbound:         true,
			RemoteAddressStart: ipToAddr("0.0.0.0"),
			RemoteAddressStop:  ipToAddr("255.255.255.255"),
			LocalAddressStart:  ipToAddr("0.0.0.0"),
			LocalAddressStop:   ipToAddr("255.255.255.255"),
			RemotePortStop:     65535,
			LocalPortStop:      65535,
		},
	}))
}

func TestVppInterfaceAddSPD(t *testing.T) {
	ctx, ipSecHandler, ifIndex := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecInterfaceAddDelSpdReply{})

	ifIndex.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	err := ipSecHandler.AddSPDInterface(10, &ipsec.SecurityPolicyDatabase_Interface{
		Name: "if1",
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecInterfaceAddDelSpd{
		IsAdd:     true,
		SpdID:     10,
		SwIfIndex: 2,
	}))
}

func TestVppInterfaceDelSPD(t *testing.T) {
	ctx, ipSecHandler, ifIndex := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecInterfaceAddDelSpdReply{})

	ifIndex.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	err := ipSecHandler.DeleteSPDInterface(10, &ipsec.SecurityPolicyDatabase_Interface{
		Name: "if1",
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecInterfaceAddDelSpd{
		IsAdd:     false,
		SpdID:     10,
		SwIfIndex: 2,
	}))
}

func ipSecTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.IPSecVppAPI, ifaceidx.IfaceMetadataIndexRW) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	ifIndex := ifaceidx.NewIfaceIndex(log, "ipsec-test-ifidx")
	ipSecHandler := vpp2005.NewIPSecVppHandler(ctx.MockChannel, ifIndex, log)
	return ctx, ipSecHandler, ifIndex
}

func TestVppAddSA(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSadEntryAddDelReply{})

	cryptoKey, err := hex.DecodeString("")
	Expect(err).To(BeNil())

	err = ipSecHandler.AddSA(&ipsec.SecurityAssociation{
		Index:         1,
		Spi:           uint32(1001),
		UseEsn:        true,
		UseAntiReplay: true,
		Protocol:      ipsec.SecurityAssociation_ESP,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSadEntryAddDel{
		IsAdd: true,
		Entry: vpp_ipsec.IpsecSadEntry{
			Protocol: ipsec_types.IPSEC_API_PROTO_ESP,
			SadID:    1,
			Spi:      1001,
			CryptoKey: vpp_ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			IntegrityKey: vpp_ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			Flags:      ipsec_types.IPSEC_API_SAD_FLAG_USE_ESN | ipsec_types.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY,
			UDPSrcPort: ^uint16(0),
			UDPDstPort: ^uint16(0),
		},
	}))
}

func TestVppDelSA(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSadEntryAddDelReply{})

	cryptoKey, err := hex.DecodeString("")
	Expect(err).To(BeNil())

	err = ipSecHandler.DeleteSA(&ipsec.SecurityAssociation{
		Index:         1,
		Spi:           uint32(1001),
		UseEsn:        true,
		UseAntiReplay: true,
		Protocol:      ipsec.SecurityAssociation_ESP,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSadEntryAddDel{
		IsAdd: false,
		Entry: vpp_ipsec.IpsecSadEntry{
			SadID: 1,
			Spi:   1001,
			CryptoKey: vpp_ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			IntegrityKey: vpp_ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			Flags:      ipsec_types.IPSEC_API_SAD_FLAG_USE_ESN | ipsec_types.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY,
			Protocol:   ipsec_types.IPSEC_API_PROTO_ESP,
			UDPSrcPort: ^uint16(0),
			UDPDstPort: ^uint16(0),
		},
	}))
}

func TestVppAddSATunnelMode(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSadEntryAddDelReply{})

	cryptoKey, err := hex.DecodeString("")
	Expect(err).To(BeNil())

	err = ipSecHandler.AddSA(&ipsec.SecurityAssociation{
		Index:         1,
		Spi:           uint32(1001),
		TunnelSrcAddr: "10.1.0.1",
		TunnelDstAddr: "20.1.0.1",
		Protocol:      ipsec.SecurityAssociation_ESP,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSadEntryAddDel{
		IsAdd: true,
		Entry: vpp_ipsec.IpsecSadEntry{
			SadID: 1,
			Spi:   1001,
			CryptoKey: vpp_ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			IntegrityKey: vpp_ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			TunnelSrc: ipsec_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ipsec_types.AddressUnion{XXX_UnionData: [16]byte{10, 1, 0, 1}},
			},
			TunnelDst: ipsec_types.Address{
				Af: ip_types.ADDRESS_IP4,
				Un: ipsec_types.AddressUnion{XXX_UnionData: [16]byte{20, 1, 0, 1}},
			},
			Flags:      ipsec_types.IPSEC_API_SAD_FLAG_IS_TUNNEL,
			Protocol:   ipsec_types.IPSEC_API_PROTO_ESP,
			UDPSrcPort: ^uint16(0),
			UDPDstPort: ^uint16(0),
		},
	}))
}

func TestVppAddSATunnelModeIPv6(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSadEntryAddDelReply{})

	cryptoKey, err := hex.DecodeString("")
	Expect(err).To(BeNil())

	err = ipSecHandler.AddSA(&ipsec.SecurityAssociation{
		Index:         1,
		Spi:           uint32(1001),
		TunnelSrcAddr: "1234::",
		TunnelDstAddr: "abcd::",
		Protocol:      ipsec.SecurityAssociation_ESP,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSadEntryAddDel{
		IsAdd: true,
		Entry: vpp_ipsec.IpsecSadEntry{
			SadID: 1,
			Spi:   1001,
			CryptoKey: vpp_ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			IntegrityKey: vpp_ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			TunnelSrc: ipsec_types.Address{
				Af: ip_types.ADDRESS_IP6,
				Un: ipsec_types.AddressUnion{XXX_UnionData: [16]byte{18, 52}},
			},
			TunnelDst: ipsec_types.Address{
				Af: ip_types.ADDRESS_IP6,
				Un: ipsec_types.AddressUnion{XXX_UnionData: [16]byte{171, 205}},
			},
			Flags:      ipsec_types.IPSEC_API_SAD_FLAG_IS_TUNNEL | ipsec_types.IPSEC_API_SAD_FLAG_IS_TUNNEL_V6,
			Protocol:   ipsec_types.IPSEC_API_PROTO_ESP,
			UDPSrcPort: ^uint16(0),
			UDPDstPort: ^uint16(0),
		},
	}))
}

func TestVppAddTunnelProtection(t *testing.T) {
	ctx, ipSecHandler, ifIndex := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndex.Put("ipip-tunnel", &ifaceidx.IfaceMetadata{SwIfIndex: 5})

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecTunnelProtectUpdateReply{})
	err := ipSecHandler.AddTunnelProtection(&ipsec.TunnelProtection{
		Interface: "ipip-tunnel",
		SaOut:     []uint32{10},
		SaIn:      []uint32{20, 30},
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecTunnelProtectUpdate{
		Tunnel: vpp_ipsec.IpsecTunnelProtect{
			SwIfIndex: 5,
			SaOut:     10,
			NSaIn:     2,
			SaIn:      []uint32{20, 30},
		},
	}))
}

func TestVppDelTunnelProtection(t *testing.T) {
	ctx, ipSecHandler, ifIndex := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndex.Put("ipip-tunnel", &ifaceidx.IfaceMetadata{SwIfIndex: 5})

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecTunnelProtectDelReply{})

	err := ipSecHandler.DeleteTunnelProtection(&ipsec.TunnelProtection{
		Interface: "ipip-tunnel",
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecTunnelProtectDel{
		SwIfIndex: 5,
	}))
}
