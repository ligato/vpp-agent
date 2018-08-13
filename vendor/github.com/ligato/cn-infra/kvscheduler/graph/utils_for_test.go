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
	. "github.com/onsi/gomega"
	"strings"

	. "github.com/ligato/cn-infra/kvscheduler/api"
	. "github.com/ligato/cn-infra/kvscheduler/test"
	"github.com/ligato/cn-infra/idxmap"
)

const (
	prefixA = "/prefixA/"
	prefixB = "/prefixB/"

	keyA1 = prefixA + "key1"
	keyA2 = prefixA + "key2"
	keyA3 = prefixA + "key3"
	keyB1 = prefixB + "key4"

	metadataMapA = "mapA"
	metadataMapB = "mapB"

	relation1 = "relation1"
	relation2 = "relation2"
)

var (
	value1 = NewStringValue(Object, "value1", "this is value1")
	value2 = NewStringValue(Object, "value2", "this is value2")
	value3 = NewStringValue(Action, "value3", "this is value3")
	value4 = NewStringValue(Property, "value4", "this is value4")
)

func prefixASelector(key string) bool {
	return strings.HasPrefix(key, prefixA)
}

func prefixBSelector(key string) bool {
	return strings.HasPrefix(key, prefixB)
}

func keySelector(keys ...string) func(key string) bool {
	return func(key string) bool {
		for _, k := range keys {
			if key == k {
				return true
			}
		}
		return false
	}
}

func selectNodesToBuild(ids ...int) map[int]struct{} {
	nodeIDs := make(map[int]struct{})
	for _, id := range ids {
		nodeIDs[id] = struct{}{}
	}
	return nodeIDs
}

func buildGraph(graph Graph, record, regMaps bool, nodes map[int]struct{}) Graph {
	if graph == nil {
		graph = NewGraph()
	}
	graphW := graph.Write(record)

	if regMaps {
		graphW.RegisterMetadataMap(metadataMapA, NewNameToInteger(metadataMapA))
		graphW.RegisterMetadataMap(metadataMapB, NewNameToInteger(metadataMapB))
	}

	if _, addNode1 := nodes[1]; addNode1 {
		node1 := graphW.SetNode(keyA1)
		node1.SetValue(value1)
		node1.SetMetadata(&OnlyInteger{Integer:1})
		node1.SetMetadataMap(metadataMapA)
		node1.SetFlags(ColorFlag(Red), AbstractFlag())
		node1.SetTargets([]RelationTarget{
			{relation1, "node2", keyA2, nil},
			{relation2, "prefixB", "", prefixBSelector},
		})
	}

	if _, addNode2 := nodes[2]; addNode2 {
		node2 := graphW.SetNode(keyA2)
		node2.SetValue(value2)
		node2.SetMetadata(&OnlyInteger{Integer:2})
		node2.SetMetadataMap(metadataMapA)
		node2.SetFlags(ColorFlag(Blue))
		node2.SetTargets([]RelationTarget{
			{relation1, "node3", keyA3, nil},
		})
	}

	if _, addNode3 := nodes[3]; addNode3 {
		node3 := graphW.SetNode(keyA3)
		node3.SetValue(value3)
		node3.SetMetadata(&OnlyInteger{Integer:3})
		node3.SetMetadataMap(metadataMapA)
		node3.SetFlags(ColorFlag(Green), AbstractFlag(), TemporaryFlag())
		node3.SetTargets([]RelationTarget{
			{relation2, "node1+node2", "", keySelector(keyA1, keyA2)},
			{relation2, "prefixB", "", prefixBSelector},
		})
	}

	if _, addNode4 := nodes[4]; addNode4 {
		node4 := graphW.SetNode(keyB1)
		node4.SetValue(value4)
		node4.SetMetadata(&OnlyInteger{Integer:1})
		node4.SetMetadataMap(metadataMapB)
		node4.SetFlags(TemporaryFlag())
		node4.SetTargets([]RelationTarget{
			{relation1, "prefixA", "", prefixASelector},
			{relation2, "non-existing-key", "non-existing-key", nil},
		})
	}

	graphW.Save()
	graphW.Release()
	return graph
}

func checkTargets(node Node, relation string, label string, targetKeys ...string) {
	targets := node.GetTargets(relation)
	Expect(targets).To(HaveKey(label))
	targetNodes := make(map[string]struct{})
	for _, targetNode := range targets[label] {
		targetNodes[targetNode.GetKey()] = struct{}{}
	}
	for _, targetKey := range targetKeys {
		Expect(targetNodes).To(HaveKey(targetKey))
	}
	Expect(targets[label]).To(HaveLen(len(targetKeys)))
}

func checkRecordedTargets(targets RecordedTargets, label string, targetKeys ...string) {
	Expect(targets).To(HaveKey(label))
	for _, targetKey := range targetKeys {
		Expect(targets[label]).To(HaveKey(targetKey))
	}
	Expect(targets[label]).To(HaveLen(len(targetKeys)))
}

func checkNodes(nodes []Node, keys ...string) {
	for _, key := range keys {
		found := false
		for _, node := range nodes {
			if node.GetKey() == key {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())
	}
	Expect(nodes).To(HaveLen(len(keys)))
}

func checkRecordedNodes(nodes []*RecordedNode, keys ...string) {
	recordedNodes := make(map[string]struct{})
	for _, node := range nodes {
		recordedNodes[node.Key] = struct{}{}
	}
	Expect(nodes).To(HaveLen(len(keys)))
}

func checkSources(node Node, relation string, sourceKeys ...string) {
	sourceNodes := make(map[string]struct{})
	for _, sourceNode := range node.GetSources(relation) {
		sourceNodes[sourceNode.GetKey()] = struct{}{}
	}
	for _, sourceKey := range sourceKeys {
		Expect(sourceNodes).To(HaveKey(sourceKey))
	}
	Expect(node.GetSources(relation)).To(HaveLen(len(sourceKeys)))
}

func checkMetadataValues(mapping idxmap.NamedMapping, labels ...string) {
	allLabels := make(map[string]struct{})
	for _, label := range mapping.ListAllNames() {
		allLabels[label] = struct{}{}
	}
	for _, label := range labels {
		Expect(allLabels).To(HaveKey(label))
	}
	Expect(mapping.ListAllNames()).To(HaveLen(len(labels)))
}
