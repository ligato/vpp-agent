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

package kvscheduler

import (
	"context"
	"sync"

	. "github.com/ligato/cn-infra/kvscheduler/api"

	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/idxmap/mem"
	"github.com/ligato/cn-infra/kvscheduler/graph"
	"github.com/ligato/cn-infra/kvscheduler/registry"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
)

const (
	// DependencyRelation identifies dependency relation for the graph.
	DependencyRelation = "depends-on"

	// DerivesRelation identifies relation of value derivation for the graph.
	DerivesRelation = "derives"
)

// Scheduler is a CN-infra plugin implementing KVScheduler.
// Detailed documentation can be found in the "api" and "docs" sub-folders.
type Scheduler struct {
	Deps

	// temporary until datasync and scheduler are properly integrated
	isInitialized bool

	// management of go routines
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// in-memory representation of all added+pending kv-pair and their dependencies
	graph graph.Graph

	// registry for descriptors
	registry registry.Registry

	// a list of key prefixed covered by registered descriptors
	keyPrefixes []string

	// TXN processing
	txnQueue     chan *queuedTxn
	errorSubs    []errorSubscription
	txnSeqNumber uint
	resyncCount  uint

	historyLock  sync.Mutex
	txnHistory   []*recordedTxn // ordered from the oldest to the latest
}

// Deps lists dependencies of the scheduler.
type Deps struct {
	infra.PluginName
	Log logging.PluginLogger
	// REST, etc.
}

// SchedulerTxn implements transaction for the KV scheduler.
type SchedulerTxn struct {
	scheduler *Scheduler
	data      *queuedTxn
	err       error
}

// errorSubscription represents one subscription for error updates.
type errorSubscription struct {
	channel  chan<- KeyWithError
	selector KeySelector
}

// Init initializes the scheduler. Single go routine is started that will process
// all the transactions synchronously.
func (scheduler *Scheduler) Init() error {
	// prepare context for all go routines
	scheduler.ctx, scheduler.cancel = context.WithCancel(context.Background())
	// initialize graph for in-memory storage of added+pending kv pairs
	scheduler.graph = graph.NewGraph()
	// initialize registry for key->descriptor lookups
	scheduler.registry = registry.NewRegistry()
	// prepare channel for serializing transactions
	scheduler.txnQueue = make(chan *queuedTxn, 100)
	// go routine processing serialized transactions
	go scheduler.consumeTransactions()
	// temporary until datasync and scheduler are properly integrated
	scheduler.isInitialized = true
	return nil
}

// IsInitialized is a method temporarily used by PropagateChanges until datasync
// and scheduler are properly integrated.
func (scheduler *Scheduler) IsInitialized() bool {
	return scheduler.isInitialized
}

// Close stops all the go routines.
func (scheduler *Scheduler) Close() error {
	scheduler.cancel()
	scheduler.wg.Wait()
	return nil
}

// RegisterKVDescriptor registers descriptor for a set of selected
// keys. It should be called in the Init phase of agent plugins.
// Every key-value pair must have at most one descriptor associated with it
// (none for values of type Property).
func (scheduler *Scheduler) RegisterKVDescriptor(descriptor KVDescriptor) {
	scheduler.registry.RegisterDescriptor(descriptor)
	scheduler.keyPrefixes = append(scheduler.keyPrefixes, descriptor.NBKeyPrefixes()...)

	withMeta, metadataMapFactory := descriptor.WithMetadata()
	if withMeta {
		var metadataMap idxmap.NamedMappingRW
		if metadataMapFactory != nil {
			metadataMap = metadataMapFactory()
		} else {
			metadataMap = mem.NewNamedMapping(scheduler.Log, descriptor.GetName(), nil)
		}
		graphW := scheduler.graph.Write(false)
		graphW.RegisterMetadataMap(descriptor.GetName(), metadataMap)
		graphW.Save()
		graphW.Release()
	}
}

// GetRegisteredNBKeyPrefixes returns a list of key prefixes from NB with values
// described by registered descriptors and therefore managed by the scheduler.
func (scheduler *Scheduler) GetRegisteredNBKeyPrefixes() []string {
	return scheduler.keyPrefixes
}

// StartNBTransaction starts a new transaction from NB to SB plane.
// The enqueued actions are scheduled for execution by Txn.Commit().
func (scheduler *Scheduler) StartNBTransaction(opts ...TxnOption) Txn {
	txn := &SchedulerTxn{
		scheduler: scheduler,
		data: &queuedTxn{
			txnType: nbTransaction,
			nb: &nbTxn{
				isBlocking: true,
				valueData:  make(map[string]interface{}),
			},
		},
	}

	for _, opt := range opts {
		switch option := opt.(type) {
		case *NonBlockingTxn:
			txn.data.nb.isBlocking = false
		case *RetryFailedOps:
			txn.data.nb.retryFailed = true
			txn.data.nb.retryPeriod = option.Period
			txn.data.nb.expBackoffRetry = option.ExpBackoff
		case *RevertOnFailure:
			txn.data.nb.revertOnFailure = true
		}
	}

	if txn.data.nb.isBlocking {
		txn.data.nb.resultChan = make(chan []KeyWithError, 1)
	}
	return txn
}

// PushSBNotification notifies about a spontaneous value change in the SB
// plane (i.e. not triggered by NB transaction).
func (scheduler *Scheduler) PushSBNotification(key string, value Value, metadata Metadata) error {
	txn := &queuedTxn{
		txnType: sbNotification,
		sb: &sbNotif{
			value: KeyValuePair{Key: key, Value: value},
		},
	}
	return scheduler.enqueueTxn(txn)
}

// GetValue currently set for the given key.
// The function can be used from within a transaction. However, if update
// of A uses the value of B, then A should be marked as dependent on B
// so that the scheduler can ensure that B is updated before A is.
func (scheduler *Scheduler) GetValue(key string) Value {
	graphR := scheduler.graph.Read()
	defer graphR.Release()

	node := graphR.GetNode(key)
	if node != nil {
		return node.GetValue()
	}
	return nil
}

// GetValues returns a set of values matched by the given selector.
func (scheduler *Scheduler) GetValues(selector KeySelector) []KeyValuePair {
	graphR := scheduler.graph.Read()
	defer graphR.Release()

	nodes := graphR.GetNodes(selector)
	return nodesToKVPairs(nodes)
}

// GetMetadataMap returns (read-only) map associating value label with value
// metadata of a given descriptor.
// Returns nil if the descriptor does not expose metadata.
func (scheduler *Scheduler) GetMetadataMap(descriptor string) idxmap.NamedMapping {
	graphR := scheduler.graph.Read()
	defer graphR.Release()

	return graphR.GetMetadataMap(descriptor)
}

// GetPendingValues returns list of values (possibly filtered by selector)
// waiting for their dependencies to be met.
func (scheduler *Scheduler) GetPendingValues(selector KeySelector) []KeyValuePair {
	graphR := scheduler.graph.Read()
	defer graphR.Release()

	nodes := graphR.GetNodes(selector, graph.WithFlags(&PendingFlag{}))
	return nodesToKVPairs(nodes)
}

// GetFailedValues returns a list of keys (possibly filtered by selector)
// whose values are in a failed state (i.e. possibly not in the state as set
// by the last transaction).
func (scheduler *Scheduler) GetFailedValues(selector KeySelector) []KeyWithError {
	graphR := scheduler.graph.Read()
	defer graphR.Release()

	nodes := graphR.GetNodes(selector, graph.WithFlags(&ErrorFlag{}))
	return nodesToKeysWithError(nodes)
}

// SubscribeForErrors allows to get notified about all failed (Error!=nil)
// and restored (Error==nil) values (possibly filtered using the selector).
func (scheduler *Scheduler) SubscribeForErrors(channel chan<- KeyWithError, selector KeySelector) {
	scheduler.errorSubs = append(scheduler.errorSubs, errorSubscription{channel: channel, selector: selector})
}

// SetValueData changes (non-derived) value data.
// NB provides untyped data which are build into the new value for the given
// key by descriptor (method BuildValue).
// If <valueData> is nil, the value will get deleted.
func (txn *SchedulerTxn) SetValueData(key string, valueData interface{}) Txn {
	if txn.data.nb.isResync {
		txn.err = ErrCombinedResyncWithChange
		return txn
	}
	txn.data.nb.valueData[key] = valueData
	return txn
}

// Resync all NB-values to match with <values>.
// The list should consist of non-derived values only - derived values will
// get created automatically using descriptors.
// Run in case the SB may be out-of-sync with NB or with the scheduler
// itself.
func (txn *SchedulerTxn) Resync(values []KeyValueDataPair) Txn {
	txn.data.nb.isResync = true
	if len(txn.data.nb.valueData) > 0 {
		txn.err = ErrCombinedResyncWithChange
		return txn
	}
	for _, value := range values {
		txn.data.nb.valueData[value.Key] = value.ValueData
	}

	return txn
}

// Commit orders scheduler to execute enqueued operations.
// Operations with unmet dependencies will get postponed and possibly
// executed later.
func (txn *SchedulerTxn) Commit(ctx context.Context) (kvErrors []KeyWithError, txnError error) {
	if txn.err != nil {
		return nil, txn.err
	}
	err := txn.scheduler.enqueueTxn(txn.data)
	if err != nil {
		return nil, err
	}
	if txn.data.nb.isBlocking {
		select {
		case <-txn.scheduler.ctx.Done():
			return nil, ErrClosedScheduler
		case <-ctx.Done():
			return nil, ErrTxnWaitCanceled
		case kvErrors = <-txn.data.nb.resultChan:
			close(txn.data.nb.resultChan)
			return kvErrors, nil
		}
	}
	return nil, nil
}
