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
	. "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/logging"
	"time"
)

// txnType differentiates between NB transaction, retry of failed operations and
// SB notification. Once queued, all three different operations are classified
// as transactions, only with different parameters.
type txnType int

const (
	sbNotification txnType = iota
	nbTransaction
	retryFailedOps
)

// String returns human-readable string representation of the transaction type.
func (txnType txnType) String() string {
	switch txnType {
	case sbNotification:
		return "SB notification"
	case nbTransaction:
		return "NB transaction"
	case retryFailedOps:
		return "RETRY"
	}
	return "UNKNOWN"
}

// sbNotif encapsulates data for SB notification.
type sbNotif struct {
	value    KeyValuePair
	metadata Metadata
}

// nbTxn encapsulates data for NB transaction.
type nbTxn struct {
	valueData       map[string]interface{} // key -> value data
	isResync        bool
	isBlocking      bool
	retryFailed     bool
	retryPeriod     time.Duration
	expBackoffRetry bool
	revertOnFailure bool
	resultChan      chan []KeyWithError
}

// retryOps encapsulates data for retry of failed operations.
type retryOps struct {
	txnSeqNum  uint
	keys       map[string]struct{}
	period     time.Duration
	expBackoff bool
}

// queuedTxn represents transaction queued for execution.
type queuedTxn struct {
	txnType txnType

	sb    *sbNotif
	nb    *nbTxn
	retry *retryOps
}

// enqueueTxn adds transaction into the FIFO queue (channel) for execution.
func (scheduler *Scheduler) enqueueTxn(txn *queuedTxn) error {
	if txn.txnType == nbTransaction && txn.nb.isBlocking {
		select {
		case <-scheduler.ctx.Done():
			return ErrClosedScheduler
		case scheduler.txnQueue <- txn:
			return nil
		}
	}
	select {
	case <-scheduler.ctx.Done():
		return ErrClosedScheduler
	case scheduler.txnQueue <- txn:
		return nil
	default:
		return ErrTxnQueueFull
	}
}

// dequeueTxn pull the oldest queued transaction.
func (scheduler *Scheduler) dequeueTxn() (txn *queuedTxn, canceled bool) {
	select {
	case <-scheduler.ctx.Done():
		return nil, true
	case txn = <-scheduler.txnQueue:
		return txn, false
	}
}

// enqueueRetry schedules retry for failed operations.
func (scheduler *Scheduler) enqueueRetry(args *retryOps) {
	go scheduler.delayRetry(args)
}

// delayRetry postpones retry until a given time period has elapsed.
func (scheduler *Scheduler) delayRetry(args *retryOps) {
	scheduler.wg.Add(1)
	defer scheduler.wg.Done()

	select {
	case <-scheduler.ctx.Done():
		return
	case <-time.After(args.period):
		err := scheduler.enqueueTxn(&queuedTxn{txnType: retryFailedOps, retry: args})
		if err != nil {
			scheduler.Log.WithFields(logging.Fields{
				"txnSeqNum": args.txnSeqNum,
				"err":       err,
			}).Warn("Failed to enqueue re-try for failed operations")
			scheduler.enqueueRetry(args) // try again with the same time period
		}
	}
}
