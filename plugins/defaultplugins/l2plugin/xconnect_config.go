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
	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// XConnectConfigurator implements PluginHandlerVPP
type XConnectConfigurator struct {
	GoVppmux    *govppmux.GOVPPPlugin
	SwIfIndexes ifaceidx.SwIfIndex
	XcIndexes   idxvpp.NameToIdxRW
	XcIndexSeq  uint32
	vppChan     *govppapi.Channel
}

// XConnectMeta meta hold info about transmit interface
type XConnectMeta struct {
	TransmitInterface string
}

// Init members (channels...) and start go routines
func (plugin *XConnectConfigurator) Init() (err error) {

	log.Debug("Initializing L2 xConnect")

	// Init VPP API channel
	plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	err = vppcalls.CheckMsgCompatibilityForL2XConnect(plugin.vppChan)
	if err != nil {
		return err
	}

	return nil
}

// Close GOVPP channel
func (plugin *XConnectConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// ConfigureXConnectPair process the NB config and propagates it to bin api calls
func (plugin *XConnectConfigurator) ConfigureXConnectPair(xConnectPairInput *l2.XConnectPairs_XConnectPair) error {
	log.Println("Configuring L2 xConnect pair", xConnectPairInput.ReceiveInterface)

	// Find interfaces
	receiveInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(xConnectPairInput.ReceiveInterface)
	if !found {
		log.WithField("Interface", xConnectPairInput.ReceiveInterface).Warn("Receive interface not found.")
		return nil
	}
	transmitInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(xConnectPairInput.TransmitInterface)
	if !found {
		log.WithField("Interface", xConnectPairInput.TransmitInterface).Warn("Transmit interface not found.")
		return nil
	}

	err := vppcalls.VppSetL2XConnect(receiveInterfaceIndex, transmitInterfaceIndex, plugin.vppChan)
	if err != nil {
		log.WithField("Error", err).Error("Failed to create l2xConnect")
		return err
	}

	// Prepare meta
	meta := XConnectMeta{
		TransmitInterface: xConnectPairInput.TransmitInterface,
	}

	// Register
	plugin.XcIndexes.RegisterName(xConnectPairInput.ReceiveInterface, plugin.XcIndexSeq, meta)
	plugin.XcIndexSeq++

	return nil
}

// ModifyXConnectPair process the NB config and propagates it to bin api calls
func (plugin *XConnectConfigurator) ModifyXConnectPair(newConfig *l2.XConnectPairs_XConnectPair, oldConfig *l2.XConnectPairs_XConnectPair) error {
	log.Println("Modifying L2 xConnect pair")

	// interfaces
	receiveInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(newConfig.ReceiveInterface)
	if !found {
		log.WithField("Interface", newConfig.ReceiveInterface).Error("Receive interface not found.")
		return nil
	}
	newTransmitInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(newConfig.TransmitInterface)
	if !found {
		log.WithField("Interface", newConfig.TransmitInterface).Error("New transmit interface not found.")
		return nil
	}
	oldTransmitInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(oldConfig.TransmitInterface)
	if !found {
		log.WithField("Interface", newConfig.TransmitInterface).Debug("Old transmit interface not found.")
		oldTransmitInterfaceIndex = 0
		// do not return, not an error
	}
	if oldTransmitInterfaceIndex == newTransmitInterfaceIndex {
		// nothing to update
		return nil
	} else if oldTransmitInterfaceIndex == 0 {
		// create new xConnect only
		err := vppcalls.VppSetL2XConnect(receiveInterfaceIndex, newTransmitInterfaceIndex, plugin.vppChan)
		if err != nil {
			log.WithField("Error", err).Error("Failed to set l2xConnect")
			return err
		}
	} else {
		errDel := vppcalls.VppUnsetL2XConnect(receiveInterfaceIndex, oldTransmitInterfaceIndex, plugin.vppChan)
		if errDel != nil {
			log.WithField("Error", errDel).Error("Failed to remove l2xConnect")
			return errDel
		}
		errCreate := vppcalls.VppSetL2XConnect(receiveInterfaceIndex, newTransmitInterfaceIndex, plugin.vppChan)
		if errCreate != nil {
			log.WithField("Error", errCreate).Error("Failed to set l2xConnect")
			return errCreate
		}
	}

	meta := XConnectMeta{
		TransmitInterface: newConfig.TransmitInterface,
	}
	plugin.XcIndexes.RegisterName(newConfig.ReceiveInterface, receiveInterfaceIndex, meta)
	plugin.XcIndexSeq++

	return nil
}

// DeleteXConnectPair process the NB config and propagates it to bin api calls
func (plugin *XConnectConfigurator) DeleteXConnectPair(xConnectPairInput *l2.XConnectPairs_XConnectPair) error {
	log.Println("Deleting L2 xConnect pair", xConnectPairInput.ReceiveInterface)

	// Find interfaces
	receiveInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(xConnectPairInput.ReceiveInterface)
	if !found {
		log.WithField("Interface", xConnectPairInput.ReceiveInterface).Warn("Receive interface not found.")
		return nil
	}
	transmitInterfaceIndex, _, found := plugin.SwIfIndexes.LookupIdx(xConnectPairInput.TransmitInterface)
	if !found {
		log.WithField("Interface", xConnectPairInput.TransmitInterface).Warn("Transmit interface not found.")
		return nil
	}

	err := vppcalls.VppUnsetL2XConnect(receiveInterfaceIndex, transmitInterfaceIndex, plugin.vppChan)
	if err != nil {
		log.WithField("Error", err).Error("Failed to remove l2xConnect")
		return err
	}

	// Unregister
	plugin.XcIndexes.UnregisterName(xConnectPairInput.ReceiveInterface)
	log.WithFields(log.Fields{"RecIface": xConnectPairInput.ReceiveInterface, "Idex": receiveInterfaceIndex}).Debug("XConnect pair unregistered.")

	return nil
}

// LookupXConnectPairs registers missing l2 xConnect pairs
func (plugin *XConnectConfigurator) LookupXConnectPairs() error {
	req := &l2ba.L2XconnectDump{}
	reqContext := plugin.vppChan.SendMultiRequest(req)
	var index uint32 = 1
	for {
		msg := &l2ba.L2XconnectDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			log.Error(err)
			return err
		}
		if stop {
			break
		}
		// Store name if missing
		_, _, found := plugin.XcIndexes.LookupName(index)
		xcIdentifier := string(msg.RxSwIfIndex)
		if !found {
			log.WithFields(log.Fields{"Name": xcIdentifier, "Index": index}).Debug("L2 xConnect pair registered.")
			plugin.XcIndexes.RegisterName(xcIdentifier, index, nil)
		} else {
			log.WithFields(log.Fields{"Name": xcIdentifier, "Index": index}).Debug("L2 xConnect pair already registered.")
		}
		index++
	}

	return nil
}
