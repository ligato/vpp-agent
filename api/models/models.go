//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package models

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

// ProtoModel represents proto.Message that returns model key.
type ProtoModel interface {
	proto.Message
	ModelKey() string
}

// Unmarshal is helper function for unmarshalling model data.
func Unmarshal(m *Model) (proto.Message, error) {
	var any types.DynamicAny
	if err := types.UnmarshalAny(m.Value, &any); err != nil {
		return nil, err
	}
	return any.Message, nil
}

// Marshal is helper function for marshalling into model data.
func Marshal(pm ProtoModel) (*Model, error) {
	any, err := types.MarshalAny(pm)
	if err != nil {
		return nil, err
	}
	return &Model{Key: pm.ModelKey(), Value: any}, nil
}
