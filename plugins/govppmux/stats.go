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
	"encoding/json"
	"expvar"
	"fmt"
	"sync"
	"time"
)

var (
	stats     Stats
	messageMu sync.RWMutex
)

func init() {
	stats.Messages = map[string]*MessageStats{}
}

func GetStats() *Stats {
	s := new(Stats)
	*s = stats
	return s
}

type Stats struct {
	ChannelsCreated uint64
	ChannelsOpen    uint64
	RequestsSent    uint64
	RequestsFailed  uint64
	Messages        map[string]*MessageStats
}

type MessageStats struct {
	Message string
	Calls   uint64
	TotalNs time.Duration
	AvgNs   time.Duration
	MaxNs   time.Duration
}

func dur(d time.Duration) string {
	return d.Round(time.Microsecond * 100).String()
}

func (m *MessageStats) MarshalJSON() ([]byte, error) {
	d := fmt.Sprintf(
		"calls: %d, total: %s, avg: %s, max: %s",
		m.Calls, dur(m.TotalNs), dur(m.AvgNs), dur(m.MaxNs),
	)
	return json.Marshal(d)
}

func (s *Stats) getOrCreateMessage(msg string) *MessageStats {
	messageMu.RLock()
	ms, ok := s.Messages[msg]
	messageMu.RUnlock()
	if !ok {
		ms = &MessageStats{Message: msg}
		messageMu.Lock()
		s.Messages[msg] = ms
		messageMu.Unlock()
	}
	return ms
}

func (m *MessageStats) increment(took time.Duration) {
	m.Calls++
	m.TotalNs += took
	m.AvgNs = m.TotalNs / time.Duration(m.Calls)
	if took > m.MaxNs {
		m.MaxNs = took
	}
}

func trackMsgRequestDur(m string, d time.Duration) {
	ms := stats.getOrCreateMessage(m)
	mall := stats.getOrCreateMessage("ALL")
	messageMu.Lock()
	ms.increment(d)
	mall.increment(d)
	messageMu.Unlock()
}

func init() {
	expvar.Publish("govppstats", expvar.Func(func() interface{} {
		return GetStats()
	}))
}
