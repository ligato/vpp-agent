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

package vpp2106_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vrrp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls/vpp2106"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

var vrrpEntries = []*l3.VRRPEntry{
	{
		Interface:   "if1",
		VrId:        4,
		Priority:    100,
		Interval:    100,
		Preempt:     false,
		Accept:      false,
		Unicast:     false,
		IpAddresses: []string{"192.168.10.21"},
		Enabled:     true,
	},
	{
		Interface:   "if1",
		VrId:        4,
		Priority:    200,
		Interval:    300,
		Preempt:     false,
		Accept:      false,
		Unicast:     false,
		IpAddresses: []string{"192.168.10.22", "192.168.10.23"},
		Enabled:     false,
	},
	{
		Interface:   "if1",
		VrId:        6,
		Priority:    50,
		Interval:    50,
		Preempt:     false,
		Accept:      false,
		Unicast:     false,
		IpAddresses: []string{"192.168.10.21"},
		Enabled:     true,
	},
}

// Test an adding of the VRRP
func TestAddVrrp(t *testing.T) {
	ctx, ifIndexes, vrrpHandler := vrrpTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	ctx.MockVpp.MockReply(&vrrp.VrrpVrAddDelReply{})
	err := vrrpHandler.VppAddVrrp(vrrpEntries[0])
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&vrrp.VrrpVrAddDelReply{})
	err = vrrpHandler.VppAddVrrp(vrrpEntries[1])
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&vrrp.VrrpVrAddDelReply{})
	err = vrrpHandler.VppAddVrrp(vrrpEntries[2])
	Expect(err).To(Succeed())
}

// Test a deletion of the VRRP
func TestDelVrrp(t *testing.T) {
	ctx, ifIndexes, vrrpHandler := vrrpTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	ctx.MockVpp.MockReply(&vrrp.VrrpVrAddDelReply{})
	err := vrrpHandler.VppDelVrrp(vrrpEntries[0])
	Expect(err).To(Succeed())
}

// Test a start of the VRRP
func TestStartVrrp(t *testing.T) {
	ctx, ifIndexes, vrrpHandler := vrrpTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	ctx.MockVpp.MockReply(&vrrp.VrrpVrStartStopReply{})
	err := vrrpHandler.VppStartVrrp(vrrpEntries[0])
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&vrrp.VrrpVrStartStopReply{})
	err = vrrpHandler.VppStartVrrp(vrrpEntries[1])
	Expect(err).To(Succeed())
}

// Test a stop of the VRRP
func TestStopVrrp(t *testing.T) {
	ctx, ifIndexes, vrrpHandler := vrrpTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndexes.Put("if1", &ifaceidx.IfaceMetadata{SwIfIndex: 1})

	ctx.MockVpp.MockReply(&vrrp.VrrpVrStartStopReply{})
	err := vrrpHandler.VppStopVrrp(vrrpEntries[0])
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&vrrp.VrrpVrStartStopReply{})
	err = vrrpHandler.VppStopVrrp(vrrpEntries[1])
	Expect(err).To(Succeed())
}

func vrrpTestSetup(t *testing.T) (*vppmock.TestCtx, ifaceidx.IfaceMetadataIndexRW, vppcalls.VrrpVppAPI) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test"), "test")
	vrrpHandler := vpp2106.NewVrrpVppHandler(ctx.MockChannel, ifIndexes, log)
	return ctx, ifIndexes, vrrpHandler
}
