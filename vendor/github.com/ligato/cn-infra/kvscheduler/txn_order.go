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
	"sort"
)

// Order by operations (in average should yield the shortest sequence of operations):
//  1. modify with re-create
//  2. add
//  3. modify
//  4. delete
//
// Furthermore, operations of the same type are ordered by dependencies to limit
// temporary pending states.
func (scheduler *Scheduler) orderValuesByOp(graphR graph.ReadAccess, values []kvForTxn) []kvForTxn {
	var recreateVals, addVals, modifyVals, deleteVals []kvForTxn
	deps := make(map[string]keySet)

	for _, kv := range values {
		descriptor := scheduler.registry.GetDescriptorForKey(kv.key)
		node := graphR.GetNode(kv.key)

		// collect dependencies among changed values
		var valDeps []Dependency
		if kv.value != nil {
			valDeps = descriptor.Dependencies(kv.key, kv.value)
		} else if node != nil {
			valDeps = descriptor.Dependencies(kv.key, node.GetValue())
		}
		deps[kv.key] = make(keySet)
		for _, kv2 := range values {
			for _, dep := range valDeps {
				if kv2.key == dep.Key || (dep.AnyOf != nil && dep.AnyOf(kv2.key)) {
					deps[kv.key][kv2.key] = struct{}{}
				}
			}
		}

		if kv.value == nil {
			deleteVals = append(deleteVals, kv)
			continue
		}
		if node == nil || node.GetFlag(PendingFlagName) != nil {
			addVals = append(addVals, kv)
			continue
		}
		if descriptor.ModifyHasToRecreate(kv.key, node.GetValue(), kv.value, node.GetMetadata()) {
			recreateVals = append(recreateVals, kv)
		} else {
			modifyVals = append(modifyVals, kv)
		}
	}

	scheduler.orderValuesByDeps(recreateVals, deps, true)
	scheduler.orderValuesByDeps(addVals, deps, true)
	scheduler.orderValuesByDeps(modifyVals, deps, true)
	scheduler.orderValuesByDeps(deleteVals, deps, false)

	var ordered []kvForTxn
	ordered = append(ordered, recreateVals...)
	ordered = append(ordered, addVals...)
	ordered = append(ordered, modifyVals...)
	ordered = append(ordered, deleteVals...)
	return ordered
}

func (scheduler *Scheduler) orderValuesByDeps(values []kvForTxn, deps map[string]keySet, depFirst bool) {
	sort.Slice(values, func(i, j int) bool {
		iDepOnJ := dependsOn(values[i].key, values[j].key, deps, len(values), 0)
		jDepOnI := dependsOn(values[j].key, values[i].key, deps, len(values), 0)
		if depFirst {
			return jDepOnI || (!iDepOnJ && values[i].key < values[j].key)

		}
		return iDepOnJ || (!jDepOnI && values[i].key < values[j].key)
	})
}
