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

//go:generate protoc --proto_path=model/interfaces --gogo_out=model/interfaces model/interfaces/interfaces.proto
//go:generate protoc --proto_path=model/bfd --gogo_out=model/bfd model/bfd/bfd.proto

//go:generate binapi-generator --input-file=/usr/share/vpp/api/af_packet.api.json --output-dir=bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/interface.api.json --output-dir=bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/ip.api.json --output-dir=bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/memif.api.json --output-dir=bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/tap.api.json --output-dir=bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/tapv2.api.json --output-dir=bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/vpe.api.json --output-dir=bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/vxlan.api.json --output-dir=bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/stats.api.json --output-dir=bin_api

// Package ifplugin implements the Interface plugin that handles management
// of VPP interfaces.
package ifplugin

import (
	"bytes"
	"errors"

	"time"

	"fmt"
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core/bin_api/vpe"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/memif"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/tap"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/vxlan"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

const dummyMode = -1

// InterfaceConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/interfaces/interfaces.proto"
// and stored in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1interface".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type InterfaceConfigurator struct {
	Log logging.Logger

	GoVppmux     govppmux.API
	ServiceLabel servicelabel.ReaderAPI
	Linux        interface{} //just flag if nil

	Stopwatch *measure.Stopwatch // timer used to measure and store time

	swIfIndexes ifaceidx.SwIfIndexRW

	uIfaceCache map[string]string // cache for not-configurable unnumbered interfaces. map[unumbered-iface-name]required-iface

	mtu uint32 // default MTU value can be read from config

	afPacketConfigurator *AFPacketConfigurator

	vppCh *govppapi.Channel

	notifChan chan govppapi.Message // to publish SwInterfaceDetails to interface_state.go

	resyncDoneOnce bool
}

// Init members (channels...) and start go routines
func (plugin *InterfaceConfigurator) Init(swIfIndexes ifaceidx.SwIfIndexRW, mtu uint32, notifChan chan govppapi.Message) (err error) {
	plugin.Log.Debug("Initializing Interface configurator")
	plugin.swIfIndexes = swIfIndexes
	plugin.notifChan = notifChan
	plugin.mtu = mtu

	plugin.vppCh, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}
	err = vppcalls.CheckMsgCompatibilityForInterface(plugin.Log, plugin.vppCh)
	if err != nil {
		return err
	}

	plugin.afPacketConfigurator = &AFPacketConfigurator{Logger: plugin.Log, Linux: plugin.Linux, SwIfIndexes: plugin.swIfIndexes, Stopwatch: plugin.Stopwatch}
	plugin.afPacketConfigurator.Init(plugin.vppCh)

	plugin.uIfaceCache = make(map[string]string)

	return nil
}

// Close GOVPP channel
func (plugin *InterfaceConfigurator) Close() error {
	return safeclose.Close(plugin.vppCh)
}

// LookupVPPInterfaces looks up all VPP interfaces and saves their name-to-index mapping and state information.
func (plugin *InterfaceConfigurator) LookupVPPInterfaces() error {
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

		// store the name-to-index mapping if it does not exist yet
		_, _, found := plugin.swIfIndexes.LookupName(msg.SwIfIndex)
		if !found {
			ifName := string(bytes.Trim(msg.InterfaceName, "\x00"))
			plugin.Log.WithFields(logging.Fields{"ifName": ifName, "swIfIndex": msg.SwIfIndex}).
				Debug("Register VPP interface name mapping.")

			plugin.swIfIndexes.RegisterName(ifName, msg.SwIfIndex, nil)
		}

		// propagate interface state information
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
		ifIdx, err = vppcalls.AddTapInterface(iface.Tap, plugin.vppCh, measure.GetTimeLog(tap.TapConnect{}, plugin.Stopwatch))
	case intf.InterfaceType_MEMORY_INTERFACE:
		ifIdx, err = vppcalls.AddMemifInterface(iface.Memif, plugin.vppCh, measure.GetTimeLog(memif.MemifCreate{}, plugin.Stopwatch))
	case intf.InterfaceType_VXLAN_TUNNEL:
		ifIdx, err = vppcalls.AddVxlanTunnel(iface.Vxlan, iface.Vrf, plugin.vppCh, measure.GetTimeLog(vxlan.VxlanAddDelTunnelReply{}, plugin.Stopwatch))
	case intf.InterfaceType_SOFTWARE_LOOPBACK:
		ifIdx, err = vppcalls.AddLoopbackInterface(plugin.vppCh, measure.GetTimeLog(vpe.CreateLoopback{}, plugin.Stopwatch))
	case intf.InterfaceType_ETHERNET_CSMACD:
		var exists bool
		if ifIdx, _, exists = plugin.swIfIndexes.LookupIdx(iface.Name); !exists {
			return errors.New("it is not yet supported to add (whitelist) a new physical interface")
		}
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		var pending bool
		if ifIdx, pending, err = plugin.afPacketConfigurator.ConfigureAfPacketInterface(iface); err != nil {
			return err
		} else if pending {
			plugin.Log.Debugf("interface %+v cannot be created yet and will be configured later", iface)
			return nil
		}
	}
	if err != nil {
		return err
	}

	var errs []error

	//rx mode
	if err := plugin.configRxModeForInterface(iface, ifIdx); err != nil {
		errs = append(errs, err)
	}

	// configure optional mac address
	if iface.PhysAddress != "" {
		err := vppcalls.SetInterfaceMac(ifIdx, iface.PhysAddress, plugin.Log, plugin.vppCh,
			measure.GetTimeLog(interfaces.SwInterfaceSetMacAddress{}, plugin.Stopwatch))
		if err != nil {
			errs = append(errs, err)
		}
	}

	// configure optional vrf
	if iface.Type != intf.InterfaceType_VXLAN_TUNNEL {
		if err := vppcalls.SetInterfaceVRF(ifIdx, iface.Vrf, plugin.Log, plugin.vppCh); err != nil {
			errs = append(errs, err)
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
		if err := plugin.addContainerIPAddress(iface, ifIdx); err != nil {
			errs = append(errs, err)
		}
	}

	// configure mtu. Prefer value in interface config, otherwise set default value if defined
	if iface.Type != intf.InterfaceType_VXLAN_TUNNEL {
		if iface.Mtu != 0 {
			err = vppcalls.SetInterfaceMtu(ifIdx, iface.Mtu, plugin.Log, plugin.vppCh,
				measure.GetTimeLog(interfaces.SwInterfaceSetMtu{}, plugin.Stopwatch))
			if err != nil {
				errs = append(errs, err)
			}
		} else if plugin.mtu != 0 {
			err = vppcalls.SetInterfaceMtu(ifIdx, plugin.mtu, plugin.Log, plugin.vppCh,
				measure.GetTimeLog(interfaces.SwInterfaceSetMtu{}, plugin.Stopwatch))
			if err != nil {
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
		err := vppcalls.InterfaceAdminUp(ifIdx, plugin.vppCh,
			measure.GetTimeLog(interfaces.SwInterfaceSetFlags{}, plugin.Stopwatch))
		if nil != err {
			l.Debugf("setting interface up failed: %v", err)
			return err
		}
	}

	// load interface state data for newly added interface (no way to filter by swIfIndex, need to dump all of them)
	plugin.LookupVPPInterfaces()

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
Call concrete vpp API method for setting rx-mode
*/
func (plugin *InterfaceConfigurator) configRxMode(iface *intf.Interfaces_Interface, ifIdx uint32, rxModeSettings intf.Interfaces_Interface_RxModeSettings) error {
	err := vppcalls.SetRxMode(ifIdx, rxModeSettings, plugin.Log, plugin.vppCh,
		measure.GetTimeLog(interfaces.SwInterfaceSetRxMode{}, plugin.Stopwatch))
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
		err := vppcalls.SetUnnumberedIP(ifIdx, ifIdxIP, plugin.Log, plugin.vppCh, measure.GetTimeLog(interfaces.SwInterfaceSetUnnumbered{}, plugin.Stopwatch))
		if err != nil {
			return err
		}
		// just log
		if len(addresses) != 0 {
			plugin.Log.Warnf("Interface %v set as un-numbered contains IP address(es)", ifName, addresses)
		}
	}

	// configure optional ip address
	var wasErr error
	for _, address := range addresses {
		err := vppcalls.AddInterfaceIP(ifIdx, address, plugin.Log, plugin.vppCh,
			measure.GetTimeLog(interfaces.SwInterfaceAddDelAddress{}, plugin.Stopwatch))
		if nil != err {
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
		err := vppcalls.UnsetUnnumberedIP(ifIdx, plugin.Log, plugin.vppCh, measure.GetTimeLog(interfaces.SwInterfaceSetUnnumbered{}, plugin.Stopwatch))
		if err != nil {
			return err
		}
	}

	// delete IP Addresses
	var wasErr error
	for _, addr := range addresses {
		err := vppcalls.DelInterfaceIP(ifIdx, addr, plugin.Log, plugin.vppCh,
			measure.GetTimeLog(interfaces.SwInterfaceAddDelAddress{}, plugin.Stopwatch))
		plugin.Log.Debug("del ip addr ", ifIdx, " ", addr, " ", err)
		if nil != err {
			wasErr = err
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
			err := vppcalls.SetUnnumberedIP(uIdx, ifIdxIP, plugin.Log, plugin.vppCh, measure.GetTimeLog(interfaces.SwInterfaceSetUnnumbered{}, plugin.Stopwatch))
			delete(plugin.uIfaceCache, uIface)
			if err != nil {
				wasErr = err
			}
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
	//rx mode
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
		err := vppcalls.SetInterfaceMac(ifIdx, newConfig.PhysAddress, plugin.Log, plugin.vppCh, nil)
		if err != nil {
			wasError = err
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
		err = plugin.removeIPAddresses(ifIdx, oldAddrs, newConfig.Unnumbered)
		if err != nil {
			wasError = err
		}

		err := vppcalls.SetInterfaceVRF(ifIdx, newConfig.Vrf, plugin.Log, plugin.vppCh)
		if err != nil {
			wasError = err
		}

		err = plugin.configureIPAddresses(newConfig.Name, ifIdx, newAddrs, newConfig.Unnumbered)
		if err != nil {
			wasError = err
		}

	} else {
		// if VRF is not changed, try to add/del only differences
		del, add := addrs.DiffAddr(newAddrs, oldAddrs)

		plugin.Log.Debug("del ip addrs: ", del)
		plugin.Log.Debug("add ip addrs: ", add)

		err = plugin.removeIPAddresses(ifIdx, del, oldConfig.Unnumbered)
		if err != nil {
			wasError = err
		}

		err = plugin.configureIPAddresses(newConfig.Name, ifIdx, add, newConfig.Unnumbered)
		if err != nil {
			wasError = err
		}
	}

	//container ip address
	err = plugin.modifyContainerIPAddress(newConfig, oldConfig, ifIdx)
	if err != nil {
		plugin.Log.WithFields(logging.Fields{"new ip": newConfig.ContainerIpAddress, "old ip": oldConfig.ContainerIpAddress,
			"ifIdx": ifIdx}).Debug("Container IP modification problem ", err)
	}

	// Set MTU if changed in interface config
	if newConfig.Mtu != 0 && newConfig.Mtu != oldConfig.Mtu {
		err := vppcalls.SetInterfaceMtu(ifIdx, newConfig.Mtu, plugin.Log, plugin.vppCh, nil)
		if err != nil {
			wasError = err
		}
	} else if newConfig.Mtu == 0 && plugin.mtu != 0 {
		err := vppcalls.SetInterfaceMtu(ifIdx, plugin.mtu, plugin.Log, plugin.vppCh, nil)
		if err != nil {
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
		if oldRxSettings == nil {
			oldRxMode = dummyMode
		} else {
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
			//reset rx-mode to default value
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
	err := vppcalls.SetRxMode(ifIdx, *newIntf.RxModeSettings, plugin.Log, plugin.vppCh,
		measure.GetTimeLog(interfaces.SwInterfaceSetRxMode{}, plugin.Stopwatch))
	plugin.Log.WithFields(
		logging.Fields{"ifName": newIntf.Name, "rxMode old": oldRxMode, "rxMode new": newRxMode.RxMode}).
		Debug("RX-mode modification for ", newIntf.Type, ".")
	return err
}

func (plugin *InterfaceConfigurator) modifyContainerIPAddress(newConfig *intf.Interfaces_Interface,
	oldConfig *intf.Interfaces_Interface, ifIdx uint32) error {
	if newConfig.ContainerIpAddress != oldConfig.ContainerIpAddress {
		plugin.Log.WithFields(logging.Fields{"ifIdx": ifIdx, "ip_new": newConfig.ContainerIpAddress,
			"ip_old": oldConfig.ContainerIpAddress}).
			Debug("Container IP address modification.")
		return plugin.addContainerIPAddress(newConfig, ifIdx)
	}
	return nil
}

// recreateVPPInterface removes and creates an interface from scratch.
func (plugin *InterfaceConfigurator) recreateVPPInterface(newConfig *intf.Interfaces_Interface,
	oldConfig *intf.Interfaces_Interface, ifIdx uint32) (wasError error) {
	var err error

	if oldConfig.Type == intf.InterfaceType_AF_PACKET_INTERFACE {
		err = plugin.afPacketConfigurator.DeleteAfPacketInterface(oldConfig)
	} else {
		err = plugin.deleteVPPInterface(oldConfig, ifIdx)
	}
	if err == nil {
		err = plugin.ConfigureVPPInterface(newConfig)
	}
	return err
}

// DeleteVPPInterface reacts to a removed NB configuration of a VPP interface.
// It results in the interface being removed from VPP.
func (plugin *InterfaceConfigurator) DeleteVPPInterface(iface *intf.Interfaces_Interface) (wasError error) {
	plugin.Log.Infof("Removing interface %v", iface.Name)

	if plugin.afPacketConfigurator.IsPendingAfPacket(iface) {
		return plugin.afPacketConfigurator.DeleteAfPacketInterface(iface)
	}

	// unregister name to init mapping (following triggers notifications for all subscribers)
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

	plugin.Log.Infof("Interface %v removed", iface.Name)

	return wasError
}

func (plugin *InterfaceConfigurator) deleteVPPInterface(oldConfig *intf.Interfaces_Interface, ifIdx uint32) (
	wasError error) {
	plugin.Log.WithFields(logging.Fields{"ifname": oldConfig.Name, "swIfIndex": ifIdx}).
		Debug("deleteVPPInterface begin")

	// let's try to do following even if previously error occurred
	err := vppcalls.InterfaceAdminDown(ifIdx, plugin.vppCh, measure.GetTimeLog(interfaces.SwInterfaceSetFlags{}, plugin.Stopwatch))
	if nil != err {
		wasError = err
	}

	// let's try to do following even if previously error occurred
	plugin.deleteContainerIPAddress(oldConfig, ifIdx)

	// let's try to do following even if previously error occurred
	oldAddrs, err := addrs.StrAddrsToStruct(oldConfig.IpAddresses)
	if err != nil {
		plugin.Log.WithFields(logging.Fields{"ifname": oldConfig.Name, "swIfIndex": ifIdx}).
			Debug("deleteVPPInterface end ", err)
		return err
	}
	for _, oldAddr := range oldAddrs {
		plugin.Log.WithField("addr", oldAddr).Info("Ip removed")
		err := vppcalls.DelInterfaceIP(ifIdx, oldAddr, plugin.Log, plugin.vppCh,
			measure.GetTimeLog(interfaces.SwInterfaceAddDelAddressReply{}, plugin.Stopwatch))
		if nil != err {
			plugin.Log.Error(err)
			wasError = err
		}
	}

	plugin.Log.Info("Ip addrs removed")

	// let's try to do following even if previously error occurred
	switch oldConfig.Type {
	case intf.InterfaceType_TAP_INTERFACE:
		err = vppcalls.DeleteTapInterface(ifIdx, oldConfig.Tap.Version, plugin.vppCh, measure.GetTimeLog(tap.TapDelete{}, plugin.Stopwatch))
	case intf.InterfaceType_MEMORY_INTERFACE:
		err = vppcalls.DeleteMemifInterface(ifIdx, plugin.vppCh, measure.GetTimeLog(memif.MemifDelete{}, plugin.Stopwatch))
	case intf.InterfaceType_VXLAN_TUNNEL:
		err = vppcalls.DeleteVxlanTunnel(oldConfig.GetVxlan(), plugin.vppCh, measure.GetTimeLog(vxlan.VxlanAddDelTunnel{}, plugin.Stopwatch))
	case intf.InterfaceType_SOFTWARE_LOOPBACK:
		err = vppcalls.DeleteLoopbackInterface(ifIdx, plugin.vppCh, measure.GetTimeLog(vpe.DeleteLoopback{}, plugin.Stopwatch))
	case intf.InterfaceType_ETHERNET_CSMACD:
		return errors.New("it is not yet supported to remove (blacklist) physical interface")
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		err = plugin.afPacketConfigurator.DeleteAfPacketInterface(oldConfig)
	}
	if nil != err {
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
		plugin.ConfigureVPPInterface(pendingAfpacket)
	}
}

// ResolveDeletedLinuxInterface reacts to a removed Linux interface.
func (plugin *InterfaceConfigurator) ResolveDeletedLinuxInterface(interfaceName, hostIfName string) {
	plugin.Log.WithFields(logging.Fields{"ifName": interfaceName, "hostIfName": hostIfName}).Info("Linux interface was deleted")

	plugin.afPacketConfigurator.ResolveDeletedLinuxInterface(interfaceName, hostIfName)
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

func (plugin *InterfaceConfigurator) addContainerIPAddress(iface *intf.Interfaces_Interface, ifIdx uint32) error {
	addr, isIpv6, err := addrs.ParseIPWithPrefix(iface.ContainerIpAddress)
	if err != nil {
		return err
	}
	return vppcalls.AddContainerIP(ifIdx, addr, isIpv6, plugin.Log, plugin.vppCh,
		measure.GetTimeLog(ip.IPContainerProxyAddDel{}, plugin.Stopwatch))
}

func (plugin *InterfaceConfigurator) deleteContainerIPAddress(oldConfig *intf.Interfaces_Interface, ifIdx uint32) error {
	addr, isIpv6, err := addrs.ParseIPWithPrefix(oldConfig.ContainerIpAddress)
	if err != nil {
		return nil
	}
	return vppcalls.DelContainerIP(ifIdx, addr, isIpv6, plugin.Log, plugin.vppCh,
		measure.GetTimeLog(ip.IPContainerProxyAddDel{}, plugin.Stopwatch))
}
