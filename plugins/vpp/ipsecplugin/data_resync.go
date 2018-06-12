//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ipsecplugin

import "github.com/ligato/vpp-agent/plugins/vpp/model/ipsec"

// Resync writes missing IPSec configs to the VPP and removes obsolete ones.
func (plugin *IPSecConfigurator) Resync(spds []*ipsec.SecurityPolicyDatabases_SPD, sas []*ipsec.SecurityAssociations_SA, tunnels []*ipsec.TunnelInterfaces_Tunnel) error {
	plugin.log.Debug("RESYNC IPSec begin.")

	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	plugin.clearMapping()

	// TODO: dump existing configuration from VPP

	for _, sa := range sas {
		if err := plugin.ConfigureSA(sa); err != nil {
			plugin.log.Error(err)
			continue
		}
	}

	for _, spd := range spds {
		if err := plugin.ConfigureSPD(spd); err != nil {
			plugin.log.Error(err)
			continue
		}
	}

	for _, tunnel := range tunnels {
		if err := plugin.ConfigureTunnel(tunnel); err != nil {
			plugin.log.Error(err)
			continue
		}
	}

	plugin.log.Debug("RESYNC IPSec end.")
	return nil
}
