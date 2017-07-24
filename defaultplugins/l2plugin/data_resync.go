package l2plugin

import (
	"fmt"
	"github.com/ligato/cn-infra/core"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/vppdump"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/cn-infra/logging/logroot"
)

// Resync writes BDs to the empty VPP
func (plugin *BDConfigurator) Resync(nbBDs []*l2.BridgeDomains_BridgeDomain) error {
	log.WithField("cfg", plugin).Debug("RESYNC BDs begin.")

	// Step 0: Dump actual state of the VPP
	vppBDs, err := vppdump.DumpBridgeDomains(plugin.vppChan)
	// old implemention: err := plugin.LookupBridgeDomains()
	if err != nil {
		return err
	}

	pluginID := core.PluginName("defaultvppplugins-l2plugin")

	var wasError error

	// Step 1: delete existing vpp configuration (current ModifyBridgeDomain does it also... need to improve that first)
	for vppIdx, vppBD := range vppBDs {
		hackIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logroot.Logger(),pluginID,
			"hack_sw_if_indexes", ifaceidx.IndexMetadata))

		// hack to reuse existing binary call wrappers
		hackBD := l2.BridgeDomains_BridgeDomain(vppBD.BridgeDomains_BridgeDomain)
		for _, vppBDIface := range vppBD.Interfaces {
			hackIfaceName := fmt.Sprintf("%d", vppBDIface.SwIfIndex)
			hackIfIndexes.RegisterName(hackIfaceName, vppBDIface.SwIfIndex, nil)
			hackBDIface := l2.BridgeDomains_BridgeDomain_Interfaces(vppBDIface.BridgeDomains_BridgeDomain_Interfaces)
			hackBDIface.Name = hackIfaceName
			hackBD.Interfaces = append(hackBD.Interfaces, &hackBDIface)
		}

		vppcalls.VppUnsetAllInterfacesFromBridgeDomain(&hackBD, vppIdx,
			hackIfIndexes, plugin.vppChan)
		err := plugin.deleteBridgeDomain(&hackBD, vppIdx)
		// TODO check if it is ok to delete the initial BD
		if err != nil {
			wasError = err
		}
	}

	// Step 2: create missing vpp configuration
	for _, nbBD := range nbBDs {
		err := plugin.ConfigureBridgeDomain(nbBD)
		if err != nil {
			wasError = err
		}
	}

	log.WithField("cfg", plugin).Debug("RESYNC BDs end. ", wasError)

	return wasError
}

// Resync writes FIBs to the empty VPP
func (plugin *FIBConfigurator) Resync(fibConfig []*l2.FibTableEntries_FibTableEntry) error {
	log.WithField("cfg", plugin).Debug("RESYNC FIBs begin.")

	for _, fib := range fibConfig {
		plugin.Add(fib, func(err2 error) {
			if err2 != nil {
				log.Error(err2)
			}
		})
	}

	activeDomains, err := vppdump.DumpBridgeDomainIDs(plugin.vppChannel)
	if err != nil {
		return err
	}
	for _, domainID := range activeDomains {
		plugin.LookupFIBEntries(domainID)
	}

	log.WithField("cfg", plugin).Debug("RESYNC FIBs end.")

	return nil
}

// Resync writes XCons to the empty VPP
func (plugin *XConnectConfigurator) Resync(xcConfig []*l2.XConnectPairs_XConnectPair) error {
	log.WithField("cfg", plugin).Debug("RESYNC XConnect begin.")

	err := plugin.LookupXConnectPairs()
	if err != nil {
		return err
	}

	var wasError error
	for _, xcon := range xcConfig {
		err = plugin.ConfigureXConnectPair(xcon)
		if err != nil {
			wasError = err
		}
	}

	log.WithField("cfg", plugin).Debug("RESYNC XConnect end. ", wasError)

	return wasError
}
