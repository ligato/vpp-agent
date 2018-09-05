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

package base

import (
	"fmt"
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

// TODO: remove printf statements

// DescriptorBase provides default(=empty) implementations for all methods
// of the KVDescriptor.
// To be used by small descriptors where auto-generation of the boiler-plate code
// produces more code than it saves.
//
// Inherit from like this:
//
// type MyDescriptor struct {
//     DescriptorBase
//     // ...
// }
//
type DescriptorBase struct {
}

// GetName returns "base".
func (bkvd *DescriptorBase) GetName() string {
	return "base"
}

// KeySelector matches no keys.
func (bkvd *DescriptorBase) KeySelector(key string) bool {
	return false
}

// NBKeyPrefixes return nil.
func (bkvd *DescriptorBase) NBKeyPrefixes() []string {
	return nil
}

// WithMetadata disables metadata.
func (bkvd *DescriptorBase) WithMetadata() (withMeta bool, customMapFactory MetadataMapFactory) {
	return false, nil
}

// Build tries to cast value data directly to Value.
func (bkvd *DescriptorBase) Build(key string, valueData interface{}) (value Value, err error) {
	var ok bool
	value, ok = valueData.(Value)
	if !ok {
		return nil, ErrInvalidValueDataType(key)
	}
	return
}

// Add does nothing.
func (bkvd *DescriptorBase) Add(key string, value Value) (metadata Metadata, err error) {
	fmt.Printf("Add for key=%s is not implemented\n", key)
	return nil, nil
}

// Delete does nothing.
func (bkvd *DescriptorBase) Delete(key string, value Value, metadata Metadata) error {
	fmt.Printf("Delete for key=%s is not implemented\n", key)
	return nil
}

// Modify does nothing.
func (bkvd *DescriptorBase) Modify(key string, oldValue, newValue Value, metadata Metadata) (Metadata, error) {
	fmt.Printf("Modify for key=%s is not implemented\n", key)
	return nil, nil
}

// ModifyHasToRecreate returns false.
func (bkvd *DescriptorBase) ModifyHasToRecreate(key string, oldValue, newValue Value, metadata Metadata) bool {
	return false
}

// Update does nothing.
func (bkvd *DescriptorBase) Update(key string, value Value, metadata Metadata) error {
	fmt.Printf("Update for key=%s is not implemented\n", key)
	return nil
}

// Dependencies returns empty list of dependencies.
func (bkvd *DescriptorBase) Dependencies(key string, value Value) []Dependency {
	return nil
}

// DerivedValues returns empty list of derived values.
func (bkvd *DescriptorBase) DerivedValues(key string, value Value) []KeyValuePair {
	return nil
}

// Dump is not supported.
func (bkvd *DescriptorBase) Dump(correlate []KVWithMetadata) ([]KVWithMetadata, error) {
	fmt.Println("Dump is not implemented")
	return nil, ErrDumpNotSupported
}

// DumpDependencies returns no dependencies.
func (bkvd *DescriptorBase) DumpDependencies() []string /* descriptor name */ {
	return nil
}
