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
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/dynamicpb"

	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

// Register registers model in DefaultRegistry.
func Register(x any, spec Spec, opts ...ModelOption) KnownModel {
	model, err := DefaultRegistry.Register(x, spec, opts...)
	if err != nil {
		panic(err)
	}
	return model
}

// RegisterModelInfos registers models in the form of ModelInfo in the DefaultRegistry.
// It returns slice of known models that were actually newly registered (didn't exist before in DefaultRegistry).
func RegisterModelInfos(modelInfos []*ModelInfo) []KnownModel {
	var knownModels []KnownModel
	for _, mi := range modelInfos {
		msg := dynamicpb.NewMessageType(mi.MessageDescriptor).New().Interface()
		spec := ToSpec(mi.Spec)
		t, _ := ModelOptionFor("nameTemplate", mi.GetOptions())
		km, err := DefaultRegistry.Register(msg, spec, WithNameTemplate(t))
		if err != nil {
			// model registration failed, try registering remaining model infos
			continue
		}
		knownModels = append(knownModels, km)
	}
	return knownModels
}

// RegisteredModels returns models registered in the DefaultRegistry.
func RegisteredModels() []KnownModel {
	return DefaultRegistry.RegisteredModels()
}

// GetModel returns registered model for given model name.
func GetModel(name string) (KnownModel, error) {
	return GetModelFromRegistry(name, DefaultRegistry)
}

// GetModelFromRegistry returns registered model in given registry for given model name.
func GetModelFromRegistry(name string, modelRegistry Registry) (KnownModel, error) {
	return modelRegistry.GetModel(name)
}

// GetModelFor returns model registered in DefaultRegistry for given proto message.
func GetModelFor(x proto.Message) (KnownModel, error) {
	return GetModelFromRegistryFor(x, DefaultRegistry)
}

// GetModelFromRegistryFor returns model registered in modelRegistry for given proto message
func GetModelFromRegistryFor(x proto.Message, modelRegistry Registry) (KnownModel, error) {
	return modelRegistry.GetModelFor(x)
}

// GetModelForKey returns model registered in DefaultRegistry which matches key.
func GetModelForKey(key string) (KnownModel, error) {
	return DefaultRegistry.GetModelForKey(key)
}

// Key is a helper for the GetKey which panics on errors.
func Key(x proto.Message) string {
	key, err := GetKey(x)
	if err != nil {
		panic(err)
	}
	return key
}

// Name is a helper for the GetName which panics on errors.
func Name(x proto.Message) string {
	name, err := GetName(x)
	if err != nil {
		panic(err)
	}
	return name
}

// GetKey returns complete key for given model,
// including key prefix defined by model specification.
// It returns error if given model is not registered.
func GetKey(x proto.Message) (string, error) {
	return GetKeyUsingModelRegistry(x, DefaultRegistry)
}

// GetKeyUsingModelRegistry returns complete key for given model from given model registry,
// including key prefix defined by model specification.
// It returns error if given model is not registered.
func GetKeyUsingModelRegistry(message proto.Message, modelRegistry Registry) (string, error) {
	// find model for message
	model, err := GetModelFromRegistryFor(message, modelRegistry)
	if err != nil {
		return "", fmt.Errorf("cannot find known model "+
			"for message (while getting key for model) due to: %w (message = %+v)", err, message)
	}

	// compute Item.ID.Name
	name, err := model.InstanceName(message)
	if err != nil {
		return "", fmt.Errorf("cannot compute model instance name due to: %v (message %+v)", err, message)
	}

	key := path.Join(model.KeyPrefix(), name)
	return key, nil
}

// GetName returns instance name for given model.
// It returns error if given model is not registered.
func GetName(x proto.Message) (string, error) {
	model, err := GetModelFor(x)
	if err != nil {
		return "", err
	}
	name, err := model.InstanceName(x)
	if err != nil {
		return "", err
	}
	return name, nil
}

// keyPrefix computes correct key prefix from given model. It
// handles correctly the case when name suffix of the key is empty
// (no template name -> key prefix does not end with "/")
func keyPrefix(modelSpec Spec, hasTemplateName bool) string {
	keyPrefix := modelSpec.KeyPrefix()
	if !hasTemplateName {
		keyPrefix = strings.TrimSuffix(keyPrefix, "/")
	}
	return keyPrefix
}

// upperFirst converts the first letter of string to upper case
func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

func resolveDynamicProtoModelName(msg *dynamicpb.Message) (any, error) {
	model, err := GetModelFor(msg)
	if err != nil {
		return nil, fmt.Errorf("can't get model "+
			"for dynamic message due to: %w (message=%v)", err, msg)
	}
	marshaller := protojson.MarshalOptions{
		EmitUnpopulated: true,
	}
	jsonData, err := marshaller.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("can't marshall message "+
			"to json due to: %w (message: %+v)", err, msg)
	}
	goType := model.LocalGoType()
	if goType != nil {
		pb := model.NewInstance()
		if err := protojson.Unmarshal(jsonData, pb); err != nil {
			return nil, fmt.Errorf("can't load json of marshalled "+
				"message to new model instance due to: %w (json=%v)", err, jsonData)
		}
		return pb, nil
	} else {
		var mapData map[string]any
		if err := json.Unmarshal(jsonData, &mapData); err != nil {
			return nil, fmt.Errorf("can't load json of marshalled "+
				"message to generic map due to: %w (json=%v)", err, jsonData)
		}
		return mapData, nil
	}
}

// DynamicLocallyKnownMessageToGeneratedMessage converts locally registered/known proto dynamic message to
// corresponding statically generated proto message. This function will fail when there is no registration
// of statically-generated proto message, i.e. dynamic message refers to remotely known model.
// This conversion method should help handling dynamic proto messages in mostly protoc-generated proto message
// oriented codebase (i.e. help for type conversions to named, help handle missing data fields as seen
// in generated proto messages,...)
func DynamicLocallyKnownMessageToGeneratedMessage(dynamicMessage *dynamicpb.Message) (proto.Message, error) {
	// get go type of statically generated proto message corresponding to locally known dynamic message
	model, err := GetModelFor(dynamicMessage)
	if err != nil {
		return nil, fmt.Errorf("can't get model "+
			"for dynamic message due to: %w (message=%v)", err, dynamicMessage)
	}
	goType := model.LocalGoType() // only for locally known models will return meaningful go type
	if goType == nil {
		return nil, fmt.Errorf("dynamic messages for remote models are not supported due to "+
			"not available go type of statically generated proto message (dynamic message=%v)", dynamicMessage)
	}

	// create empty statically-generated proto message of the same type as it was used for registration
	var registeredGoType interface{}
	if goType.Kind() == reflect.Ptr {
		registeredGoType = reflect.New(goType.Elem()).Interface()
	} else {
		registeredGoType = reflect.Zero(goType).Interface()
	}

	message := protoMessageOf(registeredGoType)

	// fill empty statically-generated proto message with data from its dynamic proto message counterpart
	// (alternative approach to this is marshalling dynamicMessage to json and unmarshalling it back to message)
	proto.Merge(message, dynamicMessage)

	return message, nil
}

// ModelOptionFor extracts value for given model detail option key
func ModelOptionFor(key string, options []*generic.ModelDetail_Option) (string, error) {
	for _, option := range options {
		if option.Key == key {
			if len(option.Values) == 0 {
				return "", fmt.Errorf("there is no value for key %v in model options", key)
			}
			if strings.TrimSpace(option.Values[0]) == "" {
				return "", fmt.Errorf("there is no value(only empty string "+
					"after trimming) for key %v in model options", key)
			}
			return option.Values[0], nil
		}
	}
	return "", fmt.Errorf("there is no model option with key %v (model options=%+v))", key, options)
}

func protoMessageOf(m interface{}) protoreflect.ProtoMessage {
	return protoimpl.X.ProtoMessageV2Of(m)
}
