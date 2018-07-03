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

package core

import (
	"fmt"
	"time"

	"errors"

	"git.fd.io/govpp.git/api"
	"github.com/sirupsen/logrus"
)

const defaultReplyTimeout = time.Second * 1 // default timeout for replies from VPP, can be changed with SetReplyTimeout

// VppRequest is a request that will be sent to VPP.
type VppRequest struct {
	SeqNum    uint16      // sequence number
	Message   api.Message // binary API message to be send to VPP
	Multipart bool        // true if multipart response is expected, false otherwise
}

// VppReply is a reply received from VPP.
type VppReply struct {
	MessageID         uint16 // ID of the message
	SeqNum            uint16 // sequence number
	Data              []byte // encoded data with the message - MessageDecoder can be used for decoding
	LastReplyReceived bool   // in case of multipart replies, true if the last reply has been already received and this one should be ignored
	Error             error  // in case of error, data is nil and this member contains error description
}

// NotifSubscribeRequest is a request to subscribe for delivery of specific notification messages.
type NotifSubscribeRequest struct {
	Subscription *NotifSubscription // subscription details
	Subscribe    bool               // true if this is a request to subscribe, false if unsubscribe
}

// NotifSubscription represents a subscription for delivery of specific notification messages.
type NotifSubscription struct {
	NotifChan  chan api.Message   // channel where notification messages will be delivered to
	MsgFactory func() api.Message // function that returns a new instance of the specific message that is expected as a notification
}

// RequestCtx is a context of a ongoing request (simple one - only one response is expected).
type RequestCtx struct {
	ch     *channel
	seqNum uint16
}

// ReceiveReply receives a reply from VPP (blocks until a reply is delivered from VPP, or until an error occurs).
// The reply will be decoded into the msg argument. Error will be returned if the response cannot be received or decoded.
func (req *RequestCtx) ReceiveReply(msg api.Message) error {
	if req == nil || req.ch == nil {
		return errors.New("invalid request context")
	}

	lastReplyReceived, err := req.ch.receiveReplyInternal(msg, req.seqNum)

	if lastReplyReceived {
		err = errors.New("multipart reply recieved while a simple reply expected")
	}
	return err
}

// MultiRequestCtx is a context of a ongoing multipart request (multiple responses are expected).
type MultiRequestCtx struct {
	ch     *channel
	seqNum uint16
}

// ReceiveReply receives a reply from VPP (blocks until a reply is delivered from VPP, or until an error occurs).
// The reply will be decoded into the msg argument. If the last reply has been already consumed, lastReplyReceived is
// set to true. Do not use the message itself if lastReplyReceived is true - it won't be filled with actual data.
// Error will be returned if the response cannot be received or decoded.
func (req *MultiRequestCtx) ReceiveReply(msg api.Message) (lastReplyReceived bool, err error) {
	if req == nil || req.ch == nil {
		return false, errors.New("invalid request context")
	}

	return req.ch.receiveReplyInternal(msg, req.seqNum)
}

// channel is the main communication interface with govpp core. It contains four Go channels, one for sending the requests
// to VPP, one for receiving the replies from it and the same set for notifications. The user can access the Go channels
// via methods provided by Channel interface in this package. Do not use the same channel from multiple goroutines
// concurrently, otherwise the responses could mix! Use multiple channels instead.
type channel struct {
	ID uint16 // channel ID

	ReqChan   chan *VppRequest // channel for sending the requests to VPP, closing this channel releases all resources in the ChannelProvider
	ReplyChan chan *VppReply   // channel where VPP replies are delivered to

	NotifSubsChan      chan *NotifSubscribeRequest // channel for sending notification subscribe requests
	NotifSubsReplyChan chan error                  // channel where replies to notification subscribe requests are delivered to

	MsgDecoder    MessageDecoder    // used to decode binary data to generated API messages
	MsgIdentifier MessageIdentifier // used to retrieve message ID of a message

	lastSeqNum uint16 // sequence number of the last sent request

	delayedReply *VppReply     // reply already taken from ReplyChan, buffered for later delivery
	replyTimeout time.Duration // maximum time that the API waits for a reply from VPP before returning an error, can be set with SetReplyTimeout
}

func (ch *channel) SendRequest(msg api.Message) *RequestCtx {
	ch.lastSeqNum++
	ch.ReqChan <- &VppRequest{
		Message: msg,
		SeqNum:  ch.lastSeqNum,
	}
	return &RequestCtx{ch: ch, seqNum: ch.lastSeqNum}
}

func (ch *channel) SendMultiRequest(msg api.Message) *MultiRequestCtx {
	ch.lastSeqNum++
	ch.ReqChan <- &VppRequest{
		Message:   msg,
		Multipart: true,
		SeqNum:    ch.lastSeqNum,
	}
	return &MultiRequestCtx{ch: ch, seqNum: ch.lastSeqNum}
}

func (ch *channel) SubscribeNotification(notifChan chan api.Message, msgFactory func() api.Message) (*NotifSubscription, error) {
	subscription := &NotifSubscription{
		NotifChan:  notifChan,
		MsgFactory: msgFactory,
	}
	ch.NotifSubsChan <- &NotifSubscribeRequest{
		Subscription: subscription,
		Subscribe:    true,
	}
	return subscription, <-ch.NotifSubsReplyChan
}

func (ch *channel) UnsubscribeNotification(subscription *NotifSubscription) error {
	ch.NotifSubsChan <- &NotifSubscribeRequest{
		Subscription: subscription,
		Subscribe:    false,
	}
	return <-ch.NotifSubsReplyChan
}

func (ch *channel) CheckMessageCompatibility(messages ...api.Message) error {
	for _, msg := range messages {
		_, err := ch.MsgIdentifier.GetMessageID(msg)
		if err != nil {
			return fmt.Errorf("message %s with CRC %s is not compatible with the VPP we are connected to",
				msg.GetMessageName(), msg.GetCrcString())
		}
	}
	return nil
}

func (ch *channel) SetReplyTimeout(timeout time.Duration) {
	ch.replyTimeout = timeout
}

func (ch *channel) GetRequestChannel() chan<- *VppRequest {
	return ch.ReqChan
}

func (ch *channel) GetReplyChannel() <-chan *VppReply {
	return ch.ReplyChan
}

func (ch *channel) GetNotificationChannel() chan<- *NotifSubscribeRequest {
	return ch.NotifSubsChan
}

func (ch *channel) GetNotificationReplyChannel() <-chan error {
	return ch.NotifSubsReplyChan
}

func (ch *channel) GetMessageDecoder() MessageDecoder {
	return ch.MsgDecoder
}

func (ch *channel) GetID() uint16 {
	return ch.ID
}

func (ch *channel) Close() {
	if ch.ReqChan != nil {
		close(ch.ReqChan)
	}
}

// NewChannelInternal returns a new channel structure.
// Note that this is just a raw channel not yet connected to VPP, it is not intended to be used directly.
// Use ChannelProvider to get an API channel ready for communication with VPP.
func NewChannelInternal(id uint16) *channel {
	return &channel{
		ID:           id,
		replyTimeout: defaultReplyTimeout,
	}
}

// receiveReplyInternal receives a reply from the reply channel into the provided msg structure.
func (ch *channel) receiveReplyInternal(msg api.Message, expSeqNum uint16) (lastReplyReceived bool, err error) {
	var ignore bool
	if msg == nil {
		return false, errors.New("nil message passed in")
	}

	if ch.delayedReply != nil {
		// try the delayed reply
		vppReply := ch.delayedReply
		ch.delayedReply = nil
		ignore, lastReplyReceived, err = ch.processReply(vppReply, expSeqNum, msg)
		if !ignore {
			return lastReplyReceived, err
		}
	}

	timer := time.NewTimer(ch.replyTimeout)
	for {
		select {
		// blocks until a reply comes to ReplyChan or until timeout expires
		case vppReply := <-ch.ReplyChan:
			ignore, lastReplyReceived, err = ch.processReply(vppReply, expSeqNum, msg)
			if ignore {
				continue
			}
			return lastReplyReceived, err

		case <-timer.C:
			err = fmt.Errorf("no reply received within the timeout period %s", ch.replyTimeout)
			return false, err
		}
	}
	return
}

func (ch *channel) processReply(reply *VppReply, expSeqNum uint16, msg api.Message) (ignore bool, lastReplyReceived bool, err error) {
	// check the sequence number
	cmpSeqNums := compareSeqNumbers(reply.SeqNum, expSeqNum)
	if cmpSeqNums == -1 {
		// reply received too late, ignore the message
		logrus.WithField("sequence-number", reply.SeqNum).Warn(
			"Received reply to an already closed binary API request")
		ignore = true
		return
	}
	if cmpSeqNums == 1 {
		ch.delayedReply = reply
		err = fmt.Errorf("missing binary API reply with sequence number: %d", expSeqNum)
		return
	}

	if reply.Error != nil {
		err = reply.Error
		return
	}
	if reply.LastReplyReceived {
		lastReplyReceived = true
		return
	}

	// message checks
	var expMsgID uint16
	expMsgID, err = ch.MsgIdentifier.GetMessageID(msg)
	if err != nil {
		err = fmt.Errorf("message %s with CRC %s is not compatible with the VPP we are connected to",
			msg.GetMessageName(), msg.GetCrcString())
		return
	}

	if reply.MessageID != expMsgID {
		var msgNameCrc string
		if nameCrc, err := ch.MsgIdentifier.LookupByID(reply.MessageID); err != nil {
			msgNameCrc = err.Error()
		} else {
			msgNameCrc = nameCrc
		}

		err = fmt.Errorf("received invalid message ID (seq-num=%d), expected %d (%s), but got %d (%s) "+
			"(check if multiple goroutines are not sharing single GoVPP channel)",
			reply.SeqNum, expMsgID, msg.GetMessageName(), reply.MessageID, msgNameCrc)
		return
	}

	// decode the message
	err = ch.MsgDecoder.DecodeMsg(reply.Data, msg)
	return
}

// compareSeqNumbers returns -1, 0, 1 if sequence number <seqNum1> precedes, equals to,
// or succeeds seq. number <seqNum2>.
// Since sequence numbers cycle in the finite set of size 2^16, the function
// must assume that the distance between compared sequence numbers is less than
// (2^16)/2 to determine the order.
func compareSeqNumbers(seqNum1, seqNum2 uint16) int {
	// calculate distance from seqNum1 to seqNum2
	var dist uint16
	if seqNum1 <= seqNum2 {
		dist = seqNum2 - seqNum1
	} else {
		dist = 0xffff - (seqNum1 - seqNum2 - 1)
	}
	if dist == 0 {
		return 0
	} else if dist <= 0x8000 {
		return -1
	}
	return 1
}
