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
)

var (
	stats   Stats
	statsMu sync.RWMutex
)

func init() {
	stats.GraphMethods.Methods = make(metrics.Calls)
	stats.AllDescriptors.Methods = make(metrics.Calls)
	stats.Descriptors = make(map[string]*StructStats)
}

/*func GetDescriptorStats() map[string]metrics.Calls {
	ss := make(map[string]metrics.Calls, len(stats.Descriptors))
	statsMu.RLock()
	for d, ds := range stats.Descriptors {
		cc := make(metrics.Calls, len(ds))
		for c, cs := range ds {
			css := *cs
			cc[c] = &css
		}
		ss[d] = cc
	}
	statsMu.RUnlock()
	return ss
}*/

/*func GetGraphStats() *metrics.CallStats {
	s := make(metrics.Calls, len(stats.Descriptors))
	statsMu.RLock()
	*s = stats.Graph
	statsMu.RUnlock()
	return s
}*/

func GetStats() *Stats {
	s := new(Stats)
	statsMu.RLock()
	*s = stats
	statsMu.RUnlock()
	return s
}

type Stats struct {
	TransactionsProcessed uint64

	GraphMethods   StructStats
	AllDescriptors StructStats
	Descriptors    map[string]*StructStats
}

func (s *Stats) addDescriptor(name string) {
	s.Descriptors[name] = &StructStats{
		Methods: make(metrics.Calls),
	}
}

type StructStats struct {
	Methods metrics.Calls `json:"-,omitempty"`
}

func (s *StructStats) MarshalJSON() ([]byte, error) {
	/*d := make(map[string]*metrics.CallStats, len(s.Methods))
	for _, ms := range s.Methods {
		m := fmt.Sprintf("%s()", ms.Name)
		d[m] = ms
	}*/
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

func init() {
	expvar.Publish("kvscheduler", expvar.Func(func() interface{} {
		return GetStats()
	}))
}
