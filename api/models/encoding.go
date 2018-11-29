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
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/ligato/vpp-agent/api"
)

// This constant is used to replace the constant from types.MarshalAny.
const LigatoApis = "models.ligato.io/"

// Unmarshal is helper function for unmarshalling model data.
func Unmarshal(m *api.Model) (ProtoModel, error) {
	protoName, err := types.AnyMessageName(m.Any)
	if err != nil {
		return nil, err
	}
	spec := registeredSpecs[protoName]
	if spec == nil {
		return nil, fmt.Errorf("model %s is not registered", protoName)
	} /*else if Spec.Version != m.Version {
		return nil, fmt.Errorf("model %s (%s) is registered with different version: %q",
			protoName, m.Version, Spec.Version)
	}*/

	var any types.DynamicAny
	if err := types.UnmarshalAny(m.Any, &any); err != nil {
		return nil, err
	}
	return any.Message.(ProtoModel), nil
}

// Marshal is helper function for marshalling into model data.
func Marshal(pb ProtoModel) (*api.Model, error) {
	protoName := proto.MessageName(pb)
	spec := registeredSpecs[protoName]
	if spec == nil {
		return nil, fmt.Errorf("model %s is not registered", protoName)
	}

	any, err := types.MarshalAny(pb)
	if err != nil {
		return nil, err
	}
	any.TypeUrl = LigatoApis + proto.MessageName(pb)

	model := &api.Model{
		//Version: Spec.Version,
		Any: any,
	}
	return model, nil
}
