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
		t.Skipf("TEIB: skipped for VPP < 20.05 (%s)", release)
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
			name: "",
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

		})
	}
}
