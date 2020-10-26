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
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-errors/errors"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	types "github.com/golang/protobuf/ptypes"
	api "go.ligato.io/vpp-agent/v3/proto/ligato/generic"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	messageDesc := proto.MessageV2(message).ProtoReflect().Descriptor()
	messageFullName := string(messageDesc.FullName())
	var knownModel *api.ModelDetail
	for _, ekm := range externallyKnownModels {
		if ekm.ProtoName == messageFullName {
			knownModel = ekm
			break
		}
	}
	if knownModel == nil {
		return nil, errors.Errorf("can't find externally known model for message %v "+
			"(All externally known models: %#v)", messageFullName, externallyKnownModels)
	}

	// compute Item.ID.Name from name template
	nameTemplate, err := modelOptionFor("nameTemplate", knownModel.Options)
	if err != nil {
		return nil, errors.Errorf("can't get name template from model options "+
			"from externally known model %v due to: %v", knownModel.ProtoName, err)
	}
	nameTemplate = replaceFieldNamesInNameTemplate(messageDesc, nameTemplate)
	marshaler := jsonpb.Marshaler{EmitDefaults: true} // using jsonbp to generate json with json name field in proto tag
	jsonData, err := marshaler.MarshalToString(message)
	if err != nil {
		return nil, errors.Errorf("can't marshall message "+
			"to json due to: %v (message: %+v)", err, message)
	}
	var mapData map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &mapData)
	if err != nil {
		return nil, errors.Errorf("can't load json of marshalled "+
			"message to generic map due to: %v (json=%v)", err, jsonData)
	}
	name, err := NameTemplate(nameTemplate)(mapData)
	if err != nil {
		return nil, errors.Errorf("can't compute name from name template by applying generic map "+
			"due to: %v (name template=%v, generic map=%v)", err, nameTemplate, mapData)
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

// GetKeyForItem returns key for given item.
func GetKeyForItem(item *api.Item) (string, error) {
	model, err := GetModelForItem(item)
	if err != nil {
		return "", err
	}
	key := path.Join(model.KeyPrefix(), item.GetId().GetName())
	return key, nil
}

// replaceFieldNamesInNameTemplate replaces JSON field names to Go Type field name in name template.
func replaceFieldNamesInNameTemplate(messageDesc protoreflect.MessageDescriptor, nameTemplate string) string {
	// FIXME this is only a good effort to map between NameTemplate variables and Proto model field names
	//  (protoName, jsonName). We can do here better (fix field names prefixing other field names or field
	//  names colliding with field names of inner reference structures), but i the end we are still guessing
	//  without knowledge of go type. Can we fix this?
	for i := 0; i < messageDesc.Fields().Len(); i++ {
		fieldDesc := messageDesc.Fields().Get(i)
		pbJSONName := fieldDesc.JSONName()
		nameTemplate = strings.ReplaceAll(nameTemplate, upperFirst(pbJSONName), pbJSONName)
		if fieldDesc.Message() != nil {
			nameTemplate = replaceFieldNamesInNameTemplate(fieldDesc.Message(), nameTemplate)
		}
	}
	return nameTemplate
}

// upperFirst converts the first letter of string to upper case
func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// modelOptionFor retrieves first value for given key in model detail options
func modelOptionFor(key string, options []*api.ModelDetail_Option) (string, error) {
	for _, option := range options {
		if option.Key == key {
			if len(option.Values) == 0 {
				return "", errors.Errorf("there is no value for key %v in model options", key)
			}
			return option.Values[0], nil
		}
	}
	return "", errors.Errorf("there is no model option with key %v (model options=%+v))", key, options)
}
