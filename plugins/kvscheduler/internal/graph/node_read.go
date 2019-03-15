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
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// maximum number of flags allowed to have defined
const maxFlags = 8

// nodeR implements Node.
type nodeR struct {
	graph *graphR

	key           string
	label         string
	value         proto.Message
	flags         [maxFlags]Flag
	metadata      interface{}
	metadataAdded bool
	metadataMap   string

	// same length and corresponding order (lexicographically by relation+label)
	targets       Targets
	targetsDef    []RelationTargetDef

	sources       sources
}

// relationSources groups all sources for a single relation.
type relationSources struct {
	relation string
	sources  utils.KeySet
}

// sources is a slice of all sources, grouped by relations.
type sources []relationSources

// String returns human-readable string representation of sourcesByRelation.
func (s sources) String() string {
	str := "{"
	for idx, rs := range s {
		if idx > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s->%s", rs.relation, rs.sources.String())
	}
	str += "}"
	return str
}

// getSourcesForRelation returns sources (keys) for the given relation.
func (s sources) getSourcesForRelation(relation string) *relationSources {
	for idx := range s {
		if s[idx].relation == relation {
			return &s[idx]
		}
	}
	return nil
}

// newNodeR creates a new instance of nodeR.
func newNodeR() *nodeR {
	return &nodeR{
		targetsDef: newTargetsDef(nil),
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
func (node *nodeR) GetFlag(flagIndex int) Flag {
	return node.flags[flagIndex]
}

// GetMetadata returns the value metadata associated with the node.
func (node *nodeR) GetMetadata() interface{} {
	return node.metadata
}

// GetTargets returns a set of nodes, indexed by relation labels, that the
// edges of the given relation points to.
func (node *nodeR) GetTargets(relation string) (runtimeTargets RuntimeTargets) {
	for i := node.targets.RelationBegin(relation); i < len(node.targets); i++ {
		if node.targets[i].Relation != relation {
			break
		}
		var nodes []Node
		for _, key := range node.targets[i].MatchingKeys.Iterate() {
			nodes = append(nodes, node.graph.nodes[key])
		}
		runtimeTargets = append(runtimeTargets, RuntimeTarget{
			Label: node.targets[i].Label,
			Nodes: nodes,
		})
	}
	return runtimeTargets
}

// GetSources returns a set of nodes with edges of the given relation
// pointing to this node.
func (node *nodeR) GetSources(relation string) (nodes []Node) {
	pgraph := node.graph.parent
	if pgraph != nil && pgraph.methodTracker != nil {
		defer pgraph.methodTracker("Node.GetSources")()
	}

	relSources := node.sources.getSourcesForRelation(relation)
	if relSources == nil {
		return nil
	}

	for _, key := range relSources.sources.Iterate() {
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

	// copy flags (arrays are passed by value)
	nodeCopy.flags = node.flags

	// shallow-copy target definitions (immutable)
	nodeCopy.targetsDef = node.targetsDef

	// copy targets
	nodeCopy.targets = node.copyTargets()

	// copy sources
	nodeCopy.sources = make(sources, len(node.sources))
	copy(nodeCopy.sources, node.sources)
	for i := range nodeCopy.sources {
		nodeCopy.sources[i].sources = nodeCopy.sources[i].sources.CopyOnWrite()
	}

	return nodeCopy
}

func (node *nodeR) copyTargets() Targets {
	targets := make(Targets, len(node.targets))
	copy(targets, node.targets)
	for i := range targets {
		targets[i].MatchingKeys = targets[i].MatchingKeys.CopyOnWrite()
	}
	return targets
}