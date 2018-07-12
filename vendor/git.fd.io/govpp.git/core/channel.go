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

// requestCtxData is a context of a ongoing request (simple one - only one response is expected).
type requestCtxData struct {
	ch     *channel
	seqNum uint16
}

// multiRequestCtxData is a context of a ongoing multipart request (multiple responses are expected).
type multiRequestCtxData struct {
	ch     *channel
	seqNum uint16
}

func (req *requestCtxData) ReceiveReply(msg api.Message) error {
	if req == nil || req.ch == nil {
		return errors.New("invalid request context")
	}

	lastReplyReceived, err := req.ch.receiveReplyInternal(msg, req.seqNum)

	if lastReplyReceived {
		err = errors.New("multipart reply recieved while a simple reply expected")
	}
	return err
}

func (req *multiRequestCtxData) ReceiveReply(msg api.Message) (lastReplyReceived bool, err error) {
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
	id uint16 // channel ID

	reqChan   chan *api.VppRequest // channel for sending the requests to VPP, closing this channel releases all resources in the ChannelProvider
	replyChan chan *api.VppReply   // channel where VPP replies are delivered to

	notifSubsChan      chan *api.NotifSubscribeRequest // channel for sending notification subscribe requests
	notifSubsReplyChan chan error                      // channel where replies to notification subscribe requests are delivered to

	msgDecoder    api.MessageDecoder    // used to decode binary data to generated API messages
	msgIdentifier api.MessageIdentifier // used to retrieve message ID of a message

	lastSeqNum uint16 // sequence number of the last sent request

	delayedReply *api.VppReply // reply already taken from ReplyChan, buffered for later delivery
	replyTimeout time.Duration // maximum time that the API waits for a reply from VPP before returning an error, can be set with SetReplyTimeout
}

func (ch *channel) SendRequest(msg api.Message) api.RequestCtx {
	ch.lastSeqNum++
	ch.reqChan <- &api.VppRequest{
		Message: msg,
		SeqNum:  ch.lastSeqNum,
	}
	return &requestCtxData{ch: ch, seqNum: ch.lastSeqNum}
}

func (ch *channel) SendMultiRequest(msg api.Message) api.MultiRequestCtx {
	ch.lastSeqNum++
	ch.reqChan <- &api.VppRequest{
		Message:   msg,
		Multipart: true,
		SeqNum:    ch.lastSeqNum,
	}
	return &multiRequestCtxData{ch: ch, seqNum: ch.lastSeqNum}
}

func (ch *channel) SubscribeNotification(notifChan chan api.Message, msgFactory func() api.Message) (*api.NotifSubscription, error) {
	subscription := &api.NotifSubscription{
		NotifChan:  notifChan,
		MsgFactory: msgFactory,
	}
	ch.notifSubsChan <- &api.NotifSubscribeRequest{
		Subscription: subscription,
		Subscribe:    true,
	}
	return subscription, <-ch.notifSubsReplyChan
}

func (ch *channel) UnsubscribeNotification(subscription *api.NotifSubscription) error {
	ch.notifSubsChan <- &api.NotifSubscribeRequest{
		Subscription: subscription,
		Subscribe:    false,
	}
	return <-ch.notifSubsReplyChan
}

func (ch *channel) CheckMessageCompatibility(messages ...api.Message) error {
	for _, msg := range messages {
		_, err := ch.msgIdentifier.GetMessageID(msg)
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

func (ch *channel) GetRequestChannel() chan<- *api.VppRequest {
	return ch.reqChan
}

func (ch *channel) GetReplyChannel() <-chan *api.VppReply {
	return ch.replyChan
}

func (ch *channel) GetNotificationChannel() chan<- *api.NotifSubscribeRequest {
	return ch.notifSubsChan
}

func (ch *channel) GetNotificationReplyChannel() <-chan error {
	return ch.notifSubsReplyChan
}

func (ch *channel) GetMessageDecoder() api.MessageDecoder {
	return ch.msgDecoder
}

func (ch *channel) GetID() uint16 {
	return ch.id
}

func (ch *channel) Close() {
	if ch.reqChan != nil {
		close(ch.reqChan)
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
		case vppReply := <-ch.replyChan:
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

func (ch *channel) processReply(reply *api.VppReply, expSeqNum uint16, msg api.Message) (ignore bool, lastReplyReceived bool, err error) {
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
	expMsgID, err = ch.msgIdentifier.GetMessageID(msg)
	if err != nil {
		err = fmt.Errorf("message %s with CRC %s is not compatible with the VPP we are connected to",
			msg.GetMessageName(), msg.GetCrcString())
		return
	}

	if reply.MessageID != expMsgID {
		var msgNameCrc string
		if nameCrc, err := ch.msgIdentifier.LookupByID(reply.MessageID); err != nil {
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
	err = ch.msgDecoder.DecodeMsg(reply.Data, msg)
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
