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
	"net"
	"reflect"
	"strings"

	"github.com/go-errors/errors"
	"github.com/gogo/protobuf/proto"
	prototypes "github.com/gogo/protobuf/types"

	"github.com/ligato/cn-infra/idxmap"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/addrs"

	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/vppcalls"
	linux_intf "github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"
	linux_ifplugin "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	linux_ifdescriptor "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/descriptor"
)

const (
	// InterfaceDescriptorName is the name of the descriptor for VPP interfaces.
	InterfaceDescriptorName = "vpp-interfaces"

	// dependency labels
	afPacketHostInterfaceDep = "afpacket-host-interface"
	vxlanMulticastDep = "vxlan-multicast-interface"
)

// A list of non-retriable errors:
var (
	// ErrUnsupportedVPPInterfaceType is returned for VPP interfaces of unknown type.
	ErrUnsupportedVPPInterfaceType = errors.New("unsupported VPP interface type")

	// ErrInterfaceWithoutName is returned when VPP interface configuration has undefined
	// Name attribute.
	ErrInterfaceWithoutName = errors.New("VPP interface defined without logical name")

	// ErrInterfaceWithoutType is returned when VPP interface configuration has undefined
	// Type attribute.
	ErrInterfaceWithoutType = errors.New("VPP interface defined without type")

	// ErrUnnumberedWithIP is returned when configuration of a VPP unnumbered interface
	// includes an IP address.
	ErrUnnumberedWithIP = errors.New("VPP unnumbered interface was defined with IP address")
)

// InterfaceDescriptor teaches KVScheduler how to configure VPP interfaces.
type InterfaceDescriptor struct {
	// config
	defaultMtu       uint32

	// dependencies
	log              logging.Logger
	ifHandler        vppcalls.IfVppAPI
	linuxIfPlugin    linux_ifplugin.API /* optional, provide if AFPacket or TAP+AUTO_TAP interfaces are used */

	// runtime
	intfIndex        ifaceidx.IfaceMetadataIndex
	ethernetIntfs    map[string]uint32 // ethernet interface name -> sw_if_index (all known physical interfaces)
	memifSocketToID  map[string]uint32 // memif socket filename to ID map (all known sockets)
}

// NewInterfaceDescriptor creates a new instance of the Interface descriptor.
func NewInterfaceDescriptor(ifHandler vppcalls.IfVppAPI, defaultMtu uint32,
	linuxIfPlugin linux_ifplugin.API, log logging.PluginLogger) *InterfaceDescriptor {

	return &InterfaceDescriptor{
		ifHandler:       ifHandler,
		defaultMtu:      defaultMtu,
		linuxIfPlugin:   linuxIfPlugin,
		log:             log.NewLogger("-if-descriptor"),
		ethernetIntfs:   make(map[string]uint32),
		memifSocketToID: make(map[string]uint32),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *InterfaceDescriptor) GetDescriptor() *adapter.InterfaceDescriptor {
	return &adapter.InterfaceDescriptor{
		Name:               InterfaceDescriptorName,
		KeySelector:        d.IsInterfaceKey,
		ValueTypeName:      proto.MessageName(&interfaces.Interface{}),
		KeyLabel:           d.InterfaceNameFromKey,
		ValueComparator:    d.EquivalentInterfaces,
		NBKeyPrefix:        interfaces.Prefix,
		WithMetadata:       true,
		MetadataMapFactory: d.MetadataFactory,
		Add:                d.Add,
		Delete:             d.Delete,
		Modify:             d.Modify,
		ModifyWithRecreate: d.ModifyWithRecreate,
		IsRetriableFailure: d.IsRetriableFailure,
		Dependencies:       d.Dependencies,
		DerivedValues:      d.DerivedValues,
		Dump:               d.Dump,
		// If Linux-IfPlugin is loaded, dump it first.
		DumpDependencies:   []string{linux_ifdescriptor.InterfaceDescriptorName},
	}
}

// SetInterfaceIndex should be used to provide interface index immediately after
// the descriptor registration.
func (d *InterfaceDescriptor) SetInterfaceIndex(intfIndex ifaceidx.IfaceMetadataIndex) {
	d.intfIndex = intfIndex
}

// IsInterfaceKey returns true if the key is identifying VPP interface configuration.
func (d *InterfaceDescriptor) IsInterfaceKey(key string) bool {
	return strings.HasPrefix(key, interfaces.Prefix)
}

// InterfaceNameFromKey returns VPP interface name from the key.
func (d *InterfaceDescriptor) InterfaceNameFromKey(key string) string {
	name, _ := interfaces.ParseNameFromKey(key)
	return name
}

// EquivalentInterfaces is case-insensitive comparison function for
// interfaces.Interface, also ignoring the order of assigned IP addresses.
func (d *InterfaceDescriptor) EquivalentInterfaces(key string, intf1, intf2 *interfaces.Interface) bool {
	// attributes compared as usually:
	if intf1.Name != intf2.Name || intf1.Type != intf2.Type || intf1.Enabled != intf2.Enabled ||
		intf1.Vrf != intf2.Vrf || intf1.SetDhcpClient != intf2.SetDhcpClient {
		return false
	}

	if !proto.Equal(intf1.Unnumbered, intf2.Unnumbered) || !proto.Equal(intf1.RxModeSettings, intf2.RxModeSettings) ||
		!proto.Equal(intf1.RxPlacementSettings, intf2.RxPlacementSettings) {
		return false
	}

	switch intf1.Type {
	case interfaces.Interface_TAP_INTERFACE:
		if !proto.Equal(intf1.GetTap(), intf2.GetTap()) {
			return false
		}
	case interfaces.Interface_VXLAN_TUNNEL:
		if !proto.Equal(intf1.GetVxlan(), intf2.GetVxlan()) {
			return false
		}
	case interfaces.Interface_AF_PACKET_INTERFACE:
		if !proto.Equal(intf1.GetAfpacket(), intf2.GetAfpacket()) {
			return false
		}
	case interfaces.Interface_MEMORY_INTERFACE:
		if !proto.Equal(intf1.GetMemif(), intf2.GetMemif()) {
			return false
		}
	}

	// handle default MTU
	if d.getInterfaceMTU(intf1) != d.getInterfaceMTU(intf2) {
		return false
	}

	// compare MAC addresses case-insensitively
	if strings.ToLower(intf1.PhysAddress) != strings.ToLower(intf2.PhysAddress) {
		return false
	}

	// order-irrelevant comparison of IP addresses
	intf1Addrs, err1 := addrs.StrAddrsToStruct(intf1.IpAddresses)
	intf2Addrs, err2 := addrs.StrAddrsToStruct(intf2.IpAddresses)
	if err1 != nil || err2 != nil {
		// one or both of the configurations are invalid, compare lazily
		return reflect.DeepEqual(intf1.IpAddresses, intf2.IpAddresses)
	}
	obsolete, new := addrs.DiffAddr(intf1Addrs, intf2Addrs)
	return len(obsolete) == 0 && len(new) == 0
}

// MetadataFactory is a factory for index-map customized for VPP interfaces.
func (d *InterfaceDescriptor) MetadataFactory() idxmap.NamedMappingRW {
	return ifaceidx.NewIfaceIndex(logrus.DefaultLogger(), "vpp-interface-index")
}

// IsRetriableFailure returns <false> for errors related to invalid configuration.
func (d *InterfaceDescriptor) IsRetriableFailure(err error) bool {
	nonRetriable := []error{
		ErrUnsupportedVPPInterfaceType,
		ErrInterfaceWithoutName,
		ErrInterfaceWithoutType,
		ErrUnnumberedWithIP,
		}
	for _, nonRetriableErr := range nonRetriable {
		if err == nonRetriableErr {
			return false
		}
	}
	return true
}

// ModifyWithRecreate returns true if Type or Type-specific attributes are different.
func (d *InterfaceDescriptor) ModifyWithRecreate(key string, oldIntf, newIntf *interfaces.Interface, metadata *ifaceidx.IfaceMetadata) bool {
	if oldIntf.Type != newIntf.Type {
		return true
	}

	switch oldIntf.Type {
	case interfaces.Interface_TAP_INTERFACE:
		if !proto.Equal(oldIntf.GetTap(), newIntf.GetTap()) {
			return true
		}
	case interfaces.Interface_MEMORY_INTERFACE:
		if !proto.Equal(oldIntf.GetMemif(), newIntf.GetMemif()) {
			return true
		}

	case interfaces.Interface_VXLAN_TUNNEL:
		if !proto.Equal(oldIntf.GetVxlan(), newIntf.GetVxlan()) {
			return true
		}

	case interfaces.Interface_AF_PACKET_INTERFACE:
		if !proto.Equal(oldIntf.GetAfpacket(), newIntf.GetAfpacket()) {
			return true
		}
	}
	return false
}

// Dependencies lists dependencies for a VPP interface.
func (d *InterfaceDescriptor) Dependencies(key string, intf *interfaces.Interface) []scheduler.Dependency {
	var dependencies []scheduler.Dependency

	if intf.Type == interfaces.Interface_AF_PACKET_INTERFACE {
		// AF-PACKET depends on a referenced Linux interface in the default namespace
		dependencies = append(dependencies, scheduler.Dependency{
			Label: afPacketHostInterfaceDep,
			Key:   linux_intf.InterfaceHostNameKey(intf.GetAfpacket().GetHostIfName()),
		})
	}

	if intf.Type == interfaces.Interface_VXLAN_TUNNEL && intf.GetVxlan().GetMulticast() != "" {
		// VXLAN referencing an interface with Multicast IP address
		dependencies = append(dependencies, scheduler.Dependency{
			Label: vxlanMulticastDep,
			AnyOf: func(key string) bool {
				ifName, ifAddr, err := interfaces.ParseInterfaceAddressKey(key)
				return err == nil && ifName == intf.GetVxlan().GetMulticast() && ifAddr.IP.IsMulticast()
			},
		})
	}

	return dependencies
}

// DerivedValues derives:
//  - key-value for unnumbered configuration sub-section
//  - empty value for enabled DHCP client
//  - empty value from a TAP interface to represent its Linux-side
//  - one empty value for every IP address assigned to the interface.
func (d *InterfaceDescriptor) DerivedValues(key string, intf *interfaces.Interface) (derValues []scheduler.KeyValuePair) {
	// unnumbered interface
	if intf.GetUnnumbered().GetIsUnnumbered() {
		derValues = append(derValues, scheduler.KeyValuePair{
			Key:   interfaces.UnnumberedKey(intf.Name),
			Value: intf.GetUnnumbered(),
		})
	}

	// DHCP client
	if intf.SetDhcpClient {
		derValues = append(derValues, scheduler.KeyValuePair{
			Key:   interfaces.DHCPClientKey(intf.Name),
			Value: &prototypes.Empty{},
		})
	}

	// TAP interface host name
	if intf.Type == interfaces.Interface_TAP_INTERFACE {
		derValues = append(derValues, scheduler.KeyValuePair{
			Key:   interfaces.TAPHostNameKey(intf.GetTap().GetHostIfName()),
			Value: &prototypes.Empty{},
		})
	}

	// IP addresses
	for _, ipAddr := range intf.IpAddresses {
		derValues = append(derValues, scheduler.KeyValuePair{
			Key:   interfaces.InterfaceAddressKey(intf.Name, ipAddr),
			Value: &prototypes.Empty{},
		})
	}
	return derValues
}

// validateInterfaceConfig validates VPP interface configuration.
func (d *InterfaceDescriptor) validateInterfaceConfig(intf *interfaces.Interface) error {
	if intf.Name == "" {
		return ErrInterfaceWithoutName
	}
	if intf.Type == interfaces.Interface_UNDEFINED {
		return ErrInterfaceWithoutType
	}
	if intf.GetUnnumbered() != nil && intf.GetUnnumbered().GetIsUnnumbered() {
		if len(intf.GetIpAddresses()) > 0 {
			return ErrUnnumberedWithIP
		}
	}
	return nil
}

// getInterfaceMTU returns the interface MTU.
func (d *InterfaceDescriptor) getInterfaceMTU(intf *interfaces.Interface) uint32 {
	mtu := intf.Mtu
	if mtu == 0 {
		return d.defaultMtu /* still can be 0, i.e. undefined */
	}
	return mtu
}

// resolveMemifSocketFilename returns memif socket filename ID.
// Registers it if does not exists yet.
func (d *InterfaceDescriptor)  resolveMemifSocketFilename(memifIf *interfaces.Interface_MemifLink) (uint32, error) {
	if memifIf.GetSocketFilename() == "" {
		return 0, errors.Errorf("memif configuration does not contain socket file name")
	}
	registeredID, registered := d.memifSocketToID[memifIf.SocketFilename]
	if !registered {
		// Register new socket. ID is generated (default filename ID is 0, first is ID 1, second ID 2, etc)
		registeredID = uint32(len(d.memifSocketToID))
		err := d.ifHandler.RegisterMemifSocketFilename([]byte(memifIf.SocketFilename), registeredID)
		if err != nil {
			return 0, errors.Errorf("error registering socket file name %s (ID %d): %v", memifIf.SocketFilename, registeredID, err)
		}
		d.memifSocketToID[memifIf.SocketFilename] = registeredID
		d.log.Debugf("Memif socket filename %s registered under ID %d", memifIf.SocketFilename, registeredID)
	}
	return registeredID, nil
}

/**
Set rx-mode on specified VPP interface

Legend:
P - polling
I - interrupt
A - adaptive

Interfaces - supported modes:
* tap interface - PIA
* memory interface - PIA
* vxlan tunnel - PIA
* software loopback - PIA
* ethernet csmad - P
* af packet - PIA
*/
func (d *InterfaceDescriptor) configRxModeForInterface(intf *interfaces.Interface, ifIdx uint32) error {
	rxModeSettings := intf.RxModeSettings
	if rxModeSettings != nil {
		switch intf.Type {
		case interfaces.Interface_ETHERNET_CSMACD:
			if rxModeSettings.RxMode == interfaces.Interface_RxModeSettings_POLLING {
				return d.ifHandler.SetRxMode(ifIdx, rxModeSettings)
			}
		default:
			return d.ifHandler.SetRxMode(ifIdx, rxModeSettings)
		}
	}
	return nil
}

// getIPAddressVersions returns two flags to tell whether the provided list of addresses
// contains IPv4 and/or IPv6 type addresses
func getIPAddressVersions(ipAddrs []*net.IPNet) (isIPv4, isIPv6 bool) {
	for _, ip := range ipAddrs {
		if ip.IP.To4() != nil {
			isIPv4 = true
		} else {
			isIPv6 = true
		}
	}

	return
}
