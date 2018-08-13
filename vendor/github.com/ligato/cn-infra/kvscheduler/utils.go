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

package kvscheduler

import (
	. "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/kvscheduler/graph"
)

type keySet map[string]struct{}

// subtract removes keys from <ks> that are in both key sets.
func (ks keySet) subtract(ks2 keySet) keySet {
	for key := range ks2 {
		delete(ks, key)
	}
	return ks
}

func nodesToKVPairs(nodes []graph.Node) (kvPairs []KeyValuePair) {
	for _, node := range nodes {
		kvPairs = append(kvPairs, KeyValuePair{
			Key:   node.GetKey(),
			Value: node.GetValue()})
	}
	return kvPairs
}

func nodesToKeysWithError(nodes []graph.Node) (kvPairs []KeyWithError) {
	for _, node := range nodes {
		kvPairs = append(kvPairs, KeyWithError{
			Key:   node.GetKey(),
			Error: getNodeError(node),
		})
	}
	return kvPairs
}

func nodesToKVPairsWithMetadata(nodes []graph.Node) (kvPairs []KVWithMetadata) {
	for _, node := range nodes {
		kvPairs = append(kvPairs, KVWithMetadata{
			Key:      node.GetKey(),
			Value:    node.GetValue(),
			Metadata: node.GetMetadata(),
			Origin:   getNodeOrigin(node),
		})
	}
	return kvPairs
}

// constructTargets builds targets for the graph based on derived values and dependencies.
func constructTargets(deps []Dependency, derives []KeyValuePair) (targets []graph.RelationTarget) {
	for _, dep := range deps {
		target := graph.RelationTarget{
			Relation: DependencyRelation,
			Label:    dep.Label,
			Key:      dep.Key,
			Selector: dep.AnyOf,
		}
		targets = append(targets, target)
	}

	for _, derived := range derives {
		target := graph.RelationTarget{
			Relation: DerivesRelation,
			Label:    derived.Value.Label(),
			Key:      derived.Key,
			Selector: nil,
		}
		targets = append(targets, target)
	}

	return targets
}

// dependsOn returns true if k1 depends on k2 based on dependencies from <deps>.
func dependsOn(k1, k2 string, deps map[string]keySet, valCount int, depth int) bool {
	if depth == valCount {
		panic("Dependency cycle!")
	}
	k1Deps := deps[k1]
	if _, depends := k1Deps[k2]; depends {
		return true
	}
	for dep := range k1Deps {
		if dependsOn(dep, k2, deps, valCount, depth+1) {
			return true
		}
	}
	return false
}

func getNodeOrigin(node graph.Node) ValueOrigin {
	flag := node.GetFlag(OriginFlagName)
	if flag != nil {
		return flag.(*OriginFlag).origin
	}
	return UnknownOrigin
}

func getNodeError(node graph.Node) error {
	errorFlag := node.GetFlag(ErrorFlagName)
	if errorFlag != nil {
		return errorFlag.(*ErrorFlag).err
	}
	return nil
}

func getNodeLastChange(node graph.Node) *LastChangeFlag {
	return node.GetFlag(LastChangeFlagName).(*LastChangeFlag)
}

func isNodeDerived(node graph.Node) bool {
	return node.GetFlag(DerivedFlagName) != nil
}

func isNodePending(node graph.Node) bool {
	return node.GetFlag(PendingFlagName) != nil
}

func isNodeReady(node graph.Node) bool {
	if getNodeOrigin(node) == FromSB {
		// for SB values dependencies are not checked
		return true
	}
	for _, targets := range node.GetTargets(DependencyRelation) {
		satisfied := false
		for _, target := range targets {
			if !isNodePending(target) {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return false
		}
	}
	return true
}

func canNodeHaveMetadata(node graph.Node) bool {
	return !isNodeDerived(node) && node.GetValue().Type() == Object
}

func getNodeBase(node graph.Node) graph.Node {
	derivedFrom := node.GetSources(DerivesRelation)
	if len(derivedFrom) == 0 {
		return node
	}
	return getNodeBase(derivedFrom[0])
}

func getDerivedNodes(node graph.Node) (derived []graph.Node) {
	for _, derivedNodes := range node.GetTargets(DerivesRelation) {
		for _, derivedNode := range derivedNodes {
			derived = append(derived, derivedNode)
		}
	}
	return derived
}

func getDerivedKeys(node graph.Node) keySet {
	set := make(keySet)
	for _, derived := range getDerivedNodes(node) {
		set[derived.GetKey()] = struct{}{}
	}
	return set
}
