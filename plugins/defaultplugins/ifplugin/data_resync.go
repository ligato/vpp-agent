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

package ifplugin

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/idxvpp/persist"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppdump"
)

// Resync writes interfaces to the empty VPP
//
// - resyncs the VPP
// - temporary: (checks wether sw_if_indexes are not obsolate - this will be swapped with master ID)
// - deletes obsolate status data
func (plugin *InterfaceConfigurator) Resync(nbIfaces []*intf.Interfaces_Interface) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC Interface begin.")
	// Calculate and log interface resync
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Step 0: Dump actual state of the VPP
	vppIfaces, err := vppdump.DumpInterfaces(plugin.Log, plugin.vppCh, plugin.Stopwatch)
	if err != nil {
		return err
	}

	plugin.Log.Debug("VPP contains len(vppIfaces)=", len(vppIfaces))

	// Step 1: Correlate vppIfaces with northbound interfaces
	// it means to find out names for vpp swIndexes
	// (temporary: correlate using persisted sw_if_indexes)

	corr := nametoidx.NewNameToIdx(logroot.StandardLogger(), core.PluginName("defaultvppplugins-ifplugin"), "iface resync corr", nil)

	if !plugin.resyncDoneOnce { //probably shortly after startup
		// we temporary load the last state from the file (in case the agent crashed)
		// later we use the VPP Master ID to correlate

		tmpCorr := nametoidx.NewNameToIdx(logroot.StandardLogger(), core.PluginName("defaultvppplugins-ifplugin"), "iface resync corr", nil)

		err = persist.Marshalling(plugin.ServiceLabel.GetAgentLabel(), plugin.swIfIndexes.GetMapping(), tmpCorr)
		if err != nil {
			return err
		}
		plugin.resyncDoneOnce = true

		// check if all loaded indexes are still in VPP (remove all sw_if_idx not contained in the VPP dump)
		for _, nbIface := range nbIfaces {
			if vppSwIfIdx, meta, found := tmpCorr.LookupIdx(nbIface.Name); found {
				corr.RegisterName(nbIface.Name, vppSwIfIdx, meta)
				plugin.Log.WithField("swIfIndex", vppSwIfIdx).Debug("Correlation ", nbIface.Name)
			}
		}
	}
	var wasError error

	// Step 2: delete obsolete vpp configuration
	for vppSwIfIdx, vppIface := range vppIfaces {
		_, _, found := corr.LookupName(vppSwIfIdx)

		if vppSwIfIdx == 0 {
			// local0 - default loopback interface
			plugin.swIfIndexes.RegisterName(vppIface.VPPInternalName, vppSwIfIdx, &vppIface.Interfaces_Interface)
		} else if vppIface.Type == intf.InterfaceType_ETHERNET_CSMACD {
			// physical interface (PCI device)
			plugin.swIfIndexes.RegisterName(vppIface.VPPInternalName, vppSwIfIdx, &vppIface.Interfaces_Interface)
		} else if !found {
			err := plugin.deleteVPPInterface(&vppIface.Interfaces_Interface, vppSwIfIdx)

			plugin.Log.WithFields(logging.Fields{"swIfIndex": vppSwIfIdx, "vppIface": vppIface}).
				Info("Interface deletion ", err)

			if err != nil {
				wasError = err
			}
		}
	}

	toBeConfigured := []*intf.Interfaces_Interface{}

	// Step 3: modify existing vpp configuration
	for _, nbIface := range nbIfaces {
		swIfIdx, _, found := corr.LookupIdx(nbIface.Name)
		vppIface, foundDump := vppIfaces[swIfIdx]
		if found && foundDump {
			err := plugin.modifyVPPInterface(nbIface, &vppIface.Interfaces_Interface, swIfIdx, vppIface.Type)
			if err != nil {
				wasError = err
			}
			if !plugin.afPacketConfigurator.IsPendingAfPacket(nbIface) {
				// even if error occurred (because there is still swIfIndex)
				plugin.swIfIndexes.RegisterName(nbIface.Name, swIfIdx, nbIface)
			}
		} else {
			toBeConfigured = append(toBeConfigured, nbIface)
		}
	}

	// Step 4: create missing vpp configuration
	for _, nbIface := range toBeConfigured {
		err := plugin.ConfigureVPPInterface(nbIface)
		if err != nil {
			wasError = err
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC Interface end. ", wasError)

	return wasError
}

// VerifyVPPConfigPresence dumps VPP interface configuration on the vpp. If there are any interfaces configured (except
// the local0), it returns false (do not interrupt the resto of the resync), otherwise returns true
func (plugin *InterfaceConfigurator) VerifyVPPConfigPresence(nbIfaces []*intf.Interfaces_Interface) bool {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC Interface begin.")
	// notify that the resync should be stopped
	var stop bool

	// Step 0: Dump actual state of the VPP
	vppIfaces, err := vppdump.DumpInterfaces(plugin.Log, plugin.vppCh, plugin.Stopwatch)
	if err != nil {
		return stop
	}

	// The strategy is optimize-cold-start, so look over all dumped VPP interfaces and check for the configured ones
	// (leave out the local0). If there are any other interfaces, return true (resync will continue).
	// If not, return a false flag which cancels the VPP resync operation.
	plugin.Log.Info("optimize-cold-start VPP resync strategy chosen, resolving...")
	if len(vppIfaces) == 0 {
		stop = true
		plugin.Log.Infof("...VPP resync interrupted assuming there is no configuration on the VPP (no interface was found)")
		return stop
	}
	// in interface exists, try to find local0 interface (index 0)
	_, ok := vppIfaces[0]
	// in case local0 is the only interface on the vpp, stop the resync
	if len(vppIfaces) == 1 && ok {
		stop = true
		plugin.Log.Infof("...VPP resync interrupted assuming there is no configuration on the VPP (only local0 was found)")
		return stop
	}
	// otherwise continue normally
	plugin.Log.Infof("... VPP configuration found, continue with VPP resync")

	return stop
}

// ResyncSession writes BFD sessions to the empty VPP
func (plugin *BFDConfigurator) ResyncSession(bfds []*bfd.SingleHopBFD_Session) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BFD Session begin.")
	// Calculate and log bfd resync
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// lookup BFD sessions
	err := plugin.LookupBfdSessions()
	if err != nil {
		return err
	}

	var wasError error

	// create BFD sessions
	for _, bfdSession := range bfds {
		err = plugin.ConfigureBfdSession(bfdSession)
		if err != nil {
			wasError = err
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BFD Session end. ", wasError)

	return wasError
}

// ResyncAuthKey writes BFD keys to the empty VPP
func (plugin *BFDConfigurator) ResyncAuthKey(bfds []*bfd.SingleHopBFD_Key) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BFD Keys begin.")
	// Calculate and log bfd resync
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// lookup BFD auth keys
	err := plugin.LookupBfdKeys()
	if err != nil {
		return err
	}

	var wasError error

	// create BFD auth keys
	for _, bfdKey := range bfds {
		err = plugin.ConfigureBfdAuthKey(bfdKey)
		if err != nil {
			wasError = err
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC BFD Keys end. ", wasError)

	return wasError
}

// ResyncEchoFunction writes BFD echo function to the empty VPP
func (plugin *BFDConfigurator) ResyncEchoFunction(bfds []*bfd.SingleHopBFD_EchoFunction) error {
	return nil
}
