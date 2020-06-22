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

package vpp

import (
	"fmt"
	"testing"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	ipsec_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin"
)

func TestIPSec(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	p2mpSupported := true // determines point-to-multipoint support
	saDumpAPIOk := true   // determines if SA dump API is working

	release := ctx.versionInfo.Release()
	if release < "20.01" {
		t.Skipf("IPSec: skipped for VPP < 20.01 (%s)", release)
	}
	if release < "20.05" {
		saDumpAPIOk = false   // tunnel protection SA dump broken in VPP 20.01
		p2mpSupported = false // point-to-multipoint support comes in VPP 20.05
	}

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test"), "test-ifidx")
	ifHandler := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test-if"))
	ipsecHandler := ipsec_vppcalls.CompatibleIPSecVppHandler(ctx.vppClient, ifIndexes, logrus.NewLogger("test-ipsec"))

	tests := []struct {
		name  string
		ipip  *interfaces.IPIPLink
		saOut *ipsec.SecurityAssociation
		saIn  *ipsec.SecurityAssociation
		tp    *ipsec.TunnelProtection
	}{
		{
			name: "Create IPSec tunnel (IPv4)",
			ipip: &interfaces.IPIPLink{
				SrcAddr: "20.30.40.50",
				DstAddr: "50.40.30.20",
			},
			saOut: &ipsec.SecurityAssociation{
				Index:          10,
				Spi:            123,
				Protocol:       ipsec.SecurityAssociation_ESP,
				CryptoAlg:      ipsec.CryptoAlg_AES_CBC_128,
				CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
				IntegAlg:       ipsec.IntegAlg_SHA1_96,
				IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
				EnableUdpEncap: true,
			},
			saIn: &ipsec.SecurityAssociation{
				Index:          20,
				Spi:            456,
				Protocol:       ipsec.SecurityAssociation_ESP,
				CryptoAlg:      ipsec.CryptoAlg_AES_CBC_128,
				CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
				IntegAlg:       ipsec.IntegAlg_SHA1_96,
				IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
				EnableUdpEncap: true,
			},
			tp: &ipsec.TunnelProtection{
				SaOut: []uint32{10},
				SaIn:  []uint32{20},
			},
		},
		{
			name: "Create IPSec tunnel (IPv6)",
			ipip: &interfaces.IPIPLink{
				SrcAddr: "2001:db8:0:1:1:1:1:1",
				DstAddr: "2002:db8:0:1:1:1:1:1",
			},
			saOut: &ipsec.SecurityAssociation{
				Index:     1,
				Spi:       789,
				Protocol:  ipsec.SecurityAssociation_ESP,
				CryptoAlg: ipsec.CryptoAlg_AES_CBC_128,
				CryptoKey: "d9a4ec50aed76f1bf80bc915d8fcfe1c",
				IntegAlg:  ipsec.IntegAlg_SHA1_96,
				IntegKey:  "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
			},
			saIn: &ipsec.SecurityAssociation{
				Index:     2,
				Spi:       321,
				Protocol:  ipsec.SecurityAssociation_ESP,
				CryptoAlg: ipsec.CryptoAlg_AES_CBC_128,
				CryptoKey: "d9a4ec50aed76f1bf80bc915d8fcfe1c",
				IntegAlg:  ipsec.IntegAlg_SHA1_96,
				IntegKey:  "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
			},
			tp: &ipsec.TunnelProtection{
				SaOut: []uint32{1},
				SaIn:  []uint32{2},
			},
		},
		{
			name: "Create multipoint IPSec tunnel",
			ipip: &interfaces.IPIPLink{
				SrcAddr:    "20.30.40.50",
				TunnelMode: interfaces.IPIPLink_POINT_TO_MULTIPOINT,
			},
			saOut: &ipsec.SecurityAssociation{
				Index:          100,
				Spi:            123,
				Protocol:       ipsec.SecurityAssociation_ESP,
				CryptoAlg:      ipsec.CryptoAlg_AES_CBC_128,
				CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
				IntegAlg:       ipsec.IntegAlg_SHA1_96,
				IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
				EnableUdpEncap: true,
			},
			saIn: &ipsec.SecurityAssociation{
				Index:          101,
				Spi:            456,
				Protocol:       ipsec.SecurityAssociation_ESP,
				CryptoAlg:      ipsec.CryptoAlg_AES_CBC_128,
				CryptoKey:      "d9a4ec50aed76f1bf80bc915d8fcfe1c",
				IntegAlg:       ipsec.IntegAlg_SHA1_96,
				IntegKey:       "bf9b150aaf5c2a87d79898b11eabd055e70abdbe",
				EnableUdpEncap: true,
			},
			tp: &ipsec.TunnelProtection{
				SaOut:       []uint32{100},
				SaIn:        []uint32{101},
				NextHopAddr: "4.5.6.7",
			},
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !p2mpSupported && test.ipip.TunnelMode == interfaces.IPIPLink_POINT_TO_MULTIPOINT {
				t.Skipf("IPIP: p2mp skipped for VPP < 20.05 (%s)", release)
			}
			// create IPIP tunnel + SAs + tunnel protection
			ifName := fmt.Sprintf("ipip%d", i)
			ifIdx, err := ifHandler.AddIpipTunnel(ifName, 0, test.ipip)
			if err != nil {
				t.Fatalf("IPIP tunnel add failed: %v", err)
			}
			ifIndexes.Clear()
			ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{
				SwIfIndex: ifIdx,
			})
			err = ipsecHandler.AddSA(test.saOut)
			if err != nil {
				t.Fatalf("IPSec SA add failed: %v", err)
			}
			err = ipsecHandler.AddSA(test.saIn)
			if err != nil {
				t.Fatalf("IPSec SA add failed: %v", err)
			}
			test.tp.Interface = ifName
			err = ipsecHandler.AddTunnelProtection(test.tp)
			if err != nil {
				t.Fatalf("add tunnel protection failed: %v\n", err)
			}

			// check created SAs + tunnel protection
			saList, err := ipsecHandler.DumpIPSecSA()
			if err != nil {
				t.Fatalf("dumping SAs failed: %v", err)
			}
			if len(saList) != 2 {
				t.Fatalf("Invalid number of SAs: %d", len(saList))
			}
			for _, sa := range saList {
				if sa.Sa.Index != test.saOut.Index && sa.Sa.Index != test.saIn.Index {
					t.Fatalf("Invalid SA index: %d", sa.Sa.Index)
				}
			}
			tpList, err := ipsecHandler.DumpTunnelProtections()
			if err != nil {
				t.Fatalf("dumping tunnel protections failed: %v", err)
			}
			if len(tpList) != 1 {
				t.Fatalf("Invalid number of tunnel protections: %d", len(tpList))
			}
			if tpList[0].Interface != ifName {
				t.Fatalf("Invalid interface name in tunnel protections: %s", tpList[0].Interface)
			}
			if saDumpAPIOk {
				if tpList[0].SaIn[0] != test.saIn.Index || tpList[0].SaOut[0] != test.saOut.Index {
					t.Fatalf("tunnel protection SA mismatch (%d != %d || %d != %d)",
						tpList[0].SaIn[0], test.saIn.Index, tpList[0].SaOut[0], test.saOut.Index)
				}
			} else {
				t.Logf("IPIP: SA index check skipped because of a broken API in VPP %s", ctx.versionInfo.Version)
			}
			if tpList[0].NextHopAddr != test.tp.NextHopAddr {
				t.Fatalf("tunnel protection next hop mismatch (%v != %v)", tpList[0].NextHopAddr, test.tp.NextHopAddr)
			}

			// delete tunnel protection, SAs and IPIP tunnel
			err = ipsecHandler.DeleteTunnelProtection(test.tp)
			if err != nil {
				t.Fatalf("delete tunnel protection failed: %v\n", err)
			}
			tpList, err = ipsecHandler.DumpTunnelProtections()
			if err != nil {
				t.Fatalf("dumping tunnel protections failed: %v", err)
			}
			if len(tpList) != 0 {
				t.Fatalf("%d tunnel protections found in dump after removing", len(tpList))
			}
			err = ipsecHandler.DeleteSA(test.saOut)
			if err != nil {
				t.Fatalf("delete SA failed: %v\n", err)
			}
			err = ipsecHandler.DeleteSA(test.saIn)
			if err != nil {
				t.Fatalf("delete SA failed: %v\n", err)
			}
			saList, err = ipsecHandler.DumpIPSecSA()
			if err != nil {
				t.Fatalf("dumping SAs failed: %v", err)
			}
			if len(saList) != 0 {
				t.Fatalf("%d SAs found in dump after removing", len(saList))
			}
			err = ifHandler.DelIpipTunnel(ifName, ifIdx)
			if err != nil {
				t.Fatalf("delete IPIP tunnel failed: %v\n", err)
			}
			ifaces, err := ifHandler.DumpInterfaces(ctx.Ctx)
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}
			if _, ok := ifaces[ifIdx]; ok {
				t.Fatalf("IPIP interface was found in dump after removing")
			}
		})
	}
}
