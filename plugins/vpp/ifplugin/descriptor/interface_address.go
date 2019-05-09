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
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

const (
	// InterfaceAddressDescriptorName is the name of the descriptor for assigning
	// IP addresses to VPP interfaces.
	InterfaceAddressDescriptorName = "vpp-interface-address"

	// dependency labels
	interfaceInVrfDep = "interface-assigned-to-vrf-table"
)

// InterfaceAddressDescriptor (un)assigns IP address to/from VPP interface.
type InterfaceAddressDescriptor struct {
	log       logging.Logger
	ifHandler vppcalls.InterfaceVppAPI
	ifIndex   ifaceidx.IfaceMetadataIndex
}

// NewInterfaceAddressDescriptor creates a new instance of InterfaceAddressDescriptor.
func NewInterfaceAddressDescriptor(ifHandler vppcalls.InterfaceVppAPI, ifIndex ifaceidx.IfaceMetadataIndex,
	log logging.PluginLogger) *kvs.KVDescriptor {

	descrCtx := &InterfaceAddressDescriptor{
		ifHandler: ifHandler,
		ifIndex:   ifIndex,
		log:       log.NewLogger("interface-address-descriptor"),
	}
	return &kvs.KVDescriptor{
		Name:         InterfaceAddressDescriptorName,
		KeySelector:  descrCtx.IsInterfaceAddressKey,
		Validate:     descrCtx.Validate,
		Create:       descrCtx.Create,
		Delete:       descrCtx.Delete,
		Dependencies: descrCtx.Dependencies,
	}
}

// IsInterfaceVrfKey returns true if the key represents assignment of an IP address
// to a VPP interface.
func (d *InterfaceAddressDescriptor) IsInterfaceAddressKey(key string) bool {
	_, _, _, _, isAddrKey := interfaces.ParseInterfaceAddressKey(key)
	return isAddrKey
}

// Validate validates IP address to be assigned to an interface.
func (d *InterfaceAddressDescriptor) Validate(key string, emptyVal proto.Message) (err error) {
	_, _, _, invalidIP, _ := interfaces.ParseInterfaceAddressKey(key)
	if invalidIP {
		return errors.New("invalid IP address")
	}
	return nil
}

// Create assigns IP address to an interface.
func (d *InterfaceAddressDescriptor) Create(key string, emptyVal proto.Message) (metadata kvs.Metadata, err error) {
	iface, ipAddr, ipAddrNet, _, _ := interfaces.ParseInterfaceAddressKey(key)
	ipAddrNet.IP = ipAddr

	ifMeta, found := d.ifIndex.LookupByName(iface)
	if !found {
		err = errors.Errorf("failed to find interface %s", iface)
		d.log.Error(err)
		return nil, err
	}

	err = d.ifHandler.AddInterfaceIP(ifMeta.SwIfIndex, ipAddrNet)
	return nil, err
}

// Delete unassigns IP address from an interface.
func (d *InterfaceAddressDescriptor) Delete(key string, emptyVal proto.Message, metadata kvs.Metadata) (err error) {
	iface, ipAddr, ipAddrNet, _, _ := interfaces.ParseInterfaceAddressKey(key)
	ipAddrNet.IP = ipAddr

	if ipAddr.IsLinkLocalUnicast() {
		return nil
	}

	ifMeta, found := d.ifIndex.LookupByName(iface)
	if !found {
		err = errors.Errorf("failed to find interface %s", iface)
		d.log.Error(err)
		return err
	}

	err = d.ifHandler.DelInterfaceIP(ifMeta.SwIfIndex, ipAddrNet)
	return err
}

// Dependencies lists assignment of the interface into the VRF table as the only dependency.
func (d *InterfaceAddressDescriptor) Dependencies(key string, emptyVal proto.Message) []kvs.Dependency {
	iface, _, _, _, _ := interfaces.ParseInterfaceAddressKey(key)
	return []kvs.Dependency{{
		Label: interfaceInVrfDep,
		AnyOf: kvs.AnyOfDependency{
			KeyPrefixes: []string{interfaces.InterfaceVrfKeyPrefix(iface)},
		},
	}}
}
