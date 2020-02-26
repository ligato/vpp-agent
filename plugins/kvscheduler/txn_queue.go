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
	"time"

	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

// enqueueTxn adds transaction into the FIFO queue (channel) for execution.
func (s *Scheduler) enqueueTxn(txn *transaction) error {
	if txn.ctx == nil {
		txn.ctx = context.TODO()
	}
	//trace.Log(txn.ctx, "txn", "enqueue")
	if txn.txnType == kvs.NBTransaction && txn.nb.isBlocking {
		select {
		case <-s.ctx.Done():
			return kvs.ErrClosedScheduler
		case s.txnQueue <- txn:
			reportQueued(1)
			return nil
		}
	}
	select {
	case <-s.ctx.Done():
		return kvs.ErrClosedScheduler
	case s.txnQueue <- txn:
		reportQueued(1)
		return nil
	default:
		reportTxnDropped()
		return kvs.ErrTxnQueueFull
	}
}

// dequeueTxn pulls the oldest queued transaction.
func (s *Scheduler) dequeueTxn() (txn *transaction, canceled bool) {
	select {
	case <-s.ctx.Done():
		return nil, true
	case txn = <-s.txnQueue:
		reportQueued(-1)
		//trace.Log(txn.ctx, "txn", "dequeue")
		return txn, false
	}
}

// enqueueRetry schedules retry for failed operations.
func (s *Scheduler) enqueueRetry(args *retryTxn) {
	go s.delayRetry(args)
}

// delayRetry postpones retry until a given time period has elapsed.
func (s *Scheduler) delayRetry(args *retryTxn) {
	s.wg.Add(1)
	defer s.wg.Done()

	select {
	case <-s.ctx.Done():
		return
	case <-time.After(args.delay):
		err := s.enqueueTxn(&transaction{
			txnType: kvs.RetryFailedOps,
			retry:   args,
			created: time.Now(),
		})
		if err != nil {
			s.Log.WithFields(logging.Fields{
				"txnSeqNum": args.txnSeqNum,
				"err":       err,
			}).Warn("Failed to enqueue retry transaction for failed operations")
			s.enqueueRetry(args) // try again with the same time period
		}
	}
}
