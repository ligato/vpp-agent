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

package protoval

import (
	"github.com/gogo/protobuf/proto"
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

// ProtoValue is an interface that value carrying proto message should implement.
type ProtoValue interface {
	Value
	GetProtoMessage() proto.Message
}

// protoValue wraps ProtoMessage to implement the Value interface for use with
// the KVScheduler.
type protoValue struct {
	protoMessage proto.Message
}

// ProtoMessageWithName is based on our convention for proto-defined objects
// to store object label under the attribute "Name".
type ProtoMessageWithName interface {
	GetName() string
}

// NewProtoValue creates a new instance of ProtoValue carrying the given proto
// message.
func NewProtoValue(protoMsg proto.Message) ProtoValue {
	if protoMsg == nil {
		return nil
	}
	return &protoValue{protoMessage: protoMsg}
}

// GetProtoMessage returns the underlying proto message.
func (pv *protoValue) GetProtoMessage() proto.Message {
	return pv.protoMessage
}

// Label tries to read and return "Name" attribute. Without the name attribute,
// the function will return empty string.
func (pv *protoValue) Label() string {
	protoWithName, hasName := pv.protoMessage.(ProtoMessageWithName)
	if hasName {
		return protoWithName.GetName()
	}
	return pv.String()
}

// Equivalent uses proto.Equal for comparison.
func (pv *protoValue) Equivalent(v2 Value) bool {
	v2Proto, ok := v2.(ProtoValue)
	if !ok {
		return false
	}
	return proto.Equal(pv.protoMessage, v2Proto.GetProtoMessage())
}

// String uses the String method from proto.Message.
func (pv *protoValue) String() string {
	return pv.protoMessage.String()
}
