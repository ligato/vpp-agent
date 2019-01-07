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
	"reflect"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/ligato/cn-infra/logging"

	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// preProcessedTxn appends un-marshalled (or filtered retry) values to a queued
// transaction and sets the sequence number.
type preProcessedTxn struct {
	seqNum uint64
	values []kvForTxn
	args   *queuedTxn
}

// kvForTxn represents a new value for a given key to be applied in a transaction.
type kvForTxn struct {
	key      string
	value    proto.Message
	metadata kvs.Metadata
	origin   kvs.ValueOrigin
	isRevert bool
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
	}
}

// processTransaction processes transaction in 6 steps:
//	1. Pre-processing: transaction parameters are initialized, retry operations
//     are filtered from the obsolete ones and for the resync the graph is refreshed
//  2. Simulation (skipped for SB notification): simulating transaction without
//     actually executing any of the Add/Delete/Modify/Update operations in order
//     to obtain the "execution plan"
//  3. Pre-recording: logging transaction arguments + plan before execution to
//     persist some information in case there is a crash during execution
//  4. Execution: executing the transaction, collecting errors
//  5. Recording: recording the finalized transaction (log + in-memory)
//  6. Post-processing: scheduling retry for failed operations, propagating errors
//     to the subscribers and to the caller of blocking commit
func (s *Scheduler) processTransaction(qTxn *queuedTxn) {
	var (
		simulatedOps kvs.RecordedTxnOps
		executedOps  kvs.RecordedTxnOps
		failed       map[string]bool
		startTime    time.Time
		stopTime     time.Time
	)
	s.txnLock.Lock()
	defer s.txnLock.Unlock()

	// 1. Pre-processing:
	startTime = time.Now()
	txn, preErrors := s.preProcessTransaction(qTxn)
	eligibleForExec := len(txn.values) > 0 && len(preErrors) == 0

	// 2. Ordering:
	txn.values = s.orderValuesByOp(txn.values)

	// 3. Simulation:
	if eligibleForExec {
		simulatedOps, _ = s.executeTransaction(txn, true)
	}

	// 4. Pre-recording
	preTxnRecord := s.preRecordTransaction(txn, simulatedOps, preErrors)

	// 5. Execution:
	if eligibleForExec {
		executedOps, failed = s.executeTransaction(txn, false)
	}
	stopTime = time.Now()

	// 6. Recording:
	s.recordTransaction(preTxnRecord, executedOps, startTime, stopTime)

	// 7. Post-processing:
	s.postProcessTransaction(txn, executedOps, failed, preErrors)
}

// preProcessTransaction initializes transaction parameters, filters obsolete retry
// operations and refreshes the graph for resync.
func (s *Scheduler) preProcessTransaction(qTxn *queuedTxn) (txn *preProcessedTxn, errors []kvs.KeyWithError) {
	// allocate new transaction sequence number
	preTxn := &preProcessedTxn{seqNum: s.txnSeqNumber, args: qTxn}
	s.txnSeqNumber++

	switch qTxn.txnType {
	case kvs.SBNotification:
		s.preProcessNotification(qTxn, preTxn)
	case kvs.NBTransaction:
		errors = s.preProcessNBTransaction(qTxn, preTxn)
	case kvs.RetryFailedOps:
		s.preProcessRetryTxn(qTxn, preTxn)
	}

	return preTxn, errors
}

// preProcessNotification filters out non-valid SB notification.
func (s *Scheduler) preProcessNotification(qTxn *queuedTxn, preTxn *preProcessedTxn) {
	graphR := s.graph.Read()
	defer graphR.Release()

	if !s.validTxnValue(graphR, qTxn.sb.value.Key, qTxn.sb.value.Value, kvs.FromSB, preTxn.seqNum) {
		return
	}
	preTxn.values = append(preTxn.values,
		kvForTxn{
			key:      qTxn.sb.value.Key,
			value:    qTxn.sb.value.Value,
			metadata: qTxn.sb.metadata,
			origin:   kvs.FromSB,
		})
}

// preProcessNBTransaction unmarshalls transaction values and for resync also refreshes the graph.
func (s *Scheduler) preProcessNBTransaction(qTxn *queuedTxn, preTxn *preProcessedTxn) (errors []kvs.KeyWithError) {
	// unmarshall all values
	graphR := s.graph.Read()
	for key, lazyValue := range qTxn.nb.value {
		descriptor := s.registry.GetDescriptorForKey(key)
		if descriptor == nil {
			// unimplemented base value
			errors = append(errors, kvs.KeyWithError{Key: key, TxnOperation: kvs.PreProcess, Error: kvs.ErrUnimplementedKey})
			continue
		}
		var value proto.Message
		if lazyValue != nil {
			// create an instance of the target proto.Message type
			valueType := proto.MessageType(descriptor.ValueTypeName)
			if valueType == nil {
				errors = append(errors, kvs.KeyWithError{Key: key, TxnOperation: kvs.PreProcess, Error: kvs.ErrUnregisteredValueType})
				continue
			}
			value = reflect.New(valueType.Elem()).Interface().(proto.Message)
			// try to deserialize the value
			err := lazyValue.GetValue(value)
			if err != nil {
				errors = append(errors, kvs.KeyWithError{Key: key, TxnOperation: kvs.PreProcess, Error: err})
				continue
			}
		}
		if !s.validTxnValue(graphR, key, value, kvs.FromNB, preTxn.seqNum) {
			continue
		}
		preTxn.values = append(preTxn.values,
			kvForTxn{
				key:    key,
				value:  value,
				origin: kvs.FromNB,
			})
	}
	graphR.Release()

	// for resync refresh the graph + collect deletes
	if len(errors) == 0 && qTxn.nb.resyncType != kvs.NotResync {
		graphW := s.graph.Write(false)
		defer graphW.Release()
		defer graphW.Save()
		s.resyncCount++

		if qTxn.nb.resyncType == kvs.DownstreamResync {
			// for downstream resync it is assumed that scheduler is in-sync with NB
			currentNodes := graphW.GetNodes(nil,
				graph.WithFlags(&OriginFlag{kvs.FromNB}),
				graph.WithoutFlags(&DerivedFlag{}))
			for _, node := range currentNodes {
				lastChange := getNodeLastChange(node)
				preTxn.values = append(preTxn.values,
					kvForTxn{
						key:      node.GetKey(),
						value:    lastChange.value,
						origin:   kvs.FromNB,
						isRevert: lastChange.revert,
					})
			}
		}

		// build the set of keys currently in NB
		nbKeys := utils.NewMapBasedKeySet()
		for _, kv := range preTxn.values {
			nbKeys.Add(kv.key)
		}

		// unless this is only UpstreamResync, refresh the graph with the current
		// state of SB
		if qTxn.nb.resyncType != kvs.UpstreamResync {
			s.refreshGraph(graphW, nil, &resyncData{
				first:   s.resyncCount == 1,
				values:  preTxn.values,
				verbose: qTxn.nb.verboseRefresh})
		}

		// collect deletes for obsolete values
		currentNodes := graphW.GetNodes(nil,
			graph.WithFlags(&OriginFlag{kvs.FromNB}),
			graph.WithoutFlags(&DerivedFlag{}))
		for _, node := range currentNodes {
			if nbKey := nbKeys.Has(node.GetKey()); nbKey {
				continue
			}
			preTxn.values = append(preTxn.values,
				kvForTxn{
					key:    node.GetKey(),
					value:  nil, // remove
					origin: kvs.FromNB,
				})
		}

		// update (record) SB values
		sbNodes := graphW.GetNodes(nil,
			graph.WithFlags(&OriginFlag{kvs.FromSB}),
			graph.WithoutFlags(&DerivedFlag{}))
		for _, node := range sbNodes {
			if nbKey := nbKeys.Has(node.GetKey()); nbKey {
				continue
			}
			preTxn.values = append(preTxn.values,
				kvForTxn{
					key:    node.GetKey(),
					value:  node.GetValue(),
					origin: kvs.FromSB,
				})
		}
	}

	return errors
}

// preProcessRetryTxn filters out obsolete retry operations.
func (s *Scheduler) preProcessRetryTxn(qTxn *queuedTxn, preTxn *preProcessedTxn) {
	graphR := s.graph.Read()
	defer graphR.Release()

	for _, key := range qTxn.retry.keys.Iterate() {
		node := graphR.GetNode(key)
		if node == nil {
			continue
		}
		lastChange := getNodeLastChange(node)
		if lastChange.txnSeqNum > qTxn.retry.txnSeqNum {
			// obsolete retry, the value has been changed since the failure
			continue
		}
		preTxn.values = append(preTxn.values,
			kvForTxn{
				key:      key,
				value:    lastChange.value,
				origin:   lastChange.origin, // FromNB
				isRevert: lastChange.revert,
			})
	}
}

// postProcessTransaction schedules retry for failed operations and propagates
// errors to the subscribers and to the caller of a blocking commit.
func (s *Scheduler) postProcessTransaction(txn *preProcessedTxn, executed kvs.RecordedTxnOps, failed map[string]bool, preErrors []kvs.KeyWithError) {
	// refresh base values with error or with a derived value that has an error
	if len(failed) > 0 {
		graphW := s.graph.Write(false)
		toRefresh := utils.NewMapBasedKeySet()
		for key := range failed {
			toRefresh.Add(key)
		}
		s.refreshGraph(graphW, toRefresh, nil)
		graphW.Save()

		// split failed values based on transactions that performed the last change
		retryTxns := make(map[uint64]*retryOps)
		for retryKey, retriable := range failed {
			if !retriable {
				continue
			}
			node := graphW.GetNode(retryKey)
			if node == nil {
				// delete returned error, but refresh showed that it is not in SB anymore anyway
				continue
			}
			lastChange := getNodeLastChange(node)
			seqNum := lastChange.txnSeqNum
			if lastChange.retryEnabled {
				if _, has := retryTxns[seqNum]; !has {
					period := lastChange.retryPeriod
					if seqNum == txn.seqNum && txn.args.txnType == kvs.RetryFailedOps && lastChange.retryExpBackoff {
						period = txn.args.retry.period * 2
					}
					retryTxns[seqNum] = &retryOps{
						txnSeqNum: seqNum,
						period:    period,
						keys:      utils.NewMapBasedKeySet(),
					}
				}
				retryTxns[seqNum].keys.Add(retryKey)
			}
		}

		// schedule a series of re-try transactions for failed values
		for _, retryTxn := range retryTxns {
			s.enqueueRetry(retryTxn)
		}
		graphW.Release()
	}

	// collect errors
	var txnErrors []kvs.KeyWithError
	txnErrors = append(txnErrors, preErrors...)

	for _, txnOp := range executed {
		if txnOp.PrevErr == nil && txnOp.NewErr == nil {
			continue
		}
		txnErrors = append(txnErrors,
			kvs.KeyWithError{
				Key:          txnOp.Key,
				TxnOperation: txnOp.Operation,
				Error:        txnOp.NewErr,
			})
	}

	// for blocking txn, send non-nil errors to the resultChan
	if txn.args.txnType == kvs.NBTransaction && txn.args.nb.isBlocking {
		var (
			errors []kvs.KeyWithError
			txnErr error
		)
		for _, kvWithError := range txnErrors {
			if kvWithError.Error != nil {
				errors = append(errors, kvWithError)
			}
		}
		if len(errors) > 0 {
			txnErr = kvs.NewTransactionError(nil, errors)
		}

		select {
		case txn.args.nb.resultChan <- txnResult{txnSeqNum: txn.seqNum, err: txnErr}:
		default:
			s.Log.WithField("txnSeq", txn.seqNum).
				Warn("Failed to deliver transaction result to the caller")
		}
	}

	// send errors to the subscribers
	for _, errSub := range s.errorSubs {
		for _, kvWithError := range txnErrors {
			if errSub.selector == nil || errSub.selector(kvWithError.Key) {
				select {
				case errSub.channel <- kvWithError:
				default:
					s.Log.WithField("txnSeq", txn.seqNum).
						Warn("Failed to deliver transaction error to a subscriber")
				}
			}
		}
	}
}

// validTxnValue checks validity of a kv-pair to be applied in a transaction.
func (s *Scheduler) validTxnValue(graphR graph.ReadAccess, key string, value proto.Message, origin kvs.ValueOrigin, txnSeqNum uint64) bool {
	if key == "" {
		s.Log.WithFields(logging.Fields{
			"txnSeqNum": txnSeqNum,
		}).Warn("Empty key for a value in the transaction")
		return false
	}
	if origin == kvs.FromSB {
		descriptor := s.registry.GetDescriptorForKey(key)
		if descriptor == nil {
			s.Log.WithFields(logging.Fields{
				"txnSeqNum": txnSeqNum,
				"key":       key,
			}).Debug("Ignoring unimplemented notification")
			return false
		}
	}
	node := graphR.GetNode(key)
	if node != nil {
		if isNodeDerived(node) {
			s.Log.WithFields(logging.Fields{
				"txnSeqNum": txnSeqNum,
				"key":       key,
			}).Warn("Transaction attempting to change a derived value")
			return false
		}
		if origin == kvs.FromSB && getNodeOrigin(node) == kvs.FromNB {
			s.Log.WithFields(logging.Fields{
				"txnSeqNum": txnSeqNum,
				"key":       key,
			}).Debug("Ignoring notification for a NB-managed value")
			return false
		}
	}
	return true
}
