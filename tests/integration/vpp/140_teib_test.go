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
	"testing"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	netalloc_mock "go.ligato.io/vpp-agent/v3/plugins/netalloc/mock"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	l3plugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

func TestTeib(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	release := ctx.versionInfo.Release()
	if release < "20.05" {
		t.Skipf("TEIB: skipped for VPP < 20.05 (%s)", release)
	}

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	l3Handler := l3plugin_vppcalls.CompatibleL3VppHandler(ctx.vppClient, ifIndexes, vrfIndexes,
		netalloc_mock.NewMockNetAlloc(), logrus.NewLogger("test-l3"))

	ifHandler := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test-if"))

	tests := []struct {
		name       string
		teib       *l3.TeibEntry
		shouldFail bool
	}{
		{
			name: "Create TEIB entry (IPv4)",
			teib: &l3.TeibEntry{
				Interface:   "if0",
				PeerAddr:    "20.30.40.50",
				NextHopAddr: "50.40.30.20",
			},
			shouldFail: false,
		},
		{
			name: "Create TEIB entry (IPv6)",
			teib: &l3.TeibEntry{
				Interface:   "if1",
				PeerAddr:    "2001:db8:0:1:1:1:1:1",
				NextHopAddr: "2002:db8:0:1:1:1:1:1",
			},
			shouldFail: false,
		},
		{
			name: "Create TEIB entry with no peer IP",
			teib: &l3.TeibEntry{
				Interface:   "if2",
				NextHopAddr: "50.40.30.20",
			},
			shouldFail: true,
		},
		{
			name: "Create TEIB entry with no next hop IP",
			teib: &l3.TeibEntry{
				Interface: "if3",
				PeerAddr:  "20.30.40.50",
			},
			shouldFail: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ifIdx, err := ifHandler.AddLoopbackInterface(test.teib.Interface)
			if err != nil {
				t.Fatalf("creating interface failed: %v", err)
			}
			ifIndexes.Put(test.teib.Interface, &ifaceidx.IfaceMetadata{SwIfIndex: ifIdx})

			err = l3Handler.VppAddTeibEntry(nil, test.teib)
			if err != nil {
				if test.shouldFail {
					return
				}
				t.Fatalf("create TEIB entry failed: %v\n", err)
			} else {
				if test.shouldFail {
					t.Fatal("create TEIB entry must fail, but it's not")
				}
			}

			entries, err := l3Handler.DumpTeib()
			if err != nil {
				t.Fatalf("dump TEIB entries failed: %v\n", err)
			}
			if len(entries) == 0 {
				t.Fatalf("no TEIB entries dumped")
			}

			if entries[0].VrfId != test.teib.VrfId {
				t.Fatalf("expected VrfId <%v>, got: <%v>", test.teib.VrfId, entries[0].VrfId)
			}
			if entries[0].Interface != test.teib.Interface {
				t.Fatalf("expected Interface <%v>, got: <%v>", test.teib.Interface, entries[0].Interface)
			}
			if entries[0].PeerAddr != test.teib.PeerAddr {
				t.Fatalf("expected PeerAddr <%s>, got: <%s>", test.teib.PeerAddr, entries[0].PeerAddr)
			}
			if entries[0].NextHopAddr != test.teib.NextHopAddr {
				t.Fatalf("expected NextHopAddr <%s>, got: <%s>", test.teib.NextHopAddr, entries[0].NextHopAddr)
			}

			err = l3Handler.VppDelTeibEntry(nil, test.teib)
			if err != nil {
				t.Fatalf("delete TEIB entry failed: %v\n", err)
			}

			entries, err = l3Handler.DumpTeib()
			if err != nil {
				t.Fatalf("dump TEIB entries failed: %v\n", err)
			}
			if len(entries) != 0 {
				t.Fatalf("%d TEIB entries dumped after delete", len(entries))
			}

			err = ifHandler.DeleteLoopbackInterface(test.teib.Interface, ifIdx)
			if err != nil {
				t.Fatalf("delete interface failed: %v", err)
			}
			ifIndexes.Delete(test.teib.Interface)
		})
	}
}
