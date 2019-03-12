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

// edgeLookup is a helper tool used internally by kvgraph for efficient lookups
// over the set of graph edges, defined using keys or key prefixes.
type edgeLookup struct {
	nodeKeys     []nodeKey
	removedNodes int

	edges          []edge
	removedTargets int
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
}

func (el *edgeLookup) copy() *edgeLookup {
	elCopy := &edgeLookup{
		nodeKeys:       make([]nodeKey, len(el.nodeKeys)),
		removedNodes:   el.removedNodes,
		edges:          make([]edge, len(el.edges)),
		removedTargets: el.removedTargets,
	}
	copy(elCopy.nodeKeys, el.nodeKeys)
	copy(elCopy.edges, el.edges)
	return elCopy
}

// O(log(n))
func (el *edgeLookup) addNodeKey(key string) {
}

// O(log(n)) amortized
func (el *edgeLookup) delNodeKey(key string) {
}

// O(log(m))
func (el *edgeLookup) addEdge(e edge) {
}

// O(log(m)) amortized
func (el *edgeLookup) delEdge(e edge) {
}

// O(log(n))
func (el *edgeLookup) iterTargets(key string, isPrefix bool, cb func(targetNode string)) {
}

// O(log(m))
func (el *edgeLookup) iterSources(targetKey string, cb func(sourceNode, relation, label string)) {
}

// O(n)
func (el *edgeLookup) gcNodeKeys() {
}

// O(m)
func (el *edgeLookup) gcEdges() {
}

func (e edge) equals(e2 edge) bool {
	return e.targetKey == e2.targetKey && e.isPrefix == e2.isPrefix &&
		e.sourceNode == e2.sourceNode && e.relation == e2.relation && e.label == e2.label
}
