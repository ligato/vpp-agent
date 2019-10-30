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

	"github.com/golang/protobuf/proto"

	"go.ligato.io/vpp-agent/v2/proto/ligato/generic"
)

// RegisteredModel represents a registered model.
type RegisteredModel struct {
	spec Spec
	modelOptions

	goType    reflect.Type
	protoName string
}

// Spec returns model specification for the model.
func (m RegisteredModel) Spec() *Spec {
	spec := m.spec
	return &spec
}

// ModelDescriptor returns descriptor for the model.
func (m RegisteredModel) ModelDescriptor() *generic.ModelDescriptor {
	return &generic.ModelDescriptor{
		Spec:      (*generic.ModelSpec)(m.Spec()),
		ProtoName: m.ProtoName(),
		Options: []*generic.ModelDescriptor_Option{
			{Key: "nameTemplate", Values: []string{m.NameTemplate()}},
			{Key: "goType", Values: []string{m.GoType()}},
		},
	}
}

// NewInstance creates new instance value for model type.
func (m RegisteredModel) NewInstance() proto.Message {
	return reflect.New(m.goType.Elem()).Interface().(proto.Message)
}

// ProtoName returns proto message name registered with the model.
func (m RegisteredModel) ProtoName() string {
	if m.protoName == "" {
		m.protoName = proto.MessageName(m.NewInstance())
	}
	return m.protoName
}

// NameTemplate returns name template for the model.
func (m RegisteredModel) NameTemplate() string {
	return m.nameTemplate
}

// GoType returns go type for the model.
func (m RegisteredModel) GoType() string {
	return m.goType.String()
}

// Path returns path for the model.
func (m RegisteredModel) Name() string {
	return m.spec.ModelName()
}

// KeyPrefix returns key prefix for the model.
func (m RegisteredModel) KeyPrefix() string {
	keyPrefix := m.spec.KeyPrefix()
	if m.nameFunc == nil {
		keyPrefix = strings.TrimSuffix(keyPrefix, "/")
	}
	return keyPrefix
}

// ParseKey parses the given key and returns item name
// or returns empty name and valid as false if the key is not valid.
func (m RegisteredModel) ParseKey(key string) (name string, valid bool) {
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
func (m RegisteredModel) IsKeyValid(key string) bool {
	_, valid := m.ParseKey(key)
	return valid
}

// StripKeyPrefix returns key with prefix stripped.
func (m RegisteredModel) StripKeyPrefix(key string) string {
	if name, valid := m.ParseKey(key); valid {
		return name
	}
	return key
}

func (m RegisteredModel) instanceName(x proto.Message) (string, error) {
	if m.nameFunc == nil {
		return "", nil
	}
	return m.nameFunc(x)
}
