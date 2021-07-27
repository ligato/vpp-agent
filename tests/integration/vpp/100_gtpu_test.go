//  Copyright (c) 2021 EMnify
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

	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

func TestGtpu(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))

	tests := []struct {
		name           string
		gtpu           *interfaces.GtpuLink
		mcastSwIfIndex uint32
		isFail         bool
	}{
		{
			name: "Create GTP-U tunnel (IP4)",
			gtpu: &interfaces.GtpuLink{
				SrcAddr:    "20.30.40.50",
				DstAddr:    "50.40.30.20",
				Teid:       101,
				EncapVrfId: 0,
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         false,
		},
		{
			name: "Create GTP-U tunnel (IP6)",
			gtpu: &interfaces.GtpuLink{
				SrcAddr:    "2001:db8:0:1:1:1:1:1",
				DstAddr:    "2002:db8:0:1:1:1:1:1",
				Teid:       102,
				EncapVrfId: 0,
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         false,
		},
		{
			name: "Create GTP-U tunnel (DecapNext: L2)",
			gtpu: &interfaces.GtpuLink{
				SrcAddr:    "20.30.40.50",
				DstAddr:    "50.40.30.20",
				Teid:       201,
				EncapVrfId: 0,
				//DecapNext:     interfaces.GtpuLink_L2,
				DecapNextNode: 1,
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         false,
		},
		{
			name: "Create GTP-U tunnel (DecapNext: IP4)",
			gtpu: &interfaces.GtpuLink{
				SrcAddr:    "20.30.40.50",
				DstAddr:    "50.40.30.20",
				Teid:       202,
				EncapVrfId: 0,
				//DecapNext:     interfaces.GtpuLink_IP4,
				DecapNextNode: 2,
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         false,
		},
		{
			name: "Create GTP-U tunnel (DecapNext: IP6)",
			gtpu: &interfaces.GtpuLink{
				SrcAddr:    "2001:db8:0:1:1:1:1:1",
				DstAddr:    "2002:db8:0:1:1:1:1:1",
				Teid:       203,
				EncapVrfId: 0,
				//DecapNext:     interfaces.GtpuLink_IP6,
				DecapNextNode: 3,
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         false,
		},
		{
			name: "Create GTP-U tunnel with same source and destination",
			gtpu: &interfaces.GtpuLink{
				SrcAddr:    "20.30.40.50",
				DstAddr:    "20.30.40.50",
				Teid:       301,
				EncapVrfId: 0,
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         true,
		},
		{
			name: "Create GTP-U tunnel with src and dst ip versions mismatch",
			gtpu: &interfaces.GtpuLink{
				SrcAddr:    "20.30.40.50",
				DstAddr:    "::1",
				Teid:       302,
				EncapVrfId: 0,
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         true,
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ifName := fmt.Sprintf("test%d", i)
			ifIdx, err := h.AddGtpuTunnel(ifName, test.gtpu, test.mcastSwIfIndex)
			if err != nil {
				if test.isFail {
					return
				}
				t.Fatalf("create GTP-U tunnel failed: %v\n", err)
			} else {
				if test.isFail {
					t.Fatal("create GTP-U tunnel must fail, but it's not")
				}
			}

			ifaces, err := h.DumpInterfaces(ctx.Ctx)
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}
			iface, ok := ifaces[ifIdx]
			if !ok {
				t.Fatalf("GTP-U interface was not found in dump")
			}

			if iface.Interface.GetType() != interfaces.Interface_GTPU_TUNNEL {
				t.Fatalf("Interface is not a GTPU tunnel")
			}

			gtpu := iface.Interface.GetGtpu()
			if test.gtpu.SrcAddr != gtpu.SrcAddr {
				t.Fatalf("expected source address <%s>, got: <%s>", test.gtpu.SrcAddr, gtpu.SrcAddr)
			}
			if test.gtpu.DstAddr != gtpu.DstAddr {
				t.Fatalf("expected destination address <%s>, got: <%s>", test.gtpu.DstAddr, gtpu.DstAddr)
			}
			if test.gtpu.Teid != gtpu.Teid {
				t.Fatalf("expected TEID <%d>, got: <%d>", test.gtpu.Teid, gtpu.Teid)
			}
			if test.gtpu.Multicast != gtpu.Multicast {
				t.Fatalf("expected multicast interface name <%s>, got: <%s>", test.gtpu.Multicast, gtpu.Multicast)
			}
			if test.gtpu.EncapVrfId != gtpu.EncapVrfId {
				t.Fatalf("expected GTP-U EncapVrfId <%d>, got: <%d>", test.gtpu.EncapVrfId, gtpu.EncapVrfId)
			}
			testDecapNext := test.gtpu.DecapNextNode
			if testDecapNext == uint32(interfaces.GtpuLink_DEFAULT) {
				testDecapNext = uint32(interfaces.GtpuLink_L2)
			}
			if testDecapNext != gtpu.DecapNextNode {
				t.Fatalf("expected GTP-U DecapNextNode <%v>, got: <%v>", testDecapNext, gtpu.DecapNextNode)
			}

			err = h.DelGtpuTunnel(ifName, test.gtpu)
			if err != nil {
				t.Fatalf("delete GTP-U tunnel failed: %v\n", err)
			}

			ifaces, err = h.DumpInterfaces(ctx.Ctx)
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}

			if _, ok := ifaces[ifIdx]; ok {
				t.Fatalf("GTP-U interface was found in dump after removing")
			}
		})
	}
}
