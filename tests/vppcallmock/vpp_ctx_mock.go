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

package vppcallmock

import (
	"testing"

	"time"

	"git.fd.io/govpp.git/adapter/mock"
	govppapi "git.fd.io/govpp.git/api"
	govpp "git.fd.io/govpp.git/core"
	. "github.com/onsi/gomega"
)

// TestCtx is helping structure for unit testing. It wraps VppAdapter which is used instead of real VPP
type TestCtx struct {
	MockVpp     *mock.VppAdapter
	conn        *govpp.Connection
	channel     govppapi.Channel
	MockChannel *mockedChannel
}

// SetupTestCtx sets up all fields of TestCtx structure at the begining of test
func SetupTestCtx(t *testing.T) *TestCtx {
	RegisterTestingT(t)

	ctx := &TestCtx{
		MockVpp: &mock.VppAdapter{},
	}

	var err error
	ctx.conn, err = govpp.Connect(ctx.MockVpp)
	Expect(err).ShouldNot(HaveOccurred())

	ctx.channel, err = ctx.conn.NewAPIChannel()
	Expect(err).ShouldNot(HaveOccurred())

	ctx.MockChannel = &mockedChannel{channel: ctx.channel}

	return ctx
}

// TeardownTestCtx politely close all used resources
func (ctx *TestCtx) TeardownTestCtx() {
	ctx.channel.Close()
	ctx.conn.Disconnect()
}

// MockedChannel implements ChannelIntf for testing purposes
type mockedChannel struct {
	channel govppapi.Channel

	// Last message which passed through method SendRequest
	Msg govppapi.Message

	// List of all messages which passed through method SendRequest
	Msgs []govppapi.Message
}

// SendRequest just save input argument to structure field for future check
func (m *mockedChannel) SendRequest(msg govppapi.Message) govppapi.RequestCtx {
	m.Msg = msg
	m.Msgs = append(m.Msgs, msg)
	return m.channel.SendRequest(msg)
}

// SendMultiRequest just save input argument to structure field for future check
func (m *mockedChannel) SendMultiRequest(msg govppapi.Message) govppapi.MultiRequestCtx {
	m.Msg = msg
	m.Msgs = append(m.Msgs, msg)
	return m.channel.SendMultiRequest(msg)
}

// CheckMessageCompatibility does nothing for mocked channel
func (m *mockedChannel) CheckMessageCompatibility(msgs ...govppapi.Message) error {
	return nil
}

// SubscribeNotification subscribes for receiving of the specified notification messages via provided Go channel
func (m *mockedChannel) SubscribeNotification(notifChan chan govppapi.Message, msgFactory func() govppapi.Message) (*govppapi.NotifSubscription, error) {
	return m.channel.SubscribeNotification(notifChan, msgFactory)
}

// UnsubscribeNotification unsubscribes from receiving the notifications tied to the provided notification subscription
func (m *mockedChannel) UnsubscribeNotification(subscription *govppapi.NotifSubscription) error {
	return m.channel.UnsubscribeNotification(subscription)
}

// SetReplyTimeout sets the timeout for replies from VPP
func (m *mockedChannel) SetReplyTimeout(timeout time.Duration) {
	m.channel.SetReplyTimeout(timeout)
}

// GetNotificationChannel returns notification channel
func (m *mockedChannel) GetReplyChannel() <-chan *govppapi.VppReply {
	return m.channel.GetReplyChannel()
}

// GetRequestChannel returns notification channel
func (m *mockedChannel) GetRequestChannel() chan<- *govppapi.VppRequest {
	return m.channel.GetRequestChannel()
}

// GetNotificationChannel returns notification channel
func (m *mockedChannel) GetNotificationChannel() chan<- *govppapi.NotifSubscribeRequest {
	return m.channel.GetNotificationChannel()
}

// GetNotificationReplyChannel returns notification reply channel
func (m *mockedChannel) GetNotificationReplyChannel() <-chan error {
	return m.channel.GetNotificationReplyChannel()
}

// GetMessageDecoder returns message decoder
func (m *mockedChannel) GetMessageDecoder() govppapi.MessageDecoder {
	return m.channel.GetMessageDecoder()
}

// GetID returns channel's ID
func (m *mockedChannel) GetID() uint16 {
	return m.channel.GetID()
}

// Close closes channel
func (m *mockedChannel) Close() {
	m.channel.Close()
}
