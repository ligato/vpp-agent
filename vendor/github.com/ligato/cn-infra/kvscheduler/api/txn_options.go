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

import "time"

// TODO: move to the localclient

// TxnOption configures NB transaction.
// The available options can be found below.
type TxnOption interface {
}

// NonBlockingTxn implements the *non-blocking* transaction option.
type NonBlockingTxn struct {
}

// WithoutBlocking returns transaction option which causes the transaction
// to be scheduled for execution, but otherwise not blocking the caller of the
// Commit() method.
// By default, commit is blocking.
func WithoutBlocking() TxnOption {
	return &NonBlockingTxn{}
}

// RetryFailedOps implements the *retry* transaction option.
type RetryFailedOps struct {
	Period     time.Duration
	ExpBackoff bool
}

// WithRetry returns transaction option which will tell the scheduler to retry
// failed operations from the transaction after given <period>. If <ExpBackoff>
// is enabled, every failed retry will double the next period.
// Can be combined with revert - even failed revert operations will be re-tried.
// By default, the scheduler will not automatically retry failed operations.
func WithRetry(period time.Duration, expBackoff bool) TxnOption {
	return &RetryFailedOps{Period: period, ExpBackoff: expBackoff}
}

// RevertOnFailure implements the *revert* transaction option.
type RevertOnFailure struct {
}

// WithRevert returns transaction option that will cause the transaction to be
// reverted if any of its operations fails.
// By default, the scheduler executes transactions in a best-effort mode - even
// in the case of an error it will keep the effects of successful operations.
func WithRevert() TxnOption {
	return &RevertOnFailure{}
}
