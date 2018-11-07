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

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/idxmap"
)

// KeySelector is used to filter keys.
type KeySelector func(key string) bool

// KeyValuePair groups key with value.
type KeyValuePair struct {
	// Key identifies value.
	Key string

	// Value may represent some object, action or property.
	//
	// Value can be added either via northbound transaction (NB-value,
	// ValueOrigin = FromNB) or pushed (as already created) through SB notification
	// (SB-value, ValueOrigin = FromSB). Values from NB take priority as they
	// overwrite existing SB values (via Modify operation), whereas notifications
	// for existing NB values are ignored. For values returned by Dump with unknown
	// origin the scheduler reviews the value's history to determine where it came
	// from.
	//
	// For descriptors the values are mutable objects - Add, Modify, Delete and
	// Update methods should reflect the value content without changing it.
	// To add and maintain extra (runtime) attributes alongside the value, descriptor
	// can use the value metadata.
	Value proto.Message
}

// Metadata are extra information carried alongside non-derived (base) value
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

// TxnOperation is one of: Pre-process, Add, Modify, Delete and Update.
type TxnOperation int

const (
	// UndefinedTxnOp represents undefined transaction operation.
	UndefinedTxnOp TxnOperation = iota
	// PreProcess key-value pair.
	PreProcess
	// Add new value.
	Add
	// Modify existing value.
	Modify
	// Delete existing value.
	Delete
	// Update (reflect modified dependencies) existing value.
	Update
)

// String returns human-readable string representation of transaction operation.
func (txnOpType TxnOperation) String() string {
	switch txnOpType {
	case UndefinedTxnOp:
		return "UNDEFINED"
	case PreProcess:
		return "PRE-PROCESS"
	case Add:
		return "ADD"
	case Modify:
		return "MODIFY"
	case Delete:
		return "DELETE"
	case Update:
		return "UPDATE"
	}
	return "INVALID"
}

// KeyWithError stores error for a key whose value failed to get updated.
type KeyWithError struct {
	Key          string
	TxnOperation TxnOperation
	Error        error
}

// KVWithMetadata encapsulates key-value pair with metadata and the origin mark.
type KVWithMetadata struct {
	Key      string
	Value    proto.Message
	Metadata Metadata
	Origin   ValueOrigin
}

// KVScheduler synchronizes the *desired* system state described by northbound
// (NB) components via transactions with the *actual* state of the southbound (SB).
// The  system state is represented as a set of inter-dependent key-value pairs
// that can be added, modified, deleted from within NB transactions or be notified
// about via notifications from the SB plane.
// The scheduling basically implements "state reconciliation" - periodically and
// on any change the scheduler attempts to update every value which has satisfied
// dependencies but is out-of-sync with the desired state given by NB.
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
//            A into the cache of values with unmet dependencies (a.k.a. pending)
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
// Base NB value without descriptor is considered unimplemented and will never
// be added (can only be pushed from SB as already created/executed).
// On the other hand, derived value is allowed to have no descriptor associated
// with it. Typically, properties of base values are implemented as derived
// (often empty) values without attached SB operations, used as targets for
// dependencies.
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
//   - descriptor API will force new SB components to follow the same
//     code structure which will make them easier to familiarize with
//   - NB components should never worry about dependencies between requests -
//     it is taken care of by the scheduler
//   - single cache for all (not only pending) values (exposed via REST,
//     easier to debug)
//
// Apart from scheduling and execution, KVScheduler also offers the following
// features:
//   - collecting and counting present and past errors individually for every
//     key
//   - retry for previously failed actions
//   - transaction reverting
//   - exposing history of actions, errors and pending values over the REST
//     interface
//   - clearly describing the sequence of actions to be executed and postponed
//     in the log file
//   - transaction execution tracing (using "runtime/trace" package)
//   - TBD: consider exposing the current config as a plotted graph (returned via
//          REST) with values as nodes (colored to distinguish cached from added
//          ones, derived from base, etc.) and dependencies as edges (unsatisfied
//          marked with red color).
type KVScheduler interface {
	// RegisterKVDescriptor registers descriptor for a set of selected
	// keys. It should be called in the Init phase of agent plugins.
	// Every key-value pair must have at most one descriptor associated with it
	// (none for derived values expressing properties).
	RegisterKVDescriptor(descriptor *KVDescriptor)

	// GetRegisteredNBKeyPrefixes returns a list of key prefixes from NB with values
	// described by registered descriptors and therefore managed by the scheduler.
	GetRegisteredNBKeyPrefixes() []string

	// StartNBTransaction starts a new transaction from NB to SB plane.
	// The enqueued actions are scheduled for execution by Txn.Commit().
	StartNBTransaction() Txn

	// TransactionBarrier ensures that all notifications received prior to the call
	// are associated with transactions that have already finalized.
	TransactionBarrier()

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
	// by reconciliation or any other operation of the scheduler/descriptor.
	PushSBNotification(key string, value proto.Message, metadata Metadata) error

	// GetValue currently set for the given key.
	// The function can be used from within a transaction. However, if update
	// of A uses the value of B, then A should be marked as dependent on B
	// so that the scheduler can ensure that B is updated before A is.
	GetValue(key string) proto.Message

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
	// whose (base) values are in a failed state (i.e. possibly not in the state as set
	// by the last transaction).
	GetFailedValues(keySelector KeySelector) []KeyWithError

	// SubscribeForErrors allows to get notified about all failed (Error!=nil)
	// and restored (Error==nil) values (possibly filtered using the selector).
	SubscribeForErrors(channel chan<- KeyWithError, keySelector KeySelector)
}

// Txn represent a single transaction.
// Scheduler starts to plan and execute actions only after Commit is called.
type Txn interface {
	// SetValue changes (non-derived) lazy value - un-marshalled during
	// transaction pre-processing using ValueTypeName given by descriptor.
	// If <value> is nil, the value will get deleted.
	SetValue(key string, value datasync.LazyValue) Txn

	// Commit orders scheduler to execute enqueued operations.
	// Operations with unmet dependencies will get postponed and possibly
	// executed later.
	// <ctx> allows to pass transaction options (see With* functions from
	// txn_options.go) or to cancel waiting for the end of a blocking transaction.
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
