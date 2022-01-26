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

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
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

// MarshalJSON marshalls proto message using the marshaller from protojson.
// The protojson package produces a different output than the standard "encoding/json"
// package, which does not operate correctly on protocol buffers.
func (p *RecordedProtoMessage) MarshalJSON() ([]byte, error) {
	var (
		msgName string
		msgData string
		err     error
	)
	if p != nil {
		msgName = string(proto.MessageName(p.Message))
		b, err := prototext.Marshal(p.Message)
		if err != nil {
			return nil, err
		}
		msgData = string(b)
	}
	pwn, err := json.Marshal(ProtoWithName{
		ProtoMsgName: msgName,
		ProtoMsgData: msgData,
	})
	if err != nil {
		return nil, err
	}
	return pwn, nil
}

// UnmarshalJSON un-marshalls proto message using the marshaller from protojson.
// The protojson package produces a different output than the standard "encoding/json"
// package, which does not operate correctly on protocol buffers.
func (p *RecordedProtoMessage) UnmarshalJSON(data []byte) error {
	var pwn ProtoWithName
	if err := json.Unmarshal(data, &pwn); err != nil {
		return err
	}
	p.ProtoMsgName = pwn.ProtoMsgName
	if p.ProtoMsgName == "" {
		return nil
	}
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(pwn.ProtoMsgName))
	if err != nil {
		return err
	}
	msg := msgType.New().Interface()
	if len(pwn.ProtoMsgData) > 0 && pwn.ProtoMsgData[0] == '{' {
		err = protojson.Unmarshal([]byte(pwn.ProtoMsgData), msg)
	} else {
		err = prototext.Unmarshal([]byte(pwn.ProtoMsgData), msg)
	}
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
		ProtoMsgName: string(proto.MessageName(msg)),
	}
}
