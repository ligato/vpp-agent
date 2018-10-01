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
	"github.com/gogo/protobuf/proto"

	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// nodeR implements Node.
type nodeR struct {
	graph *graphR

	key           string
	label         string
	value         proto.Message
	flags         []Flag
	metadata      interface{}
	metadataAdded bool
	metadataMap   string
	targetsDef    []RelationTarget
	targets       map[string]RecordedTargets // relation -> (label -> keys))
	sources       map[string]utils.KeySet    // relation -> nodes
}

// newNodeR creates a new instance of nodeR.
func newNodeR() *nodeR {
	return &nodeR{
		targets: make(map[string]RecordedTargets),
		sources: make(map[string]utils.KeySet),
	}
}

// GetKey returns the key associated with the node.
func (node *nodeR) GetKey() string {
	return node.key
}

// GetLabel returns the label associated with this node.
func (node *nodeR) GetLabel() string {
	return node.label
}

// GetKey returns the value associated with the node.
func (node *nodeR) GetValue() proto.Message {
	return node.value
}

// GetFlag returns reference to the given flag or nil if the node doesn't have
// this flag associated.
func (node *nodeR) GetFlag(name string) Flag {
	for _, flag := range node.flags {
		if flag.GetName() == name {
			return flag
		}
	}
	return nil
}

// GetMetadata returns the value metadata associated with the node.
func (node *nodeR) GetMetadata() interface{} {
	return node.metadata
}

// GetTargets returns a set of nodes, indexed by relation labels, that the
// edges of the given relation points to.
func (node *nodeR) GetTargets(relation string) RuntimeRelationTargets {
	runtimeTargets := make(RuntimeRelationTargets)

	targets, has := node.targets[relation]
	if !has {
		return runtimeTargets
	}

	for label, keys := range targets {
		var nodes []Node
		for key := range keys {
			nodes = append(nodes, node.graph.nodes[key])
		}
		runtimeTargets[label] = nodes
	}
	return runtimeTargets
}

// GetSources returns a set of nodes with edges of the given relation
// pointing to this node.
func (node *nodeR) GetSources(relation string) []Node {
	keys, has := node.sources[relation]
	if !has {
		return nil
	}
	var nodes []Node
	for key := range keys {
		nodes = append(nodes, node.graph.nodes[key])
	}
	return nodes
}

// copy returns a deep copy of the node.
func (node *nodeR) copy() *nodeR {
	nodeCopy := newNodeR()
	nodeCopy.key = node.key
	nodeCopy.label = node.label
	nodeCopy.value = node.value
	nodeCopy.metadata = node.metadata
	nodeCopy.metadataAdded = node.metadataAdded
	nodeCopy.metadataMap = node.metadataMap

	// copy flags
	for _, flag := range node.flags {
		nodeCopy.flags = append(nodeCopy.flags, flag)
	}

	// copy target definitions
	for _, targetDef := range node.targetsDef {
		nodeCopy.targetsDef = append(nodeCopy.targetsDef, targetDef)
	}

	// copy (runtime) targets
	for relation, targets := range node.targets {
		nodeCopy.targets[relation] = make(RecordedTargets)
		for label, keys := range targets {
			nodeCopy.targets[relation][label] = keys.DeepCopy()
		}
	}

	// copy sources
	for relation, keys := range node.sources {
		nodeCopy.sources[relation] = keys.DeepCopy()
	}
	return nodeCopy
}
