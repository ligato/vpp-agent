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
	"fmt"
	"sync"
	"time"
)

var (
	stats        Stats
	descriptorMu sync.RWMutex
)

func init() {
	stats.Descriptors = map[string]*DescriptorStats{
		"ALL": {},
	}
}

func GetStats() *Stats {
	s := new(Stats)
	descriptorMu.RLock()
	*s = stats
	descriptorMu.RUnlock()
	return s
}

type Stats struct {
	TransactionsProcessed uint64

	Descriptors map[string]*DescriptorStats
}

func (s *Stats) addDescriptor(name string) {
	s.Descriptors[name] = &DescriptorStats{}
}

func GetDescriptorStats() map[string]*DescriptorStats {
	s := map[string]*DescriptorStats{}
	descriptorMu.RLock()
	for d, ds := range stats.Descriptors {
		dss := *ds
		s[d] = &dss
	}
	descriptorMu.RUnlock()
	return s
}

type DescriptorStats struct {
	Methods []*MethodStats
}

type MethodStats struct {
	Method  string
	Calls   uint64
	TotalNs time.Duration
	AvgNs   time.Duration
	MaxNs   time.Duration
}

func dur(d time.Duration) string {
	return d.Round(time.Microsecond * 100).String()
}

func (s *DescriptorStats) MarshalJSON() ([]byte, error) {
	d := map[string]string{}
	for _, ms := range s.Methods {
		m := fmt.Sprintf("%s()", ms.Method)
		d[m] = fmt.Sprintf(
			"calls: %d, total: %s, avg: %s, max: %s",
			ms.Calls, dur(ms.TotalNs), dur(ms.AvgNs), dur(ms.MaxNs),
		)
	}
	return json.Marshal(d)
}

func (s *DescriptorStats) getOrCreateMethod(method string) *MethodStats {
	descriptorMu.RLock()
	for _, m := range s.Methods {
		if m.Method == method {
			descriptorMu.RUnlock()
			return m
		}
	}
	descriptorMu.RUnlock()
	ms := &MethodStats{Method: method}
	descriptorMu.Lock()
	s.Methods = append(s.Methods, ms)
	descriptorMu.Unlock()
	return ms
}

func (m *MethodStats) increment(took time.Duration) {
	descriptorMu.Lock()
	m.Calls++
	m.TotalNs += took
	m.AvgNs = m.TotalNs / time.Duration(m.Calls)
	if took > m.MaxNs {
		m.MaxNs = took
	}
	descriptorMu.Unlock()
}

func trackDescMethod(d, m string) func() {
	t := time.Now()
	method := stats.Descriptors[d].getOrCreateMethod(m)
	methodall := stats.Descriptors["ALL"].getOrCreateMethod(m)
	return func() {
		took := time.Since(t)
		method.increment(took)
		methodall.increment(took)
	}
}

func init() {
	expvar.Publish("kvdescriptors", expvar.Func(func() interface{} {
		return GetDescriptorStats()
	}))
}
