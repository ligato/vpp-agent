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

package metrics

import (
	"encoding/json"
	"sort"
	"time"
)

// RoundDuration is the default value used for rounding durations.
var RoundDuration = time.Microsecond * 10

type Calls map[string]*CallStats

// MarshalJSON implements json.Marshaler interface
func (m Calls) MarshalJSON() ([]byte, error) {
	calls := make([]*CallStats, 0, len(m))
	for _, s := range m {
		calls = append(calls, s)
	}
	sort.Slice(calls, func(i, j int) bool {
		return calls[i].Total > calls[j].Total
	})
	return json.Marshal(calls)
}

// CallStats represents generic stats for call metrics.
type CallStats struct {
	Name  string `json:",omitempty"`
	Count uint64
	Total Duration
	Avg   Duration
	Min   Duration
	Max   Duration
}

// Increment increments call count and recalculates durations
func (m *CallStats) Increment(d time.Duration) {
	took := Duration(d)
	m.Count++
	m.Total += took
	m.Avg = m.Total / Duration(m.Count)
	if took > m.Max {
		m.Max = took
	}
	if m.Min == 0 || took < m.Min {
		m.Min = took
	}
}

/*
// MarshalJSON implements json.Marshaler interface
func (m *CallStats) MarshalJSON() ([]byte, error) {
	var d string
	d = fmt.Sprintf(
		"count: %d, total: %s (avg/min/max: %s/%s/%s)",
		m.Count, durStr(m.TotalDur),
		durStr(m.AvgDur), durStr(m.MinDur), durStr(m.MaxDur),
	)
	return json.Marshal(d)
}
*/

type Duration time.Duration

// MarshalJSON implements json.Marshaler interface
func (m *Duration) MarshalJSON() ([]byte, error) {
	s := time.Duration(*m).Round(RoundDuration).String()
	return json.Marshal(s)
}
