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
	"math"
	"sort"
)

type Calls map[string]*CallStats

// MarshalJSON implements json.Marshaler interface
func (m Calls) MarshalJSON() ([]byte, error) {
	calls := make([]CallStats, 0, len(m))
	for _, s := range m {
		stat := *s
		stat.Total = round(stat.Total)
		stat.Avg = round(stat.Avg)
		stat.Min = round(stat.Min)
		stat.Max = round(stat.Max)
		calls = append(calls, stat)
	}
	sort.Slice(calls, func(i, j int) bool {
		return calls[i].Count > calls[j].Count
	})
	return json.Marshal(calls)
}

// CallStats represents generic stats for call metrics.
type CallStats struct {
	Name  string  `json:"name,omitempty"`
	Count uint64  `json:"count"`
	Total float64 `json:"total,omitempty"`
	Avg   float64 `json:"avg,omitempty"`
	Min   float64 `json:"min,omitempty"`
	Max   float64 `json:"max,omitempty"`
}

// Increment increments call count and recalculates durations
func (m *CallStats) Increment(tookSec float64) {
	m.Count++
	m.Total = m.Total + tookSec
	m.Avg = m.Total / float64(m.Count)
	if tookSec > m.Max {
		m.Max = tookSec
	}
	if m.Min == 0 || tookSec < m.Min {
		m.Min = tookSec
	}
}

func round(n float64) float64 {
	return math.Round(n*1000) / 1000
}
