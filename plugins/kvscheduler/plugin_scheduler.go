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
	"errors"
	"os"
	"runtime/trace"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/idxmap/mem"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/rpc/rest"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/graph"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/registry"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/utils"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
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

	// by default, all NB transactions and SB notifications are run without
	// simulation (Retries are always first simulated)
	defaultEnableTxnSimulation = false

	// by default, a concise summary of every processed transactions is printed
	// to stdout
	defaultPrintTxnSummary = true

	// name of the environment variable used to enable verification after every transaction
	verifyModeEnv = "KVSCHED_VERIFY_MODE"

	// name of the environment variable used to turn on automatic check for
	// the preservation of the original network namespace after descriptor operations
	checkNetNamespaceEnv = "KVSCHED_CHECK_NET_NS"

	// name of the environment variable used to trigger log messages showing
	// graph traversal
	logGraphWalkEnv = "KVSCHED_LOG_GRAPH_WALK"
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

	// in-memory representation of all created+pending kv-pairs and their dependencies
	graph graph.Graph

	// registry for descriptors
	registry registry.Registry

	// a list of key prefixed covered by registered descriptors
	keyPrefixes []string

	// TXN processing
	txnLock      sync.Mutex // can be used to pause transaction processing; always lock before the graph!
	txnQueue     chan *transaction
	txnSeqNumber uint64
	resyncCount  uint

	// value status
	updatedStates    utils.KeySet // base values with updated status
	valStateWatchers []valStateWatcher

	// TXN history
	historyLock sync.Mutex
	txnHistory  []*kvs.RecordedTxn // ordered from the oldest to the latest
	startTime   time.Time

	// debugging
	verifyMode   bool
	logGraphWalk bool
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
	EnableTxnSimulation           bool   `json:"enable-txn-simulation"`
	PrintTxnSummary               bool   `json:"print-txn-summary"`
}

// SchedulerTxn implements transaction for the KV scheduler.
type SchedulerTxn struct {
	scheduler *Scheduler
	values    map[string]proto.Message
}

// valStateWatcher represents one subscription for value state updates.
type valStateWatcher struct {
	channel  chan<- *kvscheduler.BaseValueStatus
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
		EnableTxnSimulation:           defaultEnableTxnSimulation,
		PrintTxnSummary:               defaultPrintTxnSummary,
	}

	// load configuration
	err := s.loadConfig(s.config)
	if err != nil {
		s.Log.Error(err)
		return err
	}
	s.Log.Debugf("KVScheduler configuration: %+v", *s.config)

	// prepare context for all go routines
	s.ctx, s.cancel = context.WithCancel(context.Background())
	// initialize graph for in-memory storage of key-value pairs
	graphOpts := graph.Opts{
		RecordOldRevs:       s.config.RecordTransactionHistory,
		RecordAgeLimit:      s.config.TransactionHistoryAgeLimit,
		PermanentInitPeriod: s.config.PermanentlyRecordedInitPeriod,
		MethodTracker:       trackGraphMethod,
	}
	s.graph = graph.NewGraph(graphOpts)
	// initialize registry for key->descriptor lookups
	s.registry = registry.NewRegistry()
	// prepare channel for serializing transactions
	s.txnQueue = make(chan *transaction, 100)
	reportQueueCap(cap(s.txnQueue))
	// register REST API handlers
	s.registerHandlers(s.HTTPHandlers)
	// initialize key-set used to mark values with updated status
	s.updatedStates = utils.NewSliceBasedKeySet()
	// record startup time
	s.startTime = time.Now()

	// enable or disable debugging mode
	s.verifyMode = os.Getenv(verifyModeEnv) != ""
	s.logGraphWalk = os.Getenv(logGraphWalkEnv) != ""

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

// RegisterKVDescriptor registers descriptor(s) for a set of selected
// keys. It should be called in the Init phase of agent plugins.
// Every key-value pair must have at most one descriptor associated with it
// (none for derived values expressing properties).
func (s *Scheduler) RegisterKVDescriptor(descriptors ...*kvs.KVDescriptor) error {
	for _, d := range descriptors {
		err := s.registerKVDescriptor(d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) registerKVDescriptor(descriptor *kvs.KVDescriptor) error {
	// TODO: validate descriptor
	if s.registry.GetDescriptor(descriptor.Name) != nil {
		return kvs.ErrDescriptorExists
	}

	stats.addDescriptor(descriptor.Name)

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
		graphW := s.graph.Write(true, false)
		graphW.RegisterMetadataMap(descriptor.Name, metadataMap)
		graphW.Release()
	}
	return nil
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
		values:    make(map[string]proto.Message),
	}
	return txn
}

// TransactionBarrier ensures that all notifications received prior to the call
// are associated with transactions that have already finalized.
func (s *Scheduler) TransactionBarrier() {
	s.txnLock.Lock()
	s.txnLock.Unlock()
}

// PushSBNotification notifies about a spontaneous value change(s) in the SB
// plane (i.e. not triggered by NB transaction).
func (s *Scheduler) PushSBNotification(notif ...kvs.KVWithMetadata) error {
	txn := &transaction{
		txnType: kvs.SBNotification,
		created: time.Now(),
	}
	for _, value := range notif {
		txn.values = append(txn.values, kvForTxn{
			key:      value.Key,
			value:    value.Value,
			metadata: value.Metadata,
			origin:   kvs.FromSB,
		})
	}
	return s.enqueueTxn(txn)
}

// GetMetadataMap returns (read-only) map associating value label with value
// metadata of a given descriptor.
// Returns nil if the descriptor does not expose metadata.
func (s *Scheduler) GetMetadataMap(descriptor string) idxmap.NamedMapping {
	graphR := s.graph.Read()
	defer graphR.Release()

	return graphR.GetMetadataMap(descriptor)
}

// GetValueStatus returns the status of a non-derived value with the given
// key.
func (s *Scheduler) GetValueStatus(key string) *kvscheduler.BaseValueStatus {
	graphR := s.graph.Read()
	defer graphR.Release()
	return getValueStatus(graphR.GetNode(key), key)
}

// WatchValueStatus allows to watch for changes in the status of non-derived
// values with keys selected by the selector (all if keySelector==nil).
func (s *Scheduler) WatchValueStatus(channel chan<- *kvscheduler.BaseValueStatus, keySelector kvs.KeySelector) {
	s.txnLock.Lock()
	defer s.txnLock.Unlock()
	s.valStateWatchers = append(s.valStateWatchers, valStateWatcher{
		channel:  channel,
		selector: keySelector,
	})
}

// DumpValuesByDescriptor dumps values associated with the given
// descriptor as viewed from either NB (what was requested to be applied),
// SB (what is actually applied) or from the inside (what kvscheduler's
// cached view of SB is).
func (s *Scheduler) DumpValuesByDescriptor(descriptor string, view kvs.View) (values []kvs.KVWithMetadata, err error) {
	if view == kvs.SBView {
		// pause transaction processing
		s.txnLock.Lock()
		defer s.txnLock.Unlock()
	}

	graphR := s.graph.Read()
	defer graphR.Release()

	if view == kvs.NBView {
		// return the intended state
		var kvPairs []kvs.KVWithMetadata
		nbNodes := graphR.GetNodes(nil,
			graph.WithFlags(&DescriptorFlag{descriptor}),
			graph.WithoutFlags(&DerivedFlag{}, &ValueStateFlag{kvscheduler.ValueState_OBTAINED}))

		for _, node := range nbNodes {
			lastUpdate := getNodeLastUpdate(node)
			if lastUpdate == nil || lastUpdate.value == nil {
				// filter found NB values and values requested to be deleted
				continue
			}
			kvPairs = append(kvPairs, kvs.KVWithMetadata{
				Key:      node.GetKey(),
				Value:    lastUpdate.value,
				Origin:   kvs.FromNB,
				Metadata: node.GetMetadata(),
			})
		}
		return kvPairs, nil
	}

	/* Cached/SB: */

	// retrieve from the in-memory graph first (for Retrieve it is used for correlation)
	inMemNodes := nodesToKVPairsWithMetadata(
		graphR.GetNodes(nil, descrValsSelectors(descriptor, true)...))

	if view == kvs.CachedView {
		// return the scheduler's view of SB for the given descriptor
		return inMemNodes, nil
	}

	// obtain Retrieve handler from the descriptor
	kvDescriptor := s.registry.GetDescriptor(descriptor)
	if kvDescriptor == nil {
		err = errors.New("descriptor is not registered")
		return
	}
	if kvDescriptor.Retrieve == nil {
		err = errors.New("descriptor does not support Retrieve operation")
		return
	}

	// retrieve the state directly from SB via descriptor
	values, err = kvDescriptor.Retrieve(inMemNodes)
	return
}

func (s *Scheduler) getDescriptorForKeyPrefix(keyPrefix string) string {
	var descriptorName string
	s.txnLock.Lock()
	for _, descriptor := range s.registry.GetAllDescriptors() {
		if descriptor.NBKeyPrefix == keyPrefix {
			descriptorName = descriptor.Name
		}
	}
	s.txnLock.Unlock()
	return descriptorName
}

// DumpValuesByKeyPrefix like DumpValuesByDescriptor returns a dump of values,
// but the descriptor is selected based on the key prefix.
func (s *Scheduler) DumpValuesByKeyPrefix(keyPrefix string, view kvs.View) (values []kvs.KVWithMetadata, err error) {
	descriptorName := s.getDescriptorForKeyPrefix(keyPrefix)
	if descriptorName == "" {
		err = errors.New("unknown key prefix")
		return
	}
	return s.DumpValuesByDescriptor(descriptorName, view)
}

// SetValue changes (non-derived) value.
// If <value> is nil, the value will get deleted.
func (txn *SchedulerTxn) SetValue(key string, value proto.Message) kvs.Txn {
	txn.values[key] = value
	return txn
}

// Commit orders scheduler to execute enqueued operations.
// Operations with unmet dependencies will get postponed and possibly
// executed later.
func (txn *SchedulerTxn) Commit(ctx context.Context) (txnSeqNum uint64, err error) {
	ctx, task := trace.NewTask(ctx, "scheduler.Commit")
	defer task.End()

	txnSeqNum = ^uint64(0)

	txnData := &transaction{
		ctx:     ctx,
		txnType: kvs.NBTransaction,
		nb:      &nbTxn{},
		values:  make([]kvForTxn, 0, len(txn.values)),
		created: time.Now(),
	}

	// collect values
	for key, value := range txn.values {
		txnData.values = append(txnData.values, kvForTxn{
			key:    key,
			value:  value,
			origin: kvs.FromNB,
		})
	}

	// parse transaction options
	txnData.nb.isBlocking = !kvs.IsNonBlockingTxn(ctx)
	txnData.nb.resyncType, txnData.nb.verboseRefresh = kvs.IsResync(ctx)
	txnData.nb.retryArgs, txnData.nb.retryEnabled = kvs.IsWithRetry(ctx)
	txnData.nb.revertOnFailure = kvs.IsWithRevert(ctx)
	txnData.nb.description, _ = kvs.IsWithDescription(ctx)
	txnData.nb.withSimulation = txn.scheduler.config.EnableTxnSimulation || kvs.IsWithSimulation(ctx)

	// validate transaction options
	if txnData.nb.resyncType == kvs.DownstreamResync && len(txnData.values) > 0 {
		return txnSeqNum, kvs.NewTransactionError(kvs.ErrCombinedDownstreamResyncWithChange, nil)
	}
	if txnData.nb.revertOnFailure && txnData.nb.resyncType != kvs.NotResync {
		return txnSeqNum, kvs.NewTransactionError(kvs.ErrRevertNotSupportedWithResync, nil)
	}

	// enqueue txn and for blocking Commit wait for the errors
	if txnData.nb.isBlocking {
		txnData.nb.resultChan = make(chan txnResult, 1)
	}

	err = txn.scheduler.enqueueTxn(txnData)
	if err != nil {
		return txnSeqNum, kvs.NewTransactionError(err, nil)
	}
	if txnData.nb.isBlocking {
		select {
		case <-txn.scheduler.ctx.Done():
			return txnSeqNum, kvs.NewTransactionError(kvs.ErrClosedScheduler, nil)
		case <-ctx.Done():
			return txnSeqNum, kvs.NewTransactionError(kvs.ErrTxnWaitCanceled, nil)
		case txnResult := <-txnData.nb.resultChan:
			close(txnData.nb.resultChan)
			trace.Logf(ctx, "txnSeqNum", "%d", txnResult.txnSeqNum)
			return txnResult.txnSeqNum, txnResult.err
		}
	}
	return txnSeqNum, nil
}
