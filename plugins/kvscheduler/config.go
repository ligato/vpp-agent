//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package kvscheduler

const (
	// by default, a history of processed transaction is recorded
	defaultRecordTransactionHistory = true

	// by default, only transaction processed in the last 24 hours are kept recorded
	// (with the exception of permanently recorded init period)
	defaultTransactionHistoryAgeLimit = 24 * 60 // in minutes

	// by default, transactions from the first hour of runtime stay permanently
	// recorded
	defaultPermanentlyRecordedInitPeriod = 60 // in minutes

	// by default, all NB transactions and SB notifications are run without
	// simulation (Retries are always first simulated)
	defaultEnableTxnSimulation = false

	// by default, a concise summary of every processed transactions is printed
	// to stdout
	defaultPrintTxnSummary = true
)

func DefaultConfig() *Config {
	return &Config{
		RecordTransactionHistory:      defaultRecordTransactionHistory,
		TransactionHistoryAgeLimit:    defaultTransactionHistoryAgeLimit,
		PermanentlyRecordedInitPeriod: defaultPermanentlyRecordedInitPeriod,
		EnableTxnSimulation:           defaultEnableTxnSimulation,
		PrintTxnSummary:               defaultPrintTxnSummary,
	}
}

// Config holds the KVScheduler configuration.
type Config struct {
	RecordTransactionHistory      bool   `json:"record-transaction-history"`
	TransactionHistoryAgeLimit    uint32 `json:"transaction-history-age-limit"`    // in minutes
	PermanentlyRecordedInitPeriod uint32 `json:"permanently-recorded-init-period"` // in minutes
	EnableTxnSimulation           bool   `json:"enable-txn-simulation"`
	PrintTxnSummary               bool   `json:"print-txn-summary"`
}
