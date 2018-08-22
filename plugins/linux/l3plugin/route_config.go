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

//go:generate protoc --proto_path=../model/l3 --gogo_out=../model/l3 ../model/l3/l3.proto

package l3plugin

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linux/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/linux/l3plugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linux/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linux/model/l3"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	"github.com/vishvananda/netlink"
)

const (
	ipv4AddrAny = "0.0.0.0/0"
	ipv6AddrAny = "::/0"
)

// LinuxRouteConfigurator watches for any changes in the configuration of static routes as modelled by the proto file
// "model/l3/l3.proto" and stored in ETCD under the key "/vnf-agent/{vnf-agent}/linux/config/v1/route".
// Updates received from the northbound API are compared with the Linux network configuration and differences
// are applied through the Netlink AP
type LinuxRouteConfigurator struct {
	log logging.Logger

	// Mappings
	ifIndexes        ifaceidx.LinuxIfIndexRW
	rtIndexes        l3idx.LinuxRouteIndexRW                // Index mapping for ETCD route configuration
	rtAutoIndexes    l3idx.LinuxRouteIndexRW                // Index mapping for automatic interface routes (sometimes needed to evaluate network accessibility)
	rtCachedIfRoutes l3idx.LinuxRouteIndexRW                // Cache for routes requiring interface which is missing
	rtCachedGwRoutes map[string]*l3.LinuxStaticRoutes_Route // Cache for gateway routes which cannot be created at the time due to unreachable network
	rtIdxSeq         uint32

	// Linux namespace/calls handler
	l3Handler linuxcalls.NetlinkAPI
	nsHandler nsplugin.NamespaceAPI

	// Timer used to measure and store time
	stopwatch *measure.Stopwatch
}

// Init initializes static route configurator and starts goroutines
func (plugin *LinuxRouteConfigurator) Init(logger logging.PluginLogger, l3Handler linuxcalls.NetlinkAPI, nsHandler nsplugin.NamespaceAPI,
	ifIndexes ifaceidx.LinuxIfIndexRW, stopwatch *measure.Stopwatch) error {
	// Logger
	plugin.log = logger.NewLogger("-route-conf")
	plugin.log.Debug("Initializing Linux Route configurator")

	// Mappings
	plugin.ifIndexes = ifIndexes
	plugin.rtIndexes = l3idx.NewLinuxRouteIndex(nametoidx.NewNameToIdx(plugin.log, "linux_route_indexes", nil))
	plugin.rtAutoIndexes = l3idx.NewLinuxRouteIndex(nametoidx.NewNameToIdx(plugin.log, "linux_auto_route_indexes", nil))
	plugin.rtCachedIfRoutes = l3idx.NewLinuxRouteIndex(nametoidx.NewNameToIdx(plugin.log, "linux_cached_route_indexes", nil))
	plugin.rtCachedGwRoutes = make(map[string]*l3.LinuxStaticRoutes_Route)

	// L3 and namespace handler
	plugin.l3Handler = l3Handler
	plugin.nsHandler = nsHandler

	// Configurator-wide stopwatch instance
	plugin.stopwatch = stopwatch

	return nil
}

// Close does nothing for route configurator
func (plugin *LinuxRouteConfigurator) Close() error {
	return nil
}

// GetRouteIndexes returns route in-memory indexes
func (plugin *LinuxRouteConfigurator) GetRouteIndexes() l3idx.LinuxRouteIndexRW {
	return plugin.rtIndexes
}

// GetAutoRouteIndexes returns automatic route in-memory indexes
func (plugin *LinuxRouteConfigurator) GetAutoRouteIndexes() l3idx.LinuxRouteIndexRW {
	return plugin.rtAutoIndexes
}

// GetCachedRoutes returns cached route in-memory indexes
func (plugin *LinuxRouteConfigurator) GetCachedRoutes() l3idx.LinuxRouteIndexRW {
	return plugin.rtCachedIfRoutes
}

// GetCachedGatewayRoutes returns in-memory indexes of unreachable gateway routes
func (plugin *LinuxRouteConfigurator) GetCachedGatewayRoutes() map[string]*l3.LinuxStaticRoutes_Route {
	return plugin.rtCachedGwRoutes
}

// ConfigureLinuxStaticRoute reacts to a new northbound Linux static route config by creating and configuring
// the route in the host network stack through Netlink API.
func (plugin *LinuxRouteConfigurator) ConfigureLinuxStaticRoute(route *l3.LinuxStaticRoutes_Route) error {
	plugin.log.Infof("Configuring linux static route %s", route.Name)

	// Prepare route object
	netLinkRoute := &netlink.Route{}

	if route.Interface != "" {
		// Find interface
		_, ifData, foundIface := plugin.ifIndexes.LookupIdx(route.Interface)
		if !foundIface || ifData == nil {
			plugin.log.Infof("Static route %s requires non-existing interface %s, moving to cache", route.Name, route.Interface)
			plugin.rtCachedIfRoutes.RegisterName(route.Name, plugin.rtIdxSeq, route)
			plugin.rtIdxSeq++
			return nil
		}
		netLinkRoute.LinkIndex = int(ifData.Index)
	}

	// Check gateway reachability
	if route.Default || route.GwAddr != "" {
		if !plugin.networkReachable(route.Namespace, route.GwAddr) {
			plugin.rtCachedGwRoutes[route.Name] = route
			plugin.log.Debugf("Default/Gateway route %s cached, gateway address %s is currently unreachable",
				route.Name, route.GwAddr)
			return nil
		}
	}

	// Check if route was not cached before, eventually remove it
	_, ok := plugin.rtCachedGwRoutes[route.Name]
	if ok {
		delete(plugin.rtCachedGwRoutes, route.Name)
	}

	// Default route
	if route.Default {
		err := plugin.createDefaultRoute(netLinkRoute, route)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
	} else {
		// Static route
		err := plugin.createStaticRoute(netLinkRoute, route)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
	}

	// Prepare and switch to namespace where the route belongs
	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()
	routeNs := plugin.nsHandler.RouteNsToGeneric(route.Namespace)
	revertNs, err := plugin.nsHandler.SwitchNamespace(routeNs, nsMgmtCtx)
	if err != nil {
		plugin.log.Error(err)
		return err
	}
	defer revertNs()

	err = plugin.l3Handler.AddStaticRoute(route.Name, netLinkRoute)
	if err != nil {
		plugin.log.Errorf("adding static route %s failed: %v (%+v)", route.Name, err, netLinkRoute)
		return err
	}

	plugin.rtIndexes.RegisterName(RouteIdentifier(netLinkRoute), plugin.rtIdxSeq, route)
	plugin.rtIdxSeq++
	plugin.log.Debugf("Route %s registered", route.Name)

	plugin.log.Infof("Linux static route %s configured", route.Name)

	// Retry default routes if some of them is not configurable now
	if !route.Default {
		plugin.retryDefaultRoutes(route)
	}

	return nil
}

// ModifyLinuxStaticRoute applies changes in the NB configuration of a Linux static route into the host network stack
// through Netlink API.
func (plugin *LinuxRouteConfigurator) ModifyLinuxStaticRoute(newRoute *l3.LinuxStaticRoutes_Route, oldRoute *l3.LinuxStaticRoutes_Route) error {
	plugin.log.Infof("Modifying linux static route %s", newRoute.Name)
	var err error

	// Prepare route object
	netLinkRoute := &netlink.Route{}

	if newRoute.Interface != "" {
		// Find interface
		_, ifData, foundIface := plugin.ifIndexes.LookupIdx(newRoute.Interface)
		if !foundIface || ifData == nil {
			plugin.log.Infof("Modified static route %s requires non-existing interface %s, moving to cache", newRoute.Name, newRoute.Interface)
			plugin.rtCachedIfRoutes.RegisterName(newRoute.Name, plugin.rtIdxSeq, newRoute)
			plugin.rtIdxSeq++
			return nil
		}
		netLinkRoute.LinkIndex = int(ifData.Index)
	}

	// Check gateway reachability
	if newRoute.Default || newRoute.GwAddr != "" {
		if !plugin.networkReachable(newRoute.Namespace, newRoute.GwAddr) {
			plugin.rtCachedGwRoutes[newRoute.Name] = newRoute
			plugin.log.Debugf("Default/Gateway route %s cached, gateway address %s is currently unreachable",
				newRoute.Name, newRoute.GwAddr)
			return nil
		}
	}

	// Check if route was not cached before, eventually remove it
	_, ok := plugin.rtCachedGwRoutes[newRoute.Name]
	if ok {
		delete(plugin.rtCachedGwRoutes, newRoute.Name)
	}

	// If the namespace of the new route was changed, the old route needs to be removed and the new one created in the
	// new namespace
	// If interface or destination IP address was changed, the old entry needs to be removed and recreated as well.
	// Otherwise, ModifyRouteEntry (analogy to 'ip route replace') would create a new route instead of modifying
	// the existing one
	var replace bool

	oldRouteNs := plugin.nsHandler.RouteNsToGeneric(oldRoute.Namespace)
	newRouteNs := plugin.nsHandler.RouteNsToGeneric(newRoute.Namespace)
	result := oldRouteNs.CompareNamespaces(newRouteNs)
	if result != 0 || oldRoute.Interface != newRoute.Interface {
		replace = true
	}

	// Default route
	if newRoute.Default {
		if !oldRoute.Default {
			// In this case old route has to be removed
			replace = true
		}
		err := plugin.createDefaultRoute(netLinkRoute, newRoute)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
	} else {
		if oldRoute.DstIpAddr != newRoute.Interface {
			replace = true
		}
		if err = plugin.createStaticRoute(netLinkRoute, newRoute); err != nil {
			plugin.log.Error(err)
			return err
		}
	}

	// Static route will be removed and created anew
	if replace {
		return plugin.recreateLinuxStaticRoute(netLinkRoute, newRoute)
	}

	// Prepare namespace of related interface
	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()
	routeNs := plugin.nsHandler.RouteNsToGeneric(newRoute.Namespace)

	// route has to be created in the same namespace as the interface
	revertNs, err := plugin.nsHandler.SwitchNamespace(routeNs, nsMgmtCtx)
	if err != nil {
		plugin.log.Error(err)
		return err
	}
	defer revertNs()

	// Remove old route and create a new one
	if err = plugin.DeleteLinuxStaticRoute(oldRoute); err != nil {
		plugin.log.Errorf("deleting static route %s failed: %v (%+v)", oldRoute.Name, err, oldRoute)
		return err
	}
	if err = plugin.l3Handler.AddStaticRoute(newRoute.Name, netLinkRoute); err != nil {
		plugin.log.Errorf("adding static route %s failed: %v (%+v)", newRoute.Name, err, netLinkRoute)
		return err
	}

	plugin.log.Infof("Linux static route %s modified", newRoute.Name)

	// Retry default routes if some of them is not configurable
	if !newRoute.Default {
		plugin.retryDefaultRoutes(newRoute)
	}

	return nil
}

// DeleteLinuxStaticRoute reacts to a removed NB configuration of a Linux static route entry.
func (plugin *LinuxRouteConfigurator) DeleteLinuxStaticRoute(route *l3.LinuxStaticRoutes_Route) error {
	plugin.log.Infof("Removing linux static route %s", route.Name)
	var err error

	// Check if route is in cache waiting on interface
	if _, _, found := plugin.rtCachedIfRoutes.LookupIdx(route.Name); found {
		plugin.rtCachedIfRoutes.UnregisterName(route.Name)
		plugin.log.Debugf("Route %s removed from interface cache", route.Name)
		return nil
	}
	// Check if route is in cache waiting for gateway address reachability
	for _, cachedRoute := range plugin.rtCachedGwRoutes {
		if cachedRoute.Name == route.Name {
			delete(plugin.rtCachedGwRoutes, cachedRoute.Name)
			plugin.log.Debugf("Route %s removed from gw cache", route.Name)
			return nil
		}
	}

	// Prepare route object
	netLinkRoute := &netlink.Route{}

	if route.Interface != "" {
		// Find interface
		_, ifData, foundIface := plugin.ifIndexes.LookupIdx(route.Interface)
		if !foundIface || ifData == nil {
			return fmt.Errorf("cannot delete static route %s, interface %s not found", route.Name, route.Interface)
		}
		netLinkRoute.LinkIndex = int(ifData.Index)
	}

	// Destination IP address
	if route.DstIpAddr != "" {
		addressWithPrefix := strings.Split(route.DstIpAddr, "/")
		if len(addressWithPrefix) > 1 {
			dstIPAddr := &net.IPNet{}
			_, dstIPAddr, err = net.ParseCIDR(route.DstIpAddr)
			if err != nil {
				plugin.log.Error(err)
				return err
			}
			netLinkRoute.Dst = dstIPAddr
		} else {
			plugin.log.Error("static route's dst address mask not set, route %s may not be removed", route.Name)
		}
	}
	// Gateway IP address
	if route.GwAddr != "" {
		gateway := net.ParseIP(route.GwAddr)
		if gateway != nil {
			netLinkRoute.Gw = gateway
		} else {
			plugin.log.Error("static route's gateway address %s has incorrect format, route %s may not be removed",
				route.GwAddr, route.Name)
		}
	}
	if netLinkRoute.Dst == nil && netLinkRoute.Gw == nil {
		return fmt.Errorf("cannot delete static route %s, requred at least destination or gateway address", route.Name)
	}

	// Scope
	if route.Scope != nil {
		netLinkRoute.Scope = plugin.parseRouteScope(route.Scope)
	}

	// Prepare and switch to the namespace where the route belongs
	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()
	routeNs := plugin.nsHandler.RouteNsToGeneric(route.Namespace)
	revertNs, err := plugin.nsHandler.SwitchNamespace(routeNs, nsMgmtCtx)
	if err != nil {
		plugin.log.Error(err)
		return err
	}
	defer revertNs()

	err = plugin.l3Handler.DelStaticRoute(route.Name, netLinkRoute)
	if err != nil {
		plugin.log.Errorf("deleting static route %q failed: %v (%+v)", route.Name, err, netLinkRoute)
		return err
	}

	_, _, found := plugin.rtIndexes.UnregisterName(RouteIdentifier(netLinkRoute))
	if !found {
		plugin.log.Warnf("Attempt to unregister non-registered route %s", route.Name)
	}
	plugin.log.Debugf("Route %s unregistered", route.Name)

	plugin.log.Infof("Linux static route %s removed", route.Name)

	return nil
}

// ResolveCreatedInterface manages static routes for new interface. Linux interface also creates its own route which
// can make other routes accessible and ready to create - the case is also resolved here.
func (plugin *LinuxRouteConfigurator) ResolveCreatedInterface(name string, index uint32) error {
	plugin.log.Infof("Linux static route configurator: resolve new interface %s (idx %d)", name, index)

	// Search mapping for cached routes using the new interface
	cachedIfRoutes := plugin.rtCachedIfRoutes.LookupNamesByInterface(name)
	if len(cachedIfRoutes) > 0 {
		plugin.log.Debugf("Found %d cached routes for interface %s", len(cachedIfRoutes), name)
		// Store default routes, they have to be configured as the last ones
		var defRoutes []*l3.LinuxStaticRoutes_Route
		// Static routes
		for _, cachedRoute := range cachedIfRoutes {
			if cachedRoute.Default {
				defRoutes = append(defRoutes, cachedRoute)
				continue
			}
			if err := plugin.ConfigureLinuxStaticRoute(cachedRoute); err != nil {
				plugin.log.Warn(err)
				return err
			}
			// Remove from cache
			plugin.rtCachedIfRoutes.UnregisterName(cachedRoute.Name)
		}
		// Default routes
		for _, cachedDefaultRoute := range defRoutes {
			if err := plugin.ConfigureLinuxStaticRoute(cachedDefaultRoute); err != nil {
				plugin.log.Warn(err)
				return err
			}
			// Remove from cache
			plugin.rtCachedIfRoutes.UnregisterName(cachedDefaultRoute.Name)
		}
	}

	// Interface also created its own route, so try to re-configure default routes
	err := plugin.processAutoRoutes(name, index)
	if err != nil {
		plugin.log.Error(err)
	}

	// Try to reconfigure cached gateway routes
	if len(plugin.rtCachedGwRoutes) > 0 {
		plugin.log.Debugf("Found %d cached gateway routes", len(cachedIfRoutes))
		// Store default routes, they have to be configured as the last ones
		defRoutes := make(map[string]*l3.LinuxStaticRoutes_Route)
		for _, cachedRoute := range plugin.rtCachedGwRoutes {
			// Check accessibility
			if !plugin.networkReachable(cachedRoute.Namespace, cachedRoute.GwAddr) {
				continue
			} else {
			}
			if cachedRoute.Default {
				defRoutes[cachedRoute.Name] = cachedRoute
				continue
			}
			if err := plugin.ConfigureLinuxStaticRoute(cachedRoute); err != nil {
				plugin.log.Warn(err)
				return err
			}
			// Remove from cache
			delete(plugin.rtCachedGwRoutes, cachedRoute.Name)
		}
		// Default routes
		for _, cachedDefaultRoute := range defRoutes {
			if err := plugin.ConfigureLinuxStaticRoute(cachedDefaultRoute); err != nil {
				plugin.log.Warn(err)
				return err
			}
			// Remove from cache
			delete(plugin.rtCachedGwRoutes, cachedDefaultRoute.Name)
		}
	}

	return err
}

// ResolveDeletedInterface manages static routes for removed interface
func (plugin *LinuxRouteConfigurator) ResolveDeletedInterface(name string, index uint32) error {
	plugin.log.Infof("Linux static route configurator: resolve deleted interface %v (idx %d)", name, index)

	// Search mapping for configured linux routes using the new interface
	confRoutes := plugin.rtIndexes.LookupNamesByInterface(name)
	if len(confRoutes) > 0 {
		plugin.log.Debugf("Found %d routes belonging to the removed interface %s", len(confRoutes), name)
		for _, rt := range confRoutes {
			// Add to un-configured. If the interface will be recreated, all routes are configured back
			plugin.rtCachedIfRoutes.RegisterName(rt.Name, plugin.rtIdxSeq, rt)
			plugin.rtIdxSeq++
		}
	}

	return nil
}

// RouteIdentifier generates unique route ID used in mapping
func RouteIdentifier(route *netlink.Route) string {
	if route.Dst == nil || route.Dst.String() == ipv4AddrAny || route.Dst.String() == ipv6AddrAny {
		return fmt.Sprintf("default-iface%d-table%v-%s", route.LinkIndex, route.Table, route.Gw.To4().String())
	}
	return fmt.Sprintf("dst%s-iface%d-table%v-%s", route.Dst.IP.String(), route.LinkIndex, route.Table, route.Gw.String())
}

// Create default route object with gateway address. Destination address has to be set in such a case
func (plugin *LinuxRouteConfigurator) createDefaultRoute(netLinkRoute *netlink.Route, route *l3.LinuxStaticRoutes_Route) (err error) {
	// Gateway
	gateway := net.ParseIP(route.GwAddr)
	if gateway == nil {
		return fmt.Errorf("unable to create route %s as default, gateway is nil", route.Name)
	}
	netLinkRoute.Gw = gateway

	// Destination address
	dstIPAddr := route.DstIpAddr
	if dstIPAddr == "" {
		dstIPAddr = ipv4AddrAny
	}
	if dstIPAddr != ipv4AddrAny && dstIPAddr != ipv6AddrAny {
		plugin.log.Warnf("route marked as default has dst address set to %s. The address will be ignored", dstIPAddr)
		dstIPAddr = ipv4AddrAny
	}
	_, netLinkRoute.Dst, err = net.ParseCIDR(dstIPAddr)
	if err != nil {
		plugin.log.Error(err)
		return err
	}

	// Priority
	if route.Metric != 0 {
		netLinkRoute.Priority = int(route.Metric)
	}

	plugin.log.Debugf("Creating default route with gw ip %s", netLinkRoute.Gw)
	return nil
}

// Create static route from provided data
func (plugin *LinuxRouteConfigurator) createStaticRoute(netLinkRoute *netlink.Route, route *l3.LinuxStaticRoutes_Route) error {
	var err error
	// Destination IP address
	if route.DstIpAddr != "" {
		addressWithPrefix := strings.Split(route.DstIpAddr, "/")
		dstIPAddr := &net.IPNet{}
		if len(addressWithPrefix) > 1 {
			_, dstIPAddr, err = net.ParseCIDR(route.DstIpAddr)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("cannot create static route %s, dst address net mask not set", route.Name)
		}
		plugin.log.Infof("IP address %s set as dst for route %s", route.DstIpAddr, route.Name)
		netLinkRoute.Dst = dstIPAddr
	} else {
		return fmt.Errorf("cannot create static route %s, destination addres not set", route.Name)
	}

	// Set gateway if exists
	gateway := net.ParseIP(route.GwAddr)
	if gateway != nil {
		netLinkRoute.Gw = gateway
		plugin.log.Infof("Gateway address %s set for route %s", route.GwAddr, route.Name)
	}

	// Source IP address is exists
	srcIPAddr := net.ParseIP(route.SrcIpAddr)
	if srcIPAddr != nil {
		netLinkRoute.Src = srcIPAddr
		plugin.log.Infof("IP address %s set as src for route %s", route.SrcIpAddr, route.Name)
	}

	// Scope
	if route.Scope != nil {
		netLinkRoute.Scope = plugin.parseRouteScope(route.Scope)
	}

	// Priority
	if route.Metric != 0 {
		netLinkRoute.Priority = int(route.Metric)
	}

	// Table
	netLinkRoute.Table = int(route.Table)

	plugin.log.Debugf("Creating static route with destination ip %s", netLinkRoute.Dst)
	return nil
}

// Update linux static route using modify (analogy to 'ip route replace')
func (plugin *LinuxRouteConfigurator) recreateLinuxStaticRoute(netLinkRoute *netlink.Route, route *l3.LinuxStaticRoutes_Route) error {
	plugin.log.Debugf("Route %s modification caused the route to be removed and crated again", route.Name)
	// Prepare namespace of related interface
	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()
	routeNs := plugin.nsHandler.RouteNsToGeneric(route.Namespace)

	// route has to be created in the same namespace as the interface
	revertNs, err := plugin.nsHandler.SwitchNamespace(routeNs, nsMgmtCtx)
	if err != nil {
		plugin.log.Error(err)
		return err
	}
	defer revertNs()

	// Update existing route
	return plugin.l3Handler.ReplaceStaticRoute(route.Name, netLinkRoute)
}

// Tries to configure again cached default/gateway routes (as a reaction to the new route)
func (plugin *LinuxRouteConfigurator) retryDefaultRoutes(route *l3.LinuxStaticRoutes_Route) {
	plugin.log.Debug("Retrying to configure default routes")
	for _, defRoute := range plugin.rtCachedGwRoutes {
		// Filter routes from different namespaces
		if defRoute.Namespace != nil && route.Namespace == nil || defRoute.Namespace == nil && route.Namespace != nil {
			continue
		}
		if defRoute.Namespace != nil && route.Namespace != nil && defRoute.Namespace.Name != route.Namespace.Name {
			continue
		}

		// Parse gateway and default address
		gwIPParsed := net.ParseIP(defRoute.GwAddr)
		_, dstNet, err := net.ParseCIDR(route.DstIpAddr)
		if err != nil {
			plugin.log.Errorf("Error parsing IP address %s: %v", route.DstIpAddr, err)
			continue
		}

		if dstNet.Contains(gwIPParsed) {
			// Default/Gateway route can be now configured
			if err := plugin.ConfigureLinuxStaticRoute(defRoute); err != nil {
				plugin.log.Errorf("Error while configuring route %s: %v", route.Name, err)
			}
			delete(plugin.rtCachedGwRoutes, defRoute.Name)
		} else {
			plugin.log.Debugf("%s is not within %s", defRoute.GwAddr, dstNet.IP)
		}
	}
}

// Handles automatic route created by adding interface. Method look for routes related to the interface and its
// IP address in its namespace.
// Note: read route's destination address does not contain mask. This value is determined from interfaces' IP address.
// Automatic routes are store in separate mapping and their names are generated.
func (plugin *LinuxRouteConfigurator) processAutoRoutes(ifName string, ifIdx uint32) error {
	plugin.log.Debugf("Processing automatic interfaces for %s", ifName)
	// Look for metadata
	_, ifData, found := plugin.ifIndexes.LookupIdx(ifName)
	if !found {
		return fmt.Errorf("interface %s not found in the mapping", ifName)
	}
	if ifData == nil || ifData.Data == nil {
		return fmt.Errorf("interface %s data not found in the mapping", ifName)
	}

	// Move to interface with the interface
	if ifData.Data.Namespace != nil {
		nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()
		// Switch to namespace
		ifNs := plugin.nsHandler.IfNsToGeneric(ifData.Data.Namespace)
		revertNs, err := plugin.nsHandler.SwitchNamespace(ifNs, nsMgmtCtx)
		if err != nil {
			return fmt.Errorf("RESYNC Linux route %s: failed to switch to namespace %s: %v",
				ifData.Data.Name, ifData.Data.Namespace.Name, err)
		}
		defer revertNs()
	}

	// Get interface
	link, err := netlink.LinkByName(ifData.Data.HostIfName)
	if err != nil {
		return fmt.Errorf("cannot read linux interface %s (host %s): %v", ifName, ifData.Data.HostIfName, err)
	}

	// Read all routes belonging to the interface
	linuxRts, err := netlink.RouteList(link, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("cannot read linux routes for interface %s (host %s): %v", ifName, ifData.Data.HostIfName, err)
	}

	// Iterate over link addresses and look for ones related to t
	for rtIdx, linuxRt := range linuxRts {
		if linuxRt.Dst == nil {
			continue
		}
		route := plugin.transformRoute(linuxRt, ifData.Data.HostIfName)
		// Route's destination address is read without mask. Use interface data to fill it.
		var routeFound bool
		for ipIdx, ifIP := range ifData.Data.IpAddresses {
			_, ifDst, err := net.ParseCIDR(ifIP)
			if err != nil {
				return err
			}
			if bytes.Compare(linuxRt.Dst.IP, ifDst.IP) == 0 {
				// Transform destination IP and namespace
				route.DstIpAddr = ifData.Data.IpAddresses[ipIdx]
				route.Namespace = transformNamespace(ifData.Data.Namespace)
				routeFound = true
			}
		}
		if !routeFound {
			plugin.log.Debugf("Route with IP %s skipped", linuxRt.Dst.IP.String())
			continue
		}
		// Generate name
		route.Name = ifName + strconv.Itoa(rtIdx)
		// In case there is obsolete route with the same name, remove it
		plugin.rtAutoIndexes.UnregisterName(route.Name)
		plugin.rtAutoIndexes.RegisterName(route.Name, plugin.rtIdxSeq, route)
		plugin.rtIdxSeq++

		// Also try to configure default routes
		plugin.retryDefaultRoutes(route)
	}

	return nil
}

// Transform linux netlink route type to proto message type
func (plugin *LinuxRouteConfigurator) transformRoute(linuxRt netlink.Route, ifName string) *l3.LinuxStaticRoutes_Route {
	var dstAddr, srcAddr, gwAddr string
	// Destination address
	if linuxRt.Dst != nil {
		// Transform only IP (without mask)
		dstAddr = linuxRt.Dst.IP.String()
	}
	// Source address
	if linuxRt.Src != nil {
		srcAddr = linuxRt.Src.String()
	}
	// Gateway address
	if linuxRt.Gw != nil {
		gwAddr = linuxRt.Gw.String()
	}

	if dstAddr == "" || dstAddr == ipv4AddrAny || dstAddr == ipv6AddrAny {
		// Default route
		return &l3.LinuxStaticRoutes_Route{
			Default:   true,
			Interface: ifName,
			GwAddr:    gwAddr,
			Metric:    uint32(linuxRt.Priority),
		}
	}
	// Static route
	return &l3.LinuxStaticRoutes_Route{
		Interface: ifName,
		DstIpAddr: dstAddr,
		SrcIpAddr: srcAddr,
		GwAddr:    gwAddr,
		Scope:     plugin.parseLinuxRouteScope(linuxRt.Scope),
		Metric:    uint32(linuxRt.Priority),
		Table:     uint32(linuxRt.Table),
	}
}

// Interface namespace type -> route namespace type
func transformNamespace(ifNs *interfaces.LinuxInterfaces_Interface_Namespace) *l3.LinuxStaticRoutes_Route_Namespace {
	if ifNs == nil {
		return nil
	}
	return &l3.LinuxStaticRoutes_Route_Namespace{
		Type: func(ifType interfaces.LinuxInterfaces_Interface_Namespace_NamespaceType) l3.LinuxStaticRoutes_Route_Namespace_NamespaceType {
			switch ifType {
			case interfaces.LinuxInterfaces_Interface_Namespace_PID_REF_NS:
				return l3.LinuxStaticRoutes_Route_Namespace_PID_REF_NS
			case interfaces.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS:
				return l3.LinuxStaticRoutes_Route_Namespace_MICROSERVICE_REF_NS
			case interfaces.LinuxInterfaces_Interface_Namespace_NAMED_NS:
				return l3.LinuxStaticRoutes_Route_Namespace_NAMED_NS
			case interfaces.LinuxInterfaces_Interface_Namespace_FILE_REF_NS:
				return l3.LinuxStaticRoutes_Route_Namespace_FILE_REF_NS
			default:
				return l3.LinuxStaticRoutes_Route_Namespace_PID_REF_NS
			}
		}(ifNs.Type),
		Pid:          ifNs.Pid,
		Microservice: ifNs.Microservice,
		Name:         ifNs.Name,
		Filepath:     ifNs.Filepath,
	}
}

// Agent route scope -> netlink route scope
func (plugin *LinuxRouteConfigurator) parseRouteScope(scope *l3.LinuxStaticRoutes_Route_Scope) netlink.Scope {
	switch scope.Type {
	case l3.LinuxStaticRoutes_Route_Scope_GLOBAL:
		return netlink.SCOPE_UNIVERSE
	case l3.LinuxStaticRoutes_Route_Scope_HOST:
		return netlink.SCOPE_HOST
	case l3.LinuxStaticRoutes_Route_Scope_LINK:
		return netlink.SCOPE_LINK
	case l3.LinuxStaticRoutes_Route_Scope_SITE:
		return netlink.SCOPE_SITE
	default:
		plugin.log.Infof("Unknown scope type, setting to default (link): %v", scope.Type)
		return netlink.SCOPE_LINK
	}
}

// Verifies whether address network is reachable.
func (plugin *LinuxRouteConfigurator) networkReachable(ns *l3.LinuxStaticRoutes_Route_Namespace, ipAddress string) bool {
	// Try for registered configuration routes
	registeredRoute, err := plugin.rtIndexes.LookupRouteByIP(ns, ipAddress)
	if err != nil {
		plugin.log.Errorf("Failed to resolve accessibility of %s (registered): %v", ipAddress, err)
	}
	// Try for registered automatic (interface-added) routes
	autoRoute, err := plugin.rtAutoIndexes.LookupRouteByIP(ns, ipAddress)
	if err != nil {
		plugin.log.Errorf("Failed to resolve accessibility of %s (auto): %v", ipAddress, err)
	}
	if registeredRoute != nil || autoRoute != nil {
		plugin.log.Debugf("Network %s is reachable", ipAddress)
		return true
	}
	return false
}
