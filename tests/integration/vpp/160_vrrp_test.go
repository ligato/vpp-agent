// Copyright (c) 2020 Pantheon.tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
)

func TestVrrp(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	release := ctx.versionInfo.Release()
	if release < "20.05" {
		t.Skipf("VRRP: skipped for VPP < 20.05 (%s)", release)
	}

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrrp"), "test-vrrp")
	l3Handler := l3plugin_vppcalls.CompatibleL3VppHandler(ctx.vppClient, ifIndexes, vrfIndexes,
		netalloc_mock.NewMockNetAlloc(), logrus.NewLogger("test-l3"))

	ifHandler := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test-if"))

	tests := []struct {
		name       string
		vrrp       *l3.VRRPEntry
		shouldFail bool
	}{
		{
			name: "Create VRRP entry (IPv4)",
			vrrp: &l3.VRRPEntry{
				Interface:   "if0",
				VrId:        1,
				Priority:    100,
				Interval:    150,
				PreemtpFlag: false,
				AcceptFlag:  false,
				UnicastFlag: false,
				Ipv6Flag:    false,
				Addrs:       []string{"192.168.10.21"},
				Enabled:     true,
			},
			shouldFail: false,
		},
		{
			name: "Create VRRP entry (IPv4)",
			vrrp: &l3.VRRPEntry{
				Interface:   "if1",
				VrId:        1,
				Priority:    200,
				Interval:    150,
				PreemtpFlag: false,
				AcceptFlag:  false,
				UnicastFlag: false,
				Ipv6Flag:    false,
				Addrs:       []string{"192.168.10.22"},
				Enabled:     true,
			},
			shouldFail: false,
		},
		{
			name: "Create VRRP entry (IPv6)",
			vrrp: &l3.VRRPEntry{
				Interface:   "if2",
				VrId:        33,
				Priority:    200,
				Interval:    200,
				PreemtpFlag: false,
				AcceptFlag:  false,
				UnicastFlag: false,
				Ipv6Flag:    true,
				Addrs:       []string{"2001:0db8:11a3:09d7:1f34:8a2e:07a0:765d"},
				Enabled:     true,
			},
			shouldFail: false,
		},
		{
			name: "Create TEIB entry (IPv6)",
			vrrp: &l3.VRRPEntry{
				Interface:   "if3",
				VrId:        33,
				Priority:    100,
				Interval:    0,
				PreemtpFlag: false,
				AcceptFlag:  false,
				UnicastFlag: false,
				Ipv6Flag:    true,
				Addrs:       []string{"2001:0db8:11a3:09d7:1f34:8a2e:07a0:765d"},
				Enabled:     true,
			},
			shouldFail: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ifIdx, err := ifHandler.AddLoopbackInterface(test.vrrp.Interface)
			if err != nil {
				t.Fatalf("creating interface failed: %v", err)
			}
			ifIndexes.Put(test.vrrp.Interface, &ifaceidx.IfaceMetadata{SwIfIndex: ifIdx})

			err = l3Handler.VppAddVrrp(test.vrrp)
			if err != nil {
				if test.shouldFail {
					return
				}
				t.Fatalf("create VRRP entry failed: %v\n", err)
			} else {
				if test.shouldFail {
					t.Fatal("create VRRP entry must fail, but it's not")
				}
			}

			entries, err := l3Handler.DumpVrrpEntries()
			if err != nil {
				t.Fatalf("dump VRRP entries failed: %v\n", err)
			}
			if len(entries) == 0 {
				t.Fatalf("no VRRP entries dumped")
			}

			if entries[0].Vrrp.VrId != test.vrrp.VrId {
				t.Fatalf("expected VrId <%v>, got: <%v>", test.vrrp.VrId, entries[0].Vrrp.VrId)
			}
			if entries[0].Vrrp.Interface != test.vrrp.Interface {
				t.Fatalf("expected Interface <%v>, got: <%v>", test.vrrp.Interface, entries[0].Vrrp.Interface)
			}
			if entries[0].Vrrp.Interval != test.vrrp.Interval {
				t.Fatalf("expected Interval <%v>, got: <%v>", test.vrrp.Interval, entries[0].Vrrp.Interval)
			}
			if entries[0].Vrrp.Priority != test.vrrp.Priority {
				t.Fatalf("expected Priority <%v>, got: <%v>", test.vrrp.Priority, entries[0].Vrrp.Priority)
			}
			if entries[0].Vrrp.Enabled != test.vrrp.Enabled {
				t.Fatalf("expected Enabled <%v>, got: <%v>", test.vrrp.Enabled, entries[0].Vrrp.Enabled)
			}
			if entries[0].Vrrp.Ipv6Flag != test.vrrp.Ipv6Flag {
				t.Fatalf("expected Ipv6Flag <%v>, got: <%v>", test.vrrp.Ipv6Flag, entries[0].Vrrp.Ipv6Flag)
			}
			if entries[0].Vrrp.PreemtpFlag != test.vrrp.PreemtpFlag {
				t.Fatalf("expected PreemtpFlag <%v>, got: <%v>", test.vrrp.PreemtpFlag, entries[0].Vrrp.PreemtpFlag)
			}
			if entries[0].Vrrp.UnicastFlag != test.vrrp.UnicastFlag {
				t.Fatalf("expected UnicastFlag <%v>, got: <%v>", test.vrrp.UnicastFlag, entries[0].Vrrp.UnicastFlag)
			}
			if entries[0].Vrrp.AcceptFlag != test.vrrp.AcceptFlag {
				t.Fatalf("expected AcceptFlag <%v>, got: <%v>", test.vrrp.AcceptFlag, entries[0].Vrrp.AcceptFlag)
			}

			for i := 0; i < len(test.vrrp.Addrs); i++ {
				if entries[0].Vrrp.Addrs[i] != test.vrrp.Addrs[i] {
					t.Fatalf("expected Addrs[%v]  <%v>, got: <%v>", i, test.vrrp.Addrs[i], entries[0].Vrrp.Addrs[i])
				}
			}

			err = l3Handler.VppDelVrrp(test.vrrp)
			if err != nil {
				t.Fatalf("delete VRRP entry failed: %v\n", err)
			}

			entries, err = l3Handler.DumpVrrpEntries()
			if err != nil {
				t.Fatalf("dump VRRP entries failed: %v\n", err)
			}
			if len(entries) != 0 {
				t.Fatalf("%d VRRP entries dumped after delete", len(entries))
			}

			err = ifHandler.DeleteLoopbackInterface(test.vrrp.Interface, ifIdx)
			if err != nil {
				t.Fatalf("delete interface failed: %v", err)
			}
			ifIndexes.Delete(test.vrrp.Interface)
		})
	}
}
