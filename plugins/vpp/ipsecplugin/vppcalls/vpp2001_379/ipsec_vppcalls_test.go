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

package vpp2001_379_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	ipsec "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	vpp_ipsec "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp2001_379/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls/vpp2001_379"
	"github.com/ligato/vpp-agent/plugins/vpp/vppcallmock"
	. "github.com/onsi/gomega"
)

func ipToAddr(ip string) vpp_ipsec.Address {
	addr, err := vpp2001_379.IPToAddress(ip)
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
		IsAdd: 1,
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
		IsAdd: 0,
		SpdID: 10,
	}))
}

func TestVppAddSPDEntry(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSpdEntryAddDelReply{})

	err := ipSecHandler.AddSPDEntry(10, 5, &ipsec.SecurityPolicyDatabase_PolicyEntry{
		SaIndex:    "5",
		Priority:   10,
		IsOutbound: true,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSpdEntryAddDel{
		IsAdd: 1,
		Entry: vpp_ipsec.IpsecSpdEntry{
			SpdID:              10,
			SaID:               5,
			Priority:           10,
			IsOutbound:         1,
			RemoteAddressStart: ipToAddr("0.0.0.0"),
			RemoteAddressStop:  ipToAddr("255.255.255.255"),
			LocalAddressStart:  ipToAddr("0.0.0.0"),
			LocalAddressStop:   ipToAddr("255.255.255.255"),
			RemotePortStop:     65535,
			LocalPortStop:      65535,
		},
	}))
}

func TestVppDelSPDEntry(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSpdEntryAddDelReply{})

	err := ipSecHandler.DeleteSPDEntry(10, 2, &ipsec.SecurityPolicyDatabase_PolicyEntry{
		SaIndex:    "2",
		Priority:   5,
		IsOutbound: true,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSpdEntryAddDel{
		IsAdd: 0,
		Entry: vpp_ipsec.IpsecSpdEntry{
			SpdID:              10,
			SaID:               2,
			Priority:           5,
			IsOutbound:         1,
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
		IsAdd:     1,
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
		IsAdd:     0,
		SpdID:     10,
		SwIfIndex: 2,
	}))
}

func ipSecTestSetup(t *testing.T) (*vppcallmock.TestCtx, vppcalls.IPSecVppAPI, ifaceidx.IfaceMetadataIndexRW) {
	ctx := vppcallmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	ifIndex := ifaceidx.NewIfaceIndex(log, "ipsec-test-ifidx")
	ipSecHandler := vpp2001_379.NewIPSecVppHandler(ctx.MockChannel, ifIndex, log)
	return ctx, ipSecHandler, ifIndex
}

func TestVppAddSA(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecSadEntryAddDelReply{})

	cryptoKey, err := hex.DecodeString("")
	Expect(err).To(BeNil())

	err = ipSecHandler.AddSA(&ipsec.SecurityAssociation{
		Index:         "1",
		Spi:           uint32(1001),
		UseEsn:        true,
		UseAntiReplay: true,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSadEntryAddDel{
		IsAdd: 1,
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
			Flags: vpp_ipsec.IPSEC_API_SAD_FLAG_USE_ESN | vpp_ipsec.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY,
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
		Index:         "1",
		Spi:           uint32(1001),
		UseEsn:        true,
		UseAntiReplay: true,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSadEntryAddDel{
		IsAdd: 0,
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
			Flags: vpp_ipsec.IPSEC_API_SAD_FLAG_USE_ESN | vpp_ipsec.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY,
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
		Index:         "1",
		Spi:           uint32(1001),
		TunnelSrcAddr: "10.1.0.1",
		TunnelDstAddr: "20.1.0.1",
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSadEntryAddDel{
		IsAdd: 1,
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
			TunnelSrc: vpp_ipsec.Address{
				Af: vpp_ipsec.ADDRESS_IP4,
				Un: vpp_ipsec.AddressUnion{XXX_UnionData: [16]byte{10, 1, 0, 1}},
			},
			TunnelDst: vpp_ipsec.Address{
				Af: vpp_ipsec.ADDRESS_IP4,
				Un: vpp_ipsec.AddressUnion{XXX_UnionData: [16]byte{20, 1, 0, 1}},
			},
			Flags: vpp_ipsec.IPSEC_API_SAD_FLAG_IS_TUNNEL,
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
		Index:         "1",
		Spi:           uint32(1001),
		TunnelSrcAddr: "1234::",
		TunnelDstAddr: "abcd::",
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&vpp_ipsec.IpsecSadEntryAddDel{
		IsAdd: 1,
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
			TunnelSrc: vpp_ipsec.Address{
				Af: vpp_ipsec.ADDRESS_IP6,
				Un: vpp_ipsec.AddressUnion{XXX_UnionData: [16]byte{18, 52}},
			},
			TunnelDst: vpp_ipsec.Address{
				Af: vpp_ipsec.ADDRESS_IP6,
				Un: vpp_ipsec.AddressUnion{XXX_UnionData: [16]byte{171, 205}},
			},
			Flags: vpp_ipsec.IPSEC_API_SAD_FLAG_IS_TUNNEL | vpp_ipsec.IPSEC_API_SAD_FLAG_IS_TUNNEL_V6,
		},
	}))
}
