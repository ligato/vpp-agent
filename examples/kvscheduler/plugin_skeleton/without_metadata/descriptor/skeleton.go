// Copyright (c) 2019 Cisco and/or its affiliates.
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

package descriptor

import (
	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/plugin_skeleton/without_metadata/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/plugin_skeleton/without_metadata/model"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

const (
	// SkeletonDescriptorName is the name of the descriptor skeleton.
	SKeletonDescriptorName = "skeleton"
)

// SkeletonDescriptor is only a skeleton of a descriptor, which can be used
// as a starting point to build a new descriptor from.
type SkeletonDescriptor struct {
	log logging.Logger
}

// NewSkeletonDescriptor creates a new instance of the descriptor.
func NewSkeletonDescriptor(log logging.PluginLogger) *kvs.KVDescriptor {
	// descriptors are supposed to be stateless, so use the structure only
	// as a context for things that do not change once the descriptor is
	// constructed - e.g. a reference to the logger to use within the descriptor
	descrCtx := &SkeletonDescriptor{
		log: log.NewLogger("skeleton-descriptor"),
	}

	// use adapter to convert typed descriptor into generic descriptor API
	typedDescr := &adapter.SkeletonDescriptor{
		Name:                 SKeletonDescriptorName,
		NBKeyPrefix:          model.ValueModel.KeyPrefix(),
		ValueTypeName:        model.ValueModel.ProtoName(),
		KeySelector:          model.ValueModel.IsKeyValid,
		KeyLabel:             model.ValueModel.StripKeyPrefix,
		ValueComparator:      descrCtx.EquivalentValues,
		Validate:             descrCtx.Validate,
		Create:               descrCtx.Create,
		Delete:               descrCtx.Delete,
		Update:               descrCtx.Update,
		UpdateWithRecreate:   descrCtx.UpdateWithRecreate,
		Retrieve:             descrCtx.Retrieve,
		IsRetriableFailure:   descrCtx.IsRetriableFailure,
		DerivedValues:        descrCtx.DerivedValues,
		Dependencies:         descrCtx.Dependencies,
		RetrieveDependencies: []string{}, // list the names of the descriptors to Retrieve first
	}
	return adapter.NewSkeletonDescriptor(typedDescr)
}

// EquivalentInterfaces compares two revisions of the same value for equality.
func (d *SkeletonDescriptor) EquivalentValues(key string, old, new *model.ValueSkeleton) bool {
	// compare **non-primary** attributes here (none in the ValueSkeleton)
	return true
}

// Validate validates value before it is applied.
func (d *SkeletonDescriptor) Validate(key string, value *model.ValueSkeleton) error {
	return nil
}

// Create creates new value.
func (d *SkeletonDescriptor) Create(key string, value *model.ValueSkeleton) (metadata interface{}, err error) {
	return nil, nil
}

// Delete removes an existing value.
func (d *SkeletonDescriptor) Delete(key string, value *model.ValueSkeleton, metadata interface{}) error {
	return nil
}

// Update updates existing value.
func (d *SkeletonDescriptor) Update(key string, old, new *model.ValueSkeleton, oldMetadata interface{}) (newMetadata interface{}, err error) {
	return nil, nil
}

// UpdateWithRecreate returns true if value update requires full re-creation.
func (d *SkeletonDescriptor) UpdateWithRecreate(key string, old, new *model.ValueSkeleton, metadata interface{}) bool {
	return false
}

// Retrieve retrieves values from SB.
func (d *SkeletonDescriptor) Retrieve(correlate []adapter.SkeletonKVWithMetadata) (retrieved []adapter.SkeletonKVWithMetadata, err error) {
	return retrieved, nil
}

// IsRetriableFailure returns true if the given error, returned by one of the CRUD
// operations, can be theoretically fixed by merely repeating the operation.
func (d *SkeletonDescriptor) IsRetriableFailure(err error) bool {
	return true
}

// DerivedValues breaks the value into multiple part handled/referenced
// separately.
func (d *SkeletonDescriptor) DerivedValues(key string, value *model.ValueSkeleton) (derived []kvs.KeyValuePair) {
	return derived
}

// Dependencies lists dependencies of the given value.
func (d *SkeletonDescriptor) Dependencies(key string, value *model.ValueSkeleton) (deps []kvs.Dependency) {
	return deps
}
