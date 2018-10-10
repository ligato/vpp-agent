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
	. "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

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
		txnOp, err := getNodeError(node)
		kvPairs = append(kvPairs, KeyWithError{
			Key:          node.GetKey(),
			TxnOperation: txnOp,
			Error:        err,
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
			Label:    derived.Key,
			Key:      derived.Key,
			Selector: nil,
		}
		targets = append(targets, target)
	}

	return targets
}

// getNodeOrigin returns node origin stored in Origin flag.
func getNodeOrigin(node graph.Node) ValueOrigin {
	flag := node.GetFlag(OriginFlagName)
	if flag != nil {
		return flag.(*OriginFlag).origin
	}
	return UnknownOrigin
}

// getNodeError returns node error stored in Error flag.
func getNodeError(node graph.Node) (operation TxnOperation, err error) {
	errorFlag := node.GetFlag(ErrorFlagName)
	if errorFlag != nil {
		flag := errorFlag.(*ErrorFlag)
		err = flag.err
		operation = flag.txnOp
		return
	}
	return UndefinedTxnOp, nil
}

// getNodeLastChange returns info about the last change for a given node, stored in LastChange flag.
func getNodeLastChange(node graph.Node) *LastChangeFlag {
	flag := node.GetFlag(LastChangeFlagName)
	if flag == nil {
		return nil
	}
	return flag.(*LastChangeFlag)
}

// getNodeLastUpdate returns info about the last update for a given node, stored in LastChange flag.
func getNodeLastUpdate(node graph.Node) *LastUpdateFlag {
	flag := node.GetFlag(LastUpdateFlagName)
	if flag == nil {
		return nil
	}
	return flag.(*LastUpdateFlag)
}

func isNodeDerived(node graph.Node) bool {
	return node.GetFlag(DerivedFlagName) != nil
}

func isNodePending(node graph.Node) bool {
	return node.GetFlag(PendingFlagName) != nil
}

// isNodeReady return true if the given node has all dependencies satisfied.
// Recursive calls are needed to handle circular dependencies - nodes of a strongly
// connected component are treated as if they were squashed into one.
func isNodeReady(node graph.Node) bool {
	if getNodeOrigin(node) == FromSB {
		// for SB values dependencies are not checked
		return true
	}
	ready, _ := isNodeReadyRec(node, 0, make(map[string]int))
	return ready
}

// isNodeReadyRec is a recursive call from within isNodeReady.
// visited = map{ key -> depth }
func isNodeReadyRec(node graph.Node, depth int, visited map[string]int) (ready bool, cycleDepth int) {
	if targetDepth, wasVisited := visited[node.GetKey()]; wasVisited {
		return true, targetDepth
	}
	cycleDepth = depth
	visited[node.GetKey()] = depth
	defer delete(visited, node.GetKey())

	for _, targets := range node.GetTargets(DependencyRelation) {
		satisfied := false
		for _, target := range targets {
			if isNodeBeingRemoved(target) {
				// do not consider values that are about to be removed
				continue
			}
			if !isNodePending(target) {
				satisfied = true
			}

			// test if node is inside a strongly-connected component (treated as one node)
			targetReady, targetCycleDepth := isNodeReadyRec(target, depth+1, visited)
			if targetReady && targetCycleDepth <= depth {
				// this node is reachable from the target
				satisfied = true
				if targetCycleDepth < cycleDepth {
					// update how far back in the branch this node can reach following dependencies
					cycleDepth = targetCycleDepth
				}
			}
		}
		if !satisfied {
			return false, cycleDepth
		}
	}
	return true, cycleDepth
}

// isNodeBeingRemoved returns true for a given node if it is being removed
// by a transaction or a notification (including failed removal attempt).
func isNodeBeingRemoved(node graph.Node) bool {
	base := node
	if isNodeDerived(node) {
		for {
			derivedFrom := base.GetSources(DerivesRelation)
			if len(derivedFrom) == 0 {
				break
			}
			base = derivedFrom[0]
			if isNodePending(base) {
				// one of the values from which this derives is pending
				return true
			}
		}
		if isNodeDerived(base) {
			// derived without base -> it is being removed by Modify()
			return true
		}
	}
	if getNodeLastChange(base) != nil && getNodeLastChange(base).value == nil {
		// about to be removed by transaction
		return true
	}
	return false
}

func canNodeHaveMetadata(node graph.Node) bool {
	return !isNodeDerived(node)
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

func getDerivedKeys(node graph.Node) utils.KeySet {
	set := utils.NewKeySet()
	for _, derived := range getDerivedNodes(node) {
		set.Add(derived.GetKey())
	}
	return set
}
