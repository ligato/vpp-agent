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
	"sort"
)

// orderValuesByOp orders values by operations (in average should yield the shortest
// sequence of operations):
//  1. delete
//  2. update with re-create
//  3. create
//  4. update
func (s *Scheduler) orderValuesByOp(values []kvForTxn) []kvForTxn {
	graphR := s.graph.Read()
	defer graphR.Release()

	// first order values alphabetically by keys to get deterministic behaviour and
	// output that is easier to read
	sort.Slice(values, func(i, j int) bool {
		return values[i].key < values[j].key
	})

	// sort values by operations
	var delete, recreate, create, update []kvForTxn
	for _, kv := range values {
		descriptor := s.registry.GetDescriptorForKey(kv.key)
		handler := &descriptorHandler{descriptor}
		node := graphR.GetNode(kv.key)

		if kv.value == nil {
			delete = append(delete, kv)
			continue
		}
		if node == nil || node.GetFlag(UnavailValueFlagName) != nil {
			create = append(create, kv)
			continue
		}
		if handler.updateWithRecreate(kv.key, node.GetValue(), kv.value, node.GetMetadata()) {
			recreate = append(recreate, kv)
		} else {
			update = append(update, kv)
		}
	}

	ordered := make([]kvForTxn, 0, len(values))
	ordered = append(ordered, delete...)
	ordered = append(ordered, recreate...)
	ordered = append(ordered, create...)
	ordered = append(ordered, update...)
	return ordered
}
