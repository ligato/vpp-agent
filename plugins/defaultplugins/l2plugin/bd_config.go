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
	Log      logging.Logger
	GoVppmux govppmux.API

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
	_, err := safeclose.CloseAll(plugin.vppChan, plugin.notificationChan)
	return err
}

// ConfigureBridgeDomain for newly created bridge domain.
func (plugin *BDConfigurator) ConfigureBridgeDomain(bridgeDomain *l2.BridgeDomains_BridgeDomain) error {
	plugin.Log.Println("Configuring VPP Bridge Domain", bridgeDomain.Name)

	if !plugin.vppValidateBridgeDomainBVI(bridgeDomain) {
		return nil
	}

	// Index of the new bridge domain and increment global index
	bridgeDomainIndex := plugin.BridgeDomainIDSeq
	plugin.BridgeDomainIDSeq++

	// Create bridge domain with respective index.
	err := vppcalls.VppAddBridgeDomain(bridgeDomainIndex, bridgeDomain, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(l2ba.BridgeDomainAddDel{}, plugin.Stopwatch))

	if err != nil {
		plugin.Log.WithField("Bridge domain name", bridgeDomain.Name).Error(err)
		return err
	}

	// Find all interfaces belonging to this bridge domain and set them up.
	vppcalls.VppSetAllInterfacesToBridgeDomain(bridgeDomain, bridgeDomainIndex,
		plugin.SwIfIndices, plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))

	// Resolve ARP termination table entries.
	arpTerminationTable := bridgeDomain.GetArpTerminationTable()
	if arpTerminationTable != nil && len(arpTerminationTable) != 0 {
		arpTable := bridgeDomain.ArpTerminationTable
		for _, arpEntry := range arpTable {
			err := vppcalls.VppAddArpTerminationTableEntry(bridgeDomainIndex, arpEntry.PhysAddress, arpEntry.IpAddress,
				plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.BdIPMacAddDel{}, plugin.Stopwatch))
			if err != nil {
				plugin.Log.Error(err)
			}
		}
	} else {
		plugin.Log.WithField("Bridge domain name", bridgeDomain.Name).Debug("No ARP termination entries to set")
	}

	// Register created bridge domain.
	plugin.BdIndexes.RegisterName(bridgeDomainInput.Name, bridgeDomainIndex, nil)
	plugin.Log.WithFields(logging.Fields{"Name": bridgeDomainInput.Name, "Index": bridgeDomainIndex}).Debug("Bridge domain registered.")

	// Push to bridge domain state.
	errLookup := plugin.LookupBridgeDomainDetails(bridgeDomainIndex, bridgeDomain.Name)
	if errLookup != nil {
		plugin.Log.WithField("bdName", bridgeDomain.Name).Error(errLookup)
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

	oldConfigIndex, _, found := plugin.BdIndices.LookupIdx(oldConfig.Name)
	// During update, an old bridge domain will be removed (if exists), so unregister all interfaces at first.
	if found {
		vppcalls.VppUnsetAllInterfacesFromBridgeDomain(oldConfig, oldConfigIndex,
			plugin.SwIfIndices, plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))
	}

	// In case new bridge domain does not exist, create it. But this shouldn't happen.
	newConfigIndex, _, found := plugin.BdIndices.LookupIdx(newConfig.Name)
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
	vppcalls.VppSetAllInterfacesToBridgeDomain(newConfig,
		newConfigIndex, plugin.SwIfIndices, plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))

	// Update ARP termination.
	ipMacAddDelTimeLog := measure.GetTimeLog(l2ba.BdIPMacAddDel{}, plugin.Stopwatch)
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
		oldBdIndex, _, found := plugin.BdIndices.LookupIdx(oldConfig.Name)
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

	bdIdx, _, found := plugin.BdIndices.LookupIdx(bridgeDomain.Name)
	if !found {
		plugin.Log.WithField("bdName", bridgeDomain.Name).Debug("Unable to find index for bridge domain to be deleted.")
		return nil
	}

	return plugin.deleteBridgeDomain(bridgeDomain, bdIdx)
}

func (plugin *BDConfigurator) deleteBridgeDomain(bridgeDomain *l2.BridgeDomains_BridgeDomain, bdIdx uint32) error {
	// Unmap all interfaces from removed bridge domain.
	vppcalls.VppUnsetAllInterfacesFromBridgeDomain(bridgeDomain, bdIdx,
		plugin.SwIfIndices, plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))

	err := vppcalls.VppDeleteBridgeDomain(bdIdx, plugin.Log, plugin.vppChan, measure.GetTimeLog(l2ba.BridgeDomainAddDel{}, plugin.Stopwatch))
	if err != nil {
		return err
	}

	plugin.BdIndices.UnregisterName(bridgeDomain.Name)
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

	vppcalls.VppSetInterfaceToBridgeDomain(bdIdx, ifIdx, bvi, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(l2ba.SwInterfaceSetL2Bridge{}, plugin.Stopwatch))

	// Push to bridge domain state.
	err := plugin.LookupBridgeDomainDetails(bdIdx, bd.Name)
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
	err := plugin.LookupBridgeDomainDetails(bdIdx, bd.Name)
	if err != nil {
		return err
	}

	return nil
}

// The goal of the validation is to ensure that bridge domain does not contain more than one BVI interface
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
