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
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/l2idx"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l2"
)

// XConnectConfigurator implements PluginHandlerVPP.
type XConnectConfigurator struct {
	log logging.Logger
	// Interface indexes
	ifIndexes ifaceidx.SwIfIndex
	// Cross connect indexes
	xcIndexes         l2idx.XcIndexRW
	xcAddCacheIndexes l2idx.XcIndexRW
	xcDelCacheIndexes l2idx.XcIndexRW
	xcIndexSeq        uint32

	vppChan   *govppapi.Channel
	stopwatch *measure.Stopwatch // Timer used to measure and store time
}

// Init essential configurator fields.
func (plugin *XConnectConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, swIfIndexes ifaceidx.SwIfIndex,
	enableStopwatch bool) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-xc-conf")
	plugin.log.Info("Initializing L2 xConnect configurator")

	// Mappings
	plugin.ifIndexes = swIfIndexes
	plugin.xcIndexes = l2idx.NewXcIndex(nametoidx.NewNameToIdx(plugin.log, "xc-indexes", nil))
	plugin.xcAddCacheIndexes = l2idx.NewXcIndex(nametoidx.NewNameToIdx(plugin.log, "xc-add-cache-indexes", nil))
	plugin.xcDelCacheIndexes = l2idx.NewXcIndex(nametoidx.NewNameToIdx(plugin.log, "xc-del-cache-indexes", nil))
	plugin.xcIndexSeq = 1

	// VPP channel
	plugin.vppChan, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}

	// Stopwatch
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("BFDConfigurator", plugin.log)
	}

	// Message compatibility
	if err = plugin.vppChan.CheckMessageCompatibility(vppcalls.XConnectMessages...); err != nil {
		plugin.log.Error(err)
		return err
	}

	return nil
}

// Close govpp channel.
func (plugin *XConnectConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// GetXcIndexes returns cross connect memory indexes
func (plugin *XConnectConfigurator) GetXcIndexes() l2idx.XcIndexRW {
	return plugin.xcIndexes
}

// GetXcAddCache returns cross connect 'add' cache (test purposes)
func (plugin *XConnectConfigurator) GetXcAddCache() l2idx.XcIndexRW {
	return plugin.xcAddCacheIndexes
}

// GetXcDelCache returns cross connect 'del' cache (test purposes)
func (plugin *XConnectConfigurator) GetXcDelCache() l2idx.XcIndexRW {
	return plugin.xcDelCacheIndexes
}

// ConfigureXConnectPair adds new cross connect pair
func (plugin *XConnectConfigurator) ConfigureXConnectPair(xc *l2.XConnectPairs_XConnectPair) error {
	plugin.log.Infof("Configuring L2 xConnect pair %s-%s", xc.ReceiveInterface, xc.TransmitInterface)
	if err := plugin.validateConfig(xc); err != nil {
		return err
	}
	// Verify interface presence, eventually store cross connect to cache if either is missing
	rxIfIdx, _, rxFound := plugin.ifIndexes.LookupIdx(xc.ReceiveInterface)
	if !rxFound {
		plugin.log.Debugf("XC Add: Receive interface %s not found.", xc.ReceiveInterface)
	}
	txIfIdx, _, txFound := plugin.ifIndexes.LookupIdx(xc.TransmitInterface)
	if !txFound {
		plugin.log.Debugf("XC Add: Transmit interface %s not found.", xc.TransmitInterface)
	}
	if !rxFound || !txFound {
		plugin.putOrUpdateCache(xc, true)
		return nil
	}
	// XConnect can be configured now
	if err := vppcalls.AddL2XConnect(rxIfIdx, txIfIdx, plugin.vppChan, plugin.stopwatch); err != nil {
		plugin.log.Errorf("Adding l2xConnect failed: %v", err)
		return err
	}
	// Unregister from 'del' cache in case it is present
	plugin.xcDelCacheIndexes.UnregisterName(xc.ReceiveInterface)
	// Register
	plugin.xcIndexes.RegisterName(xc.ReceiveInterface, plugin.xcIndexSeq, xc)
	plugin.xcIndexSeq++
	plugin.log.Infof("L2 xConnect pair %s-%s configured", xc.ReceiveInterface, xc.TransmitInterface)

	return nil
}

// ModifyXConnectPair modifies cross connect pair (its transmit interface). Old entry is replaced.
func (plugin *XConnectConfigurator) ModifyXConnectPair(newXc, oldXc *l2.XConnectPairs_XConnectPair) error {
	plugin.log.Infof("Modifying L2 xConnect pair %s-%s", oldXc.ReceiveInterface, oldXc.TransmitInterface)
	if err := plugin.validateConfig(newXc); err != nil {
		return err
	}
	// Verify receive interface presence
	rxIfIdx, _, rxFound := plugin.ifIndexes.LookupIdx(newXc.ReceiveInterface)
	if !rxFound {
		plugin.log.Debugf("XC Modify: Receive interface %s not found.", newXc.ReceiveInterface)
		plugin.putOrUpdateCache(newXc, true)
		plugin.xcIndexes.UnregisterName(oldXc.ReceiveInterface)
		// Can return, without receive interface the entry cannot exist
		return nil
	}
	// Verify transmit interface
	txIfIdx, _, txFound := plugin.ifIndexes.LookupIdx(newXc.TransmitInterface)
	if !txFound {
		plugin.log.Debugf("XC Modify: Transmit interface %s not found.", newXc.TransmitInterface)
		plugin.putOrUpdateCache(newXc, true)
		// If new transmit interface is missing and XConnect cannot be updated now, configurator can try to remove old
		// entry, so the VPP output won't be confusing
		oldTxIfIdx, _, oldTxFound := plugin.ifIndexes.LookupIdx(oldXc.TransmitInterface)
		if !oldTxFound {
			return nil // Nothing more can be done
		}
		plugin.log.Debugf("Removing obsolete l2xConnect %s-%s", oldXc.ReceiveInterface, oldXc.TransmitInterface)
		if err := vppcalls.DeleteL2XConnect(rxIfIdx, oldTxIfIdx, plugin.vppChan, plugin.stopwatch); err != nil {
			plugin.log.Errorf("Deleted obsolete l2xConnect failed: %v", err)
			return err
		}
		plugin.xcIndexes.UnregisterName(oldXc.ReceiveInterface)
		return nil
	}
	// Replace existing entry
	if err := vppcalls.AddL2XConnect(rxIfIdx, txIfIdx, plugin.vppChan, plugin.stopwatch); err != nil {
		plugin.log.Errorf("Replacing l2xConnect failed: %v", err)
		return err
	}
	plugin.xcIndexes.RegisterName(newXc.ReceiveInterface, plugin.xcIndexSeq, newXc)
	plugin.xcIndexSeq++
	plugin.log.Debugf("Modifying XConnect: new entry %s-%s added", newXc.ReceiveInterface, newXc.TransmitInterface)

	return nil
}

// DeleteXConnectPair removes XConnect if possible. Note: Xconnect pair cannot be removed if any interface is missing.
func (plugin *XConnectConfigurator) DeleteXConnectPair(xc *l2.XConnectPairs_XConnectPair) error {
	plugin.log.Infof("Removing L2 xConnect pair %s-%s", xc.ReceiveInterface, xc.TransmitInterface)
	if err := plugin.validateConfig(xc); err != nil {
		return err
	}
	// If receive interface is missing, XConnect entry is not configured on the VPP.
	rxIfIdx, _, rxFound := plugin.ifIndexes.LookupIdx(xc.ReceiveInterface)
	if !rxFound {
		plugin.log.Debugf("XC Del: Receive interface %s not found.", xc.ReceiveInterface)
		// Remove from all caches
		plugin.xcIndexes.UnregisterName(xc.ReceiveInterface)
		plugin.xcAddCacheIndexes.UnregisterName(xc.ReceiveInterface)
		plugin.xcDelCacheIndexes.UnregisterName(xc.ReceiveInterface)
		return nil
	}
	// Verify transmit interface. If it is missing, XConnect cannot be removed and will be put to cache for deleted
	// interfaces
	txIfIdx, _, txFound := plugin.ifIndexes.LookupIdx(xc.TransmitInterface)
	if !txFound {
		plugin.log.Debugf("XC Del: Transmit interface %s for XConnect %s not found.",
			xc.TransmitInterface, xc.ReceiveInterface)
		plugin.putOrUpdateCache(xc, false)
		// Remove from other caches
		plugin.xcIndexes.UnregisterName(xc.ReceiveInterface)
		plugin.xcAddCacheIndexes.UnregisterName(xc.ReceiveInterface)
		return nil
	}
	// XConnect can be removed now
	if err := vppcalls.DeleteL2XConnect(rxIfIdx, txIfIdx, plugin.vppChan, plugin.stopwatch); err != nil {
		plugin.log.Errorf("Removing l2xConnect failed: %v", err)
		return err
	}
	// Unregister
	plugin.xcIndexes.UnregisterName(xc.ReceiveInterface)
	plugin.log.Infof("L2 xConnect pair %s-%s removed", xc.ReceiveInterface, xc.TransmitInterface)

	return nil
}

// ResolveCreatedInterface resolves XConnects waiting for an interface.
func (plugin *XConnectConfigurator) ResolveCreatedInterface(ifName string) error {
	plugin.log.Debugf("XConnect configurator: resolving created interface %s", ifName)
	var wasErr error
	// XConnects waiting to be configured
	for _, xcRxIf := range plugin.xcAddCacheIndexes.GetMapping().ListNames() {
		_, xc, _ := plugin.xcAddCacheIndexes.LookupIdx(xcRxIf)
		if xc == nil {
			plugin.log.Errorf("XConnect entry %s has no metadata", xcRxIf)
			continue
		}
		if xc.TransmitInterface == ifName || xc.ReceiveInterface == ifName {
			plugin.xcAddCacheIndexes.UnregisterName(xcRxIf)
			if err := plugin.ConfigureXConnectPair(xc); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
		}
	}
	// XConnects waiting for removal
	for _, xcRxIf := range plugin.xcDelCacheIndexes.GetMapping().ListNames() {
		_, xc, _ := plugin.xcDelCacheIndexes.LookupIdx(xcRxIf)
		if xc == nil {
			plugin.log.Errorf("XConnect entry %s has no metadata", xcRxIf)
			continue
		}
		if xc.TransmitInterface == ifName || xc.ReceiveInterface == ifName {
			plugin.xcDelCacheIndexes.UnregisterName(xcRxIf)
			if err := plugin.DeleteXConnectPair(xc); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
		}
	}

	return wasErr
}

// ResolveDeletedInterface resolves XConnects using deleted interface
// If deleted interface is a received interface, the XConnect was removed by the VPP
// If deleted interface is a transmit interface, it will get flag 'DELETED' in VPP, but the entry will be kept
func (plugin *XConnectConfigurator) ResolveDeletedInterface(ifName string) error {
	plugin.log.Debugf("XConnect configurator: resolving deleted interface %s", ifName)
	for _, xcRxIf := range plugin.xcIndexes.GetMapping().ListNames() {
		_, xc, _ := plugin.xcIndexes.LookupIdx(xcRxIf)
		if xc == nil {
			plugin.log.Errorf("XConnect entry %s has no metadata", xcRxIf)
			continue
		}
		if xc.ReceiveInterface == ifName {
			plugin.xcIndexes.UnregisterName(xc.ReceiveInterface)
			plugin.xcAddCacheIndexes.RegisterName(xc.ReceiveInterface, plugin.xcIndexSeq, xc)
			plugin.xcIndexSeq++
			continue
		}
		// Nothing to do for transmit
	}

	return nil
}

// Add XConnect to 'add' or 'del' cache, or just update metadata
func (plugin *XConnectConfigurator) putOrUpdateCache(xc *l2.XConnectPairs_XConnectPair, cacheTypeAdd bool) {
	if cacheTypeAdd {
		if _, _, found := plugin.xcAddCacheIndexes.LookupIdx(xc.ReceiveInterface); found {
			plugin.xcAddCacheIndexes.UpdateMetadata(xc.ReceiveInterface, xc)
		} else {
			plugin.xcAddCacheIndexes.RegisterName(xc.ReceiveInterface, plugin.xcIndexSeq, xc)
			plugin.xcIndexSeq++
		}
	} else {
		if _, _, found := plugin.xcDelCacheIndexes.LookupIdx(xc.ReceiveInterface); found {
			plugin.xcDelCacheIndexes.UpdateMetadata(xc.ReceiveInterface, xc)
		} else {
			plugin.xcDelCacheIndexes.RegisterName(xc.ReceiveInterface, plugin.xcIndexSeq, xc)
			plugin.xcIndexSeq++
		}
	}
}

func (plugin *XConnectConfigurator) validateConfig(xc *l2.XConnectPairs_XConnectPair) error {
	if xc.ReceiveInterface == "" {
		return fmt.Errorf("invalid XConnect configuration, receive interface is not set")
	}
	if xc.TransmitInterface == "" {
		return fmt.Errorf("invalid XConnect configuration, transmit interface is not set")
	}
	if xc.ReceiveInterface == xc.TransmitInterface {
		return fmt.Errorf("invalid XConnect configuration, recevie interface is the same as transmit (%s)",
			xc.ReceiveInterface)
	}
	return nil
}
