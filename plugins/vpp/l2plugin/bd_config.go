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

//go:generate protoc --proto_path=../model/l2 --gogo_out=../model/l2 ../model/l2/l2.proto

// Package l2plugin implements the L2 plugin that handles Bridge Domains and L2 FIBs.
package l2plugin

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/l2"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/l2idx"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l2"
)

// BDConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of bridge domains as modelled by the proto file "../model/l2/l2.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1bd". Updates received from the northbound API
// are compared with the VPP run-time configuration and differences are applied through the VPP binary API.
type BDConfigurator struct {
	log logging.Logger

	// In-memory mappings
	ifIndexes    ifaceidx.SwIfIndex
	bdIndexes    l2idx.BDIndexRW
	bdIDSeq      uint32
	regIfCounter uint32

	// VPP channel
	vppChan govppapi.Channel

	// State notification channel
	notificationChan chan BridgeDomainStateMessage // Injected, do not close here

	// VPP API handlers
	ifHandler ifvppcalls.IfVppAPI

	// Timer used to measure and store time
	stopwatch *measure.Stopwatch
}

// BridgeDomainStateMessage is message with bridge domain state + bridge domain name (since a state message does not
// contain it). This state is sent to the bd_state.go to further processing after every change.
type BridgeDomainStateMessage struct {
	Message govppapi.Message
	Name    string
}

// GetSwIfIndexes exposes interface name-to-index mapping
func (plugin *BDConfigurator) GetBdIndexes() l2idx.BDIndexRW {
	return plugin.bdIndexes
}

// Init members (channels...) and start go routines.
func (plugin *BDConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, swIfIndexes ifaceidx.SwIfIndex, notificationChannel chan BridgeDomainStateMessage, enableStopwatch bool) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-l2-bd-conf")
	plugin.log.Debug("Initializing L2 Bridge domains configurator")

	// Mappings
	plugin.ifIndexes = swIfIndexes
	plugin.bdIndexes = l2idx.NewBDIndex(nametoidx.NewNameToIdx(plugin.log, "bd_indexes", l2idx.IndexMetadata))
	plugin.bdIDSeq = 1
	plugin.regIfCounter = 1

	// VPP channel
	plugin.vppChan, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}

	// Init notification channel.
	plugin.notificationChan = notificationChannel

	// Stopwatch
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("ACLConfigurator", plugin.log)
	}

	// VPP API handlers
	if plugin.ifHandler, err = ifvppcalls.NewIfVppHandler(plugin.vppChan, plugin.log, plugin.stopwatch); err != nil {
		return err
	}

	// Message compatibility
	err = plugin.vppChan.CheckMessageCompatibility(vppcalls.BridgeDomainMessages...)
	if err != nil {
		return err
	}

	return nil
}

// Close GOVPP channel.
func (plugin *BDConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// clearMapping prepares all in-memory-mappings and other cache fields. All previous cached entries are removed.
func (plugin *BDConfigurator) clearMapping() {
	plugin.bdIndexes.Clear()
}

// ConfigureBridgeDomain handles the creation of new bridge domain including interfaces, ARP termination
// entries and pushes status update notification.
func (plugin *BDConfigurator) ConfigureBridgeDomain(bdConfig *l2.BridgeDomains_BridgeDomain) error {
	plugin.log.Println("Configuring VPP Bridge Domain", bdConfig.Name)

	isValid, _ := plugin.vppValidateBridgeDomainBVI(bdConfig, nil)
	if !isValid {
		return nil
	}

	// Set index of the new bridge domain and increment global index.
	bdIdx := plugin.bdIDSeq
	plugin.bdIDSeq++

	// Create bridge domain with respective index.
	if err := vppcalls.VppAddBridgeDomain(bdIdx, bdConfig, plugin.vppChan, plugin.stopwatch); err != nil {
		plugin.log.Errorf("adding bridge domain %v failed: %v", bdConfig.Name, err)
		return err
	}

	// Find all interfaces belonging to this bridge domain and set them up.
	configuredIfs, err := vppcalls.SetInterfacesToBridgeDomain(bdConfig.Name, bdIdx, bdConfig.Interfaces, plugin.ifIndexes, plugin.log,
		plugin.vppChan, plugin.stopwatch)
	if err != nil {
		return err
	}

	// Resolve ARP termination table entries.
	arpTerminationTable := bdConfig.GetArpTerminationTable()
	if arpTerminationTable != nil && len(arpTerminationTable) != 0 {
		arpTable := bdConfig.ArpTerminationTable
		for _, arpEntry := range arpTable {
			err := vppcalls.VppAddArpTerminationTableEntry(bdIdx, arpEntry.PhysAddress, arpEntry.IpAddress,
				plugin.log, plugin.vppChan, plugin.stopwatch)
			if err != nil {
				plugin.log.Error(err)
			}
		}
	} else {
		plugin.log.WithField("Bridge domain name", bdConfig.Name).Debug("No ARP termination entries to set")
	}

	// Register created bridge domain.
	plugin.bdIndexes.RegisterName(bdConfig.Name, bdIdx, l2idx.NewBDMetadata(bdConfig, configuredIfs))
	plugin.log.WithFields(logging.Fields{"Name": bdConfig.Name, "Index": bdIdx}).Debug("Bridge domain registered")

	// Push to bridge domain state.
	errLookup := plugin.PropagateBdDetailsToStatus(bdIdx, bdConfig.Name)
	if errLookup != nil {
		plugin.log.WithField("bdName", bdConfig.Name).Error(errLookup)
		return errLookup
	}

	plugin.log.WithFields(logging.Fields{"bdIdx": bdIdx}).
		Infof("Bridge domain %v configured", bdConfig.Name)

	return nil
}

// ModifyBridgeDomain processes the NB config and propagates it to bin api calls.
func (plugin *BDConfigurator) ModifyBridgeDomain(newBdConfig *l2.BridgeDomains_BridgeDomain, oldBdConfig *l2.BridgeDomains_BridgeDomain) error {
	plugin.log.Infof("Modifying VPP bridge domain %v", newBdConfig.Name)

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
		plugin.log.Infof("Bridge domain %v modify done.", newBdConfig.Name)

		return nil
	}

	// Modify without recreation
	bdIdx, bdMeta, found := plugin.bdIndexes.LookupIdx(oldBdConfig.Name)
	if !found {
		// If old config is missing, the diff cannot be done. Bridge domain will be created as a new one. This
		// case should NOT happen, it means that the agent's state is inconsistent.
		plugin.log.Warnf("Bridge domain %v modify failed due to missing old configuration, will be created as a new one",
			newBdConfig.Name)
		return plugin.ConfigureBridgeDomain(newBdConfig)
	}

	// Update interfaces.
	toSet, toUnset := plugin.calculateIfaceDiff(newBdConfig.Interfaces, oldBdConfig.Interfaces)
	unConfIfs, err := vppcalls.UnsetInterfacesFromBridgeDomain(newBdConfig.Name, bdIdx, toUnset, plugin.ifIndexes, plugin.log,
		plugin.vppChan, plugin.stopwatch)
	if err != nil {
		return err
	}
	newConfIfs, err := vppcalls.SetInterfacesToBridgeDomain(newBdConfig.Name, bdIdx, toSet, plugin.ifIndexes, plugin.log,
		plugin.vppChan, plugin.stopwatch)
	if err != nil {
		return err
	}
	// Refresh configured interfaces
	configuredIfs := plugin.reckonInterfaces(bdMeta.ConfiguredInterfaces, newConfIfs, unConfIfs)

	// Update ARP termination table.
	toAdd, toRemove := plugin.calculateARPDiff(newBdConfig.ArpTerminationTable, oldBdConfig.ArpTerminationTable)
	for _, entry := range toAdd {
		vppcalls.VppAddArpTerminationTableEntry(bdIdx, entry.PhysAddress, entry.IpAddress,
			plugin.log, plugin.vppChan, plugin.stopwatch)
	}
	for _, entry := range toRemove {
		vppcalls.VppRemoveArpTerminationTableEntry(bdIdx, entry.PhysAddress, entry.IpAddress,
			plugin.log, plugin.vppChan, plugin.stopwatch)
	}

	// Push change to bridge domain state.
	errLookup := plugin.PropagateBdDetailsToStatus(bdIdx, newBdConfig.Name)
	if errLookup != nil {
		plugin.log.WithField("bdName", newBdConfig.Name).Error(errLookup)
		return errLookup
	}

	// Update bridge domain's registered metadata
	if success := plugin.bdIndexes.UpdateMetadata(newBdConfig.Name, l2idx.NewBDMetadata(newBdConfig, configuredIfs)); !success {
		plugin.log.Errorf("Failed to update metadata for bridge domain %s", newBdConfig.Name)
	}

	return nil
}

// DeleteBridgeDomain processes the NB config and propagates it to bin api calls.
func (plugin *BDConfigurator) DeleteBridgeDomain(bdConfig *l2.BridgeDomains_BridgeDomain) error {
	plugin.log.Infof("Deleting bridge domain %v", bdConfig.Name)

	bdIdx, _, found := plugin.bdIndexes.LookupIdx(bdConfig.Name)
	if !found {
		plugin.log.WithField("bdName", bdConfig.Name).Debug("Unable to find index for bridge domain to be deleted.")
		return nil
	}

	return plugin.deleteBridgeDomain(bdConfig, bdIdx)
}

func (plugin *BDConfigurator) deleteBridgeDomain(bdConfig *l2.BridgeDomains_BridgeDomain, bdIdx uint32) error {
	// Unmap all interfaces from removed bridge domain.
	if _, err := vppcalls.UnsetInterfacesFromBridgeDomain(bdConfig.Name, bdIdx, bdConfig.Interfaces, plugin.ifIndexes,
		plugin.log, plugin.vppChan, plugin.stopwatch); err != nil {
		plugin.log.Error(err) // Try to remove bridge domain anyway
	}

	if err := vppcalls.VppDeleteBridgeDomain(bdIdx, plugin.vppChan, plugin.stopwatch); err != nil {
		return err
	}

	plugin.bdIndexes.UnregisterName(bdConfig.Name)
	plugin.log.WithFields(logging.Fields{"bdIdx": bdIdx}).
		Debugf("Bridge domain %v removed", bdConfig.Name)

	// Update bridge domain state.
	if err := plugin.PropagateBdDetailsToStatus(bdIdx, bdConfig.Name); err != nil {
		return err
	}

	return nil
}

// PropagateBdDetailsToStatus looks for existing VPP bridge domain state and propagates it to the etcd bd state.
func (plugin *BDConfigurator) PropagateBdDetailsToStatus(bdID uint32, bdName string) error {
	stateMsg := BridgeDomainStateMessage{}
	var wasError error

	_, _, found := plugin.bdIndexes.LookupName(bdID)
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
	plugin.log.Infof("Assigning new interface %v to bridge domain", ifName)
	// Find bridge domain where the interface should be assigned
	bdIdx, bd, bdIf, found := plugin.bdIndexes.LookupBdForInterface(ifName)
	if !found {
		plugin.log.Debugf("Interface %s does not belong to any bridge domain", ifName)
		return nil
	}
	var bdIfs []*l2.BridgeDomains_BridgeDomain_Interfaces // Single-value
	configuredIf, err := vppcalls.SetInterfacesToBridgeDomain(bd.Name, bdIdx, append(bdIfs, bdIf), plugin.ifIndexes, plugin.log,
		plugin.vppChan, plugin.stopwatch)
	if err != nil {
		return fmt.Errorf("error while assigning interface %s to bridge domain %s", ifName, bd.Name)
	}

	// Refresh metadata
	configuredIfs, found := plugin.bdIndexes.LookupConfiguredIfsForBd(bd.Name)
	if !found {
		return fmt.Errorf("unable to get list of configured interfaces from %s", configuredIfs)
	}
	plugin.bdIndexes.UpdateMetadata(bd.Name, l2idx.NewBDMetadata(bd, append(configuredIfs, configuredIf...)))

	// Push to bridge domain state.
	if err := plugin.PropagateBdDetailsToStatus(bdIdx, bd.Name); err != nil {
		return err
	}

	return nil
}

// ResolveDeletedInterface is called by VPP if an interface is removed.
func (plugin *BDConfigurator) ResolveDeletedInterface(ifName string) error {
	plugin.log.Infof("Remove deleted interface %v from bridge domain", ifName)
	// Find bridge domain the interface should be removed from
	bdIdx, bd, _, found := plugin.bdIndexes.LookupBdForInterface(ifName)
	if !found {
		plugin.log.Debugf("Interface %s does not belong to any bridge domain", ifName)
		return nil
	}

	// If interface belonging to a bridge domain is removed, VPP handles internal bridge domain update itself.
	// However, the etcd operational state and bridge domain metadata still needs to be updated to reflect changed VPP state.
	configuredIfs, found := plugin.bdIndexes.LookupConfiguredIfsForBd(bd.Name)
	if !found {
		return fmt.Errorf("unable to get list of configured interfaces from %s", configuredIfs)
	}
	for i, configuredIf := range configuredIfs {
		if configuredIf == ifName {
			configuredIfs = append(configuredIfs[:i], configuredIfs[i+1:]...)
			break
		}
	}
	plugin.bdIndexes.UpdateMetadata(bd.Name, l2idx.NewBDMetadata(bd, configuredIfs))
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
		plugin.log.Debugf("Bridge domain %v base params changed, will be recreated", newBdConfig.Name)
	}

	if len(newBdConfig.Interfaces) == 0 {
		plugin.log.Warnf("Bridge domain %v does not contain any interface", newBdConfig.Name)
		return true, recreate
	}
	var bviCount int
	for _, bdInterface := range newBdConfig.Interfaces {
		if bdInterface.BridgedVirtualInterface {
			bviCount++
		}
	}
	if bviCount == 0 {
		plugin.log.Debugf("Bridge domain %v does not contain any bvi interface", newBdConfig.Name)
		return true, recreate
	} else if bviCount == 1 {
		return true, recreate
	} else {
		plugin.log.Warnf("Bridge domain %v contains more than one BVI interface. Correct it and create/modify bridge domain again", newBdConfig.Name)
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
	var oldBVI, newBVI *l2.BridgeDomains_BridgeDomain_Interfaces
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
	if (oldBVI == nil && newBVI != nil) || (oldBVI != nil && newBVI == nil) || (oldBVI != nil && newBVI != nil && oldBVI.Name != newBVI.Name) {
		bviChanged = true
	}

	// Resolve interfaces to unset
	for _, oldIface := range oldIfaces {
		var exists bool
		for _, newIface := range newIfaces {
			if oldIface.Name == newIface.Name && oldIface.SplitHorizonGroup == newIface.SplitHorizonGroup {
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
			if newIface.Name == oldIface.Name && newIface.SplitHorizonGroup == oldIface.SplitHorizonGroup {
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

// Recalculate configured interfaces according to output of binary API calls.
// - current is a list of interfaces present on the vpp before (read from old metadata)
// - added is a list of new configured interfaces
// - removed is a list of un-configured interfaces
// Note: resulting list of interfaces may NOT correspond with the one in bridge domain configuration.
func (plugin *BDConfigurator) reckonInterfaces(current []string, added []string, removed []string) []string {
	for _, delItem := range removed {
		for i, currItem := range current {
			if currItem == delItem {
				current = append(current[:i], current[i+1:]...)
				break
			}
		}
	}
	return append(current, added...)
}

// resolve diff of ARP entries
func (plugin *BDConfigurator) calculateARPDiff(newARPs, oldARPs []*l2.BridgeDomains_BridgeDomain_ArpTerminationEntry) (toAdd, toRemove []*l2.BridgeDomains_BridgeDomain_ArpTerminationEntry) {
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
