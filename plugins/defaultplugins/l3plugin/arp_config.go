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

package l3plugin

import (
	"fmt"
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

var msgCompatibilityARP = []govppapi.Message{
	&ip.IPNeighborAddDel{},
	&ip.IPNeighborAddDelReply{},
}

// ArpConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of L3 arp entries as modelled by the proto file "../model/l3/l3.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/arp". Updates received from the northbound API
// are compared with the VPP run-time configuration and differences are applied through the VPP binary API.
type ArpConfigurator struct {
	Log logging.Logger

	GoVppmux govppmux.API

	// ARPIndexes is a list of ARP entries which are successfully configured on the VPP
	ARPIndexes l3idx.ARPIndexRW
	// ARPCache is a list of ARP entries with are present in the ETCD, but not on VPP
	// due to missing interface
	ARPCache l3idx.ARPIndexRW
	// ARPDeleted is a list of unsuccessfully deleted ARP entries. ARP entry cannot be removed
	// if the interface is missing (it runs into 'unnassigned' state). If the interface re-appears,
	// such an ARP have to be removed
	ARPDeleted l3idx.ARPIndexRW

	ARPIndexSeq uint32
	SwIfIndexes ifaceidx.SwIfIndex
	vppChan     *govppapi.Channel

	Stopwatch *measure.Stopwatch
}

// Init initializes ARP configurator
func (plugin *ArpConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing ARP configurator")

	// Init VPP API channel
	plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	if err := plugin.vppChan.CheckMessageCompatibility(msgCompatibilityARP...); err != nil {
		plugin.Log.Error(err)
		return err
	}

	return nil
}

// Close GOVPP channel
func (plugin *ArpConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// Creates unique identifier which serves as a name in name to index mapping
func arpIdentifier(iface, ip, mac string) string {
	return fmt.Sprintf("arp-iface-%v-%v-%v", iface, ip, mac)
}

// AddArp processes the NB config and propagates it to bin api call
func (plugin *ArpConfigurator) AddArp(entry *l3.ArpTable_ArpTableEntry) error {
	plugin.Log.Infof("Configuring new ARP entry %v", *entry)

	if !isValidARP(entry, plugin.Log) {
		return fmt.Errorf("cannot configure ARP, provided data is not valid")
	}

	arpID := arpIdentifier(entry.Interface, entry.PhysAddress, entry.IpAddress)

	// look for ARP in list of deleted ARPs
	_, _, exists := plugin.ARPDeleted.UnregisterName(arpID)
	if exists {
		plugin.Log.Debugf("ARP entry %v recreated", arpID)
	}

	// verify interface presence
	ifIndex, _, exists := plugin.SwIfIndexes.LookupIdx(entry.Interface)
	if !exists {
		// Store ARP entry to cache
		plugin.Log.Debugf("Interface %v required by ARP entry not found, moving to cache", entry.Interface)
		plugin.ARPCache.RegisterName(arpID, plugin.ARPIndexSeq, entry)
		plugin.ARPIndexSeq++
		return nil
	}

	// Transform arp data
	arp, err := transformArp(entry, ifIndex)
	if err != nil {
		return err
	}
	if arp == nil {
		return nil
	}
	plugin.Log.Debugf("adding ARP: %+v", *arp)

	// Create and register new arp entry
	err = vppcalls.VppAddArp(arp, plugin.vppChan,
		measure.GetTimeLog(ip.IPNeighborAddDel{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	// Register configured ARP
	plugin.ARPIndexes.RegisterName(arpID, plugin.ARPIndexSeq, entry)
	plugin.ARPIndexSeq++
	plugin.Log.Debugf("ARP entry %v registered", arpID)

	plugin.Log.Infof("ARP entry %v configured", arpID)

	return nil
}

// ChangeArp processes the NB config and propagates it to bin api call
func (plugin *ArpConfigurator) ChangeArp(entry *l3.ArpTable_ArpTableEntry, prevEntry *l3.ArpTable_ArpTableEntry) error {
	plugin.Log.Infof("Modifying ARP entry %v to %v", *prevEntry, *entry)

	if err := plugin.DeleteArp(prevEntry); err != nil {
		return err
	}
	if err := plugin.AddArp(entry); err != nil {
		return err
	}

	plugin.Log.Infof("ARP entry %v modified to %v", *prevEntry, *entry)

	return nil
}

// DeleteArp processes the NB config and propagates it to bin api call
func (plugin *ArpConfigurator) DeleteArp(entry *l3.ArpTable_ArpTableEntry) error {
	plugin.Log.Infof("Removing ARP entry %v", *entry)

	if !isValidARP(entry, plugin.Log) {
		// Note: such an ARP cannot be configured either, so it should not happen
		return fmt.Errorf("cannot remove ARP, provided data is not valid")
	}

	// ARP entry identifier
	arpID := arpIdentifier(entry.Interface, entry.PhysAddress, entry.IpAddress)

	// Check if ARP entry is not just cached
	_, _, found := plugin.ARPCache.LookupIdx(arpID)
	if found {
		plugin.Log.Debugf("ARP entry %v found in cache, removed", arpID)
		plugin.ARPCache.UnregisterName(arpID)
		// Cached ARP is not configured on the VPP, return
		return nil
	}

	// Check interface presence
	ifIndex, _, exists := plugin.SwIfIndexes.LookupIdx(entry.Interface)
	if !exists {
		// ARP entry cannot be removed without interface. Since the data are
		// no longer in the ETCD, agent need to remember the state and remove
		// entry when possible
		plugin.Log.Infof("Cannot remove ARP entry %v due to missing interface, will be removed when possible",
			entry.Interface)
		plugin.ARPIndexes.UnregisterName(arpID)
		plugin.ARPDeleted.RegisterName(arpID, plugin.ARPIndexSeq, entry)
		plugin.ARPIndexSeq++

		return nil
	}

	// Transform arp data
	arp, err := transformArp(entry, ifIndex)
	if err != nil {
		return err
	}
	if arp == nil {
		return nil
	}
	plugin.Log.Debugf("deleting ARP: %+v", arp)

	// Delete and un-register new arp
	err = vppcalls.VppDelArp(arp, plugin.vppChan,
		measure.GetTimeLog(ip.IPNeighborAddDel{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	_, _, found = plugin.ARPIndexes.UnregisterName(arpID)
	if found {
		plugin.Log.Infof("ARP entry %v unregistered", arpID)
	} else {
		plugin.Log.Warnf("Un-register failed, ARP entry %v not found", arpID)
	}

	plugin.Log.Infof("ARP entry %v removed", arpID)

	return nil
}

// ResolveCreatedInterface handles case when new interface appears in the config
func (plugin *ArpConfigurator) ResolveCreatedInterface(interfaceName string) error {
	plugin.Log.Debugf("ARP configurator: resolving new interface %v", interfaceName)
	// find all entries which can be resolved
	entriesToAdd := plugin.ARPCache.LookupNamesByInterface(interfaceName)
	entriesToRemove := plugin.ARPDeleted.LookupNamesByInterface(interfaceName)

	var wasErr error
	// Configure all cached ARP entriesToAdd which can be configured
	for _, entry := range entriesToAdd {
		// ARP entry identifier. Every entry in cache was already validated
		arpID := arpIdentifier(entry.Interface, entry.PhysAddress, entry.IpAddress)
		if err := plugin.AddArp(entry); err != nil {
			wasErr = err
		}

		// remove from cache
		plugin.ARPCache.UnregisterName(arpID)
		plugin.Log.Infof("Previously un-configurable ARP entry %v is now configured", arpID)
	}

	// Remove all entries which should not be configured
	for _, entry := range entriesToRemove {
		arpID := arpIdentifier(entry.Interface, entry.PhysAddress, entry.IpAddress)
		if err := plugin.DeleteArp(entry); err != nil {
			wasErr = err
		}

		// remove from list of deleted
		plugin.ARPDeleted.UnregisterName(arpID)
		plugin.Log.Infof("Deprecated ARP entry %v was removed", arpID)
	}

	return wasErr
}

// ResolveDeletedInterface handles case when interface is removed from the config
func (plugin *ArpConfigurator) ResolveDeletedInterface(interfaceName string, interfaceIdx uint32) error {
	plugin.Log.Debugf("ARP configurator: resolving deleted interface %v", interfaceName)

	// Since the interface does not exist, all related ARP entries are 'un-assigned' on the VPP
	// but they cannot be removed using binary API. Nothing to do here.

	return nil
}

// Verify ARP entry contains all required fields
func isValidARP(arpInput *l3.ArpTable_ArpTableEntry, log logging.Logger) bool {
	if arpInput == nil {
		log.Info("ARP input is empty")
		return false
	}
	if arpInput.Interface == "" {
		log.Info("ARP input does not contain interface")
		return false
	}
	if arpInput.IpAddress == "" {
		log.Info("ARP input does not contain IP")
		return false
	}
	if arpInput.PhysAddress == "" {
		log.Info("ARP input does not contain MAC")
		return false
	}

	return true
}

// transformArp converts raw entry data to ARP object
func transformArp(arpInput *l3.ArpTable_ArpTableEntry, ifIndex uint32) (*vppcalls.ArpEntry, error) {
	ipAddr := net.ParseIP(arpInput.IpAddress)
	macAddr, err := net.ParseMAC(arpInput.PhysAddress)
	if err != nil {
		return nil, err
	}
	arp := &vppcalls.ArpEntry{
		Interface:  ifIndex,
		IPAddress:  ipAddr,
		MacAddress: macAddr,
		Static:     arpInput.Static,
	}
	return arp, nil
}
