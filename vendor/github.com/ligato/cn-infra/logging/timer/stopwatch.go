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
	// start time can be set as the beginning of the measurement to calculate overall time
	Overall time.Duration
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
		// Default value
		Overall:  -1,
		name: name,
		logger: log,
		timeTable: make(map[string]time.Duration),
	}
}

// LogTimeEntry stores name of the binapi call and measured duration
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
		st.logger.WithField("plugin", st.name).Infof("Stopwatch: no entries")
	}
	for k, v := range st.timeTable {
		st.logger.WithField("plugin", st.name).Infof("Calling %v took %v", k, v)
	}
	if st.Overall != -1 {
		st.logger.WithField("plugin", st.name).Infof("Resync took %v", st.Overall)
	}
	// clear map after use
	st.timeTable = make(map[string]time.Duration)
}
