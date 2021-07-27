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
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

func TestIP6ND(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(test.vppClient, logrus.NewLogger("test"))
	const ifName = "loop1"
	ifIdx, err := ih.AddLoopbackInterface(ifName)
	if err != nil {
		t.Fatalf("creating interface failed: %v", err)
	}
	if err = ih.InterfaceAdminUp(test.Ctx, ifIdx); err != nil {
		t.Fatalf("setting interface up failed: %v", err)
	}
	t.Logf("interface created %v", ifIdx)

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{SwIfIndex: ifIdx})
	vrfIndexes := vrfidx.NewVRFIndex(logrus.NewLogger("test-vrf"), "test-vrf")
	vrfIndexes.Put("vrf1-ipv4", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV4})
	vrfIndexes.Put("vrf1-ipv6", &vrfidx.VRFMetadata{Index: 0, Protocol: vpp_l3.VrfTable_IPV6})

	err = ih.SetIP6ndAutoconfig(test.Ctx, ifIdx, true, true)
	Expect(err).To(Succeed())

	out, err := test.vpp.RunCli(test.Ctx, "show ip6 fib")
	Expect(err).To(Succeed())
	Expect(out).To(ContainSubstring("default-route:1"))
	// TODO: fix test assert

	t.Logf("CLI output: %s", out)
}
