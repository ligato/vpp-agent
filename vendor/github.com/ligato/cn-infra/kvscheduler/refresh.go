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
	"github.com/ligato/cn-infra/logging"
)

// resyncData stores data to be used for resync after refresh.
type resyncData struct {
	first  bool // true if startup-resync
	values []kvForTxn
}

// refreshGraph updates all/some values in the graph to their *real* state
// using the Dump methods from descriptors.
func (scheduler *Scheduler) refreshGraph(graphW graph.RWAccess, keys keySet, resyncData *resyncData) {
	refreshedKeys := make(keySet)

	// iterate over all descriptors, in order given by dump dependencies
	for _, descriptor := range scheduler.registry.GetAllDescriptors() {
		skip := len(keys) > 0
		for key := range keys {
			if descriptor.KeySelector(key) {
				skip = false
				break
			}
		}
		if skip {
			// nothing to refresh in the key space of this descriptor
			scheduler.skipRefresh(graphW, descriptor.GetName(), nil, refreshedKeys)
			continue
		}

		// get non-pending base values for this descriptor from memory before
		// refresh
		prevAddedNodes := graphW.GetNodes(nil,
			graph.WithFlags(&DescriptorFlag{descriptor.GetName()}),
			graph.WithoutFlags(&PendingFlag{}, &DerivedFlag{}))

		// get key-value pairs for correlation
		var correlate []KVWithMetadata
		if resyncData != nil && resyncData.first {
			// for startup resync, use data received from NB
			for _, kv := range resyncData.values {
				if descriptor.KeySelector(kv.key) {
					correlate = append(correlate,
						KVWithMetadata{
							Key:    kv.key,
							Value:  kv.value,
							Origin: kv.origin,
						})
				}
			}
		} else {
			// for refresh of failed values or run-time resync, use in-memory
			// kv-pairs for correlation
			correlate = nodesToKVPairsWithMetadata(prevAddedNodes)
		}

		// execute Dump operation
		dump, err := descriptor.Dump(correlate)
		if err != nil {
			if err != ErrDumpNotSupported {
				scheduler.Log.WithField("descriptor", descriptor.GetName()).
					Error("failed to dump values, refresh for the descriptor will be skipped")
			}
			// mark un-dumpable as refreshed
			scheduler.skipRefresh(graphW, descriptor.GetName(), nil, refreshedKeys)
			continue
		}

		if len(keys) > 0 {
			// mark keys that should not be touched as refreshed
			scheduler.skipRefresh(graphW, descriptor.GetName(), keys, refreshedKeys)
		}

		// process dumped kv-pairs
		for _, dumpedKV := range dump {
			if len(keys) > 0 {
				// do no touch values that aren't meant to be refreshed
				if _, toRefresh := keys[dumpedKV.Key]; !toRefresh {
					continue
				}
			}
			if !scheduler.validDumpedKV(dumpedKV, descriptor, refreshedKeys) {
				continue
			}

			// 1st attempt to determine value origin
			if dumpedKV.Origin == UnknownOrigin {
				// determine value origin based on the values for correlation
				for _, kv := range correlate {
					if kv.Key == dumpedKV.Key {
						dumpedKV.Origin = kv.Origin
						break
					}
				}
			}

			// 2nd attempt to determine value origin
			if dumpedKV.Origin == UnknownOrigin {
				// determine value origin based on the last revision
				timeline := graphW.GetNodeTimeline(dumpedKV.Key)
				if len(timeline) > 0 {
					lastRev := timeline[len(timeline)-1]
					originFlag := lastRev.Flags[OriginFlagName]
					if originFlag == FromNB.String() {
						dumpedKV.Origin = FromNB
					} else {
						dumpedKV.Origin = FromSB
					}
				}
			}

			if dumpedKV.Origin == UnknownOrigin {
				// will assume this is from SB
				dumpedKV.Origin = FromSB
			}

			// refresh node that represents this kv-pair
			node := graphW.SetNode(dumpedKV.Key)
			node.SetValue(dumpedKV.Value)
			if withMeta, _ := descriptor.WithMetadata(); node.GetValue().Type() == Object && withMeta {
				node.SetMetadataMap(descriptor.GetName())
				node.SetMetadata(dumpedKV.Metadata)
			}

			// refresh the tree of derived values + set flags
			scheduler.unwindDumpedRelations(graphW, node, dumpedKV.Origin, refreshedKeys)
		}

		// mark non-pending, non-derived values from NB that do not actually exist as pending
		for _, node := range prevAddedNodes {
			if _, refreshed := refreshedKeys[node.GetKey()]; !refreshed {
				if getNodeOrigin(node) == FromNB && getNodeLastChange(node).value != nil {
					missingNode := graphW.SetNode(node.GetKey())
					missingNode.SetFlags(&PendingFlag{})
					if node.GetValue().Type() == Object {
						missingNode.SetMetadata(nil)
					}
				}
			}
		}

		// in-progress save to expose changes in the metadata for dumps of the following
		// descriptors
		graphW.Save()
	}

	// remove nodes that do not actually exist
	refreshable := graphW.GetNodes(nil, graph.WithoutFlags(&PendingFlag{}))
	for _, node := range refreshable {
		if _, refreshed := refreshedKeys[node.GetKey()]; !refreshed {
			graphW.DeleteNode(node.GetKey())
		}
	}
}

// skipRefresh is used to mark nodes as refreshed without actual refreshing
// if they should not (or cannot) be refreshed.
func (scheduler *Scheduler) skipRefresh(graphR graph.ReadAccess, descriptor string, except keySet, refreshed keySet) {
	skipped := graphR.GetNodes(nil,
		graph.WithFlags(&DescriptorFlag{descriptor}),
		graph.WithoutFlags(&DerivedFlag{}))
	for _, node := range skipped {
		if _, toRefresh := except[node.GetKey()]; toRefresh {
			continue
		}
		refreshed[node.GetKey()] = struct{}{}

		// BFS over derived nodes
		derived := getDerivedNodes(node)
		for len(derived) > 0 {
			var next []graph.Node
			for _, derivedNode := range derived {
				refreshed[derivedNode.GetKey()] = struct{}{}
				next = append(next, getDerivedNodes(derivedNode)...)
			}
			derived = next
		}
	}
}

// unwindDumpedRelations builds a tree of derived values based on a dumped <root> kv-pair.
func (scheduler *Scheduler) unwindDumpedRelations(graphW graph.RWAccess, root graph.NodeRW,
	origin ValueOrigin, refreshed keySet) {

	// BFS over derived values
	nodes := []graph.NodeRW{root}
	for len(nodes) > 0 {
		var next []graph.NodeRW
		for _, node := range nodes {
			descriptor := scheduler.registry.GetDescriptorForKey(node.GetKey()) // nil for properties

			// refresh flags
			if node.GetKey() == root.GetKey() {
				node.DelFlags(DerivedFlagName)
			} else {
				if !scheduler.validDumpedDerivedKV(node, descriptor, refreshed) {
					graphW.DeleteNode(node.GetKey())
					continue
				}
				node.SetFlags(&DerivedFlag{})
			}
			node.SetFlags(&OriginFlag{origin})
			if descriptor != nil {
				node.SetFlags(&DescriptorFlag{descriptor.GetName()})
			} else {
				node.DelFlags(DescriptorFlagName)
			}
			node.DelFlags(PendingFlagName)

			// refresh relations with other values
			var dependencies []Dependency
			var derives []KeyValuePair
			if node.GetValue().Type() == Object || node.GetValue().Type() == Action {
				dependencies = descriptor.Dependencies(node.GetKey(), node.GetValue())
			}
			if node.GetValue().Type() == Object {
				derives = descriptor.DerivedValues(node.GetKey(), node.GetValue())
			}
			node.SetTargets(constructTargets(dependencies, derives))

			// add derived values for the next iteration
			for _, derived := range derives {
				nextNode := graphW.SetNode(derived.Key)
				nextNode.SetValue(derived.Value)
				next = append(next, nextNode)
			}
			refreshed[node.GetKey()] = struct{}{}
		}
		nodes = next
	}
}

// validDumpedKV verifies validity of a dumped KV-pair.
func (scheduler *Scheduler) validDumpedKV(kv KVWithMetadata, descriptor KVDescriptor, refreshed keySet) bool {
	if kv.Key == "" {
		scheduler.Log.WithFields(logging.Fields{
			"descriptor": descriptor.GetName(),
		}).Warn("Descriptor dumped value with empty key")
		return false
	}
	if _, alreadyDumped := refreshed[kv.Key]; alreadyDumped {
		scheduler.Log.WithFields(logging.Fields{
			"descriptor": descriptor.GetName(),
			"key":        kv.Key,
		}).Warn("The same value was dumped more than once")
		return false
	}
	if kv.Value == nil {
		scheduler.Log.WithFields(logging.Fields{
			"descriptor": descriptor.GetName(),
			"key":        kv.Key,
		}).Warn("Descriptor dumped nil value")
		return false
	}
	if !descriptor.KeySelector(kv.Key) {
		scheduler.Log.WithFields(logging.Fields{
			"descriptor": descriptor.GetName(),
			"key":        kv.Key,
			"value":      kv.Value.Label(),
		}).Warn("Descriptor dumped value outside of its key space")
		return false
	}
	if kv.Value.Type() == Property {
		scheduler.Log.WithFields(logging.Fields{
			"descriptor": descriptor.GetName(),
			"key":        kv.Key,
			"value":      kv.Value.Label(),
		}).Warn("Descriptor dumped property value")
		return false
	}
	return true
}

// validDumpedKV verifies validity of a KV-pair derived from a dumped value.
func (scheduler *Scheduler) validDumpedDerivedKV(node graph.Node, descriptor KVDescriptor, refreshed keySet) bool {
	if node.GetValue() == nil {
		scheduler.Log.WithFields(logging.Fields{
			"descriptor": descriptor.GetName(),
			"key":        node.GetKey(),
		}).Warn("Derived nil value")
		return false
	}
	if descriptor == nil && node.GetValue().Type() != Property {
		scheduler.Log.WithFields(logging.Fields{
			"key":   node.GetKey(),
			"value": node.GetValue(),
		}).Warn("Skipping unimplemented derived value from dump")
		return false
	}
	if descriptor != nil && node.GetValue().Type() == Property {
		scheduler.Log.WithFields(logging.Fields{
			"descriptor": descriptor.GetName(),
			"key":        node.GetKey(),
			"value":      node.GetValue(),
		}).Warn("Skipping property value with descriptor")
		return false
	}
	if _, alreadyDumped := refreshed[node.GetKey()]; alreadyDumped {
		scheduler.Log.WithFields(logging.Fields{
			"descriptor": descriptor.GetName(),
			"key":        node.GetKey(),
		}).Warn("The same value was dumped more than once")
		// return true -> let's overwrite invalidly dumped derived value
	}
	return true
}
