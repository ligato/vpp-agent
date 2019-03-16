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
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/idxmap"
)

// Dependency references another kv pair that must exist before the associated
// value can be created.
type Dependency struct {
	// Label should be a short human-readable string labeling the dependency.
	// Must be unique in the list of dependencies for a value.
	Label string

	// Key of another kv pair that the associated value depends on.
	// If empty, AnyOf must be defined instead.
	Key string

	// AnyOf defines a set of keys from which at least one must reference
	// a created object for the dependency to be considered satisfied - i.e.
	// **any of** the matched keys is good enough to satisfy the dependency.
	// Either Key or AnyOf should be defined, but not both at the same time.
	// BEWARE: AnyOf comes with more overhead than a static key dependency
	// (especially when KeySelector is used without KeyPrefixes), so prefer to
	// use Key whenever possible.
	AnyOf AnyOfDependency
}

// AnyOfDependency defines a set of keys from which at least one must reference
// a created object for the dependency to be considered satisfied.
//
// KeyPrefixes jointly select a set of keys that begin with at least one of the
// defined key prefixes, potentially further filtered by KeySelector, which is
// typically parsing and evaluating key suffix to check if it satisfies the
// dependency (i.e. the selections of KeyPrefixes and KeySelector are
// **intersected**).
//
// KeyPrefixes and KeySelector can be combined, but also used as standalone.
// However, using only KeySelector without limiting the set of candidates using
// prefixes is very costly - for the scheduling algorithm the key selector is
// a black box that can potentially match any key and must be therefore checked
// with every key entering the key space (i.e. linear complexity as opposed to
// logarithmic complexity when the key space is suitably reduced with prefixes).
type AnyOfDependency struct {
	// KeyPrefixes is a list of all (longest common) key prefixes found in the
	// keys satisfying the dependency. If defined (not empty), the scheduler will
	// know that for a given key to match the dependency, it must begin with at
	// least one of the specified prefixes.
	// Keys matched by these prefixes can be further filtered with the KeySelector
	// below.
	KeyPrefixes []string

	// KeySelector, if defined (non-nil), must return true for at least one of
	// the already created keys for the dependency to be considered satisfied.
	// It is recommended to narrow down the set of candidates as much as possible
	// using KeyPrefixes, so that the KeySelector will not have to be called too
	// many times, limiting its impact on the performance.
	KeySelector KeySelector
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
	// UnknownOrigin is given to a retrieved value when it cannot be determined
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

// KVDescriptor teaches KVScheduler how to CRUD values under keys matched
// by KeySelector().
//
// Every SB component should define one or more descriptors to cover all
// (non-property) keys under its management. The descriptor is what in essence
// gives meaning to individual key-value pairs. The list of available keys and
// their purpose should be properly documented so that clients from the NB plane
// can use them correctly. The scheduler does not care what CRUD methods do,
// it only needs to call the right callbacks at the right time.
//
// Every key-value pair must have at most one descriptor associated with it.
// NB base value without descriptor is considered unimplemented and will never
// be created.
// On the other hand, derived value is allowed to have no descriptor associated
// with it. Typically, properties of base values are implemented as derived
// (often empty) values without attached SB operations, used as targets for
// dependencies.
type KVDescriptor struct {
	// Name of the descriptor unique across all registered descriptors.
	Name string

	// TODO: replace KeySelector, KeyLabel & NBKeyPrefix with methods from
	// models.Spec.

	// KeySelector selects keys described by this descriptor.
	KeySelector KeySelector

	// TODO: obsolete, remove once Orchestrator is completed
	// ValueTypeName defines name of the proto.Message type used to represent
	// described values. This attribute is mandatory, otherwise LazyValue-s
	// received from NB (e.g. datasync package) cannot be un-marshalled.
	// Note: proto Messages are registered against this type name in the generated
	// code using proto.RegisterType().
	ValueTypeName string

	// KeyLabel can be *optionally* defined to provide a *shorter* value
	// identifier, that, unlike the original key, only needs to be unique in the
	// key scope of the descriptor and not necessarily in the entire key space.
	// If defined, key label will be used as value identifier in the metadata map
	// and in the non-verbose logs.
	KeyLabel func(key string) string

	// NBKeyPrefix is a key prefix that the scheduler should watch
	// in NB to receive all NB-values described by this descriptor.
	// The key space defined by NBKeyPrefix may cover more than KeySelector
	// selects - the scheduler will filter the received values and pass
	// to the descriptor only those that are really chosen by KeySelector.
	// The opposite may be true as well - KeySelector may select some extra
	// SB-only values, which the scheduler will not watch for in NB. Furthermore,
	// the keys may already be requested for watching by another descriptor
	// within the same plugin and in such case it is not needed to mention the
	// same prefix again.
	NBKeyPrefix string

	// ValueComparator can be *optionally* provided to customize comparision
	// of values for equality.
	// Scheduler compares values to determine if Update operation is really
	// needed.
	// For NB values, <oldValue> was either previously set by NB or refreshed
	// from SB, whereas <newValue> is a new value to be applied by NB.
	ValueComparator func(key string, oldValue, newValue proto.Message) bool

	// WithMetadata tells scheduler whether to enable metadata - run-time,
	// descriptor-owned, scheduler-opaque, data carried alongside a created
	// (non-derived) value.
	// If enabled, the scheduler will maintain a map between key (-label, if
	// KeyLabel is defined) and the associated metadata.
	// If <WithMetadata> is false, metadata returned by Create will be ignored
	// and other methods will receive nil metadata.
	WithMetadata bool

	// MetadataMapFactory can be used to provide a customized map implementation
	// for value metadata, possibly extended with secondary lookups.
	// If not defined, the scheduler will use the bare NamedMapping from
	// the idxmap package.
	MetadataMapFactory MetadataMapFactory

	// Validate value handler (optional).
	// Validate is called for every new value before it is Created or Updated.
	// If the validations fails (returned <err> is non-nil), the scheduler will
	// mark the value as invalid and will not attempt to apply it.
	// The descriptor can further specify which field(s) are not valid
	// by wrapping the validation error together with a slice of invalid fields
	// using the error InvalidValueError (see errors.go).
	Validate func(key string, value proto.Message) error

	// Create new value handler.
	// For non-derived values, descriptor may return metadata to associate with
	// the value.
	// For derived values, Create+Delete+Update are optional. Typically, properties
	// of base values are implemented as derived (often empty) values without
	// attached SB operations, used as targets for dependencies.
	Create func(key string, value proto.Message) (metadata Metadata, err error)

	// Delete value handler.
	// If Create is defined, Delete handler must be provided as well.
	Delete func(key string, value proto.Message, metadata Metadata) error

	// Update value handler.
	// The handler is optional - if not defined, value change will be carried out
	// via full re-creation (Delete followed by Create with the new value).
	// <newMetadata> can re-use the <oldMetadata>.
	Update func(key string, oldValue, newValue proto.Message, oldMetadata Metadata) (newMetadata Metadata, err error)

	// UpdateWithRecreate can be defined to tell the scheduler if going from
	// <oldValue> to <newValue> requires the value to be completely re-created
	// with Delete+Create handlers.
	// If not defined, KVScheduler will decide based on the (un)availability
	// of the Update operation - if provided, it is assumed that any change
	// can be applied incrementally, otherwise a full re-creation is the only way
	// to go.
	UpdateWithRecreate func(key string, oldValue, newValue proto.Message, metadata Metadata) bool

	// Retrieve should return all non-derived values described by this descriptor
	// that *really* exist in the southbound plane (and not what the current
	// scheduler's view of SB is). Derived value will get automatically created
	// using the method DerivedValues(). If some non-derived value doesn't
	// actually exist, it shouldn't be returned by DerivedValues() for the
	// retrieved base value!
	// <correlate> represents the non-derived values currently created
	// as viewed from the northbound/scheduler point of view:
	//   -> startup resync: <correlate> = values received from NB to be applied
	//   -> run-time/downstream resync: <correlate> = values applied according
	//      to the in-memory kv-store (scheduler's view of SB)
	//
	// The callback is optional - if not defined, it is assumed that descriptor
	// is not able to read the current SB state and thus refresh cannot be
	// performed for its kv-pairs.
	Retrieve func(correlate []KVWithMetadata) ([]KVWithMetadata, error)

	// IsRetriableFailure tells scheduler if the given error, returned by one
	// of Create/Delete/Update handlers, will always be returned for the
	// the same value (non-retriable) or if the value can be theoretically
	// fixed merely by repeating the operation.
	// If the callback is not defined, every error will be considered retriable.
	IsRetriableFailure func(err error) bool

	// DerivedValues returns ("derived") values solely inferred from the current
	// state of this ("base") value. Derived values cannot be changed by NB
	// transaction.
	// While their state and existence is bound to the state of their base value,
	// they are allowed to have their own descriptors.
	//
	// Typically, derived value represents the base value's properties (that
	// other kv pairs may depend on), or extra actions taken when additional
	// dependencies are met, but otherwise not blocking the base
	// value from being created.
	//
	// The callback is optional - if not defined, there will be no values derived
	// from kv-pairs of the descriptor.
	DerivedValues func(key string, value proto.Message) []KeyValuePair

	// Dependencies are keys that must already exist for the value to be created.
	// Conversely, if a dependency is to be removed, all values that depend on it
	// are deleted first and cached for a potential future re-creation.
	// Dependencies returned in the list are AND-ed.
	// The callback is optional - if not defined, the kv-pairs of the descriptor
	// are assumed to have no dependencies.
	Dependencies func(key string, value proto.Message) []Dependency

	// RetrieveDependencies is a list of descriptors whose values are needed
	// and should be already retrieved prior to calling Retrieve for this
	// descriptor.
	// Metadata for values already retrieved are available via GetMetadataMap().
	// TODO: define dependencies as a slice of models, not descriptors.
	RetrieveDependencies []string /* descriptor name */
}
