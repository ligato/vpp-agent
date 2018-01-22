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
//go:generate binapi-generator --input-file=/usr/share/vpp/api/vpe.api.json --output-dir=../common/bin_api

// Package l2plugin implements the L2 plugin that handles Bridge Domains and L2 FIBs.
package l2plugin

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core/bin_api/vpe"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
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
	Log           logging.Logger
	GoVppmux      govppmux.API
	BdIndexes     bdidx.BDIndexRW    // bridge domains
	IfToBdIndexes idxvpp.NameToIdxRW // interface to bridge domain mapping - desired state. Metadata is boolean flag saying whether the interface is bvi or not
	//TODO use rather BdIndexes.LookupNameByIfaceName
	IfToBdRealStateIdx     idxvpp.NameToIdxRW // interface to bridge domain mapping - current state. Metadata is boolean flag saying whether the interface is bvi or not
	BridgeDomainIDSeq      uint32
	RegisteredIfaceCounter uint32
	vppChan                *govppapi.Channel
	SwIfIndexes            ifaceidx.SwIfIndex
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
	BridgeDomainIndex uint32
	IsInterfaceBvi    bool
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
	return safeclose.Close(plugin.vppChan)
}

// ConfigureBridgeDomain for newly created bridge domain.
func (plugin *BDConfigurator) ConfigureBridgeDomain(bridgeDomainInput *l2.BridgeDomains_BridgeDomain) error {
	plugin.Log.Println("Configuring VPP Bridge Domain", bridgeDomainInput.Name)

	if !plugin.vppValidateBridgeDomainBVI(bridgeDomainInput) {
		return nil
	}

	bridgeDomainIndex := plugin.BridgeDomainIDSeq

	// Create bridge domain with respective index.
	err := vppcalls.VppAddBridgeDomain(bridgeDomainIndex, bridgeDomainInput, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(l2ba.BridgeDomainAddDel{}, plugin.Stopwatch))
	// Increment global index
	plugin.BridgeDomainIDSeq++
	if err != nil {
		plugin.Log.WithField("Bridge domain name", bridgeDomainInput.Name).Error(err)
		return err
	}

	// Find all interfaces belonging to this bridge domain and set them up.
	allInterfaces, configuredInterfaces, bviInterfaceName := vppcalls.VppSetAllInterfacesToBridgeDomain(bridgeDomainInput, bridgeDomainIndex,
		plugin.SwIfIndexes, plugin.Log, plugin.vppChan, measure.GetTimeLog(vpe.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))
	plugin.registerInterfaceToBridgeDomainPairs(allInterfaces, configuredInterfaces, bviInterfaceName, bridgeDomainIndex)

	// Resolve ARP termination table entries.
	arpTerminationTable := bridgeDomainInput.GetArpTerminationTable()
	if arpTerminationTable != nil && len(arpTerminationTable) != 0 {
		arpTable := bridgeDomainInput.ArpTerminationTable
		for _, arpEntry := range arpTable {
			err := vppcalls.VppAddArpTerminationTableEntry(bridgeDomainIndex, arpEntry.PhysAddress, arpEntry.IpAddress,
				plugin.Log, plugin.vppChan, measure.GetTimeLog(vpe.BdIPMacAddDel{}, plugin.Stopwatch))
			if err != nil {
				plugin.Log.Error(err)
			}
		}
	} else {
		plugin.Log.WithField("Bridge domain name", bridgeDomainInput.Name).Debug("No ARP termination entries to set")
	}

	// Register created bridge domain.
	plugin.BdIndexes.RegisterName(bridgeDomainInput.Name, bridgeDomainIndex, nil)
	plugin.Log.WithFields(logging.Fields{"Name": bridgeDomainInput.Name, "Index": bridgeDomainIndex}).Debug("Bridge domain registered.")

	// Push to bridge domain state.
	errLookup := plugin.LookupBridgeDomainDetails(bridgeDomainIndex, bridgeDomainInput.Name)
	if errLookup != nil {
		plugin.Log.WithField("bdName", bridgeDomainInput.Name).Error(errLookup)
		return errLookup
	}

	return nil
}

// ModifyBridgeDomain processes the NB config and propagates it to bin api calls.
func (plugin *BDConfigurator) ModifyBridgeDomain(newConfig *l2.BridgeDomains_BridgeDomain, oldConfig *l2.BridgeDomains_BridgeDomain) error {
	plugin.Log.Infof("Modifying VPP bridge domain %v", newConfig.Name)

	// Validate config.
	if !plugin.vppValidateBridgeDomainBVI(newConfig) {
		return nil
	}

	oldConfigIndex, _, found := plugin.BdIndexes.LookupIdx(oldConfig.Name)
	// During update, an old bridge domain will be removed (if exists), so unregister all interfaces at first.
	if found {
		oldInterfaces := vppcalls.VppUnsetAllInterfacesFromBridgeDomain(oldConfig, oldConfigIndex,
			plugin.SwIfIndexes, plugin.Log, plugin.vppChan, measure.GetTimeLog(vpe.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))
		plugin.unregisterInterfaceToBridgeDomainPairs(oldInterfaces)
	}

	// In case new bridge domain does not exist, create it. But this shouldn't happen.
	newConfigIndex, _, found := plugin.BdIndexes.LookupIdx(newConfig.Name)
	if !found {
		err := plugin.ConfigureBridgeDomain(newConfig)
		if err != nil {
			return err
		}
	}

	// Refresh bridge domain params. Old bridge domain will be removed if exists.
	err := vppcalls.VppUpdateBridgeDomain(oldConfigIndex, newConfigIndex, newConfig, plugin.Log, plugin.vppChan, plugin.Stopwatch)
	if err != nil {
		plugin.Log.WithField("Bridge domain name", newConfig.Name).Error(err)
		return err
	}
	plugin.Log.WithField("Bridge domain name", newConfig.Name).Debug("Bridge domain params updated.")

	// Reload interfaces for new modified bridge domain, remove any out-of-date interface
	// to BD pairs and register new ones if necessary.
	allNewInterfaces, configuredNewInterfaces, bvi := vppcalls.VppSetAllInterfacesToBridgeDomain(newConfig,
		newConfigIndex, plugin.SwIfIndexes, plugin.Log, plugin.vppChan, measure.GetTimeLog(vpe.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))
	plugin.registerInterfaceToBridgeDomainPairs(allNewInterfaces, configuredNewInterfaces, bvi, newConfigIndex)

	// Update ARP termination.
	ipMacAddDelTimeLog := measure.GetTimeLog(vpe.BdIPMacAddDel{}, plugin.Stopwatch)
	if len(newConfig.ArpTerminationTable) == 0 {
		plugin.Log.Debug("No new entries to arp termination table")
	} else if len(oldConfig.ArpTerminationTable) == 0 && len(newConfig.ArpTerminationTable) != 0 {
		arpTable := newConfig.GetArpTerminationTable()
		for _, entry := range arpTable {
			vppcalls.VppAddArpTerminationTableEntry(newConfigIndex, entry.PhysAddress, entry.IpAddress, plugin.Log,
				plugin.vppChan, ipMacAddDelTimeLog)
		}
	} else if len(oldConfig.ArpTerminationTable) != 0 {
		odlArpTable := oldConfig.GetArpTerminationTable()
		newArpTable := newConfig.GetArpTerminationTable()
		// in case old BD was not removed, delete old apr entries
		oldBdIndex, _, found := plugin.BdIndexes.LookupIdx(oldConfig.Name)
		if found {
			for _, entry := range odlArpTable {
				vppcalls.VppRemoveArpTerminationTableEntry(oldBdIndex, entry.PhysAddress, entry.IpAddress, plugin.Log,
					plugin.vppChan, ipMacAddDelTimeLog)
			}
		}
		for _, entry := range newArpTable {
			vppcalls.VppAddArpTerminationTableEntry(newConfigIndex, entry.PhysAddress, entry.IpAddress, plugin.Log,
				plugin.vppChan, ipMacAddDelTimeLog)
		}
	}

	// Push change to bridge domain state.
	errLookup := plugin.LookupBridgeDomainDetails(newConfigIndex, newConfig.Name)
	if errLookup != nil {
		plugin.Log.WithField("bdName", newConfig.Name).Error(errLookup)
		return errLookup
	}

	return nil
}

// DeleteBridgeDomain processes the NB config and propagates it to bin api calls.
func (plugin *BDConfigurator) DeleteBridgeDomain(bridgeDomain *l2.BridgeDomains_BridgeDomain) error {
	plugin.Log.Infof("'Deleting' bridge domain %v", bridgeDomain.Name)

	bdIdx, _, found := plugin.BdIndexes.LookupIdx(bridgeDomain.Name)
	if !found {
		plugin.Log.WithField("bdName", bridgeDomain.Name).Debug("Unable to find index for bridge domain to be deleted.")
		return nil
	}

	return plugin.deleteBridgeDomain(bridgeDomain, bdIdx)
}

func (plugin *BDConfigurator) deleteBridgeDomain(bridgeDomain *l2.BridgeDomains_BridgeDomain, bdIdx uint32) error {
	// Unmap all interfaces from removed bridge domain.
	interfaces := vppcalls.VppUnsetAllInterfacesFromBridgeDomain(bridgeDomain, bdIdx,
		plugin.SwIfIndexes, plugin.Log, plugin.vppChan, measure.GetTimeLog(vpe.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))
	plugin.unregisterInterfaceToBridgeDomainPairs(interfaces)

	err := vppcalls.VppDeleteBridgeDomain(bdIdx, plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.BridgeDomainAddDel{}, plugin.Stopwatch))
	if err != nil {
		return err
	}

	plugin.BdIndexes.UnregisterName(bridgeDomain.Name)
	plugin.Log.WithFields(logging.Fields{"Name": bridgeDomain.Name, "bdIdx": bdIdx}).Debug("Bridge domain removed.")

	// Push to bridge domain state.
	err = plugin.LookupBridgeDomainDetails(bdIdx, bridgeDomain.Name)
	if err != nil {
		return err
	}

	return nil
}

// LookupBridgeDomainDetails looks for existing VPP bridge domain state and propagates it to the etcd bd state.
func (plugin *BDConfigurator) LookupBridgeDomainDetails(bdID uint32, bdName string) error {
	stateMsg := BridgeDomainStateMessage{}
	var wasError error

	_, _, found := plugin.BdIndexes.LookupName(bdID)
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
func (plugin *BDConfigurator) ResolveCreatedInterface(interfaceName string, interfaceIndex uint32) error {
	plugin.Log.Infof("Resolving new interface %v", interfaceName)
	// Look whether interface belongs to some bridge domain using interface-to-bd mapping.
	_, meta, found := plugin.IfToBdIndexes.LookupIdx(interfaceName)
	if !found {
		plugin.Log.Debugf("Interface %s does not belong to any bridge domain", interfaceName)
		return nil
	}
	_, _, alreadyCreated := plugin.IfToBdRealStateIdx.LookupIdx(interfaceName)
	if alreadyCreated {
		plugin.Log.Debugf("Interface %s has been already configured", interfaceName)
		return nil
	}
	bridgeDomainIndex := meta.(*BridgeDomainMeta).BridgeDomainIndex
	bvi := meta.(*BridgeDomainMeta).IsInterfaceBvi

	vppcalls.VppSetInterfaceToBridgeDomain(bridgeDomainIndex, interfaceIndex, bvi, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(vpe.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))
	// Register interface to real state.
	plugin.IfToBdRealStateIdx.RegisterName(interfaceName, interfaceIndex, meta)

	// Push to bridge domain state.
	bridgeDomainName, _, found := plugin.BdIndexes.LookupName(bridgeDomainIndex)
	if !found {
		return fmt.Errorf("unable to update status for bridge domain, index %v not found in mapping", bridgeDomainIndex)
	}
	err := plugin.LookupBridgeDomainDetails(bridgeDomainIndex, bridgeDomainName)
	if err != nil {
		return err
	}
	return nil
}

// ResolveDeletedInterface is called by VPP if an interface is removed.
func (plugin *BDConfigurator) ResolveDeletedInterface(interfaceName string) error {
	plugin.Log.Infof("Interface %v was removed. Unregister from real state ", interfaceName)
	// Lookup IfToBdIndexes in order to find a bridge domain for this interface (if exists).
	_, meta, found := plugin.IfToBdIndexes.LookupIdx(interfaceName)
	if !found {
		plugin.Log.Debugf("Removed interface %s does not belong to any bridge domain", interfaceName)
		return nil
	}
	bdID := meta.(*BridgeDomainMeta).BridgeDomainIndex
	// Find bridge domain name.
	bdName, meta, found := plugin.BdIndexes.LookupName(bdID)
	if !found {
		return fmt.Errorf("unknown bridge domain ID %v", bdID)
	}
	// If interface belonging to a bridge domain is removed, VPP handles internal bridge domain update itself.
	// However,the etcd operational state still needs to be updated to reflect changed VPP state.
	err := plugin.LookupBridgeDomainDetails(bdID, bdName)
	if err != nil {
		return err
	}
	// Unregister removed interface from real state.
	plugin.IfToBdRealStateIdx.UnregisterName(interfaceName)

	return nil
}

// Store all interface/bridge domain pairs.
func (plugin *BDConfigurator) registerInterfaceToBridgeDomainPairs(allInterfaces []string, configuredInterfaces []string, bviIface string, domainID uint32) {
	if len(allInterfaces) == 0 {
		return
	}
	for _, iface := range allInterfaces {
		bvi := false
		if iface == bviIface {
			bvi = true
		}
		// Prepare metadata.
		meta := BridgeDomainMeta{
			BridgeDomainIndex: domainID,
			IsInterfaceBvi:    bvi,
		}
		plugin.IfToBdIndexes.RegisterName(iface, plugin.RegisteredIfaceCounter, &meta)
		plugin.Log.Debugf("Iface %v to BD %v pair registered", iface, domainID)

		// Find whether interface is configured.
		ok := false
		for _, configuredIface := range configuredInterfaces {
			if configuredIface == iface {
				ok = true
				break
			}
		}
		if ok {
			plugin.Log.Debugf("Iface %v to BD %v pair configured", iface, domainID)
			plugin.IfToBdRealStateIdx.RegisterName(iface, plugin.RegisteredIfaceCounter, &meta)
		}
		plugin.RegisteredIfaceCounter++
	}
}

// Remove all interface/bridge domain pairs from database.
func (plugin *BDConfigurator) unregisterInterfaceToBridgeDomainPairs(interfaces []string) {
	if len(interfaces) == 0 {
		return
	}
	// Unregister from desired and current state.
	for _, iface := range interfaces {
		plugin.IfToBdIndexes.UnregisterName(iface)
		plugin.IfToBdRealStateIdx.UnregisterName(iface)
		plugin.Log.WithFields(logging.Fields{"Iface": iface}).Debug("Interface to bridge domain unregistered.")
	}
}

func (plugin *BDConfigurator) vppValidateBridgeDomainBVI(bridgeDomain *l2.BridgeDomains_BridgeDomain) bool {
	if len(bridgeDomain.Interfaces) == 0 {
		plugin.Log.Warnf("Bridge domain %v does not contain any interface", bridgeDomain.Name)
		return true
	}
	var bviCount int
	for _, bdInterface := range bridgeDomain.Interfaces {
		if bdInterface.BridgedVirtualInterface {
			bviCount++
		}
	}
	if bviCount == 0 {
		plugin.Log.Debugf("Bridge domain %v does not contain any bvi interface", bridgeDomain.Name)
		return true
	} else if bviCount == 1 {
		return true
	} else {
		plugin.Log.Warnf("Bridge domain %v contains more than one BVI interface. Correct it and create/modify bridge domain again", bridgeDomain.Name)
		return false
	}
}
