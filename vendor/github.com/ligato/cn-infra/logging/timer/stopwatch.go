// Copyright (c) 2017 Cisco and/or its affiliates.
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

package timer

import (
	"reflect"
	"github.com/ligato/cn-infra/logging"
	"time"
	"strconv"
	"sync"
)

// Stopwatch keeps all time measurement results
type Stopwatch struct {
	// name of the entity/plugin
	name string
	// logger used while printing
	logger logging.Logger
	// map where measurements are stored
	timeTable map[string]time.Duration
	// used to lock map
	mx sync.Mutex
}

// NewStopwatch creates a new stopwatch object with empty time map
func NewStopwatch(name string, log logging.Logger) *Stopwatch {
	return &Stopwatch{
		name: name,
		logger: log,
		timeTable: make(map[string]time.Duration),
	}
}

// LogTimeEntry stores name of the binapi call and measured duration
// <n>
// <d>
func (st *Stopwatch) LogTimeEntry(n interface{}, d time.Duration) {
	st.mx.Lock()
	defer st.mx.Unlock()

	var name string
	switch nType := n.(type) {
	case string:
		name = nType
	default:
		name = reflect.TypeOf(n).String()
	}
	// index multiple occurrences of the same name (bin_api, link)
	_, found := st.timeTable[name]
	if found {
		index := 1
		for {
			indexed := name + "#" + strconv.Itoa(index)
			_, found = st.timeTable[indexed]
			if found {
				index++
				continue
			}
			name = indexed
			break
		}
	}
	// Store time value
	st.timeTable[name] = d
}

// Print logs all entries from the map (partial times) + overall time if set
func (st *Stopwatch) Print() {
	if len(st.timeTable) == 0 {
		st.logger.WithField("conf", st.name).Infof("stopwatch has no entries")
	}
	var overall time.Duration
	for k, v := range st.timeTable {
		overall += v
		st.logger.WithFields(logging.Fields{"conf": st.name, "durationInNs": v.Nanoseconds()}).Infof("calling %v took %v", k, v)
	}
	st.logger.WithFields(logging.Fields{"conf": st.name, "durationInNs": overall.Nanoseconds()}).Infof("partial resync time is %v", overall)
	// clear map after use
	st.timeTable = make(map[string]time.Duration)
}
