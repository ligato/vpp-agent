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
	"time"

	. "github.com/ligato/vpp-agent/plugins/kvscheduler/api"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/idxmap/mem"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/registry"
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
	txnLock      sync.Mutex // can be used to pause transaction processing; always lock before the graph!
	txnQueue     chan *queuedTxn
	errorSubs    []errorSubscription
	txnSeqNumber uint
	resyncCount  uint
	lastError    map[string]error // key -> error

	// TXN history
	historyLock sync.Mutex
	txnHistory  []*recordedTxn // ordered from the oldest to the latest

	// datasync channels
	changeChan   chan datasync.ChangeEvent
	resyncChan   chan datasync.ResyncEvent
	watchDataReg datasync.WatchRegistration
}

// Deps lists dependencies of the scheduler.
type Deps struct {
	infra.PluginName
	Log          logging.PluginLogger
	HTTPHandlers rest.HTTPHandlers
	Watcher      datasync.KeyValProtoWatcher
}

// SchedulerTxn implements transaction for the KV scheduler.
type SchedulerTxn struct {
	scheduler *Scheduler
	data      *queuedTxn
}

// errorSubscription represents one subscription for error updates.
type errorSubscription struct {
	channel  chan<- KeyWithError
	selector KeySelector
}

// Init initializes the scheduler. Single go routine is started that will process
// all the transactions synchronously.
func (scheduler *Scheduler) Init() error {
	// initialize datasync channels
	scheduler.resyncChan = make(chan datasync.ResyncEvent)
	scheduler.changeChan = make(chan datasync.ChangeEvent)

	// prepare context for all go routines
	scheduler.ctx, scheduler.cancel = context.WithCancel(context.Background())
	// initialize graph for in-memory storage of added+pending kv pairs
	scheduler.graph = graph.NewGraph()
	// initialize registry for key->descriptor lookups
	scheduler.registry = registry.NewRegistry()
	// prepare channel for serializing transactions
	scheduler.txnQueue = make(chan *queuedTxn, 100)
	// map of last errors (even for nodes not in the graph anymore)
	scheduler.lastError = make(map[string]error)
	// register REST API handlers
	scheduler.registerHandlers(scheduler.HTTPHandlers)
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

// AfterInit subscribes to known NB prefixes.
func (scheduler *Scheduler) AfterInit() error {
	go scheduler.watchEvents()

	var err error
	scheduler.watchDataReg, err = scheduler.Watcher.Watch("scheduler",
		scheduler.changeChan, scheduler.resyncChan, scheduler.GetRegisteredNBKeyPrefixes()...)
	if err != nil {
		return err
	}

	return nil
}

func (scheduler *Scheduler) watchEvents() {
	for {
		select {
		case e := <-scheduler.changeChan:
			scheduler.Log.Debugf("=> SCHEDULER received CHANGE EVENT: %v changes", len(e.GetChanges()))

			txn := scheduler.StartNBTransaction()
			for _, x := range e.GetChanges() {
				scheduler.Log.Debugf("  - Change %v: %q (rev: %v)",
					x.GetChangeType(), x.GetKey(), x.GetRevision())
				if x.GetChangeType() == datasync.Delete {
					txn.SetValue(x.GetKey(), nil)
				} else {
					txn.SetValue(x.GetKey(), x)
				}
			}
			kvErrs, err := txn.Commit(WithRetry(context.Background(), time.Second, true))
			scheduler.Log.Debugf("commit result: err=%v kvErrs=%+v", err, kvErrs)
			e.Done(err)

		case e := <-scheduler.resyncChan:
			scheduler.Log.Debugf("=> SCHEDULER received RESYNC EVENT: %v prefixes", len(e.GetValues()))

			txn := scheduler.StartNBTransaction()
			for prefix, iter := range e.GetValues() {
				var keyVals []datasync.KeyVal
				for x, done := iter.GetNext(); done == false; x, done = iter.GetNext() {
					keyVals = append(keyVals, x)
					txn.SetValue(x.GetKey(), x)
				}
				scheduler.Log.Debugf(" - Resync: %q (%v key-values)", prefix, len(keyVals))
				for _, x := range keyVals {
					scheduler.Log.Debugf("\t%q: (rev: %v)", x.GetKey(), x.GetRevision())
				}
			}
			ctx := context.Background()
			ctx = WithRetry(ctx, time.Second, true)
			ctx = WithFullResync(ctx)
			kvErrs, err := txn.Commit(ctx)
			scheduler.Log.Debugf("commit result: err=%v kvErrs=%+v", err, kvErrs)
			e.Done(err)
		}
	}
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
// (none for derived values expressing properties).
func (scheduler *Scheduler) RegisterKVDescriptor(descriptor *KVDescriptor) {
	scheduler.registry.RegisterDescriptor(descriptor)
	if descriptor.NBKeyPrefix != "" {
		scheduler.keyPrefixes = append(scheduler.keyPrefixes, descriptor.NBKeyPrefix)
	}

	if descriptor.WithMetadata {
		var metadataMap idxmap.NamedMappingRW
		if descriptor.MetadataMapFactory != nil {
			metadataMap = descriptor.MetadataMapFactory()
		} else {
			metadataMap = mem.NewNamedMapping(scheduler.Log, descriptor.Name, nil)
		}
		graphW := scheduler.graph.Write(false)
		graphW.RegisterMetadataMap(descriptor.Name, metadataMap)
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
func (scheduler *Scheduler) StartNBTransaction() Txn {
	txn := &SchedulerTxn{
		scheduler: scheduler,
		data: &queuedTxn{
			txnType: nbTransaction,
			nb: &nbTxn{
				value: make(map[string]datasync.LazyValue),
			},
		},
	}
	return txn
}

// TransactionBarrier ensures that all notifications received prior to the call
// are associated with transactions that have already finalized.
func (scheduler *Scheduler) TransactionBarrier() {
	scheduler.txnLock.Lock()
	scheduler.txnLock.Unlock()
}

// PushSBNotification notifies about a spontaneous value change in the SB
// plane (i.e. not triggered by NB transaction).
func (scheduler *Scheduler) PushSBNotification(key string, value proto.Message, metadata Metadata) error {
	txn := &queuedTxn{
		txnType: sbNotification,
		sb: &sbNotif{
			value:    KeyValuePair{Key: key, Value: value},
			metadata: metadata,
		},
	}
	return scheduler.enqueueTxn(txn)
}

// GetValue currently set for the given key.
// The function can be used from within a transaction. However, if update
// of A uses the value of B, then A should be marked as dependent on B
// so that the scheduler can ensure that B is updated before A is.
func (scheduler *Scheduler) GetValue(key string) proto.Message {
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
// whose (base) values are in a failed state (i.e. possibly not in the state as set
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

// SetValue changes (non-derived) lazy value - un-marshalled during
// transaction pre-processing using ValueTypeName given by descriptor.
// If <value> is nil, the value will get deleted.
func (txn *SchedulerTxn) SetValue(key string, value datasync.LazyValue) Txn {
	txn.data.nb.value[key] = value
	return txn
}

// Commit orders scheduler to execute enqueued operations.
// Operations with unmet dependencies will get postponed and possibly
// executed later.
func (txn *SchedulerTxn) Commit(ctx context.Context) (kvErrors []KeyWithError, txnError error) {
	// parse transaction options
	txn.data.nb.isBlocking = !IsNonBlockingTxn(ctx)
	txn.data.nb.retryPeriod, txn.data.nb.expBackoffRetry, txn.data.nb.retryFailed = IsWithRetry(ctx)
	txn.data.nb.revertOnFailure = IsWithRevert(ctx)
	txn.data.nb.isFullResync = IsFullResync(ctx)
	txn.data.nb.isDownstreamResync = IsDownstreamResync(ctx)
	txn.data.nb.description, _ = IsWithDescription(ctx)

	// validate transaction options
	if txn.data.nb.isFullResync {
		// full resync overrides downstream resync
		txn.data.nb.isDownstreamResync = false
	}
	if txn.data.nb.isDownstreamResync && len(txn.data.nb.value) > 0 {
		return nil, ErrCombinedDownstreamResyncWithChange
	}
	if txn.data.nb.revertOnFailure &&
		(txn.data.nb.isDownstreamResync || txn.data.nb.isFullResync) {
		return nil, ErrRevertNotSupportedWithResync
	}

	// enqueue txn and for blocking Commit wait for the errors
	if txn.data.nb.isBlocking {
		txn.data.nb.resultChan = make(chan []KeyWithError, 1)
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
