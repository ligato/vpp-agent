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
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// Order by operations (in average should yield the shortest sequence of operations):
//  1. delete
//  2. modify with re-create
//  3. add
//  4. modify
//
// Furthermore, operations of the same type are ordered by dependencies to limit
// temporary pending states.
// Dependencies are calculated only *approximately* - ordering at this stage is just
// an *optimization*, purpose of which is to decrease the length of the transaction plan.
func (scheduler *Scheduler) orderValuesByOp(graphR graph.ReadAccess, values []kvForTxn) []kvForTxn {

	// consider at least the first-level of derived values
	derived := make(map[string]utils.KeySet) // base value key -> derived keys
	valueByKey := make(map[string]kvForTxn)
	for _, kv := range values {
		valueByKey[kv.key] = kv
		descriptor := scheduler.registry.GetDescriptorForKey(kv.key)
		handler := &descriptorHandler{descriptor}
		node := graphR.GetNode(kv.key)

		var derivedKVs []kvs.KeyValuePair
		if kv.value != nil {
			derivedKVs = handler.derivedValues(kv.key, kv.value)
		} else if node != nil {
			derivedKVs = handler.derivedValues(kv.key, node.GetValue())
		}

		derived[kv.key] = utils.NewKeySet(kv.key) // include the base key itself
		for _, derivedKV := range derivedKVs {
			derived[kv.key].Add(derivedKV.Key)
		}
	}

	// sort values by operations and collect dependencies among changed values
	recreate := utils.NewKeySet()
	add := utils.NewKeySet()
	modify := utils.NewKeySet()
	delete := utils.NewKeySet()
	deps := make(map[string]utils.KeySet)
	for _, kv := range values {
		descriptor := scheduler.registry.GetDescriptorForKey(kv.key)
		handler := &descriptorHandler{descriptor}
		node := graphR.GetNode(kv.key)

		var valDeps []kvs.Dependency
		if kv.value != nil {
			valDeps = handler.dependencies(kv.key, kv.value)
		} else if node != nil {
			valDeps = handler.dependencies(kv.key, node.GetValue())
		}
		deps[kv.key] = utils.NewKeySet()
		for _, kv2 := range values {
			for kv2DerKey := range derived[kv2.key] {
				for _, dep := range valDeps {
					if kv2DerKey == dep.Key || (dep.AnyOf != nil && dep.AnyOf(kv2DerKey)) {
						deps[kv.key].Add(kv2.key)
					}
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
	orderedKeys = append(orderedKeys, utils.TopologicalOrder(delete, deps, false, true)...)
	orderedKeys = append(orderedKeys, utils.TopologicalOrder(recreate, deps, true, true)...)
	orderedKeys = append(orderedKeys, utils.TopologicalOrder(add, deps, true, true)...)
	orderedKeys = append(orderedKeys, utils.TopologicalOrder(modify, deps, true, true)...)

	// return values in the same order as keys are in <orderedKeys>
	var ordered []kvForTxn
	for _, key := range orderedKeys {
		ordered = append(ordered, valueByKey[key])
	}
	return ordered
}
