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
	Log      logging.Logger
	GoVppmux govppmux.API
	// Injected mappings
	SwIfIndexes ifaceidx.SwIfIndex
	BdIndexes   bdidx.BDIndex
	// FIB-related mappings
	IfToBdIndexes   idxvpp.NameToIdxRW // TODO: use rather BdIndexes.LookupNameByIfaceName
	FibIndexes      bdidx.FIBIndexRW
	addCacheIndexes bdidx.FIBIndexRW // Serves as a cache for FIBs which cannot be configured immediately
	delCacheIndexes bdidx.FIBIndexRW // Serves as a cache for FIBs which cannot be removed immediately
	fibIndexSeq     uint32

	syncVppChannel  *govppapi.Channel
	asyncVppChannel *govppapi.Channel
	vppcalls        *vppcalls.L2FibVppCalls

	Stopwatch *measure.Stopwatch // timer used to measure and store time
}

// Init goroutines, mappings, channels..
func (plugin *FIBConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing L2 Bridge domains")

	// Init local mapping
	plugin.addCacheIndexes = bdidx.NewFIBIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "l2plugin", ""+
		"fib_add_indexes", nil))
	plugin.delCacheIndexes = bdidx.NewFIBIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "l2plugin", ""+
		"fib_del_indexes", nil))
	plugin.fibIndexSeq = 1

	// Init 2 VPP API channels to separate synchronous and asynchronous communication
	plugin.syncVppChannel, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}
	plugin.asyncVppChannel, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	if err := plugin.syncVppChannel.CheckMessageCompatibility(vppcalls.L2FibMessages...); err != nil {
		return err
	}

	plugin.vppcalls = vppcalls.NewL2FibVppCalls(plugin.Log, plugin.asyncVppChannel, plugin.Stopwatch)
	go plugin.vppcalls.WatchFIBReplies()

	return nil
}

// Close vpp channel.
func (plugin *FIBConfigurator) Close() error {
	_, err := safeclose.CloseAll(plugin.syncVppChannel, plugin.asyncVppChannel)
	return err
}

// Add configures provided FIB input. Every entry has to contain info about MAC address, interface, and bridge domain.
// If interface or bridge domain is missing, FIB data is cached and recalled if particular entity is registered.
func (plugin *FIBConfigurator) Add(fib *l2.FibTable_FibEntry, callback func(error)) error {
	plugin.Log.Infof("Configuring new FIB table entry with MAC %v", fib.PhysAddress)

	if fib.PhysAddress == "" {
		return fmt.Errorf("no mac address in FIB entry %s", fib)
	}
	if fib.BridgeDomain == "" {
		return fmt.Errorf("no bridge domain in FIB entry %s", fib)
	}

	// Remove FIB from (del) cache if it's there
	_, _, exists := plugin.delCacheIndexes.UnregisterName(fib.PhysAddress)
	if exists {
		plugin.Log.Debugf("FIB entry %s was removed from (del) cache before configuration")
	}

	// Validate required items and move to (add) cache if something's missing
	cached, ifIdx, bdIdx := plugin.validateFibRequirements(fib, true)
	if cached {
		return nil
	}
	plugin.Log.Debugf("Configuring FIB entry %s for bridge domain %s and interface %s", fib.PhysAddress, bdIdx, ifIdx)

	return plugin.vppcalls.Add(fib.PhysAddress, bdIdx, ifIdx, fib.BridgedVirtualInterface, fib.StaticConfig,
		func(err error) {
			// Register
			plugin.FibIndexes.RegisterName(fib.PhysAddress, plugin.fibIndexSeq, fib)
			plugin.Log.Debugf("Fib entry with MAC %v registered", fib.PhysAddress)
			plugin.fibIndexSeq++
			callback(err)
		})
}

// Modify provides changes for FIB entry. Old fib entry is removed (if possible) and a new one is registered
// if all the conditions are fulfilled (interface and bridge domain presence), otherwise new configuration is cached.
func (plugin *FIBConfigurator) Modify(oldFib *l2.FibTable_FibEntry,
	newFib *l2.FibTable_FibEntry, callback func(error)) error {
	plugin.Log.Infof("Modifying FIB table entry with MAC %s", newFib.PhysAddress)

	// Remove FIB from (add) cache if present
	_, _, exists := plugin.addCacheIndexes.UnregisterName(oldFib.PhysAddress)
	if exists {
		plugin.Log.Debugf("Modified FIB %s removed from (add) cache", oldFib.PhysAddress)
	}

	// Remove an old entry if possible
	oldIfIdx, _, ifFound := plugin.SwIfIndexes.LookupIdx(oldFib.OutgoingInterface)
	if !ifFound {
		plugin.Log.Debugf("FIB %s cannot be removed now, interface %s no longer exists",
			oldFib.PhysAddress, oldFib.OutgoingInterface)
	} else {
		oldBdIdx, _, bdFound := plugin.BdIndexes.LookupIdx(oldFib.BridgeDomain)
		if !bdFound {
			plugin.Log.Debugf("FIB %s cannot be removed, bridge domain %s no longer exists",
				oldFib.PhysAddress, oldFib.BridgeDomain)
		} else {
			if err := plugin.vppcalls.Delete(oldFib.PhysAddress, oldBdIdx, oldIfIdx, func(err error) {
				plugin.FibIndexes.UnregisterName(oldFib.PhysAddress)
				plugin.addCacheIndexes.UnregisterName(oldFib.PhysAddress)
				callback(err)
			}); err != nil {
				// Log error but continue
				plugin.Log.Errorf("FIB modify: failed to remove entry %s", oldFib.PhysAddress)
			}
		}
	}

	cached, ifIdx, bdIdx := plugin.validateFibRequirements(newFib, true)
	if cached {
		return nil
	}

	return plugin.vppcalls.Add(newFib.PhysAddress, bdIdx, ifIdx, newFib.BridgedVirtualInterface, newFib.StaticConfig,
		func(err error) {
			plugin.FibIndexes.RegisterName(oldFib.PhysAddress, plugin.fibIndexSeq, newFib)
			plugin.fibIndexSeq++
			callback(err)
		})
}

// Delete removes FIB table entry. The request to be successful, both interface and bridge domain indices
// have to be available. Request does nothing without this info. If interface (or bridge domain) was removed before,
// provided FIB data is just unregistered and agent assumes, that VPP removed FIB entry itself.
func (plugin *FIBConfigurator) Delete(fib *l2.FibTable_FibEntry, callback func(error)) error {
	plugin.Log.Infof("Deleting FIB table entry with MAC %s", fib.PhysAddress)

	// Check if FIB is in cache (add). In such a case, just remove it.
	_, _, exists := plugin.addCacheIndexes.UnregisterName(fib.PhysAddress)
	if exists {
		return nil
	}

	// Check whether the FIB can be actually removed
	cached, ifIdx, bdIdx := plugin.validateFibRequirements(fib, false)
	if cached {
		return nil
	}

	// Unregister from (del) cache and from indexes
	plugin.delCacheIndexes.UnregisterName(fib.PhysAddress)
	plugin.FibIndexes.UnregisterName(fib.PhysAddress)
	plugin.Log.Debugf("FIB %s removed from mappings", fib.PhysAddress)

	return plugin.vppcalls.Delete(fib.PhysAddress, bdIdx, ifIdx, func(err error) {
		callback(err)
	})
}

// ResolveCreatedInterface uses FIB cache to additionally configure any FIB entries for this interface. Bridge domain
// is checked for existence. If resolution is successful, new FIB entry is configured, registered and removed from cache.
func (plugin *FIBConfigurator) ResolveCreatedInterface(ifName string, ifIdx uint32, callback func(error)) error {
	plugin.Log.Infof("FIB configurator: resolving registered interface %s", ifName)

	var wasErr error
	// First, remove FIBs which cannot be removed due to missing interface
	for _, cachedFibId := range plugin.delCacheIndexes.GetMapping().ListNames() {
		_, meta, found := plugin.delCacheIndexes.LookupIdx(cachedFibId)
		if !found || meta == nil {
			// Should not happen
			continue
		}
		// Check interface
		if ifName != meta.OutgoingInterface {
			// New interface is not suitable for this FIB entry
			continue
		}
		if err := plugin.Delete(meta, func(err error) {
			plugin.Log.Debugf("Deleting obsolete FIB %s", cachedFibId)
			callback(err)
		}); err != nil {
			plugin.Log.Error(err)
			wasErr = err
		}
	}

	// Configure un-configurable FIBs
	for _, cachedFibId := range plugin.addCacheIndexes.GetMapping().ListNames() {
		_, meta, found := plugin.addCacheIndexes.LookupIdx(cachedFibId)
		if !found || meta == nil {
			// Should not happen
			continue
		}
		// Check interface
		if ifName != meta.OutgoingInterface {
			// New interface is not suitable for this FIB entry
			continue
		}

		// Check bridge domain presence
		bdIdx, _, bdFound := plugin.BdIndexes.LookupIdx(meta.BridgeDomain)
		if !bdFound {
			plugin.Log.Debugf("FIB %s still cannot be configured due to missing bridge domain %s",
				cachedFibId, meta.BridgeDomain)
			continue
		}

		if err := plugin.vppcalls.Add(cachedFibId, bdIdx, ifIdx, meta.BridgedVirtualInterface, meta.StaticConfig, func(err error) {
			plugin.Log.Infof("Configuring cached bridge domain %s", cachedFibId)
			// Handle registration
			plugin.addCacheIndexes.UnregisterName(cachedFibId)
			plugin.Log.Debugf("FIB %s removed from cache", cachedFibId)
			plugin.FibIndexes.RegisterName(cachedFibId, plugin.fibIndexSeq, meta)
			plugin.fibIndexSeq++
			callback(err)
		}); err != nil {
			plugin.Log.Error(err)
			wasErr = err
		}
	}
	plugin.Log.Infof("FIB: resolution of created interface %s is done", ifName)
	return wasErr
}

// ResolveDeletedInterface handles removed interface. In that case, FIB entry remains on the VPP but it is not possible
// to delete it.
func (plugin *FIBConfigurator) ResolveDeletedInterface(ifName string, ifIdx uint32, callback func(error)) error {
	plugin.Log.Infof("FIB configurator: resolving unregistered interface %s", ifName)

	var counter int
	for _, fib := range plugin.FibIndexes.GetMapping().ListNames() {
		_, meta, found := plugin.FibIndexes.LookupIdx(fib)
		if !found || meta == nil {
			// Should not happen
			continue
		}
		// Check interface
		if ifName != meta.OutgoingInterface {
			continue
		}
		counter++
	}

	plugin.Log.Infof("%d FIB entries belongs to removed interface %s. These FIBs cannot be deleted or changed while interface is missing",
		counter, ifName)

	return nil
}

// ResolveCreatedBridgeDomain uses FIB cache to configure any FIB entries for this bridge domain.
// Required interface is checked for existence. If resolution is successful, new FIB entry is configured,
// registered and removed from cache.
func (plugin *FIBConfigurator) ResolveCreatedBridgeDomain(bdName string, bdID uint32, callback func(error)) error {
	plugin.Log.Infof("FIB configurator: resolving registered bridge domain %s", bdName)

	var wasErr error
	// First, remove FIBs which cannot be removed due to missing interface
	for _, cachedFibId := range plugin.delCacheIndexes.GetMapping().ListNames() {
		_, meta, found := plugin.delCacheIndexes.LookupIdx(cachedFibId)
		if !found || meta == nil {
			// Should not happen
			continue
		}
		// Check interface
		if bdName != meta.BridgeDomain {
			// New bridge domain is not suitable for this FIB entry
			continue
		}
		if err := plugin.Delete(meta, func(err error) {
			plugin.Log.Debugf("Deleting obsolete FIB %s", cachedFibId)
			callback(err)
		}); err != nil {
			plugin.Log.Error(err)
			wasErr = err
		}
	}

	// Configure un-configurable FIBs
	for _, cachedFibId := range plugin.addCacheIndexes.GetMapping().ListNames() {
		_, meta, found := plugin.addCacheIndexes.LookupIdx(cachedFibId)
		if !found || meta == nil {
			// Should not happen
			continue
		}
		// Check bridge domain
		if bdName != meta.BridgeDomain {
			// New bridge domain is not suitable for this FIB entry
			continue
		}

		// Check interface presence
		ifIdx, _, ifFound := plugin.SwIfIndexes.LookupIdx(meta.OutgoingInterface)
		if !ifFound {
			plugin.Log.Debugf("FIB %s still cannot be configured due to missing interface %s",
				cachedFibId, meta.BridgeDomain)
			continue
		}

		if err := plugin.vppcalls.Add(cachedFibId, bdID, ifIdx, meta.BridgedVirtualInterface, meta.StaticConfig, func(err error) {
			plugin.Log.Infof("Configuring cached bridge domain %s", cachedFibId)
			// Handle registration
			plugin.addCacheIndexes.UnregisterName(cachedFibId)
			plugin.Log.Debugf("FIB %s removed from cache", cachedFibId)
			plugin.FibIndexes.RegisterName(cachedFibId, plugin.fibIndexSeq, meta)
			plugin.fibIndexSeq++
			callback(err)
		}); err != nil {
			plugin.Log.Error(err)
			wasErr = err
		}
	}
	plugin.Log.Infof("FIB: resolution of created bridge domain %s is done", bdName)
	return wasErr
}

// ResolveDeletedInterface handles removed bridge domain. In that case, FIB entry remains on the VPP but it is not possible
// to delete it.
func (plugin *FIBConfigurator) ResolveDeletedBridgeDomain(bdName string, bdID uint32, callback func(error)) error {
	plugin.Log.Infof("FIB configurator: resolving unregistered bridge domain %s", bdName)

	var counter int
	for _, fib := range plugin.FibIndexes.GetMapping().ListNames() {
		_, meta, found := plugin.FibIndexes.LookupIdx(fib)
		if !found || meta == nil {
			// Should not happen
			continue
		}
		// Check bridge domain
		if bdName != meta.BridgeDomain {
			continue
		}

		counter++
	}

	plugin.Log.Infof("%d FIB entries belongs to removed bridge domain %s. These FIBs cannot be deleted or changed while bridge domain is missing",
		counter, bdName)

	return nil
}

func (plugin *FIBConfigurator) validateFibRequirements(fib *l2.FibTable_FibEntry, add bool) (cached bool, ifIdx, bdIdx uint32) {
	// Check bridge domain presence
	var ifFound, bdFound bool
	bdIdx, _, bdFound = plugin.BdIndexes.LookupIdx(fib.BridgeDomain)
	if !bdFound {
		plugin.Log.Infof("FIB entry %s is configured for bridge domain %s which does not exists",
			fib.PhysAddress, fib.BridgeDomain)
	}

	// Check interface presence
	ifIdx, _, ifFound = plugin.SwIfIndexes.LookupIdx(fib.OutgoingInterface)
	if !ifFound {
		plugin.Log.Infof("FIB entry %s is configured for interface %s which does not exists",
			fib.PhysAddress, fib.OutgoingInterface)
	}

	// If either interface or bridge domain is missing, cache FIB entry
	if !bdFound || !ifFound {
		if add {
			// FIB table entry is cached and will be configured again when all required items are available
			_, _, found := plugin.addCacheIndexes.LookupIdx(fib.PhysAddress)
			if !found {
				plugin.addCacheIndexes.RegisterName(fib.PhysAddress, plugin.fibIndexSeq, fib)
				plugin.Log.Debugf("FIB entry with name %s added to cache (add)", fib.PhysAddress)
				plugin.fibIndexSeq++
			} else {
				plugin.addCacheIndexes.UpdateMetadata(fib.PhysAddress, fib)
			}
		} else {
			// FIB table entry is cached and will be removed again when all required items are available
			_, _, found := plugin.delCacheIndexes.LookupIdx(fib.PhysAddress)
			if !found {
				plugin.delCacheIndexes.RegisterName(fib.PhysAddress, plugin.fibIndexSeq, fib)
				plugin.Log.Debugf("FIB entry with name %s added to cache (del)", fib.PhysAddress)
				plugin.fibIndexSeq++
			} else {
				plugin.delCacheIndexes.UpdateMetadata(fib.PhysAddress, fib)
			}
		}
		cached = true
	}

	return
}
