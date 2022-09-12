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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/utils"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
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
		return "SBNotification"
	case NBTransaction:
		return "NBTransaction"
	case RetryFailedOps:
		return "RetryFailedOps"
	}
	return "UndefinedTxnType"
}

var txnType_value = map[string]int{
	"SBNotification": int(SBNotification),
	"NBTransaction":  int(NBTransaction),
	"RetryFailedOps": int(RetryFailedOps),
}

func (t TxnType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *TxnType) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if v, ok := txnType_value[s]; ok {
			*t = TxnType(v)
		} else {
			*t = TxnType(-1)
		}
	} else {
		var n int
		if err := json.Unmarshal(b, &n); err != nil {
			return err
		}
		*t = TxnType(n)
	}
	return nil
}

func TxnTypeToString(t TxnType) string {
	switch t {
	case NBTransaction:
		return "NB Transaction"
	case SBNotification:
		return "SB Notification"
	case RetryFailedOps:
		return "Retry Transaction"
	}
	return t.String()
}

func ResyncTypeToString(t ResyncType) string {
	switch t {
	case NotResync:
		return "Not Resync"
	case FullResync:
		return "Full Resync"
	case UpstreamResync:
		return "NB Sync"
	case DownstreamResync:
		return "SB Sync"
	}
	return t.String()
}

// RecordedTxn is used to record executed transaction.
type RecordedTxn struct {
	PreRecord      bool `json:",omitempty"` // not yet fully recorded, only args + plan + pre-processing errors
	WithSimulation bool `json:",omitempty"`

	// timestamps
	Start time.Time
	Stop  time.Time

	// arguments
	SeqNum       uint64
	TxnType      TxnType
	ResyncType   ResyncType       `json:",omitempty"`
	Description  string           `json:",omitempty"`
	RetryForTxn  uint64           `json:",omitempty"`
	RetryAttempt int              `json:",omitempty"`
	Values       []RecordedKVPair `json:",omitempty"`

	// operations
	Planned  RecordedTxnOps `json:",omitempty"`
	Executed RecordedTxnOps `json:",omitempty"`
}

// RecordedTxnOp is used to record executed/planned transaction operation.
type RecordedTxnOp struct {
	// identification
	Operation kvscheduler.TxnOperation
	Key       string

	// changes
	NewState   kvscheduler.ValueState      `json:",omitempty"`
	NewValue   *utils.RecordedProtoMessage `json:",omitempty"`
	NewErr     error                       `json:"-"`
	NewErrMsg  string                      `json:",omitempty"`
	PrevState  kvscheduler.ValueState      `json:",omitempty"`
	PrevValue  *utils.RecordedProtoMessage `json:",omitempty"`
	PrevErr    error                       `json:"-"`
	PrevErrMsg string                      `json:",omitempty"`
	NOOP       bool                        `json:",omitempty"`

	// flags
	IsDerived  bool `json:",omitempty"`
	IsProperty bool `json:",omitempty"`
	IsRevert   bool `json:",omitempty"`
	IsRetry    bool `json:",omitempty"`
	IsRecreate bool `json:",omitempty"`
}

// RecordedKVPair is used to record key-value pair.
type RecordedKVPair struct {
	Key    string
	Value  *utils.RecordedProtoMessage
	Origin ValueOrigin
}

// RecordedTxnOps is a list of recorded executed/planned transaction operations.
type RecordedTxnOps []*RecordedTxnOp

// RecordedTxns is a list of recorded transactions.
type RecordedTxns []*RecordedTxn

// RecordedKVWithMetadata is the same as KVWithMetadata but with the field Value
// of type utils.RecordedProtoMessage instead of proto.Message. This allows for
// proper JSON marshalling and unmarshalling. Values of this type are used in
// KVScheduler's REST API.
type RecordedKVWithMetadata struct {
	RecordedKVPair
	Metadata Metadata
}

// String returns a *multi-line* human-readable string representation of recorded transaction.
func (txn *RecordedTxn) String() string {
	return txn.StringWithOpts(false, false, 0)
}

// StringWithOpts allows to format string representation of recorded transaction.
func (txn *RecordedTxn) StringWithOpts(resultOnly, verbose bool, indent int) string {
	var str string
	indent1 := strings.Repeat(" ", indent)
	indent2 := strings.Repeat(" ", indent+4)
	indent3 := strings.Repeat(" ", indent+8)

	if !resultOnly {
		// transaction arguments
		str += indent1 + "* transaction arguments:\n"
		str += indent2 + fmt.Sprintf("- seqNum: %d\n", txn.SeqNum)
		if txn.TxnType == NBTransaction && txn.ResyncType != NotResync {
			str += indent2 + fmt.Sprintf("- type: %s, %s\n", TxnTypeToString(txn.TxnType), ResyncTypeToString(txn.ResyncType))
		} else {
			if txn.TxnType == RetryFailedOps {
				str += indent2 + fmt.Sprintf("- type: %s (for txn %d, attempt #%d)\n",
					TxnTypeToString(txn.TxnType), txn.RetryForTxn, txn.RetryAttempt)
			} else {
				str += indent2 + fmt.Sprintf("- type: %s\n", TxnTypeToString(txn.TxnType))
			}
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
			str += indent3 + fmt.Sprintf("  val: %s\n", utils.ProtoToString(kv.Value))
		}

	printOps:
		// planned operations
		if txn.WithSimulation {
			str += indent1 + "* planned operations:\n"
			str += txn.Planned.StringWithOpts(verbose, indent+4)
		}
	}

	if !txn.PreRecord {
		if len(txn.Executed) == 0 {
			str += indent1 + "* executed operations:\n"
		} else {
			str += indent1 + fmt.Sprintf("* executed operations (%s -> %s, dur: %s):\n",
				txn.Start.Round(time.Millisecond),
				txn.Stop.Round(time.Millisecond),
				txn.Stop.Sub(txn.Start).Round(time.Millisecond))
		}
		str += txn.Executed.StringWithOpts(verbose, indent+4)
	}

	return str
}

// String returns a *multi-line* human-readable string representation of a recorded
// transaction operation.
func (op *RecordedTxnOp) String() string {
	return op.StringWithOpts(0, false, 0)
}

// StringWithOpts allows to format string representation of a transaction operation.
func (op *RecordedTxnOp) StringWithOpts(index int, verbose bool, indent int) string {
	var str string
	indent1 := strings.Repeat(" ", indent)
	indent2 := strings.Repeat(" ", indent+4)

	var flags []string
	// operation flags
	if op.IsDerived && !op.IsProperty {
		flags = append(flags, "DERIVED")
	}
	if op.IsProperty {
		flags = append(flags, "PROPERTY")
	}
	if op.NOOP {
		flags = append(flags, "NOOP")
	}
	if op.IsRevert && !op.IsProperty {
		flags = append(flags, "REVERT")
	}
	if op.IsRetry && !op.IsProperty {
		flags = append(flags, "RETRY")
	}
	if op.IsRecreate {
		flags = append(flags, "RECREATE")
	}
	// value state transition
	//  -> OBTAINED
	if op.NewState == kvscheduler.ValueState_OBTAINED {
		flags = append(flags, "OBTAINED")
	}
	if op.PrevState == kvscheduler.ValueState_OBTAINED && op.PrevState != op.NewState {
		flags = append(flags, "WAS-OBTAINED")
	}
	//  -> UNIMPLEMENTED
	if op.NewState == kvscheduler.ValueState_UNIMPLEMENTED {
		flags = append(flags, "UNIMPLEMENTED")
	}
	if op.PrevState == kvscheduler.ValueState_UNIMPLEMENTED && op.PrevState != op.NewState {
		flags = append(flags, "WAS-UNIMPLEMENTED")
	}
	//  -> REMOVED / MISSING
	if op.PrevState == kvscheduler.ValueState_REMOVED && op.Operation == kvscheduler.TxnOperation_DELETE {
		flags = append(flags, "ALREADY-REMOVED")
	}
	if op.PrevState == kvscheduler.ValueState_MISSING {
		if op.NewState == kvscheduler.ValueState_REMOVED {
			flags = append(flags, "ALREADY-MISSING")
		} else {
			flags = append(flags, "WAS-MISSING")
		}
	}
	//  -> DISCOVERED
	if op.PrevState == kvscheduler.ValueState_DISCOVERED {
		flags = append(flags, "DISCOVERED")
	}
	//  -> PENDING
	if op.PrevState == kvscheduler.ValueState_PENDING {
		if op.NewState == kvscheduler.ValueState_PENDING {
			flags = append(flags, "STILL-PENDING")
		} else {
			flags = append(flags, "WAS-PENDING")
		}
	} else {
		if op.NewState == kvscheduler.ValueState_PENDING {
			flags = append(flags, "IS-PENDING")
		}
	}
	//  -> FAILED / INVALID
	if op.PrevState == kvscheduler.ValueState_FAILED {
		if op.NewState == kvscheduler.ValueState_FAILED {
			flags = append(flags, "STILL-FAILING")
		} else if op.NewState == kvscheduler.ValueState_CONFIGURED {
			flags = append(flags, "FIXED")
		}
	} else {
		if op.NewState == kvscheduler.ValueState_FAILED {
			flags = append(flags, "FAILED")
		}
	}
	if op.PrevState == kvscheduler.ValueState_INVALID {
		if op.NewState == kvscheduler.ValueState_INVALID {
			flags = append(flags, "STILL-INVALID")
		} else if op.NewState == kvscheduler.ValueState_CONFIGURED {
			flags = append(flags, "FIXED")
		}
	} else {
		if op.NewState == kvscheduler.ValueState_INVALID {
			flags = append(flags, "INVALID")
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
	if op.Operation == kvscheduler.TxnOperation_UPDATE {
		str += indent2 + fmt.Sprintf("- prev-value: %s \n", utils.ProtoToString(op.PrevValue))
		str += indent2 + fmt.Sprintf("- new-value: %s \n", utils.ProtoToString(op.NewValue))
	}
	if op.Operation == kvscheduler.TxnOperation_DELETE {
		str += indent2 + fmt.Sprintf("- value: %s \n", utils.ProtoToString(op.PrevValue))
	}
	if op.Operation == kvscheduler.TxnOperation_CREATE {
		str += indent2 + fmt.Sprintf("- value: %s \n", utils.ProtoToString(op.NewValue))
	}
	if op.PrevErr != nil {
		str += indent2 + fmt.Sprintf("- prev-error: %s\n", utils.ErrorToString(op.PrevErr))
	}
	if op.NewErr != nil {
		str += indent2 + fmt.Sprintf("- error: %s\n", utils.ErrorToString(op.NewErr))
	}
	if verbose {
		str += indent2 + fmt.Sprintf("- prev-state: %s \n", op.PrevState.String())
		str += indent2 + fmt.Sprintf("- new-state: %s \n", op.NewState.String())
	}

	return str
}

// String returns a *multi-line* human-readable string representation of transaction
// operations.
func (ops RecordedTxnOps) String() string {
	return ops.StringWithOpts(false, 0)
}

// StringWithOpts allows to format string representation of transaction operations.
func (ops RecordedTxnOps) StringWithOpts(verbose bool, indent int) string {
	if len(ops) == 0 {
		return strings.Repeat(" ", indent) + "<NONE>\n"
	}

	var str string
	for idx, op := range ops {
		str += op.StringWithOpts(idx+1, verbose, indent)
	}
	return str
}

// String returns a *multi-line* human-readable string representation of a transaction
// list.
func (txns RecordedTxns) String() string {
	return txns.StringWithOpts(false, false, 0)
}

// StringWithOpts allows to format string representation of a transaction list.
func (txns RecordedTxns) StringWithOpts(resultOnly, verbose bool, indent int) string {
	if len(txns) == 0 {
		return strings.Repeat(" ", indent) + "<NONE>\n"
	}

	var str string
	for idx, txn := range txns {
		str += strings.Repeat(" ", indent) + fmt.Sprintf("Transaction #%d:\n", txn.SeqNum)
		str += txn.StringWithOpts(resultOnly, verbose, indent+4)
		if idx < len(txns)-1 {
			str += "\n"
		}
	}
	return str
}
