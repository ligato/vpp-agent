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

package graph

import (
	"sort"
	"strings"
)

const (
	initNodeKeysCap = 1000
	initEdgesCap    = 10000
)

// edgeLookup is a helper tool used internally by kvgraph for **efficient** lookups
// over the set of graph edges, defined using keys or key prefixes.
type edgeLookup struct {
	nodeKeys     []nodeKey       // for O(log(n)) lookups against key prefixes
	nodeKeysMap  map[string]bool // for O(1) lookups against full keys
	removedNodes int

	edges        []edge
	removedEdges int

	overlay  *edgeLookup
	underlay *edgeLookup
}

type nodeKey struct {
	key     string
	removed bool
}

type edge struct {
	// can be empty prefix to match all
	targetKey string
	isPrefix  bool

	sourceNode string
	relation   string
	label      string

	removed bool
}

func newEdgeLookup() *edgeLookup {
	return &edgeLookup{
		nodeKeys:    make([]nodeKey, 0, initNodeKeysCap),
		nodeKeysMap: make(map[string]bool),
		edges:       make([]edge, 0, initEdgesCap),
	}
}

func (el *edgeLookup) reset() {
	el.nodeKeysMap = make(map[string]bool)
	el.nodeKeys = el.nodeKeys[:0]
	el.removedNodes = 0
	el.edges = el.edges[:0]
	el.removedEdges = 0
}

func (el *edgeLookup) makeOverlay() *edgeLookup {
	if el.overlay == nil {
		// create overlay for the first time
		el.overlay = &edgeLookup{
			nodeKeys: make([]nodeKey, 0, max(len(el.nodeKeys), initNodeKeysCap)),
			edges:    make([]edge, 0, max(len(el.edges), initEdgesCap)),
			underlay: el,
		}
	}
	// re-use previously allocated memory
	el.overlay.resizeNodeKeys(len(el.nodeKeys))
	el.overlay.resizeEdges(len(el.edges))
	copy(el.overlay.nodeKeys, el.nodeKeys)
	copy(el.overlay.edges, el.edges)
	el.overlay.nodeKeysMap = make(map[string]bool)
	el.overlay.removedEdges = el.removedEdges
	el.overlay.removedNodes = el.removedNodes
	return el.overlay
}

func (el *edgeLookup) saveOverlay() {
	if el.underlay == nil {
		panic("called saveOverlay on what is not overlay")
	}
	el.underlay.removedNodes = el.removedNodes
	el.underlay.removedEdges = el.removedEdges
	for key, add := range el.nodeKeysMap {
		if add {
			el.underlay.nodeKeysMap[key] = true
		} else {
			delete(el.underlay.nodeKeysMap, key)
		}
	}
	el.nodeKeysMap = make(map[string]bool) // clear
	el.underlay.resizeNodeKeys(len(el.nodeKeys))
	el.underlay.resizeEdges(len(el.edges))
	copy(el.underlay.nodeKeys, el.nodeKeys)
	copy(el.underlay.edges, el.edges)
}

// O(log(n))
func (el *edgeLookup) addNodeKey(key string) {
	el.nodeKeysMap[key] = true
	idx := el.nodeKeyIdx(key)
	if idx < len(el.nodeKeys) && el.nodeKeys[idx].key == key {
		if el.nodeKeys[idx].removed {
			el.nodeKeys[idx].removed = false
			el.removedNodes--
		}
		return
	}
	el.nodeKeys = append(el.nodeKeys, nodeKey{})
	if idx < len(el.nodeKeys)-1 {
		copy(el.nodeKeys[idx+1:], el.nodeKeys[idx:])
	}
	el.nodeKeys[idx].key = key
	el.nodeKeys[idx].removed = false
}

// O(log(n)) amortized
func (el *edgeLookup) delNodeKey(key string) {
	if el.underlay != nil {
		// this is overlay, remember operation
		el.nodeKeysMap[key] = false
	} else {
		// do not store false, otherwise memory usage will grow
		delete(el.nodeKeysMap, key)
	}
	idx := el.nodeKeyIdx(key)
	if idx < len(el.nodeKeys) && el.nodeKeys[idx].key == key && !el.nodeKeys[idx].removed {
		el.nodeKeys[idx].removed = true
		el.removedNodes++
		if el.removedNodes > len(el.nodeKeys)/2 {
			el.gcNodeKeys()
		}
	}
}

// O(log(n))
func (el *edgeLookup) nodeKeyIdx(key string) int {
	return sort.Search(len(el.nodeKeys),
		func(i int) bool {
			return key <= el.nodeKeys[i].key
		})
}

// O(log(m))
func (el *edgeLookup) addEdge(e edge) {
	idx := el.edgeIdx(e)
	if idx < len(el.edges) {
		equal, _ := e.compare(el.edges[idx])
		if equal {
			if el.edges[idx].removed {
				el.edges[idx].removed = false
				el.removedEdges--
			}
			return
		}
	}
	el.edges = append(el.edges, edge{})
	if idx < len(el.edges)-1 {
		copy(el.edges[idx+1:], el.edges[idx:])
	}
	el.edges[idx] = e
	el.edges[idx].removed = false
}

// O(log(m)) amortized
func (el *edgeLookup) delEdge(e edge) {
	idx := el.edgeIdx(e)
	if idx <= len(el.edges) {
		equal, _ := e.compare(el.edges[idx])
		if equal && !el.edges[idx].removed {
			el.edges[idx].removed = true
			el.removedEdges++
			if el.removedEdges > len(el.edges)/2 {
				el.gcEdges()
			}
		}
	}
}

// O(log(m))
func (el *edgeLookup) edgeIdx(e edge) int {
	return sort.Search(len(el.edges),
		func(i int) bool {
			_, order := e.compare(el.edges[i])
			return order <= 0
		})
}

// for prefix: O(log(n)) (assuming O(1) matched keys)
// for full key: O(1) average, O(n) worst-case
func (el *edgeLookup) iterTargets(key string, isPrefix bool, cb func(targetNode string)) {
	if key == "" && isPrefix {
		// iterate all
		for i := range el.nodeKeys {
			if el.nodeKeys[i].removed {
				continue
			}
			cb(el.nodeKeys[i].key)
		}
		return
	}
	if !isPrefix {
		added, known := el.nodeKeysMap[key]
		if (known && !added) || (!known && el.underlay == nil) {
			return
		}
		if !known && el.underlay != nil {
			_, added = el.underlay.nodeKeysMap[key]
			if !added {
				return
			}
		}
		cb(key)
		return
	}
	// prefix:
	idx := el.nodeKeyIdx(key)
	for i := idx; i < len(el.nodeKeys); i++ {
		if el.nodeKeys[i].removed {
			continue
		}
		if !strings.HasPrefix(el.nodeKeys[i].key, key) {
			break
		}
		cb(el.nodeKeys[i].key)
	}
}

// O(log(m)) (assuming O(1) matched sources)
func (el *edgeLookup) iterSources(targetKey string, cb func(sourceNode, relation, label string)) {
	// iter over key selectors without key prefixes
	for i := range el.edges {
		if el.edges[i].targetKey != "" {
			break
		}
		if el.edges[i].removed {
			continue
		}
		if el.edges[i].isPrefix {
			cb(el.edges[i].sourceNode, el.edges[i].relation, el.edges[i].label)
		}
	}
	idx := el.edgeIdx(edge{targetKey: targetKey})
	// iter over key prefixes (smaller than the length of targetKey)
	for i := idx - 1; i >= 0; i-- {
		if el.edges[i].removed {
			continue
		}
		if !el.edges[i].isPrefix {
			continue
		}
		if el.edges[i].targetKey == "" {
			// empty key prefixes already iterated above
			break
		}
		if !strings.HasPrefix(targetKey, el.edges[i].targetKey) {
			break
		}
		cb(el.edges[i].sourceNode, el.edges[i].relation, el.edges[i].label)
	}
	// iter over exactly matching key(-prefixes)
	for i := idx; i < len(el.edges); i++ {
		if el.edges[i].removed {
			continue
		}
		if el.edges[i].targetKey != targetKey {
			break
		}
		cb(el.edges[i].sourceNode, el.edges[i].relation, el.edges[i].label)
	}
}

// O(n)
func (el *edgeLookup) gcNodeKeys() {
	var next int
	for i := range el.nodeKeys {
		if !el.nodeKeys[i].removed {
			if next < i {
				el.nodeKeys[next] = el.nodeKeys[i]
			}
			next++
		}
	}
	el.nodeKeys = el.nodeKeys[:next]
	el.removedNodes = 0
}

// O(m)
func (el *edgeLookup) gcEdges() {
	var next int
	for i := range el.edges {
		if !el.edges[i].removed {
			if next < i {
				el.edges[next] = el.edges[i]
			}
			next++
		}
	}
	el.edges = el.edges[:next]
	el.removedEdges = 0
}

func (el *edgeLookup) resizeNodeKeys(size int) {
	if cap(el.nodeKeys) < size {
		el.nodeKeys = make([]nodeKey, size)
	}
	el.nodeKeys = el.nodeKeys[0:size]
}

func (el *edgeLookup) resizeEdges(size int) {
	if cap(el.edges) < size {
		el.edges = make([]edge, size)
	}
	el.edges = el.edges[0:size]
}

func (e edge) compare(e2 edge) (equal bool, order int) {
	if e.targetKey < e2.targetKey {
		return false, -1
	}
	if e.targetKey > e2.targetKey {
		return false, 1
	}
	if e.isPrefix != e2.isPrefix {
		if !e.isPrefix {
			return false, -1
		}
		return false, 1
	}
	if e.sourceNode < e2.sourceNode {
		return false, -1
	}
	if e.sourceNode > e2.sourceNode {
		return false, 1
	}
	if e.relation < e2.relation {
		return false, -1
	}
	if e.relation > e2.relation {
		return false, 1
	}
	if e.label < e2.label {
		return false, -1
	}
	if e.label > e2.label {
		return false, 1
	}
	return true, 0
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}