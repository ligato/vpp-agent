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

	. "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// preProcessedTxn appends un-marshalled (or filtered retry) values to a queued
// transaction and sets the sequence number.
type preProcessedTxn struct {
	seqNum uint
	values []kvForTxn
	args   *queuedTxn
}

// kvForTxn represents a new value for a given key to be applied in a transaction.
type kvForTxn struct {
	key      string
	value    proto.Message
	metadata Metadata
	origin   ValueOrigin
	isRevert bool
}

// consumeTransactions pulls the oldest queued transaction and starts the processing.
func (scheduler *Scheduler) consumeTransactions() {
	defer scheduler.wg.Done()
	for {
		txn, canceled := scheduler.dequeueTxn()
		if canceled {
			return
		}
		scheduler.processTransaction(txn)
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
func (scheduler *Scheduler) processTransaction(qTxn *queuedTxn) {
	var (
		simulatedOps recordedTxnOps
		executedOps  recordedTxnOps
		failed       map[string]bool
		execStart    time.Time
		execStop     time.Time
	)
	scheduler.txnLock.Lock()
	defer scheduler.txnLock.Unlock()

	// 1. Pre-processing:
	txn, preErrors := scheduler.preProcessTransaction(qTxn)
	eligibleForExec := len(txn.values) > 0 && len(preErrors) == 0

	// 2. Simulation:
	if eligibleForExec {
		simulatedOps, _ = scheduler.executeTransaction(txn, true)
	}

	// 3. Pre-recording
	preTxnRecord := scheduler.preRecordTransaction(txn, simulatedOps, preErrors)

	// 4. Execution:
	execStart = time.Now()
	if eligibleForExec {
		executedOps, failed = scheduler.executeTransaction(txn, false)
	}
	execStop = time.Now()

	// 5. Recording:
	scheduler.recordTransaction(preTxnRecord, executedOps, execStart, execStop)

	// 6. Post-processing:
	scheduler.postProcessTransaction(txn, executedOps, failed, preErrors)
}

// preProcessTransaction initializes transaction parameters, filters obsolete retry
// operations and refreshes the graph for resync.
func (scheduler *Scheduler) preProcessTransaction(qTxn *queuedTxn) (txn *preProcessedTxn, errors []KeyWithError) {
	// allocate new transaction sequence number
	preTxn := &preProcessedTxn{seqNum: scheduler.txnSeqNumber, args: qTxn}
	scheduler.txnSeqNumber++

	switch qTxn.txnType {
	case sbNotification:
		scheduler.preProcessNotification(qTxn, preTxn)
	case nbTransaction:
		errors = scheduler.preProcessNBTransaction(qTxn, preTxn)
	case retryFailedOps:
		scheduler.preProcessRetryTxn(qTxn, preTxn)
	}

	return preTxn, errors
}

// preProcessNotification filters out non-valid SB notification.
func (scheduler *Scheduler) preProcessNotification(qTxn *queuedTxn, preTxn *preProcessedTxn) {
	graphR := scheduler.graph.Read()
	defer graphR.Release()

	if !scheduler.validTxnValue(graphR, qTxn.sb.value.Key, qTxn.sb.value.Value, FromSB, preTxn.seqNum) {
		return
	}
	preTxn.values = append(preTxn.values,
		kvForTxn{
			key:      qTxn.sb.value.Key,
			value:    qTxn.sb.value.Value,
			metadata: qTxn.sb.metadata,
			origin:   FromSB,
		})
}

// preProcessNBTransaction unmarshalls transaction values and for resync also refreshes the graph.
func (scheduler *Scheduler) preProcessNBTransaction(qTxn *queuedTxn, preTxn *preProcessedTxn) (errors []KeyWithError) {
	// unmarshall all values
	graphR := scheduler.graph.Read()
	for key, lazyValue := range qTxn.nb.value {
		descriptor := scheduler.registry.GetDescriptorForKey(key)
		if descriptor == nil {
			// unimplemented base value
			errors = append(errors, KeyWithError{Key: key, TxnOperation: PreProcess, Error: ErrUnimplementedKey})
			continue
		}
		var value proto.Message
		if lazyValue != nil {
			// create an instance of the target proto.Message type
			valueType := proto.MessageType(descriptor.ValueTypeName)
			if valueType == nil {
				errors = append(errors, KeyWithError{Key: key, TxnOperation: PreProcess, Error: ErrUnregisteredValueType})
				continue
			}
			value = reflect.New(valueType.Elem()).Interface().(proto.Message)
			// try to deserialize the value
			err := lazyValue.GetValue(value)
			if err != nil {
				errors = append(errors, KeyWithError{Key: key, TxnOperation: PreProcess, Error: err})
				continue
			}
		}
		if !scheduler.validTxnValue(graphR, key, value, FromNB, preTxn.seqNum) {
			continue
		}
		preTxn.values = append(preTxn.values,
			kvForTxn{
				key:    key,
				value:  value,
				origin: FromNB,
			})
	}
	graphR.Release()

	// for resync refresh the graph + collect deletes
	if len(errors) == 0 && (qTxn.nb.isFullResync || qTxn.nb.isDownstreamResync) {
		graphW := scheduler.graph.Write(false)
		defer graphW.Release()
		defer graphW.Save()
		scheduler.resyncCount++

		if qTxn.nb.isDownstreamResync {
			// for downstream resync it is assumed that scheduler is in-sync with NB
			currentNodes := graphW.GetNodes(nil,
				graph.WithFlags(&OriginFlag{FromNB}),
				graph.WithoutFlags(&DerivedFlag{}))
			for _, node := range currentNodes {
				lastChange := getNodeLastChange(node)
				preTxn.values = append(preTxn.values,
					kvForTxn{
						key:      node.GetKey(),
						value:    lastChange.value,
						origin:   FromNB,
						isRevert: lastChange.revert,
					})
			}
		}

		// build the set of keys currently in NB
		nbKeys := utils.NewKeySet()
		for _, kv := range preTxn.values {
			nbKeys.Add(kv.key)
		}

		// refresh the graph with the current state of SB
		scheduler.refreshGraph(graphW, nil,
			&resyncData{first: scheduler.resyncCount == 1, values: preTxn.values})
		currentNodes := graphW.GetNodes(nil,
			graph.WithFlags(&OriginFlag{FromNB}),
			graph.WithoutFlags(&DerivedFlag{}))

		// collect deletes for obsolete values
		for _, node := range currentNodes {
			if _, nbKey := nbKeys[node.GetKey()]; nbKey {
				continue
			}
			preTxn.values = append(preTxn.values,
				kvForTxn{
					key:    node.GetKey(),
					value:  nil, // remove
					origin: FromNB,
				})
		}

		// update (record) SB values
		sbNodes := graphW.GetNodes(nil,
			graph.WithFlags(&OriginFlag{FromSB}),
			graph.WithoutFlags(&DerivedFlag{}))
		for _, node := range sbNodes {
			if _, nbKey := nbKeys[node.GetKey()]; nbKey {
				continue
			}
			preTxn.values = append(preTxn.values,
				kvForTxn{
					key:    node.GetKey(),
					value:  node.GetValue(),
					origin: FromSB,
				})
		}
	}

	return errors
}

// preProcessRetryTxn filters out obsolete retry operations.
func (scheduler *Scheduler) preProcessRetryTxn(qTxn *queuedTxn, preTxn *preProcessedTxn) {
	graphR := scheduler.graph.Read()
	defer graphR.Release()

	for key := range qTxn.retry.keys {
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
func (scheduler *Scheduler) postProcessTransaction(txn *preProcessedTxn, executed recordedTxnOps, failed map[string]bool, preErrors []KeyWithError) {
	// refresh base values with error or with a derived value that has an error
	if len(failed) > 0 {
		graphW := scheduler.graph.Write(false)
		toRefresh := utils.NewKeySet()
		for key := range failed {
			toRefresh.Add(key)
		}
		scheduler.refreshGraph(graphW, toRefresh, nil)
		graphW.Save()

		// split failed values based on transactions that performed the last change
		retryTxns := make(map[uint]*retryOps)
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
					if seqNum == txn.seqNum && txn.args.txnType == retryFailedOps && lastChange.retryExpBackoff {
						period = txn.args.retry.period * 2
					}
					retryTxns[seqNum] = &retryOps{
						txnSeqNum: seqNum,
						period:    period,
						keys:      utils.NewKeySet(),
					}
				}
				retryTxns[seqNum].keys.Add(retryKey)
			}
		}

		// schedule a series of re-try transactions for failed values
		for _, retryTxn := range retryTxns {
			scheduler.enqueueRetry(retryTxn)
		}
		graphW.Release()
	}

	// collect errors
	var txnErrors []KeyWithError
	for _, preError := range preErrors {
		txnErrors = append(txnErrors, preError)
	}
	for _, txnOp := range executed {
		if txnOp.prevErr == nil && txnOp.newErr == nil {
			continue
		}
		txnErrors = append(txnErrors,
			KeyWithError{
				Key:          txnOp.key,
				TxnOperation: txnOp.operation,
				Error:        txnOp.newErr,
			})
	}

	// for blocking txn, send non-nil errors to the resultChan
	if txn.args.txnType == nbTransaction && txn.args.nb.isBlocking {
		var errors []KeyWithError
		for _, kvWithError := range txnErrors {
			if kvWithError.Error != nil {
				errors = append(errors, kvWithError)
			}
		}
		select {
		case txn.args.nb.resultChan <- errors:
		default:
			scheduler.Log.WithField("txnSeq", txn.seqNum).
				Warn("Failed to deliver transaction result to the caller")
		}
	}

	// send errors to the subscribers
	for _, errSub := range scheduler.errorSubs {
		for _, kvWithError := range txnErrors {
			if errSub.selector == nil || errSub.selector(kvWithError.Key) {
				select {
				case errSub.channel <- kvWithError:
				default:
					scheduler.Log.WithField("txnSeq", txn.seqNum).
						Warn("Failed to deliver transaction error to a subscriber")
				}
			}
		}
	}
}

// validTxnValue checks validity of a kv-pair to be applied in a transaction.
func (scheduler *Scheduler) validTxnValue(graphR graph.ReadAccess, key string, value proto.Message, origin ValueOrigin, txnSeqNum uint) bool {
	if key == "" {
		scheduler.Log.WithFields(logging.Fields{
			"txnSeqNum": txnSeqNum,
		}).Warn("Empty key for a value in the transaction")
		return false
	}
	if origin == FromSB {
		descriptor := scheduler.registry.GetDescriptorForKey(key)
		if descriptor == nil {
			scheduler.Log.WithFields(logging.Fields{
				"txnSeqNum": txnSeqNum,
				"key":       key,
			}).Debug("Ignoring unimplemented notification")
			return false
		}
	}
	node := graphR.GetNode(key)
	if node != nil {
		if isNodeDerived(node) {
			scheduler.Log.WithFields(logging.Fields{
				"txnSeqNum": txnSeqNum,
				"key":       key,
			}).Warn("Transaction attempting to change a derived value")
			return false
		}
		if origin == FromSB && getNodeOrigin(node) == FromNB {
			scheduler.Log.WithFields(logging.Fields{
				"txnSeqNum": txnSeqNum,
				"key":       key,
			}).Debug("Ignoring notification for a NB-managed value")
			return false
		}
	}
	return true
}
