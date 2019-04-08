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
	// InterfaceVrfDescriptorName is the name of the descriptor for assigning
	// VPP interface into VRF table.
	InterfaceVrfDescriptorName = "vpp-interface-vrf"

	// dependency labels
	vrfDep = "vrf-exists"
)

// InterfaceVrfDescriptor (un)assigns VPP interface to IPv4/IPv6 VRF table. 
type InterfaceVrfDescriptor struct {
	log       logging.Logger
	ifHandler vppcalls.InterfaceVppAPI
	ifIndex   ifaceidx.IfaceMetadataIndex
}

// NewInterfaceVrfDescriptor creates a new instance of InterfaceVrfDescriptor.
func NewInterfaceVrfDescriptor(ifHandler vppcalls.InterfaceVppAPI, ifIndex ifaceidx.IfaceMetadataIndex,
	log logging.PluginLogger) *kvs.KVDescriptor {

	descrCtx := &InterfaceVrfDescriptor{
		ifHandler: ifHandler,
		ifIndex:   ifIndex,
		log:       log.NewLogger("interface-vrf-descriptor"),
	}
	return &kvs.KVDescriptor{
		Name:          InterfaceVrfDescriptorName,
		KeySelector:   descrCtx.IsInterfaceVrfKey,
		Create:        descrCtx.Create,
		Delete:        descrCtx.Delete,
		Dependencies:  descrCtx.Dependencies,
	}
}

// IsInterfaceVrfKey returns true if the key represents assignment of an interface
// into a VRF table.
func (d *InterfaceVrfDescriptor) IsInterfaceVrfKey(key string) bool {
	_, _, _, isInterfaceVrfKey := interfaces.ParseInterfaceVrfTableKey(key)
	return isInterfaceVrfKey
}

// Create puts interface into the given VRF table.
func (d *InterfaceVrfDescriptor) Create(key string, emptyVal proto.Message) (metadata kvs.Metadata, err error) {
	iface, vrf, ipv6, _ := interfaces.ParseInterfaceVrfTableKey(key)
	if vrf == 0 {
		// NOOP
		return nil, nil
	}

	ifMeta, found := d.ifIndex.LookupByName(iface)
	if !found {
		err = errors.Errorf("failed to find interface %s", iface)
		d.log.Error(err)
		return nil, err
	}

	if ipv6 {
		err = d.ifHandler.SetInterfaceVrfIPv6(ifMeta.SwIfIndex, uint32(vrf))
	} else {
		err = d.ifHandler.SetInterfaceVrf(ifMeta.SwIfIndex, uint32(vrf))
	}

	return nil, err
}

// Delete removes interface from the given VRF table.
func (d *InterfaceVrfDescriptor) Delete(key string, emptyVal proto.Message, metadata kvs.Metadata) (err error) {
	iface, vrf, ipv6, _ := interfaces.ParseInterfaceVrfTableKey(key)
	if vrf == 0 {
		// NOOP
		return nil
	}

	ifMeta, found := d.ifIndex.LookupByName(iface)
	if !found {
		err = errors.Errorf("failed to find interface %s", iface)
		d.log.Error(err)
		return err
	}

	if ipv6 {
		err = d.ifHandler.SetInterfaceVrfIPv6(ifMeta.SwIfIndex, uint32(0))
	} else {
		err = d.ifHandler.SetInterfaceVrf(ifMeta.SwIfIndex, uint32(0))
	}

	return err
}

// Dependencies lists non-zero VRF as the only dependency.
func (d *InterfaceVrfDescriptor) Dependencies(key string, emptyVal proto.Message) []kvs.Dependency {
	iface, vrf, ipv6, _ := interfaces.ParseInterfaceVrfTableKey(key)
	if vrf == 0 {
		return nil
	}
	return []kvs.Dependency{{
		Label: vrfDep,
		Key:   interfaces.InterfaceVrfTableKey(iface, vrf, ipv6),
	}}
}
