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
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	ipsec2 "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls/vpp1908"
	"github.com/ligato/vpp-agent/plugins/vpp/vppcallmock"
	. "github.com/onsi/gomega"
)

func ipToAddr(ip string) ipsec.Address {
	addr, err := vpp1908.IPToAddress(ip)
	if err != nil {
		panic(fmt.Sprintf("invalid IP: %s", ip))
	}
	return addr
}

func TestVppAddSPD(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ipsec.IpsecSpdAddDelReply{})

	err := ipSecHandler.AddSPD(10)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&ipsec.IpsecSpdAddDel{
		IsAdd: 1,
		SpdID: 10,
	}))
}

func TestVppDelSPD(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ipsec.IpsecSpdAddDelReply{})

	err := ipSecHandler.DeleteSPD(10)

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&ipsec.IpsecSpdAddDel{
		IsAdd: 0,
		SpdID: 10,
	}))
}

func TestVppAddSPDEntry(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ipsec.IpsecSpdEntryAddDelReply{})

	err := ipSecHandler.AddSPDEntry(10, 5, &ipsec2.SecurityPolicyDatabase_PolicyEntry{
		SaIndex:    "5",
		Priority:   10,
		IsOutbound: true,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&ipsec.IpsecSpdEntryAddDel{
		IsAdd: 1,
		Entry: ipsec.IpsecSpdEntry{
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

	ctx.MockVpp.MockReply(&ipsec.IpsecSpdEntryAddDelReply{})

	err := ipSecHandler.DeleteSPDEntry(10, 2, &ipsec2.SecurityPolicyDatabase_PolicyEntry{
		SaIndex:    "2",
		Priority:   5,
		IsOutbound: true,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&ipsec.IpsecSpdEntryAddDel{
		IsAdd: 0,
		Entry: ipsec.IpsecSpdEntry{
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

	ctx.MockVpp.MockReply(&ipsec.IpsecInterfaceAddDelSpdReply{})

	ifIndex.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	err := ipSecHandler.AddSPDInterface(10, &ipsec2.SecurityPolicyDatabase_Interface{
		Name: "if1",
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&ipsec.IpsecInterfaceAddDelSpd{
		IsAdd:     1,
		SpdID:     10,
		SwIfIndex: 2,
	}))
}

func TestVppInterfaceDelSPD(t *testing.T) {
	ctx, ipSecHandler, ifIndex := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ipsec.IpsecInterfaceAddDelSpdReply{})

	ifIndex.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	err := ipSecHandler.DeleteSPDInterface(10, &ipsec2.SecurityPolicyDatabase_Interface{
		Name: "if1",
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&ipsec.IpsecInterfaceAddDelSpd{
		IsAdd:     0,
		SpdID:     10,
		SwIfIndex: 2,
	}))
}

func ipSecTestSetup(t *testing.T) (*vppcallmock.TestCtx, vppcalls.IPSecVppAPI, ifaceidx.IfaceMetadataIndexRW) {
	ctx := vppcallmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	ifIndex := ifaceidx.NewIfaceIndex(log, "ipsec-test-ifidx")
	ipSecHandler := vpp1908.NewIPSecVppHandler(ctx.MockChannel, ifIndex, log)
	return ctx, ipSecHandler, ifIndex
}

func TestVppAddSA(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ipsec.IpsecSadEntryAddDelReply{})

	cryptoKey, err := hex.DecodeString("")
	Expect(err).To(BeNil())

	err = ipSecHandler.AddSA(&ipsec2.SecurityAssociation{
		Index:         "1",
		Spi:           uint32(1001),
		UseEsn:        true,
		UseAntiReplay: true,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&ipsec.IpsecSadEntryAddDel{
		IsAdd: 1,
		Entry: ipsec.IpsecSadEntry{
			SadID: 1,
			Spi:   1001,
			CryptoKey: ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			IntegrityKey: ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			Flags: ipsec.IPSEC_API_SAD_FLAG_USE_ESN | ipsec.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY,
		},
	}))
}

func TestVppDelSA(t *testing.T) {
	ctx, ipSecHandler, _ := ipSecTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ipsec.IpsecSadEntryAddDelReply{})

	cryptoKey, err := hex.DecodeString("")
	Expect(err).To(BeNil())

	err = ipSecHandler.DeleteSA(&ipsec2.SecurityAssociation{
		Index:         "1",
		Spi:           uint32(1001),
		UseEsn:        true,
		UseAntiReplay: true,
	})

	Expect(err).ShouldNot(HaveOccurred())
	Expect(ctx.MockChannel.Msg).To(BeEquivalentTo(&ipsec.IpsecSadEntryAddDel{
		IsAdd: 0,
		Entry: ipsec.IpsecSadEntry{
			SadID: 1,
			Spi:   1001,
			CryptoKey: ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			IntegrityKey: ipsec.Key{
				Length: uint8(len(cryptoKey)),
				Data:   cryptoKey,
			},
			Flags: ipsec.IPSEC_API_SAD_FLAG_USE_ESN | ipsec.IPSEC_API_SAD_FLAG_USE_ANTI_REPLAY,
		},
	}))
}
