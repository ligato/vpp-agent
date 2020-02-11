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

package test

import (
	"strings"

	"github.com/golang/protobuf/proto"
	"go.ligato.io/cn-infra/v2/idxmap"

	. "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

// WithoutOp references operation to leave undefined in the MockDescriptor.
type WithoutOp int

const (
	// WithoutCreate tells MockDescriptor to leave Create as nil.
	WithoutCreate WithoutOp = iota
	// WithoutUpdate tells MockDescriptor to leave Update as nil.
	WithoutUpdate
	// WithoutDelete tells MockDescriptor to leave Delete as nil.
	WithoutDelete
	// WithoutRetrieve tells MockDescriptor to leave Retrieve as nil.
	WithoutRetrieve
)

// mockDescriptor implements KVDescriptor for UTs.
type mockDescriptor struct {
	nextIndex int
	args      *KVDescriptor
	sb        *MockSouthbound
}

// NewMockDescriptor creates a new instance of Mock Descriptor.
func NewMockDescriptor(args *KVDescriptor, sb *MockSouthbound, firstFreeIndex int, withoutOps ...WithoutOp) *KVDescriptor {
	mock := &mockDescriptor{
		nextIndex: firstFreeIndex,
		args:      args,
		sb:        sb,
	}
	descriptor := &KVDescriptor{
		Name:                 args.Name,
		KeySelector:          args.KeySelector,
		ValueTypeName:        args.ValueTypeName,
		ValueComparator:      args.ValueComparator,
		KeyLabel:             args.KeyLabel,
		NBKeyPrefix:          args.NBKeyPrefix,
		WithMetadata:         args.WithMetadata,
		Validate:             args.Validate,
		IsRetriableFailure:   args.IsRetriableFailure,
		UpdateWithRecreate:   args.UpdateWithRecreate,
		Dependencies:         args.Dependencies,
		RetrieveDependencies: args.RetrieveDependencies,
	}
	if args.WithMetadata {
		descriptor.MetadataMapFactory = func() idxmap.NamedMappingRW {
			return NewNameToInteger(args.Name)
		}
		descriptor.KeyLabel = func(key string) string {
			return strings.TrimPrefix(key, args.NBKeyPrefix)
		}
	}
	if args.DerivedValues != nil {
		descriptor.DerivedValues = mock.DerivedValues
	}

	// operations that can be left undefined:
	withoutMap := make(map[WithoutOp]struct{})
	for _, withoutOp := range withoutOps {
		withoutMap[withoutOp] = struct{}{}
	}
	if _, withoutCreate := withoutMap[WithoutCreate]; !withoutCreate {
		descriptor.Create = mock.Create
	}
	if _, withoutDelete := withoutMap[WithoutDelete]; !withoutDelete {
		descriptor.Delete = mock.Delete
	}
	if _, withoutUpdate := withoutMap[WithoutUpdate]; !withoutUpdate {
		descriptor.Update = mock.Update
	}
	if _, withoutRetrieve := withoutMap[WithoutRetrieve]; !withoutRetrieve {
		descriptor.Retrieve = mock.Retrieve
	}
	return descriptor
}

// validateKey tests predicate for a key that should hold.
func (md *mockDescriptor) validateKey(key string, predicate bool) {
	if !predicate && md.sb != nil {
		md.sb.registerKeyWithInvalidData(key)
	}
}

// equalValues compares two values for equality
func (md *mockDescriptor) equalValues(key string, v1, v2 proto.Message) bool {
	if md.args.ValueComparator != nil {
		return md.args.ValueComparator(key, v1, v2)
	}
	return proto.Equal(v1, v2)
}

// Create executes create operation in the mock SB.
func (md *mockDescriptor) Create(key string, value proto.Message) (metadata Metadata, err error) {
	md.validateKey(key, md.args.KeySelector(key))
	withMeta := md.sb != nil && md.args.WithMetadata && !md.sb.isKeyDerived(key)
	if withMeta {
		metadata = &OnlyInteger{md.nextIndex}
	}
	if md.sb != nil {
		md.validateKey(key, md.sb.GetValue(key) == nil)
		err = md.sb.executeChange(md.args.Name, MockCreate, key, value, metadata)
	}
	if err == nil && withMeta {
		md.nextIndex++
	}
	return metadata, err
}

// Delete executes del operation in the mock SB.
func (md *mockDescriptor) Delete(key string, value proto.Message, metadata Metadata) (err error) {
	md.validateKey(key, md.args.KeySelector(key))
	if md.sb != nil {
		kv := md.sb.GetValue(key)
		md.validateKey(key, kv != nil)
		if md.sb.isKeyDerived(key) {
			// re-generated on refresh
			md.validateKey(key, md.equalValues(key, kv.Value, value))
		} else {
			md.validateKey(key, kv.Value == value)
		}
		md.validateKey(key, kv.Metadata == metadata)
		err = md.sb.executeChange(md.args.Name, MockDelete, key, nil, metadata)
	}
	return err
}

// Update executes update operation in the mock SB.
func (md *mockDescriptor) Update(key string, oldValue, newValue proto.Message, oldMetadata Metadata) (newMetadata Metadata, err error) {
	md.validateKey(key, md.args.KeySelector(key))
	newMetadata = oldMetadata
	if md.sb != nil {
		kv := md.sb.GetValue(key)
		md.validateKey(key, kv != nil)
		if md.sb.isKeyDerived(key) {
			// re-generated on refresh
			md.validateKey(key, md.equalValues(key, kv.Value, oldValue))
		} else {
			md.validateKey(key, kv.Value == oldValue)
		}
		md.validateKey(key, kv.Metadata == oldMetadata)
		err = md.sb.executeChange(md.args.Name, MockUpdate, key, newValue, newMetadata)
	}
	return newMetadata, err
}

// Dependencies uses provided DerValuesBuilder.
func (md *mockDescriptor) DerivedValues(key string, value proto.Message) []KeyValuePair {
	md.validateKey(key, md.args.KeySelector(key))
	if md.args.DerivedValues != nil {
		derivedKVs := md.args.DerivedValues(key, value)
		if md.sb != nil {
			for _, kv := range derivedKVs {
				md.sb.registerDerivedKey(kv.Key)
			}
		}
		return derivedKVs
	}
	return nil
}

// Retrieve returns non-derived values currently set in the mock SB.
func (md *mockDescriptor) Retrieve(correlate []KVWithMetadata) ([]KVWithMetadata, error) {
	if md.sb == nil {
		return nil, nil
	}
	return md.sb.retrieve(md.args.Name, correlate, md.args.KeySelector)
}
