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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"git.fd.io/govpp.git/api"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidRequestCtx = errors.New("invalid request context")
)

// requestCtx is a context for request with single reply
type requestCtx struct {
	ch     *channel
	seqNum uint16
}

// multiRequestCtx is a context for request with multiple responses
type multiRequestCtx struct {
	ch     *channel
	seqNum uint16
}

func (req *requestCtx) ReceiveReply(msg api.Message) error {
	if req == nil || req.ch == nil {
		return ErrInvalidRequestCtx
	}

	lastReplyReceived, err := req.ch.receiveReplyInternal(msg, req.seqNum)
	if err != nil {
		return err
	}
	if lastReplyReceived {
		return errors.New("multipart reply recieved while a single reply expected")
	}

	return nil
}

func (req *multiRequestCtx) ReceiveReply(msg api.Message) (lastReplyReceived bool, err error) {
	if req == nil || req.ch == nil {
		return false, ErrInvalidRequestCtx
	}

	return req.ch.receiveReplyInternal(msg, req.seqNum)
}

// vppRequest is a request that will be sent to VPP.
type vppRequest struct {
	seqNum uint16      // sequence number
	msg    api.Message // binary API message to be send to VPP
	multi  bool        // true if multipart response is expected
}

// vppReply is a reply received from VPP.
type vppReply struct {
	seqNum       uint16 // sequence number
	msgID        uint16 // ID of the message
	data         []byte // encoded data with the message
	lastReceived bool   // for multi request, true if the last reply has been already received
	err          error  // in case of error, data is nil and this member contains error
}

// NotifSubscribeRequest is a request to subscribe for delivery of specific notification messages.
type subscriptionRequest struct {
	sub       *api.NotifSubscription // subscription details
	subscribe bool                   // true if this is a request to subscribe
}

// channel is the main communication interface with govpp core. It contains four Go channels, one for sending the requests
// to VPP, one for receiving the replies from it and the same set for notifications. The user can access the Go channels
// via methods provided by Channel interface in this package. Do not use the same channel from multiple goroutines
// concurrently, otherwise the responses could mix! Use multiple channels instead.
type channel struct {
	id uint16 // channel ID

	reqChan   chan *vppRequest // channel for sending the requests to VPP
	replyChan chan *vppReply   // channel where VPP replies are delivered to

	notifSubsChan      chan *subscriptionRequest // channel for sending notification subscribe requests
	notifSubsReplyChan chan error                // channel where replies to notification subscribe requests are delivered to

	msgDecoder    api.MessageDecoder    // used to decode binary data to generated API messages
	msgIdentifier api.MessageIdentifier // used to retrieve message ID of a message

	lastSeqNum uint16 // sequence number of the last sent request

	delayedReply *vppReply     // reply already taken from ReplyChan, buffered for later delivery
	replyTimeout time.Duration // maximum time that the API waits for a reply from VPP before returning an error, can be set with SetReplyTimeout
}

func (ch *channel) GetID() uint16 {
	return ch.id
}

func (ch *channel) nextSeqNum() uint16 {
	ch.lastSeqNum++
	return ch.lastSeqNum
}

func (ch *channel) SendRequest(msg api.Message) api.RequestCtx {
	req := &vppRequest{
		msg:    msg,
		seqNum: ch.nextSeqNum(),
	}
	ch.reqChan <- req
	return &requestCtx{ch: ch, seqNum: req.seqNum}
}

func (ch *channel) SendMultiRequest(msg api.Message) api.MultiRequestCtx {
	req := &vppRequest{
		msg:    msg,
		seqNum: ch.nextSeqNum(),
		multi:  true,
	}
	ch.reqChan <- req
	return &multiRequestCtx{ch: ch, seqNum: req.seqNum}
}

func (ch *channel) SubscribeNotification(notifChan chan api.Message, msgFactory func() api.Message) (*api.NotifSubscription, error) {
	sub := &api.NotifSubscription{
		NotifChan:  notifChan,
		MsgFactory: msgFactory,
	}
	// TODO: get rid of notifSubsChan and notfSubsReplyChan,
	// it's no longer need because we know all message IDs and can store subscription right here
	ch.notifSubsChan <- &subscriptionRequest{
		sub:       sub,
		subscribe: true,
	}
	return sub, <-ch.notifSubsReplyChan
}

func (ch *channel) UnsubscribeNotification(subscription *api.NotifSubscription) error {
	ch.notifSubsChan <- &subscriptionRequest{
		sub:       subscription,
		subscribe: false,
	}
	return <-ch.notifSubsReplyChan
}

func (ch *channel) SetReplyTimeout(timeout time.Duration) {
	ch.replyTimeout = timeout
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

	if vppReply := ch.delayedReply; vppReply != nil {
		// try the delayed reply
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

func (ch *channel) processReply(reply *vppReply, expSeqNum uint16, msg api.Message) (ignore bool, lastReplyReceived bool, err error) {
	// check the sequence number
	cmpSeqNums := compareSeqNumbers(reply.seqNum, expSeqNum)
	if cmpSeqNums == -1 {
		// reply received too late, ignore the message
		logrus.WithField("sequence-number", reply.seqNum).Warn(
			"Received reply to an already closed binary API request")
		ignore = true
		return
	}
	if cmpSeqNums == 1 {
		ch.delayedReply = reply
		err = fmt.Errorf("missing binary API reply with sequence number: %d", expSeqNum)
		return
	}

	if reply.err != nil {
		err = reply.err
		return
	}
	if reply.lastReceived {
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

	if reply.msgID != expMsgID {
		var msgNameCrc string
		if replyMsg, err := ch.msgIdentifier.LookupByID(reply.msgID); err != nil {
			msgNameCrc = err.Error()
		} else {
			msgNameCrc = getMsgNameWithCrc(replyMsg)
		}

		err = fmt.Errorf("received invalid message ID (seqNum=%d), expected %d (%s), but got %d (%s) "+
			"(check if multiple goroutines are not sharing single GoVPP channel)",
			reply.seqNum, expMsgID, msg.GetMessageName(), reply.msgID, msgNameCrc)
		return
	}

	// decode the message
	if err = ch.msgDecoder.DecodeMsg(reply.data, msg); err != nil {
		return
	}

	// check Retval and convert it into VnetAPIError error
	if strings.HasSuffix(msg.GetMessageName(), "_reply") {
		// TODO: use categories for messages to avoid checking message name
		if f := reflect.Indirect(reflect.ValueOf(msg)).FieldByName("Retval"); f.IsValid() {
			if retval := f.Int(); retval != 0 {
				err = api.VPPApiError(retval)
			}
		}
	}

	return
}
