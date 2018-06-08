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
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
	"testing"
)

// Test enable/disable proxy arp
func TestProxyArp(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.ProxyArpIntfcEnableDisableReply{})
	err := vppcalls.EnableProxyArpInterface(0, ctx.MockChannel, logrus.DefaultLogger(), nil)
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&ip.ProxyArpIntfcEnableDisableReply{})
	err = vppcalls.DisableProxyArpInterface(0, ctx.MockChannel, logrus.DefaultLogger(), nil)
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&ip.ProxyArpIntfcEnableDisableReply{Retval: 1})
	err = vppcalls.VppAddArp(&arpEntries[0], ctx.MockChannel, nil)
	Expect(err).To(Not(BeNil()))
}

// Test add/delete ip range for proxy arp
func TestProxyArpRange(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err := vppcalls.AddProxyArpRange([]byte{192, 168, 10, 20}, []byte{192, 168, 10, 30}, ctx.MockChannel, logrus.DefaultLogger(), nil)
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{})
	err = vppcalls.DeleteProxyArpRange([]byte{192, 168, 10, 23}, []byte{192, 168, 10, 27}, ctx.MockChannel, logrus.DefaultLogger(), nil)
	Expect(err).To(Succeed())

	ctx.MockVpp.MockReply(&ip.ProxyArpAddDelReply{Retval: 1})
	err = vppcalls.AddProxyArpRange([]byte{192, 168, 10, 23}, []byte{192, 168, 10, 27}, ctx.MockChannel, logrus.DefaultLogger(), nil)
	Expect(err).To(Not(BeNil()))
}
