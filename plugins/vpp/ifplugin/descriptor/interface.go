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
	"fmt"
	"hash/fnv"
	"net"
	"strings"

	"github.com/gogo/protobuf/proto"
	prototypes "github.com/gogo/protobuf/types"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	linux_intf "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	linux_ns "github.com/ligato/vpp-agent/api/models/linux/namespace"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	linux_ifdescriptor "github.com/ligato/vpp-agent/plugins/linux/ifplugin/descriptor"
	linux_ifaceidx "github.com/ligato/vpp-agent/plugins/linux/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

const (
	// InterfaceDescriptorName is the name of the descriptor for VPP interfaces.
	InterfaceDescriptorName = "vpp-interface"

	// dependency labels
	afPacketHostInterfaceDep = "afpacket-host-interface-exists"
	vxlanMulticastDep        = "vxlan-multicast-interface-exists"
	vxlanVrfTableDep         = "vrf-table-for-vxlan-exists"
	microserviceDep          = "microservice-available"
	parentInterfaceDep       = "parent-interface-exists"

	// how many characters a logical interface name is allowed to have
	//  - determined by much fits into the VPP interface tag (64 null-terminated character string)
	logicalNameLengthLimit = 63

	// prefix prepended to internal names of untagged interfaces to construct unique
	// logical names
	untaggedIfPreffix = "UNTAGGED-"

	// suffix attached to logical names of dumped TAP interfaces with Linux side
	// not found by Retrieve of Linux-ifplugin
	tapMissingLinuxSideSuffix = "-MISSING_LINUX_SIDE"

	// suffix attached to logical names of dumped AF-PACKET interfaces connected
	// to missing Linux interfaces
	afPacketMissingAttachedIfSuffix = "-MISSING_ATTACHED_INTERFACE"

	// default memif attributes
	defaultMemifNumOfQueues uint32 = 1
	defaultMemifBufferSize  uint32 = 2048
	defaultMemifRingSize    uint32 = 1024
)

// A list of non-retriable errors:
var (
	// ErrUnsupportedVPPInterfaceType is returned for VPP interfaces of unknown type.
	ErrUnsupportedVPPInterfaceType = errors.New("unsupported VPP interface type")

	// ErrInterfaceWithoutName is returned when VPP interface configuration has undefined
	// Name attribute.
	ErrInterfaceWithoutName = errors.New("VPP interface defined without logical name")

	// ErrInterfaceNameTooLong is returned when VPP interface logical name exceeds the length limit.
	ErrInterfaceNameTooLong = errors.New("VPP interface logical name exceeds the length limit (63 characters)")

	// ErrInterfaceWithoutType is returned when VPP interface configuration has undefined
	// Type attribute.
	ErrInterfaceWithoutType = errors.New("VPP interface defined without type")

	// ErrUnnumberedWithIP is returned when configuration of a VPP unnumbered interface
	// includes an IP address.
	ErrUnnumberedWithIP = errors.New("VPP unnumbered interface was defined with IP address")

	// ErrAfPacketWithoutHostName is returned when AF-Packet configuration is missing host interface name.
	ErrAfPacketWithoutHostName = errors.New("VPP AF-Packet interface was defined without host interface name")

	// ErrInterfaceLinkMismatch is returned when interface type does not match the link configuration.
	ErrInterfaceLinkMismatch = errors.New("VPP interface type and link configuration do not match")

	// ErrRedefinedRxPlacement is returned when Rx placement has multiple definitions for the same queue.
	ErrRedefinedRxPlacement = errors.New("redefined RX Placement")

	// ErrSubInterfaceWithoutParent is returned when interface of type sub-interface is defined without parent.
	ErrSubInterfaceWithoutParent = errors.Errorf("subinterface with no parent interface defined")

	// ErrDPDKInterfaceMissing is returned when the expected DPDK interface does not exist on the VPP.
	ErrDPDKInterfaceMissing = errors.Errorf("DPDK interface with given name does not exists")

	// ErrBondInterfaceIDExists is returned when the bond interface uses existing ID value
	ErrBondInterfaceIDExists = errors.Errorf("Bond interface ID already exists")
)

// InterfaceDescriptor teaches KVScheduler how to configure VPP interfaces.
type InterfaceDescriptor struct {
	// config
	defaultMtu uint32

	// dependencies
	log       logging.Logger
	ifHandler vppcalls.InterfaceVppAPI

	// optional dependencies, provide if AFPacket and/or TAP+TAP_TO_VPP interfaces are used
	linuxIfPlugin  LinuxPluginAPI
	linuxIfHandler NetlinkAPI
	nsPlugin       nsplugin.API

	// runtime
	intfIndex              ifaceidx.IfaceMetadataIndex
	memifSocketToID        map[string]uint32 // memif socket filename to ID map (all known sockets)
	defaultMemifSocketPath string
	bondIDs                map[uint32]string // bond ID to name (ID != sw_if_idx)
	ethernetIfs            map[string]uint32 // name-to-index map of ethernet interfaces (entry is not
	// removed even if interface is un-configured)
}

// LinuxPluginAPI is defined here to avoid import cycles.
type LinuxPluginAPI interface {
	// GetInterfaceIndex gives read-only access to map with metadata of all configured
	// linux interfaces.
	GetInterfaceIndex() linux_ifaceidx.LinuxIfMetadataIndex
}

// NetlinkAPI here lists only those Netlink methods that are actually used by InterfaceDescriptor.
type NetlinkAPI interface {
	// InterfaceExists verifies interface existence
	InterfaceExists(ifName string) (bool, error)
}

// NewInterfaceDescriptor creates a new instance of the Interface descriptor.
func NewInterfaceDescriptor(ifHandler vppcalls.InterfaceVppAPI, defaultMtu uint32,
	linuxIfHandler NetlinkAPI, linuxIfPlugin LinuxPluginAPI, nsPlugin nsplugin.API,
	log logging.PluginLogger) (descr *kvs.KVDescriptor, ctx *InterfaceDescriptor) {

	// descriptor context
	ctx = &InterfaceDescriptor{
		ifHandler:       ifHandler,
		defaultMtu:      defaultMtu,
		linuxIfPlugin:   linuxIfPlugin,
		linuxIfHandler:  linuxIfHandler,
		nsPlugin:        nsPlugin,
		log:             log.NewLogger("if-descriptor"),
		memifSocketToID: make(map[string]uint32),
		ethernetIfs:     make(map[string]uint32),
		bondIDs:         make(map[uint32]string),
	}

	// descriptor
	typedDescr := &adapter.InterfaceDescriptor{
		Name:               InterfaceDescriptorName,
		NBKeyPrefix:        interfaces.ModelInterface.KeyPrefix(),
		ValueTypeName:      interfaces.ModelInterface.ProtoName(),
		KeySelector:        interfaces.ModelInterface.IsKeyValid,
		KeyLabel:           interfaces.ModelInterface.StripKeyPrefix,
		ValueComparator:    ctx.EquivalentInterfaces,
		WithMetadata:       true,
		MetadataMapFactory: ctx.MetadataFactory,
		Validate:           ctx.Validate,
		Create:             ctx.Create,
		Delete:             ctx.Delete,
		Update:             ctx.Update,
		UpdateWithRecreate: ctx.UpdateWithRecreate,
		Retrieve:           ctx.Retrieve,
		Dependencies:       ctx.Dependencies,
		DerivedValues:      ctx.DerivedValues,
		// If Linux-IfPlugin is loaded, dump it first.
		RetrieveDependencies: []string{linux_ifdescriptor.InterfaceDescriptorName},
	}
	descr = adapter.NewInterfaceDescriptor(typedDescr)
	return
}

// SetInterfaceIndex should be used to provide interface index immediately after
// the descriptor registration.
func (d *InterfaceDescriptor) SetInterfaceIndex(intfIndex ifaceidx.IfaceMetadataIndex) {
	d.intfIndex = intfIndex
}

// EquivalentInterfaces is case-insensitive comparison function for
// interfaces.Interface, also ignoring the order of assigned IP addresses.
func (d *InterfaceDescriptor) EquivalentInterfaces(key string, oldIntf, newIntf *interfaces.Interface) bool {
	// attributes compared as usually:
	if oldIntf.Name != newIntf.Name ||
		oldIntf.Type != newIntf.Type ||
		oldIntf.Enabled != newIntf.Enabled ||
		oldIntf.SetDhcpClient != newIntf.SetDhcpClient {
		return false
	}
	if !proto.Equal(oldIntf.Unnumbered, newIntf.Unnumbered) {
		return false
	}

	// type-specific (defaults considered)
	if !d.equivalentTypeSpecificConfig(oldIntf, newIntf) {
		return false
	}

	if newIntf.Unnumbered == nil { // unnumbered inherits VRF from numbered interface
		if oldIntf.Vrf != newIntf.Vrf {
			return false
		}
	}

	// handle default/unspecified MTU (except VxLAN and IPSec tunnel)
	if newIntf.Type != interfaces.Interface_VXLAN_TUNNEL && newIntf.Type != interfaces.Interface_IPSEC_TUNNEL {
		if d.getInterfaceMTU(newIntf) != 0 && d.getInterfaceMTU(oldIntf) != d.getInterfaceMTU(newIntf) {
			return false
		}
	}

	// compare MAC addresses case-insensitively (also handle unspecified MAC address)
	if newIntf.PhysAddress != "" &&
		strings.ToLower(oldIntf.PhysAddress) != strings.ToLower(newIntf.PhysAddress) {
		return false
	}

	if !equalStringSets(oldIntf.IpAddresses, newIntf.IpAddresses) {
		// call Update just to update IP addresses in the metadata
		return false
	}

	return true
}

// equivalentTypeSpecificConfig compares type-specific sections of two interface configurations.
func (d *InterfaceDescriptor) equivalentTypeSpecificConfig(oldIntf, newIntf *interfaces.Interface) bool {
	switch oldIntf.Type {
	case interfaces.Interface_TAP:
		if !proto.Equal(getTapConfig(oldIntf), getTapConfig(newIntf)) {
			return false
		}
	case interfaces.Interface_VXLAN_TUNNEL:
		if !proto.Equal(oldIntf.GetVxlan(), newIntf.GetVxlan()) {
			return false
		}
	case interfaces.Interface_AF_PACKET:
		if oldIntf.GetAfpacket().GetHostIfName() != newIntf.GetAfpacket().GetHostIfName() {
			return false
		}
	case interfaces.Interface_MEMIF:
		if !d.equivalentMemifs(oldIntf.GetMemif(), newIntf.GetMemif()) {
			return false
		}
	case interfaces.Interface_IPSEC_TUNNEL:
		if !d.equivalentIPSecTunnels(oldIntf.GetIpsec(), newIntf.GetIpsec()) {
			return false
		}
	case interfaces.Interface_SUB_INTERFACE:
		if !proto.Equal(oldIntf.GetSub(), newIntf.GetSub()) {
			return false
		}
	case interfaces.Interface_VMXNET3_INTERFACE:
		if !d.equivalentVmxNet3(oldIntf.GetVmxNet3(), newIntf.GetVmxNet3()) {
			return false
		}
	case interfaces.Interface_BOND_INTERFACE:
		if !d.equivalentBond(oldIntf.GetBond(), newIntf.GetBond()) {
			return false
		}
	}
	return true
}

// equivalentMemifs compares two memifs for equivalence.
func (d *InterfaceDescriptor) equivalentMemifs(oldMemif, newMemif *interfaces.MemifLink) bool {
	if oldMemif.GetMode() != newMemif.GetMode() ||
		oldMemif.GetMaster() != newMemif.GetMaster() ||
		oldMemif.GetId() != newMemif.GetId() ||
		oldMemif.GetSecret() != newMemif.GetSecret() {
		return false
	}
	// default values considered:
	if d.getMemifSocketFilename(oldMemif) != d.getMemifSocketFilename(newMemif) ||
		d.getMemifBufferSize(oldMemif) != d.getMemifBufferSize(newMemif) ||
		d.getMemifRingSize(oldMemif) != d.getMemifRingSize(newMemif) ||
		d.getMemifNumOfRxQueues(oldMemif) != d.getMemifNumOfRxQueues(newMemif) ||
		d.getMemifNumOfTxQueues(oldMemif) != d.getMemifNumOfTxQueues(newMemif) {
		return false
	}
	return true
}

// equivalentIPSecTunnels compares two IPSec tunnels for equivalence.
func (d *InterfaceDescriptor) equivalentIPSecTunnels(oldTun, newTun *interfaces.IPSecLink) bool {
	return oldTun.Esn == newTun.Esn &&
		oldTun.AntiReplay == newTun.AntiReplay &&
		oldTun.LocalSpi == newTun.LocalSpi &&
		oldTun.RemoteSpi == newTun.RemoteSpi &&
		oldTun.CryptoAlg == newTun.CryptoAlg &&
		oldTun.LocalCryptoKey == newTun.LocalCryptoKey &&
		oldTun.RemoteCryptoKey == newTun.RemoteCryptoKey &&
		oldTun.IntegAlg == newTun.IntegAlg &&
		oldTun.LocalIntegKey == newTun.LocalIntegKey &&
		oldTun.RemoteIntegKey == newTun.RemoteIntegKey &&
		oldTun.EnableUdpEncap == newTun.EnableUdpEncap
}

// equivalentVmxNets compares two vmxnet3 interfaces for equivalence.
func (d *InterfaceDescriptor) equivalentVmxNet3(oldVmxNet3, newVmxNet3 *interfaces.VmxNet3Link) bool {
	return oldVmxNet3.RxqSize == newVmxNet3.RxqSize &&
		oldVmxNet3.TxqSize == newVmxNet3.TxqSize
}

// equivalentBond compares two bond interfaces for equivalence.
func (d *InterfaceDescriptor) equivalentBond(oldBond, newBond *interfaces.BondLink) bool {
	if len(oldBond.BondedInterfaces) != len(newBond.BondedInterfaces) {
		return false
	}
	for _, oldBondSlave := range oldBond.BondedInterfaces {
		var found bool
		for _, newBondSlave := range newBond.BondedInterfaces {
			if oldBondSlave.Name == newBondSlave.Name &&
				oldBondSlave.IsPassive == newBondSlave.IsPassive &&
				oldBondSlave.IsLongTimeout == newBondSlave.IsLongTimeout {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return oldBond.Id == newBond.Id &&
		oldBond.Mode == newBond.Mode &&
		oldBond.Lb == newBond.Lb
}

// MetadataFactory is a factory for index-map customized for VPP interfaces.
func (d *InterfaceDescriptor) MetadataFactory() idxmap.NamedMappingRW {
	return ifaceidx.NewIfaceIndex(d.log, "vpp-interface-index")
}

// Validate validates VPP interface configuration.
func (d *InterfaceDescriptor) Validate(key string, intf *interfaces.Interface) error {
	// validate name
	if name := intf.GetName(); name == "" {
		return kvs.NewInvalidValueError(ErrInterfaceWithoutName, "name")
	} else if len(name) > logicalNameLengthLimit {
		return kvs.NewInvalidValueError(ErrInterfaceNameTooLong, "name")
	}

	// validate interface type defined
	if intf.GetType() == interfaces.Interface_UNDEFINED_TYPE {
		return kvs.NewInvalidValueError(ErrInterfaceWithoutType, "type")
	}

	// validate link with interface type
	linkMismatchErr := kvs.NewInvalidValueError(ErrInterfaceLinkMismatch, "link")
	switch intf.GetLink().(type) {
	case *interfaces.Interface_Sub:
		if intf.Type != interfaces.Interface_SUB_INTERFACE {
			return linkMismatchErr
		}
	case *interfaces.Interface_Memif:
		if intf.Type != interfaces.Interface_MEMIF {
			return linkMismatchErr
		}
	case *interfaces.Interface_Afpacket:
		if intf.Type != interfaces.Interface_AF_PACKET {
			return linkMismatchErr
		}
	case *interfaces.Interface_Vxlan:
		if intf.Type != interfaces.Interface_VXLAN_TUNNEL {
			return linkMismatchErr
		}
	case *interfaces.Interface_Tap:
		if intf.Type != interfaces.Interface_TAP {
			return linkMismatchErr
		}
	case *interfaces.Interface_Bond:
		if intf.Type != interfaces.Interface_BOND_INTERFACE {
			return linkMismatchErr
		}
	case *interfaces.Interface_VmxNet3:
		if intf.Type != interfaces.Interface_VMXNET3_INTERFACE {
			return linkMismatchErr
		}
	case *interfaces.Interface_Ipsec:
		if intf.Type != interfaces.Interface_IPSEC_TUNNEL {
			return linkMismatchErr
		}
	case nil:
		if intf.Type != interfaces.Interface_SOFTWARE_LOOPBACK &&
			intf.Type != interfaces.Interface_DPDK {
			return errors.Errorf("VPP interface type %s must have link defined", intf.Type)
		}
	}

	// validate type specific
	switch intf.GetType() {
	case interfaces.Interface_SUB_INTERFACE:
		if parentName := intf.GetSub().GetParentName(); parentName == "" {
			return kvs.NewInvalidValueError(ErrSubInterfaceWithoutParent, "link.sub.parent_name")
		}
	case interfaces.Interface_DPDK:
		if _, ok := d.ethernetIfs[intf.Name]; !ok {
			return kvs.NewInvalidValueError(ErrDPDKInterfaceMissing, "name")
		}
	case interfaces.Interface_AF_PACKET:
		if intf.GetAfpacket().GetHostIfName() == "" {
			return kvs.NewInvalidValueError(ErrAfPacketWithoutHostName, "link.afpacket.host_if_name")
		}
	case interfaces.Interface_BOND_INTERFACE:
		if name, ok := d.bondIDs[intf.GetBond().GetId()]; ok && name != intf.GetName() {
			return kvs.NewInvalidValueError(ErrBondInterfaceIDExists, "link.bond.id")
		}
	}

	// validate unnumbered
	if intf.GetUnnumbered() != nil {
		if len(intf.GetIpAddresses()) > 0 {
			return kvs.NewInvalidValueError(ErrUnnumberedWithIP, "unnumbered", "ip_addresses")
		}
	}

	// validate rx placements before before deriving
	for i, rxPlacement1 := range intf.GetRxPlacements() {
		for j := i + 1; j < len(intf.GetRxPlacements()); j++ {
			rxPlacement2 := intf.GetRxPlacements()[j]
			if rxPlacement1.Queue == rxPlacement2.Queue {
				return kvs.NewInvalidValueError(ErrRedefinedRxPlacement,
					fmt.Sprintf("rx_placement[.queue=%d]", rxPlacement1.Queue))
			}
		}
	}

	return nil
}

// UpdateWithRecreate returns true if Type or Type-specific attributes are different.
func (d *InterfaceDescriptor) UpdateWithRecreate(key string, oldIntf, newIntf *interfaces.Interface, metadata *ifaceidx.IfaceMetadata) bool {
	if oldIntf.Type != newIntf.Type {
		return true
	}

	// if type-specific attributes have changed, then re-create the interface
	if !d.equivalentTypeSpecificConfig(oldIntf, newIntf) {
		return true
	}

	if oldIntf.GetType() == interfaces.Interface_VXLAN_TUNNEL && oldIntf.Vrf != newIntf.Vrf {
		// for VXLAN interface a change in the VRF assignment requires full re-creation
		return true
	}

	// case for af-packet mac update (cannot be updated directly)
	if oldIntf.GetType() == interfaces.Interface_AF_PACKET && oldIntf.PhysAddress != newIntf.PhysAddress {
		return true
	}

	return false
}

// Dependencies lists dependencies for a VPP interface.
func (d *InterfaceDescriptor) Dependencies(key string, intf *interfaces.Interface) (dependencies []kvs.Dependency) {
	switch intf.Type {
	case interfaces.Interface_AF_PACKET:
		// AF-PACKET depends on a referenced Linux interface in the default namespace
		dependencies = append(dependencies, kvs.Dependency{
			Label: afPacketHostInterfaceDep,
			Key:   linux_intf.InterfaceHostNameKey(intf.GetAfpacket().GetHostIfName()),
		})
	case interfaces.Interface_TAP:
		// TAP connects VPP with microservice
		if toMicroservice := intf.GetTap().GetToMicroservice(); toMicroservice != "" {
			dependencies = append(dependencies, kvs.Dependency{
				Label: microserviceDep,
				Key:   linux_ns.MicroserviceKey(toMicroservice),
			})
		}
	case interfaces.Interface_VXLAN_TUNNEL:
		// VXLAN referencing an interface with Multicast IP address
		if vxlanMulticast := intf.GetVxlan().GetMulticast(); vxlanMulticast != "" {
			dependencies = append(dependencies, kvs.Dependency{
				Label: vxlanMulticastDep,
				AnyOf: kvs.AnyOfDependency{
					KeyPrefixes: []string{interfaces.InterfaceAddressPrefix(vxlanMulticast)},
					KeySelector: func(key string) bool {
						_, ifaceAddr, _, _, _ := interfaces.ParseInterfaceAddressKey(key)
						return ifaceAddr != nil && ifaceAddr.IsMulticast()
					},
				},
			})
		}
		if intf.GetVrf() != 0 {
			// binary API for creating VXLAN tunnel requires the VRF table
			// to be already created
			var protocol l3.VrfTable_Protocol
			srcAddr := net.ParseIP(intf.GetVxlan().GetSrcAddress()).To4()
			dstAddr := net.ParseIP(intf.GetVxlan().GetDstAddress()).To4()
			if srcAddr == nil && dstAddr == nil {
				protocol = l3.VrfTable_IPV6
			}
			dependencies = append(dependencies, kvs.Dependency{
				Label: vxlanVrfTableDep,
				Key:   l3.VrfTableKey(intf.GetVrf(), protocol),
			})
		}
	case interfaces.Interface_SUB_INTERFACE:
		// SUB_INTERFACE requires parent interface
		if parentName := intf.GetSub().GetParentName(); parentName != "" {
			dependencies = append(dependencies, kvs.Dependency{
				Label: parentInterfaceDep,
				Key:   interfaces.InterfaceKey(parentName),
			})
		}
	}

	return dependencies
}

// DerivedValues derives:
//  - key-value for unnumbered configuration sub-section
//  - empty value for enabled DHCP client
//  - configuration for every slave of a bonded interface
//  - one empty value for every IP address to be assigned to the interface
//  - one empty value for VRF table to put the interface into
//  - one value with interface configuration reduced to RxMode if set
//  - one Interface_RxPlacement for every queue with configured Rx placement
//  - one empty value which will be created once at least one IP address is
//    assigned to the interface.
func (d *InterfaceDescriptor) DerivedValues(key string, intf *interfaces.Interface) (derValues []kvs.KeyValuePair) {
	// unnumbered interface
	if intf.GetUnnumbered() != nil {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   interfaces.UnnumberedKey(intf.Name),
			Value: intf.GetUnnumbered(),
		})
	}

	// bond slave interface
	if intf.Type == interfaces.Interface_BOND_INTERFACE && intf.GetBond() != nil {
		for _, slaveIf := range intf.GetBond().GetBondedInterfaces() {
			derValues = append(derValues, kvs.KeyValuePair{
				Key:   interfaces.BondedInterfaceKey(intf.Name, slaveIf.Name),
				Value: slaveIf,
			})
		}
	}

	// DHCP client
	if intf.SetDhcpClient {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   interfaces.DHCPClientKey(intf.Name),
			Value: &prototypes.Empty{},
		})
	}

	// IP addresses
	for _, ipAddr := range intf.IpAddresses {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   interfaces.InterfaceAddressKey(intf.Name, ipAddr),
			Value: &prototypes.Empty{},
		})
	}

	// VRF assignment
	if intf.GetUnnumbered() != nil {
		// VRF inherited from the target numbered interface
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   interfaces.InterfaceInheritedVrfKey(intf.GetName(), intf.GetUnnumbered().GetInterfaceWithIp()),
			Value: &prototypes.Empty{},
		})
	} else {
		// not unnumbered
		var hasIPv4, hasIPv6 bool
		if intf.Type == interfaces.Interface_VXLAN_TUNNEL {
			srcAddr := net.ParseIP(intf.GetVxlan().GetSrcAddress()).To4()
			dstAddr := net.ParseIP(intf.GetVxlan().GetDstAddress()).To4()
			if srcAddr == nil && dstAddr == nil {
				hasIPv6 = true
			} else {
				hasIPv4 = true
			}
		} else {
			// not VXLAN tunnel
			hasIPv4, hasIPv6 = getIPAddressVersions(intf.IpAddresses)
		}
		if hasIPv4 || hasIPv6 {
			derValues = append(derValues, kvs.KeyValuePair{
				Key:   interfaces.InterfaceVrfKey(intf.GetName(), int(intf.GetVrf()), hasIPv4, hasIPv6),
				Value: &prototypes.Empty{},
			})
		}
	}

	// Rx mode
	if len(intf.GetRxModes()) > 0 {
		derValues = append(derValues, kvs.KeyValuePair{
			Key: interfaces.RxModesKey(intf.GetName()),
			Value: &interfaces.Interface{
				Name:    intf.GetName(),
				Type:    intf.GetType(),
				RxModes: intf.GetRxModes(),
			},
		})
	}

	// Rx placement
	for _, rxPlacement := range intf.GetRxPlacements() {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   interfaces.RxPlacementKey(intf.GetName(), rxPlacement.GetQueue()),
			Value: rxPlacement,
		})
	}

	// with-IP address (property)
	if len(intf.GetIpAddresses()) > 0 {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   interfaces.InterfaceWithIPKey(intf.GetName()),
			Value: &prototypes.Empty{},
		})
	}

	// TODO: define derived value for UP/DOWN state (needed for subinterfaces)

	return derValues
}

// getInterfaceMTU returns the interface MTU.
func (d *InterfaceDescriptor) getInterfaceMTU(intf *interfaces.Interface) uint32 {
	if mtu := intf.GetMtu(); mtu != 0 {
		return mtu
	}
	return d.defaultMtu /* still can be 0, i.e. undefined */
}

// resolveMemifSocketFilename returns memif socket filename ID.
// Registers it if does not exists yet.
func (d *InterfaceDescriptor) resolveMemifSocketFilename(memifIf *interfaces.MemifLink) (uint32, error) {
	socketFileName := d.getMemifSocketFilename(memifIf)
	registeredID, registered := d.memifSocketToID[socketFileName]
	if !registered {
		// Register new socket. ID is generated (default filename ID is 0, first is ID 1, second ID 2, etc)
		registeredID = uint32(len(d.memifSocketToID))
		err := d.ifHandler.RegisterMemifSocketFilename([]byte(socketFileName), registeredID)
		if err != nil {
			return 0, errors.Errorf("error registering socket file name %s (ID %d): %v", socketFileName, registeredID, err)
		}
		d.memifSocketToID[socketFileName] = registeredID
		d.log.Debugf("Memif socket filename %s registered under ID %d", socketFileName, registeredID)
	}
	return registeredID, nil
}

// getMemifSocketFilename returns the memif socket filename.
func (d *InterfaceDescriptor) getMemifSocketFilename(memif *interfaces.MemifLink) string {
	if socketFilename := memif.GetSocketFilename(); socketFilename != "" {
		return socketFilename
	}
	return d.defaultMemifSocketPath
}

// getMemifNumOfRxQueues returns the number of memif RX queues.
func (d *InterfaceDescriptor) getMemifNumOfRxQueues(memif *interfaces.MemifLink) uint32 {
	if memif.GetRxQueues() == 0 {
		return defaultMemifNumOfQueues
	}
	return memif.GetRxQueues()
}

// getMemifNumOfTxQueues returns the number of memif TX queues.
func (d *InterfaceDescriptor) getMemifNumOfTxQueues(memif *interfaces.MemifLink) uint32 {
	if memif.GetTxQueues() == 0 {
		return defaultMemifNumOfQueues
	}
	return memif.GetTxQueues()
}

// getMemifBufferSize returns the memif buffer size.
func (d *InterfaceDescriptor) getMemifBufferSize(memif *interfaces.MemifLink) uint32 {
	if memif.GetBufferSize() == 0 {
		return defaultMemifBufferSize
	}
	return memif.GetBufferSize()
}

// getMemifRingSize returns the memif ring size.
func (d *InterfaceDescriptor) getMemifRingSize(memif *interfaces.MemifLink) uint32 {
	if memif.GetRingSize() == 0 {
		return defaultMemifRingSize
	}
	return memif.GetRingSize()
}

// getTapConfig returns the TAP-specific configuration section (handling undefined attributes).
func getTapConfig(intf *interfaces.Interface) (tapLink *interfaces.TapLink) {
	tapLink = new(interfaces.TapLink)
	proto.Merge(tapLink, intf.GetTap())

	if tapLink.Version == 0 {
		tapLink.Version = 1
	}
	if tapLink.HostIfName == "" {
		tapLink.HostIfName = generateTAPHostName(intf.Name)
	}
	return tapLink
}

// generateTAPHostName (deterministically) generates the host name for a TAP interface.
func generateTAPHostName(tapName string) string {
	if tapName == "" {
		return ""
	}
	return fmt.Sprintf("tap-%d", fnvHash(tapName))
}

// fnvHash hashes string using fnv32a algorithm.
func fnvHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// equalStringSets compares two sets of strings for equality.
func equalStringSets(set1, set2 []string) bool {
	if len(set1) != len(set2) {
		return false
	}
	for _, item1 := range set1 {
		found := false
		for _, item2 := range set2 {
			if item1 == item2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// getIPAddressVersions returns two flags to tell whether the provided list of addresses
// contains IPv4 and/or IPv6 type addresses
func getIPAddressVersions(ipAddrs []string) (hasIPv4, hasIPv6 bool) {
	for _, ip := range ipAddrs {
		if strings.Contains(ip, ":") {
			hasIPv6 = true
		} else {
			hasIPv4 = true
		}
	}
	return
}
