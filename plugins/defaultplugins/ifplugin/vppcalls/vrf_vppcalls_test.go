// Copyright (c) 2018 Cisco and/or its affiliates.
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

package vppcalls_test

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/vpe"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

func TestGetInterfaceVRF(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceGetTableReply{
		VrfID: 1,
	})

	vrfID, err := vppcalls.GetInterfaceVRF(1, logrus.DefaultLogger(), ctx.MockChannel)
	Expect(err).To(BeNil())
	Expect(vrfID).To(BeEquivalentTo(1))
}

func TestGetInterfaceVRFError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceGetTable{})

	_, err := vppcalls.GetInterfaceVRF(1, logrus.DefaultLogger(), ctx.MockChannel)
	Expect(err).ToNot(BeNil())
}

func TestGetInterfaceVRFRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceGetTableReply{
		Retval: 1,
	})

	_, err := vppcalls.GetInterfaceVRF(1, logrus.DefaultLogger(), ctx.MockChannel)
	Expect(err).ToNot(BeNil())
}

func TestSetInterfaceVRF(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})

	err := vppcalls.SetInterfaceVRF(1, 2, logrus.DefaultLogger(), ctx.MockChannel)
	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.SwInterfaceSetTable)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.VrfID).To(BeEquivalentTo(2))
}

func TestSetInterfaceVRFError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTable{})

	err := vppcalls.SetInterfaceVRF(1, 2, logrus.DefaultLogger(), ctx.MockChannel)
	Expect(err).ToNot(BeNil())
}

func TestSetInterfaceVRFRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{
		Retval: 1,
	})

	err := vppcalls.SetInterfaceVRF(1, 2, logrus.DefaultLogger(), ctx.MockChannel)
	Expect(err).ToNot(BeNil())
}

func TestCreateVrfIfNeeded(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// IP FIB dump
	ctx.MockVpp.MockReply(&ip.IPFibDetails{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	// Add/del table
	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})

	err := vppcalls.CreateVrfIfNeeded(1, ctx.MockChannel)
	Expect(err).To(BeNil())
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*ip.IPTableAddDel)
		if ok {
			Expect(vppMsg.TableID).To(BeEquivalentTo(1))
			Expect(vppMsg.IsIpv6).To(BeEquivalentTo(0))
			Expect(vppMsg.IsAdd).To(BeEquivalentTo(1))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestCreateVrfIfNeededNull(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// IP FIB dump
	ctx.MockVpp.MockReply(&ip.IPFibDetails{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	// Add/del table
	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{})

	err := vppcalls.CreateVrfIfNeeded(0, ctx.MockChannel)
	Expect(err).To(BeNil())
}

func TestCreateVrfIfNeededError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// IP FIB dump
	ctx.MockVpp.MockReply(&ip.IPFibDetails{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	// Add/del table
	ctx.MockVpp.MockReply(&ip.IPTableAddDel{})

	err := vppcalls.CreateVrfIfNeeded(1, ctx.MockChannel)
	Expect(err).ToNot(BeNil())
}

func TestCreateVrfIfNeededRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// IP FIB dump
	ctx.MockVpp.MockReply(&ip.IPFibDetails{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	// Add/del table
	ctx.MockVpp.MockReply(&ip.IPTableAddDelReply{
		Retval: 1,
	})

	err := vppcalls.CreateVrfIfNeeded(1, ctx.MockChannel)
	Expect(err).ToNot(BeNil())
}
