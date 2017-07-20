// Copyright (c) 2017 Cisco and/or its affiliates.
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

package api_test

import (
	"testing"
	"time"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/adapter/mock"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"git.fd.io/govpp.git/core/bin_api/vpe"
	"git.fd.io/govpp.git/examples/bin_api/interfaces"
	"git.fd.io/govpp.git/examples/bin_api/memif"
	"git.fd.io/govpp.git/examples/bin_api/tap"

	. "github.com/onsi/gomega"
)

type testCtx struct {
	mockVpp *mock.VppAdapter
	conn    *core.Connection
	ch      *api.Channel
}

func setupTest(t *testing.T) *testCtx {
	RegisterTestingT(t)

	ctx := &testCtx{}
	ctx.mockVpp = &mock.VppAdapter{}
	govpp.SetAdapter(ctx.mockVpp)

	var err error
	ctx.conn, err = govpp.Connect()
	Expect(err).ShouldNot(HaveOccurred())

	ctx.ch, err = ctx.conn.NewAPIChannel()
	Expect(err).ShouldNot(HaveOccurred())

	return ctx
}

func (ctx *testCtx) teardownTest() {
	ctx.ch.Close()
	ctx.conn.Disconnect()
}

func TestRequestReplyTapConnect(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	ctx.mockVpp.MockReply(&tap.TapConnectReply{
		Retval:    10,
		SwIfIndex: 1,
	})
	request := &tap.TapConnect{
		TapName:      []byte("test-tap-name"),
		UseRandomMac: 1,
	}
	reply := &tap.TapConnectReply{}

	err := ctx.ch.SendRequest(request).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reply.Retval).To(BeEquivalentTo(10), "Incorrect retval value for TapConnectReply")
	Expect(reply.SwIfIndex).To(BeEquivalentTo(1), "Incorrect SwIfIndex value for TapConnectReply")
}

func TestRequestReplyTapModify(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	ctx.mockVpp.MockReply(&tap.TapModifyReply{
		Retval:    15,
		SwIfIndex: 2,
	})
	request := &tap.TapModify{
		TapName:           []byte("test-tap-modify"),
		UseRandomMac:      1,
		CustomDevInstance: 1,
	}
	reply := &tap.TapModifyReply{}

	err := ctx.ch.SendRequest(request).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reply.Retval).To(BeEquivalentTo(15), "Incorrect retval value for TapModifyReply")
	Expect(reply.SwIfIndex).To(BeEquivalentTo(2), "Incorrect SwIfIndex value for TapModifyReply")
}

func TestRequestReplyTapDelete(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	ctx.mockVpp.MockReply(&tap.TapDeleteReply{
		Retval: 20,
	})
	request := &tap.TapDelete{
		SwIfIndex: 3,
	}
	reply := &tap.TapDeleteReply{}

	err := ctx.ch.SendRequest(request).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reply.Retval).To(BeEquivalentTo(20), "Incorrect retval value for TapDeleteReply")
}

func TestRequestReplySwInterfaceTapDump(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	byteName := []byte("dev-name-test")
	ctx.mockVpp.MockReply(&tap.SwInterfaceTapDetails{
		SwIfIndex: 25,
		DevName:   byteName,
	})
	request := &tap.SwInterfaceTapDump{}
	reply := &tap.SwInterfaceTapDetails{}

	err := ctx.ch.SendRequest(request).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reply.SwIfIndex).To(BeEquivalentTo(25), "Incorrect SwIfIndex value for SwInterfaceTapDetails")
	Expect(reply.DevName).ToNot(BeNil(), "Incorrect DevName value for SwInterfaceTapDetails")
}

func TestRequestReplyMemifCreate(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	ctx.mockVpp.MockReply(&memif.MemifCreateReply{
		Retval:    22,
		SwIfIndex: 4,
	})
	request := &memif.MemifCreate{
		Role:       10,
		ID:         12,
		RingSize:   8000,
		BufferSize: 50,
	}
	reply := &memif.MemifCreateReply{}

	err := ctx.ch.SendRequest(request).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reply.Retval).To(BeEquivalentTo(22), "Incorrect Retval value for MemifCreate")
	Expect(reply.SwIfIndex).To(BeEquivalentTo(4), "Incorrect SwIfIndex value for MemifCreate")
}

func TestRequestReplyMemifDelete(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	ctx.mockVpp.MockReply(&memif.MemifDeleteReply{
		Retval: 24,
	})
	request := &memif.MemifDelete{
		SwIfIndex: 15,
	}
	reply := &memif.MemifDeleteReply{}

	err := ctx.ch.SendRequest(request).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reply.Retval).To(BeEquivalentTo(24), "Incorrect Retval value for MemifDelete")
}

func TestRequestReplyMemifDetails(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	ctx.mockVpp.MockReply(&memif.MemifDetails{
		SwIfIndex: 25,
		IfName:    []byte("memif-name"),
		Role:      0,
	})
	request := &memif.MemifDump{}
	reply := &memif.MemifDetails{}

	err := ctx.ch.SendRequest(request).ReceiveReply(reply)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reply.SwIfIndex).To(BeEquivalentTo(25), "Incorrect SwIfIndex value for MemifDetails")
	Expect(reply.IfName).ToNot(BeEmpty(), "MemifDetails IfName is empty byte array")
	Expect(reply.Role).To(BeEquivalentTo(0), "Incorrect Role value for MemifDetails")
}

func TestMultiRequestReplySwInterfaceTapDump(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	for i := 1; i <= 10; i++ {
		byteName := []byte("dev-name-test")
		ctx.mockVpp.MockReply(&tap.SwInterfaceTapDetails{
			SwIfIndex: uint32(i),
			DevName:   byteName,
		})
	}
	ctx.mockVpp.MockReply(&vpe.ControlPingReply{})

	reqCtx := ctx.ch.SendMultiRequest(&tap.SwInterfaceTapDump{})
	cnt := 0
	for {
		msg := &tap.SwInterfaceTapDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break // break out of the loop
		}
		Expect(err).ShouldNot(HaveOccurred())
		cnt++
	}
	Expect(cnt).To(BeEquivalentTo(10))
}

func TestMultiRequestReplySwInterfaceMemifDump(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// mock reply
	for i := 1; i <= 10; i++ {
		ctx.mockVpp.MockReply(&memif.MemifDetails{
			SwIfIndex: uint32(i),
		})
	}
	ctx.mockVpp.MockReply(&vpe.ControlPingReply{})

	reqCtx := ctx.ch.SendMultiRequest(&memif.MemifDump{})
	cnt := 0
	for {
		msg := &memif.MemifDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break // break out of the loop
		}
		Expect(err).ShouldNot(HaveOccurred())
		cnt++
	}
	Expect(cnt).To(BeEquivalentTo(10))
}

func TestNotifications(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// subscribe for notification
	notifChan := make(chan api.Message, 1)
	subs, err := ctx.ch.SubscribeNotification(notifChan, interfaces.NewSwInterfaceSetFlags)
	Expect(err).ShouldNot(HaveOccurred())

	// mock the notification and force its delivery
	ctx.mockVpp.MockReply(&interfaces.SwInterfaceSetFlags{
		SwIfIndex:   3,
		AdminUpDown: 1,
	})
	ctx.mockVpp.SendMsg(0, []byte(""))

	// receive the notification
	notif := (<-notifChan).(*interfaces.SwInterfaceSetFlags)

	// verify the received notifications
	Expect(notif).ShouldNot(BeNil())
	Expect(notif.SwIfIndex).To(BeEquivalentTo(3), "Incorrect SwIfIndex value for SwInterfaceSetFlags")
	Expect(notif.AdminUpDown).To(BeEquivalentTo(1), "Incorrect AdminUpDown value for SwInterfaceSetFlags")

	ctx.ch.UnsubscribeNotification(subs)
}

func TestCheckMessageCompatibility(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	err := ctx.ch.CheckMessageCompatibility(&interfaces.SwInterfaceSetFlags{})
	Expect(err).ShouldNot(HaveOccurred())
}

func TestSetReplyTimeout(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	ctx.ch.SetReplyTimeout(time.Millisecond)

	// first one request should work
	ctx.mockVpp.MockReply(&vpe.ControlPingReply{})
	err := ctx.ch.SendRequest(&vpe.ControlPing{}).ReceiveReply(&vpe.ControlPingReply{})
	Expect(err).ShouldNot(HaveOccurred())

	// no other reply ready - expect timeout
	err = ctx.ch.SendRequest(&vpe.ControlPing{}).ReceiveReply(&vpe.ControlPingReply{})
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("timeout"))
}

func TestReceiveReplyNegative(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	// invalid context 1
	reqCtx1 := &api.RequestCtx{}
	err := reqCtx1.ReceiveReply(&vpe.ControlPingReply{})
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid request context"))

	// invalid context 2
	reqCtx2 := &api.MultiRequestCtx{}
	_, err = reqCtx2.ReceiveReply(&vpe.ControlPingReply{})
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid request context"))

	// NU
	reqCtx3 := &api.RequestCtx{}
	err = reqCtx3.ReceiveReply(nil)
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("invalid request context"))
}
