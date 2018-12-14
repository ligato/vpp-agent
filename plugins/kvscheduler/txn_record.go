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
	"fmt"
	"sort"
	"strings"
	"time"

	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// preRecordTxnOp prepares txn operation record - fills attributes that we can even
// before executing the operation.
func (scheduler *Scheduler) preRecordTxnOp(args *applyValueArgs, node graph.Node) *kvs.RecordedTxnOp {
	prevOrigin := getNodeOrigin(node)
	if prevOrigin == kvs.UnknownOrigin {
		// new value
		prevOrigin = args.kv.origin
	}
	return &kvs.RecordedTxnOp{
		Key:        args.kv.key,
		Derived:    isNodeDerived(node),
		PrevValue:  utils.ProtoToString(node.GetValue()),
		NewValue:   utils.ProtoToString(args.kv.value),
		PrevOrigin: prevOrigin,
		NewOrigin:  args.kv.origin,
		WasPending: isNodePending(node),
		PrevErr:    scheduler.getNodeLastError(args.kv.key),
		IsRevert:   args.kv.isRevert,
		IsRetry:    args.isRetry,
	}
}

// preRecordTransaction logs transaction arguments + plan before execution to
// persist some information in case there is a crash during execution.
func (scheduler *Scheduler) preRecordTransaction(txn *preProcessedTxn, planned kvs.RecordedTxnOps, preErrors []kvs.KeyWithError) *kvs.RecordedTxn {
	// allocate new transaction record
	record := &kvs.RecordedTxn{
		PreRecord: true,
		SeqNum:    txn.seqNum,
		TxnType:   txn.args.txnType,
		PreErrors: preErrors,
		Planned:   planned,
	}
	if txn.args.txnType == kvs.NBTransaction {
		record.ResyncType = txn.args.nb.resyncType
		record.Description = txn.args.nb.description
	}

	// build header for the log
	var downstreamResync bool
	txnInfo := fmt.Sprintf("%s", txn.args.txnType.String())
	if txn.args.txnType == kvs.NBTransaction && txn.args.nb.resyncType != kvs.NotResync {
		ResyncType := "Full Resync"
		if txn.args.nb.resyncType == kvs.DownstreamResync {
			ResyncType = "SB Sync"
			downstreamResync = true
		}
		if txn.args.nb.resyncType == kvs.UpstreamResync {
			ResyncType = "NB Sync"
		}
		txnInfo = fmt.Sprintf("%s (%s)", txn.args.txnType.String(), ResyncType)
	}

	// record values sorted alphabetically by keys
	if !downstreamResync {
		for _, kv := range txn.values {
			record.Values = append(record.Values, kvs.RecordedKVPair{
				Key:    kv.key,
				Value:  utils.ProtoToString(kv.value),
				Origin: kv.origin,
			})
		}
		sort.Slice(record.Values, func(i, j int) bool {
			return record.Values[i].Key < record.Values[j].Key
		})
	}

	// send to the log
	var buf strings.Builder
	buf.WriteString("+======================================================================================================================+\n")
	msg := fmt.Sprintf("Transaction #%d", record.SeqNum)
	n := 115 - len(msg)
	buf.WriteString(fmt.Sprintf("| %s %"+fmt.Sprint(n)+"s |\n", msg, txnInfo))
	buf.WriteString("+======================================================================================================================+\n")
	buf.WriteString(record.StringWithOpts(false, 2))
	fmt.Println(buf.String())

	return record
}

// recordTransaction records the finalized transaction (log + in-memory).
func (scheduler *Scheduler) recordTransaction(txnRecord *kvs.RecordedTxn, executed kvs.RecordedTxnOps, start, stop time.Time) {
	txnRecord.PreRecord = false
	txnRecord.Start = start
	txnRecord.Stop = stop
	txnRecord.Executed = executed

	var buf strings.Builder
	buf.WriteString("o----------------------------------------------------------------------------------------------------------------------o\n")
	buf.WriteString(txnRecord.StringWithOpts(true, 2))
	buf.WriteString("x----------------------------------------------------------------------------------------------------------------------x\n")
	msg := fmt.Sprintf("#%d", txnRecord.SeqNum)
	msg2 := fmt.Sprintf("took %v", stop.Sub(start).Round(time.Millisecond))
	buf.WriteString(fmt.Sprintf("x %s %"+fmt.Sprint(115-len(msg))+"s x\n", msg, msg2))
	buf.WriteString("x----------------------------------------------------------------------------------------------------------------------x\n")
	fmt.Println(buf.String())

	// add transaction record into the history
	scheduler.historyLock.Lock()
	scheduler.txnHistory = append(scheduler.txnHistory, txnRecord)
	scheduler.historyLock.Unlock()
}

// GetTransactionHistory returns history of transactions started within the specified
// time window, or the full recorded history if the timestamps are zero values.
func (scheduler *Scheduler) GetTransactionHistory(since, until time.Time) (history kvs.RecordedTxns) {
	scheduler.historyLock.Lock()
	defer scheduler.historyLock.Unlock()

	if !since.IsZero() && !until.IsZero() && until.Before(since) {
		// invalid time window
		return
	}

	lastBefore := -1
	firstAfter := len(scheduler.txnHistory)

	if !since.IsZero() {
		for ; lastBefore+1 < len(scheduler.txnHistory); lastBefore++ {
			if !scheduler.txnHistory[lastBefore+1].Start.Before(since) {
				break
			}
		}
	}

	if !until.IsZero() {
		for ; firstAfter > 0; firstAfter-- {
			if !scheduler.txnHistory[firstAfter-1].Start.After(until) {
				break
			}
		}
	}

	return scheduler.txnHistory[lastBefore+1 : firstAfter]
}

// GetRecordedTransaction returns record of a transaction referenced by the sequence number.
func (scheduler *Scheduler) GetRecordedTransaction(SeqNum uint) (txn *kvs.RecordedTxn) {
	scheduler.historyLock.Lock()
	defer scheduler.historyLock.Unlock()

	for _, txn := range scheduler.txnHistory {
		if txn.SeqNum == SeqNum {
			return txn
		}
	}

	return nil
}
