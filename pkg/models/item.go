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
	"strings"

	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	types "github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	api "go.ligato.io/vpp-agent/v3/proto/ligato/generic"
	protoV2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
)

// This constant is used as prefix for TypeUrl when marshalling to Any.
const ligatoModels = "models.ligato.io/"

// Marshal is helper function for marshalling model instance into item.
func MarshalItem(pb proto.Message) (*api.Item, error) {
	model, err := GetModelFor(pb)
	if err != nil {
		return nil, err
	}
	name, err := model.instanceName(pb)
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
			Model: model.Name(),
			Name:  name,
		},
		Data: &api.Data{
			Union: &api.Data_Any{Any: any},
		},
	}
	return item, nil
}

// MarshalItemWithExternallyKnownModels is helper function for marshalling model instance into item by using
// models that are known only from external sources (=not registered in default model registry that is filled
// by variable initialization of compiled code)
func MarshalItemWithExternallyKnownModels(message proto.Message, externallyKnownModels []*api.ModelDetail) (*api.Item, error) {
	// find model for message
	knownModel, err := GetExternallyKnownModelFor(message, externallyKnownModels)
	if err != nil {
		return nil, errors.Errorf("can't find externally known model "+
			"for message due to: %v (message = %+v)", err, message)
	}

	// compute Item.ID.Name
	messageDesc := proto.MessageV2(message).ProtoReflect().Descriptor()
	messageFullName := string(messageDesc.FullName())
	name, err := instanceNameWithExternallyKnownModel(message, knownModel, messageDesc)
	if err != nil {
		return nil, errors.Errorf("can't compute model instance name due to: %v (message %+v)", err, message)
	}

	// convert message itself
	any, err := types.MarshalAny(message)
	if err != nil {
		return nil, err
	}
	any.TypeUrl = ligatoModels + messageFullName

	// create Item
	item := &api.Item{
		Id: &api.Item_ID{
			Model: fmt.Sprintf("%v.%v", knownModel.Spec.Module, knownModel.Spec.Type),
			Name:  name,
		},
		Data: &api.Data{
			Union: &api.Data_Any{Any: any},
		},
	}
	return item, nil
}

// UnmarshalItem is helper function for unmarshalling items.
func UnmarshalItem(item *api.Item) (proto.Message, error) {
	// check existence of locally registered model
	_, err := GetModelForItem(item)
	if err != nil {
		return nil, err
	}
	return unmarshalItemDataAnyV1(item.GetData().GetAny())
}

// UnmarshalItemWithExternallyKnownModels is helper function for unmarshalling items corresponding to only
// externally known models (= not present in local model registry)
func UnmarshalItemWithExternallyKnownModels(item *api.Item, externallyKnownModels []*api.ModelDetail,
	msgTypeResolver *protoregistry.Types) (proto.Message, error) {
	// check existence of remotely known model
	_, err := GetExternallyKnownModelForItem(item, externallyKnownModels)
	if err != nil {
		return nil, err
	}
	return unmarshalItemDataAny(item.GetData().GetAny(), msgTypeResolver)
}

// unmarshalItemDataAny unmarshalls the generic data part of api.Item (using new protoV2 method)
func unmarshalItemDataAny(itemAny *any.Any, msgTypeResolver *protoregistry.Types) (proto.Message, error) {
	msg, err := anypb.UnmarshalNew(itemAny, protoV2.UnmarshalOptions{
		Resolver: msgTypeResolver,
	})
	if err != nil {
		return nil, err
	}
	return proto.MessageV1(msg), nil
}

// unmarshalItemDataAnyV1 unmarshalls the generic data part of api.Item (using old working protoV1 method)
func unmarshalItemDataAnyV1(itemAny *any.Any) (proto.Message, error) {
	var any types.DynamicAny
	if err := types.UnmarshalAny(itemAny, &any); err != nil {
		return nil, err
	}
	return any.Message, nil
}

// GetModelForItem returns model for given item.
func GetModelForItem(item *api.Item) (KnownModel, error) {
	if item.GetId() == nil {
		return KnownModel{}, fmt.Errorf("item id is nil")
	}
	modelPath := item.GetId().GetModel()
	model, err := GetModel(modelPath)
	if err != nil {
		return KnownModel{}, err
	}
	// TODO: check prefix in type url?
	return model, nil
}

// GetExternallyKnownModelForItem returns model for given item.
func GetExternallyKnownModelForItem(item *api.Item, externallyKnownModels []*api.ModelDetail) (*api.ModelDetail, error) {
	if item.GetId() == nil {
		return nil, fmt.Errorf("item id is nil")
	}
	modelPath := item.GetId().GetModel()
	for _, model := range externallyKnownModels {
		externalModelPath := fmt.Sprintf("%v.%v", model.Spec.Module, model.Spec.Type)
		if externalModelPath == modelPath {
			return model, nil
		}
	}
	return nil, fmt.Errorf("can't find modelpath %v in provided "+
		"external models %+v", modelPath, externallyKnownModels)
}

// GetKeyForItem returns key for given item.
func GetKeyForItem(item *api.Item) (string, error) {
	model, err := GetModelForItem(item)
	if err != nil {
		return "", err
	}
	key := path.Join(model.KeyPrefix(), item.GetId().GetName())
	return key, nil
}

// GetKeyForItemWithExternallyKnownModels returns key for given item.
func GetKeyForItemWithExternallyKnownModels(item *api.Item, externallyKnownModels []*api.ModelDetail) (string, error) {
	model, err := GetExternallyKnownModelForItem(item, externallyKnownModels)
	if err != nil {
		return "", err
	}
	name := item.GetId().GetName()
	key := path.Join(keyPrefix(ToSpec(model.Spec), name != ""), name)
	return key, nil
}

// modelOptionFor retrieves first value for given key in model detail options
func modelOptionFor(key string, options []*api.ModelDetail_Option) (string, error) {
	for _, option := range options {
		if option.Key == key {
			if len(option.Values) == 0 {
				return "", errors.Errorf("there is no value for key %v in model options", key)
			}
			if strings.TrimSpace(option.Values[0]) == "" {
				return "", errors.Errorf("there is no value(only empty string "+
					"after trimming) for key %v in model options", key)
			}
			return option.Values[0], nil
		}
	}
	return "", errors.Errorf("there is no model option with key %v (model options=%+v))", key, options)
}
