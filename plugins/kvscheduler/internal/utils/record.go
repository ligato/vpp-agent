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

package utils

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// RecordedProtoMessage is a proto.Message suitable for recording and access via
// REST API.
type RecordedProtoMessage struct {
	proto.Message
	ProtoMsgName string
}

// ProtoWithName is used to marshall proto message data alongside the proto
// message name.
type ProtoWithName struct {
	ProtoMsgName string
	ProtoMsgData string
}

// MarshalJSON marshalls proto message using the marshaller from jsonpb.
// The jsonpb package produces a different output than the standard "encoding/json"
// package, which does not operate correctly on protocol buffers.
func (p *RecordedProtoMessage) MarshalJSON() ([]byte, error) {
	var (
		msgName string
		msgData string
		err     error
	)
	if p != nil {
		msgName = proto.MessageName(p.Message)
		marshaller := &jsonpb.Marshaler{}
		msgData, err = marshaller.MarshalToString(p.Message)
		if err != nil {
			return nil, err
		}
	}
	pwn, err := json.Marshal(
		ProtoWithName{ProtoMsgName: msgName, ProtoMsgData: msgData})
	if err != nil {
		return nil, err
	}
	return pwn, nil
}

// UnmarshalJSON un-marshalls proto message using the marshaller from jsonpb.
// The jsonpb package produces a different output than the standard "encoding/json"
// package, which does not operate correctly on protocol buffers.
func (p *RecordedProtoMessage) UnmarshalJSON(data []byte) error {
	pwn := ProtoWithName{}
	err := json.Unmarshal(data, &pwn)
	if err != nil {
		return err
	}
	p.ProtoMsgName = pwn.ProtoMsgName
	if p.ProtoMsgName == "" {
		return nil
	}
	msgType := proto.MessageType(pwn.ProtoMsgName)
	if msgType == nil {
		return fmt.Errorf("unknown proto message: %s", p.ProtoMsgName)
	}
	msg := reflect.New(msgType.Elem()).Interface().(proto.Message)
	err = jsonpb.UnmarshalString(pwn.ProtoMsgData, msg)
	if err != nil {
		return err
	}
	p.Message = msg
	return nil
}

// RecordProtoMessage prepares proto message for recording and potential
// access via REST API.
// Note: no need to clone the message - once un-marshalled, the content is never
// changed (otherwise it would break prev-new value comparisons).
func RecordProtoMessage(msg proto.Message) *RecordedProtoMessage {
	if msg == nil {
		return nil
	}
	return &RecordedProtoMessage{
		Message:      msg,
		ProtoMsgName: proto.MessageName(msg),
	}
}
