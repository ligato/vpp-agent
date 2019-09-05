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
	"github.com/gogo/protobuf/proto"
	prototypes "github.com/gogo/protobuf/types"
)

// ProtoToString converts proto message to string.
func ProtoToString(msg proto.Message) string {
	if recProto, wrapped := msg.(*RecordedProtoMessage); wrapped {
		if recProto != nil {
			msg = recProto.Message
		} else {
			msg = nil
		}
	}
	if msg == nil {
		return "<NIL>"
	}
	if _, isEmpty := msg.(*prototypes.Empty); isEmpty {
		return "<EMPTY>"
	}
	// wrap with curly braces, it is easier to read
	return "{ " + msg.String() + " }"
}

// ErrorToString converts error to string.
func ErrorToString(err error) string {
	if err == nil {
		return "<NIL>"
	}
	return err.Error()
}
