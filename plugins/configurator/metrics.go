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

package configurator

import (
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
	stats.Operations = make(metrics.Calls)
}

func GetStats() *Stats {
	s := new(Stats)
	statsMu.RLock()
	*s = stats
	statsMu.RUnlock()
	return s
}

type Stats struct {
	AllOperations metrics.CallStats
	Operations    metrics.Calls
}

func (s *Stats) getOrCreateOperation(msg string) *metrics.CallStats {
	statsMu.RLock()
	ms, ok := s.Operations[msg]
	statsMu.RUnlock()
	if !ok {
		ms = &metrics.CallStats{Name: msg}
		statsMu.Lock()
		s.Operations[msg] = ms
		statsMu.Unlock()
	}
	return ms
}

func trackOperation(m string) func() {
	t := time.Now()
	ms := stats.getOrCreateOperation(m)
	return func() {
		took := time.Since(t)
		statsMu.Lock()
		ms.Increment(took)
		stats.AllOperations.Increment(took)
		statsMu.Unlock()
	}
}

func init() {
	expvar.Publish("configurator-stats", expvar.Func(func() interface{} {
		return GetStats()
	}))
}
