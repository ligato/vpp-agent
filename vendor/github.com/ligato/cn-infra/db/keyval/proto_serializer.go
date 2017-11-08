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

// Serializer is used to make conversions between raw and formatted data.
// Currently supported formats are JSON and protobuf.
type Serializer interface {
	Unmarshal(data []byte, protoData proto.Message) error
	Marshal(message proto.Message) ([]byte, error)
}

// SerializerProto serializes proto message using proto serializer.
type SerializerProto struct{}

// SerializerJSON serializes proto message using JSON serializer.
type SerializerJSON struct{}

// Unmarshal deserializes data from slice of bytes into the provided protobuf
// message using proto marshaller.
func (sp *SerializerProto) Unmarshal(data []byte, protoData proto.Message) error {
	return proto.Unmarshal(data, protoData)
}

// Marshal serializes data from proto message to the slice of bytes using proto
// marshaller.
func (sp *SerializerProto) Marshal(message proto.Message) ([]byte, error) {
	return proto.Marshal(message)
}

// Unmarshal deserializes data from slice of bytes into the provided protobuf
// message using JSON marshaller.
func (sj *SerializerJSON) Unmarshal(data []byte, protoData proto.Message) error {
	return json.Unmarshal(data, protoData)
}

// Marshal serializes proto message to the slice of bytes using json marshaller.
func (sj *SerializerJSON) Marshal(message proto.Message) ([]byte, error) {
	return json.Marshal(message)
}
