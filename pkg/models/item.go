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
	"path"

	"github.com/golang/protobuf/proto"
	types "github.com/golang/protobuf/ptypes"

	api "github.com/ligato/vpp-agent/api/genericmanager"
)

// This constant is used as prefix for TypeUrl when marshalling to Any.
const ligatoModels = "models.ligato.io/"

// Unmarshal is helper function for unmarshalling items.
func UnmarshalItem(item *api.Item) (proto.Message, error) {
	_, err := GetModelForItem(item)
	if err != nil {
		return nil, err
	}

	var any types.DynamicAny
	if err := types.UnmarshalAny(item.GetData().GetAny(), &any); err != nil {
		return nil, err
	}
	return any.Message, nil
}

// Marshal is helper function for marshalling items.
func MarshalItem(pb proto.Message) (*api.Item, error) {
	model, err := GetModelFor(pb)
	if err != nil {
		return nil, err
	}
	name, err := model.name(pb)
	if err != nil {
		return nil, err
	}

	any, err := types.MarshalAny(pb)
	if err != nil {
		return nil, err
	}
	any.TypeUrl = ligatoModels + proto.MessageName(pb)

	item := &api.Item{
		Id: &api.Item_ID{
			Model: &api.Model{
				Module:  model.Module,
				Version: model.Version,
				Type:    model.Type,
			},
			Name: name,
		},
		Data: &api.Data{
			Any: any,
		},
	}
	return item, nil
}

type model interface {
	GetModule() string
	GetVersion() string
	GetType() string
}

func getModelPath(m model) string {
	return buildModelPath(m.GetVersion(), m.GetModule(), m.GetType())
}

// GetModelForItem returns model for given item.
func GetModelForItem(item *api.Item) (Model, error) {
	if item.GetId() == nil {
		return Model{}, fmt.Errorf("item id is nil")
	}
	itemModel := item.GetId().GetModel()
	model, err := GetModel(getModelPath(itemModel))
	if err != nil {
		return Model{}, err
	}
	if itemModel.GetModule() != model.Module ||
		itemModel.GetVersion() != model.Version ||
		itemModel.GetType() != model.Type {
		return Model{}, fmt.Errorf("item model does not match the one registered (%+v)", itemModel)
	}
	// TODO: check prefix in type url?
	return model, nil
}

// GetKeyForItem returns key for given item.
func GetKeyForItem(item *api.Item) (string, error) {
	model, err := GetModelForItem(item)
	if err != nil {
		return "", err
	}
	key := path.Join(model.keyPrefix, item.GetId().GetName())
	return key, nil
}
