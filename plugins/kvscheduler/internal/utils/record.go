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

	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"go.ligato.io/vpp-agent/v3/pkg/models"
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
	ProtoMsgData json.RawMessage
}

// MarshalJSON marshalls proto message using the marshaller from protojson.
// The protojson package produces a different output than the standard "encoding/json"
// package, which does not operate correctly on protocol buffers.
func (p *RecordedProtoMessage) MarshalJSON() ([]byte, error) {
	var (
		msgName string
		msgData []byte
		err     error
	)
	if p != nil {
		msgName = p.ProtoMsgName
		msgData, err = protojson.Marshal(p.Message)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(&ProtoWithName{
		ProtoMsgName: msgName,
		ProtoMsgData: json.RawMessage(msgData),
	})
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

	// try to find the message type in the default registry
	typeRegistry := models.DefaultRegistry.MessageTypeRegistry()
	fullMsgName := protoreflect.FullName(p.ProtoMsgName)
	msgType, err := typeRegistry.FindMessageByName(fullMsgName)
	if err != nil {
		// if not found use the proto global types registry as a fallback
		logging.Debugf("cannot get message type for message name %s from default registry: %v", fullMsgName, err)
		msgType, err = protoregistry.GlobalTypes.FindMessageByName(fullMsgName)
	}
	if err != nil {
		return err
	}

	msg := msgType.New().Interface()
	if err = protojson.Unmarshal(pwn.ProtoMsgData, msg); err != nil {
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
