// Copyright (c) 2019 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adapter

import (
	"errors"
	"fmt"
)

const (
	// DefaultStatsSocket defines a default socket file path for VPP stats API.
	DefaultStatsSocket = "/run/vpp/stats.sock"
)

var (
	ErrStatsDataBusy     = errors.New("stats data busy")
	ErrStatsDirStale     = errors.New("stats dir stale")
	ErrStatsAccessFailed = errors.New("stats access failed")
)

// StatsAPI provides connection to VPP stats API.
type StatsAPI interface {
	// Connect establishes client connection to the stats API.
	Connect() error
	// Disconnect terminates client connection.
	Disconnect() error

	// ListStats lists names for stats matching patterns.
	ListStats(patterns ...string) (names []string, err error)
	// DumpStats dumps all stat entries.
	DumpStats(patterns ...string) (entries []StatEntry, err error)

	// PrepareDir prepares new stat dir for entries that match any of prefixes.
	PrepareDir(patterns ...string) (*StatDir, error)
	// UpdateDir updates stat dir and all of their entries.
	UpdateDir(dir *StatDir) error
}

// StatType represents type of stat directory and simply
// defines what type of stat data is stored in the stat entry.
type StatType int

const (
	_                     StatType = 0
	ScalarIndex           StatType = 1
	SimpleCounterVector   StatType = 2
	CombinedCounterVector StatType = 3
	ErrorIndex            StatType = 4
	NameVector            StatType = 5
)

func (d StatType) String() string {
	switch d {
	case ScalarIndex:
		return "ScalarIndex"
	case SimpleCounterVector:
		return "SimpleCounterVector"
	case CombinedCounterVector:
		return "CombinedCounterVector"
	case ErrorIndex:
		return "ErrorIndex"
	case NameVector:
		return "NameVector"
	}
	return fmt.Sprintf("UnknownStatType(%d)", d)
}

// StatDir defines directory of stats entries created by PrepareDir.
type StatDir struct {
	Epoch   int64
	Indexes []uint32
	Entries []StatEntry
}

// StatEntry represents single stat entry. The type of stat stored in Data
// is defined by Type.
type StatEntry struct {
	Name []byte
	Type StatType
	Data Stat
}

// Counter represents simple counter with single value, which is usually packet count.
type Counter uint64

// CombinedCounter represents counter with two values, for packet count and bytes count.
type CombinedCounter [2]uint64

func (s CombinedCounter) Packets() uint64 {
	return uint64(s[0])
}

func (s CombinedCounter) Bytes() uint64 {
	return uint64(s[1])
}

// Name represents string value stored under name vector.
type Name []byte

func (n Name) String() string {
	return string(n)
}

// Stat represents some type of stat which is usually defined by StatType.
type Stat interface {
	// IsZero returns true if all of its values equal to zero.
	IsZero() bool

	// isStat is intentionally  unexported to limit implementations of interface to this package,
	isStat()
}

// ScalarStat represents stat for ScalarIndex.
type ScalarStat float64

// ErrorStat represents stat for ErrorIndex.
type ErrorStat Counter

// SimpleCounterStat represents stat for SimpleCounterVector.
// The outer array represents workers and the inner array represents interface/node/.. indexes.
// Values should be aggregated per interface/node for every worker.
// ReduceSimpleCounterStatIndex can be used to reduce specific index.
type SimpleCounterStat [][]Counter

// CombinedCounterStat represents stat for CombinedCounterVector.
// The outer array represents workers and the inner array represents interface/node/.. indexes.
// Values should be aggregated per interface/node for every worker.
// ReduceCombinedCounterStatIndex can be used to reduce specific index.
type CombinedCounterStat [][]CombinedCounter

// NameStat represents stat for NameVector.
type NameStat []Name

func (ScalarStat) isStat()          {}
func (ErrorStat) isStat()           {}
func (SimpleCounterStat) isStat()   {}
func (CombinedCounterStat) isStat() {}
func (NameStat) isStat()            {}

func (s ScalarStat) IsZero() bool {
	return s == 0
}
func (s ErrorStat) IsZero() bool {
	return s == 0
}
func (s SimpleCounterStat) IsZero() bool {
	if s == nil {
		return true
	}
	for _, ss := range s {
		for _, sss := range ss {
			if sss != 0 {
				return false
			}
		}
	}
	return true
}
func (s CombinedCounterStat) IsZero() bool {
	if s == nil {
		return true
	}
	for _, ss := range s {
		if ss == nil {
			return true
		}
		for _, sss := range ss {
			if sss[0] != 0 || sss[1] != 0 {
				return false
			}
		}
	}
	return true
}
func (s NameStat) IsZero() bool {
	if s == nil {
		return true
	}
	for _, ss := range s {
		if len(ss) > 0 {
			return false
		}
	}
	return true
}

// ReduceSimpleCounterStatIndex returns reduced SimpleCounterStat s for index i.
func ReduceSimpleCounterStatIndex(s SimpleCounterStat, i int) uint64 {
	var val uint64
	for _, w := range s {
		val += uint64(w[i])
	}
	return val
}

// ReduceCombinedCounterStatIndex returns reduced CombinedCounterStat s for index i.
func ReduceCombinedCounterStatIndex(s CombinedCounterStat, i int) [2]uint64 {
	var val [2]uint64
	for _, w := range s {
		val[0] += uint64(w[i][0])
		val[1] += uint64(w[i][1])
	}
	return val
}
