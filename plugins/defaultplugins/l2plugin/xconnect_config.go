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

	"time"

	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core/bin_api/vpe"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

const (
	transmitInterfaceKey = "TransmitInterface"
)

// XConnectConfigurator implements PluginHandlerVPP.
type XConnectConfigurator struct {
	Log         logging.Logger
	GoVppmux    govppmux.API
	SwIfIndexes ifaceidx.SwIfIndex
	XcIndexes   idxvpp.NameToIdxRW
	XcIndexSeq  uint32
	vppChan     *govppapi.Channel
	Stopwatch   *measure.Stopwatch // timer used to measure and store time
}

// XConnectMeta meta holds info about transmit interface.
type XConnectMeta struct {
	TransmitInterface string
	configured        bool // true if already configured on VPP, false if still waiting for creation of the rx or tx interface
}

// Init members (channels...) and start go routines.
func (plugin *XConnectConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing L2 xConnect configurator")

	// Init VPP API channel.
	plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	// Check bin api message compatibility.
	err = vppcalls.CheckMsgCompatibilityForL2XConnect(plugin.Log, plugin.vppChan)
	if err != nil {
		return err
	}

	plugin.XcIndexes = nametoidx.NewNameToIdx(plugin.Log, "l2plugin", "xconnect", func(meta interface{}) map[string][]string {
		res := map[string][]string{}
		if xc, ok := meta.(*XConnectMeta); ok {
			res[transmitInterfaceKey] = []string{xc.TransmitInterface}
		}
		return res
	})

	return nil
}

// Close GOVPP channel.
func (plugin *XConnectConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// ConfigureXConnectPair processes the NB config and propagates it to bin api calls.
func (plugin *XConnectConfigurator) ConfigureXConnectPair(xConnectPairInput *l2.XConnectPairs_XConnectPair) error {
	plugin.Log.Infof("Configuring L2 xConnect pair %v", xConnectPairInput.ReceiveInterface)

	// Find interfaces.
	receiveInterfaceIndex, _, rxFound := plugin.SwIfIndexes.LookupIdx(xConnectPairInput.ReceiveInterface)
	if !rxFound {
		plugin.Log.WithField("Interface", xConnectPairInput.ReceiveInterface).Warn("Receive interface not found.")
	}
	transmitInterfaceIndex, _, txFound := plugin.SwIfIndexes.LookupIdx(xConnectPairInput.TransmitInterface)
	if !txFound {
		plugin.Log.WithField("Interface", xConnectPairInput.TransmitInterface).Warn("Transmit interface not found.")
	}

	if rxFound && txFound {
		// can be configured now
		err := vppcalls.VppSetL2XConnect(receiveInterfaceIndex, transmitInterfaceIndex, plugin.Log, plugin.vppChan,
			measure.GetTimeLog(vpe.SwInterfaceSetL2Xconnect{}, plugin.Stopwatch))
		if err != nil {
			plugin.Log.WithField("Error", err).Error("Failed to create l2xConnect")
			return err
		}
	} else {
		plugin.Log.Warn("rx or tx interface not found, l2xconnect postponed")
	}

	// Prepare meta.
	meta := XConnectMeta{
		TransmitInterface: xConnectPairInput.TransmitInterface,
		configured:        rxFound && txFound,
	}

	// Register.
	plugin.XcIndexes.RegisterName(xConnectPairInput.ReceiveInterface, plugin.XcIndexSeq, &meta)
	plugin.XcIndexSeq++

	plugin.Log.Infof("L2 xConnect pair %v configured", xConnectPairInput.ReceiveInterface)

	return nil
}

// ModifyXConnectPair processes the NB config and propagates it to bin api calls.
func (plugin *XConnectConfigurator) ModifyXConnectPair(newConfig *l2.XConnectPairs_XConnectPair, oldConfig *l2.XConnectPairs_XConnectPair) error {
	plugin.Log.Infof("Modifying L2 xConnect pair %v %v", oldConfig)

	// interfaces
	receiveInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(newConfig.ReceiveInterface)
	if !found {
		plugin.Log.WithField("Interface", newConfig.ReceiveInterface).Error("Receive interface not found.")
		return nil
	}
	newTransmitInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(newConfig.TransmitInterface)
	if !found {
		plugin.Log.WithField("Interface", newConfig.TransmitInterface).Error("New transmit interface not found.")
		return nil
	}
	oldTransmitInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(oldConfig.TransmitInterface)
	if !found {
		plugin.Log.WithField("Interface", newConfig.TransmitInterface).Debug("Old transmit interface not found.")
		oldTransmitInterfaceIndex = 0
		// do not return, not an error
	}
	if oldTransmitInterfaceIndex == newTransmitInterfaceIndex {
		// nothing to update
		return nil
	} else if oldTransmitInterfaceIndex == 0 {
		// Create new xConnect only.
		err := vppcalls.VppSetL2XConnect(receiveInterfaceIndex, newTransmitInterfaceIndex, plugin.Log,
			plugin.vppChan, measure.GetTimeLog(vpe.SwInterfaceSetL2Xconnect{}, plugin.Stopwatch))
		if err != nil {
			plugin.Log.WithField("Error", err).Error("Failed to set l2xConnect")
			return err
		}
	} else {
		errDel := vppcalls.VppUnsetL2XConnect(receiveInterfaceIndex, oldTransmitInterfaceIndex, plugin.Log,
			plugin.vppChan, measure.GetTimeLog(vpe.SwInterfaceSetL2Xconnect{}, plugin.Stopwatch))
		if errDel != nil {
			plugin.Log.WithField("Error", errDel).Error("Failed to remove l2xConnect")
			return errDel
		}
		errCreate := vppcalls.VppSetL2XConnect(receiveInterfaceIndex, newTransmitInterfaceIndex, plugin.Log,
			plugin.vppChan, measure.GetTimeLog(vpe.SwInterfaceSetL2Xconnect{}, plugin.Stopwatch))
		if errCreate != nil {
			plugin.Log.WithField("Error", errCreate).Error("Failed to set l2xConnect")
			return errCreate
		}
	}

	meta := XConnectMeta{
		TransmitInterface: newConfig.TransmitInterface,
	}
	plugin.XcIndexes.RegisterName(newConfig.ReceiveInterface, receiveInterfaceIndex, &meta)
	plugin.XcIndexSeq++

	plugin.Log.Infof("L2 xConnect pair %v modified", newConfig.ReceiveInterface)

	return nil
}

// DeleteXConnectPair processes the NB config and propagates it to bin api calls.
func (plugin *XConnectConfigurator) DeleteXConnectPair(xConnectPairInput *l2.XConnectPairs_XConnectPair) error {
	plugin.Log.Infof("Removing L2 xConnect pair %v", xConnectPairInput.ReceiveInterface)

	if err := plugin.deleteL2XConnectPair(xConnectPairInput.ReceiveInterface, xConnectPairInput.TransmitInterface); err != nil {
		return err
	}

	plugin.Log.Infof("L2 xConnect pair %v removed", xConnectPairInput.ReceiveInterface)

	return nil
}

// LookupXConnectPairs registers missing l2 xConnect pairs.
func (plugin *XConnectConfigurator) LookupXConnectPairs() error {
	// L2XconnectDump time measurement
	start := time.Now()
	defer func() {
		if plugin.Stopwatch != nil {
			timeLog := measure.GetTimeLog(l2ba.L2XconnectDump{}, plugin.Stopwatch)
			if timeLog != nil {
				timeLog.LogTimeEntry(time.Since(start))
			}
		}
	}()

	req := &l2ba.L2XconnectDump{}
	reqContext := plugin.vppChan.SendMultiRequest(req)
	var index uint32 = 1
	for {
		msg := &l2ba.L2XconnectDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			plugin.Log.Error(err)
			return err
		}
		if stop {
			break
		}
		// Store name if missing.
		_, _, found := plugin.XcIndexes.LookupName(index)
		xcIdentifier := string(msg.RxSwIfIndex)
		if !found {
			plugin.Log.WithFields(logging.Fields{"Name": xcIdentifier, "Index": index}).Debug("L2 xConnect pair registered.")
			plugin.XcIndexes.RegisterName(xcIdentifier, index, nil)
		} else {
			plugin.Log.WithFields(logging.Fields{"Name": xcIdentifier, "Index": index}).Debug("L2 xConnect pair already registered.")
		}
		index++
	}

	return nil
}

// ResolveCreatedInterface configures xconnect pairs that use the interface as rx or tx and have not been configured yet.
func (plugin *XConnectConfigurator) ResolveCreatedInterface(interfaceName string, interfaceIndex uint32) error {
	plugin.Log.Infof("XConnect configurator: resolving created interface %v", interfaceName)
	var err error

	// Lookup for the interface in rx interfaces.
	err = plugin.resolveRxInterface(interfaceName, true)

	// Lookup for the interface in tx interfaces.
	rxIfs := plugin.XcIndexes.LookupNameByMetadata(transmitInterfaceKey, interfaceName)
	for _, rxIf := range rxIfs {
		err = plugin.resolveRxInterface(rxIf, true)
	}

	return err
}

// ResolveDeletedInterface deletes xconnect pairs that have not been deleted yet and use the interface as rx or tx.
func (plugin *XConnectConfigurator) ResolveDeletedInterface(interfaceName string) error {
	plugin.Log.Infof("XConnect configurator: resolving deleted interface %v", interfaceName)

	var err error

	// Lookup for the interface in rx interfaces.
	err = plugin.resolveRxInterface(interfaceName, false)

	// Lookup for the interface in tx interfaces.
	rxIfs := plugin.XcIndexes.LookupNameByMetadata(transmitInterfaceKey, interfaceName)
	for _, rxIf := range rxIfs {
		err = plugin.resolveRxInterface(rxIf, false)
	}

	return err
}

// resolveRxInterface creates/deletes 2sxconnect in XcIndexes that has not been created yet/need to be deleted
// and have the rx interface name matching with the provided argument.
func (plugin *XConnectConfigurator) resolveRxInterface(rxIfName string, create bool) error {
	var err error

	idx, meta, exists := plugin.XcIndexes.LookupIdx(rxIfName)
	if exists {
		meta := meta.(*XConnectMeta)
		if create {
			// The l2xconn needs to be created.
			if !meta.configured {
				// not yet configured, try to configure now
				err = plugin.configureL2XConnectPair(rxIfName, meta.TransmitInterface)
				if err == nil {
					// mark as configured and save
					meta.configured = true
					plugin.XcIndexes.RegisterName(rxIfName, idx, meta)
				}
			}
		} else {
			// The l2xconn needs to be deleted.
			err = plugin.deleteL2XConnectPair(rxIfName, meta.TransmitInterface)
			// meta deleted in deleteL2XConnectPair
		}
	}

	return err
}

func (plugin *XConnectConfigurator) configureL2XConnectPair(rxIf, txIf string) error {
	plugin.Log.Infof("Configuring L2 xConnect pair %v %v", rxIf, txIf)

	// Find interface idx-es.
	receiveInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(rxIf)
	if !found {
		plugin.Log.WithField("Interface", rxIf).Warn("Receive interface not found.")
		return fmt.Errorf("receive interface '%s' not found", rxIf)
	}
	transmitInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(txIf)
	if !found {
		plugin.Log.WithField("Interface", txIf).Warn("Transmit interface not found.")
		return fmt.Errorf("transmit interface '%s' not found", txIf)
	}

	// Configure l2xconnect.
	err := vppcalls.VppSetL2XConnect(receiveInterfaceIndex, transmitInterfaceIndex, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(vpe.SwInterfaceSetL2Xconnect{}, plugin.Stopwatch))
	if err != nil {
		plugin.Log.WithField("Error", err).Error("Failed to create l2xConnect")
		return err
	}

	return nil
}

func (plugin *XConnectConfigurator) deleteL2XConnectPair(rxIf, txIf string) error {
	plugin.Log.Infof("Deleting L2 xConnect pair %v %v", rxIf, txIf)

	// Find interface idx-es.
	receiveInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(rxIf)
	if !found {
		plugin.Log.WithField("Interface", rxIf).Warn("Receive interface not found.")
		return fmt.Errorf("receive interface '%s' not found", rxIf)
	}
	transmitInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(txIf)
	if !found {
		plugin.Log.WithField("Interface", txIf).Warn("Transmit interface not found.")
		return fmt.Errorf("transmit interface '%s' not found", txIf)
	}

	err := vppcalls.VppUnsetL2XConnect(receiveInterfaceIndex, transmitInterfaceIndex, plugin.Log, plugin.vppChan,
		measure.GetTimeLog(vpe.SwInterfaceSetL2Xconnect{}, plugin.Stopwatch))
	if err != nil {
		plugin.Log.WithField("Error", err).Error("Failed to remove l2xConnect")
		return err
	}

	// Unregister.
	plugin.XcIndexes.UnregisterName(rxIf)
	plugin.Log.WithFields(logging.Fields{"RecIface": rxIf, "Idex": receiveInterfaceIndex}).Debug("XConnect pair unregistered.")

	return nil
}
