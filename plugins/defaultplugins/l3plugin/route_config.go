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
//go:generate binapi-generator --input-file=/usr/share/vpp/api/ip.api.json --output-dir=../common/bin_api

// Package l3plugin implements the L3 plugin that handles L3 FIBs.
package l3plugin

import (
	"fmt"
	"strconv"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// RouteConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of L3 routes as modelled by the proto file "../model/l3/l3.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1routes". Updates received from the northbound API
// are compared with the VPP run-time configuration and differences are applied through the VPP binary API.
type RouteConfigurator struct {
	Log              logging.Logger
	GoVppmux         govppmux.API
	RouteIndexes     l3idx.RouteIndexRW
	RouteIndexSeq    uint32
	SwIfIndexes      ifaceidx.SwIfIndex
	RouteCachedIndex l3idx.RouteIndexRW
	vppChan          *govppapi.Channel
	Stopwatch        *measure.Stopwatch // timer used to measure and store time

}

// Init members (channels...) and start go routines.
func (plugin *RouteConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing L3 Route configurator")

	// Init VPP API channel.
	plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	err = plugin.checkMsgCompatibility()
	if err != nil {
		return err
	}

	return nil
}

// ConfigureRoute processes the NB config and propagates it to bin api calls.
func (plugin *RouteConfigurator) ConfigureRoute(config *l3.StaticRoutes_Route, vrfFromKey string) error {
	plugin.Log.Infof("Configuring new route %v -> %v", config.DstIpAddr, config.NextHopAddr)
	// Validate VRF index from key and it's value in data.
	if err := plugin.validateVrfFromKey(config, vrfFromKey); err != nil {
		return err
	}

	routeID := routeIdentifier(config.VrfId, config.DstIpAddr, config.NextHopAddr)
	_, _, routeExists := plugin.RouteIndexes.LookupIdx(routeID)
	if !routeExists {
		plugin.RouteIndexes.RegisterName(routeID, plugin.RouteIndexSeq, config)
		plugin.RouteIndexSeq++
		plugin.Log.Infof("Route %v registered", routeID)
	}

	swIdx, err := resolveInterfaceSwIndex(config.OutgoingInterface, plugin.SwIfIndexes)
	if err != nil {
		plugin.RouteCachedIndex.RegisterName(routeID, plugin.RouteIndexSeq, config)
		plugin.RouteIndexSeq++
		plugin.Log.Debugf("Route %v registered to cache", routeID)
		return nil
	}

	// Transform route data.
	route, err := TransformRoute(config, swIdx, plugin.Log)
	if err != nil {
		return err
	}

	// Create and register new route.
	if route != nil {
		err := vppcalls.VppAddRoute(route, plugin.vppChan, measure.GetTimeLog(ip.IPAddDelRoute{}, plugin.Stopwatch))
		if err != nil {
			return err
		}
	}

	plugin.Log.Infof("Route %v -> %v configured", config.DstIpAddr, config.NextHopAddr)

	return nil
}

// ModifyRoute processes the NB config and propagates it to bin api calls.
func (plugin *RouteConfigurator) ModifyRoute(newConfig *l3.StaticRoutes_Route, oldConfig *l3.StaticRoutes_Route, vrfFromKey string) error {
	plugin.Log.Infof("Modifying route %v -> %v", oldConfig.DstIpAddr, oldConfig.NextHopAddr)

	outgoingIfName := newConfig.OutgoingInterface
	if outgoingIfName != "" {
		_, _, existsNewOutgoing := plugin.SwIfIndexes.LookupIdx(outgoingIfName)
		routeID := routeIdentifier(oldConfig.VrfId, oldConfig.DstIpAddr, oldConfig.NextHopAddr)
		if existsNewOutgoing {
			plugin.Log.Debugf("Route %s unregistered from cache.", routeID)
			plugin.RouteCachedIndex.UnregisterName(routeID)
		} else {
			routeIdx, _, isRouteCached := plugin.RouteCachedIndex.LookupIdx(routeID)
			if isRouteCached {
				plugin.RouteCachedIndex.RegisterName(routeID, routeIdx, newConfig)
			} else {
				plugin.RouteCachedIndex.RegisterName(routeID, plugin.RouteIndexSeq, newConfig)
				plugin.RouteIndexSeq++
			}
		}
	}

	var err error
	err = plugin.deleteOldRoute(oldConfig, vrfFromKey)
	if err != nil {
		return err
	}

	err = plugin.addNewRoute(newConfig, vrfFromKey)
	if err != nil {
		return err
	}

	plugin.Log.Infof("Route %v -> %v modified", oldConfig.DstIpAddr, oldConfig.NextHopAddr)

	return nil
}

func (plugin *RouteConfigurator) deleteOldRoute(oldConfig *l3.StaticRoutes_Route, vrfFromKey string) error {
	swIdx, error := resolveInterfaceSwIndex(oldConfig.OutgoingInterface, plugin.SwIfIndexes)
	if error != nil {
		return error
	}

	// Transform old route data.
	oldRoute, err := TransformRoute(oldConfig, swIdx, plugin.Log)
	if err != nil {
		return err
	}

	// Validate old cachedRoute data Vrf.
	if err := plugin.validateVrfFromKey(oldConfig, vrfFromKey); err != nil {
		return err
	}
	// Remove and unregister old route.
	err = vppcalls.VppDelRoute(oldRoute, plugin.vppChan, measure.GetTimeLog(ip.IPAddDelRoute{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	oldRouteIdentifier := routeIdentifier(oldRoute.VrfID, oldRoute.DstAddr.String(), oldRoute.NextHopAddr.String())
	_, _, found := plugin.RouteIndexes.UnregisterName(oldRouteIdentifier)
	if found {
		plugin.Log.Infof("Old route %v unregistered", oldRouteIdentifier)
	} else {
		plugin.Log.Warnf("Unregister failed, old route %v not found", oldRouteIdentifier)
	}
	return nil
}

func (plugin *RouteConfigurator) addNewRoute(newConfig *l3.StaticRoutes_Route, vrfFromKey string) error {
	// Validate new route data Vrf.
	if err := plugin.validateVrfFromKey(newConfig, vrfFromKey); err != nil {
		return err
	}

	swIdx, error := resolveInterfaceSwIndex(newConfig.OutgoingInterface, plugin.SwIfIndexes)
	if error != nil {
		return error
	}

	// Transform new route data.
	newRoute, err := TransformRoute(newConfig, swIdx, plugin.Log)
	if err != nil {
		return err
	}
	// Create and register new route.
	err = vppcalls.VppAddRoute(newRoute, plugin.vppChan, measure.GetTimeLog(ip.IPAddDelRoute{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	newRouteIdentifier := routeIdentifier(newConfig.VrfId, newConfig.DstIpAddr, newConfig.NextHopAddr)
	plugin.RouteIndexes.RegisterName(newRouteIdentifier, plugin.RouteIndexSeq, newConfig)
	plugin.RouteIndexSeq++
	plugin.Log.Infof("New route %v registered", newRouteIdentifier)

	return nil
}

// DeleteRoute processes the NB config and propagates it to bin api calls.
func (plugin *RouteConfigurator) DeleteRoute(config *l3.StaticRoutes_Route, vrfFromKey string) (wasError error) {
	plugin.Log.Infof("Removing route %v -> %v", config.DstIpAddr, config.NextHopAddr)
	// Validate VRF index from key and it's value in data.
	if err := plugin.validateVrfFromKey(config, vrfFromKey); err != nil {
		return err
	}

	swIdx, error := resolveInterfaceSwIndex(config.OutgoingInterface, plugin.SwIfIndexes)
	if error != nil {
		return error
	}

	// Transform route data.
	route, err := TransformRoute(config, swIdx, plugin.Log)
	if err != nil {
		return err
	}
	if route == nil {
		return nil
	}
	plugin.Log.Debugf("deleting route: %+v", route)
	// Remove and unregister route.
	err = vppcalls.VppDelRoute(route, plugin.vppChan, measure.GetTimeLog(ip.IPAddDelRoute{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	routeIdentifier := routeIdentifier(config.VrfId, config.DstIpAddr, config.NextHopAddr)
	_, _, found := plugin.RouteIndexes.UnregisterName(routeIdentifier)
	if found {
		plugin.Log.Infof("Route %v unregistered", routeIdentifier)
	} else {
		plugin.Log.Warnf("Unregister failed, route %v not found", routeIdentifier)
	}

	plugin.Log.Infof("Route %v -> %v removed", config.DstIpAddr, config.NextHopAddr)

	return nil
}

func (plugin *RouteConfigurator) validateVrfFromKey(config *l3.StaticRoutes_Route, vrfFromKey string) error {
	intVrfFromKey, err := strconv.Atoi(vrfFromKey)
	if intVrfFromKey != int(config.VrfId) {
		if err != nil {
			return err
		}
		plugin.Log.Warnf("VRF index from key (%v) and from config (%v) does not match, using value from the key",
			intVrfFromKey, config.VrfId)
		config.VrfId = uint32(intVrfFromKey)
	}
	return nil
}

func (plugin *RouteConfigurator) checkMsgCompatibility() error {
	msgs := []govppapi.Message{
		&ip.IPAddDelRoute{},
		&ip.IPAddDelRouteReply{},
		&ip.IPFibDump{},
		&ip.IPFibDetails{},
		&ip.IP6FibDump{},
		&ip.IP6FibDetails{},
	}
	err := plugin.vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		plugin.Log.Error(err)
	}
	return err
}

// Close GOVPP channel.
func (plugin *RouteConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// Create unique identifier which serves as a name in name-to-index mapping.
func routeIdentifier(vrf uint32, destination string, nextHop string) string {
	return fmt.Sprintf("vrf%v-%v-%v", vrf, destination, nextHop)
}

//ResolveCreatedInterface is responsible for reconfiguring cached routes and then from removing
//them from route cache
func (plugin *RouteConfigurator) ResolveCreatedInterface(ifName string, swIdx uint32) {
	routesWithIndex := plugin.RouteCachedIndex.LookupRouteAndIDByOutgoingIfc(ifName)
	plugin.Log.Infof("Route configurator: resolving new interface %v", ifName)
	for _, routeWithIndex := range routesWithIndex {
		route := routeWithIndex.Route
		plugin.Log.WithFields(
			logging.Fields{
				"interface ifName":         ifName,
				"interface software index": swIdx,
				"vrf":            route.VrfId,
				"destination ip": route.DstIpAddr}).
			Debug("Remove routes from route cache - outgoing interface was added.")
		vrf := strconv.FormatUint(uint64(route.VrfId), 10)
		plugin.recreateRoute(route, vrf)
		plugin.RouteCachedIndex.UnregisterName(routeWithIndex.RouteID)
	}
}

/**
recreateRoute calls delete and configure route.

This is type of workaround because when outgoing interface is deleted then it isn't possible to remove
associated routes. they stay in following state:
- oper-flags:drop
- routing section: unresolved
It is neither possible to recreate interface and then create route.
It is only possible to recreate interface, delete old associated routes (like clean old mess)
and then add them again.
*/
func (plugin *RouteConfigurator) recreateRoute(route *l3.StaticRoutes_Route, vrf string) {
	plugin.DeleteRoute(route, vrf)
	plugin.ConfigureRoute(route, vrf)
}

//ResolveDeletedInterface is responsible for moving routes of deleted interface to cache
func (plugin *RouteConfigurator) ResolveDeletedInterface(ifName string, swIdx uint32) {
	routesWithIndex := plugin.RouteIndexes.LookupRouteAndIDByOutgoingIfc(ifName)
	plugin.Log.Debugf("Route configurator: resolving deleted interface %v", ifName)
	for _, routeWithIndex := range routesWithIndex {
		route := routeWithIndex.Route
		plugin.Log.WithFields(
			logging.Fields{
				"interface ifName":         ifName,
				"interface software index": swIdx,
				"vrf":            route.VrfId,
				"destination ip": route.DstIpAddr}).
			Debug("Add routes to route cache - outgoing interface was deleted.")
		plugin.moveRouteToCache(route)
	}
}

func (plugin *RouteConfigurator) moveRouteToCache(config *l3.StaticRoutes_Route) (wasError error) {
	routeID := routeIdentifier(config.VrfId, config.DstIpAddr, config.NextHopAddr)
	_, _, found := plugin.RouteIndexes.UnregisterName(routeID)
	if found {
		plugin.Log.Infof("Route %v unregistered", routeID)
	} else {
		plugin.Log.Warnf("Unregister failed, route %v not found", routeID)
	}

	plugin.Log.Infof("Route %s registrated in cache", routeID)
	plugin.RouteCachedIndex.RegisterName(routeID, plugin.RouteIndexSeq, config)
	plugin.RouteIndexSeq++

	return nil
}
