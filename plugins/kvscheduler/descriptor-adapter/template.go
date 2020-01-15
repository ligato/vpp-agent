// Copyright (c) 2018 Cisco and/or its affiliates.
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

package main

const adapterTemplate = `// Code generated by adapter-generator. DO NOT EDIT.

package adapter

import (
	"github.com/golang/protobuf/proto"
	. "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"

    {{- range $i, $path := .Imports }}
	"{{ $path }}"
	{{- end }}
)

////////// type-safe key-value pair with metadata //////////

type {{ .DescriptorName }}KVWithMetadata struct {
	Key      string
	Value    {{ .ValueT }}
	Metadata {{ .MetadataT }}
	Origin   ValueOrigin
}

////////// type-safe Descriptor structure //////////

type {{ .DescriptorName }}Descriptor struct {
	Name                 string
	KeySelector          KeySelector
	ValueTypeName        string
	KeyLabel             func(key string) string
	ValueComparator      func(key string, oldValue, newValue {{ .ValueT }}) bool
	NBKeyPrefix          string
	WithMetadata         bool
	MetadataMapFactory   MetadataMapFactory
	Validate             func(key string, value {{ .ValueT }}) error
	Create               func(key string, value {{ .ValueT }}) (metadata {{ .MetadataT }}, err error)
	Delete               func(key string, value {{ .ValueT }}, metadata {{ .MetadataT }}) error
	Update               func(key string, oldValue, newValue {{ .ValueT }}, oldMetadata {{ .MetadataT }}) (newMetadata {{ .MetadataT }}, err error)
	UpdateWithRecreate   func(key string, oldValue, newValue {{ .ValueT }}, metadata {{ .MetadataT }}) bool
	Retrieve             func(correlate []{{ .DescriptorName }}KVWithMetadata) ([]{{ .DescriptorName }}KVWithMetadata, error)
	IsRetriableFailure   func(err error) bool
	DerivedValues        func(key string, value {{ .ValueT }}) []KeyValuePair
	Dependencies         func(key string, value {{ .ValueT }}) []Dependency
	RetrieveDependencies []string /* descriptor name */
}

////////// Descriptor adapter //////////

type {{ .DescriptorName }}DescriptorAdapter struct {
	descriptor *{{ .DescriptorName }}Descriptor
}

func New{{ .DescriptorName }}Descriptor(typedDescriptor *{{ .DescriptorName }}Descriptor) *KVDescriptor {
	adapter := &{{ .DescriptorName }}DescriptorAdapter{descriptor: typedDescriptor}
	descriptor := &KVDescriptor{
		Name:                 typedDescriptor.Name,
		KeySelector:          typedDescriptor.KeySelector,
		ValueTypeName:        typedDescriptor.ValueTypeName,
		KeyLabel:             typedDescriptor.KeyLabel,
		NBKeyPrefix:          typedDescriptor.NBKeyPrefix,
		WithMetadata:         typedDescriptor.WithMetadata,
		MetadataMapFactory:   typedDescriptor.MetadataMapFactory,
		IsRetriableFailure:   typedDescriptor.IsRetriableFailure,
		RetrieveDependencies: typedDescriptor.RetrieveDependencies,
	}
	if typedDescriptor.ValueComparator != nil {
		descriptor.ValueComparator = adapter.ValueComparator
	}
	if typedDescriptor.Validate != nil {
		descriptor.Validate = adapter.Validate
	}
	if typedDescriptor.Create != nil {
		descriptor.Create = adapter.Create
	}
	if typedDescriptor.Delete != nil {
		descriptor.Delete = adapter.Delete
	}
	if typedDescriptor.Update != nil {
		descriptor.Update = adapter.Update
	}
	if typedDescriptor.UpdateWithRecreate != nil {
		descriptor.UpdateWithRecreate = adapter.UpdateWithRecreate
	}
	if typedDescriptor.Retrieve != nil {
		descriptor.Retrieve = adapter.Retrieve
	}
	if typedDescriptor.Dependencies != nil {
		descriptor.Dependencies = adapter.Dependencies
	}
	if typedDescriptor.DerivedValues != nil {
		descriptor.DerivedValues = adapter.DerivedValues
	}
	return descriptor
}

func (da *{{ .DescriptorName }}DescriptorAdapter) ValueComparator(key string, oldValue, newValue proto.Message) bool {
	typedOldValue, err1 := cast{{ .DescriptorName }}Value(key, oldValue)
	typedNewValue, err2 := cast{{ .DescriptorName }}Value(key, newValue)
	if err1 != nil || err2 != nil {
		return false
	}
	return da.descriptor.ValueComparator(key, typedOldValue, typedNewValue)
}

func (da *{{ .DescriptorName }}DescriptorAdapter) Validate(key string, value proto.Message) (err error) {
	typedValue, err := cast{{ .DescriptorName }}Value(key, value)
	if err != nil {
		return err
	}
	return da.descriptor.Validate(key, typedValue)
}

func (da *{{ .DescriptorName }}DescriptorAdapter) Create(key string, value proto.Message) (metadata Metadata, err error) {
	typedValue, err := cast{{ .DescriptorName }}Value(key, value)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Create(key, typedValue)
}

func (da *{{ .DescriptorName }}DescriptorAdapter) Update(key string, oldValue, newValue proto.Message, oldMetadata Metadata) (newMetadata Metadata, err error) {
	oldTypedValue, err := cast{{ .DescriptorName }}Value(key, oldValue)
	if err != nil {
		return nil, err
	}
	newTypedValue, err := cast{{ .DescriptorName }}Value(key, newValue)
	if err != nil {
		return nil, err
	}
	typedOldMetadata, err := cast{{ .DescriptorName }}Metadata(key, oldMetadata)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Update(key, oldTypedValue, newTypedValue, typedOldMetadata)
}

func (da *{{ .DescriptorName }}DescriptorAdapter) Delete(key string, value proto.Message, metadata Metadata) error {
	typedValue, err := cast{{ .DescriptorName }}Value(key, value)
	if err != nil {
		return err
	}
	typedMetadata, err := cast{{ .DescriptorName }}Metadata(key, metadata)
	if err != nil {
		return err
	}
	return da.descriptor.Delete(key, typedValue, typedMetadata)
}

func (da *{{ .DescriptorName }}DescriptorAdapter) UpdateWithRecreate(key string, oldValue, newValue proto.Message, metadata Metadata) bool {
	oldTypedValue, err := cast{{ .DescriptorName }}Value(key, oldValue)
	if err != nil {
		return true
	}
	newTypedValue, err := cast{{ .DescriptorName }}Value(key, newValue)
	if err != nil {
		return true
	}
	typedMetadata, err := cast{{ .DescriptorName }}Metadata(key, metadata)
	if err != nil {
		return true
	}
	return da.descriptor.UpdateWithRecreate(key, oldTypedValue, newTypedValue, typedMetadata)
}

func (da *{{ .DescriptorName }}DescriptorAdapter) Retrieve(correlate []KVWithMetadata) ([]KVWithMetadata, error) {
	var correlateWithType []{{ .DescriptorName }}KVWithMetadata
	for _, kvpair := range correlate {
		typedValue, err := cast{{ .DescriptorName }}Value(kvpair.Key, kvpair.Value)
		if err != nil {
			continue
		}
		typedMetadata, err := cast{{ .DescriptorName }}Metadata(kvpair.Key, kvpair.Metadata)
		if err != nil {
			continue
		}
		correlateWithType = append(correlateWithType,
			{{ .DescriptorName }}KVWithMetadata{
				Key:      kvpair.Key,
				Value:    typedValue,
				Metadata: typedMetadata,
				Origin:   kvpair.Origin,
			})
	}

	typedValues, err := da.descriptor.Retrieve(correlateWithType)
	if err != nil {
		return nil, err
	}
	var values []KVWithMetadata
	for _, typedKVWithMetadata := range typedValues {
		kvWithMetadata := KVWithMetadata{
			Key:      typedKVWithMetadata.Key,
			Metadata: typedKVWithMetadata.Metadata,
			Origin:   typedKVWithMetadata.Origin,
		}
		kvWithMetadata.Value = typedKVWithMetadata.Value
		values = append(values, kvWithMetadata)
	}
	return values, err
}

func (da *{{ .DescriptorName }}DescriptorAdapter) DerivedValues(key string, value proto.Message) []KeyValuePair {
	typedValue, err := cast{{ .DescriptorName }}Value(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.DerivedValues(key, typedValue)
}

func (da *{{ .DescriptorName }}DescriptorAdapter) Dependencies(key string, value proto.Message) []Dependency {
	typedValue, err := cast{{ .DescriptorName }}Value(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.Dependencies(key, typedValue)
}

////////// Helper methods //////////

func cast{{ .DescriptorName }}Value(key string, value proto.Message) ({{ .ValueT }}, error) {
	typedValue, ok := value.({{ .ValueT }})
	if !ok {
		return nil, ErrInvalidValueType(key, value)
	}
	return typedValue, nil
}

func cast{{ .DescriptorName }}Metadata(key string, metadata Metadata) ({{ .MetadataT }}, error) {
	if metadata == nil {
		return nil, nil
	}
	typedMetadata, ok := metadata.({{ .MetadataT }})
	if !ok {
		return nil, ErrInvalidMetadataType(key)
	}
	return typedMetadata, nil
}
`
