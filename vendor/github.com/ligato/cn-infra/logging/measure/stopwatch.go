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

package measure

import (
	"reflect"
	"github.com/ligato/cn-infra/logging"
	"time"
	"sync"
	"fmt"
)

// Stopwatch keeps all time measurement results
type Stopwatch struct {
	// name of the entity/plugin
	name string
	// logger used while printing
	logger logging.Logger
	// map where measurements are stored. Map is in format [string][]Duration, for every binapi/netlink api there is
	// a set of times this binapi/netlink was called
	timeTable sync.Map
	// used to lock map
	mx sync.Mutex
}

// NewStopwatch creates a new stopwatch object with empty time map
func NewStopwatch(name string, log logging.Logger) *Stopwatch {
	return &Stopwatch{
		name: name,
		logger: log,
		timeTable: sync.Map{},
	}
}

// LogTimeEntry stores name of the binapi call and measured duration
// <n> is a name of the measured entity (bin_api call object name or any other string)
// <d> is measured time
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
	// look for entries with the same name
	v, found := st.timeTable.Load(name)
	if found {
		durations, ok := v.([]time.Duration)
		if !ok {
			panic("cannot cast timeTable map value to duration")
		}
		st.timeTable.Store(name, append(durations, d))
	} else {
		// Store first time value for the specific key
		durations := make([]time.Duration, 0)
		st.timeTable.Store(name, append(durations, d))
	}
}

// Print logs all entries from the map (partial times) + overall time if set
func (st *Stopwatch) Print() {
	isMapEmpty := true
	var wasErr error
	// Calculate overall time
	var overall time.Duration
	st.timeTable.Range(func(k, v interface{}) bool {
		// Remember that the map contained entries
		isMapEmpty = false
		key, ok := k.(string)
		if !ok {
			wasErr = fmt.Errorf("cannot cast timeTable map key to string")
			// stops the iteration
			return false
		}
		value, ok := v.([]time.Duration)
		if !ok {
			wasErr = fmt.Errorf("cannot cast timeTable map value to duration")
			// stops the iteration
			return false
		}
		// Print time value for every and calculate overall time
		for _, entry := range value {
			overall += entry
			st.logger.WithFields(logging.Fields{"conf": st.name, "durationInNs": entry.Nanoseconds()}).Infof("%v call took %v", key, entry)
		}
		return true
	})

	// throw panic outside of logger.Range()
	if wasErr != nil {
		panic(wasErr)
	}

	// In case map is entry
	if isMapEmpty {
		st.logger.WithField("conf", st.name).Infof("stopwatch has no entries")
	}
	// Log overall time
	st.logger.WithFields(logging.Fields{"conf": st.name, "durationInNs": overall.Nanoseconds()}).Infof("partial resync time is %v", overall)

	// clear map after use
	st.timeTable = sync.Map{}
}
