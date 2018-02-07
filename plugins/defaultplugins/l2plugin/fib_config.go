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

package l2plugin

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bdidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// FIBConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of fib table entries as modelled by the proto file "../model/l2/l2.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/bd/<bd-label>/fib".
// Updates received from the northbound API are compared with the VPP run-time configuration
// and differences are applied through the VPP binary API.
type FIBConfigurator struct {
	Log             logging.Logger
	GoVppmux        govppmux.API
	SwIfIndexes     ifaceidx.SwIfIndex
	BdIndexes       bdidx.BDIndex
	IfToBdIndexes   idxvpp.NameToIdxRW //TODO use rather BdIndexes.LookupNameByIfaceName
	FibIndexes      idxvpp.NameToIdxRW
	FibIndexSeq     uint32
	FibDesIndexes   idxvpp.NameToIdxRW // Serves as a cache for FIBs which cannot be configured immediately
	syncVppChannel  *govppapi.Channel
	asyncVppChannel *govppapi.Channel
	vppcalls        *vppcalls.L2FibVppCalls
	Stopwatch       *measure.Stopwatch // timer used to measure and store time
}

// FIBMeta metadata holder holds information about entry interface and bridge domain.
type FIBMeta struct {
	InterfaceName    string
	BridgeDomainName string
	BVI              bool
	StaticConfig     bool
}

// Init goroutines, mappings, channels, ...
func (plugin *FIBConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing L2 Bridge domains")

	// Init local mapping.
	plugin.FibDesIndexes = nametoidx.NewNameToIdx(logrus.DefaultLogger(), "l2plugin", "fib_des_indexes", nil)

	// Init 2 VPP API channels to separate synchronous and asynchronous communication.
	plugin.syncVppChannel, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}
	plugin.asyncVppChannel, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	err = vppcalls.CheckMsgCompatibilityForL2FIB(plugin.Log, plugin.syncVppChannel)
	if err != nil {
		return err
	}

	plugin.vppcalls = vppcalls.NewL2FibVppCalls(plugin.asyncVppChannel, plugin.Stopwatch)
	go plugin.vppcalls.WatchFIBReplies(plugin.Log)

	return nil
}

// Close vpp channel.
func (plugin *FIBConfigurator) Close() error {
	_, err := safeclose.CloseAll(plugin.syncVppChannel, plugin.asyncVppChannel)
	return err
}

// Add configures provided FIB input. Every entry has to contain info about MAC address, and interface, and
// bridge domain. If interface or bridge domain is missing, FIB data is cached and recalled if particular entity is registered.
func (plugin *FIBConfigurator) Add(fib *l2.FibTableEntries_FibTableEntry, callback func(error)) error {
	plugin.Log.Infof("Configuring new FIB table entry with MAC %v", fib.PhysAddress)

	if fib.PhysAddress == "" {
		return fmt.Errorf("no mac address in FIB entry %v", fib)
	}
	if fib.BridgeDomain == "" {
		return fmt.Errorf("no bridge domain in FIB entry %v", fib)
	}
	// Prepare meta.
	meta := &FIBMeta{fib.OutgoingInterface, fib.BridgeDomain, fib.BridgedVirtualInterface, fib.StaticConfig}

	// Check bridge domain presence.
	bdIndex, _, bdFound := plugin.BdIndexes.LookupIdx(fib.BridgeDomain)
	if !bdFound {
		plugin.Log.Infof("FIB entry %v is configured for bridge domain %v which does not exists", fib.PhysAddress, fib.BridgeDomain)
	}
	// Check interface presence.
	ifIndex, _, ifFound := plugin.SwIfIndexes.LookupIdx(fib.OutgoingInterface)
	if !ifFound {
		plugin.Log.Infof("FIB entry %v is configured for interface %v which does not exists", fib.PhysAddress, fib.OutgoingInterface)
	}
	// If either interface or bridge domain is missing, cache FIB table to nc_fib_indexes.
	if !bdFound || !ifFound {
		// FIB table entry is cached and will be configured when all required configuration is available.
		plugin.FibDesIndexes.RegisterName(fib.PhysAddress, plugin.FibIndexSeq, meta)
		plugin.Log.Debugf("Uncofigured FIB entry with name %v added to cache", fib.PhysAddress)
		plugin.FibIndexSeq++
		return nil
	}

	plugin.Log.Debugf("Configuring FIB entry %v for bridge domain %v and interface %v", fib.PhysAddress, bdIndex, ifIndex)

	return plugin.vppcalls.Add(fib.PhysAddress, bdIndex, ifIndex, fib.BridgedVirtualInterface,
		fib.StaticConfig, func(err error) {
			// Register.
			plugin.FibIndexes.RegisterName(fib.PhysAddress, plugin.FibIndexSeq, meta)
			plugin.Log.Debugf("Fib entry with MAC %v registered", fib.PhysAddress)
			plugin.FibIndexSeq++
			callback(err)
		}, plugin.Log)
}

// Diff provides changes for FIB entry. Old fib entry is removed (if possible) and a new one is registered
// if all the conditions are fulfilled (interface and bridge domain presence), otherwise new configuration is cached.
func (plugin *FIBConfigurator) Diff(oldFib *l2.FibTableEntries_FibTableEntry,
	newFib *l2.FibTableEntries_FibTableEntry, callback func(error)) error {
	plugin.Log.Infof("Modifying FIB table entry with MAC ", newFib.PhysAddress)

	// Remove an old entry if necessary.
	oldIfIndex, _, ifaceFound := plugin.SwIfIndexes.LookupIdx(oldFib.OutgoingInterface)
	if !ifaceFound {
		return fmt.Errorf("FIB %v cannot be removed, interface %v does not exist",
			oldFib.PhysAddress, oldFib.OutgoingInterface)
	}
	oldBdIndex, _, bdFound := plugin.BdIndexes.LookupIdx(oldFib.BridgeDomain)
	if !bdFound {
		return fmt.Errorf("FIB %v cannot be removed, bridge domain %v does not exist",
			oldFib.PhysAddress, oldFib.BridgeDomain)
	}
	err := plugin.vppcalls.Delete(oldFib.PhysAddress, oldBdIndex, oldIfIndex, func(err error) {
		plugin.FibIndexes.UnregisterName(oldFib.PhysAddress)
		plugin.FibDesIndexes.UnregisterName(oldFib.PhysAddress)
		callback(err)
	}, plugin.Log)
	if err != nil {
		return err
	}

	// Prepare Meta.
	meta := &FIBMeta{newFib.OutgoingInterface, newFib.BridgeDomain, newFib.BridgedVirtualInterface, newFib.StaticConfig}

	// Check bridge domain presence.
	newBdIndex, _, bdFound := plugin.BdIndexes.LookupIdx(newFib.BridgeDomain)
	if !bdFound {
		plugin.Log.Infof("FIB entry %v is configured for bridge domain %v which does not exists", newFib.PhysAddress, newFib.BridgeDomain)
	}
	// Check interface presence.
	newIfIndex, _, ifFound := plugin.SwIfIndexes.LookupIdx(newFib.OutgoingInterface)
	if !ifFound {
		plugin.Log.Infof("FIB entry %v is configured for interface %v which does not exists", newFib.PhysAddress, newFib.OutgoingInterface)
	}
	if !bdFound || !ifFound {
		plugin.FibDesIndexes.RegisterName(newFib.PhysAddress, plugin.FibIndexSeq, meta)
		plugin.Log.Debugf("uncofigured FIB entry with name %v added to cache", newFib.PhysAddress)
		plugin.FibIndexSeq++
		return nil
	}

	return plugin.vppcalls.Add(newFib.PhysAddress, newBdIndex, newIfIndex, newFib.BridgedVirtualInterface,
		newFib.StaticConfig, func(err error) {
			plugin.FibIndexes.RegisterName(oldFib.PhysAddress, plugin.FibIndexSeq, meta)
			plugin.FibIndexSeq++
			callback(err)
		}, plugin.Log)
}

// Delete removes FIB table entry. The request to be successful, both interface and bridge domain indices
// have to be available. Request does nothing without this info. If interface (or bridge domain) was removed before,
// provided FIB data is just unregistered and agent assumes, that VPP removed FIB entry itself.
func (plugin *FIBConfigurator) Delete(fib *l2.FibTableEntries_FibTableEntry, callback func(error)) error {
	plugin.Log.Infof("Deleting FIB table entry with MAC ", fib.PhysAddress)

	// Remove not configured FIB from cache if exists.
	plugin.FibDesIndexes.UnregisterName(fib.PhysAddress)
	plugin.Log.Debugf("Uncofigured Fib entry with name %v removed from cache", fib.PhysAddress)
	// Unregister.
	plugin.FibIndexes.UnregisterName(fib.PhysAddress)
	plugin.Log.Debugf("FIB entry with name %v unregistered", fib.PhysAddress)

	ifIndex, _, ifaceFound := plugin.SwIfIndexes.LookupIdx(fib.OutgoingInterface)
	if !ifaceFound {
		return fmt.Errorf("FIB %v cannot be removed, interface %v does not exist",
			fib.PhysAddress, fib.OutgoingInterface)
	}
	bdIndex, _, bdFound := plugin.BdIndexes.LookupIdx(fib.BridgeDomain)
	if !bdFound {
		return fmt.Errorf("FIB %v cannot be removed, bridge domain %v does not exist",
			fib.PhysAddress, fib.BridgeDomain)
	}

	return plugin.vppcalls.Delete(fib.PhysAddress, bdIndex, ifIndex, func(err error) {
		callback(err)
	}, plugin.Log)
}

// ResolveCreatedInterface uses FIB cache to additionally configure any FIB entries for this interface. Bridge domain
// is checked for existence. If resolution is successful, new FIB entry is configured, registered and removed from cache.
func (plugin *FIBConfigurator) ResolveCreatedInterface(interfaceName string, interfaceIndex uint32,
	callback func(error)) error {
	plugin.Log.Infof("Resolve new interface %v from FIB perspective ", interfaceName)
	firstIndex := 1
	lastIndex := plugin.FibIndexSeq - 1 // Number of all registered FIB Indexes
	var wasError error
	for index := uint32(firstIndex); index <= lastIndex; index++ {
		mac, meta, found := plugin.FibDesIndexes.LookupName(index)
		if found {
			// Check interface.
			fibInterface := meta.(*FIBMeta).InterfaceName
			if interfaceName != fibInterface {
				continue
			}
			// Check bridge domain.
			fibBridgeDomain := meta.(*FIBMeta).BridgeDomainName
			bdIndex, _, bdFound := plugin.BdIndexes.LookupIdx(fibBridgeDomain)
			// Validate interface/bridge domain pair.
			validated := plugin.validateInterfaceBDPair(interfaceName, bdIndex)
			if !bdFound || !validated {
				plugin.Log.Infof("FIB entry %v - required bridge domain not found", mac)
				continue
			} else if !validated {
				plugin.Log.Infof("FIB entry %v - bridge domain %v does not contain interface %v",
					mac, bdIndex, interfaceName)
				continue
			} else {
				fibBvi := meta.(*FIBMeta).BVI
				fibStatic := meta.(*FIBMeta).StaticConfig
				err := plugin.vppcalls.Add(mac, bdIndex, interfaceIndex, fibBvi,
					fibStatic, func(err error) {
						plugin.Log.WithField("Mac", mac).
							Infof("Previously not configurable FIB entry with is now configured")
						// Resolve registration.
						plugin.FibIndexes.RegisterName(mac, plugin.FibIndexSeq, meta)
						plugin.FibIndexSeq++
						plugin.Log.Debugf("Registering FIB entry with MAC %v", mac)
						plugin.FibDesIndexes.UnregisterName(mac)
						plugin.Log.WithField("Mac", mac).
							Debugf("Uncofigured FIB entry removed from cache")
						callback(err)
					}, plugin.Log)
				if err != nil {
					wasError = err
				}
			}
		}
	}
	plugin.Log.Infof("FIB: resolution of created interface %v is done", interfaceName)
	return wasError
}

// ResolveDeletedInterface if interface was deleted. All FIB entries belonging to this interface are removed from
// configuration and added to FIB cache (from Agent perspective, FIB entry is not removed when interface is removed).
func (plugin *FIBConfigurator) ResolveDeletedInterface(interfaceName string, interfaceIndex uint32,
	callback func(error)) error {
	plugin.Log.Infof("Resolve removed interface %v from FIB perspective ", interfaceName)
	firstIndex := 1
	lastIndex := plugin.FibIndexSeq - 1 // Number of all registered FIB Indexes
	var wasError error
	for index := uint32(firstIndex); index <= lastIndex; index++ {
		mac, meta, found := plugin.FibIndexes.LookupName(index)
		if found {
			// Check interface.
			fibInterface := meta.(*FIBMeta).InterfaceName
			if interfaceName != fibInterface {
				continue
			}
			// Check bridge domain.
			fibBridgeDomain := meta.(*FIBMeta).BridgeDomainName
			bdIndex, _, bdFound := plugin.BdIndexes.LookupIdx(fibBridgeDomain)
			if !bdFound {
				wasError = fmt.Errorf("bridge domain configured for FIB no longer exists, unable to remove FIB for interface %v", interfaceName)
			} else {
				err := plugin.vppcalls.Delete(mac, bdIndex, interfaceIndex, func(err error) {
					// Resolve registration.
					plugin.FibIndexes.UnregisterName(mac)
					plugin.Log.Debugf("Unregister FIB entry with MAC %v", mac)
					plugin.FibDesIndexes.RegisterName(mac, plugin.FibIndexSeq, meta)
					plugin.FibIndexSeq++
					plugin.Log.Debugf("uncofigured FIB entry with MAC %v added to cache", mac)
					callback(err)
				}, plugin.Log)
				if err != nil {
					wasError = err
				}
			}
		}
	}
	plugin.Log.Infof("FIB: resolution of removed interface %v is done", interfaceName)
	return wasError
}

// ResolveCreatedBridgeDomain uses FIB cache to additionally configure any FIB entries
// for this bridge domain. Required interface is checked for existence. If resolution
// is successful, new FIB entry is configured, registered and removed from cache.
func (plugin *FIBConfigurator) ResolveCreatedBridgeDomain(domainName string, domainID uint32, callback func(error)) error {
	plugin.Log.Infof("Resolve created bridge domain %v from FIB perspective ", domainID)
	firstIndex := 1
	lastIndex := plugin.FibIndexSeq - 1 // Number of all registered FIB Indexes
	var wasError error
	for index := uint32(firstIndex); index <= lastIndex; index++ {
		mac, meta, found := plugin.FibDesIndexes.LookupName(index)
		if found {
			// Check interface.
			fibInterface := meta.(*FIBMeta).InterfaceName
			ifIndex, _, ifFound := plugin.SwIfIndexes.LookupIdx(fibInterface)
			// Validate interface/bridge domain pair.
			validated := plugin.validateInterfaceBDPair(fibInterface, domainID)
			if !ifFound {
				plugin.Log.Infof("FIB entry %v - required interface %v not found", mac, fibInterface)
				continue
			}
			if !validated {
				plugin.Log.Infof("FIB entry %v - required interface %v is not a part of bridge domain %v",
					mac, fibInterface, domainID)
				continue
			} else {
				fibBvi := meta.(*FIBMeta).BVI
				fibStatic := meta.(*FIBMeta).StaticConfig
				err := plugin.vppcalls.Add(mac, domainID, ifIndex, fibBvi, fibStatic, func(err error) {
					plugin.Log.Debugf("Previously not configurable FIB entry with MAC %v is now configured", mac)
					// Resolve registration.
					plugin.FibIndexes.RegisterName(mac, plugin.FibIndexSeq, meta)
					plugin.FibIndexSeq++
					plugin.Log.Debugf("Registering FIB table entry with MAC %v", mac)
					plugin.FibDesIndexes.UnregisterName(mac)
					plugin.Log.Debugf("Unconfigured FIB entry with MAC %v removed from cache", mac)
					callback(err)
				}, plugin.Log)
				if err != nil {
					wasError = err
				}
			}
		}
	}
	plugin.Log.Debugf("FIB: resolution of created bridge domain %v is done", domainName)
	return wasError
}

// ResolveDeletedBridgeDomain if BD was deleted. All FIB entries belonging to this bridge domain are removed from
// configuration and added to FIB cache (from Agent perspective, FIB entry is not removed when bridge domain vanishes).
func (plugin *FIBConfigurator) ResolveDeletedBridgeDomain(domainName string, domainID uint32, callback func(error)) error {
	plugin.Log.Infof("Resolve removed bridge domain %v from FIB perspective ", domainID)
	firstIndex := 1
	lastIndex := plugin.FibIndexSeq - 1 // Number of all registered FIB Indexes
	var wasError error
	for index := uint32(firstIndex); index <= lastIndex; index++ {
		mac, meta, found := plugin.FibIndexes.LookupName(index)
		if found {
			// Check bridge domain.
			fibBridgeDomain := meta.(*FIBMeta).BridgeDomainName
			if domainName != fibBridgeDomain {
				continue
			}
			// Check interface.
			fibInterface := meta.(*FIBMeta).InterfaceName
			ifIndex, _, ifFound := plugin.SwIfIndexes.LookupIdx(fibInterface)
			if !ifFound {
				wasError = fmt.Errorf("interface configured for FIB no longer exists, unable to remove FIB for bridge domain %v", domainName)
			} else {
				err := plugin.vppcalls.Delete(mac, domainID, ifIndex, func(err error) {
					// Resolve registration.
					plugin.FibIndexes.UnregisterName(mac)
					plugin.Log.Debug("Unregister FIB table entry with MAC ", mac)
					plugin.FibDesIndexes.UnregisterName(mac) // if exists
					plugin.Log.Debugf("uncofigured FIB entry with MAC %v removed from cache", mac)
					callback(err)
				}, plugin.Log)
				if err != nil {
					wasError = err
				}
			}
		}
	}
	plugin.Log.Infof("FIB: resolution of removed bridge domain %v is done", domainName)
	return wasError
}

// Verify that interface is assigned to bridge domain.
func (plugin *FIBConfigurator) validateInterfaceBDPair(interfaceName string, bridgeDomainIndex uint32) bool {
	_, meta, found := plugin.IfToBdIndexes.LookupIdx(interfaceName)
	if !found {
		plugin.Log.Debugf("FIB validation - Interface %v not registered as a pair with any bridge domain", interfaceName)
		return false
	}
	if meta == nil {
		plugin.Log.Errorf("Interface %v registered as a pair with bridge domain but no meta found", interfaceName)
		return false
	}
	wantedIndex := meta.(*BridgeDomainMeta).bdIdx
	if bridgeDomainIndex == wantedIndex {
		return true
	}
	return false
}
