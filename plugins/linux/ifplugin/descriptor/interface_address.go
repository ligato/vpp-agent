// Copyright (c) 2019 Cisco and/or its affiliates.
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
	"fmt"
	"syscall"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	interfaces "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	netalloc_api "github.com/ligato/vpp-agent/api/models/netalloc"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/ifaceidx"
	iflinuxcalls "github.com/ligato/vpp-agent/plugins/linux/ifplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linux/nsplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/netalloc"
)

const (
	// InterfaceAddressDescriptorName is the name of the descriptor for assigning
	// IP addresses to Linux interfaces.
	InterfaceAddressDescriptorName = "linux-interface-address"

	// DisableIPv6SysctlTemplate is used to enable ipv6 via sysctl.
	DisableIPv6SysctlTemplate = "net.ipv6.conf.%s.disable_ipv6"
)

// InterfaceAddressDescriptor (un)assigns IP address to/from Linux interface.
type InterfaceAddressDescriptor struct {
	log       logging.Logger
	ifHandler iflinuxcalls.NetlinkAPI
	nsPlugin  nsplugin.API
	addrAlloc netalloc.AddressAllocator
	intfIndex ifaceidx.LinuxIfMetadataIndex
}

// NewInterfaceAddressDescriptor creates a new instance of InterfaceAddressDescriptor.
func NewInterfaceAddressDescriptor(nsPlugin nsplugin.API, addrAlloc netalloc.AddressAllocator,
	ifHandler iflinuxcalls.NetlinkAPI, log logging.PluginLogger) (descr *kvs.KVDescriptor, ctx *InterfaceAddressDescriptor) {

	ctx = &InterfaceAddressDescriptor{
		ifHandler: ifHandler,
		nsPlugin:  nsPlugin,
		addrAlloc: addrAlloc,
		log:       log.NewLogger("interface-address-descriptor"),
	}
	descr = &kvs.KVDescriptor{
		Name:         InterfaceAddressDescriptorName,
		KeySelector:  ctx.IsInterfaceAddressKey,
		Validate:     ctx.Validate,
		Create:       ctx.Create,
		Delete:       ctx.Delete,
		Dependencies: ctx.Dependencies,
	}
	return
}

// SetInterfaceIndex should be used to provide interface index immediately after
// the descriptor registration.
func (d *InterfaceAddressDescriptor) SetInterfaceIndex(intfIndex ifaceidx.LinuxIfMetadataIndex) {
	d.intfIndex = intfIndex
}

// IsInterfaceVrfKey returns true if the key represents assignment of an IP address
// to a Linux interface (that needs to be applied). KVs representing addresses
// already allocated from netalloc plugin are excluded.
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

	ifMeta, found := d.intfIndex.LookupByName(iface)
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

	// switch to the namespace with the interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	defer revert()

	if ipAddr.IP.To4() == nil {
		// Enable IPv6 for loopback "lo" and the interface being configured
		for _, iface := range [2]string{"lo", ifMeta.HostIfName} {
			ipv6SysctlValueName := fmt.Sprintf(DisableIPv6SysctlTemplate, iface)

			// Read current sysctl value
			value, err := getSysctl(ipv6SysctlValueName)
			if err != nil || value == "0" {
				if err != nil {
					d.log.Warnf("could not read sysctl value for %v: %v",
						ifMeta.HostIfName, err)
				}
				continue
			}

			// Write sysctl to enable IPv6
			_, err = setSysctl(ipv6SysctlValueName, "0")
			if err != nil {
				return nil, fmt.Errorf("failed to enable IPv6 (%s=%s): %v",
					ipv6SysctlValueName, value, err)
			}
		}
	}

	err = d.ifHandler.AddInterfaceIP(ifMeta.HostIfName, ipAddr)

	// an attempt to add already assigned IP is not considered as error
	if err == syscall.EEXIST {
		err = nil
	}
	return nil, err
}

// Delete unassigns IP address from an interface.
func (d *InterfaceAddressDescriptor) Delete(key string, emptyVal proto.Message, metadata kvs.Metadata) (err error) {
	iface, addr, _, _, _ := interfaces.ParseInterfaceAddressKey(key)

	ifMeta, found := d.intfIndex.LookupByName(iface)
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

	// switch to the namespace with the interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
	if err != nil {
		d.log.Error(err)
		return err
	}
	defer revert()

	err = d.ifHandler.DelInterfaceIP(ifMeta.HostIfName, ipAddr)
	return err
}

// Dependencies mentions potential allocation of the IP address as dependency.
func (d *InterfaceAddressDescriptor) Dependencies(key string, emptyVal proto.Message) (deps []kvs.Dependency) {
	iface, addr, _, _, _ := interfaces.ParseInterfaceAddressKey(key)
	allocDep, hasAllocDep := d.addrAlloc.GetAddressAllocDep(addr, iface, "")
	if hasAllocDep {
		deps = append(deps, allocDep)
	}

	return deps
}
