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

package graph

import (
	"fmt"
	"strconv"
	"testing"

	. "github.com/ligato/vpp-agent/plugins/kvscheduler/internal/test"
)

/*
------------------------
 KVGraph benchmarks
------------------------

How to run:
  - build test binary	`go test -c`
  - run all benchmarks:	`./graph.test -test.run=XXX -test.bench=.`
  - with CPU profile:	`./graph.test -test.run=XXX -test.bench=. -test.cpuprofile=cpu.out`
    - analyze profile: `go tool pprof cpu.out`
  - with mem profile:	`./graph.test -test.run=XXX -test.bench=. -test.memprofile mem.out`
    - analyze profile: `go tool pprof -alloc_space mem.out`
  - with trace profile:	`./graph.test -test.run=XXX -test.bench=. -test.trace trace.out`
    - analyze profile: `go tool trace -http=:6060 trace.out`

*/

const (
	historyAgeLimit     = 5
	permanentInitPeriod = 1
)

var scale = [...]int{1, 10, 100, 1000, 10000}

type scaleCtx struct {
	keys           map[int]string
	targets        map[int]string
	targetPrefixes map[int]string
}

func init() {
	benchEl = newEdgeLookup()
}

func BenchmarkScaleWithoutRec(b *testing.B) {
	benchmarkScale(b, Opts{
		RecordOldRevs: false,
	}, false, true)
}

func BenchmarkScaleWithoutRecInPlace(b *testing.B) {
	benchmarkScale(b, Opts{
		RecordOldRevs: false,
	}, true, true)
}

func BenchmarkScaleWithoutRecInPlaceWithoutDel(b *testing.B) {
	benchmarkScale(b, Opts{
		RecordOldRevs: false,
	}, true, false)
}

func BenchmarkScaleWithRec(b *testing.B) {
	benchmarkScale(b, Opts{
		RecordOldRevs:       true,
		RecordAgeLimit:      historyAgeLimit,
		PermanentInitPeriod: permanentInitPeriod,
	}, false, true)
}

func BenchmarkScaleWithRecInPlace(b *testing.B) {
	benchmarkScale(b, Opts{
		RecordOldRevs:       true,
		RecordAgeLimit:      historyAgeLimit,
		PermanentInitPeriod: permanentInitPeriod,
	}, true, true)
}

func BenchmarkScaleWithRecInPlaceWithoutDel(b *testing.B) {
	benchmarkScale(b, Opts{
		RecordOldRevs:       true,
		RecordAgeLimit:      historyAgeLimit,
		PermanentInitPeriod: permanentInitPeriod,
	}, true, false)
}

func benchmarkScale(b *testing.B, gOpts Opts, wInPlace, withDeletes bool) {
	for _, n := range scale {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			ctx := setupScale(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				runScale(ctx, n, gOpts, wInPlace, withDeletes)
			}
		})
	}
}

func setupScale(n int) scaleCtx {
	c := scaleCtx{
		keys:           make(map[int]string),
		targets:        make(map[int]string),
		targetPrefixes: make(map[int]string),
	}
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("prefix-%d/node-%d", i/10, i%10)
		targetPrefix := fmt.Sprintf("prefix-%d/", (i/10)-1)
		target := fmt.Sprintf("%snode-%d", targetPrefix, i%10)
		c.keys[i] = key
		c.targets[i] = target
		c.targetPrefixes[i] = targetPrefix
	}
	return c
}

func runScale(c scaleCtx, n int, gOpts Opts, wInPlace bool, withDeletes bool) {
	g := NewGraph(gOpts) // re-uses benchEl

	// create n nodes
	w := g.Write(wInPlace, gOpts.RecordOldRevs)
	for i := 0; i < n; i++ {
		node := w.SetNode(c.keys[i])
		node.SetFlags(ColorFlag(Green), TemporaryFlag())
		node.DelFlags(TemporaryFlagIndex)
		node.SetMetadata(i)
		node.SetTargets([]RelationTargetDef{
			{ // static key
				Relation: "relation1",
				Label:    "label",
				Key:      c.targets[i],
			},
			{ // key prefix + key selector
				Relation: "relation2",
				Label:    "label",
				Selector: TargetSelector{
					KeyPrefixes: []string{c.targetPrefixes[i]},
					KeySelector: func(key string) bool {
						return key == c.targets[i]
					},
				},
			},
		})
	}

	// save + release write handle
	if !wInPlace {
		w.Save()
	}
	w.Release()

	// read all the nodes
	r := g.Read()
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("prefix-%d/node-%d", i/10, i%10)
		node := r.GetNode(key)
		node.GetFlag(ColorFlagIndex)
		node.GetTargets("relation1")
		node.GetTargets("relation2")
		node.GetMetadata()
		node.GetSources("relation1")
		node.GetSources("relation2")
	}
	r.Release()

	if withDeletes {
		// remove all nodes
		w = g.Write(wInPlace, gOpts.RecordOldRevs)
		for i := 0; i < n; i++ {
			w.DeleteNode(c.keys[i])
		}

		// save + release write handle
		if !wInPlace {
			w.Save()
		}
		w.Release()
	}
}
