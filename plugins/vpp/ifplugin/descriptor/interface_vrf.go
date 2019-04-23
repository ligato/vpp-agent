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
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

const (
	// InterfaceVrfDescriptorName is the name of the descriptor for assigning
	// VPP interface into VRF table.
	InterfaceVrfDescriptorName = "vpp-interface-vrf"

	// dependency labels
	vrfV4Dep        = "vrf-table-v4-exists"
	vrfV6Dep        = "vrf-table-v6-exists"
	inheritedVrfDep = "numbered-interface-assigned-to-VRF"
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
		Name:         InterfaceVrfDescriptorName,
		KeySelector:  descrCtx.IsInterfaceVrfKey,
		Create:       descrCtx.Create,
		Delete:       descrCtx.Delete,
		Dependencies: descrCtx.Dependencies,
	}
}

// IsInterfaceVrfKey returns true if the key represents assignment of an interface
// into a VRF table.
func (d *InterfaceVrfDescriptor) IsInterfaceVrfKey(key string) bool {
	_, _, _, _, isIfaceVrfKey := interfaces.ParseInterfaceVrfKey(key)
	if isIfaceVrfKey {
		return true
	}
	_, _, isIfaceInherVrfKey := interfaces.ParseInterfaceInheritedVrfKey(key)
	if isIfaceInherVrfKey {
		return true
	}
	return false
}

// Create puts interface into the given VRF table.
func (d *InterfaceVrfDescriptor) Create(key string, emptyVal proto.Message) (metadata kvs.Metadata, err error) {
	swIfIndex, vrf, ipv4, ipv6, err := d.getParametersFromKey(key)
	if err != nil {
		return nil, err
	}

	if vrf > 0 && ipv4 {
		err = d.ifHandler.SetInterfaceVrf(swIfIndex, uint32(vrf))
		if err != nil {
			d.log.Error(err)
			return nil, err
		}
	}
	if vrf > 0 && ipv6 {
		err = d.ifHandler.SetInterfaceVrfIPv6(swIfIndex, uint32(vrf))
		if err != nil {
			d.log.Error(err)
			return nil, err
		}
	}
	return nil, nil
}

// Delete removes interface from the given VRF table.
func (d *InterfaceVrfDescriptor) Delete(key string, emptyVal proto.Message, metadata kvs.Metadata) (err error) {
	swIfIndex, vrf, ipv4, ipv6, err := d.getParametersFromKey(key)
	if err != nil {
		return err
	}

	if vrf > 0 && ipv4 {
		err = d.ifHandler.SetInterfaceVrf(swIfIndex, uint32(0))
		if err != nil {
			d.log.Error(err)
			return err
		}
	}
	if vrf > 0 && ipv6 {
		err = d.ifHandler.SetInterfaceVrfIPv6(swIfIndex, uint32(0))
		if err != nil {
			d.log.Error(err)
			return err
		}
	}
	return nil
}

// Dependencies lists the target non-zero VRF as the only dependency.
func (d *InterfaceVrfDescriptor) Dependencies(key string, emptyVal proto.Message) (deps []kvs.Dependency) {
	if _, vrf, ipv4, ipv6, isIfaceVrfKey := interfaces.ParseInterfaceVrfKey(key); isIfaceVrfKey {
		if vrf > 0 && ipv4 {
			deps = append(deps, kvs.Dependency{
				Label: vrfV4Dep,
				Key:   l3.VrfTableKey(uint32(vrf), l3.VrfTable_IPV4),
			})
		}
		if vrf > 0 && ipv6 {
			deps = append(deps, kvs.Dependency{
				Label: vrfV6Dep,
				Key:   l3.VrfTableKey(uint32(vrf), l3.VrfTable_IPV6),
			})
		}
		return deps
	}

	_, fromIface, _ := interfaces.ParseInterfaceInheritedVrfKey(key)
	return []kvs.Dependency{
		{
			Label: inheritedVrfDep,
			AnyOf: kvs.AnyOfDependency{
				KeyPrefixes: []string{interfaces.InterfaceVrfKeyPrefix(fromIface)},
			},
		},
	}
}

func (d *InterfaceVrfDescriptor) getParametersFromKey(key string) (swIfIndex, vrf uint32, ipv4, ipv6 bool, err error) {
	var (
		isIfaceVrfKey    bool
		vrfTableID       int
		iface, fromIface string
	)

	iface, vrfTableID, ipv4, ipv6, isIfaceVrfKey = interfaces.ParseInterfaceVrfKey(key)
	if !isIfaceVrfKey {
		iface, fromIface, _ = interfaces.ParseInterfaceInheritedVrfKey(key)
		fromIfaceMeta, found := d.ifIndex.LookupByName(fromIface)
		if !found {
			err = errors.Errorf("failed to find interface %s", iface)
			d.log.Error(err)
			return
		}
		vrfTableID = int(fromIfaceMeta.Vrf)
		ipv4, ipv6 = getIPAddressVersions(fromIfaceMeta.IPAddresses)
	}

	ifMeta, found := d.ifIndex.LookupByName(iface)
	if !found {
		err = errors.Errorf("failed to find interface %s", iface)
		d.log.Error(err)
		return
	}

	swIfIndex = ifMeta.SwIfIndex
	vrf = uint32(vrfTableID)
	return
}
