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

//go:generate protoc --proto_path=../common/model/interfaces --gogo_out=../common/model/interfaces ../common/model/interfaces/interfaces.proto
//go:generate protoc --proto_path=../common/model/bfd --gogo_out=../common/model/bfd ../common/model/bfd/bfd.proto

//go:generate binapi-generator --input-file=/usr/share/vpp/api/af_packet.api.json --output-dir=../common/bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/bfd.api.json --output-dir=../common/bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/dhcp.api.json --output-dir=../common/bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/interface.api.json --output-dir=../common/bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/ip.api.json --output-dir=../common/bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/memif.api.json --output-dir=../common/bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/stats.api.json --output-dir=../common/bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/tap.api.json --output-dir=../common/bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/tapv2.api.json --output-dir=../common/bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/vpe.api.json --output-dir=../common/bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/vxlan.api.json --output-dir=../common/bin_api

// Package ifplugin implements the Interface plugin that handles management
// of VPP interfaces.
package ifplugin

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/dhcp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/memif"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppdump"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// InterfaceConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/interfaces/interfaces.proto"
// and stored in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1interface".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type InterfaceConfigurator struct {
	Log logging.Logger

	GoVppmux     govppmux.API
	ServiceLabel servicelabel.ReaderAPI
	Linux        interface{} // just flag if nil

	Stopwatch *measure.Stopwatch // timer used to measure and store time

	swIfIndexes ifaceidx.SwIfIndexRW
	dhcpIndices ifaceidx.DhcpIndexRW

	uIfaceCache  map[string]string // cache for not-configurable unnumbered interfaces. map[unumbered-iface-name]required-iface
	memifScCache map[string]uint32 // memif socket filename/ID cache (all known sockets). Note: do not remove items from the map

	mtu uint32 // default MTU value can be read from config

	afPacketConfigurator *AFPacketConfigurator

	vppCh *govppapi.Channel

	// Notification channels
	notifChan chan govppapi.Message // to publish SwInterfaceDetails to interface_state.go
	dhcpChan  chan govppapi.Message // channel to receive DHCP notifications
}

// Init members (channels...) and start go routines
func (plugin *InterfaceConfigurator) Init(swIfIndexes ifaceidx.SwIfIndexRW, dhcpIndices ifaceidx.DhcpIndexRW,
	mtu uint32, notifChan chan govppapi.Message) (err error) {
	plugin.Log.Debug("Initializing Interface configurator")

	plugin.swIfIndexes = swIfIndexes
	plugin.dhcpIndices = dhcpIndices
	plugin.notifChan = notifChan
	plugin.mtu = mtu

	plugin.vppCh, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}
	if err := vppcalls.CheckMsgCompatibilityForInterface(plugin.Log, plugin.vppCh); err != nil {
		return err
	}

	// Init AF-packet configurator
	plugin.afPacketConfigurator = &AFPacketConfigurator{Logger: plugin.Log, Linux: plugin.Linux, SwIfIndexes: plugin.swIfIndexes, Stopwatch: plugin.Stopwatch}
	plugin.afPacketConfigurator.Init(plugin.vppCh)

	plugin.uIfaceCache = make(map[string]string)
	// Obtain registered socket filenames
	plugin.memifScCache, err = vppdump.DumpMemifSocketDetails(plugin.Log, plugin.vppCh,
		measure.GetTimeLog(memif.MemifSocketFilenameDump{}, plugin.Stopwatch))
	if err != nil {
		return err
	}

	plugin.dhcpChan = make(chan govppapi.Message, 1)
	if _, err := plugin.vppCh.SubscribeNotification(plugin.dhcpChan, dhcp.NewDhcpComplEvent); err != nil {
		return err
	}
	go plugin.watchDHCPNotifications()

	return nil
}

// Close GOVPP channel
func (plugin *InterfaceConfigurator) Close() error {
	_, err := safeclose.CloseAll(plugin.vppCh, plugin.dhcpChan)
	return err
}

// PropagateIfDetailsToStatus looks up all VPP interfaces
func (plugin *InterfaceConfigurator) PropagateIfDetailsToStatus() error {
	start := time.Now()
	req := &interfaces.SwInterfaceDump{}
	reqCtx := plugin.vppCh.SendMultiRequest(req)

	for {
		msg := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			plugin.Log.Error(err)
			return err
		}

		_, _, found := plugin.swIfIndexes.LookupName(msg.SwIfIndex)
		if !found {
			plugin.Log.Debugf("Unregistered interface %v with ID %v found on vpp",
				string(bytes.Trim(msg.InterfaceName, "\x00")), msg.SwIfIndex)
			// Do not register unknown interface here, cuz it may cause inconsistencies in the ifplugin.
			// All new interfaces should be registered during configuration
			continue
		}

		// Propagate interface state information to notification channel.
		plugin.notifChan <- msg
	}

	// SwInterfaceSetFlags time
	if plugin.Stopwatch != nil {
		timeLog := measure.GetTimeLog(interfaces.SwInterfaceSetFlags{}, plugin.Stopwatch)
		timeLog.LogTimeEntry(time.Since(start))
	}

	return nil
}

// ConfigureVPPInterface reacts to a new northbound VPP interface config by creating and configuring
// the interface in the VPP network stack through the VPP binary API.
func (plugin *InterfaceConfigurator) ConfigureVPPInterface(iface *intf.Interfaces_Interface) (err error) {
	plugin.Log.Infof("Configuring new interface %v", iface.Name)

	var ifIdx uint32

	switch iface.Type {
	case intf.InterfaceType_TAP_INTERFACE:
		ifIdx, err = vppcalls.AddTapInterface(iface.Name, iface.Tap, plugin.vppCh, plugin.Stopwatch)
	case intf.InterfaceType_MEMORY_INTERFACE:
		var id uint32 // Memif socket id
		id, err = plugin.resolveMemifSocketFilename(iface.Memif)
		if err != nil {
			return err
		}
		ifIdx, err = vppcalls.AddMemifInterface(iface.Name, iface.Memif, id, plugin.vppCh, plugin.Stopwatch)
	case intf.InterfaceType_VXLAN_TUNNEL:
		ifIdx, err = vppcalls.AddVxlanTunnel(iface.Name, iface.Vxlan, iface.Vrf, plugin.vppCh, plugin.Stopwatch)
	case intf.InterfaceType_SOFTWARE_LOOPBACK:
		ifIdx, err = vppcalls.AddLoopbackInterface(iface.Name, plugin.vppCh, plugin.Stopwatch)
	case intf.InterfaceType_ETHERNET_CSMACD:
		var exists bool
		if ifIdx, _, exists = plugin.swIfIndexes.LookupIdx(iface.Name); !exists {
			plugin.Log.Warnf("It is not yet supported to add (whitelist) a new physical interface")
			return nil
		}
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		var pending bool
		if ifIdx, pending, err = plugin.afPacketConfigurator.ConfigureAfPacketInterface(iface); err != nil {
			return err
		}
		if pending {
			plugin.Log.Debugf("interface %+v cannot be created yet and will be configured later", iface)
			return nil
		}
	}
	if err != nil {
		plugin.Log.Error(err)
		return err
	}

	var errs []error

	// rx mode
	if err := plugin.configRxModeForInterface(iface, ifIdx); err != nil {
		errs = append(errs, err)
	}

	// configure optional mac address
	if iface.PhysAddress != "" {
		if err := vppcalls.SetInterfaceMac(ifIdx, iface.PhysAddress, plugin.vppCh, plugin.Stopwatch); err != nil {
			errs = append(errs, err)
		}
	}

	// configure optional vrf
	if iface.Type != intf.InterfaceType_VXLAN_TUNNEL {
		if err := vppcalls.SetInterfaceVRF(ifIdx, iface.Vrf, plugin.Log, plugin.vppCh); err != nil {
			errs = append(errs, err)
		}
	}

	// configure DHCP client
	if iface.SetDhcpClient {
		if err := vppcalls.SetInterfaceAsDHCPClient(ifIdx, iface.Name, plugin.vppCh, plugin.Stopwatch); err != nil {
			errs = append(errs, err)
		} else {
			plugin.Log.Debugf("Interface %v set as DHCP client", iface.Name)
		}
	}

	// configure IP addresses/un-numbered
	IPAddrs, err := addrs.StrAddrsToStruct(iface.IpAddresses)
	if err != nil {
		return err
	}
	if err := plugin.configureIPAddresses(iface.Name, ifIdx, IPAddrs, iface.Unnumbered); err != nil {
		errs = append(errs, err)
	}

	// configure container IP address
	if iface.ContainerIpAddress != "" {
		if err := vppcalls.AddContainerIP(ifIdx, iface.ContainerIpAddress, plugin.vppCh, plugin.Stopwatch); err != nil {
			errs = append(errs, err)
		} else {
			plugin.Log.WithFields(logging.Fields{"IPaddr": iface.ContainerIpAddress, "ifIdx": ifIdx}).
				Debug("Container IP address added")
		}
	}

	// configure mtu. Prefer value in interface config, otherwise set default value if defined
	if iface.Type != intf.InterfaceType_VXLAN_TUNNEL {
		if iface.Mtu != 0 {
			if err := vppcalls.SetInterfaceMtu(ifIdx, iface.Mtu, plugin.vppCh, plugin.Stopwatch); err != nil {
				errs = append(errs, err)
			}
		} else if plugin.mtu != 0 {
			if err := vppcalls.SetInterfaceMtu(ifIdx, plugin.mtu, plugin.vppCh, plugin.Stopwatch); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// register name to idx mapping if it is not an af_packet interface type (it is registered in ConfigureAfPacketInterface if needed)
	if iface.Type != intf.InterfaceType_AF_PACKET_INTERFACE {
		plugin.swIfIndexes.RegisterName(iface.Name, ifIdx, iface)
	}

	l := plugin.Log.WithFields(logging.Fields{"ifName": iface.Name, "ifIdx": ifIdx})
	l.Debug("Configured interface")

	// set interface up if enabled
	// NOTE: needs to be called after RegisterName, otherwise interface up/down notification won't map to a valid interface
	if iface.Enabled {
		if err := vppcalls.InterfaceAdminUp(ifIdx, plugin.vppCh, plugin.Stopwatch); err != nil {
			l.Debugf("setting interface up failed: %v", err)
			return err
		}
	}

	// load interface state data for newly added interface (no way to filter by swIfIndex, need to dump all of them)
	plugin.PropagateIfDetailsToStatus()

	l.Info("Interface configuration done")

	// TODO: use some error aggregator
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
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
func (plugin *InterfaceConfigurator) configRxModeForInterface(iface *intf.Interfaces_Interface, ifIdx uint32) error {
	rxModeSettings := iface.RxModeSettings
	if rxModeSettings != nil {
		switch iface.Type {
		case intf.InterfaceType_ETHERNET_CSMACD:
			if rxModeSettings.RxMode == intf.RxModeType_POLLING {
				return plugin.configRxMode(iface, ifIdx, *rxModeSettings)
			}
		default:
			return plugin.configRxMode(iface, ifIdx, *rxModeSettings)
		}
	}
	return nil
}

/**
Call specific vpp API method for setting rx-mode
*/
func (plugin *InterfaceConfigurator) configRxMode(iface *intf.Interfaces_Interface, ifIdx uint32, rxModeSettings intf.Interfaces_Interface_RxModeSettings) error {
	err := vppcalls.SetRxMode(ifIdx, rxModeSettings, plugin.vppCh, plugin.Stopwatch)
	plugin.Log.WithFields(logging.Fields{"ifName": iface.Name, "rxMode": rxModeSettings.RxMode}).
		Debug("RX-mode configuration for ", iface.Type, ".")
	return err
}

func (plugin *InterfaceConfigurator) configureIPAddresses(ifName string, ifIdx uint32, addresses []*net.IPNet, unnumbered *intf.Interfaces_Interface_Unnumbered) error {
	if unnumbered != nil && unnumbered.IsUnnumbered {
		ifWithIP := unnumbered.InterfaceWithIP
		if ifWithIP == "" {
			return fmt.Errorf("unnubered interface %s has no interface with IP address set", ifName)
		}
		ifIdxIP, _, found := plugin.swIfIndexes.LookupIdx(ifWithIP)
		if !found {
			// cache not-configurable interface
			plugin.uIfaceCache[ifName] = ifWithIP
			plugin.Log.Debugf("unnubered interface %s requires IP address from non-existing %v, moved to cache", ifName, ifWithIP)
			return nil
		}
		// Set interface as un-numbered
		if err := vppcalls.SetUnnumberedIP(ifIdx, ifIdxIP, plugin.vppCh, plugin.Stopwatch); err != nil {
			return err
		} else {
			plugin.Log.WithFields(logging.Fields{"un-numberedIface": ifIdx, "ifIdxIP": ifIdxIP}).Debug("Interface set as un-numbered")
		}
		// just log
		if len(addresses) != 0 {
			plugin.Log.Warnf("Interface %v set as un-numbered contains IP address(es)", ifName, addresses)
		}
	}

	// configure optional ip address
	var wasErr error
	for _, address := range addresses {
		if err := vppcalls.AddInterfaceIP(ifIdx, address, plugin.vppCh, plugin.Stopwatch); err != nil {
			plugin.Log.Errorf("adding interface IP address failed: %v", err)
			wasErr = err
		}
	}

	// with ip address configured, the interface can be used as a source for un-numbered interfaces (if any)
	if err := plugin.resolveDependentUnnumberedInterfaces(ifName, ifIdx); err != nil {
		wasErr = err
	}
	return wasErr
}

func (plugin *InterfaceConfigurator) removeIPAddresses(ifIdx uint32, addresses []*net.IPNet, unnumbered *intf.Interfaces_Interface_Unnumbered) error {
	if unnumbered != nil && unnumbered.IsUnnumbered {
		// Set interface as un-numbered
		if err := vppcalls.UnsetUnnumberedIP(ifIdx, plugin.vppCh, plugin.Stopwatch); err != nil {
			return err
		}
	}

	// delete IP Addresses
	var wasErr error
	for _, addr := range addresses {
		err := vppcalls.DelInterfaceIP(ifIdx, addr, plugin.vppCh, plugin.Stopwatch)
		if err != nil {
			plugin.Log.Errorf("deleting IP address failed: %v", err)
			wasErr = err
		} else {
			plugin.Log.Debug("deleted IP addr %v", addr)
		}
	}

	return wasErr
}

// Iterate over all un-numbered interfaces in cache (which could not be configured before) and find all interfaces
// dependent on the provided one
func (plugin *InterfaceConfigurator) resolveDependentUnnumberedInterfaces(ifNameIP string, ifIdxIP uint32) error {
	plugin.Log.Debugf("Looking up unnumbered interfaces dependent on %v", ifNameIP)
	var wasErr error
	for uIface, ifWithIP := range plugin.uIfaceCache {
		if ifWithIP == ifNameIP {
			// find index of the dependent interface
			uIdx, _, found := plugin.swIfIndexes.LookupIdx(uIface)
			if !found {
				plugin.Log.Debugf("Unnumbered interface %v not found, removing from cache", uIface)
				delete(plugin.uIfaceCache, uIface)
				continue
			}
			if err := vppcalls.SetUnnumberedIP(uIdx, ifIdxIP, plugin.vppCh, plugin.Stopwatch); err != nil {
				plugin.Log.Errorf("setting unnumbered IP failed: %v", err)
				wasErr = err
			} else {
				plugin.Log.WithFields(logging.Fields{"un-numberedIface": uIdx, "ifIdxIP": ifIdxIP}).Debug("Interface set as un-numbered")
			}
			delete(plugin.uIfaceCache, uIface)
		}
	}
	return wasErr
}

// ModifyVPPInterface applies changes in the NB configuration of a VPP interface into the running VPP
// through the VPP binary API.
func (plugin *InterfaceConfigurator) ModifyVPPInterface(newConfig *intf.Interfaces_Interface,
	oldConfig *intf.Interfaces_Interface) error {
	plugin.Log.Infof("Modifying Interface %v", newConfig.Name)

	if newConfig == nil {
		return errors.New("newConfig is null")
	}
	if oldConfig == nil {
		return errors.New("oldConfig is null")
	}

	if plugin.afPacketConfigurator.IsPendingAfPacket(oldConfig) {
		return plugin.recreateVPPInterface(newConfig, oldConfig, 0)
	}

	// lookup index
	ifIdx, meta, found := plugin.swIfIndexes.LookupIdx(newConfig.Name)

	if !found {
		plugin.Log.WithFields(logging.Fields{"ifName": newConfig.Name}).Debug("Mapping for interface name not found.")
		return nil
	}

	if err := plugin.modifyVPPInterface(newConfig, oldConfig, ifIdx, meta.Type); err != nil {
		return err
	}

	plugin.Log.Infof("Interface %v modified", newConfig.Name)

	return nil
}

// ModifyVPPInterface applies changes in the NB configuration of a VPP interface into the running VPP
// through the VPP binary API.
func (plugin *InterfaceConfigurator) modifyVPPInterface(newConfig *intf.Interfaces_Interface, oldConfig *intf.Interfaces_Interface,
	ifIdx uint32, ifaceType intf.InterfaceType) (err error) {

	plugin.Log.WithFields(logging.Fields{"ifname": oldConfig.Name, "swIfIndex": ifIdx}).
		Debug("modifyVPPInterface begin")

	switch ifaceType {
	case intf.InterfaceType_TAP_INTERFACE:
		if !plugin.canTapBeModifWithoutDelete(newConfig.Tap, oldConfig.Tap) {
			err := plugin.recreateVPPInterface(newConfig, oldConfig, ifIdx)
			plugin.Log.WithFields(logging.Fields{"ifName": newConfig.Name, "ifIdx": ifIdx}).
				Debug("modifyVPPInterface end. ", err)
			return err
		}
	case intf.InterfaceType_MEMORY_INTERFACE:
		if !plugin.canMemifBeModifWithoutDelete(newConfig.Memif, oldConfig.Memif) {
			err := plugin.recreateVPPInterface(newConfig, oldConfig, ifIdx)
			plugin.Log.WithFields(logging.Fields{"ifName": newConfig.Name, "ifIdx": ifIdx}).
				Debug("modifyVPPInterface end. ", err)
			return err
		}
	case intf.InterfaceType_VXLAN_TUNNEL:
		if !plugin.canVxlanBeModifWithoutDelete(newConfig.Vxlan, oldConfig.Vxlan) ||
			oldConfig.Vrf != newConfig.Vrf {
			err := plugin.recreateVPPInterface(newConfig, oldConfig, ifIdx)
			plugin.Log.WithFields(logging.Fields{"ifName": newConfig.Name, "ifIdx": ifIdx}).
				Debug("modifyVPPInterface end. ", err)
			return err
		}
	case intf.InterfaceType_SOFTWARE_LOOPBACK:
	case intf.InterfaceType_ETHERNET_CSMACD:
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		recreate, err := plugin.afPacketConfigurator.ModifyAfPacketInterface(newConfig, oldConfig)
		if err != nil || recreate {
			if err == nil {
				err = plugin.recreateVPPInterface(newConfig, oldConfig, ifIdx)
			}
			plugin.Log.WithFields(logging.Fields{"ifName": newConfig.Name, "ifIdx": ifIdx}).
				Debug("modifyVPPInterface end. ", err)
			return err
		}
	}

	var wasError error
	// rx mode
	wasError = plugin.modifyRxModeForInterfaces(oldConfig, newConfig, ifIdx)

	// admin status
	if newConfig.Enabled != oldConfig.Enabled {
		if newConfig.Enabled {
			err = vppcalls.InterfaceAdminUp(ifIdx, plugin.vppCh, nil)
		} else {
			err = vppcalls.InterfaceAdminDown(ifIdx, plugin.vppCh, nil)
		}
		if nil != err {
			wasError = err
		}
	}

	// configure new mac address if set (and only if it was changed)
	if newConfig.PhysAddress != "" && newConfig.PhysAddress != oldConfig.PhysAddress {
		if err := vppcalls.SetInterfaceMac(ifIdx, newConfig.PhysAddress, plugin.vppCh, plugin.Stopwatch); err != nil {
			plugin.Log.Errorf("setting interface MAC address failed: %v", err)
			wasError = err
		}
	}

	// reconfigure DHCP
	if oldConfig.SetDhcpClient != newConfig.SetDhcpClient {
		if newConfig.SetDhcpClient {
			if err := vppcalls.SetInterfaceAsDHCPClient(ifIdx, newConfig.Name, plugin.vppCh, plugin.Stopwatch); err != nil {
				plugin.Log.Error(err)
				wasError = err
			} else {
				plugin.Log.Debugf("Interface %v set as DHCP client", newConfig.Name)
			}
		} else {
			if err := vppcalls.UnsetInterfaceAsDHCPClient(ifIdx, newConfig.Name, plugin.vppCh, plugin.Stopwatch); err != nil {
				plugin.Log.Error(err)
				wasError = err
			} else {
				// Remove from dhcp mapping
				plugin.dhcpIndices.UnregisterName(newConfig.Name)
				plugin.Log.Debugf("Interface %v unset as DHCP client", oldConfig.Name)
			}
		}
	}

	// ip address
	newAddrs, err := addrs.StrAddrsToStruct(newConfig.IpAddresses)
	if err != nil {
		return err
	}

	oldAddrs, err := addrs.StrAddrsToStruct(oldConfig.IpAddresses)
	if err != nil {
		return err
	}

	// configure VRF if it was changed
	if oldConfig.Vrf != newConfig.Vrf &&
		ifaceType != intf.InterfaceType_VXLAN_TUNNEL {
		plugin.Log.Debugf("VRF changed: %v -> %v", oldConfig.Vrf, newConfig.Vrf)

		// interface must not have IP when setting VRF
		if err := plugin.removeIPAddresses(ifIdx, oldAddrs, newConfig.Unnumbered); err != nil {
			plugin.Log.Error(err)
			wasError = err
		}

		if err := vppcalls.SetInterfaceVRF(ifIdx, newConfig.Vrf, plugin.Log, plugin.vppCh); err != nil {
			plugin.Log.Error(err)
			wasError = err
		}

		if err = plugin.configureIPAddresses(newConfig.Name, ifIdx, newAddrs, newConfig.Unnumbered); err != nil {
			plugin.Log.Error(err)
			wasError = err
		}

	} else {
		// if VRF is not changed, try to add/del only differences
		del, add := addrs.DiffAddr(newAddrs, oldAddrs)

		plugin.Log.Debug("del ip addrs: ", del)
		plugin.Log.Debug("add ip addrs: ", add)

		if err := plugin.removeIPAddresses(ifIdx, del, oldConfig.Unnumbered); err != nil {
			plugin.Log.Error(err)
			wasError = err
		}

		if err := plugin.configureIPAddresses(newConfig.Name, ifIdx, add, newConfig.Unnumbered); err != nil {
			plugin.Log.Error(err)
			wasError = err
		}
	}

	// container ip address
	if newConfig.ContainerIpAddress != oldConfig.ContainerIpAddress {
		plugin.Log.WithFields(logging.Fields{"ifIdx": ifIdx, "ip_new": newConfig.ContainerIpAddress, "ip_old": oldConfig.ContainerIpAddress}).
			Debug("Container IP address modification.")
		if err := vppcalls.AddContainerIP(ifIdx, newConfig.ContainerIpAddress, plugin.vppCh, plugin.Stopwatch); err != nil {
			plugin.Log.WithFields(logging.Fields{"newIP": newConfig.ContainerIpAddress, "oldIP": oldConfig.ContainerIpAddress, "ifIdx": ifIdx}).
				Errorf("adding container IP failed: %v", err)
			wasError = err
		}
	}

	// Set MTU if changed in interface config
	if newConfig.Mtu != 0 && newConfig.Mtu != oldConfig.Mtu {
		if err := vppcalls.SetInterfaceMtu(ifIdx, newConfig.Mtu, plugin.vppCh, plugin.Stopwatch); err != nil {
			wasError = err
		}
	} else if newConfig.Mtu == 0 && plugin.mtu != 0 {
		if err := vppcalls.SetInterfaceMtu(ifIdx, plugin.mtu, plugin.vppCh, plugin.Stopwatch); err != nil {
			wasError = err
		}
	}

	plugin.Log.WithFields(logging.Fields{"ifName": newConfig.Name, "ifIdx": ifIdx}).Info("Modified interface")

	return wasError
}

/**
Modify rx-mode on specified VPP interface
*/
func (plugin *InterfaceConfigurator) modifyRxModeForInterfaces(oldIntf *intf.Interfaces_Interface, newIntf *intf.Interfaces_Interface,
	ifIdx uint32) error {
	oldRxSettings := oldIntf.RxModeSettings
	newRxSettings := newIntf.RxModeSettings
	if oldRxSettings != newRxSettings {
		var oldRxMode intf.RxModeType
		if oldRxSettings != nil {
			oldRxMode = oldRxSettings.RxMode
		}
		if newRxSettings != nil {
			switch newIntf.Type {
			case intf.InterfaceType_ETHERNET_CSMACD:
				if newRxSettings.RxMode == intf.RxModeType_POLLING {
					return plugin.modifyRxMode(ifIdx, newIntf, oldRxMode, *newRxSettings)
				}
				plugin.Log.WithFields(logging.Fields{"rx-mode": newRxSettings.RxMode}).
					Warn("Attempt to set unsupported rx-mode on Ethernet interface.")
			default:
				return plugin.modifyRxMode(ifIdx, newIntf, oldRxMode, *newRxSettings)
			}
		} else {
			// reset rx-mode to default value
			newRxSettings = &intf.Interfaces_Interface_RxModeSettings{}
			switch newIntf.Type {
			case intf.InterfaceType_ETHERNET_CSMACD:
				newRxSettings.RxMode = intf.RxModeType_POLLING
			case intf.InterfaceType_AF_PACKET_INTERFACE:
				newRxSettings.RxMode = intf.RxModeType_INTERRUPT
			default:
				newRxSettings.RxMode = intf.RxModeType_DEFAULT
			}
			newIntf.RxModeSettings = newRxSettings
			return plugin.modifyRxMode(ifIdx, newIntf, oldRxMode, *newRxSettings)
		}
	}
	return nil
}

/**
Direct call of vpp api to change rx-mode of specified interface
*/
func (plugin *InterfaceConfigurator) modifyRxMode(ifIdx uint32, newIntf *intf.Interfaces_Interface,
	oldRxMode intf.RxModeType, newRxMode intf.Interfaces_Interface_RxModeSettings) error {
	err := vppcalls.SetRxMode(ifIdx, *newIntf.RxModeSettings, plugin.vppCh, plugin.Stopwatch)
	plugin.Log.WithFields(
		logging.Fields{"ifName": newIntf.Name, "rxMode old": oldRxMode, "rxMode new": newRxMode.RxMode}).
		Debug("RX-mode modification for ", newIntf.Type, ".")
	return err
}

// recreateVPPInterface removes and creates an interface from scratch.
func (plugin *InterfaceConfigurator) recreateVPPInterface(newConfig *intf.Interfaces_Interface,
	oldConfig *intf.Interfaces_Interface, ifIdx uint32) (wasError error) {
	var err error

	if oldConfig.Type == intf.InterfaceType_AF_PACKET_INTERFACE {
		err = plugin.afPacketConfigurator.DeleteAfPacketInterface(oldConfig, ifIdx)
	} else {
		err = plugin.deleteVPPInterface(oldConfig, ifIdx)
	}
	if err != nil {
		return err
	}
	return plugin.ConfigureVPPInterface(newConfig)
}

// DeleteVPPInterface reacts to a removed NB configuration of a VPP interface.
// It results in the interface being removed from VPP.
func (plugin *InterfaceConfigurator) DeleteVPPInterface(iface *intf.Interfaces_Interface) (wasError error) {
	plugin.Log.Infof("Removing interface %v", iface.Name)

	if plugin.afPacketConfigurator.IsPendingAfPacket(iface) {
		ifIdx, _, found := plugin.afPacketConfigurator.SwIfIndexes.LookupIdx(iface.Name)
		if !found {
			return fmt.Errorf("cannot remove af packet interface %v, index not available from mapping", iface.Name)
		}
		return plugin.afPacketConfigurator.DeleteAfPacketInterface(iface, ifIdx)
	}

	// unregister name to init mapping (following triggers notifications for all subscribers, skip physical interfaces)
	if iface.Type != intf.InterfaceType_ETHERNET_CSMACD {
		ifIdx, prev, found := plugin.swIfIndexes.UnregisterName(iface.Name)
		if !found {
			plugin.Log.WithField("ifname", iface.Name).Debug("Unable to find index for interface to be deleted.")
			return nil
		}

		// delete from unnumbered map (if the interface is present)
		delete(plugin.uIfaceCache, iface.Name)

		if err := plugin.deleteVPPInterface(prev, ifIdx); err != nil {
			return err
		}
	} else {
		// Find index of the Physical interface and un-configure it
		ifIdx, prev, found := plugin.swIfIndexes.LookupIdx(iface.Name)
		if !found {
			plugin.Log.WithField("ifname", iface.Name).Debug("Unable to find index for interface to be deleted.")
			return nil
		}
		if err := plugin.deleteVPPInterface(prev, ifIdx); err != nil {
			return err
		}
	}

	plugin.Log.Infof("Interface %v removed", iface.Name)

	return wasError
}

func (plugin *InterfaceConfigurator) deleteVPPInterface(oldConfig *intf.Interfaces_Interface, ifIdx uint32) (wasError error) {
	plugin.Log.WithFields(logging.Fields{"ifname": oldConfig.Name, "swIfIndex": ifIdx}).
		Debug("deleteVPPInterface begin")

	// let's try to do following even if previously error occurred
	if err := vppcalls.InterfaceAdminDown(ifIdx, plugin.vppCh, plugin.Stopwatch); err != nil {
		plugin.Log.Error(err)
		wasError = err
	}

	// Remove DHCP if it was set
	if oldConfig.SetDhcpClient {
		if err := vppcalls.UnsetInterfaceAsDHCPClient(ifIdx, oldConfig.Name, plugin.vppCh, plugin.Stopwatch); err != nil {
			plugin.Log.Error(err)
			wasError = err
		}
		// Remove from dhcp mapping
		plugin.dhcpIndices.UnregisterName(oldConfig.Name)
		plugin.Log.Debugf("Interface %v unset as DHCP client", oldConfig.Name)
	}

	// let's try to do following even if previously error occurred
	if oldConfig.ContainerIpAddress != "" {
		if err := vppcalls.DelContainerIP(ifIdx, oldConfig.ContainerIpAddress, plugin.vppCh, plugin.Stopwatch); err != nil {
			plugin.Log.Error(err)
			wasError = err
		} else {
			plugin.Log.WithFields(logging.Fields{"IPaddr": oldConfig.ContainerIpAddress, "ifIdx": ifIdx}).
				Debug("Container IP address deleted")
		}
	}

	for i, oldIP := range oldConfig.IpAddresses {
		if strings.HasPrefix(oldIP, "fe80") {
			// TODO: skip link local addresses (possible workaround for af_packet)
			oldConfig.IpAddresses = append(oldConfig.IpAddresses[:i], oldConfig.IpAddresses[i+1:]...)
		}
	}
	oldAddrs, err := addrs.StrAddrsToStruct(oldConfig.IpAddresses)
	if err != nil {
		plugin.Log.WithFields(logging.Fields{"ifname": oldConfig.Name, "swIfIndex": ifIdx}).
			Debug("deleteVPPInterface end ", err)
		return err
	}
	for _, oldAddr := range oldAddrs {
		if err := vppcalls.DelInterfaceIP(ifIdx, oldAddr, plugin.vppCh, plugin.Stopwatch); err != nil {
			plugin.Log.Errorf("deleting interface IP address failed: %v", err)
			wasError = err
		}
	}

	plugin.Log.Info("IP addrs removed")

	// let's try to do following even if previously error occurred
	switch oldConfig.Type {
	case intf.InterfaceType_TAP_INTERFACE:
		err = vppcalls.DeleteTapInterface(oldConfig.Name, ifIdx, oldConfig.Tap.Version, plugin.vppCh, plugin.Stopwatch)
	case intf.InterfaceType_MEMORY_INTERFACE:
		err = vppcalls.DeleteMemifInterface(oldConfig.Name, ifIdx, plugin.vppCh, plugin.Stopwatch)
	case intf.InterfaceType_VXLAN_TUNNEL:
		err = vppcalls.DeleteVxlanTunnel(oldConfig.Name, ifIdx, oldConfig.GetVxlan(), plugin.vppCh, plugin.Stopwatch)
	case intf.InterfaceType_SOFTWARE_LOOPBACK:
		err = vppcalls.DeleteLoopbackInterface(oldConfig.Name, ifIdx, plugin.vppCh, plugin.Stopwatch)
	case intf.InterfaceType_ETHERNET_CSMACD:
		plugin.Log.Debugf("Interface removal skipped: cannot remove (blacklist) physical interface") // Not an error
		return nil
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		err = plugin.afPacketConfigurator.DeleteAfPacketInterface(oldConfig, ifIdx)
	}
	if err != nil {
		wasError = err
	}

	plugin.Log.WithFields(logging.Fields{"ifname": oldConfig.Name, "swIfIndex": ifIdx}).
		Debug("deleteVPPInterface end ", err)

	return wasError
}

// ResolveCreatedLinuxInterface reacts to a newly created Linux interface.
func (plugin *InterfaceConfigurator) ResolveCreatedLinuxInterface(interfaceName, hostIfName string, interfaceIndex uint32) {
	plugin.Log.WithFields(logging.Fields{"ifName": interfaceName, "hostIfName": hostIfName, "ifIdx": interfaceIndex}).Info("New Linux interface was created")

	pendingAfpacket := plugin.afPacketConfigurator.ResolveCreatedLinuxInterface(interfaceName, hostIfName, interfaceIndex)
	if pendingAfpacket != nil {
		// there is a pending afpacket that can be now configured
		if err := plugin.ConfigureVPPInterface(pendingAfpacket); err != nil {
			plugin.Log.Error(err)
		}
	}
}

// ResolveDeletedLinuxInterface reacts to a removed Linux interface.
func (plugin *InterfaceConfigurator) ResolveDeletedLinuxInterface(interfaceName, hostIfName string, ifIdx uint32) {
	plugin.Log.WithFields(logging.Fields{"ifName": interfaceName, "hostIfName": hostIfName}).Info("Linux interface was deleted")

	plugin.afPacketConfigurator.ResolveDeletedLinuxInterface(interfaceName, hostIfName, ifIdx)
}

// returns memif socket filename ID. Registers it if does not exists yet
func (plugin *InterfaceConfigurator) resolveMemifSocketFilename(memifIf *intf.Interfaces_Interface_Memif) (uint32, error) {
	if memifIf.SocketFilename == "" {
		return 0, fmt.Errorf("memif configuration does not contain socket file name")
	}
	registeredID, ok := plugin.memifScCache[memifIf.SocketFilename]
	if !ok {
		// Register new socket. ID is generated (default filename ID is 0, first is ID 1, second ID 2, etc)
		registeredID = uint32(len(plugin.memifScCache))
		err := vppcalls.RegisterMemifSocketFilename([]byte(memifIf.SocketFilename), registeredID, plugin.vppCh, plugin.Stopwatch)
		if err != nil {
			return 0, fmt.Errorf("error registering socket file name %s (ID %d): %v", memifIf.SocketFilename, registeredID, err)
		}
		plugin.memifScCache[memifIf.SocketFilename] = registeredID
		plugin.Log.Debugf("Memif socket filename %s registered under ID %d", memifIf.SocketFilename, registeredID)
	}
	return registeredID, nil
}

func (plugin *InterfaceConfigurator) canMemifBeModifWithoutDelete(newConfig *intf.Interfaces_Interface_Memif, oldConfig *intf.Interfaces_Interface_Memif) bool {
	if newConfig == nil || oldConfig == nil {
		return true
	}

	if *newConfig != *oldConfig {
		plugin.Log.Warnf("Difference between new & old config causing recreation of memif, old: '%+v' new: '%+v'", oldConfig, newConfig)
		return false
	}

	return true
}

func (plugin *InterfaceConfigurator) canVxlanBeModifWithoutDelete(newConfig *intf.Interfaces_Interface_Vxlan, oldConfig *intf.Interfaces_Interface_Vxlan) bool {
	if newConfig == nil || oldConfig == nil {
		return true
	}
	if *newConfig != *oldConfig {
		return false
	}

	return true
}

func (plugin *InterfaceConfigurator) canTapBeModifWithoutDelete(newConfig *intf.Interfaces_Interface_Tap, oldConfig *intf.Interfaces_Interface_Tap) bool {
	if newConfig == nil || oldConfig == nil {
		return true
	}
	if *newConfig != *oldConfig {
		return false
	}

	return true
}

// watch and process DHCP notifications. DHCP configuration is registered to dhcp mapping for every interface
func (plugin *InterfaceConfigurator) watchDHCPNotifications() {
	plugin.Log.Debug("Started watcher on DHCP notifications")

	for {
		select {
		case notification := <-plugin.dhcpChan:
			switch dhcpNotif := notification.(type) {
			case *dhcp.DhcpComplEvent:
				var ipAddr, rIPAddr net.IP = dhcpNotif.HostAddress, dhcpNotif.RouterAddress
				var hwAddr net.HardwareAddr = dhcpNotif.HostMac
				var ipStr, rIPStr string

				name := string(bytes.Trim(dhcpNotif.Hostname, "\x00"))

				if dhcpNotif.IsIpv6 == 1 {
					ipStr = ipAddr.To16().String()
					rIPStr = rIPAddr.To16().String()
				} else {
					ipStr = ipAddr[:4].To4().String()
					rIPStr = rIPAddr[:4].To4().String()
				}

				plugin.Log.Debugf("DHCP assigned %v to interface %q (router address %v)", ipStr, name, rIPStr)

				ifIdx, _, found := plugin.swIfIndexes.LookupIdx(name)
				if !found {
					plugin.Log.Warnf("Expected interface %v not found in the mapping", name)
					continue
				}

				// Register DHCP config
				plugin.dhcpIndices.RegisterName(name, ifIdx, &ifaceidx.DHCPSettings{
					IfName: name,
					IsIPv6: func(isIPv6 uint8) bool {
						if isIPv6 == 1 {
							return true
						}
						return false
					}(dhcpNotif.IsIpv6),
					IPAddress:     ipStr,
					Mask:          uint32(dhcpNotif.MaskWidth),
					PhysAddress:   hwAddr.String(),
					RouterAddress: rIPStr,
				})

				plugin.Log.Debugf("Registered dhcp metadata for interface %v", name)
			}
		}
	}
}
