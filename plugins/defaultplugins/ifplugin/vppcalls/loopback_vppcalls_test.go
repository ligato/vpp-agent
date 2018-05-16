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

	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

func TestAddLoopbackInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.CreateLoopbackReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	swIfIdx, err := vppcalls.AddLoopbackInterface("loopback", ctx.MockChannel, nil)

	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
}

func TestAddLoopbackInterfaceError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.CreateLoopback{})

	swIfIdx, err := vppcalls.AddLoopbackInterface("loopback", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(0))
}

func TestAddLoopbackInterfaceRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.CreateLoopbackReply{
		Retval: 1,
	})

	swIfIdx, err := vppcalls.AddLoopbackInterface("loopback", ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(0))
}

func TestDeleteLoopbackInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.DeleteLoopbackReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})

	err := vppcalls.DeleteLoopbackInterface("loopback", 1, ctx.MockChannel, nil)

	Expect(err).To(BeNil())
}

func TestDeleteLoopbackInterfaceError(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.DeleteLoopback{})

	err := vppcalls.DeleteLoopbackInterface("loopback", 1, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}

func TestDeleteLoopbackInterfaceRetval(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.DeleteLoopbackReply{
		Retval: 1,
	})

	err := vppcalls.DeleteLoopbackInterface("loopback", 1, ctx.MockChannel, nil)

	Expect(err).ToNot(BeNil())
}
