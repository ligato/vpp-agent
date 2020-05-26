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
)

func TestIPSec(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	if ctx.vppRelease <= "19.04" {
		t.Skipf("IPIP: skipped for VPP <= 19.04 (%s)", ctx.vppVersion)
	}

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
		CryptoAlg:      vpp_ipsec.CryptoAlg_AES_CBC_128,
		CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
		IntegAlg:       vpp_ipsec.IntegAlg_SHA1_96,
		IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
		EnableUdpEncap: true,
	}
	saIn := &vpp_ipsec.SecurityAssociation{
		Index:          20,
		Spi:            456,
		Protocol:       vpp_ipsec.SecurityAssociation_ESP,
		CryptoAlg:      vpp_ipsec.CryptoAlg_AES_CBC_128,
		CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
		IntegAlg:       vpp_ipsec.IntegAlg_SHA1_96,
		IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
		EnableUdpEncap: true,
	}
	tp := &vpp_ipsec.TunnelProtection{
		Interface: tunnelIfName,
		SaOut:     []uint32{saOut.Index},
		SaIn:      []uint32{saIn.Index},
	}

	ctx.startMicroservice(msName)
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		ipipTun,
		saOut,
		saIn,
		tp,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Sending change request failed with err")

	Eventually(ctx.getValueStateClb(ipipTun)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IPIP tunnel is not configured")
	Eventually(ctx.getValueStateClb(saOut)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"OUT SA is not configured")
	Eventually(ctx.getValueStateClb(saIn)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IN SA is not configured")
	Eventually(ctx.getValueStateClb(tp)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"tunnel protection is not configured")

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

	req2 := ctx.grpcClient.ChangeRequest()
	err = req2.
		Delete(
			saOut,
			saIn).
		Update(
			saOutNew,
			saInNew,
			tpNew,
		).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Sending change request failed with err")

	Eventually(ctx.getValueStateClb(saOut)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"old OUT SA was not removed")
	Eventually(ctx.getValueStateClb(saIn)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"old IN SA was not removed")
	Eventually(ctx.getValueStateClb(saOutNew)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"OUT SA is not configured")
	Eventually(ctx.getValueStateClb(saInNew)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"IN SA is not configured")
	Eventually(ctx.getValueStateClb(tpNew)).Should(Equal(kvscheduler.ValueState_CONFIGURED),
		"tunnel protection is not configured")

	// delete the tunnel

	req3 := ctx.grpcClient.ChangeRequest()
	err = req3.Delete(
		saOutNew,
		saInNew,
		tpNew,
		ipipTun,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred(), "Sending change request failed with err")

	Eventually(ctx.getValueStateClb(saOutNew)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"OUT SA was not removed")
	Eventually(ctx.getValueStateClb(saInNew)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"IN SA was not removed")
	Eventually(ctx.getValueStateClb(tpNew)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"tunnel protection was not removed")
	Eventually(ctx.getValueStateClb(ipipTun)).Should(Equal(kvscheduler.ValueState_NONEXISTENT),
		"IPIP tunnel was not removed")
}
