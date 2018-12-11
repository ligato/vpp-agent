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
	"sort"

	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// RecordedTxn is used to record executed transaction.
type RecordedTxn struct {
	PreRecord bool // not yet fully recorded, only args + plan + pre-processing errors

	// timestamps
	Start time.Time
	Stop  time.Time

	// arguments
	SeqNum      uint
	TxnType     TxnType
	ResyncType  kvs.ResyncType
	Description string
	Values      []RecordedKVPair

	// result
	PreErrors []kvs.KeyWithError // pre-processing errors
	Planned   RecordedTxnOps
	Executed  RecordedTxnOps
}

// RecordedTxnOp is used to record executed/planned transaction operation.
type RecordedTxnOp struct {
	// identification
	Operation kvs.TxnOperation
	Key       string
	Derived   bool

	// changes
	PrevValue  string
	NewValue   string
	PrevOrigin kvs.ValueOrigin
	NewOrigin  kvs.ValueOrigin
	WasPending bool
	IsPending  bool
	PrevErr    error
	NewErr     error

	// flags
	IsRevert bool
	IsRetry  bool
}

// RecordedKVPair is used to record key-value pair.
type RecordedKVPair struct {
	Key    string
	Value  string
	Origin kvs.ValueOrigin
}

// RecordedTxnOps is a list of recorded executed/planned transaction operations.
type RecordedTxnOps []*RecordedTxnOp

// RecordedTxns is a list of recorded transactions.
type RecordedTxns []*RecordedTxn

// String returns a *multi-line* human-readable string representation of recorded transaction.
/*func (txn *RecordedTxn) String() string {
	return txn.StringWithOpts(false, 0)
}*/

// StringWithOpts allows to format string representation of recorded transaction.
func (txn *RecordedTxn) StringWithOpts(resultOnly bool, indent int) string {
	var str string
	indent1 := strings.Repeat(" ", indent)
	indent2 := strings.Repeat(" ", indent+4)
	indent3 := strings.Repeat(" ", indent+8)

	if !resultOnly {
		// transaction arguments
		str += indent1 + "* transaction arguments:\n"
		str += indent2 + fmt.Sprintf("- seq-num: %d\n", txn.SeqNum)
		if txn.TxnType == nbTransaction && txn.ResyncType != kvs.NotResync {
			ResyncType := "Full Resync"
			if txn.ResyncType == kvs.DownstreamResync {
				ResyncType = "SB Sync"
			}
			if txn.ResyncType == kvs.UpstreamResync {
				ResyncType = "NB Sync"
			}
			str += indent2 + fmt.Sprintf("- type: %s, %s\n", txn.TxnType.String(), ResyncType)
		} else {
			str += indent2 + fmt.Sprintf("- type: %s\n", txn.TxnType.String())
		}
		if txn.Description != "" {
			descriptionLines := strings.Split(txn.Description, "\n")
			for idx, line := range descriptionLines {
				if idx == 0 {
					str += indent2 + fmt.Sprintf("- Description: %s\n", line)
				} else {
					str += indent3 + fmt.Sprintf("%s\n", line)
				}
			}
		}
		if txn.ResyncType == kvs.DownstreamResync {
			goto printOps
		}
		if len(txn.Values) == 0 {
			str += indent2 + fmt.Sprintf("- values: NONE\n")
		} else {
			str += indent2 + fmt.Sprintf("- values:\n")
		}
		for _, kv := range txn.Values {
			if txn.ResyncType != kvs.NotResync && kv.Origin == kvs.FromSB {
				// do not print SB values updated during resync
				continue
			}
			str += indent3 + fmt.Sprintf("- key: %s\n", kv.Key)
			str += indent3 + fmt.Sprintf("  value: %s\n", kv.Value)
		}

		// pre-processing errors
		if len(txn.PreErrors) > 0 {
			str += indent1 + "* pre-processing errors:\n"
			for _, preError := range txn.PreErrors {
				str += indent2 + fmt.Sprintf("- key: %s\n", preError.Key)
				str += indent2 + fmt.Sprintf("  error: %s\n", preError.Error.Error())
			}
		}

	printOps:
		// planned operations
		str += indent1 + "* planned operations:\n"
		str += txn.Planned.StringWithOpts(indent + 4)
	}

	if !txn.PreRecord {
		if len(txn.Executed) == 0 {
			str += indent1 + "* executed operations:\n"
		} else {
			str += indent1 + fmt.Sprintf("* executed operations (%s - %s, duration = %s):\n",
				txn.Start.String(), txn.Stop.String(), txn.Stop.Sub(txn.Start).String())
		}
		str += txn.Executed.StringWithOpts(indent + 4)
	}

	return str
}

// String returns a *multi-line* human-readable string representation of a recorded
// transaction operation.
func (op *RecordedTxnOp) String() string {
	return op.StringWithOpts(0, 0)
}

// StringWithOpts allows to format string representation of a transaction operation.
func (op *RecordedTxnOp) StringWithOpts(index int, indent int) string {
	var str string
	indent1 := strings.Repeat(" ", indent)
	indent2 := strings.Repeat(" ", indent+4)

	var flags []string
	if op.NewOrigin == kvs.FromSB {
		flags = append(flags, "NOTIFICATION")
	}
	if op.Derived {
		flags = append(flags, "DERIVED")
	}
	if op.IsRevert {
		flags = append(flags, "REVERT")
	}
	if op.IsRetry {
		flags = append(flags, "RETRY")
	}
	if op.WasPending {
		if op.IsPending {
			flags = append(flags, "STILL-PENDING")
		} else {
			flags = append(flags, "WAS-PENDING")
		}
	} else {
		if op.IsPending {
			flags = append(flags, "IS-PENDING")
		}
	}

	if index > 0 {
		if len(flags) == 0 {
			str += indent1 + fmt.Sprintf("%d. %s:\n", index, op.Operation.String())
		} else {
			str += indent1 + fmt.Sprintf("%d. %s %v:\n", index, op.Operation.String(), flags)
		}
	} else {
		if len(flags) == 0 {
			str += indent1 + fmt.Sprintf("%s:\n", op.Operation.String())
		} else {
			str += indent1 + fmt.Sprintf("%s %v:\n", op.Operation.String(), flags)
		}
	}

	str += indent2 + fmt.Sprintf("- key: %s\n", op.Key)
	showPrevForAdd := op.WasPending && op.PrevValue != op.NewValue
	if op.Operation == kvs.Modify || (op.Operation == kvs.Add && showPrevForAdd) {
		str += indent2 + fmt.Sprintf("- prev-value: %s \n", op.PrevValue)
		str += indent2 + fmt.Sprintf("- new-value: %s \n", op.NewValue)
	}
	if op.Operation == kvs.Delete || op.Operation == kvs.Update {
		str += indent2 + fmt.Sprintf("- value: %s \n", op.PrevValue)
	}
	if op.Operation == kvs.Add && !showPrevForAdd {
		str += indent2 + fmt.Sprintf("- value: %s \n", op.NewValue)
	}
	if op.PrevOrigin != op.NewOrigin {
		str += indent2 + fmt.Sprintf("- prev-origin: %s\n", op.PrevOrigin.String())
		str += indent2 + fmt.Sprintf("- new-origin: %s\n", op.NewOrigin.String())
	}
	if op.PrevErr != nil {
		str += indent2 + fmt.Sprintf("- prev-error: %s\n", utils.ErrorToString(op.PrevErr))
	}
	if op.NewErr != nil {
		str += indent2 + fmt.Sprintf("- error: %s\n", utils.ErrorToString(op.NewErr))
	}

	return str
}

// String returns a *multi-line* human-readable string representation of transaction
// operations.
func (ops RecordedTxnOps) String() string {
	return ops.StringWithOpts(0)
}

// StringWithOpts allows to format string representation of transaction operations.
func (ops RecordedTxnOps) StringWithOpts(indent int) string {
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
func (txns RecordedTxns) String() string {
	return txns.StringWithOpts(false, 0)
}

// StringWithOpts allows to format string representation of a transaction list.
func (txns RecordedTxns) StringWithOpts(resultOnly bool, indent int) string {
	if len(txns) == 0 {
		return strings.Repeat(" ", indent) + "<NONE>\n"
	}

	var str string
	for idx, txn := range txns {
		str += strings.Repeat(" ", indent) + fmt.Sprintf("Transaction #%d:\n", txn.SeqNum)
		str += txn.StringWithOpts(resultOnly, indent+4)
		if idx < len(txns)-1 {
			str += "\n"
		}
	}
	return str
}

// preRecordTxnOp prepares txn operation record - fills attributes that we can even
// before executing the operation.
func (scheduler *Scheduler) preRecordTxnOp(args *applyValueArgs, node graph.Node) *RecordedTxnOp {
	prevOrigin := getNodeOrigin(node)
	if prevOrigin == kvs.UnknownOrigin {
		// new value
		prevOrigin = args.kv.origin
	}
	return &RecordedTxnOp{
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
func (scheduler *Scheduler) preRecordTransaction(txn *preProcessedTxn, planned RecordedTxnOps, preErrors []kvs.KeyWithError) *RecordedTxn {
	// allocate new transaction record
	record := &RecordedTxn{
		PreRecord:          true,
		SeqNum:             txn.seqNum,
		TxnType:            txn.args.txnType,
		PreErrors:          preErrors,
		Planned:            planned,
	}
	if txn.args.txnType == nbTransaction {
		record.ResyncType = txn.args.nb.resyncType
		record.Description = txn.args.nb.description
	}

	// build header for the log
	var downstreamResync bool
	txnInfo := fmt.Sprintf("%s", txn.args.txnType.String())
	if txn.args.txnType == nbTransaction && txn.args.nb.resyncType != kvs.NotResync {
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
			record.Values = append(record.Values, RecordedKVPair{
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
func (scheduler *Scheduler) recordTransaction(txnRecord *RecordedTxn, executed RecordedTxnOps, start, stop time.Time) {
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

// getTransactionHistory returns history of transactions started within the specified
// time window, or the full recorded history if the timestamps are zero values.
func (scheduler *Scheduler) getTransactionHistory(since, until time.Time) (history RecordedTxns) {
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

// getRecordedTransaction returns record of a transaction referenced by the sequence number.
func (scheduler *Scheduler) getRecordedTransaction(SeqNum uint) (txn *RecordedTxn) {
	scheduler.historyLock.Lock()
	defer scheduler.historyLock.Unlock()

	for _, txn := range scheduler.txnHistory {
		if txn.SeqNum == SeqNum {
			return txn
		}
	}

	return nil
}
