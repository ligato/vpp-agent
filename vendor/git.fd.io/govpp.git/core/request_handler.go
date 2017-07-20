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

	logger "github.com/Sirupsen/logrus"

	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core/bin_api/vpe"
)

// watchRequests watches for requests on the request API channel and forwards them as messages to VPP.
func (c *Connection) watchRequests(ch *api.Channel, chMeta *channelMetadata) {
	for {
		select {
		case req, ok := <-ch.ReqChan:
			// new request on the request channel
			if !ok {
				// after closing the request channel, release API channel and return
				c.releaseAPIChannel(ch, chMeta)
				return
			}
			c.processRequest(ch, chMeta, req)

		case req := <-ch.NotifSubsChan:
			// new request on the notification subscribe channel
			c.processNotifSubscribeRequest(ch, req)
		}
	}
}

// processRequest processes a single request received on the request channel.
func (c *Connection) processRequest(ch *api.Channel, chMeta *channelMetadata, req *api.VppRequest) error {
	// check whether we are connected to VPP
	if atomic.LoadUint32(&c.connected) == 0 {
		error := errors.New("not connected to VPP, ignoring the request")
		log.Error(error)
		sendReply(ch, &api.VppReply{Error: error})
		return error
	}

	// retrieve message ID
	msgID, err := c.GetMessageID(req.Message)
	if err != nil {
		error := fmt.Errorf("unable to retrieve message ID: %v", err)
		log.WithFields(logger.Fields{
			"msg_name": req.Message.GetMessageName(),
			"msg_crc":  req.Message.GetCrcString(),
		}).Error(err)
		sendReply(ch, &api.VppReply{Error: error})
		return error
	}

	// encode the message into binary
	data, err := c.codec.EncodeMsg(req.Message, msgID)
	if err != nil {
		error := fmt.Errorf("unable to encode the messge: %v", err)
		log.WithFields(logger.Fields{
			"context": chMeta.id,
			"msg_id":  msgID,
		}).Error(error)
		sendReply(ch, &api.VppReply{Error: error})
		return error
	}

	if log.Level == logger.DebugLevel { // for performance reasons - logrus does some processing even if debugs are disabled
		log.WithFields(logger.Fields{
			"context":  chMeta.id,
			"msg_id":   msgID,
			"msg_size": len(data),
		}).Debug("Sending a message to VPP.")
	}

	// send the message
	if req.Multipart {
		// expect multipart response
		atomic.StoreUint32(&chMeta.multipart, 1)
	}

	// send the request to VPP
	c.vpp.SendMsg(chMeta.id, data)

	if req.Multipart {
		// send a control ping to determine end of the multipart response
		ping := &vpe.ControlPing{}
		pingData, _ := c.codec.EncodeMsg(ping, c.pingReqID)

		log.WithFields(logger.Fields{
			"context":  chMeta.id,
			"msg_id":   c.pingReqID,
			"msg_size": len(pingData),
		}).Debug("Sending a control ping to VPP.")

		c.vpp.SendMsg(chMeta.id, pingData)
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

	if log.Level == logger.DebugLevel { // for performance reasons - logrus does some processing even if debugs are disabled
		log.WithFields(logger.Fields{
			"context":  context,
			"msg_id":   msgID,
			"msg_size": len(data),
		}).Debug("Received a message from VPP.")
	}

	if context == 0 || conn.isNotificationMessage(msgID) {
		// process the message as a notification
		conn.sendNotifications(msgID, data)
		return
	}

	// match ch according to the context
	conn.channelsLock.RLock()
	ch, ok := conn.channels[context]
	conn.channelsLock.RUnlock()

	if !ok {
		log.WithFields(logger.Fields{
			"context": context,
			"msg_id":  msgID,
		}).Error("Context ID not known, ignoring the message.")
		return
	}

	chMeta := ch.Metadata().(*channelMetadata)
	lastReplyReceived := false
	// if this is a control ping reply and multipart request is being processed, treat this as a last part of the reply
	if msgID == conn.pingReplyID && atomic.CompareAndSwapUint32(&chMeta.multipart, 1, 0) {
		lastReplyReceived = true
	}

	// send the data to the channel
	sendReply(ch, &api.VppReply{
		MessageID:         msgID,
		Data:              data,
		LastReplyReceived: lastReplyReceived,
	})
}

// sendReply sends the reply into the go channel, if it cannot be completed without blocking, otherwise
// it logs the error and do not send the message.
func sendReply(ch *api.Channel, reply *api.VppReply) {
	select {
	case ch.ReplyChan <- reply:
		// reply sent successfully
	default:
		// unable to write into the channel without blocking
		log.WithFields(logger.Fields{
			"channel": ch,
			"msg_id":  reply.MessageID,
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
	// try to get the ID from the map
	c.msgIDsLock.RLock()
	id, ok := c.msgIDs[msgName+msgCrc]
	c.msgIDsLock.RUnlock()
	if ok {
		return id, nil
	}

	// get the ID using VPP API
	id, err := c.vpp.GetMsgID(msgName, msgCrc)
	if err != nil {
		error := fmt.Errorf("unable to retrieve message ID: %v", err)
		log.WithFields(logger.Fields{
			"msg_name": msgName,
			"msg_crc":  msgCrc,
		}).Errorf("unable to retrieve message ID: %v", err)
		return id, error
	}

	c.msgIDsLock.Lock()
	c.msgIDs[msgName+msgCrc] = id
	c.msgIDsLock.Unlock()

	return id, nil
}
