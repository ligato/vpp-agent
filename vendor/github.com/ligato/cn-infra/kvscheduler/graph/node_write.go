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
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

type node struct {
	*nodeR

	metaInSync     bool
	dataUpdated    bool
	targetsUpdated bool
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

// SetValue associates given value with this node.
func (node *node) SetValue(value Value) {
	node.value = value
	node.dataUpdated = true
}

// SetFlags associates given flag with this node.
func (node *node) SetFlags(flags ...Flag) {
	toBeSet := make(map[string]struct{})
	for _, flag := range flags {
		toBeSet[flag.GetName()] = struct{}{}
	}

	var otherFlags []Flag
	for _, flag := range node.flags {
		if _, set := toBeSet[flag.GetName()]; !set {
			otherFlags = append(otherFlags, flag)
		}
	}

	node.flags = append(otherFlags, flags...)
	node.dataUpdated = true
}

// DelFlags removes given flag from this node.
func (node *node) DelFlags(names ...string) {
	var otherFlags []Flag
	for _, flag := range node.flags {
		delete := false
		for _, flagName := range names {
			if flag.GetName() == flagName {
				delete = true
				break
			}
		}
		if !delete {
			otherFlags = append(otherFlags, flag)
		}
	}

	node.flags = otherFlags
	node.dataUpdated = true
}

// SetMetadataMap chooses metadata map to be used to store the association
// between this node's value label and metadata.
func (node *node) SetMetadataMap(mapName string) {
	if node.metadataMap == "" { // cannot be changed
		node.metadataMap = mapName
		node.dataUpdated = true
		node.metaInSync = false
	}
}

// SetMetadata associates given value metadata with this node.
func (node *node) SetMetadata(metadata interface{}) {
	node.metadata = metadata
	node.dataUpdated = true
	node.metaInSync = false
}

// SetTargets provides definition of all edges pointing from this node.
func (node *node) SetTargets(targets []RelationTarget) {
	node.targetsDef = targets
	node.dataUpdated = true

	// remove from sources of current targets
	node.removeThisFromSources()

	// re-init targets
	node.initRuntimeTarget()

	// build new targets
	for _, otherNode := range node.graph.nodes {
		if otherNode.key == node.key {
			continue
		}
		node.checkPotentialTarget(otherNode)
	}
}

// initRuntimeTarget re-initialize targets to empty key-sets.
func (node *node) initRuntimeTarget() {
	node.targets = make(map[string]RecordedTargets)

	for _, targetDef := range node.targetsDef {
		if _, hasRelation := node.targets[targetDef.Relation]; !hasRelation {
			node.targets[targetDef.Relation] = make(RecordedTargets)
		}
		if _, hasLabel := node.targets[targetDef.Relation][targetDef.Label]; !hasLabel {
			node.targets[targetDef.Relation][targetDef.Label] = make(KeySet)
		}
	}
}

// checkPotentialTarget checks if node2 is target of node in any of the relations.
func (node *node) checkPotentialTarget(node2 *node) {
	for _, targetDef := range node.targetsDef {
		if targetDef.Key == node2.key || (targetDef.Key == "" && targetDef.Selector(node2.key)) {
			node.targets[targetDef.Relation][targetDef.Label][node2.key] = struct{}{}
			node.targetsUpdated = true
			if _, hasRelation := node2.sources[targetDef.Relation]; !hasRelation {
				node2.sources[targetDef.Relation] = make(KeySet)
			}
			node2.sources[targetDef.Relation][node.key] = struct{}{}
		}
	}
}

// removeFromTargets removes given key from the map of targets.
func (node *node) removeFromTargets(key string) {
	for relation, targets := range node.targets {
		for label := range targets {
			if _, has := node.targets[relation][label][key]; has {
				delete(node.targets[relation][label], key)
				node.targetsUpdated = true
			}
		}
	}
}

// removeFromTargets removes this node from the set of sources of all the other nodes.
func (node *node) removeThisFromSources() {
	for relation, targets := range node.targets {
		for _, targetNodes := range targets {
			for key := range targetNodes {
				targetNode := node.graph.nodes[key]
				delete(targetNode.sources[relation], node.GetKey())
			}
		}
	}
}