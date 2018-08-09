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

package api

import (
	"github.com/ligato/cn-infra/idxmap"
)

// Dependency references another kv pair that must exist before the associated
// value can be added.
type Dependency struct {
	// Label should be a short human-readable string labeling the dependency.
	// Must be unique in the list of dependencies for a value.
	Label string

	// Key of another kv pair that the associated value depends on.
	// If empty, AnyOf must be defined instead.
	Key string

	// AnyOf, if not nil, must return true for at least one of the already added
	// keys for the dependency to be considered satisfied.
	// Either Key or AnyOf should be defined, but not both at the same time.
	// Note: AnyOf comes with more overhead than a static key dependency,
	// so prefer to use the latter whenever possible.
	AnyOf KeySelector
}

// MetadataMapFactory can be used by descriptor to define a custom map associating
// value labels with value metadata, potentially extending the basic in-memory
// implementation (memNamedMapping) with secondary indexes, type-safe watch, etc.
// If metadata are enabled (by WithMetadata method), the scheduler will create
// an instance of the map using the provided factory during the descriptor
// registration (RegisterKVDescriptor). Immediately afterwards, the mapping
// is available read-only via scheduler's method GetMetadataMap. The returned
// map can be then casted to the customized implementation, but it should remain
// read-only (i.e. define read-only interface for the customized implementation).
type MetadataMapFactory func() idxmap.NamedMappingRW

// ValueOrigin is one of: FromNB, FromSB, UnknownOrigin.
type ValueOrigin int

const (
	// UnknownOrigin is returned by Dump for a value when it cannot be determine
	// if the value was previously created by NB or not.
	// Scheduler will then look into its history to find out if the value was
	// ever managed by NB to determine the origin heuristically.
	UnknownOrigin ValueOrigin = iota

	// FromNB marks value created via NB transaction.
	FromNB

	// FromSB marks value not managed by NB - i.e. created automatically or
	// externally in SB.
	FromSB
)

// String converts ValueOrigin to string.
func (vo ValueOrigin) String() string {
	switch vo {
	case FromNB:
		return "from-NB"
	case FromSB:
		return "from-SB"
	default:
		return "unknown"
	}
}

// KVDescriptor teaches KVScheduler how to build/add/delete/modify/update & dump
// values under keys matched by KeySelector().
//
// Every SB component should define one or more descriptors to cover all
// (non-property) keys under its management. The descriptor is what in essence
// gives meaning to individual key-value pairs. The list of available keys and
// their purpose should be described in the API of SB components so that NB plane
// can use them correctly. The scheduler does not care what Add/Delete/...
// methods do, it only needs to call the right callbacks at the right time.
//
// Every key-value pair must have at most one descriptor associated with it.
// NB values of type Object or Action without descriptor are considered
// unimplemented and will never be added or even built (can only be pushed from
// SB as already created/executed). On the other hand, for values of type Property
// it is actually required to have no descriptor associated with them.
type KVDescriptor interface {
	// GetName returns name of the descriptor unique across all registered
	// descriptors.
	GetName() string

	// KeySelector returns true for keys described by this descriptor.
	KeySelector(key string) bool

	// NBKeyPrefixes lists all key prefixes that the scheduler should watch
	// in NB to receive all NB-values described by this descriptor.
	// The key space defined by NBKeyPrefixes may cover more than KeySelector
	// selects - the scheduler will filter the received values and pass
	// to the descriptor only those that are really chosen by KeySelector.
	// The opposite may be true as well - KeySelector may select some extra
	// SB-only values, which the scheduler will not watch for in NB. Furthermore,
	// the keys may already be requested for watching by another descriptor
	// within the same plugin and in such case it is not needed to mention the
	// same prefixes again.
	NBKeyPrefixes() []string

	// WithMetadata tells scheduler whether to enable metadata - run-time,
	// descriptor-owned, scheduler-opaque, data carried alongside a created
	// (non-derived Object) value.
	// The scheduler maintains the association between the value (label) and
	// metadata using NamedMapping from the idxmap package. To define a customized
	// map implementation, possibly extended with secondary indexes, provide
	// map factory as the second return value.
	// If <withMeta> is false, metadata returned by Add will be ignored and
	// other methods will receive nil metadata.
	WithMetadata() (withMeta bool, customMapFactory MetadataMapFactory)

	// Build a fresh new value from data received from NB.
	// For every added/modified (untyped) value received in NB transaction,
	// BuildValue must be called first to get an implementation of Value which
	// can be then used with other operations of the descriptor.
	Build(key string, valueData interface{}) (value Value, err error)

	// Add new value.
	// For non-derived Object values, descriptor may return metadata to associate
	// with the value.
	Add(key string, value Value) (metadata Metadata, err error)

	// Delete existing value.
	Delete(key string, value Value, metadata Metadata) error

	// Modify existing value.
	// <newMetadata> can re-use the <oldMetadata>.
	Modify(key string, oldValue, newValue Value, oldMetadata Metadata) (newMetadata Metadata, err error)

	// ModifyHasToRecreate should return true if going from <oldValue> to
	// <newValue> requires the value to be completely re-created with
	// Delete+Add.
	ModifyHasToRecreate(key string, oldValue, newValue Value, metadata Metadata) bool

	// Update is called every time the "context" of the value changes - whenever
	// a dependency is modified or the set of dependencies changes without
	// preventing the existence of this value.
	Update(key string, value Value, metadata Metadata) error

	// List dependencies of the given value.
	// Dependencies are keys that must already exist for the value to be added.
	// Conversely, if a dependency is to be removed, all values that depend on it
	// are deleted first and cached for a potential future re-creation.
	// Dependencies returned in the list are AND-ed.
	Dependencies(key string, value Value) []Dependency

	// DerivedValues returns ("derived") values solely inferred from the current
	// state of this ("base") value. Derived values cannot be changed by NB
	// transaction.
	// While their state and existence is bound to the state of their base value,
	// they are allowed to have their own descriptors.
	//
	// Typically, derived value represents the base value's properties (that
	// other kv pairs may depend on), or extra actions taken when additional
	// dependencies are met, but otherwise not blocking the base
	// value from being added.
	DerivedValues(key string, value Value) []KeyValuePair

	// Dump should return all non-derived values described by this descriptor
	// that *really* exist in the southbound plane (and not what the current
	// scheduler's view of SB is). Derived value will get automatically created
	// using the method DerivedValues(). If some non-derived value doesn't
	// actually exist, it shouldn't be returned by DerivedValues() for the dumped
	// value!
	// <correlate> represents the non-derived values currently created
	// as viewed from the northbound/scheduler point of view:
	//   -> startup resync: <correlate> = values received from NB to be applied
	//   -> run-time resync: <correlate> = values applied according to the
	//      in-memory kv-store
	// Return ErrDumpNotSupported if dumping is not supported by this descriptor.
	Dump(correlate []KVWithMetadata) ([]KVWithMetadata, error)

	// DumpDependencies returns a list of descriptors that have to be dumped
	// before this descriptor. Values already dumped are available for reading
	// via scheduler methods GetValue(), GetValues() and runtime data using
	// GetMetadataMap().
	DumpDependencies() []string /* descriptor name */
}
