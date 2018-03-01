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

//go:generate protoc --proto_path=../common/model/l3 --gogo_out=../common/model/l3 ../common/model/l3/l3.proto

package l3plugin

import (
	"fmt"
	"net"
	"strings"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/l3"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/ifaceidx"
	linuxcalls2 "github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/linuxcalls"
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
	Log logging.Logger

	LinuxIfIdx      ifaceidx.LinuxIfIndexRW
	RouteIdxSeq     uint32
	rtIndexes       l3idx.LinuxRouteIndexRW
	rtCachedIndexes l3idx.LinuxRouteIndexRW
	// cache for default routes which cannot be created due to unreachable network
	rtCachedRoutes map[string]*l3.LinuxStaticRoutes_Route

	// Time measurement
	Stopwatch *measure.Stopwatch // timer used to measure and store time

}

// Init initializes static route configurator and starts goroutines
func (plugin *LinuxRouteConfigurator) Init(rtIndexes l3idx.LinuxRouteIndexRW, rtCachedIndexes l3idx.LinuxRouteIndexRW) error {
	plugin.Log.Debug("Initializing Linux Route configurator")
	plugin.rtIndexes = rtIndexes
	plugin.rtCachedIndexes = rtCachedIndexes

	// Default route cache
	plugin.rtCachedRoutes = make(map[string]*l3.LinuxStaticRoutes_Route)

	return nil
}

// Close closes all goroutines started during Init
func (plugin *LinuxRouteConfigurator) Close() error {
	return nil
}

// ConfigureLinuxStaticRoute reacts to a new northbound Linux static route config by creating and configuring
// the route in the host network stack through Netlink API.
func (plugin *LinuxRouteConfigurator) ConfigureLinuxStaticRoute(route *l3.LinuxStaticRoutes_Route) error {
	plugin.Log.Infof("Configuring linux static route %v", route.Name)

	// Prepare route object
	netLinkRoute := &netlink.Route{}

	if route.Interface != "" {
		// Find interface
		idx, _, foundIface := plugin.LinuxIfIdx.LookupIdx(route.Interface)
		if !foundIface {
			plugin.Log.Infof("Static route %v requires non-existing interface %v, moving to cache", route.Name, route.Interface)
			plugin.rtCachedIndexes.RegisterName(route.Name, plugin.RouteIdxSeq, route)
			plugin.RouteIdxSeq++
			return nil
		}
		netLinkRoute.LinkIndex = int(idx)
	}

	// default route
	if route.Default {
		cached, err := plugin.createDefaultRoute(netLinkRoute, route)
		if err != nil {
			plugin.Log.Error(err)
			return err
		}
		if cached {
			// If route was cached, skip rest of the resolution
			return nil
		}
	} else {
		// static route
		err := plugin.createStaticRoute(netLinkRoute, route)
		if err != nil {
			plugin.Log.Error(err)
			return err
		}
	}

	// Prepare namespace of related interface
	nsMgmtCtx := linuxcalls2.NewNamespaceMgmtCtx()
	routeNs := linuxcalls.ToGenericRouteNs(route.Namespace)

	// route has to be created in the same namespace as the interface
	revertNs, err := routeNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
	if err != nil {
		plugin.Log.Error(err)
		return err
	}
	defer revertNs()

	err = linuxcalls.AddStaticRoute(route.Name, netLinkRoute, plugin.Log, measure.GetTimeLog("add-linux-route", plugin.Stopwatch))
	if err != nil {
		plugin.Log.Errorf("adding static route %q failed: %v (%+v)", route.Name, err, netLinkRoute)
		return err
	}

	plugin.rtIndexes.RegisterName(routeIdentifier(netLinkRoute), plugin.RouteIdxSeq, route)
	plugin.RouteIdxSeq++
	plugin.Log.Debugf("Route %v registered", route.Name)

	plugin.Log.Infof("Linux static route %v configured", route.Name)

	// Retry default routes if some of them is not configurable
	if !route.Default {
		plugin.retryDefaultRoutes(route)
	}

	return nil
}

// ModifyLinuxStaticRoute applies changes in the NB configuration of a Linux static route into the host network stack
// through Netlink API.
func (plugin *LinuxRouteConfigurator) ModifyLinuxStaticRoute(newRoute *l3.LinuxStaticRoutes_Route, oldRoute *l3.LinuxStaticRoutes_Route) error {
	plugin.Log.Infof("Modifying linux static route %v", newRoute.Name)
	var err error

	// Prepare route object
	netLinkRoute := &netlink.Route{}

	if newRoute.Interface != "" {
		// Find interface
		idx, _, foundIface := plugin.LinuxIfIdx.LookupIdx(newRoute.Interface)
		if !foundIface {
			plugin.Log.Infof("Modified static route %v requires non-existing interface %v, moving to cache", newRoute.Name, newRoute.Interface)
			plugin.rtCachedIndexes.RegisterName(newRoute.Name, plugin.RouteIdxSeq, newRoute)
			plugin.RouteIdxSeq++
			return nil
		}
		netLinkRoute.LinkIndex = int(idx)
	}

	// If the namespace of the new route was changed, the old route needs to be removed and the new one created in the
	// new namespace
	// If interface or destination IP address was changed, the old entry needs to be removed and recreated as well.
	// In such a case, ModifyRouteEntry (analogy to 'ip route replace') would create a new route instead of modifying
	// the existing one
	replace := true

	oldRouteNs := linuxcalls.ToGenericRouteNs(oldRoute.Namespace)
	newRouteNs := linuxcalls.ToGenericRouteNs(newRoute.Namespace)
	result := oldRouteNs.CompareNamespaces(newRouteNs)
	if result != 0 || oldRoute.Interface != newRoute.Interface {
		replace = false
	}

	// Default route
	if newRoute.Default {
		if !oldRoute.Default {
			// In this case old route has to be removed
			replace = false
		}
		cached, err := plugin.createDefaultRoute(netLinkRoute, newRoute)
		if err != nil {
			plugin.Log.Error(err)
			return err
		}
		if cached {
			// If route was cached, skip rest of the resolution
			return nil
		}
	} else {
		if oldRoute.DstIpAddr != newRoute.Interface {
			replace = false
		}
		if err = plugin.createStaticRoute(netLinkRoute, newRoute); err != nil {
			plugin.Log.Error(err)
			return err
		}
	}

	if replace {
		return plugin.updateLinuxStaticRoute(netLinkRoute, newRoute)
	}

	// Prepare namespace of related interface
	nsMgmtCtx := linuxcalls2.NewNamespaceMgmtCtx()
	routeNs := linuxcalls.ToGenericRouteNs(newRoute.Namespace)

	// route has to be created in the same namespace as the interface
	revertNs, err := routeNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
	if err != nil {
		plugin.Log.Error(err)
		return err
	}
	defer revertNs()

	// Remove old route and create a new one
	if err = plugin.DeleteLinuxStaticRoute(oldRoute); err != nil {
		plugin.Log.Errorf("deleting static route %q failed: %v (%+v)", oldRoute.Name, err, oldRoute)
		return err
	}
	if err = linuxcalls.AddStaticRoute(newRoute.Name, netLinkRoute, plugin.Log, measure.GetTimeLog("add-linux-route", plugin.Stopwatch)); err != nil {
		plugin.Log.Errorf("adding static route %q failed: %v (%+v)", newRoute.Name, err, netLinkRoute)
		return err
	}

	plugin.Log.Infof("Linux static route %v modified", newRoute.Name)

	// Retry default routes if some of them is not configurable
	if !newRoute.Default {
		plugin.retryDefaultRoutes(newRoute)
	}

	return nil
}

// DeleteLinuxStaticRoute reacts to a removed NB configuration of a Linux static route entry.
func (plugin *LinuxRouteConfigurator) DeleteLinuxStaticRoute(route *l3.LinuxStaticRoutes_Route) error {
	plugin.Log.Infof("Removing linux static route %v", route.Name)
	var err error

	// Prepare route object
	netLinkRoute := &netlink.Route{}

	if route.Interface != "" {
		// Find interface
		idx, _, foundIface := plugin.LinuxIfIdx.LookupIdx(route.Interface)
		if !foundIface {
			return fmt.Errorf("cannot delete static route %v, interface %v not found", route.Name, route.Interface)
		}
		netLinkRoute.LinkIndex = int(idx)
	}

	// Destination IP address
	if route.DstIpAddr != "" {
		addressWithPrefix := strings.Split(route.DstIpAddr, "/")
		dstIPAddr := &net.IPNet{}
		if len(addressWithPrefix) > 1 {
			_, dstIPAddr, err = net.ParseCIDR(route.DstIpAddr)
			if err != nil {
				plugin.Log.Error(err)
				return err
			}
		} else {
			return fmt.Errorf("cannot remove static route %v, dst address net mask not set", route.Name)
		}
		netLinkRoute.Dst = dstIPAddr
	} else if route.GwAddr != "" {
		gateway := net.ParseIP(route.GwAddr)
		if gateway == nil {
			return fmt.Errorf("cannot remove static route %v, gateway address has incorrect format: %v",
				route.Name, route.GwAddr)
		}
		netLinkRoute.Gw = gateway
	} else {
		return fmt.Errorf("cannot remove static route %v, destination/gateway address not available", route.Name)
	}

	// Scope
	netLinkRoute.Scope = plugin.parseRouteScope(route.Scope)

	// Prepare namespace of related interface
	nsMgmtCtx := linuxcalls2.NewNamespaceMgmtCtx()
	routeNs := linuxcalls.ToGenericRouteNs(route.Namespace)

	// route has to be created in the same namespace as the interface
	revertNs, err := routeNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
	if err != nil {
		plugin.Log.Error(err)
		return err
	}
	defer revertNs()

	err = linuxcalls.DeleteStaticRoute(route.Name, netLinkRoute, plugin.Log, measure.GetTimeLog("del-linux-route", plugin.Stopwatch))
	if err != nil {
		plugin.Log.Errorf("deleting static route %q failed: %v (%+v)", route.Name, err, netLinkRoute)
		return err
	}

	_, _, found := plugin.rtIndexes.UnregisterName(routeIdentifier(netLinkRoute))
	if !found {
		plugin.Log.Warnf("Attempt to unregister non-registered route %v", route.Name)
	}
	plugin.Log.Debugf("Route %v unregistered", route.Name)

	plugin.Log.Infof("Linux static route %v removed", route.Name)

	return nil
}

// LookupLinuxRoutes reads all routes and registers them if needed
func (plugin *LinuxRouteConfigurator) LookupLinuxRoutes() error {
	plugin.Log.Infof("Browsing Linux routes")

	// read all routes
	routes, err := linuxcalls.ReadStaticRoutes(nil, noFamilyFilter, plugin.Log, nil)
	if err != nil {
		plugin.Log.Error(err)
		return err
	}
	for _, rt := range routes {
		plugin.Log.WithField("interface", rt.LinkIndex).Debugf("Found new static linux route")
		_, _, found := plugin.rtIndexes.LookupIdx(routeIdentifier(&rt))
		if !found {
			plugin.rtIndexes.RegisterName(routeIdentifier(&rt), plugin.RouteIdxSeq, nil)
			plugin.RouteIdxSeq++
			plugin.Log.Debugf("route registered as %v", routeIdentifier(&rt))
		}
	}

	return nil
}

// ResolveCreatedInterface manages cached static routes for new interface
func (plugin *LinuxRouteConfigurator) ResolveCreatedInterface(name string, index uint32) error {
	plugin.Log.Infof("Linux static route configurator: resolve new interface %v", name)

	// Search mapping for cached routes using the new interface
	cachedRoutes := plugin.rtCachedIndexes.LookupNamesByInterface(name)
	if len(cachedRoutes) > 0 {
		plugin.Log.Debugf("Found %v cached routes for interface %v", len(cachedRoutes), name)
		// store default routes, they have to be configured as the last ones
		var defRoutes []*l3.LinuxStaticRoutes_Route
		// static routes
		for _, cachedRoute := range cachedRoutes {
			if cachedRoute.Default {
				defRoutes = append(defRoutes, cachedRoute)
				continue
			}
			if err := plugin.ConfigureLinuxStaticRoute(cachedRoute); err != nil {
				plugin.Log.Warn(err)
				return err
			}
			// Remove from cache
			plugin.rtCachedIndexes.UnregisterName(cachedRoute.Name)
		}
		// default routes
		for _, cachedDefaultRoute := range defRoutes {
			if err := plugin.ConfigureLinuxStaticRoute(cachedDefaultRoute); err != nil {
				plugin.Log.Warn(err)
				return err
			}
			// Remove from cache
			plugin.rtCachedIndexes.UnregisterName(cachedDefaultRoute.Name)
		}
	}

	return nil
}

// ResolveDeletedInterface manages static routes for removed interface
func (plugin *LinuxRouteConfigurator) ResolveDeletedInterface(name string, index uint32) error {
	plugin.Log.Infof("Linux static route configurator: resolve deleted interface %v", name)

	// Search mapping for configured application namespaces using the new interface
	confRoutes := plugin.rtIndexes.LookupNamesByInterface(name)
	if len(confRoutes) > 0 {
		plugin.Log.Debugf("Found %v routes belonging to the removed interface %v", len(confRoutes), name)
		for _, rt := range confRoutes {
			// Add to un-configured. If the interface will be recreated, all routes are configured back
			plugin.rtCachedIndexes.RegisterName(rt.Name, plugin.RouteIdxSeq, rt)
			plugin.RouteIdxSeq++
		}
	}

	return nil
}

// Create default route object with gateway address. Destination address has to be set in such a case
func (plugin *LinuxRouteConfigurator) createDefaultRoute(netLinkRoute *netlink.Route, route *l3.LinuxStaticRoutes_Route) (cached bool, err error) {
	// Gateway
	if !plugin.networkReachable(route.Namespace, route.GwAddr) {
		plugin.rtCachedRoutes[route.GwAddr] = route
		plugin.Log.Debugf("Default route %v cached, gateway is currently unreachable", route.Name)
		return true, nil
	}
	// Check if route was not cached before, eventually remove it
	_, ok := plugin.rtCachedRoutes[route.GwAddr]
	if ok {
		delete(plugin.rtCachedRoutes, route.GwAddr)
	}
	gateway := net.ParseIP(route.GwAddr)
	if gateway == nil {
		return false, fmt.Errorf("unable to create route %v as default, gateway is nil", route.Name)
	}
	netLinkRoute.Gw = gateway

	// Destination address
	dstIPAddr := route.DstIpAddr
	if dstIPAddr == "" {
		dstIPAddr = ipv4AddrAny
	}
	if dstIPAddr != ipv4AddrAny && dstIPAddr != ipv6AddrAny {
		plugin.Log.Warnf("route marked as default has dst address set to %v. The address will be ignored", dstIPAddr)
		dstIPAddr = ipv4AddrAny
	}
	_, netLinkRoute.Dst, err = net.ParseCIDR(dstIPAddr)
	if err != nil {
		plugin.Log.Error(err)
		return false, err
	}

	// Priority
	if route.Metric != 0 {
		netLinkRoute.Priority = int(route.Metric)
	}

	plugin.Log.Debugf("created default route with gw ip %v", netLinkRoute.Gw)
	return false, nil
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
				plugin.Log.Error(err)
				return err
			}
		} else {
			return fmt.Errorf("cannot create static route %v, dst address net mask not set", route.Name)
		}
		plugin.Log.Infof("IP address %v set as dst for route %v", route.DstIpAddr, route.Name)
		netLinkRoute.Dst = dstIPAddr
	} else {
		return fmt.Errorf("cannot create static route %v, destination addres not set", route.Name)
	}

	// Set gateway if exists
	gateway := net.ParseIP(route.GwAddr)
	if gateway != nil {
		netLinkRoute.Gw = gateway
		plugin.Log.Infof("Gateway address %v set for route %v", route.GwAddr, route.Name)
	}

	netLinkRoute.Gw = gateway

	// Source IP address is exists
	srcIPAddr := net.ParseIP(route.SrcIpAddr)
	if srcIPAddr != nil {
		netLinkRoute.Src = srcIPAddr
		plugin.Log.Infof("IP address %v set as src for route %v", route.SrcIpAddr, route.Name)
	}

	// Scope
	netLinkRoute.Scope = plugin.parseRouteScope(route.Scope)

	// Priority
	if route.Metric != 0 {
		netLinkRoute.Priority = int(route.Metric)
	}

	// Table
	netLinkRoute.Table = int(route.Table)

	plugin.Log.Debugf("created static route with destination ip %v", netLinkRoute.Dst)
	return nil
}

// Update linux static route using modify (analogy to 'ip route replace')
func (plugin *LinuxRouteConfigurator) updateLinuxStaticRoute(netLinkRoute *netlink.Route, route *l3.LinuxStaticRoutes_Route) error {
	// Prepare namespace of related interface
	nsMgmtCtx := linuxcalls2.NewNamespaceMgmtCtx()
	routeNs := linuxcalls.ToGenericRouteNs(route.Namespace)

	// route has to be created in the same namespace as the interface
	revertNs, err := routeNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
	if err != nil {
		plugin.Log.Error(err)
		return err
	}
	defer revertNs()

	// Update existing route
	return linuxcalls.ModifyStaticRoute(route.Name, netLinkRoute, plugin.Log, measure.GetTimeLog("modify-linux-route", plugin.Stopwatch))
}

// Tries to reconfigure cached default routes
func (plugin *LinuxRouteConfigurator) retryDefaultRoutes(route *l3.LinuxStaticRoutes_Route) {
	plugin.Log.Debug("Retrying to configure default routes")
	for gwIP, defRoute := range plugin.rtCachedRoutes {
		// Parse gateway and default address
		gwIPParsed := net.ParseIP(gwIP)
		_, dstNet, err := net.ParseCIDR(route.DstIpAddr)
		if err != nil {
			plugin.Log.Errorf("Error parsing IP address %s: %v", route.DstIpAddr, err)
			continue
		}
		if dstNet.Contains(gwIPParsed) {
			// default route can be configured
			if err := plugin.ConfigureLinuxStaticRoute(defRoute); err != nil {
				plugin.Log.Errorf("Error while configuring route %s: %v", route.Name, err)
			}
			delete(plugin.rtCachedRoutes, gwIP)
		}
	}
}

func (plugin *LinuxRouteConfigurator) parseRouteScope(scope *l3.LinuxStaticRoutes_Route_Scope) netlink.Scope {
	if scope == nil {
		plugin.Log.Info("Scope type not defined, seting to default (link)")
		return netlink.SCOPE_LINK
	}
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
		plugin.Log.Infof("Unknown scope type, setting to default (link): %v", scope.Type)
		return netlink.SCOPE_LINK
	}
}

// Verifies whether address network is reachable.
func (plugin *LinuxRouteConfigurator) networkReachable(ns *l3.LinuxStaticRoutes_Route_Namespace, ipAddress string) bool {
	route, err := plugin.rtIndexes.LookupRouteByIP(ns, ipAddress)
	if err != nil {
		plugin.Log.Errorf("Failed to resolve accessibility of %s: %v", ipAddress, err)
		return false
	}
	if route == nil {
		plugin.Log.Debugf("IP address %s is not accessible", ipAddress)
		return false
	}
	plugin.Log.Debugf("IP address %s is accessible", ipAddress)
	return true
}

func routeIdentifier(route *netlink.Route) string {
	if route.Dst == nil {
		return fmt.Sprintf("default-iface%v-table%v-%v", route.LinkIndex, route.Table, route.Gw.String())
	}
	return fmt.Sprintf("dst%v-iface%v-table%v-%v", route.Dst.IP.String(), route.LinkIndex, route.Table, route.Gw.String())
}
