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
	"context"
	"github.com/ligato/cn-infra/idxmap"
)

// KeySelector is used to filter keys.
type KeySelector func(key string) bool

// ValueType is one of: Object, Action or Property.
type ValueType int

const (
	// Object is something that exists.
	//
	// Objects are typically base (non-derived) values, but it is allowed
	// to derive (sub)object from another object. It is actually allowed
	// to derive/ any value type from an object.
	// Without associated descriptor, object cannot be created (build + add)
	// via NB transaction (considered unimplemented) - it can only be pushed
	// from SB.
	// Descriptor is allowed to associate metadata with created non-derived
	// object values and let the scheduler to maintain and expose the mapping
	// between value labels and metadata (possibly extended with secondary
	// indices).
	Object ValueType = iota

	// Action is something to be executed when dependencies are met and reverted
	// when it is no longer the case. Actions are often derived from objects.
	//
	// Like objects, actions can be base or derived values, but it is not allowed
	// to further derive more values from actions. Also, actions are not expected
	// to have metadata associated with them.
	// Without associated descriptor, there are obviously no actions to execute,
	// thus un-described action values are never create - they can only be pushed
	// (as already executed) from SB.
	Action

	// Property is something an object has and other kv pairs may depend on.
	//
	// Property is always derived from object - it cannot be created/pushed as
	// a base value. Property cannot have descriptor associated with it
	// (i.e. no methods to execute, no dependencies or further derived values,
	// no associated metadata).
	Property
)

// String converts ValueType to string.
func (vt ValueType) String() string {
	switch vt {
	case Object:
		return "object"
	case Action:
		return "action"
	default:
		return "property"
	}
}

// Value may represent some object, action or property.
//
// Value can be built+added either via northbound transaction (NB-value,
// ValueOrigin = FromNB) or pushed (as already created) through SB notification
// (SB-value, ValueOrigin = FromSB). Values from NB take priority as they
// overwrite existing SB values (via Modify operation), whereas notifications
// for existing NB values are ignored. For values returned by Dump with unknown
// origin the scheduler reviews the value's history to determine where it came
// from.
//
// For descriptors the values (once built) are mutable objects - Add, Modify,
// Delete and Update method should reflect the value content without changing it.
// To add and maintain extra (runtime) attributes alongside the value, descriptor
// can use the value metadata.
type Value interface {
	// Label should return a short string labeling the value.
	// In the scope of a single descriptor, every described value should have
	// a unique label.
	Label() string

	// String returns a human-readable string for logging & REST interface.
	String() string

	// Equivalent is used by the scheduler to determine if Modify operation is
	// needed, i.e. it tells if this value if effectively the same as <v2> from
	// the NB point of view.
	// The implementation can be omitted simply by always returning false,
	// which will result in scheduler calling Modify every time the value was
	// updated/re-synced.
	Equivalent(v2 Value) bool

	// Type should classify the value.
	Type() ValueType
}

// Metadata are extra information carried alongside non-derived Object value
// that descriptor may use for runtime attributes, secondary lookups, etc. This
// data are opaque for the scheduler and fully owned by the descriptor.
// Descriptor is supposed to create/edit (and use) metadata inside the Add,
// Modify, Update methods and return the latest state in the dump.
// Metadata, however, should not be used to determine the list of derived values
// and dependencies for a value - this needs to be fixed for a given value
// (Modify is effectively replace) and known even before the value is added.
//
// The only way how scheduler can learn anything from metadata, is if MetadataMap
// is enabled by descriptor (using WithMetadata method) and a custom NamedMapping
// implementation is provided that defines secondary indexes (over metadata).
// The scheduler exposes the current snapshot of secondary indexes, but otherwise
// is not familiar with their semantics.
type Metadata interface{}

// KeyValuePair groups key with value.
type KeyValuePair struct {
	Key   string
	Value Value
}

// KeyValueDataPair groups key with value data.
type KeyValueDataPair struct {
	Key       string
	ValueData interface{}
}

// KeyWithError stores error for a key whose value failed to get updated.
type KeyWithError struct {
	Key   string
	Error error
}

// KVWithMetadata encapsulates key-value pair with metadata and the origin mark.
type KVWithMetadata struct {
	Key      string
	Value    Value
	Metadata Metadata
	Origin   ValueOrigin
}

// KVScheduler synchronizes data-oriented requests flowing from the northbound
// (NB) to the southbound (SB), by representing objects, actions over objects
// and object properties from the SB plane as key-value pairs with dependencies
// defined between them.
// The values can be built + added/modified/deleted via transactions issued from
// the NB interface or spontaneously pushed through notifications from the SB plane.
// The scheduling is then defined as follows: on any change the scheduler
// attempts to update every value which has satisfied dependencies but is
// out-of-sync with NB.
//
// For the scheduler, the key-value pairs are just abstract items that need
// to be managed in a synchronized fashion according to the described relations.
// It is up to the SB components to assign actual meaning to the individual
// values (via methods Add, Delete, Modify & Update of the KVDescriptor).
//
// The idea behind scheduler is based on the Mediator pattern - SB components
// do not communicate directly, but instead interact through the mediator.
// This reduces the dependencies between communicating objects, thereby reducing
// coupling.
//
// The values are described for scheduler by registered KVDescriptor-s.
// The scheduler learns two kinds of relations between values that have to be
// respected by the scheduling algorithm:
//   -> A depends on B:
//          - A cannot exist without B
//          - request to add A without B existing must be postponed by storing
//            A into the cache (a.k.a. pending) of values with unmet dependencies
//          - if B is to be removed and A exists, A must be removed first
//            and cached in case B is restored in the future
//          - Note: values pushed from SB are not checked for dependencies
//   -> B is derived from A:
//          - value B is not added directly (by NB or SB) but gets derived
//            from base value A (using the DerivedValues() method of the base
//            value's descriptor)
//          - derived value exists only as long as its base does and gets removed
//            (without caching) once the base value goes away
//          - derived value may be described by a different descriptor than
//            the base and usually represents property of the base value (that
//            other values may depend on) or an extra action to be taken
//            when additional dependencies are met.
//
// Every key-value pair must have at most one descriptor associated with it.
// NB values of type Object or Action without descriptor are considered
// unimplemented and will never be added or even built (can only be pushed from
// SB as already created/executed). On the other hand, for values of type Property
// it is actually required to have no descriptor associated with them.
//
// For descriptors the values are mutable objects - Add, Modify, Delete and
// Update method should reflect the value content without changing it.
// To add and maintain extra (runtime) attributes alongside the value, scheduler
// allows descriptors to append metadata - of any type - to each created
// non-derived Object value. Descriptor can also use the metadata to define
// secondary lookups, exposed via MetadataMap.
//
// Advantages of the centralized scheduling are:
//   - easy to add new descriptors and dependencies
//   - decreases the likelihood of race conditions and deadlocks in systems with
//     complex dependencies
//   - allows to write loosely-coupled SB components (mediator pattern)
//   - descriptor interface will force new SB components to follow the same
//     code structure which will make them easier to familiarize with
//   - NB components should never worry about dependencies between requests -
//     it is taken care of by the scheduler
//   - single cache for all pending values (exposed via REST, easier to debug)
//
// Apart from scheduling and execution, KVScheduler also offers the following
// features:
//   - collecting and counting present and past errors individually for every
//     value
//   - retry for previously failed actions
//   - transaction revert
//   - exposing history of actions, errors and pending values over the REST
//     interface
//   - clearly describing the sequence of actions to be executed and postponed
//     in the log file
//   - deadlock (circular dependency) detection
//   - TBD: consider exposing the current config as a plotted graph (returned via
//          REST) with values as nodes (colored to distinguish cached from added
//          ones) and dependencies as edges (unsatisfied marked with red color).
type KVScheduler interface {
	// RegisterKVDescriptor registers descriptor for a set of selected
	// keys. It should be called in the Init phase of agent plugins.
	// Every key-value pair must have at most one descriptor associated with it
	// (none for values of type Property).
	RegisterKVDescriptor(descriptor KVDescriptor)

	// GetRegisteredNBKeyPrefixes returns a list of key prefixes from NB with values
	// described by registered descriptors and therefore managed by the scheduler.
	GetRegisteredNBKeyPrefixes() []string

	// StartNBTransaction starts a new transaction from NB to SB plane.
	// The enqueued actions are scheduled for execution by Txn.Commit().
	StartNBTransaction(opts ...TxnOption) Txn

	// PushSBNotification notifies about a spontaneous value change in the SB
	// plane (i.e. not triggered by NB transaction).
	//
	// Pass <value> as nil if the value was removed, non-nil otherwise.
	//
	// Values pushed from SB do not trigger Add/Modify/Delete operations
	// on the descriptors - the change has already happened in SB - only
	// dependencies and derived values are updated.
	//
	// Values pushed from SB are overwritten by those created via NB transactions,
	// however. For example, notifications for values already created by NB
	// are ignored. But otherwise, SB values (not managed by NB) are untouched
	// by resync or any other operation of the scheduler/descriptor.
	PushSBNotification(key string, value Value, metadata Metadata) error

	// GetValue currently set for the given key.
	// The function can be used from within a transaction. However, if update
	// of A uses the value of B, then A should be marked as dependent on B
	// so that the scheduler can ensure that B is updated before A is.
	GetValue(key string) Value

	// GetValues returns a set of values matched by the given selector.
	GetValues(selector KeySelector) []KeyValuePair

	// GetMetadataMap returns (read-only) map associating value label with value
	// metadata of a given descriptor.
	// Returns nil if the descriptor does not expose metadata.
	GetMetadataMap(descriptor string) idxmap.NamedMapping

	// GetPendingValues returns list of values (possibly filtered by selector)
	// waiting for their dependencies to be met.
	GetPendingValues(keySelector KeySelector) []KeyValuePair

	// GetFailedValues returns a list of keys (possibly filtered by selector)
	// whose values are in a failed state (i.e. possibly not in the state as set
	// by the last transaction).
	GetFailedValues(keySelector KeySelector) []KeyWithError

	// SubscribeForErrors allows to get notified about all failed (Error!=nil)
	// and restored (Error==nil) values (possibly filtered using the selector).
	SubscribeForErrors(channel chan<- KeyWithError, keySelector KeySelector)
}

// Txn represent a single transaction.
// Scheduler starts to plan and execute actions only after Commit is called.
type Txn interface {
	// SetValueData changes (non-derived) value data.
	// NB provides untyped data which are build into the new value for the given
	// key by descriptor (method BuildValue).
	// If <valueData> is nil, the value will get deleted.
	SetValueData(key string, valueData interface{}) Txn

	// Resync all NB-values to match with <values>.
	// The list should consist of non-derived values only - derived values will
	// get created automatically using descriptors.
	// Run in case the SB may be out-of-sync with NB or with the scheduler
	// itself.
	Resync(values []KeyValueDataPair) Txn

	// Commit orders scheduler to execute enqueued operations.
	// Operations with unmet dependencies will get postponed and possibly
	// executed later.
	// <ctx> allows to cancel waiting for the end of a blocking transaction.
	// <txnError> covers validity of the transaction and the preparedness
	// of the scheduler to execute it.
	// <kvErrors> are related to operations from this transaction that
	// could be immediately executed or from previous transactions that have
	// got their dependencies satisfied by this txn.
	// Non-blocking transactions return immediately and always without errors.
	// Subscribe with KVScheduler.SubscribeForErrors() to get notified about all
	// errors, including those returned by action triggered later or asynchronously
	// by a SB notification.
	Commit(ctx context.Context) (kvErrors []KeyWithError, txnError error)
}
