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
	"github.com/pkg/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

const (
	// RxPlacementDescriptorName is the name of the descriptor for the rx-placement
	// config-subsection of VPP interfaces.
	RxPlacementDescriptorName = "vpp-interface-rx-placement"
)

// RxPlacementDescriptor configures Rx placement for VPP interface queues.
type RxPlacementDescriptor struct {
	log       logging.Logger
	ifHandler vppcalls.InterfaceVppAPI
	ifIndex   ifaceidx.IfaceMetadataIndex
}

// NewRxPlacementDescriptor creates a new instance of RxPlacementDescriptor.
func NewRxPlacementDescriptor(ifHandler vppcalls.InterfaceVppAPI, ifIndex ifaceidx.IfaceMetadataIndex,
	log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &RxPlacementDescriptor{
		ifHandler: ifHandler,
		ifIndex:   ifIndex,
		log:       log.NewLogger("rx-placement-descriptor"),
	}

	typedDescr := &adapter.RxPlacementDescriptor{
		Name:            RxPlacementDescriptorName,
		KeySelector:     ctx.IsInterfaceRxPlacementKey,
		ValueComparator: ctx.EquivalentRxPlacement,
		ValueTypeName:   proto.MessageName(&interfaces.Interface{}),
		Create:          ctx.Create,
		Delete:          ctx.Delete,
		Dependencies:    ctx.Dependencies,
	}

	return adapter.NewRxPlacementDescriptor(typedDescr)
}

// IsInterfaceRxPlacementKey returns true if the key is identifying RxPlacement
// configuration.
func (d *RxPlacementDescriptor) IsInterfaceRxPlacementKey(key string) bool {
	_, _, isValid := interfaces.ParseRxPlacementKey(key)
	return isValid
}

// EquivalentRxMode compares Rx placements for equivalency.
func (d *RxPlacementDescriptor) EquivalentRxPlacement(key string,
	oldRxPl, newRxPl *interfaces.Interface_RxPlacement) bool {

	if (oldRxPl.MainThread != newRxPl.MainThread) ||
		(!oldRxPl.MainThread && oldRxPl.Worker != newRxPl.Worker) {
		return false
	}
	return true
}

// Create configures RxPlacement for a given interface queue.
// Please note the proto message Interface is only used as container for RxMode.
// Only interface name, type and Rx mode are set.
func (d *RxPlacementDescriptor) Create(key string, rxPlacement *interfaces.Interface_RxPlacement) (metadata interface{}, err error) {
	ifaceName, _, _ := interfaces.ParseRxPlacementKey(key)
	ifMeta, found := d.ifIndex.LookupByName(ifaceName)
	if !found {
		err = errors.Errorf("failed to find interface %s", ifaceName)
		d.log.Error(err)
		return nil, err
	}

	if err = d.ifHandler.SetRxPlacement(ifMeta.SwIfIndex, rxPlacement); err != nil {
		err = errors.Errorf("failed to set rx-placement for queue %d of the interface %s: %v",
			rxPlacement.Queue, ifaceName, err)
		d.log.Error(err)
		return nil, err
	}
	return nil, err
}

// Delete is NOOP (Rx placement cannot be returned back to default).
func (d *RxPlacementDescriptor) Delete(key string, rxPlacement *interfaces.Interface_RxPlacement, metadata interface{}) error {
	return nil
}

// Dependencies informs scheduler that Rx placement configuration cannot be applied
// until the interface link is UP.
func (d *RxPlacementDescriptor) Dependencies(key string, rxPlacement *interfaces.Interface_RxPlacement) (deps []kvs.Dependency) {
	ifaceName, _, _ := interfaces.ParseRxPlacementKey(key)
	return []kvs.Dependency{
		{
			Label: linkIsUpDep,
			Key:   interfaces.LinkStateKey(ifaceName, true),
		},
	}
}