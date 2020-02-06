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
	"math"
	"net"
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"

	netalloc_mock "go.ligato.io/vpp-agent/v3/plugins/netalloc/mock"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	l3plugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
)

func TestL3XC(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	if test.versionInfo.Release() <= "19.04" {
		t.Skipf("SKIP for VPP %s<=19.04", test.versionInfo.Release())
	}

	// Setup indexers
	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	vrfIndexes.Put("vrf1-ipv4", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV4})
	vrfIndexes.Put("vrf1-ipv6", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV6})

	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(test.vppClient, logrus.NewLogger("test"))

	// Create interfaces
	const iface1 = "loop1"
	ifIdx1, err := ih.AddLoopbackInterface(iface1)
	if err != nil {
		t.Fatalf("creating interface failed: %v", err)
	}
	ifIndexes.Put(iface1, &ifaceidx.IfaceMetadata{SwIfIndex: ifIdx1})
	ipNet1 := net.IPNet{
		IP:   net.IPv4(10, 10, 0, 1),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	if err := ih.AddInterfaceIP(ifIdx1, &ipNet1); err != nil {
		t.Fatalf("adding interface IP failed: %v", err)
	}
	if err := ih.InterfaceAdminUp(test.Context, ifIdx1); err != nil {
		t.Fatalf("setting interface admin up failed: %v", err)
	}

	const iface2 = "loop2"
	ifIdx2, err := ih.AddLoopbackInterface(iface2)
	if err != nil {
		t.Fatalf("creating interface failed: %v", err)
	}
	ifIndexes.Put(iface2, &ifaceidx.IfaceMetadata{SwIfIndex: ifIdx2})
	ipNet2 := net.IPNet{
		IP:   net.IPv4(10, 20, 0, 1),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	if err := ih.AddInterfaceIP(ifIdx2, &ipNet2); err != nil {
		t.Fatalf("adding interface IP failed: %v", err)
	}
	if err := ih.InterfaceAdminUp(test.Context, ifIdx2); err != nil {
		t.Fatalf("setting interface admin up failed: %v", err)
	}

	// Run test cases
	l3Handler := l3plugin_vppcalls.CompatibleL3VppHandler(test.vppClient, ifIndexes, vrfIndexes,
		netalloc_mock.NewMockNetAlloc(), logrus.NewLogger("test"))

	l3xcs, err := l3Handler.DumpL3XC(test.Context, math.MaxUint32)
	if err != nil {
		t.Fatalf("dumping l3xcs failed: %v", err)
	} else if len(l3xcs) != 0 {
		t.Fatalf("expected empty dump, but got: %+v", l3xcs)
	}

	err = l3Handler.UpdateL3XC(test.Context, &l3plugin_vppcalls.L3XC{
		SwIfIndex: ifIdx1,
		IsIPv6:    false,
		Paths: []l3plugin_vppcalls.Path{
			{
				SwIfIndex: ifIdx2,
				NextHop:   ipNet2.IP,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error on updating l3xcs: %v", err)
	}

	l3xcs, err = l3Handler.DumpL3XC(test.Context, math.MaxUint32)
	if err != nil {
		t.Fatalf("dumping l3xcs failed: %v", err)
	} else if n := len(l3xcs); n != 1 {
		t.Fatalf("expected 1 l3xc, but got %d", n)
	}

	if l3xcs[0].SwIfIndex != ifIdx1 {
		t.Fatalf("expected SwIfIndex to be %d, but got %d", ifIdx1, l3xcs[0].SwIfIndex)
	} else if l3xcs[0].IsIPv6 != false {
		t.Fatalf("expected IsIPv6 to be false")
	} else if len(l3xcs[0].Paths) != 1 {
		t.Fatalf("expected 1 path, but got %d", len(l3xcs[0].Paths))
	}
	path := l3xcs[0].Paths[0]
	if path.SwIfIndex != ifIdx2 {
		t.Fatalf("expected path SwIfIndex to be %d, but got %d", ifIdx2, path.SwIfIndex)
	} else if !path.NextHop.Equal(ipNet2.IP) {
		t.Fatalf("expected path Nh to be %v, but got %v", ipNet2.IP, path.NextHop)
	}

	err = l3Handler.DeleteL3XC(test.Context, ifIdx1, false)
	if err != nil {
		t.Fatalf("deleting l3xc failed: %v", err)
	}

	l3xcs, err = l3Handler.DumpL3XC(test.Context, math.MaxUint32)
	if err != nil {
		t.Fatalf("dumping l3xcs failed: %v", err)
	} else if len(l3xcs) != 0 {
		t.Fatalf("expected empty dump, but got: %+v", l3xcs)
	}
}
