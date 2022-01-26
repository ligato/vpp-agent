//  Copyright (c) 2021 Cisco and/or its affiliates.
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
	"google.golang.org/protobuf/proto"

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

	spdIfaceDumpOk := false // determines if ipsec_spd_interface_dump works correctly

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test"), "test-ifidx")
	ifHandler := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test-if"))
	ipsecHandler := ipsec_vppcalls.CompatibleIPSecVppHandler(ctx.vppClient, ifIndexes, logrus.NewLogger("test-ipsec"))

	tests := []struct {
		name  string
		ipip  *interfaces.IPIPLink
		saOut *ipsec.SecurityAssociation
		saIn  *ipsec.SecurityAssociation
		spd   *ipsec.SecurityPolicyDatabase
		spOut *ipsec.SecurityPolicy
		spIn  *ipsec.SecurityPolicy
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
			spd: &ipsec.SecurityPolicyDatabase{
				Index: 100,
			},
			spOut: &ipsec.SecurityPolicy{
				SpdIndex:        100,
				SaIndex:         10,
				Priority:        0,
				IsOutbound:      true,
				RemoteAddrStart: "10.10.1.1",
				RemoteAddrStop:  "10.10.1.255",
				LocalAddrStart:  "10.10.2.1",
				LocalAddrStop:   "10.10.2.255",
				Protocol:        0,
				Action:          ipsec.SecurityPolicy_PROTECT,
			},
			spIn: &ipsec.SecurityPolicy{
				SpdIndex:        100,
				SaIndex:         20,
				Priority:        0,
				IsOutbound:      false,
				RemoteAddrStart: "10.10.1.1",
				RemoteAddrStop:  "10.10.1.255",
				LocalAddrStart:  "10.10.2.1",
				LocalAddrStop:   "10.10.2.255",
				Protocol:        0,
				RemotePortStart: 1000,
				RemotePortStop:  5000,
				LocalPortStart:  2000,
				LocalPortStop:   7000,
				Action:          ipsec.SecurityPolicy_PROTECT,
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
			spd: &ipsec.SecurityPolicyDatabase{
				Index: 101,
			},
			spOut: &ipsec.SecurityPolicy{
				SpdIndex:        101,
				SaIndex:         1,
				Priority:        0,
				IsOutbound:      true,
				RemoteAddrStart: "2001:1000::1",
				RemoteAddrStop:  "2001:1000::1000",
				LocalAddrStart:  "2001:2000::1",
				LocalAddrStop:   "2001:2000::1000",
				Protocol:        0,
				Action:          ipsec.SecurityPolicy_PROTECT,
			},
			spIn: &ipsec.SecurityPolicy{
				SpdIndex:        101,
				SaIndex:         2,
				Priority:        0,
				IsOutbound:      false,
				RemoteAddrStart: "2001:1000::1",
				RemoteAddrStop:  "2001:1000::1000",
				LocalAddrStart:  "2001:2000::1",
				LocalAddrStop:   "2001:2000::1000",
				Protocol:        0,
				RemotePortStart: 1000,
				RemotePortStop:  5000,
				LocalPortStart:  2000,
				LocalPortStop:   7000,
				Action:          ipsec.SecurityPolicy_PROTECT,
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
			spd: &ipsec.SecurityPolicyDatabase{
				Index: 102,
			},
			spOut: &ipsec.SecurityPolicy{
				SpdIndex:        102,
				SaIndex:         100,
				Priority:        0,
				IsOutbound:      true,
				RemoteAddrStart: "10.10.1.1",
				RemoteAddrStop:  "10.10.1.255",
				LocalAddrStart:  "10.10.2.1",
				LocalAddrStop:   "10.10.2.255",
				Protocol:        0,
				Action:          ipsec.SecurityPolicy_PROTECT,
			},
			spIn: &ipsec.SecurityPolicy{
				SpdIndex:        102,
				SaIndex:         101,
				Priority:        0,
				IsOutbound:      false,
				RemoteAddrStart: "10.10.1.1",
				RemoteAddrStop:  "10.10.1.255",
				LocalAddrStart:  "10.10.2.1",
				LocalAddrStop:   "10.10.2.255",
				Protocol:        0,
				RemotePortStart: 1000,
				RemotePortStop:  5000,
				LocalPortStart:  2000,
				LocalPortStop:   7000,
				Action:          ipsec.SecurityPolicy_PROTECT,
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
			// create IPIP tunnel + SAs + tunnel protection + SPs
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
			err = ipsecHandler.AddSPD(test.spd.Index)
			if err != nil {
				t.Fatalf("add SPD failed: %v\n", err)
			}
			err = ipsecHandler.AddSPDInterface(test.spd.Index, &ipsec.SecurityPolicyDatabase_Interface{Name: ifName})
			if err != nil {
				t.Fatalf("add SPD-Interface failed: %v\n", err)
			}
			err = ipsecHandler.AddSP(test.spOut)
			if err != nil {
				t.Fatalf("add SP failed: %v\n", err)
			}
			err = ipsecHandler.AddSP(test.spIn)
			if err != nil {
				t.Fatalf("add SP failed: %v\n", err)
			}

			// check created SAs + tunnel protection + SPs
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
			if tpList[0].SaIn[0] != test.saIn.Index || tpList[0].SaOut[0] != test.saOut.Index {
				t.Fatalf("tunnel protection SA mismatch (%d != %d || %d != %d)",
					tpList[0].SaIn[0], test.saIn.Index, tpList[0].SaOut[0], test.saOut.Index)
			}
			if tpList[0].NextHopAddr != test.tp.NextHopAddr {
				t.Fatalf("tunnel protection next hop mismatch (%v != %v)", tpList[0].NextHopAddr, test.tp.NextHopAddr)
			}
			spdList, err := ipsecHandler.DumpIPSecSPD()
			if err != nil {
				t.Fatalf("dumping of SPDs failed: %v", err)
			}
			if len(spdList) != 1 {
				t.Fatalf("Invalid number of SPDs: %d", len(spdList))
			}
			if spdList[0].Index != test.spd.Index {
				t.Fatalf("Invalid SPD index: %d", spdList[0].Index)
			}
			if spdIfaceDumpOk {
				if len(spdList[0].Interfaces) != 1 {
					t.Fatalf("Invalid number of interfaces inside SPDs: %d", len(spdList[0].Interfaces))
				}
				if spdList[0].Interfaces[0].Name != ifName {
					t.Fatalf("Invalid interface name in tunnel protections: %s", spdList[0].Interfaces[0].Name)
				}
			}
			spList, err := ipsecHandler.DumpIPSecSP()
			if err != nil {
				t.Fatalf("dumping of SPs failed: %v", err)
			}
			if len(spList) != 2 {
				t.Fatalf("Invalid number of SPs: %d", len(spList))
			}
			for _, sp := range spList {
				if !proto.Equal(sp, test.spOut) && !proto.Equal(sp, test.spIn) {
					t.Fatalf("Invalid SP: %+v", sp)
				}
			}

			// delete SPs, tunnel protection, SAs and IPIP tunnel
			err = ipsecHandler.DeleteSP(test.spIn)
			if err != nil {
				t.Fatalf("delete of security policy failed: %v\n", err)
			}
			err = ipsecHandler.DeleteSP(test.spOut)
			if err != nil {
				t.Fatalf("delete of security policy failed: %v\n", err)
			}
			spList, err = ipsecHandler.DumpIPSecSP()
			if err != nil {
				t.Fatalf("dumping of security policies failed: %v", err)
			}
			if len(spList) != 0 {
				t.Fatalf("%d SPs found in dump after removing", len(spList))
			}
			err = ipsecHandler.DeleteSPDInterface(test.spd.Index, &ipsec.SecurityPolicyDatabase_Interface{Name: ifName})
			if err != nil {
				t.Fatalf("delete of SPD failed: %v\n", err)
			}
			err = ipsecHandler.DeleteSPD(test.spd.Index)
			if err != nil {
				t.Fatalf("delete of SPD failed: %v\n", err)
			}
			spdList, err = ipsecHandler.DumpIPSecSPD()
			if err != nil {
				t.Fatalf("dumping of SPDs failed: %v", err)
			}
			if len(spdList) != 0 {
				t.Fatalf("%d SPDs found in dump after removing", len(spdList))
			}
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
