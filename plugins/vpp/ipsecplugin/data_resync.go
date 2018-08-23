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

import (
	"github.com/ligato/vpp-agent/plugins/vpp/model/ipsec"
	"github.com/go-errors/errors"
)

// Resync writes missing IPSec configs to the VPP and removes obsolete ones.
func (c *IPSecConfigurator) Resync(spds []*ipsec.SecurityPolicyDatabases_SPD, sas []*ipsec.SecurityAssociations_SA, tunnels []*ipsec.TunnelInterfaces_Tunnel) error {
	defer func() {
		if c.stopwatch != nil {
			c.stopwatch.PrintLog()
		}
	}()

	c.clearMapping()

	// TODO: dump existing configuration from VPP

	for _, sa := range sas {
		if err := c.ConfigureSA(sa); err != nil {
			return errors.Errorf("IPSec resync error: failed to configure SA %d: %v", sa.Name, err)
		}
	}

	for _, spd := range spds {
		if err := c.ConfigureSPD(spd); err != nil {
			return errors.Errorf("IPSec resync error: failed to configure SPD %d: %v", spd.Name, err)
		}
	}

	for _, tunnel := range tunnels {
		if err := c.ConfigureTunnel(tunnel); err != nil {
			return errors.Errorf("IPSec resync error: failed to configure tunnel interface %d: %v", tunnel.Name, err)
		}
	}

	c.log.Debug("IPSec resync done")
	return nil
}
