//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package vppcalls_test

import (
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
	"net"
	"testing"
)

var routes = []vppcalls.Route{
	{
		VrfID:       1,
		DstAddr:     net.IPNet{IP: []byte{192, 168, 10, 21}, Mask: []byte{255, 255, 255, 0}},
		NextHopAddr: []byte{192, 168, 30, 1},
	},
	{
		VrfID:       2,
		DstAddr:     net.IPNet{IP: []byte{0xde, 0xad, 0, 0, 0, 0, 0, 0, 0xde, 0xad, 0, 0, 0, 0, 0, 1}, Mask: []byte{}},
		NextHopAddr: []byte{192, 168, 30, 1},
	},
}

// Test adding routes
func TestAddRoute(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPFibDetails{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})
	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err := vppcalls.VppAddRoute(&routes[0], ctx.MockChannel, nil)
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err = vppcalls.VppAddRoute(&routes[0], ctx.MockChannel, nil)
	Expect(err).To(Not(BeNil()))
}

// Test deleteing routes
func TestDeleteRoute(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err := vppcalls.VppDelRoute(&routes[0], ctx.MockChannel, nil)
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{})
	err = vppcalls.VppDelRoute(&routes[1], ctx.MockChannel, nil)
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&ip.IPAddDelRouteReply{1})
	err = vppcalls.VppDelRoute(&routes[0], ctx.MockChannel, nil)
	Expect(err).To(Not(BeNil()))
}
