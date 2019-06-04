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

package descriptor

import (
	"sync"

	prototypes "github.com/gogo/protobuf/types"
	"github.com/ligato/cn-infra/logging"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

const (
	// LinkStateDescriptorName is the name of the descriptor notifying about the
	// link state changes of VPP interfaces.
	LinkStateDescriptorName = "vpp-interface-link-state"
)

// LinkStateDescriptor notifies kvscheduler about the link state changes of VPP
// interfaces.
type LinkStateDescriptor struct {
	// input arguments
	log          logging.Logger
	kvscheduler  kvs.KVScheduler
	ifaceHandler vppcalls.InterfaceVppAPI
	ifaceIdx     ifaceidx.IfaceMetadataIndex

	linkStatesMx sync.Mutex
	linkStates   map[string]bool // interface name -> link is up
}

// NewLinkStateDescriptor creates a new instance of the Link-State descriptor.
func NewLinkStateDescriptor(kvscheduler kvs.KVScheduler, ifaceHandler vppcalls.InterfaceVppAPI,
	ifaceIdx ifaceidx.IfaceMetadataIndex, log logging.PluginLogger) (descr *kvs.KVDescriptor, ctx *LinkStateDescriptor) {

	descrCtx := &LinkStateDescriptor{
		log:          log.NewLogger("interface-link-state"),
		kvscheduler:  kvscheduler,
		ifaceHandler: ifaceHandler,
		ifaceIdx:     ifaceIdx,
		linkStates:   make(map[string]bool),
	}
	return &kvs.KVDescriptor{
		Name:                 LinkStateDescriptorName,
		KeySelector:          descrCtx.IsInterfaceLinkStateKey,
		Retrieve:             descrCtx.Retrieve,
		// Retrieve depends on the interface descriptor: interface index is used
		// to convert sw_if_index to logical interface name
		RetrieveDependencies: []string{InterfaceDescriptorName},
	}, descrCtx
}

// IsInterfaceLinkStateKey returns <true> for keys representing
// link-state of VPP interfaces.
func (w *LinkStateDescriptor) IsInterfaceLinkStateKey(key string) bool {
	_, _, isLinkStateKey := interfaces.ParseLinkStateKey(key)
	return isLinkStateKey
}

// Retrieve returns key for every VPP interface describing the state of the link
// (value is empty).
func (w *LinkStateDescriptor) Retrieve(correlate []kvs.KVWithMetadata) (values []kvs.KVWithMetadata, err error) {
	// TODO: avoid dumping interface details when it was already done in the interface
	//       descriptor within the same Refresh (e.g. during full resync)
	//       - e.g. add context to allow sharing of information across Retrieve-s of the same Refresh

	ifaceStates, err := w.ifaceHandler.DumpInterfaceStates()
	if err != nil {
		w.log.Error(err)
		return nil, err
	}

	w.linkStatesMx.Lock()
	defer w.linkStatesMx.Unlock()
	w.linkStates = make(map[string]bool) // clear the map

	for ifaceIdx, ifaceState := range ifaceStates {
		ifaceName, _, found := w.ifaceIdx.LookupBySwIfIndex(ifaceIdx)
		if !found {
			// skip interface not configured by NB (e.g. untouched Gbe interface)
			continue
		}
		linkIsUp := ifaceState.LinkState == interfaces.InterfaceState_UP
		w.linkStates[ifaceName] = linkIsUp
		values = append(values, kvs.KVWithMetadata{
			Key:    interfaces.LinkStateKey(ifaceName, linkIsUp),
			Value:  &prototypes.Empty{},
			Origin: kvs.FromSB,
		})
	}

	return values, nil
}

// UpdateLinkState notifies scheduler about a change in the link state of an interface.
func (w *LinkStateDescriptor) UpdateLinkState(ifaceState *interfaces.InterfaceNotification) {
	w.linkStatesMx.Lock()
	defer w.linkStatesMx.Unlock()

	w.log.Debugf("Updating link state: %+v", ifaceState)

	var notifs []kvs.KVWithMetadata

	operStatus := ifaceState.State.OperStatus
	ifaceName := ifaceState.State.Name
	linkWasUp, hadLinkState := w.linkStates[ifaceName]
	linkIsUp := operStatus == interfaces.InterfaceState_UP
	toDelete := operStatus == interfaces.InterfaceState_DELETED ||
		operStatus == interfaces.InterfaceState_UNKNOWN_STATUS

	if toDelete || (hadLinkState && (linkIsUp != linkWasUp)) {
		if hadLinkState {
			// remove now obsolete key-value pair
			notifs = append(notifs, kvs.KVWithMetadata{
				Key:      interfaces.LinkStateKey(ifaceState.State.Name, linkWasUp),
				Value:    nil,
				Metadata: nil,
			})
			delete(w.linkStates, ifaceName)
		}
	}

	if !toDelete && (!hadLinkState || (linkIsUp != linkWasUp)) {
		// push new key-value pair
		notifs = append(notifs, kvs.KVWithMetadata{
			Key:      interfaces.LinkStateKey(ifaceState.State.Name, linkIsUp),
			Value:    &prototypes.Empty{},
			Metadata: nil,
		})
		w.linkStates[ifaceName] = linkIsUp
	}

	if len(notifs) != 0 {
		err := w.kvscheduler.PushSBNotification(notifs...)
		if err != nil {
			w.log.Errorf("failed to send notifications to KVScheduler: %v", err)
		}
	}
}