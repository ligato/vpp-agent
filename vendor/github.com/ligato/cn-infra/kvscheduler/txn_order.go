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
	"github.com/ligato/cn-infra/kvscheduler/internal/graph"
	"github.com/ligato/cn-infra/kvscheduler/internal/utils"
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
	recreate := utils.NewKeySet()
	add := utils.NewKeySet()
	modify := utils.NewKeySet()
	delete := utils.NewKeySet()
	valueByKey := make(map[string]kvForTxn)
	deps := make(map[string]utils.KeySet)

	for _, kv := range values {
		valueByKey[kv.key] = kv
		descriptor := scheduler.registry.GetDescriptorForKey(kv.key)
		handler := &descriptorHandler{descriptor}
		node := graphR.GetNode(kv.key)

		// collect dependencies among changed values
		var valDeps []Dependency
		if kv.value != nil {
			valDeps = handler.dependencies(kv.key, kv.value)
		} else if node != nil {
			valDeps = handler.dependencies(kv.key, node.GetValue())
		}
		deps[kv.key] = utils.NewKeySet()
		for _, kv2 := range values {
			for _, dep := range valDeps {
				if kv2.key == dep.Key || (dep.AnyOf != nil && dep.AnyOf(kv2.key)) {
					deps[kv.key].Add(kv2.key)
				}
			}
		}

		if kv.value == nil {
			delete.Add(kv.key)
			continue
		}
		if node == nil || node.GetFlag(PendingFlagName) != nil {
			add.Add(kv.key)
			continue
		}
		if handler.modifyWithRecreate(kv.key, node.GetValue(), kv.value, node.GetMetadata()) {
			recreate.Add(kv.key)
		} else {
			modify.Add(kv.key)
		}
	}

	// order keys by operation + dependencies
	var orderedKeys []string
	orderedKeys = append(orderedKeys, utils.TopologicalOrder(recreate, deps, true, true)...)
	orderedKeys = append(orderedKeys, utils.TopologicalOrder(add, deps, true, true)...)
	orderedKeys = append(orderedKeys, utils.TopologicalOrder(modify, deps, true, true)...)
	orderedKeys = append(orderedKeys, utils.TopologicalOrder(delete, deps, false, true)...)

	// return values in the same order as keys are in <orderedKeys>
	var ordered []kvForTxn
	for _, key := range orderedKeys {
		ordered = append(ordered, valueByKey[key])
	}
	return ordered
}
