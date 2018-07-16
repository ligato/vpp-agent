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
	"time"

	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linux/model/l3"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	"github.com/vishvananda/netlink"
)

// Resync configures an initial set of ARPs. Existing Linux ARPs are registered and potentially re-configured.
func (plugin *LinuxArpConfigurator) Resync(arpEntries []*l3.LinuxStaticArpEntries_ArpEntry) (errs []error) {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC ARPs begin.")

	defer func(t time.Time) {
		plugin.stopwatch.TimeLog("configure-linux-arp").LogTimeEntry(time.Since(t))
	}(time.Now())

	// Create missing arp entries and update existing ones
	for _, entry := range arpEntries {
		err := plugin.ConfigureLinuxStaticArpEntry(entry)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Dump pre-existing not managed arp entries
	err := plugin.LookupLinuxArpEntries()
	if err != nil {
		errs = append(errs, err)
	}

	plugin.log.WithField("cfg", plugin).Debug("RESYNC ARPs end. ")

	return
}

// Resync configures an initial set of static routes. Existing Linux static routes are registered and potentially
// re-configured. Resync does not remove any linux route.
func (plugin *LinuxRouteConfigurator) Resync(nbRoutes []*l3.LinuxStaticRoutes_Route) (errs []error) {
	plugin.log.WithField("cfg", plugin).Debug("RESYNC static routes begin.")

	defer func(t time.Time) {
		plugin.stopwatch.TimeLog("configure-linux-route").LogTimeEntry(time.Since(t))
	}(time.Now())

	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()

	// First step is to find a linux equivalent for NB route config
	for _, nbRoute := range nbRoutes {
		// Route interface exists
		if nbRoute.Interface != "" {
			_, _, found := plugin.ifIndexes.LookupIdx(nbRoute.Interface)
			if !found {
				// If route interface does not exist, cache it
				plugin.log.Debugf("RESYNC static route %v: interface %s does not exists, moving to cache",
					nbRoute.Name, nbRoute.Interface)
				plugin.rtCachedIfRoutes.RegisterName(nbRoute.Name, plugin.rtIdxSeq, nbRoute)
				plugin.rtIdxSeq++
				continue
			}
		}

		// There can be several routes found according to matching parameters
		linuxRtList, err := plugin.findLinuxRoutes(nbRoute, nsMgmtCtx)
		if err != nil {
			plugin.log.Error(err)
			errs = append(errs, err)
			continue
		}
		plugin.log.Debugf("found %d linux routes to compare for %s", len(linuxRtList), nbRoute.Name)
		// Find at least one route which has the same parameters
		var rtFound bool
		for rtIdx, linuxRtEntry := range linuxRtList {
			// Route interface interface
			var hostName string
			var ifData *ifaceidx.IndexedLinuxInterface
			if linuxRtEntry.LinkIndex != 0 {
				var found bool
				var nsName string
				if nbRoute.Namespace == nil {
					nsName = ifaceidx.DefNs
				} else {
					nsName = nbRoute.Namespace.Name
				}
				_, ifData, found = plugin.ifIndexes.LookupNameByNamespace(uint32(linuxRtEntry.LinkIndex), nsName)
				if !found || ifData == nil {
					plugin.log.Debugf("Interface %d (data %v) not found for route", linuxRtEntry.LinkIndex, ifData)
				} else {
					hostName = ifData.Data.HostIfName
				}
			}
			linuxRt := plugin.transformRoute(linuxRtEntry, hostName)
			if plugin.isRouteEqual(rtIdx, nbRoute, linuxRt) {
				rtFound = true
				break
			}
		}
		if rtFound {
			// Register route if found
			plugin.log.Debugf("RESYNC Linux routes: %s was found and will be registered without additional changes", nbRoute.Name)
			plugin.rtIndexes.RegisterName(nbRoute.Name, plugin.rtIdxSeq, nbRoute)
			plugin.rtIdxSeq++
			// Resolve cached routes
			if !nbRoute.Default {
				plugin.retryDefaultRoutes(nbRoute)
			}
		} else {
			// Configure route if not found
			plugin.log.Debugf("RESYNC Linux routes: %s was not found and will be configured", nbRoute.Name)
			if err := plugin.ConfigureLinuxStaticRoute(nbRoute); err != nil {
				plugin.log.Error(err)
				errs = append(errs, err)
			}
		}
	}

	return
}

// Look for routes similar to provided NB config in respective namespace. Routes can be read using destination address
// or interface. FOr every config, both ways are used.
func (plugin *LinuxRouteConfigurator) findLinuxRoutes(nbRoute *l3.LinuxStaticRoutes_Route, nsMgmtCtx *nsplugin.NamespaceMgmtCtx) ([]netlink.Route, error) {
	plugin.log.Debugf("Looking for equivalent linux routes for %s", nbRoute.Name)

	// Move to proper namespace
	if nbRoute.Namespace != nil {
		// Switch to namespace
		routeNs := plugin.nsHandler.RouteNsToGeneric(nbRoute.Namespace)
		revertNs, err := plugin.nsHandler.SwitchNamespace(routeNs, nsMgmtCtx)
		if err != nil {
			return nil, fmt.Errorf("RESYNC Linux route %s: failed to switch to namespace %s: %v",
				nbRoute.Name, nbRoute.Namespace.Name, err)
		}
		defer revertNs()
	}
	var linuxRoutes []netlink.Route
	// Look for routes using destination IP address
	if nbRoute.DstIpAddr != "" && plugin.networkReachable(nbRoute.Namespace, nbRoute.DstIpAddr) {
		_, dstNetIP, err := net.ParseCIDR(nbRoute.DstIpAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse destination IP address %s: %v", nbRoute.DstIpAddr, err)
		}
		linuxRts, err := netlink.RouteGet(dstNetIP.IP)
		if err != nil {
			return nil, fmt.Errorf("failed to read linux route %s using address %s: %v",
				nbRoute.Name, nbRoute.DstIpAddr, err)
		}
		if linuxRts != nil {
			linuxRoutes = append(linuxRoutes, linuxRts...)
		}
	}
	// Look for routes using interface
	if nbRoute.Interface != "" {
		// Look whether interface is registered
		_, meta, found := plugin.ifIndexes.LookupIdx(nbRoute.Interface)
		if !found {
			// Should not happen, was successfully checked before
			plugin.log.Errorf("Route %s interface %s is missing from the mapping", nbRoute.Name, nbRoute.Interface)
		} else if meta == nil || meta.Data == nil {
			plugin.log.Errorf("Interface %s data missing", nbRoute.Interface)
		} else {
			// Look for interface using host name
			link, err := netlink.LinkByName(meta.Data.HostIfName)
			if err != nil {
				return nil, fmt.Errorf("failed to read interface %s: %v", meta.Data.HostIfName, err)
			}
			linuxRts, err := netlink.RouteList(link, netlink.FAMILY_ALL)
			if err != nil {
				return nil, fmt.Errorf("failed to read linux route %s using interface %s: %v",
					nbRoute.Name, meta.Data.HostIfName, err)
			}
			if linuxRts != nil {
				linuxRoutes = append(linuxRoutes, linuxRts...)
			}
		}
	}

	if len(linuxRoutes) == 0 {
		plugin.log.Debugf("Equivalent for route %s was not found", nbRoute.Name)
	}

	return linuxRoutes, nil
}

// Compare all route parameters and returns true if routes are equal, false otherwise
func (plugin *LinuxRouteConfigurator) isRouteEqual(rtIdx int, nbRoute, linuxRt *l3.LinuxStaticRoutes_Route) bool {
	// Interface (if exists)
	if nbRoute.Interface != "" && nbRoute.Interface != linuxRt.Interface {
		plugin.log.Debugf("Linux route %d: interface is different (NB: %s, Linux: %s)",
			rtIdx, nbRoute.Interface, linuxRt.Interface)
		return false
	}
	// Default route
	if nbRoute.Default {
		if !linuxRt.Default {
			plugin.log.Debugf("Linux route %d: NB route is default, but linux route is not", rtIdx)
			return false
		}
		if nbRoute.GwAddr != linuxRt.GwAddr {
			plugin.log.Debugf("Linux route %d: gateway is different (NB: %s, Linux: %s)",
				rtIdx, nbRoute.GwAddr, linuxRt.GwAddr)
			return false
		}
		if nbRoute.Metric != linuxRt.Metric {
			plugin.log.Debugf("Linux route %d: metric is different (NB: %s, Linux: %s)",
				rtIdx, nbRoute.Metric, linuxRt.Metric)
			return false
		}
		return true
	}
	// Static route
	_, nbIPNet, err := net.ParseCIDR(nbRoute.DstIpAddr)
	if err != nil {
		plugin.log.Error(err)
		return false
	}
	if nbIPNet.IP.String() != linuxRt.DstIpAddr {
		plugin.log.Debugf("Linux route %d: destination address is different (NB: %s, Linux: %s)",
			rtIdx, nbIPNet.IP.String(), linuxRt.DstIpAddr)
		return false
	}
	// Compare source IP/gateway
	if nbRoute.SrcIpAddr == "" && linuxRt.SrcIpAddr != "" || nbRoute.SrcIpAddr != "" && linuxRt.SrcIpAddr == "" {
		if nbRoute.SrcIpAddr == "" && nbRoute.SrcIpAddr != linuxRt.GwAddr {
			plugin.log.Debugf("Linux route %d: source does not match gateway (NB: %s, Linux: %s)",
				rtIdx, nbRoute.SrcIpAddr, linuxRt.SrcIpAddr)
			return false
		} else if linuxRt.SrcIpAddr == "" && nbRoute.GwAddr != linuxRt.SrcIpAddr {
			plugin.log.Debugf("Linux route %d: source does not match gateway (NB: %s, Linux: %s)",
				rtIdx, nbRoute.SrcIpAddr, linuxRt.SrcIpAddr)
			return false
		}
	} else if nbRoute.SrcIpAddr != "" && linuxRt.SrcIpAddr != "" && nbRoute.SrcIpAddr != linuxRt.SrcIpAddr {
		plugin.log.Debugf("Linux route %d: source address is different (NB: %s, Linux: %s)",
			rtIdx, nbRoute.SrcIpAddr, linuxRt.SrcIpAddr)
		return false
	}

	if nbRoute.SrcIpAddr != "" && nbRoute.SrcIpAddr != linuxRt.SrcIpAddr {
		plugin.log.Debugf("Linux route %d: source address is different (NB: %s, Linux: %s)",
			rtIdx, nbRoute.SrcIpAddr, linuxRt.SrcIpAddr)
		return false
	}
	// If NB scope is nil, set scope type LINK (default value)
	if nbRoute.Scope == nil {
		nbRoute.Scope = &l3.LinuxStaticRoutes_Route_Scope{
			Type: l3.LinuxStaticRoutes_Route_Scope_LINK,
		}
	} else if linuxRt.Scope != nil {
		if nbRoute.Scope.Type != linuxRt.Scope.Type {
			plugin.log.Debugf("Linux route %d: scope is different (NB: %s, Linux: %s)",
				rtIdx, nbRoute.Scope.Type, linuxRt.Scope.Type)
			return false
		}
	}

	return true
}

// Parse netlink type scope to proto
func (plugin *LinuxRouteConfigurator) parseLinuxRouteScope(scope netlink.Scope) *l3.LinuxStaticRoutes_Route_Scope {
	switch scope {
	case netlink.SCOPE_UNIVERSE:
		return &l3.LinuxStaticRoutes_Route_Scope{
			Type: l3.LinuxStaticRoutes_Route_Scope_GLOBAL,
		}
	case netlink.SCOPE_HOST:
		return &l3.LinuxStaticRoutes_Route_Scope{
			Type: l3.LinuxStaticRoutes_Route_Scope_HOST,
		}
	case netlink.SCOPE_LINK:
		return &l3.LinuxStaticRoutes_Route_Scope{
			Type: l3.LinuxStaticRoutes_Route_Scope_LINK,
		}
	case netlink.SCOPE_SITE:
		return &l3.LinuxStaticRoutes_Route_Scope{
			Type: l3.LinuxStaticRoutes_Route_Scope_SITE,
		}
	default:
		plugin.log.Infof("Unknown scope type, setting to default (link): %v", scope)
		return &l3.LinuxStaticRoutes_Route_Scope{
			Type: l3.LinuxStaticRoutes_Route_Scope_LINK,
		}
	}
}
