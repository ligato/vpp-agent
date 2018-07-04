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

package core

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	logger "github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/api"
)

var (
	ErrNotConnected = errors.New("not connected to VPP, ignoring the request")
	ErrProbeTimeout = errors.New("probe reply not received within timeout period")
)

// watchRequests watches for requests on the request API channel and forwards them as messages to VPP.
func (c *Connection) watchRequests(ch *api.Channel) {
	for {
		select {
		case req, ok := <-ch.ReqChan:
			// new request on the request channel
			if !ok {
				// after closing the request channel, release API channel and return
				c.releaseAPIChannel(ch)
				return
			}
			c.processRequest(ch, req)

		case req := <-ch.NotifSubsChan:
			// new request on the notification subscribe channel
			c.processNotifSubscribeRequest(ch, req)
		}
	}
}

// processRequest processes a single request received on the request channel.
func (c *Connection) processRequest(ch *api.Channel, req *api.VppRequest) error {
	// check whether we are connected to VPP
	if atomic.LoadUint32(&c.connected) == 0 {
		err := ErrNotConnected
		log.Error(err)
		sendReply(ch, &api.VppReply{SeqNum: req.SeqNum, Error: err})
		return err
	}

	// retrieve message ID
	msgID, err := c.GetMessageID(req.Message)
	if err != nil {
		err = fmt.Errorf("unable to retrieve message ID: %v", err)
		log.WithFields(logger.Fields{
			"msg_name": req.Message.GetMessageName(),
			"msg_crc":  req.Message.GetCrcString(),
			"seq_num":  req.SeqNum,
		}).Error(err)
		sendReply(ch, &api.VppReply{SeqNum: req.SeqNum, Error: err})
		return err
	}

	// encode the message into binary
	data, err := c.codec.EncodeMsg(req.Message, msgID)
	if err != nil {
		err = fmt.Errorf("unable to encode the messge: %v", err)
		log.WithFields(logger.Fields{
			"channel": ch.ID,
			"msg_id":  msgID,
			"seq_num": req.SeqNum,
		}).Error(err)
		sendReply(ch, &api.VppReply{SeqNum: req.SeqNum, Error: err})
		return err
	}

	if log.Level == logger.DebugLevel { // for performance reasons - logrus does some processing even if debugs are disabled
		log.WithFields(logger.Fields{
			"channel":  ch.ID,
			"msg_id":   msgID,
			"msg_size": len(data),
			"msg_name": req.Message.GetMessageName(),
			"seq_num":  req.SeqNum,
		}).Debug("Sending a message to VPP.")
	}

	// send the request to VPP
	context := packRequestContext(ch.ID, req.Multipart, req.SeqNum)
	err = c.vpp.SendMsg(context, data)
	if err != nil {
		err = fmt.Errorf("unable to send the message: %v", err)
		log.WithFields(logger.Fields{
			"context": context,
			"msg_id":  msgID,
			"seq_num": req.SeqNum,
		}).Error(err)
		sendReply(ch, &api.VppReply{SeqNum: req.SeqNum, Error: err})
		return err
	}

	if req.Multipart {
		// send a control ping to determine end of the multipart response
		pingData, _ := c.codec.EncodeMsg(msgControlPing, c.pingReqID)

		log.WithFields(logger.Fields{
			"context":  context,
			"msg_id":   c.pingReqID,
			"msg_size": len(pingData),
			"seq_num":  req.SeqNum,
		}).Debug("Sending a control ping to VPP.")

		c.vpp.SendMsg(context, pingData)
	}

	return nil
}

// msgCallback is called whenever any binary API message comes from VPP.
func msgCallback(context uint32, msgID uint16, data []byte) {
	connLock.RLock()
	defer connLock.RUnlock()

	if conn == nil {
		log.Warn("Already disconnected, ignoring the message.")
		return
	}

	chanID, isMultipart, seqNum := unpackRequestContext(context)
	if log.Level == logger.DebugLevel { // for performance reasons - logrus does some processing even if debugs are disabled
		log.WithFields(logger.Fields{
			"msg_id":       msgID,
			"msg_size":     len(data),
			"channel_id":   chanID,
			"is_multipart": isMultipart,
			"seq_num":      seqNum,
		}).Debug("Received a message from VPP.")
	}

	if context == 0 || conn.isNotificationMessage(msgID) {
		// process the message as a notification
		conn.sendNotifications(msgID, data)
		return
	}

	// match ch according to the context
	conn.channelsLock.RLock()
	ch, ok := conn.channels[chanID]
	conn.channelsLock.RUnlock()

	if !ok {
		log.WithFields(logger.Fields{
			"channel_id": chanID,
			"msg_id":     msgID,
		}).Error("Channel ID not known, ignoring the message.")
		return
	}

	lastReplyReceived := false
	// if this is a control ping reply to a multipart request, treat this as a last part of the reply
	if msgID == conn.pingReplyID && isMultipart {
		lastReplyReceived = true
	}

	// send the data to the channel
	sendReply(ch, &api.VppReply{
		MessageID:         msgID,
		SeqNum:            seqNum,
		Data:              data,
		LastReplyReceived: lastReplyReceived,
	})

	// store actual time of this reply
	conn.lastReplyLock.Lock()
	conn.lastReply = time.Now()
	conn.lastReplyLock.Unlock()
}

// sendReply sends the reply into the go channel, if it cannot be completed without blocking, otherwise
// it logs the error and do not send the message.
func sendReply(ch *api.Channel, reply *api.VppReply) {
	select {
	case ch.ReplyChan <- reply:
		// reply sent successfully
	case <-time.After(time.Millisecond * 100):
		// receiver still not ready
		log.WithFields(logger.Fields{
			"channel": ch,
			"msg_id":  reply.MessageID,
			"seq_num": reply.SeqNum,
		}).Warn("Unable to send the reply, reciever end not ready.")
	}
}

// GetMessageID returns message identifier of given API message.
func (c *Connection) GetMessageID(msg api.Message) (uint16, error) {
	if c == nil {
		return 0, errors.New("nil connection passed in")
	}
	return c.messageNameToID(msg.GetMessageName(), msg.GetCrcString())
}

// messageNameToID returns message ID of a message identified by its name and CRC.
func (c *Connection) messageNameToID(msgName string, msgCrc string) (uint16, error) {
	msgKey := msgName + "_" + msgCrc

	// try to get the ID from the map
	c.msgIDsLock.RLock()
	id, ok := c.msgIDs[msgKey]
	c.msgIDsLock.RUnlock()
	if ok {
		return id, nil
	}

	// get the ID using VPP API
	id, err := c.vpp.GetMsgID(msgName, msgCrc)
	if err != nil {
		err = fmt.Errorf("unable to retrieve message ID: %v", err)
		log.WithFields(logger.Fields{
			"msg_name": msgName,
			"msg_crc":  msgCrc,
		}).Error(err)
		return id, err
	}

	c.msgIDsLock.Lock()
	c.msgIDs[msgKey] = id
	c.msgIDsLock.Unlock()

	return id, nil
}

// LookupByID looks up message name and crc by ID.
func (c *Connection) LookupByID(ID uint16) (string, error) {
	if c == nil {
		return "", errors.New("nil connection passed in")
	}

	c.msgIDsLock.Lock()
	defer c.msgIDsLock.Unlock()

	for key, id := range c.msgIDs {
		if id == ID {
			return key, nil
		}
	}

	return "", fmt.Errorf("unknown message ID: %d", ID)
}

// +------------------+-------------------+-----------------------+
// | 15b = channel ID | 1b = is multipart | 16b = sequence number |
// +------------------+-------------------+-----------------------+
func packRequestContext(chanID uint16, isMultipart bool, seqNum uint16) uint32 {
	context := uint32(chanID) << 17
	if isMultipart {
		context |= 1 << 16
	}
	context |= uint32(seqNum)
	return context
}

func unpackRequestContext(context uint32) (chanID uint16, isMulipart bool, seqNum uint16) {
	chanID = uint16(context >> 17)
	if ((context >> 16) & 0x1) != 0 {
		isMulipart = true
	}
	seqNum = uint16(context & 0xffff)
	return
}
