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

package govppmux

import (
	"expvar"
	"os"
	"sync"
	"time"

	"go.ligato.io/vpp-agent/v3/pkg/metrics"
	"go.ligato.io/vpp-agent/v3/proto/ligato/govppmux"
)

// DisableOldStats is used to disabled old way of collecting stats.
var DisableOldStats = os.Getenv("GOVPPMUX_OLDSTATS_DISABLED") != ""

var (
	stats   Stats
	statsMu sync.RWMutex
)

func init() {
	if DisableOldStats {
		return
	}
	stats.Errors = make(metrics.Calls)
	stats.Messages = make(metrics.Calls)
	stats.Replies = make(metrics.Calls)
	metrics.Register(&govppmux.Metrics{}, func() interface{} {
		return &GetStats().Metrics
	})
	expvar.Publish("govppmux.stats", expvar.Func(func() interface{} {
		return &GetStats().Metrics
	}))
}

func GetStats() *Stats {
	if DisableOldStats {
		return nil
	}
	s := new(Stats)
	statsMu.RLock()
	*s = stats
	statsMu.RUnlock()
	return s
}

// Stats defines various statistics for govppmux plugin.
type Stats struct {
	govppmux.Metrics

	Errors metrics.Calls

	AllMessages metrics.CallStats
	Messages    metrics.Calls

	Replies metrics.Calls
}

func (s *Stats) getOrCreateMessage(msg string) *metrics.CallStats {
	statsMu.RLock()
	ms, ok := s.Messages[msg]
	statsMu.RUnlock()
	if !ok {
		ms = &metrics.CallStats{Name: msg}
		statsMu.Lock()
		s.Messages[msg] = ms
		statsMu.Unlock()
	}
	return ms
}

func trackMsgRequestDur(m string, d time.Duration) {
	ms := stats.getOrCreateMessage(m)
	statsMu.Lock()
	ms.Increment(d)
	stats.AllMessages.Increment(d)
	statsMu.Unlock()
}

func (s *Stats) getOrCreateReply(msg string) *metrics.CallStats {
	statsMu.RLock()
	ms, ok := s.Replies[msg]
	statsMu.RUnlock()
	if !ok {
		ms = &metrics.CallStats{Name: msg}
		statsMu.Lock()
		s.Replies[msg] = ms
		statsMu.Unlock()
	}
	return ms
}

func trackMsgReply(m string) {
	ms := stats.getOrCreateReply(m)
	statsMu.Lock()
	ms.Increment(0)
	statsMu.Unlock()
}

func (s *Stats) getOrCreateError(msg string) *metrics.CallStats {
	statsMu.RLock()
	ms, ok := s.Errors[msg]
	statsMu.RUnlock()
	if !ok {
		ms = &metrics.CallStats{Name: msg}
		statsMu.Lock()
		s.Errors[msg] = ms
		statsMu.Unlock()
	}
	return ms
}

func trackError(m string) {
	ms := stats.getOrCreateError(m)
	statsMu.Lock()
	ms.Increment(0)
	statsMu.Unlock()
}
