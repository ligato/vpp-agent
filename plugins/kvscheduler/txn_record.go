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
	"strings"
	"time"

	. "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// recordedTxn is used to record executed transaction.
type recordedTxn struct {
	preRecord bool // not yet fully recorded, only args + plan + pre-processing errors

	// timestamps (zero if len(executed) == 0)
	start time.Time
	stop  time.Time

	// arguments
	seqNum             uint
	txnType            txnType
	isFullResync       bool
	isDownstreamResync bool
	description        string
	values             []recordedKVPair

	// result
	preErrors []KeyWithError // pre-processing errors
	planned   recordedTxnOps
	executed  recordedTxnOps
}

// recorderTxnOp is used to record executed/planned transaction operation.
type recordedTxnOp struct {
	// identification
	operation TxnOperation
	key       string
	derived   bool

	// changes
	prevValue  string
	newValue   string
	prevOrigin ValueOrigin
	newOrigin  ValueOrigin
	wasPending bool
	isPending  bool
	prevErr    error
	newErr     error

	// flags
	isRevert bool
	isRetry  bool
}

// recordedKVPair is used to record key-value pair.
type recordedKVPair struct {
	key    string
	value  string
	origin ValueOrigin
}

// recordedTxnOps is a list of recorded executed/planned transaction operations.
type recordedTxnOps []*recordedTxnOp

// recordedTxns is a list of recorded transactions.
type recordedTxns []*recordedTxn

// String returns a *multi-line* human-readable string representation of recorded transaction.
func (txn *recordedTxn) String() string {
	return txn.StringWithOpts(false, 0)
}

// StringWithOpts allows to format string representation of recorded transaction.
func (txn *recordedTxn) StringWithOpts(resultOnly bool, indent int) string {
	var str string
	indent1 := strings.Repeat(" ", indent)
	indent2 := strings.Repeat(" ", indent+4)
	indent3 := strings.Repeat(" ", indent+8)

	if !resultOnly {
		// transaction arguments
		str += indent1 + "* transaction arguments:\n"
		str += indent2 + fmt.Sprintf("- seq-num: %d\n", txn.seqNum)
		if txn.txnType == nbTransaction && (txn.isFullResync || txn.isDownstreamResync) {
			resyncType := "Full-Resync"
			if txn.isDownstreamResync {
				resyncType = "Downstream-Resync"
			}
			str += indent2 + fmt.Sprintf("- type: %s, %s\n", txn.txnType.String(), resyncType)
		} else {
			str += indent2 + fmt.Sprintf("- type: %s\n", txn.txnType.String())
		}
		if txn.description != "" {
			descriptionLines := strings.Split(txn.description, "\n")
			for idx, line := range descriptionLines {
				if idx == 0 {
					str += indent2 + fmt.Sprintf("- description: %s\n", line)
				} else {
					str += indent3 + fmt.Sprintf("%s\n", line)
				}
			}
		}
		if txn.isDownstreamResync {
			goto printOps
		}
		if len(txn.values) == 0 {
			str += indent2 + fmt.Sprintf("- values: NONE\n")
		} else {
			str += indent2 + fmt.Sprintf("- values:\n")
		}
		for _, kv := range txn.values {
			resync := txn.isFullResync || txn.isDownstreamResync
			if resync && kv.origin == FromSB {
				// do not print SB values updated during resync
				continue
			}
			str += indent3 + fmt.Sprintf("- key: %s\n", kv.key)
			str += indent3 + fmt.Sprintf("  value: %s\n", kv.value)
		}

		// pre-processing errors
		if len(txn.preErrors) > 0 {
			str += indent1 + "* pre-processing errors:\n"
			for _, preError := range txn.preErrors {
				str += indent2 + fmt.Sprintf("- key: %s\n", preError.Key)
				str += indent2 + fmt.Sprintf("  error: %s\n", preError.Error.Error())
			}
		}

	printOps:
		// planned operations
		str += indent1 + "* planned operations:\n"
		str += txn.planned.StringWithOpts(indent + 4)
	}

	if !txn.preRecord {
		if len(txn.executed) == 0 {
			str += indent1 + "* executed operations:\n"
		} else {
			str += indent1 + fmt.Sprintf("* executed operations (%s - %s, duration = %s):\n",
				txn.start.String(), txn.stop.String(), txn.stop.Sub(txn.start).String())
		}
		str += txn.executed.StringWithOpts(indent + 4)
	}

	return str
}

// String returns a *multi-line* human-readable string representation of a recorded
// transaction operation.
func (op *recordedTxnOp) String() string {
	return op.StringWithOpts(0, 0)
}

// StringWithOpts allows to format string representation of a transaction operation.
func (op *recordedTxnOp) StringWithOpts(index int, indent int) string {
	var str string
	indent1 := strings.Repeat(" ", indent)
	indent2 := strings.Repeat(" ", indent+4)

	var flags []string
	if op.newOrigin == FromSB {
		flags = append(flags, "NOTIFICATION")
	}
	if op.derived {
		flags = append(flags, "DERIVED")
	}
	if op.isRevert {
		flags = append(flags, "REVERT")
	}
	if op.isRetry {
		flags = append(flags, "RETRY")
	}
	if op.wasPending {
		if op.isPending {
			flags = append(flags, "STILL-PENDING")
		} else {
			flags = append(flags, "WAS-PENDING")
		}
	} else {
		if op.isPending {
			flags = append(flags, "IS-PENDING")
		}
	}

	if index > 0 {
		if len(flags) == 0 {
			str += indent1 + fmt.Sprintf("%d. %s:\n", index, op.operation.String())
		} else {
			str += indent1 + fmt.Sprintf("%d. %s %v:\n", index, op.operation.String(), flags)
		}
	} else {
		if len(flags) == 0 {
			str += indent1 + fmt.Sprintf("%s:\n", op.operation.String())
		} else {
			str += indent1 + fmt.Sprintf("%s %v:\n", op.operation.String(), flags)
		}
	}

	str += indent2 + fmt.Sprintf("- key: %s\n", op.key)
	showPrevForAdd := op.wasPending && op.prevValue != op.newValue
	if op.operation == Modify || (op.operation == Add && showPrevForAdd) {
		str += indent2 + fmt.Sprintf("- prev-value: %s \n", op.prevValue)
		str += indent2 + fmt.Sprintf("- new-value: %s \n", op.newValue)
	}
	if op.operation == Delete || op.operation == Update {
		str += indent2 + fmt.Sprintf("- value: %s \n", op.prevValue)
	}
	if op.operation == Add && !showPrevForAdd {
		str += indent2 + fmt.Sprintf("- value: %s \n", op.newValue)
	}
	if op.prevOrigin != op.newOrigin {
		str += indent2 + fmt.Sprintf("- prev-origin: %s\n", op.prevOrigin.String())
		str += indent2 + fmt.Sprintf("- new-origin: %s\n", op.newOrigin.String())
	}
	if op.prevErr != nil {
		str += indent2 + fmt.Sprintf("- prev-error: %s\n", utils.ErrorToString(op.prevErr))
	}
	if op.newErr != nil {
		str += indent2 + fmt.Sprintf("- error: %s\n", utils.ErrorToString(op.newErr))
	}

	return str
}

// String returns a *multi-line* human-readable string representation of transaction
// operations.
func (ops recordedTxnOps) String() string {
	return ops.StringWithOpts(0)
}

// StringWithOpts allows to format string representation of transaction operations.
func (ops recordedTxnOps) StringWithOpts(indent int) string {
	if len(ops) == 0 {
		return strings.Repeat(" ", indent) + "<NONE>\n"
	}

	var str string
	for idx, op := range ops {
		str += op.StringWithOpts(idx+1, indent)
	}
	return str
}

// String returns a *multi-line* human-readable string representation of a transaction
// list.
func (txns recordedTxns) String() string {
	return txns.StringWithOpts(false, 0)
}

// StringWithOpts allows to format string representation of a transaction list.
func (txns recordedTxns) StringWithOpts(resultOnly bool, indent int) string {
	if len(txns) == 0 {
		return strings.Repeat(" ", indent) + "<NONE>\n"
	}

	var str string
	for idx, txn := range txns {
		str += strings.Repeat(" ", indent) + fmt.Sprintf("Transaction #%d:\n", txn.seqNum)
		str += txn.StringWithOpts(resultOnly, indent+4)
		if idx < len(txns)-1 {
			str += "\n"
		}
	}
	return str
}

// preRecordTxnOp prepares txn operation record - fills attributes that we can even
// before executing the operation.
func (scheduler *Scheduler) preRecordTxnOp(args *applyValueArgs, node graph.Node) *recordedTxnOp {
	prevOrigin := getNodeOrigin(node)
	if prevOrigin == UnknownOrigin {
		// new value
		prevOrigin = args.kv.origin
	}
	return &recordedTxnOp{
		key:        args.kv.key,
		derived:    isNodeDerived(node),
		prevValue:  utils.ProtoToString(node.GetValue()),
		newValue:   utils.ProtoToString(args.kv.value),
		prevOrigin: prevOrigin,
		newOrigin:  args.kv.origin,
		wasPending: isNodePending(node),
		prevErr:    scheduler.getNodeLastError(args.kv.key),
		isRevert:   args.kv.isRevert,
		isRetry:    args.isRetry,
	}
}

// preRecordTransaction logs transaction arguments + plan before execution to
// persist some information in case there is a crash during execution.
func (scheduler *Scheduler) preRecordTransaction(txn *preProcessedTxn, planned recordedTxnOps, preErrors []KeyWithError) *recordedTxn {
	// allocate new transaction record
	record := &recordedTxn{
		preRecord:          true,
		seqNum:             txn.seqNum,
		txnType:            txn.args.txnType,
		isFullResync:       txn.args.txnType == nbTransaction && txn.args.nb.isFullResync,
		isDownstreamResync: txn.args.txnType == nbTransaction && txn.args.nb.isDownstreamResync,
		preErrors:          preErrors,
		planned:            planned,
	}
	if txn.args.txnType == nbTransaction {
		record.description = txn.args.nb.description
	}

	// record values
	for _, kv := range txn.values {
		record.values = append(record.values, recordedKVPair{
			key:    kv.key,
			value:  utils.ProtoToString(kv.value),
			origin: kv.origin,
		})
	}

	// send to the log
	logMsg := "Processing new transaction:\n" + record.StringWithOpts(false, 2)
	//scheduler.Log.Info(logMsg)
	fmt.Println(logMsg)

	return record
}

// recordTransaction records the finalized transaction (log + in-memory).
func (scheduler *Scheduler) recordTransaction(txnRecord *recordedTxn, executed recordedTxnOps, start, stop time.Time) {
	txnRecord.preRecord = false
	txnRecord.start = start
	txnRecord.stop = stop
	txnRecord.executed = executed

	// log txn result
	logMsg := fmt.Sprintf("Finalized transaction (seq-num=%d):\n%s",
		txnRecord.seqNum, txnRecord.StringWithOpts(true, 2))
	//scheduler.Log.Info(logMsg)
	fmt.Println(logMsg)

	// add transaction record into the history
	scheduler.historyLock.Lock()
	scheduler.txnHistory = append(scheduler.txnHistory, txnRecord)
	scheduler.historyLock.Unlock()
}

// getTransactionHistory returns history of transactions started within the specified
// time window, or the full recorded history if the timestamps are zero values.
func (scheduler *Scheduler) getTransactionHistory(since, until time.Time) (history recordedTxns) {
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
			if !scheduler.txnHistory[lastBefore+1].start.Before(since) {
				break
			}
		}
	}

	if !until.IsZero() {
		for ; firstAfter > 0; firstAfter-- {
			if !scheduler.txnHistory[firstAfter-1].start.After(until) {
				break
			}
		}
	}

	return scheduler.txnHistory[lastBefore+1 : firstAfter]
}

// getRecordedTransaction returns record of a transaction referenced by the sequence number.
func (scheduler *Scheduler) getRecordedTransaction(seqNum uint) (txn *recordedTxn) {
	scheduler.historyLock.Lock()
	defer scheduler.historyLock.Unlock()

	for _, txn := range scheduler.txnHistory {
		if txn.seqNum == seqNum {
			return txn
		}
	}

	return nil
}
