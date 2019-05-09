//  Copyright (c) 2019 Cisco and/or its affiliates.
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

import (
	"encoding/json"
	"expvar"
	"sync"
	"time"

	"github.com/ligato/vpp-agent/pkg/metrics"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

var (
	stats   Stats
	statsMu sync.RWMutex
)

func init() {
	stats.GraphMethods.Methods = make(metrics.Calls)
	stats.AllDescriptors.Methods = make(metrics.Calls)
	stats.Descriptors = make(map[string]*StructStats)
	stats.TxnStats.Methods = make(metrics.Calls)
	stats.TxnStats.OperationCount = make(map[string]uint64)
	stats.TxnStats.ValueStateCount = make(map[string]uint64)
	for state := range kvs.ValueState_value {
		stats.TxnStats.ValueStateCount[state] = 0
	}
	for op, opVal := range kvs.TxnOperation_name {
		if op == int32(kvs.TxnOperation_UNDEFINED) ||
			op == int32(kvs.TxnOperation_VALIDATE) {
			continue
		}
		stats.TxnStats.OperationCount[opVal] = 0
	}
}

func GetStats() *Stats {
	s := new(Stats)
	statsMu.RLock()
	*s = stats
	statsMu.RUnlock()
	return s
}

type Stats struct {
	TxnStats       TxnStats
	GraphMethods   StructStats
	AllDescriptors StructStats
	Descriptors    map[string]*StructStats
}

func (s *Stats) addDescriptor(name string) {
	s.Descriptors[name] = &StructStats{
		Methods: make(metrics.Calls),
	}
}

type TxnStats struct {
	TotalProcessed  uint64
	OperationCount  map[string]uint64
	ValueStateCount map[string]uint64
	ErrorCount      uint64
	Methods         metrics.Calls
}

type StructStats struct {
	Methods metrics.Calls `json:"-,omitempty"`
}

func (s *StructStats) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Methods)
}

func (s *StructStats) getOrCreateMethod(method string) *metrics.CallStats {
	statsMu.RLock()
	ms, ok := s.Methods[method]
	statsMu.RUnlock()
	if !ok {
		ms = &metrics.CallStats{Name: method}
		statsMu.Lock()
		s.Methods[method] = ms
		statsMu.Unlock()
	}
	return ms
}

func trackDescMethod(d, m string) func() {
	t := time.Now()
	method := stats.Descriptors[d].getOrCreateMethod(m)
	methodall := stats.AllDescriptors.getOrCreateMethod(m)
	return func() {
		took := time.Since(t)
		statsMu.Lock()
		method.Increment(took)
		methodall.Increment(took)
		statsMu.Unlock()
	}
}

func trackGraphMethod(m string) func() {
	t := time.Now()
	method := stats.GraphMethods.getOrCreateMethod(m)
	return func() {
		took := time.Since(t)
		statsMu.Lock()
		method.Increment(took)
		statsMu.Unlock()
	}
}

func trackTransactionMethod(m string) func() {
	t := time.Now()
	s := stats.TxnStats
	return func() {
		took := time.Since(t)
		statsMu.Lock()
		ms, tracked := s.Methods[m]
		if !tracked {
			ms = &metrics.CallStats{Name: m}
			s.Methods[m] = ms
		}
		ms.Increment(took)
		statsMu.Unlock()
	}
}

func updateTransactionStats(execOps kvs.RecordedTxnOps) {
	statsMu.Lock()
	defer statsMu.Unlock()
	stats.TxnStats.TotalProcessed++
	for _, op := range execOps {
		if op.NewErr != nil {
			stats.TxnStats.ErrorCount++
		}
		stats.TxnStats.OperationCount[op.Operation.String()]++
		stats.TxnStats.ValueStateCount[op.NewState.String()]++
	}
}

func init() {
	expvar.Publish("kvscheduler", expvar.Func(func() interface{} {
		return GetStats()
	}))
}
