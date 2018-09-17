// Copyright (c) 2018 Cisco and/or its affiliates.
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

//go:generate protoc --proto_path=model/apitrace --gogo_out=model/apitrace model/apitrace/apitrace.proto

import (
	"sync"
	"time"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure/model/apitrace"
)

// Tracer allows to measure, store and list measured time entries.
type Tracer interface {
	// Starts the time measurement. All logged entries are calculated since the start.
	Start()
	// LogTime puts measured time to the table and resets the time.
	LogTime(entity string)
	// Print logs all entries
	Print()
	// Get all trace entries stored
	Get() *apitrace.Trace
	// Clear removes entries from the log database
	Clear()
}

// NewTracer creates new tracer object
func NewTracer(name string, log logging.Logger) Tracer {
	return &tracer{
		name:   name,
		log:    log,
		index:  1,
		timedb: make(map[int]*entry),
	}
}

// Inner structure handling database and measure results
type tracer struct {
	sync.Mutex

	name      string
	log       logging.Logger
	startTime time.Time
	// Entry index, used in database as key and increased after every entry. Never resets since the tracer object is
	// created or the database is cleared
	index int
	// Time database, uses index as key and entry as value
	timedb map[int]*entry
}

// Single time entry
type entry struct {
	entryName  string
	loggedTime time.Duration
}

func (t *tracer) Start() {
	t.startTime = time.Now()
}

func (t *tracer) LogTime(entity string) {
	if t == nil {
		return
	}
	// Skip cli-inband
	if entity == "cli_inband" {
		return
	}

	t.Lock()
	defer t.Unlock()

	// Store time
	t.timedb[t.index] = &entry{
		entryName:  entity,
		loggedTime: time.Since(t.startTime),
	}
	t.index++
}

func (t *tracer) Print() {
	t.process()
}

func (t *tracer) Get() *apitrace.Trace {
	return t.process()
}

func (t *tracer) Clear() {
	t.timedb = make(map[int]*entry)
}

func (t *tracer) process() *apitrace.Trace {
	t.Lock()
	defer t.Unlock()

	var (
		trace = &apitrace.Trace{
			TracedEntries: make([]*apitrace.Trace_TracedEntry, 0),
		}
		data  []*apitrace.Trace_TracedEntry
		total time.Duration
	)

	for idx := 1; idx <= len(t.timedb); idx++ {
		entry, ok := t.timedb[idx]
		if !ok {
			t.log.Errorf("failed to print tracer: timedb processing error")
			return nil
		}
		total += entry.loggedTime
		message := &apitrace.Trace_TracedEntry{
			Index:    uint32(idx),
			MsgName:  entry.entryName,
			Duration: entry.loggedTime.String(),
		}
		data = append(data, message)
	}

	// Log overall time
	trace.TracedEntries = data
	trace.Overall = total.String()

	return trace
}
