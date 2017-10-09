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
)

// Stopwatch keeps all time measurement results
type Stopwatch struct {
	// start of the resync
	Overall time.Duration
	// map where measurements are stored
	timeTable map[string]time.Duration
}

func NewStopwatch() *Stopwatch {
	return &Stopwatch{
		// Default value
		Overall:  -1,
		timeTable: make(map[string]time.Duration),
	}
}

func (st *Stopwatch) LogTime(n interface{}, d time.Duration) {
	name := reflect.TypeOf(n).String()
	_, found := st.timeTable[name]
	if found {
		// index more occurences of the same binapi
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
	st.timeTable[name] = d
}

func (st *Stopwatch) Print(pluginName string, log logging.Logger) {
	if len(st.timeTable) == 0 {
		log.WithField("plugin", pluginName).Infof("Timer: no entries")
	}
	for k, v := range st.timeTable {
		log.WithField("plugin", pluginName).Infof("Calling %v took %v", k, v)
	}
	if st.Overall != -1 {
		log.WithField("plugin", pluginName).Infof("Resync took %v", st.Overall)
	}
	// purge map
	st.timeTable = make(map[string]time.Duration)
}
