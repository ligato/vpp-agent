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
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"strings"
)

var (
	// ErrCombinedDownstreamResyncWithChange is returned when transaction combines downstream-resync with data changes.
	ErrCombinedDownstreamResyncWithChange = errors.New("downstream resync combined with data changes in one transaction")

	// ErrRevertNotSupportedWithResync is returned when transaction combines resync with revert.
	ErrRevertNotSupportedWithResync = errors.New("it is not supported to combine resync with revert")

	// ErrClosedScheduler is returned when scheduler is closed during transaction execution.
	ErrClosedScheduler = errors.New("scheduler was closed")

	// ErrTxnWaitCanceled is returned when waiting for result of blocking transaction is canceled.
	ErrTxnWaitCanceled = errors.New("waiting for result of blocking transaction was canceled")

	// ErrTxnQueueFull is returned when the queue of pending transactions is full.
	ErrTxnQueueFull = errors.New("transaction queue is full")

	// ErrUnregisteredValueType is returned for non-derived values whose proto.Message type
	// is not registered.
	ErrUnregisteredValueType = errors.New("protobuf message type is not registered")

	// ErrUnimplementedKey is returned for non-derived values without provided descriptor.
	ErrUnimplementedKey = errors.New("unimplemented key")

	// ErrUnimplementedAdd is returned when NB transaction attempts to Add value
	// for which there is a descriptor, but Add operation is not implemented.
	ErrUnimplementedAdd = errors.New("Add operation is not implemented")

	// ErrUnimplementedDelete is returned when NB transaction attempts to Delete value
	// for which there is a descriptor, but Delete operation is not implemented.
	ErrUnimplementedDelete = errors.New("Delete operation is not implemented")

	// ErrUnimplementedModify is returned when NB transaction attempts to Modify value
	// for which there is a descriptor, but Modify operation is not implemented.
	ErrUnimplementedModify = errors.New("Modify operation is not implemented")
)

// ErrInvalidValueType is returned to scheduler by auto-generated descriptor adapter
// when value does not match expected type.
func ErrInvalidValueType(key string, value proto.Message) error {
	if key == "" {
		return fmt.Errorf("value (%s) has invalid type", value.String())
	}
	return fmt.Errorf("value (%s) has invalid type for key: %s", value.String(), key)
}

// ErrInvalidMetadataType is returned to scheduler by auto-generated descriptor adapter
// when value metadata does not match expected type.
func ErrInvalidMetadataType(key string) error {
	if key == "" {
		return errors.New("metadata has invalid type")
	}
	return fmt.Errorf("metadata has invalid type for key: %s", key)
}

/****************************** Transaction Error *****************************/

// TransactionError implements Error interface, wrapping all errors encountered
// during the processing of a single transaction.
type TransactionError struct {
	txnInitError error
	kvErrors     []KeyWithError
}

// NewTransactionError is a constructor for transaction error.
func NewTransactionError(txnInitError error, kvErrors []KeyWithError) *TransactionError {
	return &TransactionError{txnInitError: txnInitError, kvErrors: kvErrors}
}

// Error returns a string representation of all errors encountered during
// the transaction processing.
func (e *TransactionError) Error() string {
	if e == nil {
		return ""
	}
	if e.txnInitError != nil {
		return e.txnInitError.Error()
	}
	if len(e.kvErrors) > 0 {
		var kvErrMsgs []string
		for _, kvError := range e.kvErrors {
			kvErrMsgs = append(kvErrMsgs,
				fmt.Sprintf("%s (%v): %v", kvError.Key, kvError.TxnOperation, kvError.Error))
			return fmt.Sprintf("failed key-value pairs: [%s]", strings.Join(kvErrMsgs, ", "))
		}
	}
	return ""
}

// GetKVErrors returns errors for key-value pairs that failed to get applied.
func (e *TransactionError) GetKVErrors() (kvErrors []KeyWithError) {
	if e == nil {
		return kvErrors
	}
	return e.kvErrors
}

// GetTxnInitError returns error thrown during the transaction initialization.
// If the transaction initialization fails, the other stages of the transaction
// processing are not even started, therefore either GetTxnInitError or GetKVErrors
// may return some errors, but not both.
func (e *TransactionError) GetTxnInitError() error {
	if e == nil {
		return nil
	}
	return e.txnInitError
}
