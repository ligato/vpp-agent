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

package keyval

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
)

// Serializer is responsible for transformation of data stored in etcd.
type Serializer interface {
	Unmarshal(data []byte, protoData proto.Message) error
	Marshal(message proto.Message) ([]byte, error)
}

// SerializerProto serializes proto message using proto serializer
type SerializerProto struct{}

// SerializerJSON serialize proto message using json serializer
type SerializerJSON struct{}

// Unmarshal unmarshals data from slice of bytes into the provided protobuf message
func (sp *SerializerProto) Unmarshal(data []byte, protoData proto.Message) error {
	return proto.Unmarshal(data, protoData)
}

// Marshal transforms data from proto message to the slice of bytes using proto marshaller
func (sp *SerializerProto) Marshal(message proto.Message) ([]byte, error) {
	return proto.Marshal(message)
}

// Unmarshal unmarshals data from slice of bytes into the provided protobuf message
func (sj *SerializerJSON) Unmarshal(data []byte, protoData proto.Message) error {
	return json.Unmarshal(data, protoData)
}

// Marshal marshals proto message using json marshaller
func (sj *SerializerJSON) Marshal(message proto.Message) ([]byte, error) {
	return json.Marshal(message)
}
