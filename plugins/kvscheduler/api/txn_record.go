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

package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// TxnType differentiates between NB transaction, retry of failed operations and
// SB notification. Once queued, all three different operations are classified
// as transactions, only with different parameters.
type TxnType int

const (
	// SBNotification is notification from southbound.
	SBNotification TxnType = iota

	// NBTransaction is transaction from northbound.
	NBTransaction

	// RetryFailedOps is a transaction re-trying failed operations from previous
	// northbound transaction.
	RetryFailedOps
)

// String returns human-readable string representation of the transaction type.
func (t TxnType) String() string {
	switch t {
	case SBNotification:
		return "SB notification"
	case NBTransaction:
		return "NB transaction"
	case RetryFailedOps:
		return "RETRY"
	}
	return "UNKNOWN"
}

// RecordedTxn is used to record executed transaction.
type RecordedTxn struct {
	PreRecord bool // not yet fully recorded, only args + plan + pre-processing errors

	// timestamps
	Start time.Time
	Stop  time.Time

	// arguments
	SeqNum      uint64
	TxnType     TxnType
	ResyncType  ResyncType
	Description string
	Values      []RecordedKVPair

	// result
	PreErrors []KeyWithError // pre-processing errors
	Planned   RecordedTxnOps
	Executed  RecordedTxnOps
}

// RecordedTxnOp is used to record executed/planned transaction operation.
type RecordedTxnOp struct {
	// identification
	Operation TxnOperation
	Key       string
	Derived   bool

	// changes
	PrevValue  proto.Message
	NewValue   proto.Message
	PrevOrigin ValueOrigin
	NewOrigin  ValueOrigin
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
	Value  proto.Message
	Origin ValueOrigin
}

// RecordedTxnOps is a list of recorded executed/planned transaction operations.
type RecordedTxnOps []*RecordedTxnOp

// RecordedTxns is a list of recorded transactions.
type RecordedTxns []*RecordedTxn

// String returns a *multi-line* human-readable string representation of recorded transaction.
func (txn *RecordedTxn) String() string {
	return txn.StringWithOpts(false, 0)
}

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
		if txn.TxnType == NBTransaction && txn.ResyncType != NotResync {
			ResyncType := "Full Resync"
			if txn.ResyncType == DownstreamResync {
				ResyncType = "SB Sync"
			}
			if txn.ResyncType == UpstreamResync {
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
		if txn.ResyncType == DownstreamResync {
			goto printOps
		}
		if len(txn.Values) == 0 {
			str += indent2 + fmt.Sprintf("- values: NONE\n")
		} else {
			str += indent2 + fmt.Sprintf("- values:\n")
		}
		for _, kv := range txn.Values {
			if txn.ResyncType != NotResync && kv.Origin == FromSB {
				// do not print SB values updated during resync
				continue
			}
			str += indent3 + fmt.Sprintf("- key: %s\n", kv.Key)
			str += indent3 + fmt.Sprintf("  value: %s\n", utils.ProtoToString(kv.Value))
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
	if op.NewOrigin == FromSB {
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
	showPrevForAdd := op.WasPending && !proto.Equal(op.PrevValue, op.NewValue)
	if op.Operation == Modify || (op.Operation == Add && showPrevForAdd) {
		str += indent2 + fmt.Sprintf("- prev-value: %s \n", utils.ProtoToString(op.PrevValue))
		str += indent2 + fmt.Sprintf("- new-value: %s \n", utils.ProtoToString(op.NewValue))
	}
	if op.Operation == Delete || op.Operation == Update {
		str += indent2 + fmt.Sprintf("- value: %s \n", utils.ProtoToString(op.PrevValue))
	}
	if op.Operation == Add && !showPrevForAdd {
		str += indent2 + fmt.Sprintf("- value: %s \n", utils.ProtoToString(op.NewValue))
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
