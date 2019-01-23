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

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/idxmap/mem"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/rpc/rest"

	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/registry"
)

const (
	// DependencyRelation identifies dependency relation for the graph.
	DependencyRelation = "depends-on"

	// DerivesRelation identifies relation of value derivation for the graph.
	DerivesRelation = "derives"

	// how often the transaction history gets trimmed to remove records too old to keep
	txnHistoryTrimmingPeriod = 1 * time.Minute

	// by default, a history of processed transaction is recorded
	defaultRecordTransactionHistory = true

	// by default, only transaction processed in the last 24 hours are kept recorded
	// (with the exception of permanently recorded init period)
	defaultTransactionHistoryAgeLimit = 24 * 60 // in minutes

	// by default, transactions from the first hour of runtime stay permanently
	// recorded
	defaultPermanentlyRecordedInitPeriod = 60 // in minutes
)

// Scheduler is a CN-infra plugin implementing KVScheduler.
// Detailed documentation can be found in the "api" and "docs" sub-folders.
type Scheduler struct {
	Deps

	// configuration
	config *Config

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
	txnSeqNumber uint64
	resyncCount  uint
	lastError    map[string]error // key -> error

	// TXN history
	historyLock sync.Mutex
	txnHistory  []*kvs.RecordedTxn // ordered from the oldest to the latest
	startTime   time.Time
}

// Deps lists dependencies of the scheduler.
type Deps struct {
	infra.PluginDeps
	HTTPHandlers rest.HTTPHandlers
}

// Config holds the KVScheduler configuration.
type Config struct {
	RecordTransactionHistory      bool   `json:"record-transaction-history"`
	TransactionHistoryAgeLimit    uint32 `json:"transaction-history-age-limit"`    // in minutes
	PermanentlyRecordedInitPeriod uint32 `json:"permanently-recorded-init-period"` // in minutes
}

// SchedulerTxn implements transaction for the KV scheduler.
type SchedulerTxn struct {
	scheduler *Scheduler
	data      *queuedTxn
}

// errorSubscription represents one subscription for error updates.
type errorSubscription struct {
	channel  chan<- kvs.KeyWithError
	selector kvs.KeySelector
}

// Init initializes the scheduler. Single go routine is started that will process
// all the transactions synchronously.
func (s *Scheduler) Init() error {
	// default configuration
	s.config = &Config{
		RecordTransactionHistory:      defaultRecordTransactionHistory,
		TransactionHistoryAgeLimit:    defaultTransactionHistoryAgeLimit,
		PermanentlyRecordedInitPeriod: defaultPermanentlyRecordedInitPeriod,
	}

	// load configuration
	err := s.loadConfig(s.config)
	if err != nil {
		s.Log.Error(err)
		return err
	}
	s.Log.Infof("KVScheduler configuration: %+v", *s.config)

	// prepare context for all go routines
	s.ctx, s.cancel = context.WithCancel(context.Background())
	// initialize graph for in-memory storage of added+pending kv pairs
	s.graph = graph.NewGraph(s.config.RecordTransactionHistory, s.config.TransactionHistoryAgeLimit,
		s.config.PermanentlyRecordedInitPeriod)
	// initialize registry for key->descriptor lookups
	s.registry = registry.NewRegistry()
	// prepare channel for serializing transactions
	s.txnQueue = make(chan *queuedTxn, 100)
	// map of last errors (even for nodes not in the graph anymore)
	s.lastError = make(map[string]error)
	// register REST API handlers
	s.registerHandlers(s.HTTPHandlers)
	// record startup time
	s.startTime = time.Now()

	// go routine processing serialized transactions
	s.wg.Add(1)
	go s.consumeTransactions()

	// go routine periodically removing transaction records too old to keep
	if s.config.RecordTransactionHistory {
		s.wg.Add(1)
		go s.transactionHistoryTrimming()
	}
	return nil
}

// loadConfig loads configuration file.
func (s *Scheduler) loadConfig(config *Config) error {
	found, err := s.Cfg.LoadValue(config)
	if err != nil {
		return err
	} else if !found {
		s.Log.Debugf("%v config not found", s.PluginName)
		return nil
	}
	s.Log.Debugf("%v config found: %+v", s.PluginName, config)
	return err
}

// Close stops all the go routines.
func (s *Scheduler) Close() error {
	s.cancel()
	s.wg.Wait()
	return nil
}

// RegisterKVDescriptor registers descriptor for a set of selected
// keys. It should be called in the Init phase of agent plugins.
// Every key-value pair must have at most one descriptor associated with it
// (none for derived values expressing properties).
func (s *Scheduler) RegisterKVDescriptor(descriptor *kvs.KVDescriptor) {
	s.registry.RegisterDescriptor(descriptor)
	if descriptor.NBKeyPrefix != "" {
		s.keyPrefixes = append(s.keyPrefixes, descriptor.NBKeyPrefix)
	}

	if descriptor.WithMetadata {
		var metadataMap idxmap.NamedMappingRW
		if descriptor.MetadataMapFactory != nil {
			metadataMap = descriptor.MetadataMapFactory()
		} else {
			metadataMap = mem.NewNamedMapping(s.Log, descriptor.Name, nil)
		}
		graphW := s.graph.Write(false)
		graphW.RegisterMetadataMap(descriptor.Name, metadataMap)
		graphW.Save()
		graphW.Release()
	}
}

// GetRegisteredNBKeyPrefixes returns a list of key prefixes from NB with values
// described by registered descriptors and therefore managed by the scheduler.
func (s *Scheduler) GetRegisteredNBKeyPrefixes() []string {
	return s.keyPrefixes
}

// StartNBTransaction starts a new transaction from NB to SB plane.
// The enqueued actions are scheduled for execution by Txn.Commit().
func (s *Scheduler) StartNBTransaction() kvs.Txn {
	txn := &SchedulerTxn{
		scheduler: s,
		data: &queuedTxn{
			txnType: kvs.NBTransaction,
			nb: &nbTxn{
				value: make(map[string]datasync.LazyValue),
			},
		},
	}
	return txn
}

// TransactionBarrier ensures that all notifications received prior to the call
// are associated with transactions that have already finalized.
func (s *Scheduler) TransactionBarrier() {
	s.txnLock.Lock()
	s.txnLock.Unlock()
}

// PushSBNotification notifies about a spontaneous value change in the SB
// plane (i.e. not triggered by NB transaction).
func (s *Scheduler) PushSBNotification(key string, value proto.Message, metadata kvs.Metadata) error {
	txn := &queuedTxn{
		txnType: kvs.SBNotification,
		sb: &sbNotif{
			value:    kvs.KeyValuePair{Key: key, Value: value},
			metadata: metadata,
		},
	}
	return s.enqueueTxn(txn)
}

// GetValue currently set for the given key.
// The function can be used from within a transaction. However, if update
// of A uses the value of B, then A should be marked as dependent on B
// so that the scheduler can ensure that B is updated before A is.
func (s *Scheduler) GetValue(key string) proto.Message {
	graphR := s.graph.Read()
	defer graphR.Release()

	node := graphR.GetNode(key)
	if node != nil {
		return node.GetValue()
	}
	return nil
}

// GetValues returns a set of values matched by the given selector.
func (s *Scheduler) GetValues(selector kvs.KeySelector) []kvs.KeyValuePair {
	graphR := s.graph.Read()
	defer graphR.Release()

	nodes := graphR.GetNodes(selector)
	return nodesToKVPairs(nodes)
}

// GetMetadataMap returns (read-only) map associating value label with value
// metadata of a given descriptor.
// Returns nil if the descriptor does not expose metadata.
func (s *Scheduler) GetMetadataMap(descriptor string) idxmap.NamedMapping {
	graphR := s.graph.Read()
	defer graphR.Release()

	return graphR.GetMetadataMap(descriptor)
}

// GetPendingValues returns list of values (possibly filtered by selector)
// waiting for their dependencies to be met.
func (s *Scheduler) GetPendingValues(selector kvs.KeySelector) []kvs.KeyValuePair {
	graphR := s.graph.Read()
	defer graphR.Release()

	nodes := graphR.GetNodes(selector, graph.WithFlags(&PendingFlag{}))
	return nodesToKVPairs(nodes)
}

// GetFailedValues returns a list of keys (possibly filtered by selector)
// whose (base) values are in a failed state (i.e. possibly not in the state as set
// by the last transaction).
func (s *Scheduler) GetFailedValues(selector kvs.KeySelector) []kvs.KeyWithError {
	graphR := s.graph.Read()
	defer graphR.Release()

	nodes := graphR.GetNodes(selector, graph.WithFlags(&ErrorFlag{}))
	return nodesToKeysWithError(nodes)
}

// SubscribeForErrors allows to get notified about all failed (Error!=nil)
// and restored (Error==nil) values (possibly filtered using the selector).
func (s *Scheduler) SubscribeForErrors(channel chan<- kvs.KeyWithError, selector kvs.KeySelector) {
	s.errorSubs = append(s.errorSubs, errorSubscription{channel: channel, selector: selector})
}

// SetValue changes (non-derived) lazy value - un-marshalled during
// transaction pre-processing using ValueTypeName given by descriptor.
// If <value> is nil, the value will get deleted.
func (txn *SchedulerTxn) SetValue(key string, value datasync.LazyValue) kvs.Txn {
	txn.data.nb.value[key] = value
	return txn
}

// Commit orders scheduler to execute enqueued operations.
// Operations with unmet dependencies will get postponed and possibly
// executed later.
func (txn *SchedulerTxn) Commit(ctx context.Context) (txnSeqNum uint64, err error) {
	txnSeqNum = ^uint64(0)

	// parse transaction options
	txn.data.nb.isBlocking = !kvs.IsNonBlockingTxn(ctx)
	txn.data.nb.resyncType, txn.data.nb.verboseRefresh = kvs.IsResync(ctx)
	txn.data.nb.retryPeriod, txn.data.nb.expBackoffRetry, txn.data.nb.retryFailed = kvs.IsWithRetry(ctx)
	txn.data.nb.revertOnFailure = kvs.IsWithRevert(ctx)
	txn.data.nb.description, _ = kvs.IsWithDescription(ctx)

	// validate transaction options
	if txn.data.nb.resyncType == kvs.DownstreamResync && len(txn.data.nb.value) > 0 {
		return txnSeqNum, kvs.NewTransactionError(kvs.ErrCombinedDownstreamResyncWithChange, nil)
	}
	if txn.data.nb.revertOnFailure && txn.data.nb.resyncType != kvs.NotResync {
		return txnSeqNum, kvs.NewTransactionError(kvs.ErrRevertNotSupportedWithResync, nil)
	}

	// enqueue txn and for blocking Commit wait for the errors
	if txn.data.nb.isBlocking {
		txn.data.nb.resultChan = make(chan txnResult, 1)
	}
	err = txn.scheduler.enqueueTxn(txn.data)
	if err != nil {
		return txnSeqNum, kvs.NewTransactionError(err, nil)
	}
	if txn.data.nb.isBlocking {
		select {
		case <-txn.scheduler.ctx.Done():
			return txnSeqNum, kvs.NewTransactionError(kvs.ErrClosedScheduler, nil)
		case <-ctx.Done():
			return txnSeqNum, kvs.NewTransactionError(kvs.ErrTxnWaitCanceled, nil)
		case txnResult := <-txn.data.nb.resultChan:
			close(txn.data.nb.resultChan)
			return txnResult.txnSeqNum, txnResult.err
		}
	}
	return txnSeqNum, nil
}
