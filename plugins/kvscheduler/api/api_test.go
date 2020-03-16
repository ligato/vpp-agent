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

package api_test

import (
	"encoding/json"
	"testing"

	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

func TestTxnTypeEncode(t *testing.T) {
	tests := []struct {
		name string

		txntype api.TxnType

		expectOut string
		expectErr error
	}{
		{"SBNotification", api.SBNotification, `"SBNotification"`, nil},
		{"NBTransaction", api.NBTransaction, `"NBTransaction"`, nil},
		{"RetryFailedOps", api.RetryFailedOps, `"RetryFailedOps"`, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, err := json.Marshal(test.txntype)
			if err != test.expectErr {
				t.Fatalf("expected error: %v, got %v", test.expectErr, err)
			}
			out := string(b)
			if out != test.expectOut {
				t.Fatalf("expected output: %q, got %q", test.expectOut, out)
			}
		})
	}
}

func TestTxnTypeDecode(t *testing.T) {
	tests := []struct {
		name string

		input string

		expectTxnType api.TxnType
		expectErr     error
	}{
		{"RetryFailedOps", `"RetryFailedOps"`, api.RetryFailedOps, nil},
		{"NBTransaction", `"NBTransaction"`, api.NBTransaction, nil},
		{"1 (NBTransaction)", `1`, api.NBTransaction, nil},
		{"0 (SBNotification)", `0`, api.SBNotification, nil},
		{"invalid", `"INVALID"`, api.TxnType(-1), nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var txntype api.TxnType
			err := json.Unmarshal([]byte(test.input), &txntype)
			if err != test.expectErr {
				t.Fatalf("expected error: %v, got %v", test.expectErr, err)
			}
			if txntype != test.expectTxnType {
				t.Fatalf("expected TxnType: %v, got %v", test.expectTxnType, txntype)
			}
		})
	}
}

func TestResyncTypeEncode(t *testing.T) {
	tests := []struct {
		name string

		resynctype api.ResyncType

		expectOut string
		expectErr error
	}{
		{"FullResync", api.FullResync, `"FullResync"`, nil},
		{"UpstreamResync", api.UpstreamResync, `"UpstreamResync"`, nil},
		{"DownstreamResync", api.DownstreamResync, `"DownstreamResync"`, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, err := json.Marshal(test.resynctype)
			if err != test.expectErr {
				t.Fatalf("expected error: %v, got %v", test.expectErr, err)
			}
			out := string(b)
			if out != test.expectOut {
				t.Fatalf("expected output: %q, got %q", test.expectOut, out)
			}
		})
	}
}

func TestResyncTypeDecode(t *testing.T) {
	tests := []struct {
		name string

		input string

		expectResyncType api.ResyncType
		expectErr        error
	}{
		{"FullResync", `"FullResync"`, api.FullResync, nil},
		{"UpstreamResync", `"UpstreamResync"`, api.UpstreamResync, nil},
		{"DownstreamResync", `"DownstreamResync"`, api.DownstreamResync, nil},
		{"1 (FullResync)", `1`, api.FullResync, nil},
		{"2 (UpstreamResync)", `2`, api.UpstreamResync, nil},
		{"3 (DownstreamResync)", `3`, api.DownstreamResync, nil},
		{"invalid", `"INVALID"`, api.ResyncType(0), nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var resyncType api.ResyncType
			err := json.Unmarshal([]byte(test.input), &resyncType)
			if err != test.expectErr {
				t.Fatalf("expected error: %v, got %v", test.expectErr, err)
			}
			if resyncType != test.expectResyncType {
				t.Fatalf("expected ResyncType: %v, got %v", test.expectResyncType, resyncType)
			}
		})
	}
}
