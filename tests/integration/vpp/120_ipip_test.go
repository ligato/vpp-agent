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

	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

func TestIPIP(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	dumpAPIOk := true
	release := ctx.versionInfo.Release()
	if release < "20.01" {
		t.Logf("IPIP: skipped for VPP < 20.01 (%s)", release)
		return
	}

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))

	tests := []struct {
		name       string
		ipip       *interfaces.IPIPLink
		shouldFail bool
	}{
		{
			name: "Create IPIP tunnel (IPv4)",
			ipip: &interfaces.IPIPLink{
				SrcAddr: "20.30.40.50",
				DstAddr: "50.40.30.20",
			},
			shouldFail: false,
		},
		{
			name: "Create IPIP tunnel (IPv6)",
			ipip: &interfaces.IPIPLink{
				SrcAddr: "2001:db8:0:1:1:1:1:1",
				DstAddr: "2002:db8:0:1:1:1:1:1",
			},
			shouldFail: false,
		},
		{
			name: "Create IPIP tunnel with same src and dst address",
			ipip: &interfaces.IPIPLink{
				SrcAddr: "20.30.40.50",
				DstAddr: "20.30.40.50",
			},
			shouldFail: true,
		},
		{
			name: "Create IPIP tunnel with missing src address",
			ipip: &interfaces.IPIPLink{
				DstAddr: "20.30.40.50",
			},
			shouldFail: true,
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ifName := fmt.Sprintf("ipip%d", i)
			ifIdx, err := h.AddIpipTunnel(ifName, 0, test.ipip)

			if err != nil {
				if test.shouldFail {
					return
				}
				t.Fatalf("create IPIP tunnel failed: %v\n", err)
			} else {
				if test.shouldFail {
					t.Fatal("create IPIP tunnel must fail, but it's not")
				}
			}

			if dumpAPIOk {
				ifaces, err := h.DumpInterfaces(ctx.Ctx)
				if err != nil {
					t.Fatalf("dumping interfaces failed: %v", err)
				}
				iface, ok := ifaces[ifIdx]
				if !ok {
					t.Fatalf("IPIP interface was not found in dump")
				}

				if iface.Interface.GetType() != interfaces.Interface_IPIP_TUNNEL {
					t.Fatalf("Interface is not an IPIP tunnel")
				}

				ipip := iface.Interface.GetIpip()
				if test.ipip.SrcAddr != ipip.SrcAddr {
					t.Fatalf("expected source address <%s>, got: <%s>", test.ipip.SrcAddr, ipip.SrcAddr)
				}
				if test.ipip.DstAddr != ipip.DstAddr {
					t.Fatalf("expected destination address <%s>, got: <%s>", test.ipip.DstAddr, ipip.DstAddr)
				}
			} else {
				t.Logf("IPIP: DumpInterfaces skipped because of a broken API in VPP %s", ctx.versionInfo.Version)
			}

			err = h.DelIpipTunnel(ifName, ifIdx)
			if err != nil {
				t.Fatalf("delete IPIP tunnel failed: %v\n", err)
			}

			if dumpAPIOk {
				ifaces, err := h.DumpInterfaces(ctx.Ctx)
				if err != nil {
					t.Fatalf("dumping interfaces failed: %v", err)
				}

				if _, ok := ifaces[ifIdx]; ok {
					t.Fatalf("IPIP interface was found in dump after removing")
				}
			}
		})
	}
}
