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
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/utils"
)

// resyncData stores data to be used for resync after refresh.
type resyncData struct {
	first   bool // true if startup-resync
	values  []kvForTxn
	verbose bool
}

// refreshGraph updates all/some values in the graph to their *real* state
// using the Dump methods from descriptors.
func (s *Scheduler) refreshGraph(graphW graph.RWAccess, keys utils.KeySet, resyncData *resyncData) {
	refreshedKeys := utils.NewMapBasedKeySet()

	// iterate over all descriptors, in order given by dump dependencies
	for _, descriptor := range s.registry.GetAllDescriptors() {
		handler := &descriptorHandler{descriptor}

		// check if this descriptor's key space should be refreshed as well
		var skip bool
		if keys != nil {
			skip = keys.Length() > 0
			for _, key := range keys.Iterate() {
				if descriptor.KeySelector(key) {
					skip = false
					break
				}
			}
		}
		if skip {
			// nothing to refresh in the key space of this descriptor
			s.skipRefresh(graphW, descriptor.Name, nil, refreshedKeys)
			continue
		}

		// get non-pending base values for this descriptor from memory before
		// refresh
		prevAddedNodes := graphW.GetNodes(nil,
			graph.WithFlags(&DescriptorFlag{descriptor.Name}),
			graph.WithoutFlags(&PendingFlag{}, &DerivedFlag{}))

		// get key-value pairs for correlation
		var correlate []kvs.KVWithMetadata
		if resyncData != nil && resyncData.first {
			// for startup resync, use data received from NB
			for _, kv := range resyncData.values {
				if descriptor.KeySelector(kv.key) {
					correlate = append(correlate,
						kvs.KVWithMetadata{
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
		dump, ableToDump, err := handler.dump(correlate)

		// mark un-dumpable as refreshed
		if !ableToDump || err != nil {
			if err != nil {
				s.Log.WithField("descriptor", descriptor.Name).
					Errorf("failed to dump values (skipping descriptor refresh): %s", err)
			}
			s.skipRefresh(graphW, descriptor.Name, nil, refreshedKeys)
			continue
		} else if resyncData == nil || resyncData.verbose {
			plural := "s"
			if len(dump) == 1 {
				plural = ""
			}

			var dumpList string
			for _, d := range dump {
				dumpList += fmt.Sprintf("\n - %+v", d)
			}

			s.Log.Debugf("Descriptor %s dumped %d item%s: %v",
				descriptor.Name, len(dump), plural, dumpList)

		}

		if keys != nil && keys.Length() > 0 {
			// mark keys that should not be touched as refreshed
			s.skipRefresh(graphW, descriptor.Name, keys, refreshedKeys)
		}

		// process dumped kv-pairs
		for _, dumpedKV := range dump {
			if keys != nil && keys.Length() > 0 {
				// do no touch values that aren't meant to be refreshed
				if toRefresh := keys.Has(dumpedKV.Key); !toRefresh {
					continue
				}
			}
			if !s.validDumpedKV(dumpedKV, descriptor, refreshedKeys) {
				continue
			}

			// 1st attempt to determine value origin
			if dumpedKV.Origin == kvs.UnknownOrigin {
				// determine value origin based on the values for correlation
				for _, kv := range correlate {
					if kv.Key == dumpedKV.Key {
						dumpedKV.Origin = kv.Origin
						break
					}
				}
			}

			// 2nd attempt to determine value origin
			if dumpedKV.Origin == kvs.UnknownOrigin {
				// determine value origin based on the last revision
				timeline := graphW.GetNodeTimeline(dumpedKV.Key)
				if len(timeline) > 0 {
					lastRev := timeline[len(timeline)-1]
					originFlag := lastRev.Flags.GetFlag(OriginFlagName)
					dumpedKV.Origin = originFlag.(*OriginFlag).origin
				}
			}

			if dumpedKV.Origin == kvs.UnknownOrigin {
				// will assume this is from SB
				dumpedKV.Origin = kvs.FromSB
			}

			// refresh node that represents this kv-pair
			node := graphW.SetNode(dumpedKV.Key)
			node.SetLabel(handler.keyLabel(node.GetKey()))
			node.SetValue(dumpedKV.Value)
			if descriptor.WithMetadata {
				node.SetMetadataMap(descriptor.Name)
				node.SetMetadata(dumpedKV.Metadata)
			}

			// refresh the tree of derived values + set flags
			s.unwindDumpedRelations(graphW, node, dumpedKV.Origin, refreshedKeys)
		}

		// mark non-pending, non-derived values from NB that do not actually exist as pending
		for _, node := range prevAddedNodes {
			if refreshed := refreshedKeys.Has(node.GetKey()); !refreshed {
				if getNodeOrigin(node) == kvs.FromNB && getNodeLastChange(node).value != nil {
					missingNode := graphW.SetNode(node.GetKey())
					missingNode.SetFlags(&PendingFlag{})
					missingNode.SetMetadata(nil)
				}
			}
		}

		// in-progress save to expose changes in the metadata for dumps of the following
		// descriptors
		graphW.Save()
	}

	// remove nodes that do not actually exist
	for _, node := range graphW.GetNodes(nil) {
		if !isNodeDerived(node) && isNodePending(node) {
			// keep pending base values
			continue
		}
		if refreshed := refreshedKeys.Has(node.GetKey()); !refreshed {
			graphW.DeleteNode(node.GetKey())
		}
	}

	if resyncData == nil || resyncData.verbose {
		fmt.Println(dumpGraph(graphW))
	}
}

// skipRefresh is used to mark nodes as refreshed without actual refreshing
// if they should not (or cannot) be refreshed.
func (s *Scheduler) skipRefresh(graphR graph.ReadAccess, descriptor string, except utils.KeySet, refreshed utils.KeySet) {
	skipped := graphR.GetNodes(nil,
		graph.WithFlags(&DescriptorFlag{descriptor}),
		graph.WithoutFlags(&DerivedFlag{}))
	for _, node := range skipped {
		if except != nil {
			if toRefresh := except.Has(node.GetKey()); toRefresh {
				continue
			}
		}
		refreshed.Add(node.GetKey())

		// BFS over derived nodes
		derived := getDerivedNodes(node)
		for len(derived) > 0 {
			var next []graph.Node
			for _, derivedNode := range derived {
				refreshed.Add(derivedNode.GetKey())
				next = append(next, getDerivedNodes(derivedNode)...)
			}
			derived = next
		}
	}
}

// unwindDumpedRelations builds a tree of derived values based on a dumped <root> kv-pair.
func (s *Scheduler) unwindDumpedRelations(graphW graph.RWAccess, root graph.NodeRW,
	origin kvs.ValueOrigin, refreshed utils.KeySet) {

	// BFS over derived values
	nodes := []graph.NodeRW{root}
	for len(nodes) > 0 {
		var next []graph.NodeRW
		for _, node := range nodes {
			descriptor := s.registry.GetDescriptorForKey(node.GetKey()) // nil for properties
			handler := descriptorHandler{descriptor}

			// refresh flags
			if node.GetKey() == root.GetKey() {
				node.DelFlags(DerivedFlagName)
			} else {
				if !s.validDumpedDerivedKV(node, descriptor, refreshed) {
					graphW.DeleteNode(node.GetKey())
					continue
				}
				node.SetFlags(&DerivedFlag{})
			}
			node.SetFlags(&OriginFlag{origin})
			if descriptor != nil {
				node.SetFlags(&DescriptorFlag{descriptor.Name})
			} else {
				node.DelFlags(DescriptorFlagName)
			}
			node.DelFlags(PendingFlagName)

			// refresh relations with other values
			dependencies := handler.dependencies(node.GetKey(), node.GetValue())
			derives := handler.derivedValues(node.GetKey(), node.GetValue())
			node.SetTargets(constructTargets(dependencies, derives))

			// add derived values for the next iteration
			for _, derived := range derives {
				nextNode := graphW.SetNode(derived.Key)
				nextNode.SetValue(derived.Value)
				next = append(next, nextNode)
			}
			refreshed.Add(node.GetKey())
		}
		nodes = next
	}
}

func dumpGraph(g graph.RWAccess) string {
	keys := g.GetKeys()

	var buf strings.Builder
	graphInfo := fmt.Sprintf("%d nodes", len(keys))
	buf.WriteString("+======================================================================================================================+\n")
	buf.WriteString(fmt.Sprintf("| GRAPH DUMP %105s |\n", graphInfo))
	buf.WriteString("+======================================================================================================================+\n")
	writeLine := func(left, right string) {
		n := 115 - len(left)
		buf.WriteString(fmt.Sprintf("| %s %"+fmt.Sprint(n)+"s |\n", left, right))

	}
	writeLines := func(linesStr string, prefix string) {
		lines := strings.Split(linesStr, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			writeLine(fmt.Sprintf("%s%s", prefix, line), "")
		}
	}
	for i, key := range keys {
		node := g.GetNode(key)
		keyLabel := key
		if label := node.GetLabel(); label != key && label != "" {
			keyLabel = fmt.Sprintf("%s (%s)", key, label)
		}
		descriptor := ""
		if f := node.GetFlag(DescriptorFlagName); f != nil {
			descriptor = fmt.Sprintf("[%s] ", f.GetValue())
		}
		lastChange := "-"
		if f := node.GetFlag(LastChangeFlagName); f != nil {
			lastChange = f.GetValue()
		}
		lastUpdate := "-"
		if f := node.GetFlag(LastUpdateFlagName); f != nil {
			lastUpdate = f.GetValue()
		}
		pending := ""
		if f := node.GetFlag(PendingFlagName); f != nil {
			pending = "<PENDING> "
		}
		writeLine(fmt.Sprintf("%s%s", descriptor, keyLabel), fmt.Sprintf("%s%s/%s %s",
			pending,
			lastChange, lastUpdate,
			node.GetFlag(OriginFlagName).GetValue(),
		))
		writeLines(proto.MarshalTextString(node.GetValue()), "  ")

		if f := node.GetTargets(DependencyRelation); f != nil && len(f) > 0 {
			writeLine("Depends on:", "")
			for _, dep := range f {
				var nodeDeps []string
				for _, node := range dep.Nodes {
					nodeDeps = append(nodeDeps, node.GetKey())
				}
				if len(nodeDeps) > 1 {
					writeLine(fmt.Sprintf(" - %s", dep.Label), "")
					writeLines(strings.Join(nodeDeps, "\n"), "  -> ")
				} else {
					writeLine(fmt.Sprintf(" - %s -> %v", dep.Label, strings.Join(nodeDeps, " ")), "")
				}
			}
		}
		if f := node.GetTargets(DerivesRelation); f != nil && len(f) > 0 {
			writeLine("Derives:", "")
			var nodeDers []string
			for _, der := range f {
				if len(der.Nodes) == 0 {
					nodeDers = append(nodeDers, "%s", der.Label)
				} else {
					for _, node := range der.Nodes {
						desc := ""
						if d := node.GetFlag(DescriptorFlagName); d != nil {
							desc = fmt.Sprintf("[%s] ", d.GetValue())
						}
						nodeDers = append(nodeDers, fmt.Sprintf("%s%s", desc, node.GetKey()))
					}
				}
			}
			writeLines(strings.Join(nodeDers, "\n"), " - ")
		}
		if f := node.GetSources(DependencyRelation); len(f) > 0 {
			writeLine("Dependency for:", "")
			var nodeDeps []string
			for _, node := range f {
				desc := ""
				if d := node.GetFlag(DescriptorFlagName); d != nil {
					desc = fmt.Sprintf("[%s] ", d.GetValue())
				}
				nodeDeps = append(nodeDeps, fmt.Sprintf("%s%s", desc, node.GetKey()))
			}
			writeLines(strings.Join(nodeDeps, "\n"), " - ")
		}
		if f := node.GetSources(DerivesRelation); len(f) > 0 {
			var nodeDers []string
			for _, der := range f {
				nodeDers = append(nodeDers, der.GetKey())
			}
			writeLine(fmt.Sprintf("Derived from: %s", strings.Join(nodeDers, " ")), "")
		}
		if f := node.GetMetadata(); f != nil {
			writeLine(fmt.Sprintf("Metadata: %+v", f), "")
		}
		if f := node.GetFlag(ErrorFlagName); f != nil {
			writeLine(fmt.Sprintf("Errors: %+v", f.GetValue()), "")
		}

		if i+1 != len(keys) {
			buf.WriteString("+======================================================================================================================+\n")
		}
	}
	buf.WriteString("+======================================================================================================================+\n")

	return buf.String()
}

// validDumpedKV verifies validity of a dumped KV-pair.
func (s *Scheduler) validDumpedKV(kv kvs.KVWithMetadata, descriptor *kvs.KVDescriptor, refreshed utils.KeySet) bool {
	if kv.Key == "" {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptor.Name,
		}).Warn("Descriptor dumped value with empty key")
		return false
	}
	if alreadyDumped := refreshed.Has(kv.Key); alreadyDumped {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptor.Name,
			"key":        kv.Key,
		}).Warn("The same value was dumped more than once")
		return false
	}
	if kv.Value == nil {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptor.Name,
			"key":        kv.Key,
		}).Warn("Descriptor dumped nil value")
		return false
	}
	if !descriptor.KeySelector(kv.Key) {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptor.Name,
			"key":        kv.Key,
			"value":      kv.Value,
		}).Warn("Descriptor dumped value outside of its key space")
		return false
	}
	return true
}

// validDumpedKV verifies validity of a KV-pair derived from a dumped value.
func (s *Scheduler) validDumpedDerivedKV(node graph.Node, descriptor *kvs.KVDescriptor, refreshed utils.KeySet) bool {
	descriptorName := "<NONE>"
	if descriptor != nil {
		descriptorName = descriptor.Name
	}
	if node.GetValue() == nil {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptorName,
			"key":        node.GetKey(),
		}).Warn("Derived nil value")
		return false
	}
	if alreadyDumped := refreshed.Has(node.GetKey()); alreadyDumped {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptorName,
			"key":        node.GetKey(),
		}).Warn("The same value was dumped more than once")
		// return true -> let's overwrite invalidly dumped derived value
	}
	return true
}
