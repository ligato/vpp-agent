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

// This constant is used as prefix for TypeUrl when marshalling to Any.
const ligatoUrl = "models.ligato.io/"

// Unmarshal is helper function for unmarshalling items.
func UnmarshalItem(m *api.Item) (proto.Message, error) {
	protoName, err := types.AnyMessageName(m.GetValue().GetAny())
	if err != nil {
		return nil, err
	}
	_, found := registeredModels[protoName]
	if !found {
		return nil, fmt.Errorf("model %s is not registered as model", protoName)
	}

	var any types.DynamicAny
	if err := types.UnmarshalAny(m.GetValue().GetAny(), &any); err != nil {
		return nil, err
	}
	return any.Message, nil
}

// Marshal is helper function for marshalling items.
func MarshalItem(pb proto.Message) (*api.Item, error) {
	protoName := proto.MessageName(pb)
	_, found := registeredModels[protoName]
	if !found {
		return nil, fmt.Errorf("proto %s is not registered as model", protoName)
	}

	any, err := types.MarshalAny(pb)
	if err != nil {
		return nil, err
	}
	any.TypeUrl = ligatoUrl + proto.MessageName(pb)

	object := &api.Item{
		Key: Key(pb),
		Value: &api.Value{
			Any: any,
		},
	}
	return object, nil
}

// RegisteredModels returns all registered modules.
func RegisteredModels() (models []*api.Model) {
	for _, s := range registeredModels {
		models = append(models, &api.Model{
			Module:  s.Module,
			Type:    s.Type,
			Version: s.Version,
			Meta: map[string]string{
				"nameTemplate": s.NameTemplate,
				"protoName":    s.protoName,
				"modelPath":    s.modelPath,
			},
		})
	}
	return
}
