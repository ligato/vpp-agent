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

	"github.com/golang/protobuf/proto"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	vpe_vppcalls "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	netalloc_mock "go.ligato.io/vpp-agent/v3/plugins/netalloc/mock"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	l3plugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin"
)

func TestIPNeighbor(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	if test.versionInfo.Release() >= "20.01" {
		t.Skipf("SKIP for VPP %s>=20.01", test.versionInfo.Release())
	}

	vpp := vpe_vppcalls.CompatibleHandler(test.vppClient)

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	vrfIndexes.Put("vrf1-ipv4", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV4})
	vrfIndexes.Put("vrf1-ipv6", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV6})

	h := l3plugin_vppcalls.CompatibleL3VppHandler(
		test.vppClient, ifIndexes, vrfIndexes, netalloc_mock.NewMockNetAlloc(), logrus.NewLogger("test"),
	)

	cliShowConfig := func() {
		out, err := vpp.RunCli(test.Ctx, "show ip scan-neighbor")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("cli config:\n%v", out)
	}

	def := h.DefaultIPScanNeighbor()
	if def == nil {
		t.Fatal("default config is nil")
	}

	cliShowConfig()

	ipneigh, err := h.GetIPScanNeighbor()
	if err != nil {
		t.Fatal("getting ip neighbor config failed:", err)
	}
	t.Logf("dump config:\n%+v", proto.MarshalTextString(ipneigh))
	if ipneigh.Mode != vpp_l3.IPScanNeighbor_DISABLED {
		t.Fatal("expected Mode to be DISABLED")
	}

	if err := h.SetIPScanNeighbor(&vpp_l3.IPScanNeighbor{
		Mode:           vpp_l3.IPScanNeighbor_IPV4,
		MaxProcTime:    20,
		MaxUpdate:      0, // MaxUpdate will be set to 10
		ScanInterval:   1,
		ScanIntDelay:   1,
		StaleThreshold: 4,
	}); err != nil {
		t.Fatal(err)
	}

	cliShowConfig()

	ipneigh, err = h.GetIPScanNeighbor()
	if err != nil {
		t.Fatal("getting ip neighbor config failed:", err)
	}
	t.Logf("dump config:\n%+v", proto.MarshalTextString(ipneigh))
	if ipneigh.Mode != vpp_l3.IPScanNeighbor_IPV4 {
		t.Fatalf("expected Mode to be IPV4, got %v", ipneigh.Mode)
	}
	if ipneigh.MaxProcTime != 20 {
		t.Fatalf("expected MaxProcTime to be 20, got %v", ipneigh.MaxProcTime)
	}
	if ipneigh.MaxUpdate != 10 {
		t.Logf("expected MaxUpdate to be 10, got %v", ipneigh.MaxUpdate)
	}
	if ipneigh.ScanInterval != 1 {
		t.Fatalf("expected ScanInterval to be 1, got %v", ipneigh.ScanInterval)
	}
	if ipneigh.ScanIntDelay != 1 {
		t.Fatalf("expected ScanIntDelay to be 5, got %v", ipneigh.ScanIntDelay)
	}
	if ipneigh.StaleThreshold != 4 {
		t.Fatalf("expected ScanInterval to be 1, got %v", ipneigh.StaleThreshold)
	}
}
