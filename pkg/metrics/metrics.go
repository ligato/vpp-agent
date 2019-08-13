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
	"time"
)

// RoundDuration is the default value used for rounding durations.
var RoundDuration = time.Millisecond * 1

type Calls map[string]*CallStats

// MarshalJSON implements json.Marshaler interface
func (m Calls) MarshalJSON() ([]byte, error) {
	calls := make([]*CallStats, 0, len(m))
	for _, s := range m {
		calls = append(calls, s)
	}
	sort.Slice(calls, func(i, j int) bool {
		return calls[i].Count > calls[j].Count
	})

	/*var buf bytes.Buffer
	buf.WriteByte('{')
	for i, c := range calls {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(fmt.Sprintf(`"%s":{"count":%d}`, c.Name, c.Count))
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil*/

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
func (m *CallStats) Increment(d time.Duration) {
	took := d.Round(RoundDuration).Seconds()
	m.Count++
	m.Total = round(m.Total + took)
	m.Avg = round(m.Total / float64(m.Count))
	if took > m.Max {
		m.Max = took
	}
	if m.Min == 0 || took < m.Min {
		m.Min = took
	}
}

func round(n float64) float64 {
	return math.Round(n*1000) / 1000
}

// MarshalJSON implements json.Marshaler interface
/*func (m *CallStats) MarshalJSON() ([]byte, error) {
	var d string
	d = fmt.Sprintf(
		"%s - count: %d, total: %s (avg/min/max: %s/%s/%s)",
		m.Name, m.Count, durStr(m.Total),
		durStr(m.Avg), durStr(m.Min), durStr(m.Max),
	)
	return json.Marshal(d)
}*/
