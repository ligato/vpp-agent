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
	plugin.Log.Debugf("Found %d routes configured on the VPP", len(vppRoutes))

	// Correlate NB and VPP configuration
	for _, nbRoute := range nbRoutes {
		nbRouteID := routeIdentifier(nbRoute.VrfId, nbRoute.DstIpAddr, nbRoute.NextHopAddr)
		nbIfIdx, _, found := plugin.SwIfIndexes.LookupIdx(nbRoute.OutgoingInterface)
		if !found {
			plugin.Log.Debugf("RESYNC routes: outgoing interface not found for %s", nbRouteID)
			plugin.RouteCachedIndex.RegisterName(nbRouteID, plugin.RouteIndexSeq, nbRoute)
			plugin.RouteIndexSeq++
			continue
		}
		// Default VPP value for weight in case it is not set
		if nbRoute.Weight == 0 {
			nbRoute.Weight = 1
		}
		// Look for the same route in the configuration
		for _, vppRoute := range vppRoutes {
			vppRouteID := routeIdentifier(vppRoute.VrfID, vppRoute.DstAddr.String(), vppRoute.NextHopAddr.String())
			plugin.Log.Debugf("RESYNC routes: comparing %s and %s", nbRouteID, vppRouteID)
			if vppRoute.OutIface != nbIfIdx {
				plugin.Log.Debugf("RESYNC routes: interface index is different (NB: %d, VPP %d)",
					nbIfIdx, vppRoute.OutIface)
				continue
			}
			if vppRoute.DstAddr.String() != nbRoute.DstIpAddr {
				plugin.Log.Debugf("RESYNC routes: dst address is different (NB: %s, VPP %s)",
					nbRoute.DstIpAddr, vppRoute.DstAddr.String())
				continue
			}
			if vppRoute.VrfID != nbRoute.VrfId {
				plugin.Log.Debugf("RESYNC routes: VRF ID is different (NB: %d, VPP %d)",
					nbRoute.VrfId, vppRoute.VrfID)
				continue
			}
			if vppRoute.Weight != nbRoute.Weight {
				plugin.Log.Debugf("RESYNC routes: weight is different (NB: %d, VPP %d)",
					nbRoute.Weight, vppRoute.Weight)
				continue
			}
			if vppRoute.Preference != nbRoute.Preference {
				plugin.Log.Debugf("RESYNC routes: preference is different (NB: %d, VPP %d)",
					nbRoute.Preference, vppRoute.Preference)
				continue
			}
			if vppRoute.NextHopAddr.String() != nbRoute.NextHopAddr {
				plugin.Log.Debugf("RESYNC routes: next hop address is different (NB: %d, VPP %d)",
					nbRoute.NextHopAddr, vppRoute.NextHopAddr.String())
				continue
			}
			// Register existing routes
			plugin.RouteIndexes.RegisterName(nbRouteID, plugin.RouteIndexSeq, nbRoute)
			plugin.RouteIndexSeq++
			plugin.Log.Debugf("RESYNC routes: route %s registered without additional changes", nbRouteID)
			break
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
				plugin.Log.Debugf("RESYNC routes: route %s not found and will be configured", routeID)
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
func (plugin *ArpConfigurator) Resync(arpEntries []*l3.ArpTable_ArpEntry) error {
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

// Resync confgures the empty VPP (overwrites the proxy arp entries)
func (plugin *ProxyArpConfigurator) ResyncInterfaces(nbProxyArpIfs []*l3.ProxyArpInterfaces_InterfaceList) error {
	plugin.Log.Debug("RESYNC proxy ARP interfaces begin. ")
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Note: there is no API to dump already configured proxy ARP interfaces

	var wasError error
	if len(nbProxyArpIfs) > 0 {
		for _, entry := range nbProxyArpIfs {
			wasError = plugin.AddInterface(entry)
		}
	}

	plugin.Log.Debug("RESYNC proxy ARP interface end. ", wasError)
	return nil
}

// Resync confgures the empty VPP (overwrites the proxy arp ranges)
func (plugin *ProxyArpConfigurator) ResyncRanges(nbProxyArpRanges []*l3.ProxyArpRanges_RangeList) error {
	plugin.Log.Debug("RESYNC proxy ARP ranges begin. ")
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Note: there is no API to dump already configured proxy ARP ranges

	var wasError error
	if len(nbProxyArpRanges) > 0 {
		for _, entry := range nbProxyArpRanges {
			wasError = plugin.AddRange(entry)
		}
	}

	plugin.Log.Debug("RESYNC proxy ARP ranges end. ", wasError)
	return nil
}
