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
	"path"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-errors/errors"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Register registers model in DefaultRegistry.
func Register(pb proto.Message, spec Spec, opts ...ModelOption) *KnownModel {
	model, err := DefaultRegistry.Register(pb, spec, opts...)
	if err != nil {
		panic(err)
	}
	return model
}

// RegisteredModels returns models registered in the DefaultRegistry.
func RegisteredModels() []KnownModel {
	return DefaultRegistry.RegisteredModels()
}

// GetModel returns registered model for given model name.
func GetModel(name string) (KnownModel, error) {
	return DefaultRegistry.GetModel(name)
}

// GetModelFor returns model registered in DefaultRegistry for given proto message.
func GetModelFor(x proto.Message) (KnownModel, error) {
	return DefaultRegistry.GetModelFor(x)
}

// GetExternallyKnownModelFor returns externally known model (from given externallyKnownModels) corresponding
// to given proto message
func GetExternallyKnownModelFor(message proto.Message, externallyKnownModels []*ModelInfo) (*ModelInfo, error) {
	messageDesc := proto.MessageV2(message).ProtoReflect().Descriptor()
	messageFullName := string(messageDesc.FullName())
	var knownModel *ModelInfo
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
	return knownModel, nil
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
	model, err := GetModelFor(x)
	if err != nil {
		return "", err
	}
	name, err := model.instanceName(x)
	if err != nil {
		return "", err
	}
	key := path.Join(model.KeyPrefix(), name)
	return key, nil
}

// GetKeyWithExternallyKnownModels returns complete
// key for given model, including key prefix defined
// by externally known model specification.
func GetKeyWithExternallyKnownModels(message proto.Message, externallyKnownModels []*ModelInfo) (string, error) {
	// find model for message
	knownModel, err := GetExternallyKnownModelFor(message, externallyKnownModels)
	if err != nil {
		return "", errors.Errorf("can't find externally known model "+
			"for message due to: %v (message = %+v)", err, message)
	}

	// compute Item.ID.Name
	name, err := instanceNameWithExternallyKnownModel(message, knownModel)
	if err != nil {
		return "", errors.Errorf("can't compute model instance name due to: %v (message %+v)", err, message)
	}

	key := path.Join(keyPrefix(ToSpec(knownModel.Spec), name != ""), name)
	return key, nil
}

// GetName returns instance name for given model.
// It returns error if given model is not registered.
func GetName(x proto.Message) (string, error) {
	model, err := GetModelFor(x)
	if err != nil {
		return "", err
	}
	name, err := model.instanceName(x)
	if err != nil {
		return "", err
	}
	return name, nil
}

// instanceNameWithExternallyKnownModel computes message name using name template (if present).
// This is the equivalent to models.KnownModel's instanceName(...) using not locally registered
// model but using externally acquired models.
func instanceNameWithExternallyKnownModel(message proto.Message, knownModel *ModelInfo) (string, error) {
	nameTemplate, err := modelOptionFor("nameTemplate", knownModel.Options)
	if err != nil {
		logrus.DefaultLogger().Debugf("no nameTemplate model "+
			"option for model %v, using empty instance name", knownModel.ProtoName)
		return "", nil // having no name template is valid case for some models
	}
	nameTemplate = replaceFieldNamesInNameTemplate(knownModel.MessageDescriptor, nameTemplate)
	marshaler := jsonpb.Marshaler{EmitDefaults: true} // using jsonbp to generate json with json name field in proto tag
	jsonData, err := marshaler.MarshalToString(message)
	if err != nil {
		return "", errors.Errorf("can't marshall message "+
			"to json due to: %v (message: %+v)", err, message)
	}
	var mapData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &mapData); err != nil {
		return "", errors.Errorf("can't load json of marshalled "+
			"message to generic map due to: %v (json=%v)", err, jsonData)
	}
	name, err := NameTemplate(nameTemplate)(mapData)
	if err != nil {
		return "", errors.Errorf("can't compute name from name template by applying generic map "+
			"due to: %v (name template=%v, generic map=%v)", err, nameTemplate, mapData)
	}
	return name, nil
}

// replaceFieldNamesInNameTemplate replaces JSON field names to Go Type field name in name template.
func replaceFieldNamesInNameTemplate(messageDesc protoreflect.MessageDescriptor, nameTemplate string) string {
	// FIXME this is only a good effort to map between NameTemplate variables and Proto model field names
	//  (protoName, jsonName). We can do here better (fix field names prefixing other field names or field
	//  names colliding with field names of inner reference structures), but i the end we are still guessing
	//  without knowledge of go type. Can we fix this?
	//  Try check message type fields for possible information about Go Type field names
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
