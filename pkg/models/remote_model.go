// Copyright (c) 2020 Pantheon.tech
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

package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-errors/errors"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	api "go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

// RemotelyKnownModel represents a registered remote model (remote model has only information about model
// from remote source, i.e. missing go type because VPP-Agent meta service doesn't provide it)
type RemotelyKnownModel struct {
	model *ModelInfo
}

// Spec returns model specification for the model.
func (m *RemotelyKnownModel) Spec() *Spec {
	spec := ToSpec(m.model.Spec)
	return &spec
}

// ModelDetail returns descriptor for the model.
func (m *RemotelyKnownModel) ModelDetail() *api.ModelDetail {
	return m.model.ModelDetail
}

// NewInstance creates new instance value for model type. Due to missing go type for remote models, the created
// instance won't have the same go type as in case of local models, but dynamic proto message's go type
// (the proto descriptor will be the same).
func (m *RemotelyKnownModel) NewInstance() proto.Message {
	return dynamicpb.NewMessageType(m.model.MessageDescriptor).New().Interface()
}

// ProtoName returns proto message name registered with the model.
func (m *RemotelyKnownModel) ProtoName() string {
	if strings.TrimSpace(m.model.ProtoName) == "" {
		return string(m.model.MessageDescriptor.FullName())
	}
	return m.model.ProtoName
}

// ProtoFile returns proto file name for the model.
func (m *RemotelyKnownModel) ProtoFile() string {
	return m.model.MessageDescriptor.ParentFile().Path()
}

// NameTemplate returns name template for the model.
func (m *RemotelyKnownModel) NameTemplate() string {
	nameTemplate, _ := m.modelOptionFor("nameTemplate", m.model.Options)
	return nameTemplate
}

// GoType returns go type for the model.
func (m *RemotelyKnownModel) GoType() string {
	goType, _ := m.modelOptionFor("goType", m.model.Options)
	return goType
}

// LocalGoType should returns reflect go type for the model, but remotely known model doesn't have
// locally known reflect go type. It always returns nil.
func (m *RemotelyKnownModel) LocalGoType() reflect.Type {
	return nil
}

// PkgPath returns package import path for the model definition.
func (m *RemotelyKnownModel) PkgPath() string {
	pkgPath, _ := m.modelOptionFor("pkgPath", m.model.Options)
	return pkgPath
}

// Name returns name for the model.
func (m *RemotelyKnownModel) Name() string {
	return ToSpec(m.model.Spec).ModelName()
}

// KeyPrefix returns key prefix for the model.
func (m *RemotelyKnownModel) KeyPrefix() string {
	return keyPrefix(ToSpec(m.model.Spec), m.NameTemplate() != "")
}

// ParseKey parses the given key and returns item name
// or returns empty name and valid as false if the key is not valid.
func (m *RemotelyKnownModel) ParseKey(key string) (name string, valid bool) {
	name = strings.TrimPrefix(key, m.KeyPrefix())
	if name == key || (name == "" && m.NameTemplate() != "") {
		name = strings.TrimPrefix(key, m.Name())
	}
	// key had the prefix and also either
	// non-empty name or no name template
	if name != key && (name != "" || m.NameTemplate() == "") {
		// TODO: validate name?
		return name, true
	}
	return "", false
}

// IsKeyValid returns true if given key is valid for this model.
func (m *RemotelyKnownModel) IsKeyValid(key string) bool {
	_, valid := m.ParseKey(key)
	return valid
}

// StripKeyPrefix returns key with prefix stripped.
func (m *RemotelyKnownModel) StripKeyPrefix(key string) string {
	if name, valid := m.ParseKey(key); valid {
		return name
	}
	return key
}

// InstanceName computes message name for given proto message using name template (if present).
func (m *RemotelyKnownModel) InstanceName(x interface{}) (string, error) {
	message := protoMessageOf(x)

	// extract data from message and use them with name template to get the name
	nameTemplate, err := m.modelOptionFor("nameTemplate", m.model.Options)
	if err != nil {
		logrus.DefaultLogger().Debugf("no nameTemplate model "+
			"option for model %v, using empty instance name", m.model.ProtoName)
		return "", nil // having no name template is valid case for some models
	}
	nameTemplate = m.replaceFieldNamesInNameTemplate(m.model.MessageDescriptor, nameTemplate)
	marshaller := protojson.MarshalOptions{
		EmitUnpopulated: true,
	}
	jsonData, err := marshaller.Marshal(message)
	if err != nil {
		return "", errors.Errorf("can't marshall message "+
			"to json due to: %v (message: %+v)", err, message)
	}
	var mapData map[string]interface{}
	if err := json.Unmarshal(jsonData, &mapData); err != nil {
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
func (m *RemotelyKnownModel) replaceFieldNamesInNameTemplate(messageDesc protoreflect.MessageDescriptor, nameTemplate string) string {
	// FIXME this is only a good effort to map between NameTemplate variables and Proto model field names
	//  (protoName, jsonName). We can do here better (fix field names prefixing other field names or field
	//  names colliding with field names of inner reference structures), but i the end we are still guessing
	//  without knowledge of go type. Can we fix this?
	//  (The dynamicpb.NewMessageType(messageDesc) should return MessageType that joins message descriptor and
	//  go type information, but for dynamicpb package the go type means always dynamicpb.Message and not real
	//  go type of generated models. We could use some other MessageType implementation, but they always need
	//  the go type informations(reflect.Type) so without it the MessageType is useless for solving this)
	for i := 0; i < messageDesc.Fields().Len(); i++ {
		fieldDesc := messageDesc.Fields().Get(i)
		pbJSONName := fieldDesc.JSONName()
		nameTemplate = strings.ReplaceAll(nameTemplate, "."+m.upperFirst(pbJSONName), "."+pbJSONName)
		if fieldDesc.Message() != nil {
			nameTemplate = m.replaceFieldNamesInNameTemplate(fieldDesc.Message(), nameTemplate)
		}
	}
	return nameTemplate
}

// upperFirst converts the first letter of string to upper case
func (m *RemotelyKnownModel) upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// modelOptionFor retrieves first value for given key in model detail options
func (m *RemotelyKnownModel) modelOptionFor(key string, options []*api.ModelDetail_Option) (string, error) {
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
