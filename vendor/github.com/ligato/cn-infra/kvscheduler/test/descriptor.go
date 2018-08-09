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
	. "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/idxmap"
)

// ValueBuilder is type of the callback used to build values.
type ValueBuilder      func(key string, valueData interface{}) (value Value, err error)

// DependencyBuilder is type of the callback used to build dependencies.
type DependencyBuilder func(key string, value Value) []Dependency

// DerValuesBuilder is type of the callback used to build derived values.
type DerValuesBuilder  func(key string, value Value) []KeyValuePair

// RecreateChecker is type of the callback used to tell if a value needs to be re-created.
type RecreateChecker   func(key string, oldValue, newValue Value, metadata Metadata) bool

// MockDescriptorArgs encapsulates arguments for the descriptor.
type MockDescriptorArgs struct {
	Name              string
	KeySelector       KeySelector
	NBKeyPrefixes     []string
	WithMetadata      bool
	ValueBuilder      ValueBuilder
	DependencyBuilder DependencyBuilder
	DerValuesBuilder  DerValuesBuilder
	RecreateChecker   RecreateChecker
	DumpIsSupported   bool
	DumpDependencies  []string
}

// mockDescriptor implements KVDescriptor for UTs.
type mockDescriptor struct {
	nextIndex        int
	args             *MockDescriptorArgs
	sb               *MockSouthbound
}

// NewMockDescriptor creates a new instance of Mock Descriptor.
func NewMockDescriptor(args *MockDescriptorArgs, sb *MockSouthbound, firstFreeIndex int) KVDescriptor {
	return &mockDescriptor{
		nextIndex: firstFreeIndex,
		args:      args,
		sb:        sb,
		}
}

// validateKey tests predicate for a key that should hold.
func (md *mockDescriptor) validateKey(key string, predicate bool) {
	if !predicate && md.sb != nil {
		md.sb.registerKeyWithInvalidData(key)
	}
}

// GetName return name from the input arguments.
func (md *mockDescriptor) GetName() string {
	return md.args.Name
}

// KeySelector uses selector from the input arguments.
func (md *mockDescriptor) KeySelector(key string) bool {
	return md.args.KeySelector(key)
}

// NBKeyPrefixes returns NB key prefixes from the input arguments.
func (md *mockDescriptor) NBKeyPrefixes() []string {
	return md.args.NBKeyPrefixes
}

// WithMetadata returns factory for NameToInteger map if metadata are enabled
// by the input arguments.
func (md *mockDescriptor) WithMetadata() (withMeta bool, customMapFactory MetadataMapFactory) {
	if md.sb != nil && md.args.WithMetadata {
		return true, func() idxmap.NamedMappingRW {return NewNameToInteger(md.args.Name)}
	}
	return false, nil
}

// Build uses provided value builder or simply tries to cast valueData to Value.
func (md *mockDescriptor) Build(key string, valueData interface{}) (value Value, err error) {
	md.validateKey(key, md.args.KeySelector(key))
	if md.args.ValueBuilder != nil {
		return md.args.ValueBuilder(key, valueData)
	}
	// if ValueBuilder is not defined, try to cast the data directly to value
	var ok bool
	value, ok = valueData.(Value)
	if !ok {
		return nil, ErrInvalidValueDataType(key)
	}
	return
}

// Add executes add operation in the mock SB.
func (md *mockDescriptor) Add(key string, value Value) (metadata Metadata, err error) {
	md.validateKey(key, md.args.KeySelector(key))
	if md.sb != nil && md.args.WithMetadata && !md.sb.isKeyDerived(key) && value.Type() == Object {
		metadata = &OnlyInteger{md.nextIndex}
		md.nextIndex++
	}
	if md.sb != nil {
		md.validateKey(key, md.sb.GetValue(key) == nil)
		err = md.sb.executeChange(md.GetName(), Add, key, value, metadata)
	}
	return metadata, err
}

// Delete executes del operation in the mock SB.
func (md *mockDescriptor) Delete(key string, value Value, metadata Metadata) (err error) {
	md.validateKey(key, md.args.KeySelector(key))
	if md.sb != nil {
		kv := md.sb.GetValue(key)
		md.validateKey(key, kv != nil)
		if md.sb.isKeyDerived(key) {
			// re-generated on refresh
			md.validateKey(key, kv.Value.Equivalent(value))
		} else {
			md.validateKey(key, kv.Value == value)
		}
		md.validateKey(key, kv.Metadata == metadata)
		err = md.sb.executeChange(md.GetName(), Delete, key, nil, metadata)
	}
	return nil
}

// Modify executes modify operation in the mock SB.
func (md *mockDescriptor) Modify(key string, oldValue, newValue Value, oldMetadata Metadata) (newMetadata Metadata, err error) {
	md.validateKey(key, md.args.KeySelector(key))
	newMetadata = oldMetadata
	if md.sb != nil {
		kv := md.sb.GetValue(key)
		md.validateKey(key, kv != nil)
		if md.sb.isKeyDerived(key) {
			// re-generated on refresh
			md.validateKey(key, kv.Value.Equivalent(oldValue))
		} else {
			md.validateKey(key, kv.Value == oldValue)
		}
		md.validateKey(key, kv.Metadata == oldMetadata)
		err = md.sb.executeChange(md.GetName(), Modify, key, newValue, newMetadata)
	}
	return newMetadata, err
}

// ModifyHasToRecreate uses provided RecreateChecker.
func (md *mockDescriptor) ModifyHasToRecreate(key string, oldValue, newValue Value, metadata Metadata) bool {
	md.validateKey(key, md.args.KeySelector(key))
	if md.args.RecreateChecker != nil {
		return md.args.RecreateChecker(key, oldValue, newValue, metadata)
	}
	return false
}

// Update executes update operation in the mock SB.
func (md *mockDescriptor) Update(key string, value Value, metadata Metadata) (err error) {
	md.validateKey(key, md.args.KeySelector(key))
	if md.sb != nil {
		kv := md.sb.GetValue(key)
		md.validateKey(key, kv != nil)
		md.validateKey(key, kv.Value.Equivalent(value))
		md.validateKey(key, kv.Metadata == metadata)
		err = md.sb.executeChange(md.GetName(), Update, key, value, metadata)
	}
	return nil
}

// Dependencies uses provided DependencyBuilder.
func (md *mockDescriptor) Dependencies(key string, value Value) []Dependency {
	md.validateKey(key, md.args.KeySelector(key))
	if md.args.DependencyBuilder != nil {
		return md.args.DependencyBuilder(key, value)
	}
	return nil
}

// Dependencies uses provided DerValuesBuilder.
func (md *mockDescriptor) DerivedValues(key string, value Value) []KeyValuePair {
	md.validateKey(key, md.args.KeySelector(key))
	if md.args.DerValuesBuilder != nil {
		derivedKVs := md.args.DerValuesBuilder(key, value)
		if md.sb != nil {
			for _, kv := range derivedKVs {
				md.sb.registerDerivedKey(kv.Key)
			}
		}
		return derivedKVs
	}
	return nil
}

// Dump returns non-derived values currently set in the mock SB.
func (md *mockDescriptor) Dump(correlate []KVWithMetadata) ([]KVWithMetadata, error) {
	if !md.args.DumpIsSupported || md.sb == nil {
		return nil, ErrDumpNotSupported
	}
	return md.sb.dump(correlate, md.args.KeySelector)
}

// DumpDependencies returns dump dependencies from the input arguments.
func (md *mockDescriptor) DumpDependencies() []string {
	return md.args.DumpDependencies
}
