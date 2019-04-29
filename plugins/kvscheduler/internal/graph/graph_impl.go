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

package graph

import (
	"sync"
	"time"
)

const (
	// how often (at most) the log of previous revisions gets trimmed to remove
	// records too old to keep
	oldRevsTrimmingPeriod = 1 * time.Minute
)

// kvgraph implements Graph interface.
type kvgraph struct {
	rwLock sync.RWMutex
	graph  *graphR

	startTime           time.Time
	lastRevTrimming     time.Time // last time the history of revisions was trimmed
	recordOldRevs       bool
	recordAgeLimit      time.Duration
	permanentInitPeriod time.Duration

	methodTracker MethodTracker
}

// MethodTracker can be optionally supplied to track beginning and end of calls
// for (non-trivial) graph methods.
type MethodTracker func(method string) (onReturn func())

// Opts groups input options for the graph constructor.
type Opts struct {
	RecordOldRevs       bool
	RecordAgeLimit      uint32
	PermanentInitPeriod uint32

	MethodTracker MethodTracker
}

// NewGraph creates and new instance of key-value graph.
// <recordOldRevs> if enabled, will cause the graph to record the previous
// revisions of every node that have ever existed. <recordAgeLimit> is in minutes
// and allows to limit the maximum age of a record to keep, avoiding infinite
// memory usage growth. The initial phase of the execution is, however, of greater
// significance and <permanentInitPeriod> allows to keep records from that period
// permanently in memory.
func NewGraph(opts Opts) Graph {
	kvgraph := &kvgraph{
		startTime:           time.Now(),
		lastRevTrimming:     time.Now(),
		recordOldRevs:       opts.RecordOldRevs,
		recordAgeLimit:      time.Duration(opts.RecordAgeLimit) * time.Minute,
		permanentInitPeriod: time.Duration(opts.PermanentInitPeriod) * time.Minute,
		methodTracker:       opts.MethodTracker,
	}
	kvgraph.graph = newGraphR(opts.MethodTracker)
	kvgraph.graph.parent = kvgraph
	return kvgraph
}

// Read returns a graph handle for read-only access.
// The graph supports multiple concurrent readers.
// Release eventually using Release() method.
func (kvgraph *kvgraph) Read() ReadAccess {
	kvgraph.rwLock.RLock()
	return kvgraph.graph
}

// Write returns a graph handle for read-write access.
// The graph supports at most one writer at a time - i.e. it is assumed
// there is no write-concurrency.
// If <inPlace> is enabled, the changes are applied with immediate effect,
// otherwise they are propagated to the graph using Save().
// In-place Write handle holds write lock, therefore reading is blocked until
// the handle is released.
// If <record> is true, the changes will be recorded once the handle is
// released.
// Release eventually using Release() method.
func (kvgraph *kvgraph) Write(inPlace, record bool) RWAccess {
	if kvgraph.methodTracker != nil {
		defer kvgraph.methodTracker("Write")()
	}
	if inPlace {
		kvgraph.rwLock.Lock()
	}
	return newGraphRW(kvgraph.graph, inPlace, record)
}
