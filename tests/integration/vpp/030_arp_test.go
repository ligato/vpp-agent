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

package vpp

import (
	"strings"
	"testing"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	netalloc_mock "go.ligato.io/vpp-agent/v3/plugins/netalloc/mock"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	l3plugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
)

func TestArp(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))
	const ifName = "loop1"
	ifIdx, err := ih.AddLoopbackInterface(ifName)
	if err != nil {
		t.Fatalf("creating interface failed: %v", err)
	}
	t.Logf("interface created %v", ifIdx)

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{SwIfIndex: ifIdx})
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	vrfIndexes.Put("vrf1-ipv4", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV4})
	vrfIndexes.Put("vrf1-ipv6", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV6})

	h := l3plugin_vppcalls.CompatibleL3VppHandler(ctx.vppClient, ifIndexes, vrfIndexes,
		netalloc_mock.NewMockNetAlloc(), logrus.NewLogger("test"))

	tests := []struct {
		name        string
		newArpEntry vpp_l3.ARPEntry
	}{
		{"static arp for ipv4", vpp_l3.ARPEntry{
			Interface:   ifName,
			IpAddress:   "192.168.10.21",
			PhysAddress: "59:6C:45:59:8E:BD",
			Static:      true,
		}},
		{"nonstatic arp for ipv4", vpp_l3.ARPEntry{
			Interface:   ifName,
			IpAddress:   "192.168.10.22",
			PhysAddress: "6C:45:59:59:8E:BD",
			Static:      false,
		}},
		{"nonstatic arp for ipv6", vpp_l3.ARPEntry{
			Interface:   ifName,
			IpAddress:   "dead::1",
			PhysAddress: "8E:BD:6C:45:59:59",
			Static:      false,
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			arpentries, err := h.DumpArpEntries()
			if err != nil {
				t.Fatalf("dumping arpentries failed: %v", err)
			}
			arpentriescnt := len(arpentries)
			t.Logf("%d arpentries dumped", arpentriescnt)

			err = h.VppAddArp(&test.newArpEntry)
			if err != nil {
				t.Fatalf("adding arpentry failed: %v", err)
			}
			t.Logf("arpentry added %+v", test.newArpEntry)

			arpentries, err = h.DumpArpEntries()
			if err != nil {
				t.Fatalf("dumping arpentries failed: %v", err)
			}
			arpentriescnt2 := len(arpentries)
			t.Logf("%d arpentries dumped", arpentriescnt2)

			if arpentriescnt+1 != arpentriescnt2 {
				t.Errorf("Number of arp entries after adding of one arp entry is not incremented by 1")
			}

			newArpEntryIsPresent := false
			for _, arpentry := range arpentries {
				if (arpentry.Arp.Interface == test.newArpEntry.Interface) && (arpentry.Arp.IpAddress == test.newArpEntry.IpAddress) && (strings.ToLower(arpentry.Arp.PhysAddress) == strings.ToLower(test.newArpEntry.PhysAddress)) {
					t.Logf("dumped arpentry %+v", arpentry)
					newArpEntryIsPresent = true
					break
				}
			}

			if !newArpEntryIsPresent {
				t.Error("Added arp entry is not present in arp dump")
			}

			err = h.VppDelArp(&test.newArpEntry)
			if err != nil {
				t.Fatalf("deleting arpentry failed: %v", err)
			}
			t.Logf("arpentry deleted")

			arpentries, err = h.DumpArpEntries()
			if err != nil {
				t.Fatalf("dumping arpentries failed: %v", err)
			}
			arpentriescnt3 := len(arpentries)
			t.Logf("%d arpentries dumped", arpentriescnt3)

			if arpentriescnt2-1 != arpentriescnt3 {
				t.Errorf("Number of arp entries after deleting of one arp entry is not decremented by 1")
			}

			for _, arpentry := range arpentries {
				if (arpentry.Arp.Interface == test.newArpEntry.Interface) && (arpentry.Arp.IpAddress == test.newArpEntry.IpAddress) && (strings.ToLower(arpentry.Arp.PhysAddress) == strings.ToLower(test.newArpEntry.PhysAddress)) {
					t.Error("Added arp entry is still present in arp dump - should be deleted")
				}
			}
		})
	}
}
