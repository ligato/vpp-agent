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
//go:generate binapi-generator --input-file=/usr/share/vpp/api/vpe.api.json --output-dir=bin_api
//go:generate binapi-generator --input-file=/usr/share/vpp/api/vxlan.api.json --output-dir=bin_api

// Package ifplugin is the implementation of the Interface plugin.deps.
package ifplugin

import (
	"bytes"
	"errors"

	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linuxplugin"
)

// InterfaceConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/interfaces/interfaces.proto"
// and stored in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1interface".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type InterfaceConfigurator struct {
	GoVppmux     govppmux.API
	ServiceLabel servicelabel.ReaderAPI
	Linux        linuxplugin.API

	swIfIndexes ifaceidx.SwIfIndexRW

	afPacketConfigurator *AFPacketConfigurator

	vppCh     *govppapi.Channel
	notifChan chan govppapi.Message // to publish SwInterfaceDetails to interface_state.go

	resyncDoneOnce bool
}

// Init members (channels...) and start go routines
func (plugin *InterfaceConfigurator) Init(swIfIndexes ifaceidx.SwIfIndexRW, notifChan chan govppapi.Message) (err error) {
	log.DefaultLogger().Debug("Initializing InterfaceConfigurator")
	plugin.swIfIndexes = swIfIndexes
	plugin.notifChan = notifChan

	plugin.vppCh, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}
	err = vppcalls.CheckMsgCompatibilityForInterface(plugin.vppCh)
	if err != nil {
		return err
	}

	plugin.afPacketConfigurator = &AFPacketConfigurator{Linux: plugin.Linux}
	plugin.afPacketConfigurator.Init(plugin.vppCh)

	return nil
}

// Close GOVPP channel
func (plugin *InterfaceConfigurator) Close() error {
	return safeclose.Close(plugin.vppCh)
}

// LookupVPPInterfaces looks up all VPP interfaces and saves their name-to-index mapping and state information.
func (plugin *InterfaceConfigurator) LookupVPPInterfaces() error {
	req := &interfaces.SwInterfaceDump{}
	reqCtx := plugin.vppCh.SendMultiRequest(req)

	for {
		msg := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			log.DefaultLogger().Error(err)
			return err
		}

		// store the name to index mapping if it does not exist yet
		_, _, found := plugin.swIfIndexes.LookupName(msg.SwIfIndex)
		if !found {
			ifName := string(bytes.Trim(msg.InterfaceName, "\x00"))
			log.DefaultLogger().WithFields(log.Fields{"ifName": ifName, "swIfIndex": msg.SwIfIndex}).
				Debug("Register VPP interface name mapping.")

			plugin.swIfIndexes.RegisterName(ifName, msg.SwIfIndex, nil)
		}

		// propagate interface state information
		plugin.notifChan <- msg
	}

	return nil
}

// ConfigureVPPInterface reacts to a new northbound VPP interface config by creating and configuring
// the interface in the VPP network stack through the VPP binary API.
func (plugin *InterfaceConfigurator) ConfigureVPPInterface(iface *intf.Interfaces_Interface) error {
	log.DefaultLogger().WithField("ifName", iface.Name).Debug("Configuring VPP interface")
	var ifIdx uint32
	var err error
	var exists bool
	var pending bool

	switch iface.Type {
	case intf.InterfaceType_TAP_INTERFACE:
		ifIdx, err = vppcalls.AddTapInterface(iface.Tap, plugin.vppCh)
	case intf.InterfaceType_MEMORY_INTERFACE:
		ifIdx, err = vppcalls.AddMemifInterface(iface.Memif, plugin.vppCh)
	case intf.InterfaceType_VXLAN_TUNNEL:
		ifIdx, err = vppcalls.AddVxlanTunnel(iface.Vxlan, plugin.vppCh)
	case intf.InterfaceType_SOFTWARE_LOOPBACK:
		ifIdx, err = vppcalls.AddLoopbackInterface(iface.Name, plugin.vppCh)
	case intf.InterfaceType_ETHERNET_CSMACD:
		ifIdx, _, exists = plugin.swIfIndexes.LookupIdx(iface.Name)
		if !exists {
			return errors.New("it is not yet supported to add (whitelist) a new physical interface")
		}
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		ifIdx, pending, err = plugin.afPacketConfigurator.ConfigureAfPacketInterface(iface)
	}

	if nil != err {
		return err
	}
	if pending {
		// interface cannot be created yet and will be configured later
		return nil
	}

	var wasError error

	// configure optional mac address
	if iface.PhysAddress != "" {
		err := vppcalls.SetInterfaceMac(ifIdx, iface.PhysAddress, plugin.vppCh)
		if err != nil {
			wasError = err
		}
	}

	// configure optional ip address
	newAddrs, err := addrs.StrAddrsToStruct(iface.IpAddresses)
	if err != nil {
		return err
	}
	for i := range newAddrs {
		err := vppcalls.AddInterfaceIP(ifIdx, newAddrs[i], plugin.vppCh)
		if nil != err {
			wasError = err
		}
	}

	// register name to idx mapping
	plugin.swIfIndexes.RegisterName(iface.Name, ifIdx, iface)
	log.DefaultLogger().WithFields(log.Fields{"ifName": iface.Name, "ifIdx": ifIdx}).Info("Configured interface")

	// set interface up if enabled
	// NOTE: needs to be called after RegisterName, otherwise interface up/down notification won't map to a valid interface
	if iface.Enabled {
		err := vppcalls.InterfaceAdminUp(ifIdx, plugin.vppCh)
		if nil != err {
			return err
		}
	}

	// load interface state data for newly added interface (no way to filter by swIfIndex, need to dump all of them)
	plugin.LookupVPPInterfaces()

	return wasError
}

// ModifyVPPInterface applies changes in the NB configuration of a VPP interface into the running VPP
// through the VPP binary API.
func (plugin *InterfaceConfigurator) ModifyVPPInterface(newConfig *intf.Interfaces_Interface,
	oldConfig *intf.Interfaces_Interface) error {

	if newConfig == nil {
		return errors.New("newConfig is null")
	}
	if oldConfig == nil {
		return errors.New("oldConfig is null")
	}

	log.DefaultLogger().Debug("'Modifying' VPP interface", newConfig.Name)

	if plugin.afPacketConfigurator.IsPendingAfPacket(oldConfig) {
		return plugin.recreateVPPInterface(newConfig, oldConfig, 0)
	}

	// lookup index
	ifIdx, meta, found := plugin.swIfIndexes.LookupIdx(newConfig.Name)

	if !found {
		log.DefaultLogger().WithFields(log.Fields{"ifName": newConfig.Name}).Debug("Mapping for interface name not found.")
		return nil
	}

	return plugin.modifyVPPInterface(newConfig, oldConfig, ifIdx, meta.Type)
}

// ModifyVPPInterface applies changes in the NB configuration of a VPP interface into the running VPP
// through the VPP binary API.
func (plugin *InterfaceConfigurator) modifyVPPInterface(newConfig *intf.Interfaces_Interface,
	oldConfig *intf.Interfaces_Interface, ifIdx uint32, ifaceType intf.InterfaceType) (err error) {

	log.DefaultLogger().WithFields(log.Fields{"ifname": oldConfig.Name, "swIfIndex": ifIdx}).
		Debug("modifyVPPInterface begin")

	switch ifaceType {
	case intf.InterfaceType_TAP_INTERFACE:
	case intf.InterfaceType_MEMORY_INTERFACE:
		if !canMemifBeModifWithoutDelete(newConfig.Memif, oldConfig.Memif) {
			err := plugin.recreateVPPInterface(newConfig, oldConfig, ifIdx)
			log.DefaultLogger().WithFields(log.Fields{"ifName": newConfig.Name, "ifIdx": ifIdx}).
				Debug("modifyVPPInterface end. ", err)
			return err
		}
	case intf.InterfaceType_VXLAN_TUNNEL:
		if !canVxlanBeModifWithoutDelete(newConfig.Vxlan, oldConfig.Vxlan) {
			err := plugin.recreateVPPInterface(newConfig, oldConfig, ifIdx)
			log.DefaultLogger().WithFields(log.Fields{"ifName": newConfig.Name, "ifIdx": ifIdx}).
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
			log.DefaultLogger().WithFields(log.Fields{"ifName": newConfig.Name, "ifIdx": ifIdx}).
				Debug("modifyVPPInterface end. ", err)
			return err
		}
	}

	var wasError error

	// admin status
	if newConfig.Enabled != oldConfig.Enabled {
		if newConfig.Enabled {
			err = vppcalls.InterfaceAdminUp(ifIdx, plugin.vppCh)
		} else {
			err = vppcalls.InterfaceAdminDown(ifIdx, plugin.vppCh)
		}
		if nil != err {
			wasError = err
		}
	}

	// configure new mac address if set (and only if it was changed)
	if newConfig.PhysAddress != "" && newConfig.PhysAddress != oldConfig.PhysAddress {
		err := vppcalls.SetInterfaceMac(ifIdx, newConfig.PhysAddress, plugin.vppCh)
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

	del, add := addrs.DiffAddr(newAddrs, oldAddrs)

	log.DefaultLogger().Debug("del ip addrs: ", del)
	log.DefaultLogger().Debug("add ip addrs: ", add)

	for i := range del {
		err := vppcalls.DelInterfaceIP(ifIdx, del[i], plugin.vppCh)
		log.DefaultLogger().Debug("del ip addr ", ifIdx, " ", del[i], " ", err)
		if nil != err {
			wasError = err
		}
	}

	for i := range add {
		err := vppcalls.AddInterfaceIP(ifIdx, add[i], plugin.vppCh)
		log.DefaultLogger().Debug("add ip addr ", ifIdx, " ", add[i], " ", err)
		if nil != err {
			wasError = err
		}
	}

	log.DefaultLogger().WithFields(log.Fields{"ifName": newConfig.Name, "ifIdx": ifIdx}).Debug("modifyVPPInterface end. ", err)

	return wasError
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
	log.DefaultLogger().Debug("'Deleting' VPP interface", iface.Name)

	if plugin.afPacketConfigurator.IsPendingAfPacket(iface) {
		return plugin.afPacketConfigurator.DeleteAfPacketInterface(iface)
	}

	// unregister name to init mapping (following triggers notifications for all subscribers)
	ifIdx, prev, found := plugin.swIfIndexes.UnregisterName(iface.Name)
	if !found {
		log.DefaultLogger().WithField("ifname", iface.Name).Debug("Unable to find index for interface to be deleted.")
		return nil
	}

	return plugin.deleteVPPInterface(prev, ifIdx)
}

func (plugin *InterfaceConfigurator) deleteVPPInterface(oldConfig *intf.Interfaces_Interface, ifIdx uint32) (
	wasError error) {
	log.DefaultLogger().WithFields(log.Fields{"ifname": oldConfig.Name, "swIfIndex": ifIdx}).
		Debug("deleteVPPInterface begin")

	// let's try to do following even if previously error occurred
	err := vppcalls.InterfaceAdminDown(ifIdx, plugin.vppCh)
	if nil != err {
		wasError = err
	}

	// let's try to do following even if previously error occurred
	oldAddrs, err := addrs.StrAddrsToStruct(oldConfig.IpAddresses)
	if err != nil {
		log.DefaultLogger().WithFields(log.Fields{"ifname": oldConfig.Name, "swIfIndex": ifIdx}).
			Debug("deleteVPPInterface end ", err)
		return err
	}
	for i := range oldAddrs {
		log.DefaultLogger().WithField("addr", oldAddrs[i]).Info("Ip removed")
		err := vppcalls.DelInterfaceIP(ifIdx, oldAddrs[i], plugin.vppCh)
		if nil != err {
			log.DefaultLogger().Error(err)
			wasError = err
		}
	}

	log.DefaultLogger().Info("Ip addrs removed")

	// let's try to do following even if previously error occurred
	switch oldConfig.Type {
	case intf.InterfaceType_TAP_INTERFACE:
		err = vppcalls.DeleteTapInterface(ifIdx, plugin.vppCh)
	case intf.InterfaceType_MEMORY_INTERFACE:
		err = vppcalls.DeleteMemifInterface(ifIdx, plugin.vppCh)
	case intf.InterfaceType_VXLAN_TUNNEL:
		err = vppcalls.DeleteVxlanTunnel(oldConfig.GetVxlan(), plugin.vppCh)
	case intf.InterfaceType_SOFTWARE_LOOPBACK:
		err = vppcalls.DeleteLoopbackInterface(ifIdx, plugin.vppCh)
	case intf.InterfaceType_ETHERNET_CSMACD:
		return errors.New("it is not yet supported to remove (blacklist) physical interface")
	case intf.InterfaceType_AF_PACKET_INTERFACE:
		err = plugin.afPacketConfigurator.DeleteAfPacketInterface(oldConfig)
	}
	if nil != err {
		wasError = err
	}

	log.DefaultLogger().WithFields(log.Fields{"ifname": oldConfig.Name, "swIfIndex": ifIdx}).
		Debug("deleteVPPInterface end ", err)

	return wasError
}

// ResolveCreatedLinuxInterface reacts to a newly created Linux interface.
func (plugin *InterfaceConfigurator) ResolveCreatedLinuxInterface(interfaceName string, interfaceIndex uint32) {
	log.DefaultLogger().WithFields(log.Fields{"ifName": interfaceName, "ifIdx": interfaceIndex}).Info("New Linux interface was created")

	pendingAfpacket := plugin.afPacketConfigurator.ResolveCreatedLinuxInterface(interfaceName, interfaceIndex)
	if pendingAfpacket != nil {
		// there is a pending afpacket that can be now configured
		plugin.ConfigureVPPInterface(pendingAfpacket)
	}
}

// ResolveDeletedLinuxInterface reacts to a removed Linux interface.
func (plugin *InterfaceConfigurator) ResolveDeletedLinuxInterface(interfaceName string) {
	log.DefaultLogger().WithFields(log.Fields{"ifName": interfaceName}).Info("Linux interface was deleted")

	plugin.afPacketConfigurator.ResolveDeletedLinuxInterface(interfaceName)
}

func canMemifBeModifWithoutDelete(newConfig *intf.Interfaces_Interface_Memif, oldConfig *intf.Interfaces_Interface_Memif) bool {
	if newConfig.RingSize == 0 {
		newConfig.RingSize = 1 // default
	}

	if newConfig.BufferSize != oldConfig.BufferSize || newConfig.Id != oldConfig.Id || newConfig.Secret != oldConfig.Secret ||
		newConfig.RingSize != oldConfig.RingSize || newConfig.Master != oldConfig.Master || newConfig.SocketFilename != oldConfig.SocketFilename ||
		newConfig.RxQueues != oldConfig.RxQueues || newConfig.TxQueues != oldConfig.TxQueues {

		log.DefaultLogger().Warn("Difference between new & old config causing recreation of memif ", oldConfig)

		return false
	}

	return true
}

func canVxlanBeModifWithoutDelete(newConfig *intf.Interfaces_Interface_Vxlan, oldConfig *intf.Interfaces_Interface_Vxlan) bool {
	if newConfig.SrcAddress != oldConfig.SrcAddress || newConfig.DstAddress != oldConfig.DstAddress || newConfig.Vni != oldConfig.Vni {
		return false
	}

	return true
}
