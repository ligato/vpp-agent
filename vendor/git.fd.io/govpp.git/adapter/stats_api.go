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

var (
	ErrStatDirBusy  = errors.New("stat dir busy")
	ErrStatDumpBusy = errors.New("stat dump busy")
)

var (
	// DefaultStatsSocket is the default path for the VPP stat socket file.
	DefaultStatsSocket = "/run/vpp/stats.sock"
)

// StatsAPI provides connection to VPP stats API.
type StatsAPI interface {
	// Connect establishes client connection to the stats API.
	Connect() error

	// Disconnect terminates client connection.
	Disconnect() error

	// ListStats lists names for all stats.
	ListStats(patterns ...string) (statNames []string, err error)

	// DumpStats dumps all stat entries.
	DumpStats(patterns ...string) ([]*StatEntry, error)
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

// StatEntry represents single stat entry. The type of stat stored in Data
// is defined by Type.
type StatEntry struct {
	Name string
	Type StatType
	Data Stat
}

// Counter represents simple counter with single value.
type Counter uint64

// CombinedCounter represents counter with two values, for packet count and bytes count.
type CombinedCounter struct {
	Packets Counter
	Bytes   Counter
}

// Name represents string value stored under name vector.
type Name string

// ScalarStat represents stat for ScalarIndex.
type ScalarStat float64

// ErrorStat represents stat for ErrorIndex.
type ErrorStat Counter

// SimpleCounterStat represents stat for SimpleCounterVector.
// The outer array represents workers and the inner array represents interface/node/.. indexes.
// Values should be aggregated per interface/node for every worker.
type SimpleCounterStat [][]Counter

// CombinedCounterStat represents stat for CombinedCounterVector.
// The outer array represents workers and the inner array represents interface/node/.. indexes.
// Values should be aggregated per interface/node for every worker.
type CombinedCounterStat [][]CombinedCounter

// NameStat represents stat for NameVector.
type NameStat []Name

// Data represents some type of stat which is usually defined by StatType.
type Stat interface {
	// isStat is unexported to limit implementations of Data interface to this package,
	isStat()
}

func (ScalarStat) isStat()          {}
func (ErrorStat) isStat()           {}
func (SimpleCounterStat) isStat()   {}
func (CombinedCounterStat) isStat() {}
func (NameStat) isStat()            {}
