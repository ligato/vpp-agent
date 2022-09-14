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

package vppmock

import (
	"context"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"go.fd.io/govpp/adapter/mock"
	govppapi "go.fd.io/govpp/api"
	"go.fd.io/govpp/core"
	govpp "go.fd.io/govpp/core"
	log "go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
)

// TestCtx is a helper for unit testing.
// It wraps VppAdapter which is used instead of real VPP.
type TestCtx struct {
	// *GomegaWithT
	Context       context.Context
	MockVpp       *mock.VppAdapter
	MockStats     *mock.StatsAdapter
	MockChannel   *mockedChannel
	MockVPPClient *mockVPPClient
	PingReplyMsg  govppapi.Message
}

// SetupTestCtx sets up all fields of TestCtx structure at the begining of test
func SetupTestCtx(t *testing.T) *TestCtx {
	RegisterTestingT(t)
	// g := NewGomegaWithT(t) // TODO: use GomegaWithT

	ctx := &TestCtx{
		// GomegaWithT: g,
		Context:      context.Background(),
		MockVpp:      mock.NewVppAdapter(),
		MockStats:    mock.NewStatsAdapter(),
		PingReplyMsg: &govpp.ControlPingReply{VpePID: 123},
	}

	ctx.MockVPPClient = newMockVPPClient(ctx)

	return ctx
}

// TeardownTestCtx politely close all used resources
func (ctx *TestCtx) TeardownTestCtx() {
	ctx.MockVPPClient.mockedChannel.Close()
	ctx.MockVPPClient.conn.Disconnect()
}

// HandleReplies represents spec for MockReplyHandler.
type HandleReplies struct {
	Name     string
	Ping     bool
	Message  govppapi.Message
	Messages []govppapi.Message
}

// MockReplies sets up reply handler for give HandleReplies.
func (ctx *TestCtx) MockReplies(dataList []*HandleReplies) {
	var sendControlPing bool

	ctx.MockVpp.MockReplyHandler(func(request mock.MessageDTO) (reply []byte, msgID uint16, prepared bool) {
		// Following types are not automatically stored in mock adapter's map and will be sent with empty MsgName
		// TODO: initialize mock adapter's map with these
		switch request.MsgID {
		case 100:
			request.MsgName = "control_ping"
		case 101:
			request.MsgName = "control_ping_reply"
		case 200:
			request.MsgName = "sw_interface_dump"
		case 201:
			request.MsgName = "sw_interface_details"
		}

		if request.MsgName == "" {
			log.DefaultLogger().Fatalf("mockHandler received request (ID: %v) with empty MsgName, check if compatibility check is done before using this request", request.MsgID)
		}

		if sendControlPing {
			sendControlPing = false
			data := ctx.PingReplyMsg
			reply, err := ctx.MockVpp.ReplyBytes(request, data)
			Expect(err).To(BeNil())
			msgID, err := ctx.MockVpp.GetMsgID(data.GetMessageName(), data.GetCrcString())
			Expect(err).To(BeNil())
			return reply, msgID, true
		}

		for _, dataMock := range dataList {
			if request.MsgName == dataMock.Name {
				// Send control ping next iteration if set
				sendControlPing = dataMock.Ping
				if len(dataMock.Messages) > 0 {
					log.DefaultLogger().Infof(" MOCK HANDLER: mocking %d messages", len(dataMock.Messages))
					ctx.MockVpp.MockReply(dataMock.Messages...)
					return nil, 0, false
				}
				if dataMock.Message == nil {
					return nil, 0, false
				}
				msgID, err := ctx.MockVpp.GetMsgID(dataMock.Message.GetMessageName(), dataMock.Message.GetCrcString())
				Expect(err).To(BeNil())
				reply, err := ctx.MockVpp.ReplyBytes(request, dataMock.Message)
				Expect(err).To(BeNil())
				return reply, msgID, true
			}
		}

		if strings.HasSuffix(request.MsgName, "_dump") {
			sendControlPing = true
			return nil, 0, false
		}

		var err error
		replyMsg, id, ok := ctx.MockVpp.ReplyFor("", request.MsgName)
		if ok {
			reply, err = ctx.MockVpp.ReplyBytes(request, replyMsg)
			Expect(err).To(BeNil())
			msgID = id
			prepared = true
		} else {
			log.DefaultLogger().Warnf("NO REPLY FOR %v FOUND", request.MsgName)
		}

		return reply, msgID, prepared
	})
}

// MockedChannel implements ChannelIntf for testing purposes
type mockedChannel struct {
	govppapi.Channel

	// Last message which passed through method SendRequest
	Msg govppapi.Message

	// List of all messages which passed through method SendRequest
	Msgs []govppapi.Message

	RetErrs []error

	notifications chan govppapi.Message
}

// SendRequest just save input argument to structure field for future check
func (m *mockedChannel) SendRequest(msg govppapi.Message) govppapi.RequestCtx {
	m.Msg = msg
	m.Msgs = append(m.Msgs, msg)
	reqCtx := m.Channel.SendRequest(msg)
	var retErr error
	if retErrsLen := len(m.RetErrs); retErrsLen > 0 {
		retErr = m.RetErrs[retErrsLen-1]
		m.RetErrs = m.RetErrs[:retErrsLen-1]
	}
	return &mockedContext{reqCtx, retErr}
}

// SendMultiRequest just save input argument to structure field for future check
func (m *mockedChannel) SendMultiRequest(msg govppapi.Message) govppapi.MultiRequestCtx {
	m.Msg = msg
	m.Msgs = append(m.Msgs, msg)
	return m.Channel.SendMultiRequest(msg)
}

func (m *mockedChannel) SubscribeNotification(notifChan chan govppapi.Message, event govppapi.Message) (govppapi.SubscriptionCtx, error) {
	m.notifications = notifChan
	return &mockSubscription{}, nil
}

func (m *mockedChannel) GetChannel() chan govppapi.Message {
	return m.notifications
}

type mockSubscription struct{}

func (s *mockSubscription) Unsubscribe() error {
	return nil
}

type mockedContext struct {
	requestCtx govppapi.RequestCtx
	retErr     error
}

// ReceiveReply returns prepared error or nil
func (m *mockedContext) ReceiveReply(msg govppapi.Message) error {
	if m.retErr != nil {
		return m.retErr
	}
	return m.requestCtx.ReceiveReply(msg)
}

type mockStream struct {
	mockVPPClient *mockVPPClient
	stream        govppapi.Stream
}

func (m *mockStream) SendMsg(msg govppapi.Message) error {
	m.mockVPPClient.mockedChannel.Msg = msg
	m.mockVPPClient.mockedChannel.Msgs = append(m.mockVPPClient.mockedChannel.Msgs, msg)
	if retErrsLen := len(m.mockVPPClient.mockedChannel.RetErrs); retErrsLen > 0 {
		retErr := m.mockVPPClient.mockedChannel.RetErrs[retErrsLen-1]
		m.mockVPPClient.mockedChannel.RetErrs = m.mockVPPClient.mockedChannel.RetErrs[:retErrsLen-1]
		return retErr
	}
	return m.stream.SendMsg(msg)
}

func (m *mockStream) RecvMsg() (govppapi.Message, error) {
	return m.stream.RecvMsg()
}

func (m *mockStream) Close() error {
	return m.stream.Close()
}

type mockVPPClient struct {
	ctx  *TestCtx
	conn *core.Connection
	*mockedChannel
	version         vpp.Version
	unloadedPlugins map[string]bool
}

func newMockVPPClient(ctx *TestCtx) *mockVPPClient {
	conn, err := govpp.Connect(ctx.MockVpp)
	Expect(err).ShouldNot(HaveOccurred())
	channel, err := conn.NewAPIChannel()
	Expect(err).ShouldNot(HaveOccurred())

	ctx.MockChannel = &mockedChannel{
		Channel: channel,
	}
	return &mockVPPClient{
		ctx:             ctx,
		conn:            conn,
		mockedChannel:   ctx.MockChannel,
		unloadedPlugins: map[string]bool{},
	}
}

func (m *mockVPPClient) NewStream(ctx context.Context, options ...govppapi.StreamOption) (govppapi.Stream, error) {
	stream, err := m.conn.NewStream(ctx, options...)
	if err != nil {
		return nil, err
	}
	return &mockStream{
		mockVPPClient: m,
		stream:        stream,
	}, nil
}

func (m *mockVPPClient) Invoke(ctx context.Context, req govppapi.Message, reply govppapi.Message) error {
	m.Msg = req
	m.Msgs = append(m.Msgs, req)
	return m.conn.Invoke(ctx, req, reply)
}

func (m *mockVPPClient) Version() vpp.Version {
	return m.version
}

func (m *mockVPPClient) NewAPIChannel() (govppapi.Channel, error) {
	return m.mockedChannel, nil
}

func (m *mockVPPClient) NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (govppapi.Channel, error) {
	return m.mockedChannel, nil
}

/*func (m *mockVPPClient) CheckCompatiblity(msgs ...govppapi.Message) error {
	return m.mockedChannel.CheckCompatiblity(msgs...)
}*/

func (m *mockVPPClient) IsPluginLoaded(plugin string) bool {
	return !m.unloadedPlugins[plugin]
}

func (m *mockVPPClient) BinapiVersion() vpp.Version {
	return ""
}

func (m *mockVPPClient) Stats() govppapi.StatsProvider {
	panic("implement me")
}

func (m *mockVPPClient) OnReconnect(h func()) {
	panic("implement me")
}
