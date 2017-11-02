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

//go:generate protoc --proto_path=model --gogo_out=model model/l3/l3.proto

package l3plugin

import (
	"fmt"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/ifaceidx"
	linuxcalls2 "github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/model/l3"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
)

// LinuxRouteConfigurator watches for any changes in the configuration of static routes as modelled by the proto file
// "model/l3/l3.proto" and stored in ETCD under the key "/vnf-agent/{vnf-agent}/linux/config/v1/route".
// Updates received from the northbound API are compared with the Linux network configuration and differences
// are applied through the Netlink AP
type LinuxRouteConfigurator struct {
	Log logging.Logger

	LinuxIfIdx  ifaceidx.LinuxIfIndexRW
	RouteIdxSeq uint32
	rtIndexes   l3idx.LinuxRouteIndexRW

	// Time measurement
	Stopwatch *measure.Stopwatch // timer used to measure and store time

}

// Init initializes static route configurator and starts goroutines
func (plugin *LinuxRouteConfigurator) Init(rtIndexes l3idx.LinuxRouteIndexRW) error {
	plugin.Log.Debug("Initializing LinuxRouteConfigurator")
	plugin.rtIndexes = rtIndexes

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
	var err error

	// Prepare route object
	netLinkRoute := &netlink.Route{}

	// Find interface
	idx, _, foundIface := plugin.LinuxIfIdx.LookupIdx(route.Interface)
	if !foundIface {
		return fmt.Errorf("cannot create static route %v, interface %v not found", route.Name, route.Interface)
	}
	netLinkRoute.LinkIndex = int(idx)

	// default route
	if route.Default {
		err = plugin.createDefaultRoute(netLinkRoute, route)
		if err != nil {
			return err
		}
	} else {
		// static route
		err = plugin.createStaticRoute(netLinkRoute, route)
		if err != nil {
			return err
		}
	}

	// Prepare namespace of related interface
	nsMgmtCtx := linuxcalls2.NewNamespaceMgmtCtx()
	routeNs := linuxcalls.ToGenericRouteNs(route.Namespace)

	// route has to be created in the same namespace as the interface
	revertNs, err := routeNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
	if err != nil {
		return err
	}
	defer revertNs()

	err = linuxcalls.AddStaticRoute(route.Name, netLinkRoute, plugin.Log, measure.GetTimeLog("add-linux-route", plugin.Stopwatch))

	plugin.rtIndexes.RegisterName(routeIdentifier(netLinkRoute), plugin.RouteIdxSeq, nil)
	plugin.RouteIdxSeq++
	plugin.Log.Debugf("Route %v registered", route.Name)

	return err
}

// ModifyLinuxStaticRoute applies changes in the NB configuration of a Linux static route into the host network stack
// through Netlink API.
func (plugin *LinuxRouteConfigurator) ModifyLinuxStaticRoute(newRoute *l3.LinuxStaticRoutes_Route, oldRoute *l3.LinuxStaticRoutes_Route) error {
	plugin.Log.Infof("Modifying linux static route %v", newRoute.Name)
	var err error

	// Prepare route object
	netLinkRoute := &netlink.Route{}

	// Find interface
	idx, _, foundIface := plugin.LinuxIfIdx.LookupIdx(newRoute.Interface)
	if !foundIface {
		return fmt.Errorf("cannot update static route %v, interface %v not found", newRoute.Name, newRoute.Interface)
	}
	netLinkRoute.LinkIndex = int(idx)

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
		if err = plugin.createDefaultRoute(netLinkRoute, newRoute); err != nil {
			return nil
		}
	} else {
		if oldRoute.DstIpAddr != newRoute.Interface {
			replace = false
		}
		if err = plugin.createStaticRoute(netLinkRoute, newRoute); err != nil {
			return nil
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
		return err
	}
	defer revertNs()

	// Remove old route and create a new one
	if err = linuxcalls.AddStaticRoute(newRoute.Name, netLinkRoute, plugin.Log, measure.GetTimeLog("add-linux-route", plugin.Stopwatch)); err != nil {
		return err
	}
	return plugin.DeleteLinuxStaticRoute(oldRoute)
}

// DeleteLinuxStaticRoute reacts to a removed NB configuration of a Linux static route entry.
func (plugin *LinuxRouteConfigurator) DeleteLinuxStaticRoute(route *l3.LinuxStaticRoutes_Route) error {
	plugin.Log.Infof("Removing linux static route %v", route.Name)
	var err error

	// Prepare route object
	netLinkRoute := &netlink.Route{}

	// Find interface
	idx, _, foundIface := plugin.LinuxIfIdx.LookupIdx(route.Interface)
	if !foundIface {
		return fmt.Errorf("cannot delete static route %v, interface %v not found", route.Name, route.Interface)
	}
	netLinkRoute.LinkIndex = int(idx)

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
			return fmt.Errorf("cannot remove static route %v, dst address net mask not set", route.Name)
		}
		netLinkRoute.Dst = dstIPAddr
	} else {
		return fmt.Errorf("cannot remove static route %v, destination addres not set", route.Name)
	}

	// Scope
	netLinkRoute.Scope = plugin.parseRouteScope(route.Scope)

	// Prepare namespace of related interface
	nsMgmtCtx := linuxcalls2.NewNamespaceMgmtCtx()
	routeNs := linuxcalls.ToGenericRouteNs(route.Namespace)

	// route has to be created in the same namespace as the interface
	revertNs, err := routeNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
	if err != nil {
		return err
	}
	defer revertNs()

	err = linuxcalls.DeleteStaticRoute(route.Name, netLinkRoute, plugin.Log, measure.GetTimeLog("del-linux-route", plugin.Stopwatch))

	_, _, found := plugin.rtIndexes.UnregisterName(routeIdentifier(netLinkRoute))
	if !found {
		plugin.Log.Warnf("Attempt to unregister non-registered route %v", route.Name)
	}
	plugin.Log.Debugf("Route %v unregistered", route.Name)

	return err
}

// LookupLinuxRoutes reads all routes and registers them if needed
func (plugin *LinuxRouteConfigurator) LookupLinuxRoutes() error {
	plugin.Log.Infof("Browsing Linux routes")

	// read all routes
	routes, err := linuxcalls.ReadStaticRoutes(nil, noFamilyFilter, plugin.Log, nil)
	if err != nil {
		return err
	}
	for _, route := range routes {
		plugin.Log.WithField("interface", route.LinkIndex).Debugf("Found new static linux route")
		_, _, found := plugin.rtIndexes.LookupIdx(routeIdentifier(&route))
		if !found {
			plugin.rtIndexes.RegisterName(routeIdentifier(&route), plugin.RouteIdxSeq, nil)
			plugin.RouteIdxSeq++
			plugin.Log.Debug("route registered as %v", routeIdentifier(&route))
		}
	}

	return nil
}

// Create default route object with gateway address. Destination address has to be set in such a case
func (plugin *LinuxRouteConfigurator) createDefaultRoute(netLinkRoute *netlink.Route, route *l3.LinuxStaticRoutes_Route) error {
	// Destination address
	if route.DstIpAddr != "" {
		plugin.Log.Warnf("route marked as default has dst address set to %v. The address will be ignored", route.DstIpAddr)
	}
	netLinkRoute.Dst = nil
	// Gateway
	gateway := net.ParseIP(route.GwAddr)
	if gateway == nil {
		return fmt.Errorf("unable to create route %v as default, gateway is nil", route.Name)
	}
	netLinkRoute.Gw = gateway

	// Priority
	if route.Metric != 0 {
		netLinkRoute.Priority = int(route.Metric)
	}

	plugin.Log.Debugf("created default route with gw ip %v", netLinkRoute.Gw)
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
		return err
	}
	defer revertNs()

	// Update existing route
	return linuxcalls.ModifyStaticRoute(route.Name, netLinkRoute, plugin.Log, measure.GetTimeLog("modify-linux-route", plugin.Stopwatch))
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

func routeIdentifier(route *netlink.Route) string {
	if route.Dst == nil {
		return fmt.Sprintf("default-iface%v-table%v-%v", route.LinkIndex, route.Table, route.Gw.String())
	}
	return fmt.Sprintf("dst%v-iface%v-table%v-%v", route.Dst.IP.String(), route.LinkIndex, route.Table, route.Gw.String())
}
