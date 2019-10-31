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

// KnownModel represents a registered model.
type KnownModel struct {
	spec Spec
	modelOptions

	goType    reflect.Type
	protoName string
}

// Spec returns model specification for the model.
func (m KnownModel) Spec() *Spec {
	spec := m.spec
	return &spec
}

// ModelDescriptor returns descriptor for the model.
func (m KnownModel) ModelDetail() *generic.ModelDetail {
	return &generic.ModelDetail{
		Spec:      m.Spec().Proto(),
		ProtoName: proto.String(m.ProtoName()),
		Options: []*generic.ModelDetail_Option{
			{Key: proto.String("nameTemplate"), Values: []string{m.NameTemplate()}},
			{Key: proto.String("goType"), Values: []string{m.GoType()}},
		},
	}
}

// NewInstance creates new instance value for model type.
func (m KnownModel) NewInstance() proto.Message {
	return reflect.New(m.goType.Elem()).Interface().(proto.Message)
}

// ProtoName returns proto message name registered with the model.
func (m KnownModel) ProtoName() string {
	if m.protoName == "" {
		m.protoName = proto.MessageName(m.NewInstance())
	}
	return m.protoName
}

// NameTemplate returns name template for the model.
func (m KnownModel) NameTemplate() string {
	return m.nameTemplate
}

// GoType returns go type for the model.
func (m KnownModel) GoType() string {
	return m.goType.String()
}

// Path returns path for the model.
func (m KnownModel) Name() string {
	return m.spec.ModelName()
}

// KeyPrefix returns key prefix for the model.
func (m KnownModel) KeyPrefix() string {
	keyPrefix := m.spec.KeyPrefix()
	if m.nameFunc == nil {
		keyPrefix = strings.TrimSuffix(keyPrefix, "/")
	}
	return keyPrefix
}

// ParseKey parses the given key and returns item name
// or returns empty name and valid as false if the key is not valid.
func (m KnownModel) ParseKey(key string) (name string, valid bool) {
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
func (m KnownModel) IsKeyValid(key string) bool {
	_, valid := m.ParseKey(key)
	return valid
}

// StripKeyPrefix returns key with prefix stripped.
func (m KnownModel) StripKeyPrefix(key string) string {
	if name, valid := m.ParseKey(key); valid {
		return name
	}
	return key
}

func (m KnownModel) instanceName(x proto.Message) (string, error) {
	if m.nameFunc == nil {
		return "", nil
	}
	return m.nameFunc(x)
}
