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
	"fmt"
	"reflect"
	"sort"
	"strings"
)

const (
	initNodeKeysCap       = 1000
	initEdgesCap          = 10000
	initDirDepthBoundsCap = 10

	dirSeparator = "/"
)

// edgeLookup is a helper tool used internally by kvgraph for **efficient** lookups
// over the set of graph edges, defined using keys or key prefixes.
type edgeLookup struct {
	nodeKeyData   []nodeKey
	nodeKeyOffset []int           // for O(log(n)) lookups against key prefixes
	nodeKeyMap    map[string]bool // for O(1) lookups against full keys
	removedNodes  int

	edgeData     []edge // unordered
	edgeOffset   []int  // ordered first by directory depth, then lexicographically by edge data
	removedEdges int
	// edges are first sorted (split) by the number of target key components (directories)
	// -> dirDepthBounds[dirCount] = index of the first edge in <edgeOffset> whose
	//    targetKey consists of <dirCount> directories (incl. the last suffix)
	dirDepthBounds []int

	overlay  *edgeLookup
	underlay *edgeLookup

	methodTracker MethodTracker
}

type nodeKey struct {
	key     string
	removed bool

	decOffset int // used internally in gcNodeKeys()
}

type edge struct {
	// can be empty prefix to match all
	targetKey string
	isPrefix  bool

	sourceNode string
	relation   string
	label      string

	removed bool

	decOffset int // used internally in gcEdges()
}

func newEdgeLookup(mt MethodTracker) *edgeLookup {
	return &edgeLookup{
		nodeKeyData:    make([]nodeKey, 0, initNodeKeysCap),
		nodeKeyOffset:  make([]int, 0, initNodeKeysCap),
		nodeKeyMap:     make(map[string]bool),
		edgeData:       make([]edge, 0, initEdgesCap),
		edgeOffset:     make([]int, 0, initEdgesCap),
		dirDepthBounds: make([]int, 0, initDirDepthBoundsCap),
		methodTracker:  mt,
	}
}

func (el *edgeLookup) reset() {
	el.nodeKeyMap = make(map[string]bool)
	el.nodeKeyData = el.nodeKeyData[:0]
	el.nodeKeyOffset = el.nodeKeyOffset[:0]
	el.removedNodes = 0
	el.edgeData = el.edgeData[:0]
	el.edgeOffset = el.edgeOffset[:0]
	el.dirDepthBounds = el.dirDepthBounds[:0]
	el.removedEdges = 0
}

func (el *edgeLookup) makeOverlay() *edgeLookup {
	if el.overlay == nil {
		// create overlay for the first time
		el.overlay = &edgeLookup{
			nodeKeyData:    make([]nodeKey, 0, max(len(el.nodeKeyData), initNodeKeysCap)),
			nodeKeyOffset:  make([]int, 0, max(len(el.nodeKeyOffset), initNodeKeysCap)),
			edgeData:       make([]edge, 0, max(len(el.edgeData), initEdgesCap)),
			edgeOffset:     make([]int, 0, max(len(el.edgeOffset), initEdgesCap)),
			dirDepthBounds: make([]int, 0, max(len(el.dirDepthBounds), initDirDepthBoundsCap)),
			underlay:       el,
		}
	}
	// re-use previously allocated memory
	el.overlay.resizeNodeKeys(len(el.nodeKeyOffset))
	el.overlay.resizeEdges(len(el.edgeOffset))
	el.overlay.resizeDirDepthBounds(len(el.dirDepthBounds))
	copy(el.overlay.nodeKeyData, el.nodeKeyData)
	copy(el.overlay.nodeKeyOffset, el.nodeKeyOffset)
	copy(el.overlay.edgeData, el.edgeData)
	copy(el.overlay.edgeOffset, el.edgeOffset)
	copy(el.overlay.dirDepthBounds, el.dirDepthBounds)
	el.overlay.nodeKeyMap = make(map[string]bool)
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
	for key, add := range el.nodeKeyMap {
		if add {
			el.underlay.nodeKeyMap[key] = true
		} else {
			delete(el.underlay.nodeKeyMap, key)
		}
	}
	el.nodeKeyMap = make(map[string]bool) // clear
	el.underlay.resizeNodeKeys(len(el.nodeKeyOffset))
	el.underlay.resizeEdges(len(el.edgeOffset))
	el.underlay.resizeDirDepthBounds(len(el.dirDepthBounds))
	copy(el.underlay.nodeKeyData, el.nodeKeyData)
	copy(el.underlay.nodeKeyOffset, el.nodeKeyOffset)
	copy(el.underlay.edgeData, el.edgeData)
	copy(el.underlay.edgeOffset, el.edgeOffset)
	copy(el.underlay.dirDepthBounds, el.dirDepthBounds)
}

// O(log(n))
func (el *edgeLookup) addNodeKey(key string) {
	if el.methodTracker != nil {
		defer el.methodTracker("edgeLookup.addNodeKey")()
	}

	el.nodeKeyMap[key] = true

	// find the corresponding index in nodeKeyOffset
	idx := el.nodeKeyIdx(key)
	if idx < len(el.nodeKeyOffset) {
		offset := el.nodeKeyOffset[idx]
		if el.nodeKeyData[offset].key == key {
			if el.nodeKeyData[offset].removed {
				el.nodeKeyData[offset].removed = false
				el.removedNodes--
			}
			return
		}
	}

	// add to both nodeKeyOffset and nodeKeyData
	el.nodeKeyData = append(el.nodeKeyData, nodeKey{})
	offset := len(el.nodeKeyData) - 1
	el.nodeKeyData[offset].key = key
	el.nodeKeyData[offset].removed = false
	el.nodeKeyOffset = append(el.nodeKeyOffset, -1)
	if idx < len(el.nodeKeyOffset)-1 {
		copy(el.nodeKeyOffset[idx+1:], el.nodeKeyOffset[idx:])
	}
	el.nodeKeyOffset[idx] = offset
}

// O(log(n)) amortized
func (el *edgeLookup) delNodeKey(key string) {
	if el.methodTracker != nil {
		defer el.methodTracker("edgeLookup.delNodeKey")()
	}

	if el.underlay != nil {
		// this is overlay, remember operation
		el.nodeKeyMap[key] = false
	} else {
		// do not store false, otherwise memory usage will grow
		delete(el.nodeKeyMap, key)
	}
	idx := el.nodeKeyIdx(key)
	if idx < len(el.nodeKeyOffset) {
		offset := el.nodeKeyOffset[idx]
		if el.nodeKeyData[offset].key == key && !el.nodeKeyData[offset].removed {
			el.nodeKeyData[offset].removed = true
			el.removedNodes++
			if el.removedNodes > len(el.nodeKeyData)/2 {
				el.gcNodeKeys()
			}
		}
	}
}

// O(log(n))
func (el *edgeLookup) nodeKeyIdx(key string) int {
	return sort.Search(len(el.nodeKeyOffset),
		func(i int) bool {
			return key <= el.nodeKeyData[el.nodeKeyOffset[i]].key
		})
}

// O(log(m))
func (el *edgeLookup) addEdge(e edge) {
	if el.methodTracker != nil {
		defer el.methodTracker("edgeLookup.addEdge")()
	}

	// find the corresponding index in edgeOffset
	e.targetKey = trimTrailingDirSep(e.targetKey)
	dirDepth := getDirDepth(e.targetKey)
	idx := el.edgeIdx(e, dirDepth)
	if idx < len(el.edgeOffset) {
		offset := el.edgeOffset[idx]
		equal, _ := e.compare(el.edgeData[offset])
		if equal {
			if el.edgeData[offset].removed {
				el.edgeData[offset].removed = false
				el.removedEdges--
			}
			return
		}
	}

	// add to both edgeOffset and edgeData
	el.edgeData = append(el.edgeData, e)
	offset := len(el.edgeData) - 1
	el.edgeData[offset].removed = false
	el.edgeOffset = append(el.edgeOffset, -1)
	if idx < len(el.edgeOffset)-1 {
		copy(el.edgeOffset[idx+1:], el.edgeOffset[idx:])
	}
	el.edgeOffset[idx] = offset

	// update directory boundaries
	for i := dirDepth + 1; i < len(el.dirDepthBounds); i++ {
		el.dirDepthBounds[i]++
	}
	for i := len(el.dirDepthBounds); i <= dirDepth; i++ {
		el.dirDepthBounds = append(el.dirDepthBounds, len(el.edgeOffset)-1)
	}
}

// O(log(m)) amortized
func (el *edgeLookup) delEdge(e edge) {
	if el.methodTracker != nil {
		defer el.methodTracker("edgeLookup.delEdge")()
	}

	e.targetKey = trimTrailingDirSep(e.targetKey)
	dirDepth := getDirDepth(e.targetKey)
	idx := el.edgeIdx(e, dirDepth)
	if idx < len(el.edgeOffset) {
		offset := el.edgeOffset[idx]
		equal, _ := e.compare(el.edgeData[offset])
		if equal && !el.edgeData[offset].removed {
			el.edgeData[offset].removed = true
			el.removedEdges++
			if el.removedEdges > len(el.edgeData)/2 {
				el.gcEdges()
			}
		}
	}
}

// O(log(m))
func (el *edgeLookup) edgeIdx(e edge, dirDepth int) int {
	begin, end := el.getDirDepthBounds(dirDepth)
	if begin == end {
		return begin
	}
	return begin + sort.Search(end-begin,
		func(i int) bool {
			e2 := el.edgeData[el.edgeOffset[begin+i]]
			_, order := e.compare(e2)
			return order <= 0
		})
}

func (el *edgeLookup) getDirDepthBounds(dirDepth int) (begin, end int) {
	if dirDepth < len(el.dirDepthBounds) {
		begin = el.dirDepthBounds[dirDepth]
	} else {
		begin = len(el.edgeOffset)
	}
	if dirDepth < len(el.dirDepthBounds)-1 {
		end = el.dirDepthBounds[dirDepth+1]
	} else {
		end = len(el.edgeOffset)
	}
	return
}

// for prefix: O(log(n)) (assuming O(1) matched keys)
// for full key: O(1) average, O(n) worst-case
func (el *edgeLookup) iterTargets(key string, isPrefix bool, cb func(targetNode string)) {
	if el.methodTracker != nil {
		defer el.methodTracker("edgeLookup.iterTargets")()
	}

	if key == "" && isPrefix {
		// iterate all
		for i := range el.nodeKeyData {
			if el.nodeKeyData[i].removed {
				continue
			}
			cb(el.nodeKeyData[i].key)
		}
		return
	}
	if !isPrefix {
		added, known := el.nodeKeyMap[key]
		if (known && !added) || (!known && el.underlay == nil) {
			return
		}
		if !known && el.underlay != nil {
			_, added = el.underlay.nodeKeyMap[key]
			if !added {
				return
			}
		}
		cb(key)
		return
	}
	// prefix:
	idx := el.nodeKeyIdx(key)
	for i := idx; i < len(el.nodeKeyOffset); i++ {
		offset := el.nodeKeyOffset[i]
		if el.nodeKeyData[offset].removed {
			continue
		}
		if !strings.HasPrefix(el.nodeKeyData[offset].key, key) {
			break
		}
		cb(el.nodeKeyData[offset].key)
	}
}

// O(log(m)) (assuming O(1) matched sources)
func (el *edgeLookup) iterSources(targetKey string, cb func(sourceNode, relation, label string)) {
	if el.methodTracker != nil {
		defer el.methodTracker("edgeLookup.iterSources")()
	}
	targetKey = trimTrailingDirSep(targetKey)

	var dirDepth int
	for i := 0; i <= len(targetKey); i++ {
		prefix := i < len(targetKey)
		if i == 0 || !prefix || targetKey[i] == dirSeparator[0] {
			idx := el.edgeIdx(edge{targetKey: targetKey[:i]}, dirDepth)
			_, end := el.getDirDepthBounds(dirDepth)
			for j := idx; j < end; j++ {
				offset := el.edgeOffset[j]
				if el.edgeData[offset].targetKey != targetKey[:i] {
					break
				}
				if prefix && !el.edgeData[offset].isPrefix {
					continue
				}
				if el.edgeData[offset].removed {
					continue
				}
				cb(el.edgeData[offset].sourceNode, el.edgeData[offset].relation, el.edgeData[offset].label)
			}
			dirDepth++
		}
	}
}

// O(n)
func (el *edgeLookup) gcNodeKeys() {
	// for each offset determine how much it will decrease
	var decOffset int
	for i := 0; i < len(el.nodeKeyData); i++ {
		if el.nodeKeyData[i].removed {
			decOffset++
		} else {
			el.nodeKeyData[i].decOffset = decOffset
		}
	}

	// GC node-key offsets
	var next int
	for i := range el.nodeKeyOffset {
		offset := el.nodeKeyOffset[i]
		if !el.nodeKeyData[offset].removed {
			if next < i {
				el.nodeKeyOffset[next] = el.nodeKeyOffset[i]
			}
			el.nodeKeyOffset[next] -= el.nodeKeyData[offset].decOffset
			next++
		}
	}
	el.nodeKeyOffset = el.nodeKeyOffset[:next]

	// GC node-key data
	next = 0
	for i := 0; i < len(el.nodeKeyData); i++ {
		if !el.nodeKeyData[i].removed {
			if next < i {
				el.nodeKeyData[next] = el.nodeKeyData[i]
			}
			next++
		}
	}
	el.nodeKeyData = el.nodeKeyData[:next]

	el.removedNodes = 0
	if len(el.nodeKeyOffset) != len(el.nodeKeyData) {
		panic("len(el.nodeKeyOffset) != len(el.nodeKeyData)")
	}
}

// O(m)
func (el *edgeLookup) gcEdges() {
	// for each offset determine how much it will decrease
	var decOffset int
	for i := 0; i < len(el.edgeData); i++ {
		if el.edgeData[i].removed {
			decOffset++
		} else {
			el.edgeData[i].decOffset = decOffset
		}
	}

	// GC edge offsets
	var next int
	for dIdx, curBound := range el.dirDepthBounds {
		newBound := next
		nextBound := len(el.edgeOffset)
		if dIdx < len(el.dirDepthBounds)-1 {
			nextBound = el.dirDepthBounds[dIdx+1]
		}
		for i := curBound; i < nextBound; i++ {
			offset := el.edgeOffset[i]
			if !el.edgeData[offset].removed {
				if next < i {
					el.edgeOffset[next] = el.edgeOffset[i]
				}
				// update offset to reflect the post-GC situation
				el.edgeOffset[next] -= el.edgeData[offset].decOffset
				next++
			}
		}
		el.dirDepthBounds[dIdx] = newBound
	}
	el.edgeOffset = el.edgeOffset[:next]

	// GC edge data
	next = 0
	for i := 0; i < len(el.edgeData); i++ {
		if !el.edgeData[i].removed {
			if next < i {
				el.edgeData[next] = el.edgeData[i]
			}
			next++
		}
	}
	el.edgeData = el.edgeData[:next]

	// GC directory boundaries
	dIdx := len(el.dirDepthBounds) - 1
	for ; dIdx >= 0; dIdx-- {
		if el.dirDepthBounds[dIdx] < len(el.edgeOffset) {
			break
		}
	}
	el.dirDepthBounds = el.dirDepthBounds[:dIdx+1]

	el.removedEdges = 0
	if len(el.edgeOffset) != len(el.edgeData) {
		panic("len(el.edgeOffset) != len(el.edgeData)")
	}
}

func (el *edgeLookup) resizeNodeKeys(size int) {
	if cap(el.nodeKeyData) < size {
		el.nodeKeyData = make([]nodeKey, size)
		el.nodeKeyOffset = make([]int, size)
	}
	el.nodeKeyData = el.nodeKeyData[0:size]
	el.nodeKeyOffset = el.nodeKeyOffset[0:size]
}

func (el *edgeLookup) resizeEdges(size int) {
	if cap(el.edgeData) < size {
		el.edgeData = make([]edge, size)
		el.edgeOffset = make([]int, size)
	}
	el.edgeData = el.edgeData[0:size]
	el.edgeOffset = el.edgeOffset[0:size]
}

func (el *edgeLookup) resizeDirDepthBounds(size int) {
	if cap(el.dirDepthBounds) < size {
		el.dirDepthBounds = make([]int, size)
	}
	el.dirDepthBounds = el.dirDepthBounds[0:size]
}

// for UTs
func (el *edgeLookup) verifyDirDepthBounds() error {
	if len(el.edgeData) != len(el.edgeOffset) {
		return fmt.Errorf("len(edgeData) != len(edgeOffset) (%d != %d)",
			len(el.edgeData), len(el.edgeOffset))
	}
	if cap(el.edgeData) != cap(el.edgeOffset) {
		return fmt.Errorf("cap(edgeData) != cap(edgeOffset) (%d != %d)",
			cap(el.edgeData), cap(el.edgeOffset))
	}
	for i := 0; i < len(el.edgeOffset); i++ {
		found := false
		for j := 0; j < len(el.edgeOffset); j++ {
			if el.edgeOffset[j] == i {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing entry for edge offset %d (offsets=%+v)",
				i, el.edgeOffset)
		}
	}

	if len(el.nodeKeyData) != len(el.nodeKeyOffset) {
		return fmt.Errorf("len(nodeKeyData) != len(nodeKeyOffset) (%d != %d)",
			len(el.nodeKeyData), len(el.nodeKeyOffset))
	}
	if cap(el.nodeKeyData) != cap(el.nodeKeyOffset) {
		return fmt.Errorf("cap(nodeKeyData) != cap(nodeKeyOffset) (%d != %d)",
			cap(el.nodeKeyData), cap(el.nodeKeyOffset))
	}
	for i := 0; i < len(el.nodeKeyOffset); i++ {
		found := false
		for j := 0; j < len(el.nodeKeyOffset); j++ {
			if el.nodeKeyOffset[j] == i {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing entry for node-key offset %d (offsets=%+v)",
				i, el.nodeKeyOffset)
		}
	}

	expBounds := []int{}
	dirDepth := -1
	for i := range el.edgeOffset {
		tk := el.edgeData[el.edgeOffset[i]].targetKey
		if len(tk) > 0 && tk[len(tk)-1] == dirSeparator[0] {
			return fmt.Errorf("edge with targetKey ending with dir separator: %s", tk)
		}
		var tkDirDepth int
		if tk != "" {
			tkDirDepth = len(strings.Split(tk, dirSeparator))
		}

		if tkDirDepth < dirDepth {
			return fmt.Errorf("edge with targetKey inserted at a wrong dir depth (%d): %s",
				dirDepth, tk)
		}
		for j := dirDepth + 1; j <= tkDirDepth; j++ {
			expBounds = append(expBounds, i)
		}
		dirDepth = tkDirDepth
	}
	// bad performance of this is OK, the method is used only in unit tests
	if !reflect.DeepEqual(el.dirDepthBounds, expBounds) {
		return fmt.Errorf("unexpected dir-depth bounds: expected=%v, actual=%v "+
			"(offsets=%+v, data=%+v)",
			expBounds, el.dirDepthBounds, el.edgeOffset, el.edgeData)
	}
	return nil
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

func trimTrailingDirSep(s string) string {
	for len(s) > 0 && s[0] == dirSeparator[0] {
		s = s[1:]
	}
	for len(s) > 0 && s[len(s)-1] == dirSeparator[0] {
		s = s[:len(s)-1]
	}
	return s
}

func getDirDepth(s string) int {
	var depth int
	if len(s) > 0 {
		depth++ // include last suffix (assuming no trailing separator)
	}
	depth += strings.Count(s, dirSeparator)
	return depth
}
