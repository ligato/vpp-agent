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

	. "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/kvscheduler/graph"
)

// txnOperationType differentiates between add, modify (incl. re-create), delete
// and update operations.
type txnOperationType int

const (
	add txnOperationType = iota
	modify
	del
	update
)

// String returns human-readable string representation of transaction operation.
func (txnOpType txnOperationType) String() string {
	switch txnOpType {
	case add:
		return "ADD"
	case modify:
		return "MODIFY"
	case del:
		return "DELETE"
	case update:
		return "UPDATE"
	}
	return "UNKNOWN"
}

// recordedTxn is used to record executed transaction.
type recordedTxn struct {
	preRecord bool // not yet fully recorded, only args + plan + pre-processing errors

	// timestamps (zero if len(executed) == 0)
	start time.Time
	stop  time.Time

	// arguments
	seqNum   uint
	txnType  txnType
	isResync bool
	values   []recordedKVPair

	// result
	preErrors []KeyWithError // pre-processing errors
	planned   recordedTxnOps
	executed  recordedTxnOps
}

// recorderTxnOp is used to record executed/planned transaction operation.
type recordedTxnOp struct {
	// identification
	operation txnOperationType
	key       string

	// changes
	prevValue  *recordedValue
	newValue   *recordedValue
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

// recordedValue is used to record value.
type recordedValue struct {
	valueType ValueType
	label     string
	string    string
}

// recordedKVPair is used to record key-value pair.
type recordedKVPair struct {
	key    string
	value  *recordedValue
	origin ValueOrigin
}

// recordedTxnOps is a list of recorded executed/planned transaction operations.
type recordedTxnOps []*recordedTxnOp

// recordedTxns is a list of recorded transactions.
type recordedTxns []*recordedTxn

// String returns a human-readable string representation of recorded value.
func (value *recordedValue) String() string {
	return value.StringWithOpts(false)
}

// StringWithOpts allows to format string representation of recorded value.
func (value *recordedValue) StringWithOpts(verbose bool) string {
	if value == nil {
		return "NIL"
	}
	if verbose {
		return fmt.Sprintf("%s [label=%s, type=%s]", value.string, value.label, value.valueType)
	}
	return fmt.Sprintf("%s [type=%s]", value.label, value.valueType)
}

// String returns a *multi-line* human-readable string representation of recorded transaction.
func (txn *recordedTxn) String() string {
	return txn.StringWithOpts(false, 0, false)
}

// StringWithOpts allows to format string representation of recorded transaction.
func (txn *recordedTxn) StringWithOpts(resultOnly bool, indent int, verbose bool) string {
	var str string
	indent1 := strings.Repeat(" ", indent)
	indent2 := strings.Repeat(" ", indent+4)
	indent3 := strings.Repeat(" ", indent+8)

	if !resultOnly {
		// transaction arguments
		str += indent1 + "* transaction arguments:\n"
		str += indent2 + fmt.Sprintf("- seq-num: %d\n", txn.seqNum)
		str += indent2 + fmt.Sprintf("- type: %s\n", txn.txnType.String())
		if txn.txnType == nbTransaction {
			str += indent2 + fmt.Sprintf("- is-resync: %t\n", txn.isResync)
		}
		if len(txn.values) == 0 {
			str += indent2 + fmt.Sprintf("- values: NONE\n")
		} else {
			str += indent2 + fmt.Sprintf("- values:\n")
		}
		for _, kv := range txn.values {
			str += indent3 + fmt.Sprintf("- key: %s\n", kv.key)
			str += indent3 + fmt.Sprintf("  value: %s\n", kv.value.StringWithOpts(verbose))
			if txn.isResync {
				str += indent3 + fmt.Sprintf("  origin: %s\n", kv.origin.String())
			}
		}

		// pre-processing errors
		if len(txn.preErrors) > 0 {
			str += indent1 + "* pre-processing errors:\n"
			for _, preError := range txn.preErrors {
				str += indent2 + fmt.Sprintf("- key: %s\n", preError.Key)
				str += indent2 + fmt.Sprintf("  error: %s\n", preError.Error.Error())
			}
		}

		// planned operations
		str += indent1 + "* planned operations:\n"
		str += txn.planned.StringWithOpts(indent+4, verbose)
	}

	if !txn.preRecord {
		if len(txn.executed) == 0 {
			str += indent1 + "* executed operations:\n"
		} else {
			str += indent1 + fmt.Sprintf("* executed operations (%s - %s):\n",
				txn.start.String(), txn.stop.String())
		}
		str += txn.executed.StringWithOpts(indent+4, verbose)
	}

	return str
}

// String returns a *multi-line* human-readable string representation of a recorded
// transaction operation.
func (op *recordedTxnOp) String() string {
	return op.StringWithOpts(0, 0, false)
}

// StringWithOpts allows to format string representation of a transaction operation.
func (op *recordedTxnOp) StringWithOpts(index int, indent int, verbose bool) string {
	var str string
	indent1 := strings.Repeat(" ", indent)
	indent2 := strings.Repeat(" ", indent+4)

	if index > 0 {
		str += indent1 + fmt.Sprintf("%d. %s:\n", index, op.operation.String())
	} else {
		str += indent1 + fmt.Sprintf("%s:\n", op.operation.String())
	}

	str += indent2 + fmt.Sprintf("- key: %s\n", op.key)
	str += indent2 + fmt.Sprintf("- prev-value: %s\n", op.prevValue.StringWithOpts(verbose))
	str += indent2 + fmt.Sprintf("- new-value: %s\n", op.newValue.StringWithOpts(verbose))
	str += indent2 + fmt.Sprintf("- prev-origin: %s\n", op.prevOrigin.String())
	str += indent2 + fmt.Sprintf("- new-origin: %s\n", op.newOrigin.String())
	str += indent2 + fmt.Sprintf("- was-pending: %t\n", op.wasPending)
	str += indent2 + fmt.Sprintf("- is-pending: %t\n", op.isPending)
	str += indent2 + fmt.Sprintf("- prev-error: %s\n", errorToString(op.prevErr))
	str += indent2 + fmt.Sprintf("- new-error: %s\n", errorToString(op.newErr))
	str += indent2 + fmt.Sprintf("- is-revert: %t\n", op.isRevert)
	str += indent2 + fmt.Sprintf("- is-retry: %t\n", op.isRetry)

	return str
}

// String returns a *multi-line* human-readable string representation of transaction
// operations.
func (ops recordedTxnOps) String() string {
	return ops.StringWithOpts(0, false)
}

// StringWithOpts allows to format string representation of transaction operations.
func (ops recordedTxnOps) StringWithOpts(indent int, verbose bool) string {
	if len(ops) == 0 {
		return strings.Repeat(" ", indent) + "<NONE>\n"
	}

	var str string
	for idx, op := range ops {
		str += op.StringWithOpts(idx+1, indent, verbose)
	}
	return str
}

// String returns a *multi-line* human-readable string representation of a transaction
// list.
func (txns recordedTxns) String() string {
	return txns.StringWithOpts(false, 0, false)
}

// StringWithOpts allows to format string representation of a transaction list.
func (txns recordedTxns) StringWithOpts(resultOnly bool, indent int, verbose bool) string {
	if len(txns) == 0 {
		return strings.Repeat(" ", indent) + "<NONE>\n"
	}

	var str string
	for idx, txn := range txns {
		str += strings.Repeat(" ", indent) + fmt.Sprintf("Transaction #%d:\n", txn.seqNum)
		str += txn.StringWithOpts(resultOnly, indent+4, verbose)
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
		prevValue:  scheduler.recordValue(node.GetValue()),
		newValue:   scheduler.recordValue(args.kv.value),
		prevOrigin: prevOrigin,
		newOrigin:  args.kv.origin,
		wasPending: isNodePending(node),
		prevErr:    getNodeError(node),
		isRevert:   args.kv.isRevert,
		isRetry:    args.isRetry,
	}
}

func (scheduler *Scheduler) recordValue(value Value) *recordedValue {
	if value == nil {
		return nil
	}
	return &recordedValue{
		valueType: value.Type(),
		label:     value.Label(),
		string:    value.String(),
	}
}

// preRecordTransaction logs transaction arguments + plan before execution to
// persist some information in case there is a crash during execution.
func (scheduler *Scheduler) preRecordTransaction(txn *preProcessedTxn, planned recordedTxnOps, preErrors []KeyWithError) *recordedTxn {
	// allocate new transaction record
	record := &recordedTxn{
		preRecord: true,
		seqNum:    txn.seqNum,
		txnType:   txn.args.txnType,
		isResync:  txn.args.txnType == nbTransaction && txn.args.nb.isResync,
		preErrors: preErrors,
		planned:   planned,
	}

	// record values
	for _, kv := range txn.values {
		record.values = append(record.values, recordedKVPair{
			key:    kv.key,
			value:  scheduler.recordValue(kv.value),
			origin: kv.origin,
		})
	}

	// send to the log
	logMsg := "Processing new transaction:\n" + record.StringWithOpts(false, 2, true)
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
		txnRecord.seqNum, txnRecord.StringWithOpts(true, 2, true))
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

	return scheduler.txnHistory[lastBefore+1:firstAfter]
}

func errorToString(err error) string {
	if err == nil {
		return "<NIL>"
	}
	return err.Error()
}