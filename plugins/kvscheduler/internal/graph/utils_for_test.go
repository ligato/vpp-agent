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
	"strings"

	. "github.com/onsi/gomega"

	"go.ligato.io/cn-infra/v2/idxmap"

	. "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/test"
)

const (
	value1Label = "value1"
	value2Label = "value2"
	value3Label = "value3"
	value4Label = "value4"

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
	value1 = NewStringValue("this is value1")
	value2 = NewStringValue("this is value2")
	value3 = NewStringValue("this is value3")
	value4 = NewStringValue("this is value4")

	commonOpts = Opts{RecordOldRevs: true, RecordAgeLimit: minutesInOneDay, PermanentInitPeriod: minutesInOneHour}
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

func buildGraph(graph Graph, wInPlace bool, record, regMaps bool, nodes map[int]struct{}) Graph {
	if graph == nil {
		graph = NewGraph(commonOpts)
	}
	graphW := graph.Write(wInPlace, record)
	Expect(graphW.ValidateEdges()).To(BeNil())

	if regMaps {
		graphW.RegisterMetadataMap(metadataMapA, NewNameToInteger(metadataMapA))
		graphW.RegisterMetadataMap(metadataMapB, NewNameToInteger(metadataMapB))
	}

	var (
		node1, node2, node3, node4 NodeRW
	)

	if _, addNode1 := nodes[1]; addNode1 {
		node1 = graphW.SetNode(keyA1)
		node1.SetLabel(value1Label)
		node1.SetValue(value1)
		node1.SetMetadata(&OnlyInteger{Integer: 1})
		node1.SetMetadataMap(metadataMapA)
		node1.SetFlags(ColorFlag(Red), AbstractFlag())
		node1.SetTargets([]RelationTargetDef{
			{relation1, "node3", keyA3, TargetSelector{}},
			{relation2, "node2", keyA2, TargetSelector{}},
		})
		Expect(graphW.ValidateEdges()).To(BeNil())
		// targets changed
		node1.SetTargets([]RelationTargetDef{
			{relation1, "node2", keyA2, TargetSelector{}},
			{relation2, "prefixB", "", TargetSelector{KeySelector: prefixBSelector}},
		})
		Expect(graphW.ValidateEdges()).To(BeNil())
	}

	if _, addNode2 := nodes[2]; addNode2 {
		node2 = graphW.SetNode(keyA2)
		Expect(graphW.ValidateEdges()).To(BeNil())
		node2.SetLabel(value2Label)
		node2.SetValue(value2)
		node2.SetMetadata(&OnlyInteger{Integer: 2})
		node2.SetMetadataMap(metadataMapA)
		node2.SetFlags(ColorFlag(Blue))
		node2.SetTargets([]RelationTargetDef{
			{relation1, "node3", keyA1, TargetSelector{}},
		})
		Expect(graphW.ValidateEdges()).To(BeNil())
		// targets changed
		node2.SetTargets([]RelationTargetDef{
			{relation1, "node3", keyA3, TargetSelector{}},
		})
		Expect(graphW.ValidateEdges()).To(BeNil())
	}

	if _, addNode3 := nodes[3]; addNode3 {
		node3 = graphW.SetNode(keyA3)
		Expect(graphW.ValidateEdges()).To(BeNil())
		node3.SetLabel(value3Label)
		node3.SetValue(value3)
		node3.SetMetadata(&OnlyInteger{Integer: 3})
		node3.SetMetadataMap(metadataMapA)
		node3.SetFlags(ColorFlag(Green), AbstractFlag(), TemporaryFlag())
		node3.SetTargets([]RelationTargetDef{
			{relation2, "node1+node2", "", TargetSelector{KeySelector: keySelector(keyA1, keyA2)}},
			{relation2, "prefixB", keyB1, TargetSelector{}},
		})
		Expect(graphW.ValidateEdges()).To(BeNil())
		// targets changed
		node3.SetTargets([]RelationTargetDef{
			{relation2, "node1+node2", "", TargetSelector{KeySelector: keySelector(keyA1, keyA2)}},
			{relation2, "prefixB", "", TargetSelector{KeySelector: prefixBSelector}},
		})
		Expect(graphW.ValidateEdges()).To(BeNil())
	}

	if _, addNode4 := nodes[4]; addNode4 {
		node4 = graphW.SetNode(keyB1)
		Expect(graphW.ValidateEdges()).To(BeNil())
		node4.SetLabel(value4Label)
		node4.SetValue(value4)
		node4.SetMetadata(&OnlyInteger{Integer: 1})
		node4.SetMetadataMap(metadataMapB)
		node4.SetFlags(TemporaryFlag())
		node4.SetTargets([]RelationTargetDef{
			{relation1, "prefixA", "", TargetSelector{KeySelector: prefixASelector}},
			{relation2, "non-existing-key", "non-existing-key", TargetSelector{}},
			{relation2, "non-existing-key2", "non-existing-key2", TargetSelector{}},
		})
		Expect(graphW.ValidateEdges()).To(BeNil())
		// targets changed
		node4.SetTargets([]RelationTargetDef{
			{relation1, "prefixA", "", TargetSelector{KeySelector: prefixASelector}},
			{relation2, "non-existing-key", "non-existing-key", TargetSelector{}},
		})
		Expect(graphW.ValidateEdges()).To(BeNil())
	}

	if !wInPlace {
		graphW.Save()

		// make changes that will not be saved and thus should have no effect
		if node1 != nil {
			node1.SetTargets([]RelationTargetDef{
				{relation1, "node3", keyA3, TargetSelector{}},
				{relation2, "node2", keyA2, TargetSelector{}},
			})
			Expect(graphW.ValidateEdges()).To(BeNil())
		}
		if node3 != nil {
			node3.SetTargets([]RelationTargetDef{})
			Expect(graphW.ValidateEdges()).To(BeNil())
		}
		if node4 != nil {
			node4.SetTargets([]RelationTargetDef{
				{relation1, "prefixA", "use-key-instead-of-selector", TargetSelector{}},
				{relation2, "non-existing-key", keyA3, TargetSelector{}},
			})
			Expect(graphW.ValidateEdges()).To(BeNil())
		}
	}

	graphW.Release()
	return graph
}

func flags(flags ...Flag) (flagArray [maxFlags]Flag) {
	for _, f := range flags {
		flagArray[f.GetIndex()] = f
	}
	return
}

func checkTargets(node Node, relation string, label string, targetKeys ...string) {
	targets := node.GetTargets(relation)
	forLabel := targets.GetTargetForLabel(label)
	targetNodes := make(map[string]struct{})
	for _, targetNode := range forLabel.Nodes {
		targetNodes[targetNode.GetKey()] = struct{}{}
	}
	for _, targetKey := range targetKeys {
		Expect(targetNodes).To(HaveKey(targetKey))
	}
	Expect(targetNodes).To(HaveLen(len(targetKeys)))
}

func checkRecordedTargets(recordedTargets Targets, relation string, labelCnt int, label string, targetKeys ...string) {
	cnt := 0
	for i := recordedTargets.RelationBegin(relation); i < len(recordedTargets); i++ {
		if recordedTargets[i].Relation != relation {
			break
		}
		cnt++
	}
	Expect(cnt).To(Equal(labelCnt))
	t, _ := recordedTargets.GetTargetForLabel(relation, label)
	Expect(t).ToNot(BeNil())
	Expect(t.Label).To(Equal(label))
	for _, targetKey := range targetKeys {
		Expect(t.MatchingKeys.Has(targetKey)).To(BeTrue())
	}
	Expect(t.MatchingKeys.Length()).To(Equal(len(targetKeys)))
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
	for _, perLabel := range node.GetSources(relation) {
		for _, sourceNode := range perLabel.Nodes {
			sourceNodes[sourceNode.GetKey()] = struct{}{}
		}
	}
	for _, sourceKey := range sourceKeys {
		Expect(sourceNodes).To(HaveKey(sourceKey))
	}
	Expect(sourceNodes).To(HaveLen(len(sourceKeys)))
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
