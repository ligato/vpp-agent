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
	prototypes "github.com/gogo/protobuf/types"
	"github.com/ligato/cn-infra/logging"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
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
	log         logging.Logger
	kvscheduler kvs.KVScheduler
	ifaceIdx    ifaceidx.IfaceMetadataIndex
}

// NewLinkStateDescriptor creates a new instance of the Link-State descriptor.
func NewLinkStateDescriptor(kvscheduler kvs.KVScheduler, ifaceIdx ifaceidx.IfaceMetadataIndex,
	log logging.PluginLogger) (descr *kvs.KVDescriptor, ctx *LinkStateDescriptor) {

	descrCtx := &LinkStateDescriptor{
		log:          log.NewLogger("interface-link-state"),
		kvscheduler:  kvscheduler,
		ifaceIdx:     ifaceIdx,
	}
	return &kvs.KVDescriptor{
		Name:                 LinkStateDescriptorName,
		KeySelector:          descrCtx.IsInterfaceLinkStateKey,
		Retrieve:             descrCtx.Retrieve,
		RetrieveDependencies: []string{InterfaceDescriptorName}, // link state read from interface metadata
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
	for _, ifaceName := range w.ifaceIdx.ListAllInterfaces() {
		ifaceMeta, _ := w.ifaceIdx.LookupByName(ifaceName)
		values = append(values, kvs.KVWithMetadata{
			Key:    interfaces.LinkStateKey(ifaceName, ifaceMeta.LinkIsUp),
			Value:  &prototypes.Empty{},
			Origin: kvs.FromSB,
		})
	}

	return values, nil
}

// UpdateLinkState notifies scheduler about a change in the link state of an interface.
func (w *LinkStateDescriptor) UpdateLinkState(ifaceState *interfaces.InterfaceNotification) {

	operStatus := ifaceState.State.OperStatus
	if operStatus == interfaces.InterfaceState_DELETED ||
		operStatus == interfaces.InterfaceState_UNKNOWN_STATUS {
		// interface link is neither up nor down
		w.kvscheduler.PushSBNotification(
			interfaces.LinkStateKey(ifaceState.State.Name, true),
			nil,
			nil)
		w.kvscheduler.PushSBNotification(
			interfaces.LinkStateKey(ifaceState.State.Name, false),
			nil,
			nil)
		return
	}

	w.kvscheduler.PushSBNotification(
		interfaces.LinkStateKey(ifaceState.State.Name, operStatus == interfaces.InterfaceState_UP),
		&prototypes.Empty{},
		nil)
}