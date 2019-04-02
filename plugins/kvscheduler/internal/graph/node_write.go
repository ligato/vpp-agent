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
	"reflect"
	"sort"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

type node struct {
	*nodeR

	metaInSync     bool
	dataUpdated    bool
	targetsUpdated bool
	sourcesUpdated bool
}

// newNode creates a new instance of node, either built from the scratch or
// extending existing nodeR.
func newNode(nodeR *nodeR) *node {
	if nodeR == nil {
		return &node{
			nodeR:       newNodeR(),
			metaInSync:  true,
			dataUpdated: true, /* completely new node */
		}
	}
	return &node{
		nodeR:      nodeR,
		metaInSync: true,
	}
}

// SetLabel associates given label with this node.
func (node *node) SetLabel(label string) {
	node.label = label
	node.dataUpdated = true
}

// SetValue associates given value with this node.
func (node *node) SetValue(value proto.Message) {
	node.value = value
	node.dataUpdated = true
}

// SetFlags associates given flag with this node.
func (node *node) SetFlags(flags ...Flag) {
	for _, flag := range flags {
		node.flags[flag.GetIndex()] = flag
	}
	node.dataUpdated = true
}

// DelFlags removes given flag from this node.
func (node *node) DelFlags(flagIndexes ...int) {
	for _, idx := range flagIndexes {
		node.flags[idx] = nil
	}
	node.dataUpdated = true
}

// SetMetadataMap chooses metadata map to be used to store the association
// between this node's value label and metadata.
func (node *node) SetMetadataMap(mapName string) {
	if node.metadataMap == "" { // cannot be changed
		node.metadataMap = mapName
		node.dataUpdated = true
		node.metaInSync = false
		if !node.graph.wCopy {
			node.syncMetadata()
		}
	}
}

// SetMetadata associates given value metadata with this node.
func (node *node) SetMetadata(metadata interface{}) {
	node.metadata = metadata
	node.dataUpdated = true
	node.metaInSync = false
	if !node.graph.wCopy {
		node.syncMetadata()
	}
}

// syncMetadata applies metadata changes into the associated mapping.
func (node *node) syncMetadata() {
	if node.metaInSync {
		return
	}
	// update metadata map
	if mapping, hasMapping := node.graph.mappings[node.metadataMap]; hasMapping {
		if node.metadataAdded {
			if node.metadata == nil {
				mapping.Delete(node.label)
				node.metadataAdded = false
			} else {
				prevMeta, _ := mapping.GetValue(node.label)
				if !reflect.DeepEqual(prevMeta, node.metadata) {
					mapping.Update(node.label, node.metadata)
				}
			}
		} else if node.metadata != nil {
			mapping.Put(node.label, node.metadata)
			node.metadataAdded = true
		}
	}
	node.metaInSync = true
}

// SetTargets updates definitions of all edges pointing from this node.
func (node *node) SetTargets(targetsDef []RelationTargetDef) {

	pgraph := node.graph.parent
	if pgraph != nil && pgraph.methodTracker != nil {
		defer pgraph.methodTracker("Node.SetTargets")()
	}

	sort.Slice(targetsDef, func(i, j int) bool {
		_, order := targetsDef[i].Compare(targetsDef[j])
		return order == -1
	})

	var i,j int
	for i < len(targetsDef) || j < len(node.targetsDef) {
		var equal bool
		var order int
		if i < len(targetsDef) && j < len(node.targetsDef) {
			equal, order = targetsDef[i].Compare(node.targetsDef[j])
		} else if i < len(targetsDef) {
			equal = false
			order = -1
		} else {
			equal = false
			order = 1
		}
		if equal {
			if targetsDef[i].WithKeySelector() {
				// re-run key selector
				target := &node.targets[i]
				// -> remove obsolete targets
				var obsolete []string
				for _, key := range target.MatchingKeys.Iterate() {
					if targetsDef[i].Selector.KeySelector(key) == false {
						obsolete = append(obsolete, key)
					}
				}
				for _, key := range obsolete {
					target.MatchingKeys.Del(key)
					targetNode := node.graph.nodes[key]
					targetNode.removeFromSources(target.Relation, target.Label, node.key)
				}
				// -> check for new targets
				node.iterEveryEdge(targetsDef[i], func(key string) {
					targetNode := node.graph.nodes[key]
					node.addToTargets(targetNode, target)
				})
			}
			i++
			j++
			continue
		}

		// not equal, process the first in the order
		if order == 0 {
			// updated target definition
			target := &node.targets[i]
			target.ExpectedKey = expectedKey(targetsDef[i])
			// remove previous edges
			for _, key := range target.MatchingKeys.Iterate() {
				targetNode := node.graph.nodes[key]
				targetNode.removeFromSources(target.Relation, target.Label, node.key)
			}
			node.addDelEdges(node.targetsDef[j], true)
			// create new edges
			if !targetsDef[i].WithKeySelector() {
				target.MatchingKeys = utils.NewSingletonKeySet("")
			} else {
				// selector
				target.MatchingKeys = utils.NewSliceBasedKeySet()
			}
			node.addDelEdges(targetsDef[i], false)
			node.iterEveryEdge(targetsDef[i], func(key string) {
				targetNode := node.graph.nodes[key]
				node.addToTargets(targetNode, target)
			})
			i++
			j++
			continue
		}
		if order == -1 {
			// new target definition
			node.addDelEdges(targetsDef[i], false)
			node.addTargetEntry(i, targetsDef[i].Relation, targetsDef[i].Label,
				targetsDef[i].WithKeySelector())
			target := &node.targets[i]
			target.ExpectedKey = expectedKey(targetsDef[i])
			node.iterEveryEdge(targetsDef[i], func(key string) {
				targetNode := node.graph.nodes[key]
				node.addToTargets(targetNode, target)

			})
			i++
			continue
		}
		if order == 1 {
			// obsolete target definition
			target := &node.targets[i]
			for _, key := range target.MatchingKeys.Iterate() {
				targetNode := node.graph.nodes[key]
				targetNode.removeFromSources(target.Relation, target.Label, node.key)
			}
			node.addDelEdges(node.targetsDef[j], true)
			node.removeTargetEntry(i)
			j++
			continue
		}
	}

	node.targetsDef = targetsDef
	node.dataUpdated = true
	// check implementation:
	if len(node.targetsDef) != len(node.targets) {
		panic("SetTargets: len(node.targetsDef) != len(node.targets)")
	}
}

// addTargetEntry adds new target entry at the given index.
func (node *node) addTargetEntry(index int, relation, label string, withSelector bool) {
	node.targets = append(node.targets, Target{})
	if index < len(node.targets)-1 {
		copy(node.targets[index+1:], node.targets[index:])
	}
	node.targets[index].Relation = relation
	node.targets[index].Label = label
	node.targets[index].ExpectedKey = ""
	node.targets[index].MatchingKeys = utils.NewSliceBasedKeySet()
	if !withSelector {
		node.targets[index].MatchingKeys = utils.NewSingletonKeySet("")
	} else {
		// selector
		node.targets[index].MatchingKeys = utils.NewSliceBasedKeySet()
	}
}

// removeTargetEntry removes target entry at the given index
func (node *node) removeTargetEntry(index int) {
	if index < len(node.targets)-1 {
		copy(node.targets[index:], node.targets[index+1:])
	}
	node.targets = node.targets[0:len(node.targets)-1]
}

func (node *node) addDelEdges(target RelationTargetDef, del bool) {
	cb := node.graph.edgeLookup.addEdge
	if del {
		cb = node.graph.edgeLookup.delEdge
	}
	if target.Key != "" {
		cb(edge{
			targetKey:  target.Key,
			isPrefix:   false,
			sourceNode: node.key,
			relation:   target.Relation,
			label:      target.Label,
		})
	} else {
		for _, keyPrefix := range target.Selector.KeyPrefixes {
			cb(edge{
				targetKey:  keyPrefix,
				isPrefix:   true,
				sourceNode: node.key,
				relation:   target.Relation,
				label:      target.Label,
			})
		}
		if len(target.Selector.KeyPrefixes) == 0 {
			cb(edge{
				targetKey:  "",
				isPrefix:   true,
				sourceNode: node.key,
				relation:   target.Relation,
				label:      target.Label,
			})
		}
	}
}

// iterEveryEdge iterates over every outgoing edge.
func (node *node) iterEveryEdge(target RelationTargetDef, cb func(targetKey string)) {
	checkTarget := func(key string) {
		if !target.WithKeySelector() || target.Selector.KeySelector(key) == true {
			cb(key)
		}
	}
	if target.Key != "" {
		node.graph.edgeLookup.iterTargets(target.Key, false, checkTarget)
		return
	}
	if len(target.Selector.KeyPrefixes) == 0 {
		node.graph.edgeLookup.iterTargets("", true, checkTarget)
	}
	for _, keyPrefix := range target.Selector.KeyPrefixes {
		node.graph.edgeLookup.iterTargets(keyPrefix, true, checkTarget)
	}
}

// addToTargets adds node2 into the set of targets for this node.
// Sources of node2 are also updated accordingly.
func (node *node) addToTargets(node2 *node, target *Target) {
	// update targets of node
	updated := target.MatchingKeys.Add(node2.key)
	node.targetsUpdated = updated || node.targetsUpdated
	if !updated {
		return
	}

	// update sources of node2
	node2.addToSources(node, target)
}

// addToSources adds node2 into the set of sources for this node.
func (node *node) addToSources(node2 *node, target *Target) {
	s, idx := node.sources.GetTargetForLabel(target.Relation, target.Label)
	if s == nil {
		node.sources = append(node.sources, Target{})
		if idx < len(node.sources)-1 {
			copy(node.sources[idx+1:], node.sources[idx:])
		}
		node.sources[idx].Relation = target.Relation
		node.sources[idx].Label = target.Label
		node.sources[idx].MatchingKeys = utils.NewSliceBasedKeySet()
		s = &(node.sources[idx])
	}
	updated := s.MatchingKeys.Add(node2.key)
	node.sourcesUpdated = updated || node.sourcesUpdated
	if updated {
		node.graph.unsaved.Add(node.key)
	}
}

// removeFromTarget removes given key from the given target.
// Note: sources are not updated!
func (node *node) removeFromTarget(key, relation, label string) {
	target, _ := node.targets.GetTargetForLabel(relation, label)
	updated := target.MatchingKeys.Del(key)
	node.targetsUpdated = updated || node.targetsUpdated
	if updated {
		node.graph.unsaved.Add(node.key)
	}
}

// removeFromSources removes given key from the sources for the given relation.
func (node *node) removeFromSources(relation, label, key string) {
	t, idx := node.sources.GetTargetForLabel(relation, label)
	updated := t.MatchingKeys.Del(key)
	if updated {
		if t.MatchingKeys.Length() == 0 {
			if idx < len(node.sources)-1 {
				copy(node.sources[idx:], node.sources[idx+1:])
			}
			node.sources = node.sources[0:len(node.sources)-1]
		}
		node.sourcesUpdated = true
		node.graph.unsaved.Add(node.key)
	}
}

func expectedKey(target RelationTargetDef) (expKey string) {
	if target.Key != "" {
		return target.Key
	}
	for idx, prefix := range target.Selector.KeyPrefixes {
		if idx > 0 {
			expKey += " | "
		}
		expKey += prefix + "*"
	}
	return expKey
}