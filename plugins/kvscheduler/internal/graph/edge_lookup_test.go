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
	. "github.com/onsi/gomega"
	"testing"
)

type mockIter struct {
	visitedNodes map[string]struct{}
	visitedEdges map[visitedEdge]struct{}
}

type visitedEdge struct {
	sourceNode, relation, label string
}

func newMockIter() *mockIter {
	m := &mockIter{}
	m.reset()
	return m
}

func (m *mockIter) reset() {
	m.visitedNodes = make(map[string]struct{})
	m.visitedEdges = make(map[visitedEdge]struct{})
}

func (m *mockIter) visitNode(key string) {
	_, alreadyVisited := m.visitedNodes[key]
	Expect(alreadyVisited).To(BeFalse())
	m.visitedNodes[key] = struct{}{}
}

func (m *mockIter) visitEdge(sourceNode, relation, label string) {
	edge := visitedEdge{
		sourceNode: sourceNode,
		relation:   relation,
		label:      label,
	}
	_, alreadyVisited := m.visitedEdges[edge]
	Expect(alreadyVisited).To(BeFalse())
	m.visitedEdges[edge] = struct{}{}
}

func TestLookupOverNodeKeys(t *testing.T) {
	RegisterTestingT(t)

	mi := newMockIter()

	el := newEdgeLookup()
	Expect(el).ToNot(BeNil())

	el.iterTargets("some-key", false, mi.visitNode)
	Expect(mi.visitedNodes).To(BeEmpty())

	el.addNodeKey("prefix1/node1")
	el.addNodeKey("prefix2/node3")
	el.addNodeKey("prefix1/node2")

	// static key which exists
	el.iterTargets("prefix1/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()

	// static key which does not exist
	el.iterTargets("prefix2/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(BeEmpty())

	// prefix1
	el.iterTargets("prefix1/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(2))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node2"))
	mi.reset()

	// prefix2
	el.iterTargets("prefix2/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix2/node3"))
	mi.reset()

	// empty prefix (select all)
	el.iterTargets("", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(3))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node2"))
	Expect(mi.visitedNodes).To(HaveKey("prefix2/node3"))
	mi.reset()

	el.delNodeKey("prefix1/node2")

	// static key which no longer exists
	el.iterTargets("prefix1/node2", false, mi.visitNode)
	Expect(mi.visitedNodes).To(BeEmpty())

	// prefix1
	el.iterTargets("prefix1/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()

	// remove the rest
	el.delNodeKey("prefix1/node1")
	Expect(el.nodeKeys).To(HaveLen(1)) // gc-ed
	el.delNodeKey("prefix2/node3")
	Expect(el.nodeKeys).To(HaveLen(0)) // gc-ed

	// empty prefix (select all)
	el.iterTargets("", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(0))
	mi.reset()

	// test "attribute"
	el.addNodeKey("prefix1/node1/attr1")
	el.addNodeKey("prefix1/node1")
	el.addNodeKey("prefix1/node2")
	el.addNodeKey("prefix1/node1/attr2")

	// prefix1/node1 as key
	el.iterTargets("prefix1/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()

	// prefix1/node1 as prefix
	el.iterTargets("prefix1/node1", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(3))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1/attr1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1/attr2"))
	mi.reset()
}

func TestLookupOverNodeKeysWithOverlay(t *testing.T) {
	RegisterTestingT(t)

	mi := newMockIter()

	el := newEdgeLookup()
	Expect(el).ToNot(BeNil())

	el.addNodeKey("prefix1/node1")
	el.addNodeKey("prefix2/node3")

	elOver := el.makeOverlay()
	Expect(elOver).ToNot(BeNil())
	Expect(elOver.underlay).To(Equal(el))
	Expect(el.overlay).To(Equal(elOver))

	elOver.addNodeKey("prefix1/node2")

	// check in overlay
	elOver.iterTargets("prefix1/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	elOver.iterTargets("prefix2/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(BeEmpty())
	elOver.iterTargets("prefix1/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(2))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node2"))
	mi.reset()
	elOver.iterTargets("prefix2/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix2/node3"))
	mi.reset()
	elOver.iterTargets("", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(3))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node2"))
	Expect(mi.visitedNodes).To(HaveKey("prefix2/node3"))
	mi.reset()

	// check in the underlay before saving
	el.iterTargets("prefix1/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	el.iterTargets("prefix1/node2", false, mi.visitNode)
	Expect(mi.visitedNodes).To(BeEmpty())
	el.iterTargets("prefix1/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	el.iterTargets("prefix2/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix2/node3"))
	mi.reset()
	el.iterTargets("", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(2))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix2/node3"))
	mi.reset()

	// save and re-check in the underlay
	elOver.saveOverlay()
	el.iterTargets("prefix1/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	el.iterTargets("prefix1/node2", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node2"))
	mi.reset()
	el.iterTargets("prefix1/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(2))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node2"))
	mi.reset()
	el.iterTargets("prefix2/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix2/node3"))
	mi.reset()
	el.iterTargets("", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(3))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node2"))
	Expect(mi.visitedNodes).To(HaveKey("prefix2/node3"))
	mi.reset()

	// delete two in overlay
	elOver.delNodeKey("prefix1/node2")
	elOver.delNodeKey("prefix2/node3")

	// check in overlay:
	elOver.iterTargets("prefix1/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	elOver.iterTargets("prefix1/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	elOver.iterTargets("prefix2/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(BeEmpty())
	elOver.iterTargets("", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()

	// check in underlay before save:
	el.iterTargets("prefix1/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	el.iterTargets("prefix1/node2", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node2"))
	mi.reset()
	el.iterTargets("prefix1/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(2))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node2"))
	mi.reset()
	el.iterTargets("prefix2/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix2/node3"))
	mi.reset()
	el.iterTargets("", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(3))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node2"))
	Expect(mi.visitedNodes).To(HaveKey("prefix2/node3"))
	mi.reset()

	// save and re-check underlay
	elOver.saveOverlay()
	el.iterTargets("prefix1/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	el.iterTargets("prefix1/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	el.iterTargets("prefix2/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(BeEmpty())
	el.iterTargets("", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	Expect(elOver.nodeKeys).To(HaveLen(1))
	Expect(el.nodeKeys).To(HaveLen(1))

	// remove the last key in overlay, but do not save
	Expect(el.makeOverlay()).To(Equal(elOver))
	elOver.delNodeKey("prefix1/node1")
	Expect(elOver.nodeKeys).To(HaveLen(0))
	Expect(el.nodeKeys).To(HaveLen(1))
	elOver.iterTargets("", true, mi.visitNode)
	Expect(mi.visitedNodes).To(BeEmpty())

	// no changes
	elOver = el.makeOverlay()
	elOver.saveOverlay()
	el.iterTargets("prefix1/node1", false, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	el.iterTargets("prefix1/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	el.iterTargets("prefix2/", true, mi.visitNode)
	Expect(mi.visitedNodes).To(BeEmpty())
	el.iterTargets("", true, mi.visitNode)
	Expect(mi.visitedNodes).To(HaveLen(1))
	Expect(mi.visitedNodes).To(HaveKey("prefix1/node1"))
	mi.reset()
	Expect(elOver.nodeKeys).To(HaveLen(1))
	Expect(el.nodeKeys).To(HaveLen(1))
}

func TestLookupOverEdges(t *testing.T) {
	RegisterTestingT(t)

	mi := newMockIter()

	el := newEdgeLookup()
	Expect(el).ToNot(BeNil())

	el.iterSources("some-key", mi.visitEdge)
	Expect(mi.visitedEdges).To(BeEmpty())

	el.addEdge(edge{ // "prefix1/node2" -> "prefix1/node1"
		targetKey:  "prefix1/node1",
		isPrefix:   false,
		sourceNode: "prefix1/node2",
		relation:   "depends-on",
		label:      "node1 exists",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	el.addEdge(edge{ // "prefix1/node2" -> "prefix2/node3"
		targetKey:  "prefix2/node3",
		isPrefix:   false,
		sourceNode: "prefix1/node2",
		relation:   "depends-on",
		label:      "node3 exists",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())

	el.addEdge(edge{ // "prefix2/node3" -> "prefix1/*"
		targetKey:  "prefix1/",
		isPrefix:   true,
		sourceNode: "prefix2/node3",
		relation:   "depends-on",
		label:      "prefix1 non-empty",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	el.addEdge(edge{ // "prefix2/node3" -> "prefix2/node3"
		targetKey:  "prefix2/node3",
		isPrefix:   false,
		sourceNode: "prefix2/node3",
		relation:   "myself",
		label:      "edge to itself",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())

	el.addEdge(edge{
		targetKey:  "", // "prefix1/node1" -> *
		isPrefix:   true,
		sourceNode: "prefix1/node1",
		relation:   "all",
		label:      "all",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())

	Expect(el.edges).To(HaveLen(5))

	el.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(3))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()

	el.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(2))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()

	el.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(3))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node3 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "myself", "edge to itself"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()

	// delete 2 edges
	el.delEdge(edge{ // "prefix1/node2" -> "prefix2/node3"
		targetKey:  "prefix2/node3",
		isPrefix:   false,
		sourceNode: "prefix1/node2",
		relation:   "depends-on",
		label:      "node3 exists",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	el.delEdge(edge{
		targetKey:  "", // "prefix1/node1" -> *
		isPrefix:   true,
		sourceNode: "prefix1/node1",
		relation:   "all",
		label:      "all",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	Expect(el.edges).To(HaveLen(5)) // not gc-ed yet

	el.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(2))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	mi.reset()

	el.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	mi.reset()

	el.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "myself", "edge to itself"}))
	mi.reset()

	// delete another edge
	el.delEdge(edge{ // "prefix2/node3" -> "prefix1/*"
		targetKey:  "prefix1/",
		isPrefix:   true,
		sourceNode: "prefix2/node3",
		relation:   "depends-on",
		label:      "prefix1 non-empty",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	Expect(el.edges).To(HaveLen(2)) // gc-ed

	el.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	mi.reset()

	el.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(BeEmpty())
	mi.reset()

	el.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "myself", "edge to itself"}))
	mi.reset()

	// delete the remaining edges
	el.delEdge(edge{ // "prefix1/node2" -> "prefix1/node1"
		targetKey:  "prefix1/node1",
		isPrefix:   false,
		sourceNode: "prefix1/node2",
		relation:   "depends-on",
		label:      "node1 exists",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	el.delEdge(edge{ // "prefix2/node3" -> "prefix2/node3"
		targetKey:  "prefix2/node3",
		isPrefix:   false,
		sourceNode: "prefix2/node3",
		relation:   "myself",
		label:      "edge to itself",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	Expect(el.edges).To(BeEmpty()) // gc-ed

	el.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(BeEmpty())

	el.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(BeEmpty())

	el.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(BeEmpty())
}

func TestLookupOverEdgesWithOverlay(t *testing.T) {
	RegisterTestingT(t)

	mi := newMockIter()

	el := newEdgeLookup()
	Expect(el).ToNot(BeNil())

	el.addEdge(edge{ // "prefix1/node2" -> "prefix1/node1"
		targetKey:  "prefix1/node1",
		isPrefix:   false,
		sourceNode: "prefix1/node2",
		relation:   "depends-on",
		label:      "node1 exists",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	el.addEdge(edge{ // "prefix1/node2" -> "prefix2/node3"
		targetKey:  "prefix2/node3",
		isPrefix:   false,
		sourceNode: "prefix1/node2",
		relation:   "depends-on",
		label:      "node3 exists",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	el.addEdge(edge{ // "prefix2/node3" -> "prefix1/*"
		targetKey:  "prefix1/",
		isPrefix:   true,
		sourceNode: "prefix2/node3",
		relation:   "depends-on",
		label:      "prefix1 non-empty",
	})
	Expect(el.verifyDirDepthBounds()).To(BeNil())

	elOver := el.makeOverlay()
	Expect(elOver).ToNot(BeNil())
	Expect(elOver.underlay).To(Equal(el))
	Expect(el.overlay).To(Equal(elOver))
	Expect(elOver.verifyDirDepthBounds()).To(BeNil())

	elOver.addEdge(edge{ // "prefix2/node3" -> "prefix2/node3"
		targetKey:  "prefix2/node3",
		isPrefix:   false,
		sourceNode: "prefix2/node3",
		relation:   "myself",
		label:      "edge to itself",
	})
	Expect(elOver.verifyDirDepthBounds()).To(BeNil())

	elOver.addEdge(edge{
		targetKey:  "", // "prefix1/node1" -> *
		isPrefix:   true,
		sourceNode: "prefix1/node1",
		relation:   "all",
		label:      "all",
	})
	Expect(elOver.verifyDirDepthBounds()).To(BeNil())

	Expect(el.edges).To(HaveLen(3))
	Expect(elOver.edges).To(HaveLen(5))

	// check overlay
	elOver.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(3))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()
	elOver.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(2))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()
	elOver.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(3))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node3 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "myself", "edge to itself"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()

	// check underlay
	el.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(2))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	mi.reset()
	el.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	mi.reset()
	el.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node3 exists"}))
	mi.reset()

	// save and re-check underlay
	elOver.saveOverlay()
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	el.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(3))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()
	el.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(2))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()
	el.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(3))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node3 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "myself", "edge to itself"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()

	// delete 2 edges in overlay
	elOver.delEdge(edge{ // "prefix1/node2" -> "prefix2/node3"
		targetKey:  "prefix2/node3",
		isPrefix:   false,
		sourceNode: "prefix1/node2",
		relation:   "depends-on",
		label:      "node3 exists",
	})
	Expect(elOver.verifyDirDepthBounds()).To(BeNil())
	elOver.delEdge(edge{
		targetKey:  "", // "prefix1/node1" -> *
		isPrefix:   true,
		sourceNode: "prefix1/node1",
		relation:   "all",
		label:      "all",
	})
	Expect(elOver.verifyDirDepthBounds()).To(BeNil())
	Expect(elOver.edges).To(HaveLen(5)) // not gc-ed yet

	// check overlay
	elOver.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(2))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	mi.reset()
	elOver.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	mi.reset()
	elOver.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "myself", "edge to itself"}))
	mi.reset()

	// underlay before save
	el.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(3))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()
	el.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(2))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()
	el.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(3))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node3 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "myself", "edge to itself"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node1", "all", "all"}))
	mi.reset()

	// save and re-check underlay
	elOver.saveOverlay()
	Expect(el.verifyDirDepthBounds()).To(BeNil())
	el.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(2))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	mi.reset()
	el.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	mi.reset()
	el.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "myself", "edge to itself"}))
	mi.reset()

	// delete another edge, but do not save
	elOver.delEdge(edge{ // "prefix2/node3" -> "prefix1/*"
		targetKey:  "prefix1/",
		isPrefix:   true,
		sourceNode: "prefix2/node3",
		relation:   "depends-on",
		label:      "prefix1 non-empty",
	})
	Expect(elOver.verifyDirDepthBounds()).To(BeNil())
	Expect(elOver.edges).To(HaveLen(2)) // gc-ed
	Expect(el.edges).To(HaveLen(5))

	// check overlay
	elOver.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	mi.reset()
	elOver.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(BeEmpty())
	mi.reset()
	elOver.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "myself", "edge to itself"}))
	mi.reset()

	// throw away changes
	Expect(el.makeOverlay()).To(Equal(elOver))
	elOver = el.makeOverlay()
	elOver.saveOverlay()
	Expect(el.verifyDirDepthBounds()).To(BeNil())

	// check underlay
	el.iterSources("prefix1/node1", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(2))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix1/node2", "depends-on", "node1 exists"}))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	mi.reset()
	el.iterSources("prefix1/node2", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "depends-on", "prefix1 non-empty"}))
	mi.reset()
	el.iterSources("prefix2/node3", mi.visitEdge)
	Expect(mi.visitedEdges).To(HaveLen(1))
	Expect(mi.visitedEdges).To(HaveKey(visitedEdge{"prefix2/node3", "myself", "edge to itself"}))
	mi.reset()
}
