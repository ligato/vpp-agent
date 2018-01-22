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

package l3plugin

import (
	"github.com/ligato/cn-infra/logging/measure"
	l3ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppdump"
)

// Resync configures the VPP static routes.
func (plugin *RouteConfigurator) Resync(nbRoutes []*l3.StaticRoutes_Route) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC routes begin. ")
	// Calculate and log route resync.
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Retrieve VPP route configuration
	vppRoutes, err := vppdump.DumpStaticRoutes(plugin.Log, plugin.vppChan, measure.GetTimeLog(l3ba.IPFibDump{}, plugin.Stopwatch))
	if err != nil {
		return err
	}

	// Correlate VPP and NB configuration
	for _, vppRoute := range vppRoutes {
		// Look for the same route in the configuration
		for _, nbRoute := range nbRoutes {
			ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(nbRoute.OutgoingInterface)
			if !found {
				continue
			}
			if vppRoute.OutIface != ifIdx {
				continue
			}
			if vppRoute.DstAddr.String() != nbRoute.DstIpAddr {
				continue
			}
			if vppRoute.VrfID != nbRoute.VrfId {
				continue
			}
			if vppRoute.Weight != nbRoute.Weight {
				continue
			}
			if vppRoute.Preference != nbRoute.Preference {
				continue
			}
			if vppRoute.NextHopAddr.String() != nbRoute.NextHopAddr {
				continue
			}
			// Register existing routes
			routeID := routeIdentifier(nbRoute.VrfId, nbRoute.DstIpAddr, nbRoute.NextHopAddr)
			plugin.RouteIndexes.RegisterName(routeID, plugin.RouteIndexSeq, nbRoute)
			plugin.RouteIndexSeq++
		}

	}

	// Add missing route configuration
	var wasError error
	if len(nbRoutes) > 0 {
		for _, nbRoute := range nbRoutes {
			routeID := routeIdentifier(nbRoute.VrfId, nbRoute.DstIpAddr, nbRoute.NextHopAddr)
			_, _, found := plugin.RouteIndexes.LookupIdx(routeID)
			if !found {
				// create new route if does not exist yet. VRF ID is already validated at this point.
				if err := plugin.ConfigureRoute(nbRoute, string(nbRoute.VrfId)); err != nil {
					plugin.Log.Error(err)
					wasError = err
				}
			}
		}
	}
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC routes end. ", wasError)
	return wasError
}

// Resync confgures the empty VPP (overwrites the arp entries)
func (plugin *ArpConfigurator) Resync(arpEntries []*l3.ArpTable_ArpTableEntry) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC arp begin. ")
	// Calculate and log arp resync
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	var wasError error
	if len(arpEntries) > 0 {
		for _, entry := range arpEntries {
			wasError = plugin.AddArp(entry)
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC arp end. ", wasError)
	return nil
}
