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
	"context"
	"time"
)

type schedulerCtxKey int

const (
	// fullResyncCtxKey is a key under which *full-resync* txn option is stored
	// into the context.
	fullResyncCtxKey schedulerCtxKey = iota

	// downstreamResyncCtxKey is a key under which *downstream-resync* txn option is
	// stored into the context.
	downstreamResyncCtxKey

	// nonBlockingTxnCtxKey is a key under which *non-blocking* txn option is
	// stored into the context.
	nonBlockingTxnCtxKey

	// retryCtxKey is a key under which *retry* txn option is stored into
	// the context.
	retryCtxKey

	// revertCtxKey is a key under which *revert* txn option is stored into
	// the context.
	revertCtxKey

	// txnDescriptionKey is a key under which transaction description is stored
	// into the context.
	txnDescriptionKey
)

/* Full-Resync */

// fullResyncOpt represents the *full-resync* transaction option.
type fullResyncOpt struct {
	// no attributes
}

// WithFullResync prepares context for transaction carrying up-to-date *full*
// snapshot of NB key-value pairs that SB should be reconciled against.
// Such transaction should only carry non-NIL values - existing NB values
// not included in the transaction are automatically removed.
func WithFullResync(ctx context.Context) context.Context {
	return context.WithValue(ctx, fullResyncCtxKey, &fullResyncOpt{})
}

// IsFullResync returns true if the transaction context is configured
// to trigger full-resync.
func IsFullResync(ctx context.Context) bool {
	_, isFullResync := ctx.Value(fullResyncCtxKey).(*fullResyncOpt)
	return isFullResync
}

/* Downstream-Resync */

// downstreamResyncOpt represents the *downstream-resync* transaction option.
type downstreamResyncOpt struct {
	// no attributes
}

// WithDownstreamResync prepares context for transaction that will trigger resync
// between scheduler and SB - i.e. without NB providing up-to-date snapshot of
// key-value pairs, hence "downstream" reconciliation.
// Transaction is thus expected to carry no key-value pairs.
func WithDownstreamResync(ctx context.Context) context.Context {
	return context.WithValue(ctx, downstreamResyncCtxKey, &downstreamResyncOpt{})
}

// IsDownstreamResync returns true if the transaction context is configured
// to trigger downstream-resync.
func IsDownstreamResync(ctx context.Context) bool {
	_, isDownstreamResync := ctx.Value(downstreamResyncCtxKey).(*downstreamResyncOpt)
	return isDownstreamResync
}

/* Non-blocking Txn */

// nonBlockingTxnOpt represents the *non-blocking* transaction option.
type nonBlockingTxnOpt struct {
	// no attributes
}

// WithoutBlocking prepares context for transaction that should be scheduled
// for execution without blocking the caller of the Commit() method.
// By default, commit is blocking.
func WithoutBlocking(ctx context.Context) context.Context {
	return context.WithValue(ctx, nonBlockingTxnCtxKey, &nonBlockingTxnOpt{})
}

// IsNonBlockingTxn returns true if transaction context is configured for
// non-blocking Commit.
func IsNonBlockingTxn(ctx context.Context) bool {
	_, nonBlocking := ctx.Value(nonBlockingTxnCtxKey).(*nonBlockingTxnOpt)
	return nonBlocking
}

/* Retry */

// retryOpt represents the *retry* transaction option.
type retryOpt struct {
	period     time.Duration
	expBackoff bool
}

// WithRetry prepares context for transaction for which the scheduler will retry
// any (retriable) failed operations after given <period>. If <expBackoff>
// is enabled, every failed retry will double the next delay.
// Can be combined with revert - even failed revert operations will be re-tried.
// By default, the scheduler will not automatically retry failed operations.
func WithRetry(ctx context.Context, period time.Duration, expBackoff bool) context.Context {
	return context.WithValue(ctx, retryCtxKey, &retryOpt{
		period:     period,
		expBackoff: expBackoff,
	})
}

// IsWithRetry returns true if transaction context is configured to allow retry,
// including the option parameters, or zero values if retry is not enabled.
func IsWithRetry(ctx context.Context) (period time.Duration, expBackoff, withRetry bool) {
	retryArgs, withRetry := ctx.Value(retryCtxKey).(*retryOpt)
	if !withRetry {
		return 0, false, withRetry
	}
	return retryArgs.period, retryArgs.expBackoff, withRetry
}

/* Revert */

// revertOpt represents the *revert* transaction option.
type revertOpt struct {
	// no attributes
}

// WithRevert prepares context for transaction that will be reverted if any
// of its operations fails.
// By default, the scheduler executes transactions in a best-effort mode - even
// in the case of an error it will keep the effects of successful operations.
func WithRevert(ctx context.Context) context.Context {
	return context.WithValue(ctx, revertCtxKey, &revertOpt{})
}

// IsWithRevert returns true if the transaction context is configured
// to revert transaction if any of its operations fails.
func IsWithRevert(ctx context.Context) bool {
	_, isWithRevert := ctx.Value(revertCtxKey).(*revertOpt)
	return isWithRevert
}

/* Txn Description */

// txnDescriptionOpt represents the *txn-description* transaction option.
type txnDescriptionOpt struct {
	description string
}

// WithDescription prepares context for transaction that will have description
// provided.
// By default, transactions are without description.
func WithDescription(ctx context.Context, description string) context.Context {
	return context.WithValue(ctx, txnDescriptionKey, &txnDescriptionOpt{description: description})
}

// IsWithDescription returns true if the transaction context is configured
// to include transaction description.
func IsWithDescription(ctx context.Context) (description string, withDescription bool) {
	descriptionOpt, withDescription := ctx.Value(txnDescriptionKey).(*txnDescriptionOpt)
	if !withDescription {
		return "", false
	}
	return descriptionOpt.description, true
}
