//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package descriptor

import (
	"fmt"

	vpp_abf "github.com/ligato/vpp-agent/api/models/vpp/abf"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin/abfidx"

	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"

	"github.com/go-errors/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
)

const (
	// ABFToInterfaceDescriptorName is name for descriptor
	ABFToInterfaceDescriptorName = "vpp-abf-to-interface"

	// dependency labels
	interfaceDep = "interface-exists"
)

// ABFToInterfaceDescriptor represents assignment of interface to ABF policy.
type ABFToInterfaceDescriptor struct {
	log        logging.Logger
	abfHandler vppcalls.ABFVppAPI
	abfIndex   abfidx.ABFMetadataIndex
	ifPlugin   ifplugin.API
}

// NewABFToInterfaceDescriptor returns new ABFInterface descriptor
func NewABFToInterfaceDescriptor(abfIndex abfidx.ABFMetadataIndex, abfHandler vppcalls.ABFVppAPI, ifPlugin ifplugin.API, log logging.PluginLogger) *ABFToInterfaceDescriptor {
	return &ABFToInterfaceDescriptor{
		log:        log,
		abfHandler: abfHandler,
		abfIndex:   abfIndex,
		ifPlugin:   ifPlugin,
	}
}

// GetDescriptor returns descriptor suitable for registration with the KVScheduler.
func (d *ABFToInterfaceDescriptor) GetDescriptor() *api.KVDescriptor {
	return &api.KVDescriptor{
		Name:         ABFToInterfaceDescriptorName,
		KeySelector:  d.IsABFInterfaceKey,
		Create:       d.Create,
		Delete:       d.Delete,
		Dependencies: d.Dependencies,
	}
}

// IsABFInterfaceKey returns true if the key is identifying ABF policy interface (derived value)
func (d *ABFToInterfaceDescriptor) IsABFInterfaceKey(key string) bool {
	_, _, isABFToInterfaceKey := vpp_abf.ParseABFToInterfaceKey(key)
	return isABFToInterfaceKey
}

// Create binds interface to ABF.
func (d *ABFToInterfaceDescriptor) Create(key string, emptyVal proto.Message) (metadata api.Metadata, err error) {
	// validate and get all required values
	isIPv6, abfIdx, ifIdx, priority, err := d.process(key)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	// attach interface to ABF policy
	if isIPv6 {
		return nil, d.abfHandler.AbfAttachInterfaceIPv6(abfIdx, ifIdx, priority)
	}
	return nil, d.abfHandler.AbfAttachInterfaceIPv4(abfIdx, ifIdx, priority)
}

// Delete unbinds interface from ABF.
func (d *ABFToInterfaceDescriptor) Delete(key string, emptyVal proto.Message, metadata api.Metadata) (err error) {
	// validate and get all required values
	isIPv6, abfIdx, ifIdx, priority, err := d.process(key)
	if err != nil {
		d.log.Error(err)
		return err
	}

	// detach interface to ABF policy
	if isIPv6 {
		return d.abfHandler.AbfDetachInterfaceIPv6(abfIdx, ifIdx, priority)
	}
	return d.abfHandler.AbfDetachInterfaceIPv4(abfIdx, ifIdx, priority)
}

// Dependencies lists the interface as the only dependency for the binding.
func (d *ABFToInterfaceDescriptor) Dependencies(key string, emptyVal proto.Message) []api.Dependency {
	_, ifName, _ := vpp_abf.ParseABFToInterfaceKey(key)
	return []api.Dependency{
		{
			Label: interfaceDep,
			Key:   vpp_interfaces.InterfaceKey(ifName),
		},
	}
}

// returns a bunch of values needed to attach/detach interface to/from ABF
func (d *ABFToInterfaceDescriptor) process(key string) (isIPv6 bool, abfIdx, ifIdx, priority uint32, err error) {
	// parse ABF and interface name
	abfIndex, ifName, isValid := vpp_abf.ParseABFToInterfaceKey(key)
	if !isValid {
		err = fmt.Errorf("ABF to interface key %s is not valid", key)
		return
	}
	// obtain ABF index
	abfData, exists := d.abfIndex.LookupByName(abfIndex)
	if !exists {
		err = errors.Errorf("failed to obtain metadata for ABF %s", abfIndex)
		return
	}

	// obtain interface index
	ifData, exists := d.ifPlugin.GetInterfaceIndex().LookupByName(ifName)
	if !exists {
		err = errors.Errorf("failed to obtain metadata for interface %s", ifName)
		return
	}

	// find other interface parameters from metadata
	for _, attachedIf := range abfData.Attached {
		if attachedIf.InputInterface == ifName {
			isIPv6, priority = attachedIf.IsIpv6, attachedIf.Priority
		}
	}
	return isIPv6, abfData.Index, ifData.SwIfIndex, priority, nil
}
