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
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	netalloc_api "go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

const (
	// InterfaceAddressDescriptorName is the name of the descriptor for assigning
	// IP addresses to VPP interfaces.
	InterfaceAddressDescriptorName = "vpp-interface-address"

	// dependency labels
	interfaceInVrfDep = "interface-assigned-to-vrf-table"
)

// InterfaceAddressDescriptor (un)assigns (static) IP address to/from VPP interface.
type InterfaceAddressDescriptor struct {
	log       logging.Logger
	ifHandler vppcalls.InterfaceVppAPI
	ifIndex   ifaceidx.IfaceMetadataIndex
	addrAlloc netalloc.AddressAllocator
}

// NewInterfaceAddressDescriptor creates a new instance of InterfaceAddressDescriptor.
func NewInterfaceAddressDescriptor(ifHandler vppcalls.InterfaceVppAPI, addrAlloc netalloc.AddressAllocator,
	ifIndex ifaceidx.IfaceMetadataIndex, log logging.PluginLogger) *kvs.KVDescriptor {

	descrCtx := &InterfaceAddressDescriptor{
		ifHandler: ifHandler,
		ifIndex:   ifIndex,
		addrAlloc: addrAlloc,
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
// to a VPP interface (that needs to be applied). KVs representing addresses
// already allocated from netalloc plugin or obtained from a DHCP server are
// excluded.
func (d *InterfaceAddressDescriptor) IsInterfaceAddressKey(key string) bool {
	_, _, source, _, isAddrKey := interfaces.ParseInterfaceAddressKey(key)
	return isAddrKey &&
		(source == netalloc_api.IPAddressSource_STATIC || source == netalloc_api.IPAddressSource_ALLOC_REF)
}

// Validate validates IP address to be assigned to an interface.
func (d *InterfaceAddressDescriptor) Validate(key string, emptyVal proto.Message) (err error) {
	iface, addr, _, invalidKey, _ := interfaces.ParseInterfaceAddressKey(key)
	if invalidKey {
		return errors.New("invalid key")
	}

	return d.addrAlloc.ValidateIPAddress(addr, iface, "ip_addresses", netalloc.GwRefUnexpected)
}

// Create assigns IP address to an interface.
func (d *InterfaceAddressDescriptor) Create(key string, emptyVal proto.Message) (metadata kvs.Metadata, err error) {
	iface, addr, _, _, _ := interfaces.ParseInterfaceAddressKey(key)

	ifMeta, found := d.ifIndex.LookupByName(iface)
	if !found {
		err = errors.Errorf("failed to find interface %s", iface)
		d.log.Error(err)
		return nil, err
	}

	ipAddr, err := d.addrAlloc.GetOrParseIPAddress(addr, iface, netalloc_api.IPAddressForm_ADDR_WITH_MASK)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	err = d.ifHandler.AddInterfaceIP(ifMeta.SwIfIndex, ipAddr)
	return nil, err
}

// Delete unassigns IP address from an interface.
func (d *InterfaceAddressDescriptor) Delete(key string, emptyVal proto.Message, metadata kvs.Metadata) (err error) {
	iface, addr, _, _, _ := interfaces.ParseInterfaceAddressKey(key)

	ifMeta, found := d.ifIndex.LookupByName(iface)
	if !found {
		err = errors.Errorf("failed to find interface %s", iface)
		d.log.Error(err)
		return err
	}

	ipAddr, err := d.addrAlloc.GetOrParseIPAddress(addr, iface, netalloc_api.IPAddressForm_ADDR_WITH_MASK)
	if err != nil {
		d.log.Error(err)
		return err
	}

	if ipAddr.IP.IsLinkLocalUnicast() {
		return nil
	}

	err = d.ifHandler.DelInterfaceIP(ifMeta.SwIfIndex, ipAddr)
	return err
}

// Dependencies lists assignment of the interface into the VRF table and potential
// allocation of the IP address as dependencies.
func (d *InterfaceAddressDescriptor) Dependencies(key string, emptyVal proto.Message) []kvs.Dependency {
	iface, addr, _, _, _ := interfaces.ParseInterfaceAddressKey(key)
	deps := []kvs.Dependency{{
		Label: interfaceInVrfDep,
		AnyOf: kvs.AnyOfDependency{
			KeyPrefixes: []string{interfaces.InterfaceVrfKeyPrefix(iface)},
		},
	}}

	allocDep, hasAllocDep := d.addrAlloc.GetAddressAllocDep(addr, iface, "")
	if hasAllocDep {
		deps = append(deps, allocDep)
	}

	return deps
}
