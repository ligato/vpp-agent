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
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"

	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

// knownModel represents a registered local model (local model has go types compiled into program binary)
type knownModel struct {
	modelOptions
	spec   Spec
	pb     proto.Message
	goType reflect.Type

	// cache
	keyPrefix *string
	modelName *string
}

// Spec returns model specification for the model.
func (m *knownModel) Spec() Spec {
	return m.spec
}

// ModelDetail returns descriptor for the model.
func (m *knownModel) ModelDetail() *generic.ModelDetail {
	return &generic.ModelDetail{
		Spec:      m.Spec().Proto(),
		ProtoName: m.ProtoName(),
		Options: []*generic.ModelDetail_Option{
			{Key: "nameTemplate", Values: []string{m.NameTemplate()}},
			{Key: "protoFile", Values: []string{m.ProtoFile()}},
		},
	}
}

// NewInstance creates new instance value for model type.
func (m *knownModel) NewInstance() proto.Message {
	if m.goType != nil {
		return reflect.New(m.goType.Elem()).Interface().(proto.Message)
	}
	return dynamicpb.NewMessageType(m.pb.ProtoReflect().Descriptor()).New().Interface()
}

// ProtoName returns proto message name registered with the model.
func (m *knownModel) ProtoName() string {
	if m.pb != nil {
		return string(m.pb.ProtoReflect().Descriptor().FullName())
	}
	return ""
}

// ProtoFile returns proto file name for the model.
func (m *knownModel) ProtoFile() string {
	if m.pb != nil {
		return m.pb.ProtoReflect().Descriptor().ParentFile().Path()
	}
	return ""
}

// NameTemplate returns name template for the model.
func (m *knownModel) NameTemplate() string {
	return m.nameTemplate
}

// LocalGoType returns reflect go type for the model.
func (m *knownModel) LocalGoType() reflect.Type {
	return m.goType
}

// Name returns name for the model.
func (m *knownModel) Name() string {
	if m.modelName == nil {
		modelName := m.spec.ModelName()
		m.modelName = &modelName
	}
	return *m.modelName
}

// KeyPrefix returns key prefix for the model.
func (m *knownModel) KeyPrefix() string {
	if m.keyPrefix == nil {
		keyPrefix := keyPrefix(m.spec, m.nameFunc != nil)
		m.keyPrefix = &keyPrefix
	}
	return *m.keyPrefix
}

// ParseKey parses the given key and returns item name
// or returns empty name and valid as false if the key is not valid.
func (m *knownModel) ParseKey(key string) (name string, valid bool) {
	name = strings.TrimPrefix(key, m.KeyPrefix())
	if name == key || (name == "" && m.nameFunc != nil) {
		name = strings.TrimPrefix(key, m.Name())
	}
	// key had the prefix and also either
	// non-empty name or no name template
	if name != key && (name != "" || m.nameFunc == nil) {
		// TODO: validate name?
		return name, true
	}
	return "", false
}

// IsKeyValid returns true if given key is valid for this model.
func (m *knownModel) IsKeyValid(key string) bool {
	_, valid := m.ParseKey(key)
	return valid
}

// StripKeyPrefix returns key with prefix stripped.
func (m *knownModel) StripKeyPrefix(key string) string {
	if name, valid := m.ParseKey(key); valid {
		return name
	}
	return key
}

// InstanceName computes message name for given proto message using name template (if present).
func (m *knownModel) InstanceName(x any) (string, error) {
	if m.nameFunc == nil {
		return "", nil
	}
	return m.nameFunc(x)
}
