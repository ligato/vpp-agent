//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package e2e

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

func TestIPSec(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		msName       = "microservice1"
		tunnelIfName = "ipsec-tunnel"
	)

	// configure IPIP tunnel with IPSec tunnel protection

	ipipTun := &vpp_interfaces.Interface{
		Name:    tunnelIfName,
		Enabled: true,
		Type:    vpp_interfaces.Interface_IPIP_TUNNEL,

		Link: &vpp_interfaces.Interface_Ipip{
			Ipip: &vpp_interfaces.IPIPLink{
				DstAddr: "8.8.8.8",
				SrcAddr: "1.2.3.4",
			},
		},
	}
	saOut := &vpp_ipsec.SecurityAssociation{
		Index:          10,
		Spi:            123,
		Protocol:       vpp_ipsec.SecurityAssociation_ESP,
		CryptoAlg:      vpp_ipsec.CryptoAlg_AES_GCM_128,
		CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
		CryptoSalt:     1500,
		IntegAlg:       vpp_ipsec.IntegAlg_SHA1_96,
		IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
		EnableUdpEncap: true,
		TunnelSrcPort:  4500,
		TunnelDstPort:  8777,
	}
	saIn := &vpp_ipsec.SecurityAssociation{
		Index:          20,
		Spi:            456,
		Protocol:       vpp_ipsec.SecurityAssociation_ESP,
		CryptoAlg:      vpp_ipsec.CryptoAlg_AES_GCM_128,
		CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
		CryptoSalt:     8900,
		IntegAlg:       vpp_ipsec.IntegAlg_SHA1_96,
		IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
		EnableUdpEncap: true,
		TunnelSrcPort:  8777,
		TunnelDstPort:  4500,
	}
	spOut := &vpp_ipsec.SecurityPolicy{
		SpdIndex:        100,
		SaIndex:         10,
		Priority:        0,
		IsOutbound:      true,
		RemoteAddrStart: "10.10.1.1",
		RemoteAddrStop:  "10.10.1.255",
		LocalAddrStart:  "10.10.2.1",
		LocalAddrStop:   "10.10.2.255",
		Protocol:        0,
		RemotePortStart: 100,
		RemotePortStop:  2000,
		LocalPortStart:  0,
		LocalPortStop:   65535,
		Action:          vpp_ipsec.SecurityPolicy_PROTECT,
	}
	spIn := &vpp_ipsec.SecurityPolicy{
		SpdIndex:        100,
		SaIndex:         20,
		Priority:        0,
		IsOutbound:      false,
		RemoteAddrStart: "10.10.1.1",
		RemoteAddrStop:  "10.10.1.255",
		LocalAddrStart:  "10.10.2.1",
		LocalAddrStop:   "10.10.2.255",
		Protocol:        0,
		RemotePortStart: 0,
		RemotePortStop:  65535,
		LocalPortStart:  0,
		LocalPortStop:   65535,
		Action:          vpp_ipsec.SecurityPolicy_PROTECT,
	}
	spd := &vpp_ipsec.SecurityPolicyDatabase{
		Index: 100,
		Interfaces: []*vpp_ipsec.SecurityPolicyDatabase_Interface{
			{
				Name: tunnelIfName,
			},
		},
	}
	tp := &vpp_ipsec.TunnelProtection{
		Interface: tunnelIfName,
		SaOut:     []uint32{saOut.Index},
		SaIn:      []uint32{saIn.Index},
	}

	ctx.StartMicroservice(msName)
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		ipipTun,
		saOut,
		saIn,
		tp,
		spd,
		spIn,
		spOut,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Sending change request failed with err")

	ctx.Eventually(ctx.GetValueStateClb(ipipTun)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IPIP tunnel is not configured")
	ctx.Eventually(ctx.GetValueStateClb(saOut)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"OUT SA is not configured")
	ctx.Eventually(ctx.GetValueStateClb(saIn)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IN SA is not configured")
	ctx.Eventually(ctx.GetValueStateClb(tp)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"tunnel protection is not configured")
	ctx.Eventually(ctx.GetValueStateClb(spd)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"SPD is not configured")
	ctx.Eventually(ctx.GetValueStateClb(spIn)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IN SP is not configured")
	ctx.Eventually(ctx.GetValueStateClb(spOut)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"OUT SP is not configured")

	if ctx.VppRelease() >= "20.05" {
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())
	}

	// rekey - delete old SAs, create new SAs and modify tunnel protection

	saOutNew := &vpp_ipsec.SecurityAssociation{
		Index:          11,
		Spi:            888,
		Protocol:       vpp_ipsec.SecurityAssociation_ESP,
		CryptoAlg:      vpp_ipsec.CryptoAlg_AES_CBC_128,
		CryptoKey:      "a9a4ec50aed76f1bf80bc915d8fcfe1d",
		IntegAlg:       vpp_ipsec.IntegAlg_SHA1_96,
		IntegKey:       "cf9b150aaf5c2a87d79898b11eabd055e70abdbf",
		EnableUdpEncap: true,
	}
	saInNew := &vpp_ipsec.SecurityAssociation{
		Index:          21,
		Spi:            999,
		Protocol:       vpp_ipsec.SecurityAssociation_ESP,
		CryptoAlg:      vpp_ipsec.CryptoAlg_AES_CBC_128,
		CryptoKey:      "a9a4ec50aed76f1bf80bc915d8fcfe1d",
		IntegAlg:       vpp_ipsec.IntegAlg_SHA1_96,
		IntegKey:       "cf9b150aaf5c2a87d79898b11eabd055e70abdbf",
		EnableUdpEncap: true,
	}
	tpNew := &vpp_ipsec.TunnelProtection{
		Interface: tunnelIfName,
		SaOut:     []uint32{saOutNew.Index},
		SaIn:      []uint32{saInNew.Index},
	}
	spOutNew := &vpp_ipsec.SecurityPolicy{
		SpdIndex:        100,
		SaIndex:         11,
		Priority:        0,
		IsOutbound:      true,
		RemoteAddrStart: "10.10.1.1",
		RemoteAddrStop:  "10.10.1.255",
		LocalAddrStart:  "10.10.2.1",
		LocalAddrStop:   "10.10.2.255",
		Protocol:        0,
		RemotePortStart: 0,
		RemotePortStop:  65535,
		LocalPortStart:  0,
		LocalPortStop:   65535,
		Action:          vpp_ipsec.SecurityPolicy_PROTECT,
	}
	spInNew := &vpp_ipsec.SecurityPolicy{
		SpdIndex:        100,
		SaIndex:         21,
		Priority:        0,
		IsOutbound:      false,
		RemoteAddrStart: "10.10.1.1",
		RemoteAddrStop:  "10.10.1.255",
		LocalAddrStart:  "10.10.2.1",
		LocalAddrStop:   "10.10.2.255",
		Protocol:        0,
		RemotePortStart: 0,
		RemotePortStop:  65535,
		LocalPortStart:  0,
		LocalPortStop:   65535,
		Action:          vpp_ipsec.SecurityPolicy_PROTECT,
	}

	req2 := ctx.GenericClient().ChangeRequest()
	err = req2.
		Delete(
			saOut,
			saIn,
			spOut,
			spIn).
		Update(
			saOutNew,
			saInNew,
			spOutNew,
			spInNew,
			tpNew,
		).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Sending change request failed with err")

	ctx.Eventually(ctx.GetValueStateClb(saOut)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"old OUT SA was not removed")
	ctx.Eventually(ctx.GetValueStateClb(saIn)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"old IN SA was not removed")
	ctx.Eventually(ctx.GetValueStateClb(saOutNew)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"OUT SA is not configured")
	ctx.Eventually(ctx.GetValueStateClb(saInNew)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IN SA is not configured")
	ctx.Eventually(ctx.GetValueStateClb(tpNew)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"tunnel protection is not configured")
	ctx.Eventually(ctx.GetValueStateClb(spOut)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"old OUT SP was not removed")
	ctx.Eventually(ctx.GetValueStateClb(spIn)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"old IN SP was not removed")
	ctx.Eventually(ctx.GetValueStateClb(spOutNew)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"OUT SP is not configured")
	ctx.Eventually(ctx.GetValueStateClb(spInNew)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IN SP is not configured")

	if ctx.VppRelease() >= "20.05" {
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())
	}

	// delete the tunnel

	req3 := ctx.GenericClient().ChangeRequest()
	err = req3.Delete(
		saOutNew,
		saInNew,
		tpNew,
		ipipTun,
		spInNew,
		spOutNew,
		spd,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Sending change request failed with err")

	ctx.Eventually(ctx.GetValueStateClb(saOutNew)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"OUT SA was not removed")
	ctx.Eventually(ctx.GetValueStateClb(saInNew)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"IN SA was not removed")
	ctx.Eventually(ctx.GetValueStateClb(spOutNew)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"OUT SP was not removed")
	ctx.Eventually(ctx.GetValueStateClb(spInNew)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"IN SP was not removed")
	ctx.Eventually(ctx.GetValueStateClb(spd)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"SPD was not removed")
	ctx.Eventually(ctx.GetValueStateClb(tpNew)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"tunnel protection was not removed")
	ctx.Eventually(ctx.GetValueStateClb(ipipTun)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"IPIP tunnel was not removed")

	if ctx.VppRelease() >= "20.05" {
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())
	}
}

func TestIPSecMultiPoint(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	if ctx.VppRelease() < "20.05" {
		t.Skipf("IPSec MP: skipped for VPP < 20.05 (%s)", ctx.VppRelease())
	}

	const (
		msName       = "microservice1"
		tunnelIfName = "ipsec-tunnel"
	)

	ipipTun := &vpp_interfaces.Interface{
		Name:    tunnelIfName,
		Enabled: true,
		Type:    vpp_interfaces.Interface_IPIP_TUNNEL,

		Link: &vpp_interfaces.Interface_Ipip{
			Ipip: &vpp_interfaces.IPIPLink{
				SrcAddr:    "1.2.3.4",
				TunnelMode: vpp_interfaces.IPIPLink_POINT_TO_MULTIPOINT,
			},
		},
		IpAddresses: []string{"192.168.0.1/24"},
	}
	saOut1 := &vpp_ipsec.SecurityAssociation{
		Index:          10,
		Spi:            123,
		Protocol:       vpp_ipsec.SecurityAssociation_ESP,
		CryptoAlg:      vpp_ipsec.CryptoAlg_AES_CBC_128,
		CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
		IntegAlg:       vpp_ipsec.IntegAlg_SHA1_96,
		IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
		EnableUdpEncap: true,
	}
	saIn1 := &vpp_ipsec.SecurityAssociation{
		Index:          20,
		Spi:            456,
		Protocol:       vpp_ipsec.SecurityAssociation_ESP,
		CryptoAlg:      vpp_ipsec.CryptoAlg_AES_CBC_128,
		CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
		IntegAlg:       vpp_ipsec.IntegAlg_SHA1_96,
		IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
		EnableUdpEncap: true,
	}
	saOut2 := &vpp_ipsec.SecurityAssociation{
		Index:          30,
		Spi:            789,
		Protocol:       vpp_ipsec.SecurityAssociation_ESP,
		CryptoAlg:      vpp_ipsec.CryptoAlg_AES_CBC_128,
		CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
		IntegAlg:       vpp_ipsec.IntegAlg_SHA1_96,
		IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
		EnableUdpEncap: true,
	}
	saIn2 := &vpp_ipsec.SecurityAssociation{
		Index:          40,
		Spi:            111,
		Protocol:       vpp_ipsec.SecurityAssociation_ESP,
		CryptoAlg:      vpp_ipsec.CryptoAlg_AES_CBC_128,
		CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
		IntegAlg:       vpp_ipsec.IntegAlg_SHA1_96,
		IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
		EnableUdpEncap: true,
	}
	tp1 := &vpp_ipsec.TunnelProtection{
		Interface:   tunnelIfName,
		SaOut:       []uint32{saOut1.Index},
		SaIn:        []uint32{saIn1.Index},
		NextHopAddr: "192.168.0.2",
	}
	tp2 := &vpp_ipsec.TunnelProtection{
		Interface:   tunnelIfName,
		SaOut:       []uint32{saOut2.Index},
		SaIn:        []uint32{saIn2.Index},
		NextHopAddr: "192.168.0.3",
	}
	teib1 := &vpp_l3.TeibEntry{
		Interface:   tunnelIfName,
		PeerAddr:    tp1.NextHopAddr,
		NextHopAddr: "8.8.8.8",
	}
	teib2 := &vpp_l3.TeibEntry{
		Interface:   tunnelIfName,
		PeerAddr:    tp2.NextHopAddr,
		NextHopAddr: "8.8.8.9",
	}
	spOut1 := &vpp_ipsec.SecurityPolicy{
		SpdIndex:        100,
		SaIndex:         10,
		Priority:        0,
		IsOutbound:      true,
		RemoteAddrStart: "10.10.1.1",
		RemoteAddrStop:  "10.10.1.255",
		LocalAddrStart:  "10.10.2.1",
		LocalAddrStop:   "10.10.2.255",
		Protocol:        0,
		RemotePortStart: 0,
		RemotePortStop:  65535,
		LocalPortStart:  0,
		LocalPortStop:   65535,
		Action:          vpp_ipsec.SecurityPolicy_PROTECT,
	}
	spIn1 := &vpp_ipsec.SecurityPolicy{
		SpdIndex:        100,
		SaIndex:         20,
		Priority:        0,
		IsOutbound:      false,
		RemoteAddrStart: "10.10.1.1",
		RemoteAddrStop:  "10.10.1.255",
		LocalAddrStart:  "10.10.2.1",
		LocalAddrStop:   "10.10.2.255",
		Protocol:        0,
		RemotePortStart: 0,
		RemotePortStop:  65535,
		LocalPortStart:  0,
		LocalPortStop:   65535,
		Action:          vpp_ipsec.SecurityPolicy_PROTECT,
	}
	spOut2 := &vpp_ipsec.SecurityPolicy{
		SpdIndex:        100,
		SaIndex:         30,
		Priority:        0,
		IsOutbound:      true,
		RemoteAddrStart: "10.20.1.1",
		RemoteAddrStop:  "10.20.1.255",
		LocalAddrStart:  "10.20.2.1",
		LocalAddrStop:   "10.20.2.255",
		Protocol:        0,
		RemotePortStart: 0,
		RemotePortStop:  65535,
		LocalPortStart:  0,
		LocalPortStop:   65535,
		Action:          vpp_ipsec.SecurityPolicy_PROTECT,
	}
	spIn2 := &vpp_ipsec.SecurityPolicy{
		SpdIndex:        100,
		SaIndex:         40,
		Priority:        0,
		IsOutbound:      false,
		RemoteAddrStart: "10.20.1.1",
		RemoteAddrStop:  "10.20.1.255",
		LocalAddrStart:  "10.20.2.1",
		LocalAddrStop:   "10.20.2.255",
		Protocol:        0,
		RemotePortStart: 0,
		RemotePortStop:  65535,
		LocalPortStart:  0,
		LocalPortStop:   65535,
		Action:          vpp_ipsec.SecurityPolicy_PROTECT,
	}
	spd := &vpp_ipsec.SecurityPolicyDatabase{
		Index: 100,
		Interfaces: []*vpp_ipsec.SecurityPolicyDatabase_Interface{
			{
				Name: tunnelIfName,
			},
		},
	}

	ctx.StartMicroservice(msName)
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		ipipTun,
		saOut1, saIn1, saOut2, saIn2,
		spOut1, spIn1, spOut2, spIn2, spd,
		tp1, tp2,
		teib1, teib2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Sending change request failed with err")

	ctx.Eventually(ctx.GetValueStateClb(ipipTun)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IPIP tunnel is not configured")
	ctx.Eventually(ctx.GetValueStateClb(saOut1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"OUT SA 1 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(saIn1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IN SA 1 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(saOut2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"OUT SA 2 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(saIn2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IN SA 2 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(tp1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"tunnel protection 1 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(tp2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"tunnel protection 2 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(teib1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TEIB 1 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(teib2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"TEIB 2 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(spOut1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"OUT SP 1 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(spIn1)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IN SP 1 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(spOut2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"OUT SP 2 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(spIn2)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IN SP 2 is not configured")
	ctx.Eventually(ctx.GetValueStateClb(spd)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"SPD is not configured")

	if ctx.VppRelease() >= "20.05" {
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())
	}

	req3 := ctx.GenericClient().ChangeRequest()
	err = req3.Delete(
		ipipTun,
		saOut1, saIn1, saOut2, saIn2,
		spOut1, spIn1, spOut2, spIn2, spd,
		tp1, tp2,
		teib1, teib2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred(), "Sending change request failed with err")

	ctx.Eventually(ctx.GetValueStateClb(teib1)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"TEIB 1 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(teib2)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"TEIB 2 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(saOut1)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"OUT SA 1 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(saIn1)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"IN SA 1 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(saOut2)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"OUT SA 2 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(saIn2)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"IN SA 2 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(tp2)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"tunnel protection 2 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(tp1)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"tunnel protection 1 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(ipipTun)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"IPIP tunnel was not removed")
	ctx.Eventually(ctx.GetValueStateClb(spOut1)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"OUT SP 1 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(spIn1)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"IN SP 1 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(spOut2)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"OUT SP 2 was not removed")
	ctx.Eventually(ctx.GetValueStateClb(spIn2)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"IN SP 2 was not removed")

	if ctx.VppRelease() >= "20.05" {
		ctx.Expect(ctx.AgentInSync()).To(BeTrue())
	}
}
