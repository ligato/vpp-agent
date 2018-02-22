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

package ifplugin

import (
	"errors"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/af_packet"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
)

// AFPacketConfigurator is used by InterfaceConfigurator to execute afpacket-specific management operations.
// Most importantly it needs to ensure that Afpacket interface is create AFTER the associated host interface.
type AFPacketConfigurator struct {
	logging.Logger
	Linux       interface{}        //just flag if nil
	Stopwatch   *measure.Stopwatch // from InterfaceConfigurator
	SwIfIndexes ifaceidx.SwIfIndexRW

	afPacketByHostIf map[string]*AfPacketConfig // host interface name -> Af Packet interface configuration
	afPacketByName   map[string]*AfPacketConfig // af packet name -> Af Packet interface configuration
	hostInterfaces   map[string]struct{}        // a set of available host interfaces

	vppCh *govppapi.Channel // govpp channel used by InterfaceConfigurator
}

// AfPacketConfig wraps the proto formatted configuration of an Afpacket interface together with a flag
// that tells if the interface is waiting for a host interface to get created.
type AfPacketConfig struct {
	config  *intf.Interfaces_Interface
	pending bool
}

// Init members of AFPacketConfigurator.
func (plugin *AFPacketConfigurator) Init(vppCh *govppapi.Channel) (err error) {
	plugin.Infof("Initializing AF-Packet configurator")
	plugin.vppCh = vppCh

	plugin.afPacketByHostIf = make(map[string]*AfPacketConfig)
	plugin.afPacketByName = make(map[string]*AfPacketConfig)
	plugin.hostInterfaces = make(map[string]struct{})
	return nil
}

// ConfigureAfPacketInterface creates a new Afpacket interface or marks it as pending if the target host interface doesn't exist yet.
func (plugin *AFPacketConfigurator) ConfigureAfPacketInterface(afpacket *intf.Interfaces_Interface) (swIndex uint32, pending bool, err error) {
	if afpacket.Type != intf.InterfaceType_AF_PACKET_INTERFACE {
		return 0, false, errors.New("expecting AfPacket interface")
	}

	if plugin.Linux != nil {
		_, hostIfAvail := plugin.hostInterfaces[afpacket.Afpacket.HostIfName]
		if !hostIfAvail {
			plugin.addToCache(afpacket, true)
			return 0, true, nil
		}
	}
	swIdx, err := vppcalls.AddAfPacketInterface(afpacket.Name, afpacket.Afpacket, plugin.vppCh, measure.GetTimeLog(af_packet.AfPacketCreate{}, plugin.Stopwatch))
	if err != nil {
		plugin.addToCache(afpacket, true)
		return 0, true, err
	}
	plugin.addToCache(afpacket, false)
	// If the interface is not in pending state, register it
	plugin.SwIfIndexes.RegisterName(afpacket.Name, swIdx, afpacket)
	return swIdx, false, nil
}

// ModifyAfPacketInterface updates the cache with afpacket configurations and tells InterfaceConfigurator if the interface
// needs to be recreated for the changes to be applied.
func (plugin *AFPacketConfigurator) ModifyAfPacketInterface(newConfig *intf.Interfaces_Interface,
	oldConfig *intf.Interfaces_Interface) (recreate bool, err error) {

	if oldConfig.Type != intf.InterfaceType_AF_PACKET_INTERFACE ||
		newConfig.Type != intf.InterfaceType_AF_PACKET_INTERFACE {
		return false, errors.New("expecting AfPacket interface")
	}

	afpacket, found := plugin.afPacketByName[oldConfig.Name]
	if !found || afpacket.pending || (newConfig.Afpacket.HostIfName != oldConfig.Afpacket.HostIfName) {
		return true, nil
	}

	// rewrite cached configuration
	plugin.addToCache(newConfig, false)
	return false, nil
}

// DeleteAfPacketInterface removes Afpacket interface from VPP and from the cache.
func (plugin *AFPacketConfigurator) DeleteAfPacketInterface(afpacket *intf.Interfaces_Interface, ifIdx uint32) (err error) {

	if afpacket.Type != intf.InterfaceType_AF_PACKET_INTERFACE {
		return errors.New("expecting AfPacket interface")
	}

	config, found := plugin.afPacketByName[afpacket.Name]
	if !found || !config.pending {
		err = vppcalls.DeleteAfPacketInterface(afpacket.Name, ifIdx, afpacket.GetAfpacket(), plugin.vppCh, measure.GetTimeLog(af_packet.AfPacketDelete{}, plugin.Stopwatch))
		// unregister interface to let other plugins know that it is removed from the vpp
		plugin.SwIfIndexes.UnregisterName(afpacket.Name)
	}
	plugin.removeFromCache(afpacket)
	return err
}

// ResolveCreatedLinuxInterface reacts to a newly created Linux interface.
func (plugin *AFPacketConfigurator) ResolveCreatedLinuxInterface(interfaceName, hostIfName string, interfaceIndex uint32) *intf.Interfaces_Interface {
	if plugin.Linux == nil {
		plugin.WithFields(logging.Fields{"ifName": interfaceName, "hostIfName": hostIfName}).Warn("Unexpectedly learned about a new Linux interface")
		return nil
	}
	plugin.hostInterfaces[hostIfName] = struct{}{}

	afpacket, found := plugin.afPacketByHostIf[hostIfName]
	if found {
		if !afpacket.pending {
			// this should not happen, log as warning
			plugin.WithFields(logging.Fields{"ifName": interfaceName, "hostIfName": hostIfName}).Warn(
				"Re-creating already configured AFPacket interface")
			// remove the existing afpacket and let the interface configurator to re-create it
			plugin.DeleteAfPacketInterface(afpacket.config, interfaceIndex)
		}
		// afpacket is now free to get created
		return afpacket.config
	}
	return nil // nothing to configure
}

// ResolveDeletedLinuxInterface reacts to a removed Linux interface.
func (plugin *AFPacketConfigurator) ResolveDeletedLinuxInterface(interfaceName, hostIfName string, ifIdx uint32) {
	if plugin.Linux == nil {
		plugin.WithFields(logging.Fields{"ifName": interfaceName, "hostIfName": hostIfName}).Warn("Unexpectedly learned about removed Linux interface")
		return
	}
	delete(plugin.hostInterfaces, hostIfName)

	afpacket, found := plugin.afPacketByHostIf[hostIfName]
	if found {
		// remove the interface and re-add as pending
		plugin.DeleteAfPacketInterface(afpacket.config, ifIdx)
		plugin.ConfigureAfPacketInterface(afpacket.config)
	}
}

// IsPendingAfPacket returns true if the given config belongs to pending Afpacket interface.
func (plugin *AFPacketConfigurator) IsPendingAfPacket(iface *intf.Interfaces_Interface) (pending bool) {
	afpacket, found := plugin.afPacketByName[iface.Name]
	return found && afpacket.pending
}

func (plugin *AFPacketConfigurator) addToCache(afpacket *intf.Interfaces_Interface, pending bool) {
	config := &AfPacketConfig{config: afpacket, pending: pending}
	plugin.afPacketByHostIf[afpacket.Afpacket.HostIfName] = config
	plugin.afPacketByName[afpacket.Name] = config
	plugin.Debugf("Afpacket interface with name %v added to cache (hostIf: %s, pending: %t)",
		afpacket.Name, afpacket.Afpacket.HostIfName, pending)
}

func (plugin *AFPacketConfigurator) removeFromCache(afpacket *intf.Interfaces_Interface) {
	delete(plugin.afPacketByName, afpacket.Name)
	delete(plugin.afPacketByHostIf, afpacket.Afpacket.HostIfName)
	plugin.Debugf("Afpacket interface with name %v removed from cache", afpacket.Name)
}
