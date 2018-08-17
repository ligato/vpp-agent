// Copyright (c) 2017 Cisco and/or its affiliates.
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

//go:generate protoc --proto_path=../model/interfaces --gogo_out=../model/interfaces ../model/interfaces/interfaces.proto
//go:generate protoc --proto_path=../model/bfd --gogo_out=../model/bfd ../model/bfd/bfd.proto

// Package ifplugin implements the Interface plugin that handles management
// of VPP interfaces.
package ifplugin

import (
	"bytes"
	"net"
	"strings"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/go-errors/errors"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
)

// InterfaceConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/interfaces/interfaces.proto"
// and stored in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1interface".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type InterfaceConfigurator struct {
	log logging.Logger

	linux interface{} // just flag if nil

	stopwatch *measure.Stopwatch // timer used to measure and store time

	swIfIndexes ifaceidx.SwIfIndexRW
	dhcpIndexes ifaceidx.DhcpIndexRW

	uIfaceCache         map[string]string                     // cache for not-configurable unnumbered interfaces. map[unumbered-iface-name]required-iface
	memifScCache        map[string]uint32                     // memif socket filename/ID cache (all known sockets). Note: do not remove items from the map
	vxlanMulticastCache map[string]*intf.Interfaces_Interface // cache for multicast VxLANs expecting another interface

	defaultMtu uint32 // default MTU value can be read from config

	afPacketConfigurator *AFPacketConfigurator

	vppCh govppapi.Channel

	// VPP API handler
	ifHandler vppcalls.IfVppAPI

	// Notification channels
	NotifChan chan govppapi.Message // to publish SwInterfaceDetails to interface_state.go
	DhcpChan  chan govppapi.Message // channel to receive DHCP notifications
}

// Init members (channels...) and start go routines
func (ic *InterfaceConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, linux interface{},
	notifChan chan govppapi.Message, defaultMtu uint32, enableStopwatch bool) (err error) {
	// Logger
	ic.log = logger.NewLogger("-if-conf")

	// Configurator-wide stopwatch instance
	if enableStopwatch {
		ic.stopwatch = measure.NewStopwatch("Interface-configurator", ic.log)
	}

	// State notification channel
	ic.NotifChan = notifChan

	// Config file data
	ic.defaultMtu = defaultMtu

	// VPP channel
	if ic.vppCh, err = goVppMux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create API channel: %v", err)
	}

	// VPP API handler
	ic.ifHandler = vppcalls.NewIfVppHandler(ic.vppCh, ic.log, ic.stopwatch)

	// Mappings
	ic.swIfIndexes = ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(ic.log, "sw_if_indexes", ifaceidx.IndexMetadata))
	ic.dhcpIndexes = ifaceidx.NewDHCPIndex(nametoidx.NewNameToIdx(ic.log, "dhcp_indices", ifaceidx.IndexDHCPMetadata))
	ic.uIfaceCache = make(map[string]string)
	ic.vxlanMulticastCache = make(map[string]*intf.Interfaces_Interface)
	ic.memifScCache = make(map[string]uint32)

	// Init AF-packet configurator
	ic.linux = linux
	ic.afPacketConfigurator = &AFPacketConfigurator{}
	ic.afPacketConfigurator.Init(ic.log, ic.ifHandler, ic.linux, ic.swIfIndexes)

	// DHCP channel
	ic.DhcpChan = make(chan govppapi.Message, 1)
	if _, err := ic.vppCh.SubscribeNotification(ic.DhcpChan, dhcp.NewDHCPComplEvent); err != nil {
		return err
	}

	go ic.watchDHCPNotifications()

	ic.log.Info("Interface configurator initialized")

	return nil
}

// Close GOVPP channel
func (ic *InterfaceConfigurator) Close() error {
	if err := safeclose.Close(ic.vppCh, ic.DhcpChan); err != nil {
		return errors.Errorf("failed to safeclose: %v", err)
	}
	return nil
}

// clearMapping prepares all in-memory-mappings and other cache fields. All previous cached entries are removed.
func (ic *InterfaceConfigurator) clearMapping() error {
	ic.swIfIndexes.Clear()
	ic.dhcpIndexes.Clear()
	ic.uIfaceCache = make(map[string]string)
	ic.vxlanMulticastCache = make(map[string]*intf.Interfaces_Interface)
	ic.memifScCache = make(map[string]uint32)

	ic.log.Debugf("interface configurator mapping cleared")
	return nil
}

// GetSwIfIndexes exposes interface name-to-index mapping
func (ic *InterfaceConfigurator) GetSwIfIndexes() ifaceidx.SwIfIndexRW {
	return ic.swIfIndexes
}

// GetDHCPIndexes exposes DHCP name-to-index mapping
func (ic *InterfaceConfigurator) GetDHCPIndexes() ifaceidx.DhcpIndexRW {
	return ic.dhcpIndexes
}

// IsSocketFilenameCached returns true if provided filename is presented in the cache
func (ic *InterfaceConfigurator) IsSocketFilenameCached(filename string) bool {
	_, ok := ic.memifScCache[filename]
	return ok
}

// IsUnnumberedIfCached returns true if provided interface is cached as unconfigurabel unnubered interface
func (ic *InterfaceConfigurator) IsUnnumberedIfCached(ifName string) bool {
	_, ok := ic.uIfaceCache[ifName]
	return ok
}

// IsMulticastVxLanIfCached returns true if provided interface is cached as VxLAN with missing multicast interface
func (ic *InterfaceConfigurator) IsMulticastVxLanIfCached(ifName string) bool {
	_, ok := ic.vxlanMulticastCache[ifName]
	return ok
}

// ConfigureVPPInterface reacts to a new northbound VPP interface config by creating and configuring
// the interface in the VPP network stack through the VPP binary API.
func (ic *InterfaceConfigurator) ConfigureVPPInterface(iface *intf.Interfaces_Interface) (err error) {
	var ifIdx uint32

	switch iface.Type {
	case intf.InterfaceType_TAP_INTERFACE:
		ifIdx, err = ic.ifHandler.AddTapInterface(iface.Name, iface.Tap)
	case intf.InterfaceType_MEMORY_INTERFACE:
		var id uint32 // Memif socket id
		if id, err = ic.resolveMemifSocketFilename(iface.Memif); err != nil {
			return err
		}
		ifIdx, err = ic.ifHandler.AddMemifInterface(iface.Name, iface.Memif, id)
	case intf.InterfaceType_VXLAN_TUNNEL:
		// VxLAN multicast interface. Interrupt the processing if there is an error or interface was cached
		multicastIfIdx, cached, err := ic.getVxLanMulticast(iface)
		if err != nil || cached {
			return err
		}
		ifIdx, err = ic.ifHandler.AddVxlanTunnel(iface.Name, iface.Vxlan, iface.Vrf, multicastIfIdx)
	case intf.InterfaceType_SOFTWARE_LOOPBACK:
		ifIdx, err = ic.ifHandler.AddLoopbackInterface(iface.Name)
	case intf.InterfaceType_ETHERNET_CSMACD:
		var exists bool
		if ifIdx, _, exists = ic.swIfIndexes.LookupIdx(iface.Name); !exists {
			ic.log.Warnf("It is not yet supported to add (whitelist) a new physical interface")
			return nil
		}
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		var pending bool
		if ifIdx, pending, err = ic.afPacketConfigurator.ConfigureAfPacketInterface(iface); err != nil {
			return err
		}
		if pending {
			ic.log.Debugf("Af-packet interface %s cannot be created yet and will be configured later", iface)
			return nil
		}
	}
	if err != nil {
		return err
	}

	// Rx-mode
	if err := ic.configRxModeForInterface(iface, ifIdx); err != nil {
		return err
	}

	// Rx-placement TODO: simplify implementation for rx placement when the binary api call will be available (remove dump)
	if iface.RxPlacementSettings != nil {
		// Required in order to get vpp internal name. Must be called from here, calling in vppcalls causes
		// import cycle
		ifMap, err := ic.ifHandler.DumpInterfaces()
		if err != nil {
			return errors.Errorf("failed to dump interfaces: %v", err)
		}
		ifData, ok := ifMap[ifIdx]
		if !ok || ifData == nil {
			return errors.Errorf("set rx-placement failed, no data available for interface index %d", ifIdx)
		}
		if err := ic.ifHandler.SetRxPlacement(ifData.Meta.InternalName, iface.RxPlacementSettings); err != nil {
			return errors.Errorf("failed to set rx-placement for interface %s: %v", ifData.Interface.Name, err)
		}
	}

	// MAC address (optional, for af-packet is configured in different way)
	if iface.PhysAddress != "" && iface.Type != intf.InterfaceType_AF_PACKET_INTERFACE {
		if err := ic.ifHandler.SetInterfaceMac(ifIdx, iface.PhysAddress); err != nil {
			return errors.Errorf("failed to set MAC address %s to interface %s: %v",
				iface.PhysAddress, iface.Name, err)
		}
	}

	// DHCP client
	if iface.SetDhcpClient {
		if err := ic.ifHandler.SetInterfaceAsDHCPClient(ifIdx, iface.Name); err != nil {
			return errors.Errorf("failed to set interface %s as DHCP client", iface.Name)
		}
	}

	// Get IP addresses
	IPAddrs, err := addrs.StrAddrsToStruct(iface.IpAddresses)
	if err != nil {
		return errors.Errorf("failed to convert %s IP address list to IPNet structures: %v", iface.Name, err)
	}

	// VRF (optional, unavailable for VxLAN interfaces), has to be done before IP addresses are configured
	if iface.Type != intf.InterfaceType_VXLAN_TUNNEL {
		// Configured separately for IPv4/IPv6
		isIPv4, isIPv6 := getIPAddressVersions(IPAddrs)
		if isIPv4 {
			if err := ic.ifHandler.SetInterfaceVrf(ifIdx, iface.Vrf); err != nil {
				return errors.Errorf("failed to set interface %s as IPv4 VRF %d: %v", iface.Name, iface.Vrf, err)
			}
		}
		if isIPv6 {
			if err := ic.ifHandler.SetInterfaceVrfIPv6(ifIdx, iface.Vrf); err != nil {
				return errors.Errorf("failed to set interface %s as IPv6 VRF %d: %v", iface.Name, iface.Vrf, err)
			}
		}
	}

	// Configure IP addresses or unnumbered config
	if err := ic.configureIPAddresses(iface.Name, ifIdx, IPAddrs, iface.Unnumbered); err != nil {
		return err
	}

	// configure container IP address
	if iface.ContainerIpAddress != "" {
		if err := ic.ifHandler.AddContainerIP(ifIdx, iface.ContainerIpAddress); err != nil {
			return errors.Errorf("failed to add container IP address %s to interface %s: %v",
				iface.ContainerIpAddress, iface.Name, err)
		}
	}

	// configure mtu. Prefer value in interface config, otherwise set default value if defined
	if iface.Type != intf.InterfaceType_VXLAN_TUNNEL {
		mtuToConfigure := iface.Mtu
		if mtuToConfigure == 0 && ic.defaultMtu != 0 {
			mtuToConfigure = ic.defaultMtu
		}
		if mtuToConfigure != 0 {
			iface.Mtu = mtuToConfigure
			if err := ic.ifHandler.SetInterfaceMtu(ifIdx, mtuToConfigure); err != nil {
				return errors.Errorf("failed to set MTU %d to interface %s: %v", mtuToConfigure, iface.Name, err)
			}
		}
	}

	// register name to idx mapping if it is not an af_packet interface type (it is registered in ConfigureAfPacketInterface if needed)
	if iface.Type != intf.InterfaceType_AF_PACKET_INTERFACE {
		ic.swIfIndexes.RegisterName(iface.Name, ifIdx, iface)
		ic.log.Debugf("Interface %s registered to interface mapping", iface.Name)
	}

	// set interface up if enabled
	// NOTE: needs to be called after RegisterName, otherwise interface up/down notification won't map to a valid interface
	if iface.Enabled {
		if err := ic.ifHandler.InterfaceAdminUp(ifIdx); err != nil {
			return errors.Errorf("failed to set interface %s up: %v", iface.Name, err)
		}
	}

	// load interface state data for newly added interface (no way to filter by swIfIndex, need to dump all of them)
	if err := ic.propagateIfDetailsToStatus(); err != nil {
		return err
	}

	// Check whether there is no VxLAN interface waiting on created one as a multicast
	if err := ic.resolveCachedVxLANMulticasts(iface.Name); err != nil {
		return err
	}

	ic.log.Infof("Interface %s configured", iface.Name)

	return nil
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
func (ic *InterfaceConfigurator) configRxModeForInterface(iface *intf.Interfaces_Interface, ifIdx uint32) error {
	rxModeSettings := iface.RxModeSettings
	if rxModeSettings != nil {
		switch iface.Type {
		case intf.InterfaceType_ETHERNET_CSMACD:
			if rxModeSettings.RxMode == intf.RxModeType_POLLING {
				return ic.configRxMode(iface, ifIdx, rxModeSettings)
			}
		default:
			return ic.configRxMode(iface, ifIdx, rxModeSettings)
		}
	}
	return nil
}

// Call specific vpp API method for setting rx-mode
func (ic *InterfaceConfigurator) configRxMode(iface *intf.Interfaces_Interface, ifIdx uint32, rxModeSettings *intf.Interfaces_Interface_RxModeSettings) error {
	if err := ic.ifHandler.SetRxMode(ifIdx, rxModeSettings); err != nil {
		return errors.Errorf("failed to set Rx-mode for interface %s: %v", iface.Name, err)
	}
	return nil
}

func (ic *InterfaceConfigurator) configureIPAddresses(ifName string, ifIdx uint32, addresses []*net.IPNet, unnumbered *intf.Interfaces_Interface_Unnumbered) error {
	if unnumbered != nil && unnumbered.IsUnnumbered {
		ifWithIP := unnumbered.InterfaceWithIp
		if ifWithIP == "" {
			return errors.Errorf("unnubered interface %s has no interface with IP address set", ifName)
		}
		ifIdxIP, _, found := ic.swIfIndexes.LookupIdx(ifWithIP)
		if !found {
			// cache not-configurable interface
			ic.uIfaceCache[ifName] = ifWithIP
			ic.log.Debugf("unnubered interface %s moved to cache (requires IP address from non-existing %s)", ifName, ifWithIP)
			return nil
		}
		// Set interface as un-numbered
		if err := ic.ifHandler.SetUnnumberedIP(ifIdx, ifIdxIP); err != nil {
			return errors.Errorf("failed to set interface %d as unnumbered for %d: %v", ifIdxIP, ifName, err)
		}
	}

	// configure optional ip address
	for _, address := range addresses {
		if err := ic.ifHandler.AddInterfaceIP(ifIdx, address); err != nil {
			return errors.Errorf("adding IP address %s to interface %s failed: %v", address.String(), ifName, err)
		}
	}

	// with ip address configured, the interface can be used as a source for un-numbered interfaces (if any)
	if err := ic.resolveDependentUnnumberedInterfaces(ifName, ifIdx); err != nil {
		return err
	}
	return nil
}

func (ic *InterfaceConfigurator) removeIPAddresses(ifIdx uint32, addresses []*net.IPNet, unnumbered *intf.Interfaces_Interface_Unnumbered) error {
	if unnumbered != nil && unnumbered.IsUnnumbered {
		// Set interface as un-numbered
		if err := ic.ifHandler.UnsetUnnumberedIP(ifIdx); err != nil {
			return errors.Errorf("faield to unset unnumbered IP for interface %d: %v", ifIdx, err)
		}
	}

	// delete IP Addresses
	for _, addr := range addresses {
		err := ic.ifHandler.DelInterfaceIP(ifIdx, addr)
		if err != nil {
			return errors.Errorf("deleting IP address %s from interface %d failed: %v", addr, ifIdx, err)
		}
	}

	return nil
}

// Iterate over all un-numbered interfaces in cache (which could not be configured before) and find all interfaces
// dependent on the provided one
func (ic *InterfaceConfigurator) resolveDependentUnnumberedInterfaces(ifNameIP string, ifIdxIP uint32) error {
	for uIface, ifWithIP := range ic.uIfaceCache {
		if ifWithIP == ifNameIP {
			// find index of the dependent interface
			uIdx, _, found := ic.swIfIndexes.LookupIdx(uIface)
			if !found {
				delete(ic.uIfaceCache, uIface)
				ic.log.Debugf("Unnumbered interface %s removed from cache (not found)", uIface)
				continue
			}
			if err := ic.ifHandler.SetUnnumberedIP(uIdx, ifIdxIP); err != nil {
				return errors.Errorf("setting unnumbered IP %d for %s failed: %v", ifIdxIP, uIdx, err)
			}
			delete(ic.uIfaceCache, uIface)
			ic.log.Debugf("Unnumbered interface %s set and removed from cache", uIface)
		}
	}
	return nil
}

// ModifyVPPInterface applies changes in the NB configuration of a VPP interface into the running VPP
// through the VPP binary API.
func (ic *InterfaceConfigurator) ModifyVPPInterface(newConfig *intf.Interfaces_Interface,
	oldConfig *intf.Interfaces_Interface) error {

	// Recreate pending Af-packet
	if newConfig.Type == intf.InterfaceType_AF_PACKET_INTERFACE && ic.afPacketConfigurator.IsPendingAfPacket(oldConfig) {
		return ic.recreateVPPInterface(newConfig, oldConfig, 0)
	}

	// Re-create cached VxLAN
	if newConfig.Type == intf.InterfaceType_VXLAN_TUNNEL {
		if _, ok := ic.vxlanMulticastCache[newConfig.Name]; ok {
			delete(ic.vxlanMulticastCache, newConfig.Name)
			ic.log.Debugf("Interface %s removed from VxLAN multicast cache, will be configured", newConfig.Name)
			return ic.ConfigureVPPInterface(newConfig)
		}
	}

	// Lookup index. If not found, create interface a a new on.
	ifIdx, meta, found := ic.swIfIndexes.LookupIdx(newConfig.Name)
	if !found {
		ic.log.Warnf("Modify interface %s: index was not found in the mapping, creating as a new one", newConfig.Name)
		return ic.ConfigureVPPInterface(newConfig)
	}

	if err := ic.modifyVPPInterface(newConfig, oldConfig, ifIdx, meta.Type); err != nil {
		return err
	}

	ic.log.Infof("Interface %s modified", newConfig.Name)

	return nil
}

// ModifyVPPInterface applies changes in the NB configuration of a VPP interface into the running VPP
// through the VPP binary API.
func (ic *InterfaceConfigurator) modifyVPPInterface(newConfig, oldConfig *intf.Interfaces_Interface,
	ifIdx uint32, ifaceType intf.InterfaceType) (err error) {

	switch ifaceType {
	case intf.InterfaceType_TAP_INTERFACE:
		if !ic.canTapBeModifWithoutDelete(newConfig.Tap, oldConfig.Tap) {
			return ic.recreateVPPInterface(newConfig, oldConfig, ifIdx)
		}
	case intf.InterfaceType_MEMORY_INTERFACE:
		if !ic.canMemifBeModifWithoutDelete(newConfig.Memif, oldConfig.Memif) {
			return ic.recreateVPPInterface(newConfig, oldConfig, ifIdx)
		}
	case intf.InterfaceType_VXLAN_TUNNEL:
		if !ic.canVxlanBeModifWithoutDelete(newConfig.Vxlan, oldConfig.Vxlan) ||
			oldConfig.Vrf != newConfig.Vrf {
			return ic.recreateVPPInterface(newConfig, oldConfig, ifIdx)
		}
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		recreate, err := ic.afPacketConfigurator.ModifyAfPacketInterface(newConfig, oldConfig)
		if err != nil {
			return err
		}
		if recreate {
			return ic.recreateVPPInterface(newConfig, oldConfig, ifIdx)
		}
	case intf.InterfaceType_SOFTWARE_LOOPBACK:
	case intf.InterfaceType_ETHERNET_CSMACD:
	}

	// Rx-mode
	if !(oldConfig.RxModeSettings == nil && newConfig.RxModeSettings == nil) {
		if err := ic.modifyRxModeForInterfaces(oldConfig, newConfig, ifIdx); err != nil {
			return err
		}
	}

	// Rx-placement
	if newConfig.RxPlacementSettings != nil {
		// Required in order to get vpp internal name. Must be called from here, calling in vppcalls causes
		// import cycle
		ifMap, err := ic.ifHandler.DumpInterfaces()
		if err != nil {
			return errors.Errorf("failed to dump interfaces: %v", err)
		}
		ifData, ok := ifMap[ifIdx]
		if !ok || ifData == nil {
			return errors.Errorf("set rx-placement for new config failed, no data available for interface index %d", ifIdx)
		}
		if err := ic.ifHandler.SetRxPlacement(ifData.Meta.InternalName, newConfig.RxPlacementSettings); err != nil {
			return errors.Errorf("failed to set rx-placement for interface %s: %v", newConfig.Name, err)
		}
	}

	// Admin status
	if newConfig.Enabled != oldConfig.Enabled {
		if newConfig.Enabled {
			if err = ic.ifHandler.InterfaceAdminUp(ifIdx); err != nil {
				return errors.Errorf("failed to set interface %s up: %v", newConfig.Name, err)
			}
		} else {
			if err = ic.ifHandler.InterfaceAdminDown(ifIdx); err != nil {
				return errors.Errorf("failed to set interface %s down: %v", newConfig.Name, err)
			}
		}
	}

	// Configure new mac address if set (and only if it was changed)
	if newConfig.PhysAddress != "" && newConfig.PhysAddress != oldConfig.PhysAddress {
		if err := ic.ifHandler.SetInterfaceMac(ifIdx, newConfig.PhysAddress); err != nil {
			return errors.Errorf("setting interface %s MAC address %s failed: %v",
				newConfig.Name, newConfig.PhysAddress, err)
		}
	}

	// Reconfigure DHCP
	if oldConfig.SetDhcpClient != newConfig.SetDhcpClient {
		if newConfig.SetDhcpClient {
			if err := ic.ifHandler.SetInterfaceAsDHCPClient(ifIdx, newConfig.Name); err != nil {
				return errors.Errorf("failed to set interface %s as DHCP client: %v", newConfig.Name, err)
			}
		} else {
			if err := ic.ifHandler.UnsetInterfaceAsDHCPClient(ifIdx, newConfig.Name); err != nil {
				return errors.Errorf("failed to unset interface %s as DHCP client: %v", newConfig.Name, err)
			} else {
				// Remove from DHCP mapping
				ic.dhcpIndexes.UnregisterName(newConfig.Name)
				ic.log.Debugf("Interface %s unregistered as DHCP client", oldConfig.Name)
			}
		}
	}

	// Ip addresses
	newAddrs, err := addrs.StrAddrsToStruct(newConfig.IpAddresses)
	if err != nil {
		return errors.Errorf("failed to convert %s IP address list to IPNet structures: %v", newConfig.Name, err)
	}
	oldAddrs, err := addrs.StrAddrsToStruct(oldConfig.IpAddresses)
	if err != nil {
		return errors.Errorf("failed to convert %s IP address list to IPNet structures: %v", oldConfig.Name, err)
	}

	// Reconfigure VRF
	if ifaceType != intf.InterfaceType_VXLAN_TUNNEL {
		// Interface must not have IP when setting VRF
		if err := ic.removeIPAddresses(ifIdx, oldAddrs, oldConfig.Unnumbered); err != nil {
			return err
		}

		// Get VRF IP version using new list of addresses. During modify, interface VRF IP version
		// should be updated as well.
		isIPv4, isIPv6 := getIPAddressVersions(newAddrs)
		if isIPv4 {
			if err := ic.ifHandler.SetInterfaceVrf(ifIdx, newConfig.Vrf); err != nil {
				return errors.Errorf("failed to set IPv4 VRF %d for interface %s: %v",
					newConfig.Vrf, newConfig.Name, err)
			}
		}
		if isIPv6 {
			if err := ic.ifHandler.SetInterfaceVrfIPv6(ifIdx, newConfig.Vrf); err != nil {
				return errors.Errorf("failed to set IPv6 VRF %d for interface %s: %v",
					newConfig.Vrf, newConfig.Name, err)
			}
		}

		if err = ic.configureIPAddresses(newConfig.Name, ifIdx, newAddrs, newConfig.Unnumbered); err != nil {
			return err
		}
	}

	// Container ip address
	if newConfig.ContainerIpAddress != oldConfig.ContainerIpAddress {
		if err := ic.ifHandler.AddContainerIP(ifIdx, newConfig.ContainerIpAddress); err != nil {
			return errors.Errorf("failed to add container IP %s to interface %s: %v",
				newConfig.ContainerIpAddress, newConfig.Name, err)
		}
	}

	// Set MTU if changed in interface config
	if newConfig.Mtu != 0 && newConfig.Mtu != oldConfig.Mtu {
		if err := ic.ifHandler.SetInterfaceMtu(ifIdx, newConfig.Mtu); err != nil {
			return errors.Errorf("failed to set MTU to interface %s: %v", newConfig.Name, err)
		}
	} else if newConfig.Mtu == 0 && ic.defaultMtu != 0 {
		if err := ic.ifHandler.SetInterfaceMtu(ifIdx, ic.defaultMtu); err != nil {
			return errors.Errorf("failed to set MTU to interface %s: %v", newConfig.Name, err)
		}
	}

	ic.swIfIndexes.UpdateMetadata(newConfig.Name, newConfig)
	ic.log.Debugf("Metadata updated in interface mapping for %s", newConfig.Name)

	return nil
}

/**
Modify rx-mode on specified VPP interface
*/
func (ic *InterfaceConfigurator) modifyRxModeForInterfaces(oldIntf, newIntf *intf.Interfaces_Interface, ifIdx uint32) error {
	oldRx := oldIntf.RxModeSettings
	newRx := newIntf.RxModeSettings

	if oldRx == nil && newRx != nil || oldRx != nil && newRx == nil || *oldRx != *newRx {
		// If new rx mode is nil, value is reset to default version (differs for interface types)
		switch newIntf.Type {
		case intf.InterfaceType_ETHERNET_CSMACD:
			if newRx == nil {
				return ic.modifyRxMode(newIntf.Name, ifIdx, &intf.Interfaces_Interface_RxModeSettings{RxMode: intf.RxModeType_POLLING})
			} else if newRx.RxMode != intf.RxModeType_POLLING {
				return errors.Errorf("attempt to set unsupported rx-mode %s to Ethernet interface %s", newRx.RxMode, newIntf.Name)
			}
		case intf.InterfaceType_AF_PACKET_INTERFACE:
			if newRx == nil {
				return ic.modifyRxMode(newIntf.Name, ifIdx, &intf.Interfaces_Interface_RxModeSettings{RxMode: intf.RxModeType_INTERRUPT})
			}
		default: // All the other interface types
			if newRx == nil {
				return ic.modifyRxMode(newIntf.Name, ifIdx, &intf.Interfaces_Interface_RxModeSettings{RxMode: intf.RxModeType_DEFAULT})
			}
		}
		return ic.modifyRxMode(newIntf.Name, ifIdx, newRx)
	}

	return nil
}

/**
Direct call of vpp api to change rx-mode of specified interface
*/
func (ic *InterfaceConfigurator) modifyRxMode(ifName string, ifIdx uint32, rxMode *intf.Interfaces_Interface_RxModeSettings) error {
	if err := ic.ifHandler.SetRxMode(ifIdx, rxMode); err != nil {
		return errors.Errorf("failed to set rx-mode for interface %s: %v", ifName, err)
	}
	return nil
}

// recreateVPPInterface removes and creates an interface from scratch.
func (ic *InterfaceConfigurator) recreateVPPInterface(newConfig *intf.Interfaces_Interface,
	oldConfig *intf.Interfaces_Interface, ifIdx uint32) error {

	if oldConfig.Type == intf.InterfaceType_AF_PACKET_INTERFACE {
		if err := ic.afPacketConfigurator.DeleteAfPacketInterface(oldConfig, ifIdx); err != nil {
			return err
		}
	} else {
		if err := ic.deleteVPPInterface(oldConfig, ifIdx); err != nil {
			return err
		}
	}
	return ic.ConfigureVPPInterface(newConfig)
}

// DeleteVPPInterface reacts to a removed NB configuration of a VPP interface.
// It results in the interface being removed from VPP.
func (ic *InterfaceConfigurator) DeleteVPPInterface(iface *intf.Interfaces_Interface) error {
	// Remove VxLAN from cache if exists
	if iface.Type == intf.InterfaceType_VXLAN_TUNNEL {
		if _, ok := ic.vxlanMulticastCache[iface.Name]; ok {
			delete(ic.vxlanMulticastCache, iface.Name)
			ic.log.Debugf("Interface %s removed from VxLAN multicast cache, will be removed", iface.Name)
			return nil
		}
	}

	if ic.afPacketConfigurator.IsPendingAfPacket(iface) {
		ifIdx, _, found := ic.afPacketConfigurator.ifIndexes.LookupIdx(iface.Name)
		if !found {
			// Just remove from cache
			ic.afPacketConfigurator.removeFromCache(iface)
			return nil
		}

		return ic.afPacketConfigurator.DeleteAfPacketInterface(iface, ifIdx)
	}

	// unregister name to init mapping (following triggers notifications for all subscribers, skip physical interfaces)
	if iface.Type != intf.InterfaceType_ETHERNET_CSMACD {
		ifIdx, prev, found := ic.swIfIndexes.UnregisterName(iface.Name)
		if !found {
			return errors.Errorf("Unable to find interface %s in the mapping", iface.Name)
		}
		ic.log.Debugf("Interface %s unregistered from interface mapping", iface.Name)

		// delete from unnumbered map (if the interface is present)
		delete(ic.uIfaceCache, iface.Name)
		ic.log.Debugf("Unnumbered interface %s removed from cache (will be removed)", iface.Name)

		if err := ic.deleteVPPInterface(prev, ifIdx); err != nil {
			return err
		}
	} else {
		// Find index of the Physical interface and un-configure it
		ifIdx, prev, found := ic.swIfIndexes.LookupIdx(iface.Name)
		if !found {
			return errors.Errorf("unable to find index for physical interface %s, cannot delete", iface.Name)
		}
		if err := ic.deleteVPPInterface(prev, ifIdx); err != nil {
			return err
		}
	}

	ic.log.Infof("Interface %v removed", iface.Name)

	return nil
}

func (ic *InterfaceConfigurator) deleteVPPInterface(oldConfig *intf.Interfaces_Interface, ifIdx uint32) error {
	// Skip setting interface to ADMIN_DOWN unless the type AF_PACKET_INTERFACE
	if oldConfig.Type != intf.InterfaceType_AF_PACKET_INTERFACE {
		if err := ic.ifHandler.InterfaceAdminDown(ifIdx); err != nil {
			return errors.Errorf("failed to set interface %s down: %v", oldConfig.Name, err)
		}
	}

	// Remove DHCP if it was set
	if oldConfig.SetDhcpClient {
		if err := ic.ifHandler.UnsetInterfaceAsDHCPClient(ifIdx, oldConfig.Name); err != nil {
			return errors.Errorf("failed to unset interface %s as DHCP client: %v", oldConfig.Name, err)
		}
		// Remove from DHCP mapping
		ic.dhcpIndexes.UnregisterName(oldConfig.Name)
		ic.log.Debugf("Interface %v unregistered as DHCP client", oldConfig.Name)
	}

	if oldConfig.ContainerIpAddress != "" {
		if err := ic.ifHandler.DelContainerIP(ifIdx, oldConfig.ContainerIpAddress); err != nil {
			return errors.Errorf("failed to delete container IP %s from interface %s: %v",
				oldConfig.ContainerIpAddress, oldConfig.Name, err)
		}
	}

	for i, oldIP := range oldConfig.IpAddresses {
		if strings.HasPrefix(oldIP, "fe80") {
			// TODO: skip link local addresses (possible workaround for af_packet)
			oldConfig.IpAddresses = append(oldConfig.IpAddresses[:i], oldConfig.IpAddresses[i+1:]...)
			ic.log.Debugf("delete vpp interface %s: link local address %s skipped", oldConfig.Name, oldIP)
		}
	}
	oldAddrs, err := addrs.StrAddrsToStruct(oldConfig.IpAddresses)
	if err != nil {
		return errors.Errorf("failed to convert %s IP address list to IPNet structures: %v", oldConfig.Name, err)
	}
	for _, oldAddr := range oldAddrs {
		if err := ic.ifHandler.DelInterfaceIP(ifIdx, oldAddr); err != nil {
			return errors.Errorf("failed to remove IP address %s from interface %s: %v",
				oldAddr, oldConfig.Name, err)
		}
	}

	// let's try to do following even if previously error occurred
	switch oldConfig.Type {
	case intf.InterfaceType_TAP_INTERFACE:
		err = ic.ifHandler.DeleteTapInterface(oldConfig.Name, ifIdx, oldConfig.Tap.Version)
	case intf.InterfaceType_MEMORY_INTERFACE:
		err = ic.ifHandler.DeleteMemifInterface(oldConfig.Name, ifIdx)
	case intf.InterfaceType_VXLAN_TUNNEL:
		err = ic.ifHandler.DeleteVxlanTunnel(oldConfig.Name, ifIdx, oldConfig.GetVxlan())
	case intf.InterfaceType_SOFTWARE_LOOPBACK:
		err = ic.ifHandler.DeleteLoopbackInterface(oldConfig.Name, ifIdx)
	case intf.InterfaceType_ETHERNET_CSMACD:
		ic.log.Debugf("Interface removal skipped: cannot remove (blacklist) physical interface") // Not an error
		return nil
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		err = ic.afPacketConfigurator.DeleteAfPacketInterface(oldConfig, ifIdx)
	}
	if err != nil {
		return errors.Errorf("failed to remove interface %s, index %d: %v", oldConfig.Name, ifIdx, err)
	}

	return nil
}

// ResolveCreatedLinuxInterface reacts to a newly created Linux interface.
func (ic *InterfaceConfigurator) ResolveCreatedLinuxInterface(ifName, hostIfName string, ifIdx uint32) error {
	pendingAfpacket, err := ic.afPacketConfigurator.ResolveCreatedLinuxInterface(ifName, hostIfName, ifIdx)
	if err != nil {
		return err
	}
	if pendingAfpacket != nil {
		// there is a pending af-packet that can be now configured
		return ic.ConfigureVPPInterface(pendingAfpacket)
	}
	return nil
}

// ResolveDeletedLinuxInterface reacts to a removed Linux interface.
func (ic *InterfaceConfigurator) ResolveDeletedLinuxInterface(ifName, hostIfName string, ifIdx uint32) error {
	return ic.afPacketConfigurator.ResolveDeletedLinuxInterface(ifName, hostIfName, ifIdx)
}

// PropagateIfDetailsToStatus looks up all VPP interfaces
func (ic *InterfaceConfigurator) propagateIfDetailsToStatus() error {
	start := time.Now()
	req := &interfaces.SwInterfaceDump{}
	reqCtx := ic.vppCh.SendMultiRequest(req)

	for {
		msg := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			return errors.Errorf("failed to receive interface dump details: %v", err)
		}

		_, _, found := ic.swIfIndexes.LookupName(msg.SwIfIndex)
		if !found {
			ic.log.Warnf("Unregistered interface %v with ID %v found on vpp",
				string(bytes.SplitN(msg.InterfaceName, []byte{0x00}, 2)[0]), msg.SwIfIndex)
			// Do not register unknown interface here, cuz it may cause inconsistencies in the ifplugin.
			// All new interfaces should be registered during configuration
			continue
		}

		// Propagate interface state information to notification channel.
		ic.NotifChan <- msg
	}

	// SwInterfaceSetFlags time
	if ic.stopwatch != nil {
		timeLog := measure.GetTimeLog(interfaces.SwInterfaceSetFlags{}, ic.stopwatch)
		timeLog.LogTimeEntry(time.Since(start))
	}

	return nil
}

// returns memif socket filename ID. Registers it if does not exists yet
func (ic *InterfaceConfigurator) resolveMemifSocketFilename(memifIf *intf.Interfaces_Interface_Memif) (uint32, error) {
	if memifIf.SocketFilename == "" {
		return 0, errors.Errorf("memif configuration does not contain socket file name")
	}
	registeredID, ok := ic.memifScCache[memifIf.SocketFilename]
	if !ok {
		// Register new socket. ID is generated (default filename ID is 0, first is ID 1, second ID 2, etc)
		registeredID = uint32(len(ic.memifScCache))
		err := ic.ifHandler.RegisterMemifSocketFilename([]byte(memifIf.SocketFilename), registeredID)
		if err != nil {
			return 0, errors.Errorf("error registering socket file name %s (ID %d): %v", memifIf.SocketFilename, registeredID, err)
		}
		ic.memifScCache[memifIf.SocketFilename] = registeredID
		ic.log.Debugf("Memif socket filename %s registered under ID %d", memifIf.SocketFilename, registeredID)
	}
	return registeredID, nil
}

// Returns VxLAN multicast interface index if set and exists. Returns index of the interface an whether the vxlan was cached.
func (ic *InterfaceConfigurator) getVxLanMulticast(vxlan *intf.Interfaces_Interface) (ifIdx uint32, cached bool, err error) {
	if vxlan.Vxlan == nil {
		ic.log.Debugf("VxLAN multicast: no data available for %s", vxlan.Name)
		return 0, false, nil
	}
	if vxlan.Vxlan.Multicast == "" {
		ic.log.Debugf("VxLAN %s has no multicast interface defined", vxlan.Name)
		return 0, false, nil
	}
	mcIfIdx, mcIf, found := ic.swIfIndexes.LookupIdx(vxlan.Vxlan.Multicast)
	if !found {
		ic.log.Infof("multicast interface %s not found, %s is cached", vxlan.Vxlan.Multicast, vxlan.Name)
		ic.vxlanMulticastCache[vxlan.Name] = vxlan
		ic.log.Debugf("Interface %s added to VxLAN multicast cache", vxlan.Name)
		return 0, true, nil
	}
	// Check wheteher at least one of the addresses is from multicast range
	if len(mcIf.IpAddresses) == 0 {
		return 0, false, errors.Errorf("VxLAN %s refers to multicast interface %s which does not have any IP address",
			vxlan.Name, mcIf.Name)
	}
	var IPVerified bool
	for _, mcIfAddr := range mcIf.IpAddresses {
		mcIfAddrWithoutMask := strings.Split(mcIfAddr, "/")[0]
		IPVerified = net.ParseIP(mcIfAddrWithoutMask).IsMulticast()
		if IPVerified {
			if vxlan.Vxlan.DstAddress != mcIfAddr {
				ic.log.Warn("VxLAN %s contains destination address %s which will be replaced with multicast %s",
					vxlan.Name, vxlan.Vxlan.DstAddress, mcIfAddr)
			}
			vxlan.Vxlan.DstAddress = mcIfAddrWithoutMask
			break
		}
	}
	if !IPVerified {
		return 0, false, errors.Errorf("VxLAN %s refers to multicast interface %s which does not have multicast IP address",
			vxlan.Name, mcIf.Name)
	}

	return mcIfIdx, false, nil
}

// Look over cached VxLAN multicast interfaces and configure them if possible
func (ic *InterfaceConfigurator) resolveCachedVxLANMulticasts(createdIfName string) error {
	for vxlanName, vxlan := range ic.vxlanMulticastCache {
		if vxlan.Vxlan.Multicast == createdIfName {
			delete(ic.vxlanMulticastCache, vxlanName)
			ic.log.Debugf("Interface %s removed from VxLAN multicast cache, will be configured", vxlanName)
			if err := ic.ConfigureVPPInterface(vxlan); err != nil {
				return errors.Errorf("failed to configure VPP interface %s as VxLAN multicast: %v",
					createdIfName, err)
			}
		}
	}

	return nil
}

func (ic *InterfaceConfigurator) canMemifBeModifWithoutDelete(newConfig *intf.Interfaces_Interface_Memif, oldConfig *intf.Interfaces_Interface_Memif) bool {
	if newConfig == nil || oldConfig == nil {
		return true
	}

	if *newConfig != *oldConfig {
		ic.log.Debug("Difference between new & old config causing recreation of memif")
		return false
	}

	return true
}

func (ic *InterfaceConfigurator) canVxlanBeModifWithoutDelete(newConfig *intf.Interfaces_Interface_Vxlan, oldConfig *intf.Interfaces_Interface_Vxlan) bool {
	if newConfig == nil || oldConfig == nil {
		return true
	}
	if *newConfig != *oldConfig {
		ic.log.Debug("Difference between new & old config causing recreation of VxLAN")
		return false
	}

	return true
}

func (ic *InterfaceConfigurator) canTapBeModifWithoutDelete(newConfig *intf.Interfaces_Interface_Tap, oldConfig *intf.Interfaces_Interface_Tap) bool {
	if newConfig == nil || oldConfig == nil {
		return true
	}
	if *newConfig != *oldConfig {
		ic.log.Debug("Difference between new & old config causing recreation of tap")
		return false
	}

	return true
}

// watch and process DHCP notifications. DHCP configuration is registered to dhcp mapping for every interface
func (ic *InterfaceConfigurator) watchDHCPNotifications() {
	ic.log.Debug("Started watcher on DHCP notifications")

	for {
		select {
		case notification := <-ic.DhcpChan:
			switch dhcpNotif := notification.(type) {
			case *dhcp.DHCPComplEvent:
				var ipAddr, rIPAddr net.IP = dhcpNotif.Lease.HostAddress, dhcpNotif.Lease.RouterAddress
				var hwAddr net.HardwareAddr = dhcpNotif.Lease.HostMac
				var ipStr, rIPStr string

				name := string(bytes.SplitN(dhcpNotif.Lease.Hostname, []byte{0x00}, 2)[0])

				if dhcpNotif.Lease.IsIPv6 == 1 {
					ipStr = ipAddr.To16().String()
					rIPStr = rIPAddr.To16().String()
				} else {
					ipStr = ipAddr[:4].To4().String()
					rIPStr = rIPAddr[:4].To4().String()
				}

				ic.log.Debugf("DHCP assigned %v to interface %q (router address %v)", ipStr, name, rIPStr)

				ifIdx, _, found := ic.swIfIndexes.LookupIdx(name)
				if !found {
					ic.log.Warnf("Expected interface %v not found in the mapping", name)
					continue
				}

				// Register DHCP config
				ic.dhcpIndexes.RegisterName(name, ifIdx, &ifaceidx.DHCPSettings{
					IfName: name,
					IsIPv6: func(isIPv6 uint8) bool {
						if isIPv6 == 1 {
							return true
						}
						return false
					}(dhcpNotif.Lease.IsIPv6),
					IPAddress:     ipStr,
					Mask:          uint32(dhcpNotif.Lease.MaskWidth),
					PhysAddress:   hwAddr.String(),
					RouterAddress: rIPStr,
				})
				ic.log.Debugf("Interface %s registered as DHCP client", name)
			}
		}
	}
}

// If not nil, prints error including stack trace. The same value is also returned, so it can be easily propagated further
func (ic *InterfaceConfigurator) LogError(err error) error {
	if err == nil {
		return nil
	}
	ic.log.WithField("logger", ic.log).Errorf(string(err.Error() + "\n" + string(err.(*errors.Error).Stack())))
	return err
}

// Returns two flags, whether provided list of addresses contains IPv4 and/or IPv6 type addresses
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
