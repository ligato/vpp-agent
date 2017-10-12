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
	"strconv"
)

// StopWatchEntry provides method to log measured time entries
type StopWatchEntry interface {
	LogTimeEntry(d time.Duration)
}

// Stopwatch keeps all time measurement results
type Stopwatch struct {
	// name of the entity/plugin
	name string
	// logger used while printing
	logger logging.Logger
	// map where measurements are stored. Map is in format [string]TimeLog (string is a name related to the measured time(s)
	// which are stored in timelog), for every binapi/netlink api there is
	// a set of times this binapi/netlink was called
	timeTable sync.Map
}

// NewStopwatch creates a new stopwatch object with empty time map
func NewStopwatch(name string, log logging.Logger) *Stopwatch {
	return &Stopwatch{
		name: name,
		logger: log,
		timeTable: sync.Map{},
	}
}

// TimeLog is a wrapper for the measured data for specific name
type TimeLog struct {
	entries []time.Duration
}

// GetTimeLog returns a pointer to the TimeLog object related to the provided name (derived from the <n> parameter).
// If stopwatch is not used, returns nil
func GetTimeLog(n interface{}, s *Stopwatch) *TimeLog {
	// return nil if does not exist
	if s == nil {
		return nil
	}
	return s.timeLog(n)
}

// looks over stopwatch timeTable map in order to find a TimeLog object for provided name. If the object does not exist,
// it is created anew, stored in the map and returned
func (st *Stopwatch) timeLog(n interface{}) *TimeLog {
	// derive name
	var name string
	switch nType := n.(type) {
	case string:
		name = nType
	default:
		name = reflect.TypeOf(n).String()
	}
	// create and initialize new TimeLog in case it does not exist
	timer := &TimeLog{}
	timer.entries = make([]time.Duration, 0)
	// if there is no TimeLog under the name, store the created one. Otherwise, existing timer is returned
	existingTimer, loaded := st.timeTable.LoadOrStore(name, timer)
	if loaded {
		// cast to object which can be returned
		existing, ok := existingTimer.(*TimeLog)
		if !ok {
			panic(fmt.Errorf("cannot cast timeTable map value to duration"))
		}
		return existing
	} else {
		return timer
	}
}

// Store time entry in TimeLog (the time log itself is stored in the stopwatch sync.Map)
func (t *TimeLog) LogTimeEntry(d time.Duration) {
	if t != nil && t.entries != nil {
		t.entries = append(t.entries, d)
	}
}

// Print all entries from TimeLog and reset it
func (st *Stopwatch) PrintLog() {
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
		value, ok := v.(*TimeLog)
		if !ok {
			wasErr = fmt.Errorf("cannot cast timeTable map value to duration")
			// stops the iteration
			return false
		}
		// Print time value for every and calculate overall time
		for index, entry := range value.entries {
			name := key
			if index != 0 {
				name = key + "#" + strconv.Itoa(index)
			}
			st.logger.WithFields(logging.Fields{"conf": st.name, "durationInNs": entry.Nanoseconds()}).Infof("%v call took %v", name, entry)
			overall += entry
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