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
	"runtime/trace"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/ligato/cn-infra/logging"

	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// transaction represents kscheduler transaction that is being queued/processed.
// Once finalized, it is recorded as instance of RecordedTxn and these data
// are thrown away.
type transaction struct {
	ctx     context.Context
	seqNum  uint64
	txnType kvs.TxnType
	values  []kvForTxn
	nb      *nbTxn    // defined for NB transactions
	retry   *retryTxn // defined for retry of failed operations
}

// kvForTxn represents a new value for a given key to be applied in a transaction.
type kvForTxn struct {
	key      string
	value    proto.Message
	metadata kvs.Metadata
	origin   kvs.ValueOrigin
	isRevert bool
}

// nbTxn encapsulates data for NB transaction.
type nbTxn struct {
	resyncType     kvs.ResyncType
	verboseRefresh bool
	isBlocking     bool

	retryEnabled bool
	retryArgs    *kvs.RetryOpt

	revertOnFailure bool
	withSimulation  bool
	description     string
	resultChan      chan txnResult
}

// retryTxn encapsulates data for retry of failed operations.
type retryTxn struct {
	retryTxnMeta
	keys map[string]uint64 // key -> value revision (last update) when the retry was enqueued
}

// retryTxnMeta contains metadata for Retry transaction.
type retryTxnMeta struct {
	txnSeqNum uint64
	delay     time.Duration
	attempt   int
}

// txnResult represents transaction result.
type txnResult struct {
	err       error
	txnSeqNum uint64
}

// consumeTransactions pulls the oldest queued transaction and starts the processing.
func (s *Scheduler) consumeTransactions() {
	defer s.wg.Done()
	for {
		txn, canceled := s.dequeueTxn()
		if canceled {
			return
		}
		s.processTransaction(txn)
		atomic.AddUint64(&stats.TransactionsProcessed, 1)
	}
}

// processTransaction processes transaction in 6 steps:
//	1. Pre-processing: transaction parameters are initialized, retry operations
//     are filtered from the obsolete ones and for the resync the graph is refreshed
//  2. Simulation: simulating transaction without actually executing any of the
//     Create/Delete/Update operations in order to obtain the "execution plan"
//  3. Pre-recording: logging transaction arguments + plan before execution to
//     persist some information in case there is a crash during execution
//  4. Execution: executing the transaction, collecting errors
//  5. Recording: recording the finalized transaction (log + in-memory)
//  6. Post-processing: scheduling retry for failed operations, propagating value
//     state updates to the subscribers and returning error/nil to the caller
//     of blocking commit
func (s *Scheduler) processTransaction(txn *transaction) {
	s.txnLock.Lock()
	defer s.txnLock.Unlock()

	startTime := time.Now()

	// 1. Pre-processing:
	skipExec, skipSimulation, record := s.preProcessTransaction(txn)

	// 2. Ordering:
	if !skipExec {
		txn.values = s.orderValuesByOp(txn.values)
	}

	// 3. Simulation:
	var simulatedOps kvs.RecordedTxnOps
	if !skipSimulation {
		graphW := s.graph.Write(false, record)
		simulatedOps = s.executeTransaction(txn, graphW, true)
		if len(simulatedOps) == 0 {
			// nothing to execute
			graphW.Save()
			skipExec = true
		}
		graphW.Release()
	}

	// 4. Pre-recording
	preTxnRecord := s.preRecordTransaction(txn, simulatedOps, skipSimulation)

	// 5. Execution:
	var executedOps kvs.RecordedTxnOps
	if !skipExec {
		graphW := s.graph.Write(true, record)
		executedOps = s.executeTransaction(txn, graphW, false)
		graphW.Release()
	}

	stopTime := time.Now()

	// 6. Recording:
	s.recordTransaction(txn, preTxnRecord, executedOps, startTime, stopTime)

	// 7. Post-processing:
	s.postProcessTransaction(txn, executedOps)
}

// preProcessTransaction initializes transaction parameters, filters obsolete retry
// operations and refreshes the graph for resync.
func (s *Scheduler) preProcessTransaction(txn *transaction) (skipExec, skipSimulation, record bool) {
	defer trace.StartRegion(txn.ctx, "preProcessTransaction").End()

	// allocate new transaction sequence number
	txn.seqNum = s.txnSeqNumber
	s.txnSeqNumber++

	switch txn.txnType {
	case kvs.SBNotification:
		skipExec = s.preProcessNotification(txn)
		skipSimulation = !s.config.EnableTxnSimulation
		record = true
	case kvs.NBTransaction:
		skipExec = s.preProcessNBTransaction(txn)
		skipSimulation = skipExec || !txn.nb.withSimulation
		record = txn.nb.resyncType != kvs.DownstreamResync
	case kvs.RetryFailedOps:
		skipExec = s.preProcessRetryTxn(txn)
		skipSimulation = skipExec
		record = true
	}

	return
}

// preProcessNotification filters out non-valid SB notification.
func (s *Scheduler) preProcessNotification(txn *transaction) (skip bool) {
	graphR := s.graph.Read()
	defer graphR.Release()

	kv := txn.values[0]
	skip = s.filterNotification(graphR, kv.key, kv.value, txn.seqNum)
	return
}

// preProcessNBTransaction refreshes the graph for resync.
func (s *Scheduler) preProcessNBTransaction(txn *transaction) (skip bool) {
	if txn.nb.resyncType == kvs.NotResync {
		// nothing to do in the pre-processing stage
		return false
	}

	// for resync refresh the graph + collect deletes
	graphW := s.graph.Write(true,false)
	defer graphW.Release()
	s.resyncCount++

	if txn.nb.resyncType == kvs.DownstreamResync {
		// for downstream resync it is assumed that scheduler is in-sync with NB
		currentNodes := graphW.GetNodes(nil, nbBaseValsSelectors()...)
		for _, node := range currentNodes {
			lastUpdate := getNodeLastUpdate(node)
			txn.values = append(txn.values,
				kvForTxn{
					key:      node.GetKey(),
					value:    lastUpdate.value,
					origin:   kvs.FromNB,
					isRevert: lastUpdate.revert,
				})
		}
	}

	// build the set of keys currently in NB
	nbKeys := utils.NewMapBasedKeySet()
	for _, kv := range txn.values {
		nbKeys.Add(kv.key)
	}

	// unless this is only UpstreamResync, refresh the graph with the current
	// state of SB
	if txn.nb.resyncType != kvs.UpstreamResync {
		s.refreshGraph(graphW, nil, &resyncData{
			first:  s.resyncCount == 1,
			values: txn.values,
		}, txn.nb.verboseRefresh)
	}

	// collect deletes for obsolete values
	currentNodes := graphW.GetNodes(nil, nbBaseValsSelectors()...)
	for _, node := range currentNodes {
		if nbKey := nbKeys.Has(node.GetKey()); nbKey {
			continue
		}
		txn.values = append(txn.values,
			kvForTxn{
				key:    node.GetKey(),
				value:  nil, // remove
				origin: kvs.FromNB,
			})
	}

	// update (record) SB values
	sbNodes := graphW.GetNodes(nil, sbBaseValsSelectors()...)
	for _, node := range sbNodes {
		if nbKey := nbKeys.Has(node.GetKey()); nbKey {
			continue
		}
		txn.values = append(txn.values,
			kvForTxn{
				key:    node.GetKey(),
				value:  node.GetValue(),
				origin: kvs.FromSB,
			})
	}

	skip = len(txn.values) == 0
	return
}

// preProcessRetryTxn filters out obsolete retry operations.
func (s *Scheduler) preProcessRetryTxn(txn *transaction) (skip bool) {
	graphR := s.graph.Read()
	defer graphR.Release()

	for key, retryRev := range txn.retry.keys {
		node := graphR.GetNode(key)
		if node == nil {
			continue
		}
		lastUpdate := getNodeLastUpdate(node)
		if lastUpdate == nil || lastUpdate.txnSeqNum > retryRev {
			// obsolete retry, the value has been updated since the failure
			continue
		}
		txn.values = append(txn.values,
			kvForTxn{
				key:      key,
				value:    lastUpdate.value,
				origin:   kvs.FromNB,
				isRevert: lastUpdate.revert,
			})
	}
	skip = len(txn.values) == 0
	return
}

// postProcessTransaction schedules retry for failed operations and propagates
// value state updates to the subscribers and error/nil to the caller of a blocking
// commit.
func (s *Scheduler) postProcessTransaction(txn *transaction, executed kvs.RecordedTxnOps) {
	defer trace.StartRegion(txn.ctx, "postProcessTransaction").End()

	// collect new failures (combining derived with base)
	toRetry := utils.NewSliceBasedKeySet()
	toRefresh := utils.NewSliceBasedKeySet()

	var verboseRefresh bool
	graphR := s.graph.Read()
	for _, op := range executed {
		node := graphR.GetNode(op.Key)
		if node == nil {
			continue
		}
		state := getNodeState(node)
		baseKey := getNodeBaseKey(node)
		if state == kvs.ValueState_UNIMPLEMENTED {
			continue
		}
		if state == kvs.ValueState_FAILED {
			toRefresh.Add(baseKey)
			verboseRefresh = true
		}
		if state == kvs.ValueState_RETRYING {
			toRefresh.Add(baseKey)
			toRetry.Add(baseKey)
			verboseRefresh = true
		}
		if s.verifyMode {
			toRefresh.Add(baseKey)
		}
	}
	graphR.Release()

	// refresh base values which themselves are in a failed state or have derived failed values
	// - in verifyMode all updated values are re-freshed
	if toRefresh.Length() > 0 {
		graphW := s.graph.Write(true,false)
		s.refreshGraph(graphW, toRefresh, nil, verboseRefresh)

		// split values based on the retry metadata
		retryTxns := make(map[retryTxnMeta]*retryTxn)
		for _, retryKey := range toRetry.Iterate() {
			node := graphW.GetNode(retryKey)
			lastUpdate := getNodeLastUpdate(node)
			// did retry fail?
			var alreadyRetried bool
			if txn.txnType == kvs.RetryFailedOps {
				_, alreadyRetried = txn.retry.keys[retryKey]
			}
			// determine how long to delay the retry
			delay := lastUpdate.retryArgs.Period
			if alreadyRetried && lastUpdate.retryArgs.ExpBackoff {
				delay = txn.retry.delay * 2
			}
			// determine which attempt this is
			attempt := 1
			if alreadyRetried {
				attempt = txn.retry.attempt + 1
			}
			// determine which transaction this retry is for
			seqNum := txn.seqNum
			if alreadyRetried {
				seqNum = txn.retry.txnSeqNum
			}
			// add key into the set to retry within a single transaction
			retryMeta := retryTxnMeta{
				txnSeqNum: seqNum,
				delay:     delay,
				attempt:   attempt,
			}
			if _, has := retryTxns[retryMeta]; !has {
				retryTxns[retryMeta] = &retryTxn{
					retryTxnMeta: retryMeta,
					keys:         make(map[string]uint64),
				}
			}
			retryTxns[retryMeta].keys[retryKey] = lastUpdate.txnSeqNum
		}

		// schedule a series of re-try transactions for failed values
		for _, retryTxn := range retryTxns {
			s.enqueueRetry(retryTxn)
		}
		graphW.Release()
	}

	// collect state updates
	var stateUpdates []*kvs.BaseValueStatus
	removed := utils.NewSliceBasedKeySet()
	graphR = s.graph.Read()
	for _, key := range s.updatedStates.Iterate() {
		node := graphR.GetNode(key)
		status := getValueStatus(node, key)
		if status.Value.State == kvs.ValueState_REMOVED {
			removed.Add(key)
		}
		stateUpdates = append(stateUpdates, status)
	}
	graphR.Release()
	// clear the set of updated states
	s.updatedStates = utils.NewSliceBasedKeySet()

	// if enabled, verify transaction effects
	var kvErrors []kvs.KeyWithError
	if s.verifyMode {
		graphR = s.graph.Read()
		for _, op := range executed {
			key := op.Key
			node := graphR.GetNode(key)
			if node == nil {
				continue
			}
			state := getNodeState(node)
			if state == kvs.ValueState_RETRYING || state == kvs.ValueState_FAILED {
				// effects of failed operations are uncertain and cannot be therefore verified
				continue
			}
			expValue := getNodeLastAppliedValue(node)
			lastOp := getNodeLastOperation(node)
			expToNotExist := expValue == nil || state == kvs.ValueState_PENDING || state == kvs.ValueState_INVALID
			if expToNotExist && isNodeAvailable(node) {
				kvErrors = append(kvErrors, kvs.KeyWithError{
					Key:          key,
					Error:        kvs.NewVerificationError(key, kvs.ExpectedToNotExist),
					TxnOperation: lastOp,
				})
				continue
			}
			if expValue == nil {
				// properly removed
				continue
			}
			if !expToNotExist && !isNodeAvailable(node) {
				kvErrors = append(kvErrors, kvs.KeyWithError{
					Key:          key,
					Error:        kvs.NewVerificationError(key, kvs.ExpectedToExist),
					TxnOperation: lastOp,
				})
				continue
			}
			descriptor := s.registry.GetDescriptorForKey(key)
			handler := &descriptorHandler{descriptor}
			equivalent := handler.equivalentValues(key, node.GetValue(), expValue)
			if !equivalent {
				kvErrors = append(kvErrors, kvs.KeyWithError{
					Key:          key,
					Error:        kvs.NewVerificationError(key, kvs.NotEquivalent),
					TxnOperation: lastOp,
				})
				s.Log.WithFields(
					logging.Fields{
						"applied":   expValue,
						"refreshed": node.GetValue(),
					}).Warn("Detected non-equivalent applied vs. refreshed values")
			}
		}
		graphR.Release()
	}

	// build transaction error
	var txnErr error
	for _, txnOp := range executed {
		if txnOp.NewErr == nil {
			continue
		}
		kvErrors = append(kvErrors,
			kvs.KeyWithError{
				Key:          txnOp.Key,
				TxnOperation: txnOp.Operation,
				Error:        txnOp.NewErr,
			})
	}
	if len(kvErrors) > 0 {
		txnErr = kvs.NewTransactionError(nil, kvErrors)
	}
	if txn.txnType == kvs.NBTransaction && txn.nb.isBlocking {
		// for blocking txn, send non-nil errors to the resultChan
		select {
		case txn.nb.resultChan <- txnResult{txnSeqNum: txn.seqNum, err: txnErr}:
		default:
			s.Log.WithField("txnSeq", txn.seqNum).
				Warn("Failed to deliver transaction result to the caller")
		}
	} else {
		// for asynchronous events, just log the transaction error
		if txnErr == nil {
			s.Log.Infof("Transaction %d successful!", txn.seqNum)
		} else {
			s.Log.Error(txnErr.Error())
		}
	}

	// send value status updates to the watchers
	for _, watcher := range s.valStateWatchers {
		for _, stateUpdate := range stateUpdates {
			if watcher.selector == nil || watcher.selector(stateUpdate.Value.Key) {
				select {
				case watcher.channel <- stateUpdate:
				default:
					s.Log.WithField("txnSeq", txn.seqNum).
						Warn("Failed to deliver value status update to a watcher")
				}
			}
		}
	}

	// delete removed values from the graph after the notifications have been sent
	if removed.Length() > 0 {
		graphW := s.graph.Write(true,true)
		for _, key := range removed.Iterate() {
			graphW.DeleteNode(key)
		}
		graphW.Release()
	}
}

// filterNotification checks if the received notification should be filtered
// or normally applied.
func (s *Scheduler) filterNotification(graphR graph.ReadAccess, key string, value proto.Message, txnSeqNum uint64) bool {
	descriptor := s.registry.GetDescriptorForKey(key)
	if descriptor == nil {
		s.Log.WithFields(logging.Fields{
			"txnSeqNum": txnSeqNum,
			"key":       key,
		}).Debug("Ignoring unimplemented notification")
		return true
	}
	node := graphR.GetNode(key)
	if node != nil {
		if getNodeOrigin(node) == kvs.FromNB {
			s.Log.WithFields(logging.Fields{
				"txnSeqNum": txnSeqNum,
				"key":       key,
			}).Debug("Ignoring notification for a NB-managed value")
			return true
		}
	}
	return false
}
