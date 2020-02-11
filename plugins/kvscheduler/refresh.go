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
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/graph"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/internal/utils"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

var enableGraphDump = os.Getenv("KVSCHEDULER_GRAPHDUMP") != ""

const (
	nodeVisitBeginMark = "[BEGIN]"
	nodeVisitEndMark   = "[END]"
)

// resyncData stores data to be used for resync after refresh.
type resyncData struct {
	first  bool // true if startup-resync
	values []kvForTxn
}

// refreshGraph updates all/some values in the graph to their *real* state
// using the Retrieve methods from descriptors.
func (s *Scheduler) refreshGraph(graphW graph.RWAccess,
	keys utils.KeySet, resyncData *resyncData, verbose bool,
) {
	if s.logGraphWalk {
		keysToRefresh := "<ALL>"
		if keys != nil && keys.Length() > 0 {
			keysToRefresh = keys.String()
		}
		msg := fmt.Sprintf("refreshGrap (keys=%s)", keysToRefresh)
		fmt.Printf("%s %s\n", nodeVisitBeginMark, msg)
		defer fmt.Printf("%s %s\n", nodeVisitEndMark, msg)
	}
	refreshedKeys := utils.NewMapBasedKeySet()

	// iterate over all descriptors, in order given by retrieve dependencies
	for _, descriptor := range s.registry.GetAllDescriptors() {
		handler := newDescriptorHandler(descriptor)

		// get base values for this descriptor from memory before refresh
		// (including those marked as unavailable which may need metadata update)
		descrNodes := graphW.GetNodes(nil,
			descrValsSelectors(descriptor.Name, false)...)

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
			s.skipRefresh(descrNodes, nil, refreshedKeys)
			continue
		}

		// get key-value pairs for correlation
		var correlateCap int
		if resyncData != nil && resyncData.first {
			correlateCap = len(resyncData.values)
		} else {
			correlateCap = len(descrNodes)
		}
		correlate := make([]kvs.KVWithMetadata, 0, correlateCap)
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
			for _, node := range descrNodes {
				if isNodeAvailable(node) {
					correlate = append(correlate, nodeToKVPairWithMetadata(node))
				}
			}
		}

		// execute Retrieve operation
		retrieved, ableToRetrieve, err := handler.retrieve(correlate)

		// mark un-retrievable as refreshed
		if !ableToRetrieve || err != nil {
			l := s.Log.WithField("descriptor", descriptor.Name)
			if err != nil {
				l.Errorf("failed to retrieve values: %v", err)
				l.Debugf("skipping refresh for the descriptor")
			}
			s.skipRefresh(descrNodes, nil, refreshedKeys)
			continue
		} else if verbose {
			plural := "s"
			if len(retrieved) == 1 {
				plural = ""
			}

			var list strings.Builder
			for i, d := range retrieved {
				num := fmt.Sprintf("%d.", i+1)
				list.WriteString(fmt.Sprintf("\n - %3s [%s]: %q (%s)\n   %v",
					num, descriptor.Name, d.Key, d.Origin, utils.ProtoToString(d.Value)))
				if d.Metadata != nil {
					list.WriteString(fmt.Sprintf("\n   Metadata: %+v", d.Metadata))
				}
			}
			s.Log.Debugf("%s descriptor retrieved %d item%s: %v",
				descriptor.Name, len(retrieved), plural, list.String())

		}

		if keys != nil && keys.Length() > 0 {
			// mark keys that should not be touched as refreshed
			s.skipRefresh(descrNodes, keys, refreshedKeys)
		}

		// process retrieved kv-pairs
		for _, retrievedKV := range retrieved {
			if keys != nil && keys.Length() > 0 {
				// do no touch values that aren't meant to be refreshed
				if toRefresh := keys.Has(retrievedKV.Key); !toRefresh {
					continue
				}
			}
			if !s.validRetrievedKV(retrievedKV, descriptor, refreshedKeys) {
				continue
			}

			// 1st attempt to determine value origin
			if retrievedKV.Origin == kvs.UnknownOrigin {
				// determine value origin based on the values for correlation
				for _, kv := range correlate {
					if kv.Key == retrievedKV.Key {
						retrievedKV.Origin = kv.Origin
						break
					}
				}
			}

			// 2nd attempt to determine value origin
			if retrievedKV.Origin == kvs.UnknownOrigin {
				// determine value origin based on the last revision
				timeline := graphW.GetNodeTimeline(retrievedKV.Key)
				if len(timeline) > 0 {
					lastRev := timeline[len(timeline)-1]
					valueStateFlag := lastRev.Flags.GetFlag(ValueStateFlagIndex)
					valueState := valueStateFlag.(*ValueStateFlag).valueState
					retrievedKV.Origin = valueStateToOrigin(valueState)
				}
			}

			if retrievedKV.Origin == kvs.UnknownOrigin {
				// will assume this is from SB
				retrievedKV.Origin = kvs.FromSB
			}

			// refresh node that represents this kv-pair
			s.refreshValue(graphW, retrievedKV, handler, refreshedKeys, 2)
		}

		// unset the metadata from base NB values that do not actually exists
		for _, node := range descrNodes {
			if refreshed := refreshedKeys.Has(node.GetKey()); !refreshed {
				if getNodeOrigin(node) == kvs.FromNB && node.GetMetadata() != nil {
					if s.logGraphWalk {
						fmt.Printf("  -> unset metadata for key=%s\n", node.GetKey())
					}
					missingNode := graphW.SetNode(node.GetKey())
					missingNode.SetMetadata(nil)
				}
			}
		}
	}

	// update state of values that do not actually exist
	for _, node := range graphW.GetNodes(nil) {
		if refreshed := refreshedKeys.Has(node.GetKey()); refreshed {
			continue
		}
		s.refreshUnavailNode(graphW, node, refreshedKeys, 2)
	}

	if enableGraphDump && verbose && s.config.PrintTxnSummary {
		fmt.Println(dumpGraph(graphW))
	}
}

// refreshValue refreshes node that represents the given retrieved key-value pair.
func (s *Scheduler) refreshValue(graphW graph.RWAccess, retrievedKV kvs.KVWithMetadata,
	handler *descriptorHandler, refreshed utils.KeySet, indent int) {
	if s.logGraphWalk {
		indentStr := strings.Repeat(" ", indent)
		msg := fmt.Sprintf("refreshValue (key=%s)", retrievedKV.Key)
		fmt.Printf("%s%s %s\n", indentStr, nodeVisitBeginMark, msg)
		defer fmt.Printf("%s%s %s\n", indentStr, nodeVisitEndMark, msg)
	}

	// refresh node that represents this kv-pair
	node := graphW.SetNode(retrievedKV.Key)
	node.SetLabel(handler.keyLabel(node.GetKey()))
	node.SetValue(retrievedKV.Value)
	if handler.descriptor.WithMetadata {
		node.SetMetadataMap(handler.descriptor.Name)
		node.SetMetadata(retrievedKV.Metadata)
	}
	s.refreshAvailNode(graphW, node, retrievedKV.Origin, false, node.GetKey(), refreshed, indent+2)

	// determine the set of unavailable derived values
	obsolete := getDerivedKeys(node)
	derives := handler.derivedValues(node.GetKey(), node.GetValue())
	for _, newDerived := range derives {
		obsolete.Del(newDerived.Key)
	}

	// keep obsolete derived values still in the relation
	for _, key := range obsolete.Iterate() {
		derives = append(derives, kvs.KeyValuePair{Key: key}) // value unused
	}

	// refresh relations
	dependencies := handler.dependencies(node.GetKey(), node.GetValue())
	node.SetTargets(constructTargets(dependencies, derives))

	// refresh derived values
	for _, kv := range derives {
		isObsolete := obsolete.Has(kv.Key)
		derNode := graphW.SetNode(kv.Key)
		if !isObsolete {
			derDescr := s.registry.GetDescriptorForKey(kv.Key)
			derHandler := newDescriptorHandler(derDescr)
			derNode.SetValue(kv.Value)
			dependencies := derHandler.dependencies(derNode.GetKey(), derNode.GetValue())
			derNode.SetTargets(constructTargets(dependencies, nil))
			s.refreshAvailNode(graphW, derNode, retrievedKV.Origin, true, node.GetKey(), refreshed, indent+2)
		} else {
			s.refreshUnavailNode(graphW, derNode, refreshed, indent+2)
		}
	}
}

// refreshAvailNode refreshes state of a node whose value was returned by Retrieve.
func (s *Scheduler) refreshAvailNode(graphW graph.RWAccess, node graph.NodeRW,
	origin kvs.ValueOrigin, derived bool, baseKey string, refreshed utils.KeySet, indent int) {
	if s.logGraphWalk {
		indentStr := strings.Repeat(" ", indent)
		var derivedMark string
		if derived {
			derivedMark = ", is-derived"
		}
		msg := fmt.Sprintf("refreshAvailNode (key=%s%s)", node.GetKey(), derivedMark)
		fmt.Printf("%s%s %s\n", indentStr, nodeVisitBeginMark, msg)
		defer fmt.Printf("%s%s %s\n", indentStr, nodeVisitEndMark, msg)
	}

	// validate first
	descriptor := s.registry.GetDescriptorForKey(node.GetKey()) // nil for properties
	if derived && !s.validRetrievedDerivedKV(node, descriptor, refreshed) {
		graphW.DeleteNode(node.GetKey())
		return
	}

	// update availability
	if !isNodeAvailable(node) {
		s.updatedStates.Add(baseKey)
		node.DelFlags(UnavailValueFlagIndex)
	}
	refreshed.Add(node.GetKey())

	// refresh state
	if getNodeState(node) == kvscheduler.ValueState_NONEXISTENT {
		// newly found node
		if origin == kvs.FromSB {
			s.refreshNodeState(node, kvscheduler.ValueState_OBTAINED, indent)
		} else {
			s.refreshNodeState(node, kvscheduler.ValueState_DISCOVERED, indent)
		}
	}
	if getNodeState(node) == kvscheduler.ValueState_PENDING {
		// no longer pending apparently
		s.refreshNodeState(node, kvscheduler.ValueState_CONFIGURED, indent)
	}

	// update descriptor flag
	if descriptor != nil {
		node.SetFlags(&DescriptorFlag{descriptor.Name})
	} else {
		node.DelFlags(DescriptorFlagIndex)
	}

	// updated flags for derived values
	if !derived {
		node.DelFlags(DerivedFlagIndex)
	} else {
		node.SetFlags(&DerivedFlag{baseKey})
	}
}

// refreshUnavailNode refreshes state of a node whose value is found to be unavailable.
func (s *Scheduler) refreshUnavailNode(graphW graph.RWAccess, node graph.Node, refreshed utils.KeySet, indent int) {
	if s.logGraphWalk {
		indentStr := strings.Repeat(" ", indent)
		msg := fmt.Sprintf("refreshUnavailNode (key=%s, isDerived=%t)", node.GetKey(), isNodeDerived(node))
		fmt.Printf("%s%s %s\n", indentStr, nodeVisitBeginMark, msg)
		defer fmt.Printf("%s%s %s\n", indentStr, nodeVisitEndMark, msg)
	}

	refreshed.Add(node.GetKey())
	if isNodeAvailable(node) {
		s.updatedStates.Add(getNodeBaseKey(node))
	}
	state := getNodeState(node)
	if getNodeOrigin(node) == kvs.FromSB || state == kvscheduler.ValueState_DISCOVERED {
		// just remove from the graph
		graphW.DeleteNode(node.GetKey())
		return
	}

	// mark node as unavailable, but do not delete
	nodeW := graphW.SetNode(node.GetKey())
	if isNodeAvailable(node) {
		nodeW.SetFlags(&UnavailValueFlag{})
	}

	// update state
	if state == kvscheduler.ValueState_UNIMPLEMENTED {
		// it is expected that unimplemented value is not retrieved
		return
	}
	if state == kvscheduler.ValueState_CONFIGURED {
		if getNodeLastUpdate(node).value == nil {
			s.refreshNodeState(nodeW, kvscheduler.ValueState_REMOVED, indent)
		} else {
			s.refreshNodeState(nodeW, kvscheduler.ValueState_MISSING, indent)
		}
	}
}

func (s *Scheduler) refreshNodeState(node graph.NodeRW, newState kvscheduler.ValueState, indent int) {
	if getNodeState(node) != newState {
		if s.logGraphWalk {
			fmt.Printf("%s  -> change value state from %v to %v\n",
				strings.Repeat(" ", indent), getNodeState(node), newState)
		}
		node.SetFlags(&ValueStateFlag{valueState: newState})
	}
}

// skipRefresh is used to mark nodes as refreshed without actual refreshing
// if they should not (or cannot) be refreshed.
func (s *Scheduler) skipRefresh(nodes []graph.Node, except utils.KeySet, refreshed utils.KeySet) {
	for _, node := range nodes {
		if except != nil {
			if toRefresh := except.Has(node.GetKey()); toRefresh {
				continue
			}
		}
		refreshed.Add(node.GetKey())

		// skip refresh for derived nodes
		for _, derivedNode := range getDerivedNodes(node) {
			refreshed.Add(derivedNode.GetKey())
		}
	}
}

func dumpGraph(g graph.RWAccess) string {
	keys := g.GetKeys()

	var buf strings.Builder
	graphInfo := fmt.Sprintf("%d nodes", len(keys))
	buf.WriteString("+======================================================================================================================+\n")
	buf.WriteString(fmt.Sprintf("| GRAPH %110s |\n", graphInfo))
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
		if f := node.GetFlag(DescriptorFlagIndex); f != nil {
			descriptor = fmt.Sprintf("[%s] ", f.GetValue())
		}
		lastUpdate := "-"
		if f := node.GetFlag(LastUpdateFlagIndex); f != nil {
			lastUpdate = f.GetValue()
		}
		unavailable := ""
		if f := node.GetFlag(UnavailValueFlagIndex); f != nil {
			unavailable = "<UNAVAILABLE> "
		}
		writeLine(fmt.Sprintf("%s%s", descriptor, keyLabel), fmt.Sprintf("%s %s %s",
			unavailable,
			lastUpdate,
			getNodeState(node).String(),
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
				} else if len(nodeDeps) == 1 {
					writeLine(fmt.Sprintf(" - %s -> %v", dep.Label, strings.Join(nodeDeps, " ")), "")
				} else {
					writeLine(fmt.Sprintf(" - %s -> <UNAVAILABLE>", dep.Label), "")
				}
			}
		}
		if f := node.GetTargets(DerivesRelation); f != nil && len(f) > 0 {
			writeLine("Derives:", "")
			var nodeDers []string
			for _, der := range f {
				if len(der.Nodes) == 0 {
					nodeDers = append(nodeDers, fmt.Sprintf("%s", der.Label))
				} else {
					for _, node := range der.Nodes {
						desc := ""
						if d := node.GetFlag(DescriptorFlagIndex); d != nil {
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
			for _, perLabel := range f {
				for _, node := range perLabel.Nodes {
					desc := ""
					if d := node.GetFlag(DescriptorFlagIndex); d != nil {
						desc = fmt.Sprintf("[%s] ", d.GetValue())
					}
					nodeDeps = append(nodeDeps, fmt.Sprintf("%s%s", desc, node.GetKey()))
				}
			}
			writeLines(strings.Join(nodeDeps, "\n"), " - ")
		}
		if f := node.GetSources(DerivesRelation); len(f) > 0 {
			var nodeDers []string
			for _, perLabel := range f {
				for _, der := range perLabel.Nodes {
					nodeDers = append(nodeDers, der.GetKey())
				}
			}
			writeLine(fmt.Sprintf("Derived from: %s", strings.Join(nodeDers, " ")), "")
		}
		if f := node.GetMetadata(); f != nil {
			writeLine(fmt.Sprintf("Metadata: %+v", f), "")
		}
		if f := node.GetFlag(ErrorFlagIndex); f != nil {
			writeLine(fmt.Sprintf("Errors: %+v", f.GetValue()), "")
		}

		if i+1 != len(keys) {
			buf.WriteString("+----------------------------------------------------------------------------------------------------------------------+\n")
		}
	}
	buf.WriteString("+======================================================================================================================+\n")

	return buf.String()
}

// validRetrievedKV verifies validity of a retrieved KV-pair.
func (s *Scheduler) validRetrievedKV(kv kvs.KVWithMetadata, descriptor *kvs.KVDescriptor, refreshed utils.KeySet) bool {
	if kv.Key == "" {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptor.Name,
		}).Warn("Descriptor retrieved value with empty key")
		return false
	}
	if alreadyRetrieved := refreshed.Has(kv.Key); alreadyRetrieved {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptor.Name,
			"key":        kv.Key,
		}).Warn("The same value was retrieved more than once")
		return false
	}
	if kv.Value == nil {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptor.Name,
			"key":        kv.Key,
		}).Warn("Descriptor retrieved nil value")
		return false
	}
	if !descriptor.KeySelector(kv.Key) {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptor.Name,
			"key":        kv.Key,
			"value":      kv.Value,
		}).Warn("Descriptor retrieved value outside of its key space")
		return false
	}
	return true
}

// validRetrievedDerivedKV verifies validity of a KV-pair derived from a retrieved value.
func (s *Scheduler) validRetrievedDerivedKV(node graph.Node, descriptor *kvs.KVDescriptor, refreshed utils.KeySet) bool {
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
	if alreadyRetrieved := refreshed.Has(node.GetKey()); alreadyRetrieved {
		s.Log.WithFields(logging.Fields{
			"descriptor": descriptorName,
			"key":        node.GetKey(),
		}).Warn("The same value was retrieved more than once")
		// return true -> let's overwrite invalidly retrieved derived value
	}
	return true
}
