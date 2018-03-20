package ipsecplugin

import "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/ipsec"

// Resync writes missing IPSec configs to the VPP and removes obsolete ones.
func (plugin *IPSecConfigurator) Resync(spds []*ipsec.SecurityPolicyDatabases_SPD, sas []*ipsec.SecurityAssociations_SA) error {
	plugin.Log.Debug("RESYNC IPSec begin.")

	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// TODO: dump existing configuration from VPP

	for _, sa := range sas {
		if err := plugin.ConfigureSA(sa); err != nil {
			plugin.Log.Error(err)
			continue
		}
	}

	for _, spd := range spds {
		if err := plugin.ConfigureSPD(spd); err != nil {
			plugin.Log.Error(err)
			continue
		}
	}

	plugin.Log.Debug("RESYNC IPSec end.")
	return nil
}
