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

// Package mock is an alternative VPP adapter aimed for unit/integration testing where the
// actual communication with VPP is not demanded.
package mock

import (
	"bytes"
	"log"
	"reflect"
	"sync"

	"git.fd.io/govpp.git/core"
	"github.com/lunixbochs/struc"

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/adapter/mock/binapi"
	"git.fd.io/govpp.git/api"
)

// VppAdapter represents a mock VPP adapter that can be used for unit/integration testing instead of the vppapiclient adapter.
type VppAdapter struct {
	callback func(context uint32, msgId uint16, data []byte)

	msgNameToIds map[string]uint16
	msgIDsToName map[uint16]string
	msgIDSeq     uint16
	binAPITypes  map[string]reflect.Type
	access       sync.RWMutex
}

// defaultReply is a default reply message that mock adapter returns for a request.
type defaultReply struct {
	Retval int32
}

// MessageDTO is a structure used for propagating information to ReplyHandlers.
type MessageDTO struct {
	MsgID    uint16
	MsgName  string
	ClientID uint32
	Data     []byte
}

// ReplyHandler is a type that allows to extend the behaviour of VPP mock.
// Return value ok is used to signalize that mock reply is calculated and ready to be used.
type ReplyHandler func(request MessageDTO) (reply []byte, msgID uint16, ok bool)

const (
	defaultReplyMsgID = 1 // default message ID for the reply to be sent back via callback
)

var replies []api.Message        // FIFO queue of messages
var replyHandlers []ReplyHandler // callbacks that are able to calculate mock responses
var repliesLock sync.Mutex       // mutex for the queue
var mode = 0                     // mode in which the mock operates

const useRepliesQueue = 1  // use replies in the queue instead of the default one
const useReplyHandlers = 2 //use ReplyHandler

// NewVppAdapter returns a new mock adapter.
func NewVppAdapter() adapter.VppAdapter {
	return &VppAdapter{}
}

// Connect emulates connecting the process to VPP.
func (a *VppAdapter) Connect() error {
	return nil
}

// Disconnect emulates disconnecting the process from VPP.
func (a *VppAdapter) Disconnect() {
	// no op
}

// GetMsgNameByID returns message name for specified message ID.
func (a *VppAdapter) GetMsgNameByID(msgID uint16) (string, bool) {
	switch msgID {
	case 100:
		return "control_ping", true
	case 101:
		return "control_ping_reply", true
	case 200:
		return "sw_interface_dump", true
	case 201:
		return "sw_interface_details", true
	}

	a.access.Lock()
	defer a.access.Unlock()
	a.initMaps()
	msgName, found := a.msgIDsToName[msgID]

	return msgName, found
}

// RegisterBinAPITypes registers binary API message types in the mock adapter.
func (a *VppAdapter) RegisterBinAPITypes(binAPITypes map[string]reflect.Type) {
	a.access.Lock()
	defer a.access.Unlock()
	a.initMaps()
	for _, v := range binAPITypes {
		if msg, ok := reflect.New(v).Interface().(api.Message); ok {
			a.binAPITypes[msg.GetMessageName()] = v
		}
	}
}

// ReplyTypeFor returns reply message type for given request message name.
func (a *VppAdapter) ReplyTypeFor(requestMsgName string) (reflect.Type, uint16, bool) {
	replyName, foundName := binapi.ReplyNameFor(requestMsgName)
	if foundName {
		if reply, found := a.binAPITypes[replyName]; found {
			msgID, err := a.GetMsgID(replyName, "")
			if err == nil {
				return reply, msgID, found
			}
		}
	}

	return nil, 0, false
}

// ReplyFor returns reply message for given request message name.
func (a *VppAdapter) ReplyFor(requestMsgName string) (api.Message, uint16, bool) {
	replType, msgID, foundReplType := a.ReplyTypeFor(requestMsgName)
	if foundReplType {
		msgVal := reflect.New(replType)
		if msg, ok := msgVal.Interface().(api.Message); ok {
			log.Println("FFF ", replType, msgID, foundReplType)
			return msg, msgID, true
		}
	}

	return nil, 0, false
}

// ReplyBytes encodes the mocked reply into binary format.
func (a *VppAdapter) ReplyBytes(request MessageDTO, reply api.Message) ([]byte, error) {
	replyMsgID, err := a.GetMsgID(reply.GetMessageName(), reply.GetCrcString())
	if err != nil {
		log.Println("ReplyBytesE ", replyMsgID, " ", reply.GetMessageName(), " clientId: ", request.ClientID,
			" ", err)
		return nil, err
	}
	log.Println("ReplyBytes ", replyMsgID, " ", reply.GetMessageName(), " clientId: ", request.ClientID)

	buf := new(bytes.Buffer)
	struc.Pack(buf, &core.VppReplyHeader{VlMsgID: replyMsgID, Context: request.ClientID})
	struc.Pack(buf, reply)

	return buf.Bytes(), nil
}

// GetMsgID returns mocked message ID for the given message name and CRC.
func (a *VppAdapter) GetMsgID(msgName string, msgCrc string) (uint16, error) {
	switch msgName {
	case "control_ping":
		return 100, nil
	case "control_ping_reply":
		return 101, nil
	case "sw_interface_dump":
		return 200, nil
	case "sw_interface_details":
		return 201, nil
	}

	a.access.Lock()
	defer a.access.Unlock()
	a.initMaps()

	msgID, found := a.msgNameToIds[msgName]
	if found {
		return msgID, nil
	}

	a.msgIDSeq++
	msgID = a.msgIDSeq
	a.msgNameToIds[msgName] = msgID
	a.msgIDsToName[msgID] = msgName

	log.Println("VPP GetMessageId ", msgID, " name:", msgName, " crc:", msgCrc)

	return msgID, nil
}

// initMaps initializes internal maps (if not already initialized).
func (a *VppAdapter) initMaps() {
	if a.msgIDsToName == nil {
		a.msgIDsToName = map[uint16]string{}
		a.msgNameToIds = map[string]uint16{}
		a.msgIDSeq = 1000
	}

	if a.binAPITypes == nil {
		a.binAPITypes = map[string]reflect.Type{}
	}
}

// SendMsg emulates sending a binary-encoded message to VPP.
func (a *VppAdapter) SendMsg(clientID uint32, data []byte) error {
	switch mode {
	case useReplyHandlers:
		a.initMaps()
		for i := len(replyHandlers) - 1; i >= 0; i-- {
			replyHandler := replyHandlers[i]

			buf := bytes.NewReader(data)
			reqHeader := core.VppRequestHeader{}
			struc.Unpack(buf, &reqHeader)

			a.access.Lock()
			reqMsgName, _ := a.msgIDsToName[reqHeader.VlMsgID]
			a.access.Unlock()

			reply, msgID, finished := replyHandler(MessageDTO{reqHeader.VlMsgID, reqMsgName,
				clientID, data})
			if finished {
				a.callback(clientID, msgID, reply)
				return nil
			}
		}
		fallthrough
	case useRepliesQueue:
		repliesLock.Lock()
		defer repliesLock.Unlock()

		// pop all replies from queue
		for i, reply := range replies {
			if i > 0 && reply.GetMessageName() == "control_ping_reply" {
				// hack - do not send control_ping_reply immediately, leave it for the the next callback
				replies = []api.Message{}
				replies = append(replies, reply)
				return nil
			}
			msgID, _ := a.GetMsgID(reply.GetMessageName(), reply.GetCrcString())
			buf := new(bytes.Buffer)
			if reply.GetMessageType() == api.ReplyMessage {
				struc.Pack(buf, &core.VppReplyHeader{VlMsgID: msgID, Context: clientID})
			} else if reply.GetMessageType() == api.EventMessage {
				struc.Pack(buf, &core.VppEventHeader{VlMsgID: msgID, Context: clientID})
			} else if reply.GetMessageType() == api.RequestMessage {
				struc.Pack(buf, &core.VppRequestHeader{VlMsgID: msgID, Context: clientID})
			} else {
				struc.Pack(buf, &core.VppOtherHeader{VlMsgID: msgID})
			}
			struc.Pack(buf, reply)
			a.callback(clientID, msgID, buf.Bytes())
		}
		if len(replies) > 0 {
			replies = []api.Message{}
			if len(replyHandlers) > 0 {
				// Switch back to handlers once the queue is empty to revert back
				// the fallthrough effect.
				mode = useReplyHandlers
			}
			return nil
		}

		//fallthrough
	default:
		// return default reply
		buf := new(bytes.Buffer)
		msgID := uint16(defaultReplyMsgID)
		struc.Pack(buf, &core.VppReplyHeader{VlMsgID: msgID, Context: clientID})
		struc.Pack(buf, &defaultReply{})
		a.callback(clientID, msgID, buf.Bytes())
	}
	return nil
}

// SetMsgCallback sets a callback function that will be called by the adapter whenever a message comes from the mock.
func (a *VppAdapter) SetMsgCallback(cb func(context uint32, msgID uint16, data []byte)) {
	a.callback = cb
}

// WaitReady mocks waiting for VPP
func (a *VppAdapter) WaitReady() error {
	return nil
}

// MockReply stores a message to be returned when the next request comes. It is a FIFO queue - multiple replies
// can be pushed into it, the first one will be popped when some request comes.
// Using of this method automatically switches the mock into th useRepliesQueue mode.
func (a *VppAdapter) MockReply(msg api.Message) {
	repliesLock.Lock()
	defer repliesLock.Unlock()

	replies = append(replies, msg)
	mode = useRepliesQueue
}

// MockReplyHandler registers a handler function that is supposed to generate mock responses to incoming requests.
// Using of this method automatically switches the mock into th useReplyHandlers mode.
func (a *VppAdapter) MockReplyHandler(replyHandler ReplyHandler) {
	repliesLock.Lock()
	defer repliesLock.Unlock()

	replyHandlers = append(replyHandlers, replyHandler)
	mode = useReplyHandlers
}
