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
	targetsDef    *targetsDef
	targets       TargetsByRelation
	sources       sourcesByRelation
}

// targetsDef implements effective lookups over target definitions.
type targetsDef struct {
	defs []RelationTargetDef

	selectors  []RelationTargetDef
	staticKeys map[string][]RelationTargetDef
	byLabel    map[relationLabel]RelationTargetDef
}

// relationLabel groups relation + target label.
type relationLabel struct {
	relation, label string
}

// relationSources groups all sources for a single relation.
type relationSources struct {
	relation string
	sources  utils.KeySet
}

// sourcesByRelation is a slice of all sources, grouped by relations.
type sourcesByRelation []*relationSources

// newTargetsDef is a constructor for targetsDef.
func newTargetsDef(defs []RelationTargetDef) *targetsDef {
	tdef := &targetsDef{
		defs:       defs,
		staticKeys: make(map[string][]RelationTargetDef),
		byLabel:    make(map[relationLabel]RelationTargetDef),
	}
	for _, def := range defs {
		tdef.byLabel[relationLabel{relation: def.Relation, label: def.Label}] = def
		if def.Key == "" && def.Selector != nil {
			tdef.selectors = append(tdef.selectors, def)
		} else {
			if _, hasKey := tdef.staticKeys[def.Key]; !hasKey {
				tdef.staticKeys[def.Key] = []RelationTargetDef{}
			}
			tdef.staticKeys[def.Key] = append(tdef.staticKeys[def.Key], def)
		}
	}
	return tdef
}

// getDefinition retrieves definition for the given relation and label.
func (td *targetsDef) getForLabel(relation, label string) (def RelationTargetDef, exists bool) {
	def, exists = td.byLabel[relationLabel{relation: relation, label: label}]
	return
}

// getDefinition retrieves definition(s) selecting the given key.
func (td *targetsDef) getForKey(relation string, key string) (defs []RelationTargetDef) {
	if staticDefs, hasStaticDefs := td.staticKeys[key]; hasStaticDefs {
		for _, def := range staticDefs {
			if relation == "" || def.Relation == relation {
				defs = append(defs, def)
			}
		}
	}
	for _, def := range td.selectors {
		if relation != "" && def.Relation != relation {
			continue
		}
		if def.Selector(key) {
			defs = append(defs, def)
		}
	}
	return defs
}

// String returns human-readable string representation of sourcesByRelation.
func (s sourcesByRelation) String() string {
	str := "{"
	for idx, sources := range s {
		if idx > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s->%s", sources.relation, sources.sources.String())
	}
	str += "}"
	return str
}

// getSourcesForRelation returns sources (keys) for the given relation.
func (s sourcesByRelation) getSourcesForRelation(relation string) *relationSources {
	for _, relSources := range s {
		if relSources.relation == relation {
			return relSources
		}
	}
	return nil
}

// newNodeR creates a new instance of nodeR.
func newNodeR() *nodeR {
	return &nodeR{}
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
func (node *nodeR) GetTargets(relation string) (runtimeTargets RuntimeTargetsByLabel) {
	relTargets := node.targets.GetTargetsForRelation(relation)
	if relTargets == nil {
		return nil
	}
	for _, targets := range relTargets.Targets {
		var nodes []Node
		for _, key := range targets.MatchingKeys.Iterate() {
			nodes = append(nodes, node.graph.nodes[key])
		}
		runtimeTargets = append(runtimeTargets, &RuntimeTargets{
			Label: targets.Label,
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
	nodeCopy.sources = make(sourcesByRelation, 0, len(node.sources))
	for _, relSources := range node.sources {
		nodeCopy.sources = append(nodeCopy.sources, &relationSources{
			relation: relSources.relation,
			sources:  relSources.sources.CopyOnWrite(),
		})
	}
	return nodeCopy
}

func (node *nodeR) copyTargets() TargetsByRelation {
	tCopy := make(TargetsByRelation, 0, len(node.targets))
	for _, relTargets := range node.targets {
		targets := make(TargetsByLabel, 0, len(relTargets.Targets))
		for _, target := range relTargets.Targets {
			targets = append(targets, &Targets{
				Label:        target.Label,
				ExpectedKey:  target.ExpectedKey,
				MatchingKeys: target.MatchingKeys.CopyOnWrite(),
			})
		}
		tCopy = append(tCopy, &RelationTargets{
			Relation: relTargets.Relation,
			Targets:  targets,
		})
	}
	return tCopy
}
