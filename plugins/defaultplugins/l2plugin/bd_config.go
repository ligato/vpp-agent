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

//go:generate protoc --proto_path=../common/model/l2 --gogo_out=../common/model/l2 ../common/model/l2/l2.proto

//go:generate binapi-generator --input-file=/usr/share/vpp/api/l2.api.json --output-dir=../common/bin_api

// Package l2plugin implements the L2 plugin that handles Bridge Domains and L2 FIBs.
package l2plugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/safeclose"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bdidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// BDConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of bridge domains as modelled by the proto file "../model/l2/l2.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1bd". Updates received from the northbound API
// are compared with the VPP run-time configuration and differences are applied through the VPP binary API.
type BDConfigurator struct {
	Log          logging.Logger
	GoVppmux     govppmux.API
	ServiceLabel servicelabel.ReaderAPI

	// bridge domains
	BdIndices bdidx.BDIndexRW
	// interface indices
	SwIfIndices ifaceidx.SwIfIndex

	BridgeDomainIDSeq      uint32
	RegisteredIfaceCounter uint32
	vppChan                *govppapi.Channel
	notificationChan       chan BridgeDomainStateMessage
	Stopwatch              *measure.Stopwatch // timer used to measure and store time
}

// BridgeDomainStateMessage is message with bridge domain state + bridge domain name (since a state message does not
// contain it). This state is sent to the bd_state.go to further processing after every change.
type BridgeDomainStateMessage struct {
	Message govppapi.Message
	Name    string
}

// BridgeDomainMeta holds info about interfaces's bridge domain index and BVI.
type BridgeDomainMeta struct {
	bdIdx          uint32
	IsInterfaceBvi bool
}

// Init members (channels...) and start go routines.
func (plugin *BDConfigurator) Init(notificationChannel chan BridgeDomainStateMessage) (err error) {
	plugin.Log.Debug("Initializing L2 Bridge domains configurator")

	// Init VPP API channel
	plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	// Init notification channel.
	plugin.notificationChan = notificationChannel

	err = vppcalls.CheckMsgCompatibilityForBridgeDomains(plugin.Log, plugin.vppChan)
	if err != nil {
		return err
	}

	return nil
}

// Close GOVPP channel.
func (plugin *BDConfigurator) Close() error {
	_, err := safeclose.CloseAll(plugin.vppChan, plugin.notificationChan)
	return err
}

// ConfigureBridgeDomain handles the creation of new bridge domain including interfaces, ARP termination
// entries and pushes status update notification.
func (plugin *BDConfigurator) ConfigureBridgeDomain(bdConfig *l2.BridgeDomains_BridgeDomain) error {
	plugin.Log.Println("Configuring VPP Bridge Domain", bdConfig.Name)

	isValid, _ := plugin.vppValidateBridgeDomainBVI(bdConfig, nil)
	if !isValid {
		return nil
	}

	// Index of the new bridge domain and increment global index.
	bdIdx := plugin.BridgeDomainIDSeq
	plugin.BridgeDomainIDSeq++

	// Create bridge domain with respective index.
	err := vppcalls.VppAddBridgeDomain(bdIdx, bdConfig, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(l2ba.BridgeDomainAddDel{}, plugin.Stopwatch))

	if err != nil {
		plugin.Log.WithField("Bridge domain name", bdConfig.Name).Error(err)
		return err
	}

	// Find all interfaces belonging to this bridge domain and set them up.
	vppcalls.SetInterfacesToBridgeDomain(bdConfig, bdIdx, bdConfig.Interfaces, plugin.SwIfIndices, plugin.Log,
		plugin.vppChan, measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))

	// Resolve ARP termination table entries.
	arpTerminationTable := bdConfig.GetArpTerminationTable()
	if arpTerminationTable != nil && len(arpTerminationTable) != 0 {
		arpTable := bdConfig.ArpTerminationTable
		for _, arpEntry := range arpTable {
			err := vppcalls.VppAddArpTerminationTableEntry(bdIdx, arpEntry.PhysAddress, arpEntry.IpAddress,
				plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.BdIPMacAddDel{}, plugin.Stopwatch))
			if err != nil {
				plugin.Log.Error(err)
			}
		}
	} else {
		plugin.Log.WithField("Bridge domain name", bdConfig.Name).Debug("No ARP termination entries to set")
	}

	// Register created bridge domain.
	plugin.BdIndices.RegisterName(bdConfig.Name, bdIdx, bdConfig)
	plugin.Log.WithFields(logging.Fields{"Name": bdConfig.Name, "Index": bdIdx}).Debug("Bridge domain registered.")

	// Push to bridge domain state.
	errLookup := plugin.PropagateBdDetailsToStatus(bdIdx, bdConfig.Name)
	if errLookup != nil {
		plugin.Log.WithField("bdName", bdConfig.Name).Error(errLookup)
		return errLookup
	}

	return nil
}

// ModifyBridgeDomain processes the NB config and propagates it to bin api calls.
func (plugin *BDConfigurator) ModifyBridgeDomain(newBdConfig *l2.BridgeDomains_BridgeDomain, oldBdConfig *l2.BridgeDomains_BridgeDomain) error {
	plugin.Log.Infof("Modifying VPP bridge domain %v", newBdConfig.Name)

	// Validate updated config.
	isValid, recreate := plugin.vppValidateBridgeDomainBVI(newBdConfig, oldBdConfig)
	if !isValid {
		return nil
	}

	// In case bridge domain params changed, it needs to be recreated
	if recreate {
		if err := plugin.DeleteBridgeDomain(oldBdConfig); err != nil {
			return err
		}
		if err := plugin.ConfigureBridgeDomain(newBdConfig); err != nil {
			return err
		}
		plugin.Log.Infof("Bridge domain %v modify done.", newBdConfig.Name)
	} else {
		// Modify without recreation
		bdIdx, _, found := plugin.BdIndices.LookupIdx(oldBdConfig.Name)
		if !found {
			// If old config is missing, the diff cannot be done. Bridge domain will be created as a new one. This
			// case should NOT happen, it means that the agent's state is inconsistent.
			plugin.Log.Warnf("Bridge domain %v modify failed due to missing old configuration, will be created as a new one",
				newBdConfig.Name)
			return plugin.ConfigureBridgeDomain(newBdConfig)
		}

		// Update interfaces.
		toSet, toUnset := plugin.calculateIfaceDiff(newBdConfig.Interfaces, oldBdConfig.Interfaces)
		vppcalls.UnsetInterfacesFromBridgeDomain(newBdConfig, bdIdx, toUnset, plugin.SwIfIndices, plugin.Log,
			plugin.vppChan, measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))
		vppcalls.SetInterfacesToBridgeDomain(newBdConfig, bdIdx, toSet, plugin.SwIfIndices, plugin.Log,
			plugin.vppChan, measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))

		// Update ARP termination table.
		toAdd, toRemove := plugin.calculateARPDiff(newBdConfig.ArpTerminationTable, oldBdConfig.ArpTerminationTable)
		ipMacAddDelTimeLog := measure.GetTimeLog(l2ba.BdIPMacAddDel{}, plugin.Stopwatch)
		for _, entry := range toAdd {
			vppcalls.VppAddArpTerminationTableEntry(bdIdx, entry.PhysAddress, entry.IpAddress, plugin.Log,
				plugin.vppChan, ipMacAddDelTimeLog)
		}
		for _, entry := range toRemove {
			vppcalls.VppRemoveArpTerminationTableEntry(bdIdx, entry.PhysAddress, entry.IpAddress, plugin.Log,
				plugin.vppChan, ipMacAddDelTimeLog)
		}

		// Push change to bridge domain state.
		errLookup := plugin.PropagateBdDetailsToStatus(bdIdx, newBdConfig.Name)
		if errLookup != nil {
			plugin.Log.WithField("bdName", newBdConfig.Name).Error(errLookup)
			return errLookup
		}
	}

	return nil
}

// DeleteBridgeDomain processes the NB config and propagates it to bin api calls.
func (plugin *BDConfigurator) DeleteBridgeDomain(bdConfig *l2.BridgeDomains_BridgeDomain) error {
	plugin.Log.Infof("Deleting bridge domain %v", bdConfig.Name)

	bdIdx, _, found := plugin.BdIndices.LookupIdx(bdConfig.Name)
	if !found {
		plugin.Log.WithField("bdName", bdConfig.Name).Debug("Unable to find index for bridge domain to be deleted.")
		return nil
	}

	return plugin.deleteBridgeDomain(bdConfig, bdIdx)
}

func (plugin *BDConfigurator) deleteBridgeDomain(bdConfig *l2.BridgeDomains_BridgeDomain, bdIdx uint32) error {
	// Unmap all interfaces from removed bridge domain.
	vppcalls.UnsetInterfacesFromBridgeDomain(bdConfig, bdIdx, bdConfig.Interfaces,
		plugin.SwIfIndices, plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))

	err := vppcalls.VppDeleteBridgeDomain(bdIdx, plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.BridgeDomainAddDel{}, plugin.Stopwatch))
	if err != nil {
		return err
	}

	plugin.BdIndices.UnregisterName(bdConfig.Name)
	plugin.Log.WithFields(logging.Fields{"Name": bdConfig.Name, "bdIdx": bdIdx}).Debug("Bridge domain removed.")

	// Update bridge domain state.
	err = plugin.PropagateBdDetailsToStatus(bdIdx, bdConfig.Name)
	if err != nil {
		return err
	}

	return nil
}

// PropagateBdDetailsToStatus looks for existing VPP bridge domain state and propagates it to the etcd bd state.
func (plugin *BDConfigurator) PropagateBdDetailsToStatus(bdID uint32, bdName string) error {
	stateMsg := BridgeDomainStateMessage{}
	var wasError error

	_, _, found := plugin.BdIndices.LookupName(bdID)
	if !found {
		// If bridge domain does not exist in mapping, the lookup treats it as a removed bridge domain,
		// and ID in message is set to 0. Name has to be passed further in order
		// to be able to construct the key to remove the status from ETCD.
		stateMsg.Message = &l2ba.BridgeDomainDetails{
			BdID: 0,
		}
		stateMsg.Name = bdName
	} else {
		// Put current state data to status message.
		req := &l2ba.BridgeDomainDump{
			BdID: bdID,
		}
		reqContext := plugin.vppChan.SendRequest(req)
		msg := &l2ba.BridgeDomainDetails{}
		err := reqContext.ReceiveReply(msg)
		if err != nil {
			wasError = err
		}
		stateMsg.Message = msg
		stateMsg.Name = bdName
	}

	// Propagate bridge domain state information to the bridge domain state updater.
	plugin.notificationChan <- stateMsg

	return wasError
}

// ResolveCreatedInterface looks for bridge domain this interface is assigned to and sets it up.
func (plugin *BDConfigurator) ResolveCreatedInterface(ifName string, ifIdx uint32) error {
	plugin.Log.Infof("Assigning new interface %v to bridge domain", ifName)
	// Find bridge domain where the interface should be assigned
	bdIdx, bd, bvi, found := plugin.BdIndices.LookupBdForInterface(ifName)
	if !found {
		plugin.Log.Debugf("Interface %s does not belong to any bridge domain", ifName)
		return nil
	}

	vppcalls.SetInterfaceToBridgeDomain(bdIdx, ifIdx, bvi, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))

	// Push to bridge domain state.
	err := plugin.PropagateBdDetailsToStatus(bdIdx, bd.Name)
	if err != nil {
		return err
	}

	return nil
}

// ResolveDeletedInterface is called by VPP if an interface is removed.
func (plugin *BDConfigurator) ResolveDeletedInterface(ifName string) error {
	plugin.Log.Infof("Remove deleted interface %v from bridge domain", ifName)
	// Find bridge domain the interface should be removed from
	bdIdx, bd, _, found := plugin.BdIndices.LookupBdForInterface(ifName)
	if !found {
		plugin.Log.Debugf("Interface %s does not belong to any bridge domain", ifName)
		return nil
	}
	// If interface belonging to a bridge domain is removed, VPP handles internal bridge domain update itself.
	// However,the etcd operational state still needs to be updated to reflect changed VPP state.
	err := plugin.PropagateBdDetailsToStatus(bdIdx, bd.Name)
	if err != nil {
		return err
	}

	return nil
}

// The goal of the validation is to ensure that bridge domain does not contain more than one BVI interface
func (plugin *BDConfigurator) vppValidateBridgeDomainBVI(newBdConfig, oldBdConfig *l2.BridgeDomains_BridgeDomain) (bool, bool) {
	recreate := plugin.calculateBdParamsDiff(newBdConfig, oldBdConfig)
	if recreate {
		plugin.Log.Debugf("Bridge domain %v base params changed, will be recreated", newBdConfig.Name)
	}

	if len(newBdConfig.Interfaces) == 0 {
		plugin.Log.Warnf("Bridge domain %v does not contain any interface", newBdConfig.Name)
		return true, recreate
	}
	var bviCount int
	for _, bdInterface := range newBdConfig.Interfaces {
		if bdInterface.BridgedVirtualInterface {
			bviCount++
		}
	}
	if bviCount == 0 {
		plugin.Log.Debugf("Bridge domain %v does not contain any bvi interface", newBdConfig.Name)
		return true, recreate
	} else if bviCount == 1 {
		return true, recreate
	} else {
		plugin.Log.Warnf("Bridge domain %v contains more than one BVI interface. Correct it and create/modify bridge domain again", newBdConfig.Name)
		return false, recreate
	}
}

// Compares all base bridge domain params. Returns true if there is a difference, false otherwise.
func (plugin *BDConfigurator) calculateBdParamsDiff(newBdConfig, oldBdConfig *l2.BridgeDomains_BridgeDomain) bool {
	if oldBdConfig == nil {
		// nothing to compare
		return false
	}

	if newBdConfig.ArpTermination != oldBdConfig.ArpTermination {
		return true
	}
	if newBdConfig.Flood != oldBdConfig.Flood {
		return true
	}
	if newBdConfig.Forward != oldBdConfig.Forward {
		return true
	}
	if newBdConfig.Learn != oldBdConfig.Learn {
		return true
	}
	if newBdConfig.MacAge != oldBdConfig.MacAge {
		return true
	}
	if newBdConfig.UnknownUnicastFlood != oldBdConfig.UnknownUnicastFlood {
		return true
	}
	return false
}

// Returns lists of interfaces which will be set or unset to bridge domain
// Unset:
//	* all interfaces which are no longer part of the bridge domain
//	* interface which will be set as BVI (in case the BVI was changed)
//	* interface which will be set as non-BVI
// Set:
// 	* all new interfaces added to bridge domain
//  * interface which was a BVI before
//  * interface which will be the new BVI
func (plugin *BDConfigurator) calculateIfaceDiff(newIfaces, oldIfaces []*l2.BridgeDomains_BridgeDomain_Interfaces) (toSet, toUnset []*l2.BridgeDomains_BridgeDomain_Interfaces) {
	// Find BVI interfaces (it may not be configured)
	var oldBVI *l2.BridgeDomains_BridgeDomain_Interfaces
	var newBVI *l2.BridgeDomains_BridgeDomain_Interfaces
	for _, newIface := range newIfaces {
		if newIface.BridgedVirtualInterface {
			newBVI = newIface
			break
		}
	}
	for _, oldIface := range oldIfaces {
		if oldIface.BridgedVirtualInterface {
			oldBVI = oldIface
			break
		}
	}

	// If BVI was set/unset in general or the BVI interface was changed, pass the knowledge to the diff
	// resolution
	var bviChanged bool
	if oldBVI == nil && newBVI == nil {
		bviChanged = false
	} else if (oldBVI == nil && newBVI != nil) || (oldBVI != nil && newBVI == nil) || (oldBVI.Name != newBVI.Name) {
		bviChanged = true
	}

	// Resolve interfaces to unset
	for _, oldIface := range oldIfaces {
		var exists bool
		for _, newIface := range newIfaces {
			if oldIface.Name == newIface.Name {
				exists = true
			}
		}
		// Unset interface as an obsolete one
		if !exists {
			toUnset = append(toUnset, oldIface)
			continue
		}
		if bviChanged {
			// unset deprecated BVI interface
			if oldBVI != nil && oldBVI.Name == oldIface.Name {
				toUnset = append(toUnset, oldIface)
				continue
			}
			// unset non-BVI interface which will be subsequently set as BVI
			if newBVI != nil && newBVI.Name == oldIface.Name {
				toUnset = append(toUnset, oldIface)
			}
		}
	}

	// Resolve interfaces to set
	for _, newIface := range newIfaces {
		var exists bool
		for _, oldIface := range oldIfaces {
			if newIface.Name == oldIface.Name {
				exists = true
			}
		}
		// Set newly added interface
		if !exists {
			toSet = append(toSet, newIface)
			continue
		}
		if bviChanged {
			// Set non-BVI interface which was BVI before
			if oldBVI != nil && oldBVI.Name == newIface.Name {
				toSet = append(toSet, newIface)
				continue
			}
			// Set new BVI interface
			if newBVI != nil && newBVI.Name == newIface.Name {
				toSet = append(toSet, newIface)
			}
		}
	}

	return toSet, toUnset
}

// resolve diff of ARP entries
func (plugin *BDConfigurator) calculateARPDiff(newARPs, oldARPs []*l2.BridgeDomains_BridgeDomain_ArpTerminationTable) (toAdd, toRemove []*l2.BridgeDomains_BridgeDomain_ArpTerminationTable) {
	// Resolve ARPs to add
	for _, newARP := range newARPs {
		var exists bool
		for _, oldARP := range oldARPs {
			if newARP.IpAddress == oldARP.IpAddress && newARP.PhysAddress == oldARP.PhysAddress {
				exists = true
			}
		}
		if !exists {
			toAdd = append(toAdd, newARP)
		}
	}
	// Resolve ARPs to remove
	for _, oldARP := range oldARPs {
		var exists bool
		for _, newARP := range newARPs {
			if oldARP.IpAddress == newARP.IpAddress && oldARP.PhysAddress == newARP.PhysAddress {
				exists = true
			}
		}
		if !exists {
			toRemove = append(toRemove, oldARP)
		}
	}

	return toAdd, toRemove
}
