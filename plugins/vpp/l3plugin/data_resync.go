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
	"fmt"

	"net"

	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
)

// Resync configures the VPP static routes.
func (plugin *RouteConfigurator) Resync(nbRoutes []*l3.StaticRoutes_Route) error {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC routes begin. ")
	// Calculate and log route resync.
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// Re-initialize cache
	plugin.clearMapping()

	// Retrieve VPP route configuration
	vppRouteDetails, err := plugin.rtHandler.DumpStaticRoutes()
	if err != nil {
		return err
	}
	plugin.log.Debugf("Found %d routes configured on the VPP", len(vppRouteDetails))

	// Correlate NB and VPP configuration
	for _, nbRoute := range nbRoutes {
		nbRouteID := routeIdentifier(nbRoute.VrfId, nbRoute.DstIpAddr, nbRoute.NextHopAddr)
		nbIfIdx, _, found := plugin.ifIndexes.LookupIdx(nbRoute.OutgoingInterface)
		if !found {
			if nbRoute.Type == l3.StaticRoutes_Route_INTER_VRF {
				// expected by inter VRF-routes
				nbIfIdx = vppcalls.NextHopOutgoingIfUnset
			} else {
				plugin.log.Debugf("RESYNC routes: outgoing interface not found for %s", nbRouteID)
				plugin.rtCachedIndexes.RegisterName(nbRouteID, plugin.rtIndexSeq, nbRoute)
				plugin.rtIndexSeq++
				continue
			}
		}
		// Default VPP value for weight in case it is not set
		if nbRoute.Weight == 0 {
			nbRoute.Weight = 1
		}
		// Look for the same route in the configuration
		for _, vppRouteDetail := range vppRouteDetails {
			if vppRouteDetail.Route == nil {
				continue
			}
			vppRoute := vppRouteDetail.Route
			vppRouteID := routeIdentifier(vppRoute.VrfId, vppRoute.DstIpAddr, vppRoute.NextHopAddr)
			plugin.log.Debugf("RESYNC routes: comparing %s and %s", nbRouteID, vppRouteID)
			if int32(vppRoute.Type) != int32(nbRoute.Type) {
				plugin.log.Debugf("RESYNC routes: route type is different (NB: %d, VPP %d)",
					nbRoute.Type, vppRoute.Type)
				continue
			}
			if vppRoute.OutgoingInterface != nbRoute.OutgoingInterface {
				plugin.log.Debugf("RESYNC routes: interface index is different (NB: %d, VPP %d)",
					nbIfIdx, vppRoute.OutgoingInterface)
				continue
			}
			if vppRoute.DstIpAddr != nbRoute.DstIpAddr {
				plugin.log.Debugf("RESYNC routes: dst address is different (NB: %s, VPP %s)",
					nbRoute.DstIpAddr, vppRoute.DstIpAddr)
				continue
			}
			if vppRoute.VrfId != nbRoute.VrfId {
				plugin.log.Debugf("RESYNC routes: VRF ID is different (NB: %d, VPP %d)",
					nbRoute.VrfId, vppRoute.VrfId)
				continue
			}
			if vppRoute.Weight != nbRoute.Weight {
				plugin.log.Debugf("RESYNC routes: weight is different (NB: %d, VPP %d)",
					nbRoute.Weight, vppRoute.Weight)
				continue
			}
			if vppRoute.Preference != nbRoute.Preference {
				plugin.log.Debugf("RESYNC routes: preference is different (NB: %d, VPP %d)",
					nbRoute.Preference, vppRoute.Preference)
				continue
			}
			if vppRoute.NextHopAddr != nbRoute.NextHopAddr {
				if nbRoute.NextHopAddr == "" {
					plugin.log.Debugf("RESYNC routes: empty next hop address matched (NB: %s, VPP %s)",
						nbRoute.NextHopAddr, vppRoute.NextHopAddr)
				} else {
					plugin.log.Debugf("RESYNC routes: next hop address is different (NB: %s, VPP %s)",
						nbRoute.NextHopAddr, vppRoute.NextHopAddr)
					continue
				}
			}
			if vppRoute.ViaVrfId != nbRoute.ViaVrfId {
				plugin.log.Debugf("RESYNC routes: via VRF ID is different (NB: %d, VPP %d)",
					nbRoute.ViaVrfId, vppRoute.ViaVrfId)
				continue
			}
			// Register existing routes
			plugin.rtIndexes.RegisterName(nbRouteID, plugin.rtIndexSeq, nbRoute)
			plugin.rtIndexSeq++
			plugin.log.Debugf("RESYNC routes: route %s registered without additional changes", nbRouteID)
			break
		}
	}

	// Add missing route configuration
	var wasError error
	for _, nbRoute := range nbRoutes {
		routeID := routeIdentifier(nbRoute.VrfId, nbRoute.DstIpAddr, nbRoute.NextHopAddr)
		_, _, found := plugin.rtIndexes.LookupIdx(routeID)
		if !found {
			// create new route if does not exist yet. VRF ID is already validated at this point.
			if err := plugin.ConfigureRoute(nbRoute, fmt.Sprintf("%d", nbRoute.VrfId)); err != nil {
				plugin.log.Error(err)
				wasError = err
			}
			plugin.log.Debugf("RESYNC routes: route %s was configured", routeID)
		}
	}

	// Remove other routes except DROP type
	for _, vppRoute := range vppRouteDetails {
		if routeMayBeRemoved(vppRoute) {
			route := vppRoute.Route
			routeID := routeIdentifier(route.VrfId, route.DstIpAddr, route.NextHopAddr)
			_, _, found := plugin.rtIndexes.LookupIdx(routeID)
			if !found {
				// Register before removal
				plugin.rtIndexes.RegisterName(routeID, plugin.rtIndexSeq, route)
				plugin.rtIndexSeq++
				if err := plugin.DeleteRoute(route, fmt.Sprintf("%d", route.VrfId)); err != nil {
					plugin.log.Error(err)
					wasError = err
				}
				plugin.log.Debugf("RESYNC routes: vpp route %s removed", routeID)
			}
		}
	}

	plugin.log.WithField("cfg", plugin).Debug("RESYNC routes end. ", wasError)
	return wasError
}

// Following rules are currently applied:
// - no DROP type route can be removed in order to prevent removal of VPP default routes
// - IPv6 link local route cannot be removed
func routeMayBeRemoved(route *vppcalls.RouteDetails) bool {
	if route.Route.Type == l3.StaticRoutes_Route_DROP {
		return false
	}
	if route.Meta.IsIPv6 && net.ParseIP(route.Route.DstIpAddr).IsLinkLocalUnicast() {
		return false
	}
	return true
}

// Resync confgures the empty VPP (overwrites the arp entries)
func (plugin *ArpConfigurator) Resync(arpEntries []*l3.ArpTable_ArpEntry) error {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC arp begin. ")
	// Calculate and log arp resync
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// Re-initialize cache
	plugin.clearMapping()

	var wasError error
	if len(arpEntries) > 0 {
		for _, entry := range arpEntries {
			wasError = plugin.AddArp(entry)
		}
	}

	plugin.log.WithField("cfg", plugin).Debug("RESYNC arp end. ", wasError)
	return nil
}

// Resync confgures the empty VPP (overwrites the proxy arp entries)
func (plugin *ProxyArpConfigurator) ResyncInterfaces(nbProxyArpIfs []*l3.ProxyArpInterfaces_InterfaceList) error {
	plugin.log.Debug("RESYNC proxy ARP interfaces begin. ")
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// Re-initialize cache
	plugin.clearMapping()

	// Todo: dump proxy arp

	var wasError error
	if len(nbProxyArpIfs) > 0 {
		for _, entry := range nbProxyArpIfs {
			wasError = plugin.AddInterface(entry)
		}
	}

	plugin.log.Debug("RESYNC proxy ARP interface end. ", wasError)
	return nil
}

// Resync confgures the empty VPP (overwrites the proxy arp ranges)
func (plugin *ProxyArpConfigurator) ResyncRanges(nbProxyArpRanges []*l3.ProxyArpRanges_RangeList) error {
	plugin.log.Debug("RESYNC proxy ARP ranges begin. ")
	defer func() {
		if plugin.stopwatch != nil {
			plugin.stopwatch.PrintLog()
		}
	}()

	// Todo: dump proxy arp

	var wasError error
	if len(nbProxyArpRanges) > 0 {
		for _, entry := range nbProxyArpRanges {
			wasError = plugin.AddRange(entry)
		}
	}

	plugin.log.Debug("RESYNC proxy ARP ranges end. ", wasError)
	return nil
}

// Resync configures the empty VPP (adds IP scan neigh config)
func (p *IPNeighConfigurator) Resync(config *l3.IPScanNeighbor) error {
	p.log.Debug("RESYNC IP neighbor begin. ")
	defer func() {
		if p.stopwatch != nil {
			p.stopwatch.PrintLog()
		}
	}()

	var wasError error
	if err := p.Set(config); err != nil {
		return err
	}

	p.log.Debug("RESYNC IP neighbor end. ", wasError)
	return nil
}
