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

// MarshalItem is helper function for marshalling model instance into item
func MarshalItem(pb proto.Message) (*api.Item, error) {
	return MarshalItemUsingModelRegistry(pb, DefaultRegistry)
}

// MarshalItemUsingModelRegistry is helper function for marshalling model instance
// into item (using given model registry)
func MarshalItemUsingModelRegistry(pb proto.Message, modelRegistry Registry) (*api.Item, error) {
	model, err := GetModelFromRegistryFor(pb, modelRegistry)
	if err != nil {
		return nil, errors.Errorf("can't find known model "+
			"for message due to: %v (message = %+v)", err, pb)
	}
	name, err := model.InstanceName(pb)
	if err != nil {
		return nil, errors.Errorf("can't compute model instance name due to: %v (message %+v)", err, pb)
	}

	any, err := types.MarshalAny(pb)
	if err != nil {
		return nil, err
	}
	any.TypeUrl = ligatoModels + string(proto.MessageV2(pb).ProtoReflect().Descriptor().FullName())

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

// UnmarshalItem is helper function for unmarshalling items.
func UnmarshalItem(item *api.Item) (proto.Message, error) {
	return UnmarshalItemUsingModelRegistry(item, DefaultRegistry)
}

// UnmarshalItemUsingModelRegistry is helper function for unmarshalling items (using given model registry)
func UnmarshalItemUsingModelRegistry(item *api.Item, modelRegistry Registry) (proto.Message, error) {
	// check existence of known model
	_, err := GetModelFromModelRegistryForItem(item, modelRegistry)
	if err != nil {
		return nil, err
	}

	// unmarshal item's inner data
	// (we must distinguish between model registries due to different go types produced by using different
	// model registry. The LocalRegistry use cases need go types as generated from models, but
	// the RemoteRegistry can't produce such go typed instances (we know the name of go type, but can't
	// produce it from remote information) so dynamic proto message must be enough (*dynamicpb.Message))
	if _, ok := modelRegistry.(*LocalRegistry); ok {
		return unmarshalItemDataAnyOfLocalModel(item.GetData().GetAny())
	}
	return unmarshalItemDataAnyOfRemoteModel(item.GetData().GetAny(), modelRegistry.MessageTypeRegistry())
}

// unmarshalItemDataAnyOfRemoteModel unmarshalls the generic data part of api.Item that has remote model.
// The unmarshalled proto.Message will have dynamic type (*dynamicpb.Message).
func unmarshalItemDataAnyOfRemoteModel(itemAny *any.Any, msgTypeResolver *protoregistry.Types) (proto.Message, error) {
	msg, err := anypb.UnmarshalNew(itemAny, protoV2.UnmarshalOptions{
		Resolver: msgTypeResolver,
	})
	if err != nil {
		return nil, err
	}
	return proto.MessageV1(msg), nil
}

// unmarshalItemDataAnyOfLocalModel unmarshalls the generic data part of api.Item that has local model.
// The unmarshalled proto.Message will have the go type of model generated go structures (that is due to
// go type registering in init() method of generated go structures file).
func unmarshalItemDataAnyOfLocalModel(itemAny *any.Any) (proto.Message, error) {
	var any types.DynamicAny // local
	if err := types.UnmarshalAny(itemAny, &any); err != nil {
		return nil, err
	}
	return any.Message, nil
}

// GetModelForItem returns model for given item.
func GetModelForItem(item *api.Item) (KnownModel, error) {
	return GetModelFromModelRegistryForItem(item, DefaultRegistry)
}

// GetModelFromModelRegistryForItem returns model for given item (using given model registry)
func GetModelFromModelRegistryForItem(item *api.Item, modelRegistry Registry) (KnownModel, error) {
	if item.GetId() == nil {
		return nil, fmt.Errorf("item id is nil")
	}
	modelPath := item.GetId().GetModel()
	model, err := GetModelFromRegistry(modelPath, modelRegistry)
	if err != nil {
		return nil, fmt.Errorf("can't find modelpath %v in provided "+
			"models %+v", modelPath, modelRegistry)
	}
	// TODO: check prefix in type url?
	return model, nil
}

// GetKeyForItem returns key for given item.
func GetKeyForItem(item *api.Item) (string, error) {
	return GetKeyForItemUsingModelRegistry(item, DefaultRegistry)
}

// GetKeyForItem returns key for given item (using given model registry)
func GetKeyForItemUsingModelRegistry(item *api.Item, modelRegistry Registry) (string, error) {
	model, err := GetModelFromModelRegistryForItem(item, modelRegistry)
	if err != nil {
		return "", err
	}
	key := path.Join(model.KeyPrefix(), item.GetId().GetName())
	return key, nil
}
