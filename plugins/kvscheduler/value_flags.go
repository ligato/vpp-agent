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
	"time"

	"github.com/gogo/protobuf/proto"

	. "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

const (
	////// updated by transactions:

	// LastChangeFlagName is the name of the LastChange flag.
	LastChangeFlagName = "last-change"

	// LastUpdateFlagName is the name of the LastUpdate flag.
	LastUpdateFlagName = "last-update"

	// ErrorFlagName is the name of the Error flag.
	ErrorFlagName = "error"

	////// updated by transactions + refresh:

	// PendingFlagName is the name of the Pending flag.
	PendingFlagName = "pending"

	// OriginFlagName is the name of the Origin flag.
	OriginFlagName = "origin"

	// DescriptorFlagName is the name of the Descriptor flag.
	DescriptorFlagName = "descriptor"

	// DerivedFlagName is the name of the Derived flag.
	DerivedFlagName = "derived"
)

// LastChangeFlag is set to all base values to remember the last change from
// a NB transaction or a SB notification for a potential retry.
type LastChangeFlag struct {
	txnSeqNum uint
	value     proto.Message
	origin    ValueOrigin
	revert    bool

	// NB txn options
	retryEnabled    bool
	retryPeriod     time.Duration
	retryExpBackoff bool
}

// GetName return name of the LastChange flag.
func (flag *LastChangeFlag) GetName() string {
	return LastChangeFlagName
}

// GetValue describes the last change (txn-seq number only).
func (flag *LastChangeFlag) GetValue() string {
	return fmt.Sprintf("TXN-%d", flag.txnSeqNum)
}

// LastUpdateFlag is set to all values to remember the last transaction which
// has changed/updated the value.
type LastUpdateFlag struct {
	txnSeqNum uint
}

// GetName return name of the LastUpdate flag.
func (flag *LastUpdateFlag) GetName() string {
	return LastUpdateFlagName
}

// GetValue return the sequence number of the last transaction that performed
// update.
func (flag *LastUpdateFlag) GetValue() string {
	return fmt.Sprintf("TXN-%d", flag.txnSeqNum)
}

// PendingFlag is used to mark values that cannot be created because dependencies
// are not satisfied or the Add operation has failed.
type PendingFlag struct {
}

// GetName return name of the Pending flag.
func (flag *PendingFlag) GetName() string {
	return PendingFlagName
}

// GetValue return empty string (presence of the flag is the only information).
func (flag *PendingFlag) GetValue() string {
	return ""
}

// OriginFlag is used to remember the origin of the value.
type OriginFlag struct {
	origin ValueOrigin
}

// GetName return name of the Origin flag.
func (flag *OriginFlag) GetName() string {
	return OriginFlagName
}

// GetValue returns the value origin (as string).
func (flag *OriginFlag) GetValue() string {
	return flag.origin.String()
}

// ErrorFlag is used to mark base values that are in a failed state
// (or their derived values). It is used for KVScheduler.GetFailedValues(),
// also to inform user in the graph dump about currently failing values and
// finally for statistical purposes.
type ErrorFlag struct {
	err   error
	txnOp TxnOperation
}

// GetName return name of the Origin flag.
func (flag *ErrorFlag) GetName() string {
	return ErrorFlagName
}

// GetValue returns the error as string.
func (flag *ErrorFlag) GetValue() string {
	if flag.err == nil {
		return ""
	}
	return flag.err.Error()
}

// DescriptorFlag is used to lookup values by their descriptor.
type DescriptorFlag struct {
	descriptorName string
}

// GetName return name of the Descriptor flag.
func (flag *DescriptorFlag) GetName() string {
	return DescriptorFlagName
}

// GetValue returns the descriptor name.
func (flag *DescriptorFlag) GetValue() string {
	return flag.descriptorName
}

// DerivedFlag is used to mark derived values.
type DerivedFlag struct {
}

// GetName return name of the Derived flag.
func (flag *DerivedFlag) GetName() string {
	return DerivedFlagName
}

// GetValue return empty string (presence of the flag is the only information).
func (flag *DerivedFlag) GetValue() string {
	return ""
}
