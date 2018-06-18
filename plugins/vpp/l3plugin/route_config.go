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

// Package l3plugin implements the L3 plugin that handles L3 FIBs.
package l3plugin

import (
	"fmt"
	"strconv"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
)

// RouteConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of L3 routes as modelled by the proto file "../model/l3/l3.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1routes". Updates received from the northbound API
// are compared with the VPP run-time configuration and differences are applied through the VPP binary API.
type RouteConfigurator struct {
	log logging.Logger

	// In-memory mappings
	ifIndexes       ifaceidx.SwIfIndex
	rtIndexes       l3idx.RouteIndexRW
	rtCachedIndexes l3idx.RouteIndexRW
	rtIndexSeq      uint32

	// VPP channels
	vppChan *govppapi.Channel

	// Timer used to measure and store time
	stopwatch *measure.Stopwatch
}

// Init members (channels...) and start go routines.
func (plugin *RouteConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, swIfIndexes ifaceidx.SwIfIndex,
	enableStopwatch bool) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-l3-route-conf")
	plugin.log.Debug("Initializing L3 Route configurator")

	// Mappings
	plugin.ifIndexes = swIfIndexes
	plugin.rtIndexes = l3idx.NewRouteIndex(nametoidx.NewNameToIdx(plugin.log, "route_indexes", nil))
	plugin.rtCachedIndexes = l3idx.NewRouteIndex(nametoidx.NewNameToIdx(plugin.log, "route_cached_indexes", nil))
	plugin.rtIndexSeq = 1

	// VPP channel
	plugin.vppChan, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}

	// Stopwatch
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("RouteConfigurator", plugin.log)
	}

	// Message compatibility
	if err := plugin.vppChan.CheckMessageCompatibility(vppcalls.RouteMessages...); err != nil {
		plugin.log.Error(err)
		return err
	}

	return nil
}

// Close GOVPP channel.
func (plugin *RouteConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// Create unique identifier which serves as a name in name-to-index mapping.
func routeIdentifier(vrf uint32, destination string, nextHop string) string {
	return fmt.Sprintf("vrf%v-%v-%v", vrf, destination, nextHop)
}

// ConfigureRoute processes the NB config and propagates it to bin api calls.
func (plugin *RouteConfigurator) ConfigureRoute(config *l3.StaticRoutes_Route, vrfFromKey string) error {
	plugin.log.Infof("Configuring new route %v -> %v", config.DstIpAddr, config.NextHopAddr)

	// Validate VRF index from key and it's value in data.
	if err := plugin.validateVrfFromKey(config, vrfFromKey); err != nil {
		return err
	}

	routeID := routeIdentifier(config.VrfId, config.DstIpAddr, config.NextHopAddr)

	swIdx, err := resolveInterfaceSwIndex(config.OutgoingInterface, plugin.ifIndexes)
	if err != nil {
		plugin.rtCachedIndexes.RegisterName(routeID, plugin.rtIndexSeq, config)
		plugin.rtIndexSeq++
		plugin.log.Debugf("Route %v registered to cache", routeID)
		return nil
	}

	// Transform route data.
	route, err := TransformRoute(config, swIdx, plugin.log)
	if err != nil {
		return err
	}

	// Create and register new route.
	if route != nil {
		err := vppcalls.VppAddRoute(route, plugin.vppChan, plugin.stopwatch)
		if err != nil {
			return err
		}
	}

	// Register configured route
	_, _, routeExists := plugin.rtIndexes.LookupIdx(routeID)
	if !routeExists {
		plugin.rtIndexes.RegisterName(routeID, plugin.rtIndexSeq, config)
		plugin.rtIndexSeq++
		plugin.log.Infof("Route %v registered", routeID)
	}

	plugin.log.Infof("Route %v -> %v configured", config.DstIpAddr, config.NextHopAddr)
	return nil
}

// ModifyRoute processes the NB config and propagates it to bin api calls.
func (plugin *RouteConfigurator) ModifyRoute(newConfig *l3.StaticRoutes_Route, oldConfig *l3.StaticRoutes_Route, vrfFromKey string) error {
	plugin.log.Infof("Modifying route %v -> %v", oldConfig.DstIpAddr, oldConfig.NextHopAddr)

	routeID := routeIdentifier(oldConfig.VrfId, oldConfig.DstIpAddr, oldConfig.NextHopAddr)

	if newConfig.OutgoingInterface != "" {
		_, _, existsNewOutgoing := plugin.ifIndexes.LookupIdx(newConfig.OutgoingInterface)
		if existsNewOutgoing {
			plugin.log.Debugf("Route %s unregistered from cache", routeID)
			plugin.rtCachedIndexes.UnregisterName(routeID)
		} else {
			if routeIdx, _, isCached := plugin.rtCachedIndexes.LookupIdx(routeID); isCached {
				plugin.rtCachedIndexes.RegisterName(routeID, routeIdx, newConfig)
			} else {
				plugin.rtCachedIndexes.RegisterName(routeID, plugin.rtIndexSeq, newConfig)
				plugin.rtIndexSeq++
			}
		}
	}

	if err := plugin.deleteOldRoute(oldConfig, vrfFromKey); err != nil {
		return err
	}

	if err := plugin.addNewRoute(newConfig, vrfFromKey); err != nil {
		return err
	}

	plugin.log.Infof("Route %v -> %v modified", oldConfig.DstIpAddr, oldConfig.NextHopAddr)
	return nil
}

func (plugin *RouteConfigurator) deleteOldRoute(oldConfig *l3.StaticRoutes_Route, vrfFromKey string) error {
	swIdx, err := resolveInterfaceSwIndex(oldConfig.OutgoingInterface, plugin.ifIndexes)
	if err != nil {
		return err
	}

	// Transform old route data.
	oldRoute, err := TransformRoute(oldConfig, swIdx, plugin.log)
	if err != nil {
		return err
	}

	// Validate old cachedRoute data Vrf.
	if err := plugin.validateVrfFromKey(oldConfig, vrfFromKey); err != nil {
		return err
	}
	// Remove and unregister old route.
	if err := vppcalls.VppDelRoute(oldRoute, plugin.vppChan, plugin.stopwatch); err != nil {
		return err
	}

	oldRouteIdentifier := routeIdentifier(oldRoute.VrfID, oldRoute.DstAddr.String(), oldRoute.NextHopAddr.String())

	_, _, found := plugin.rtIndexes.UnregisterName(oldRouteIdentifier)
	if found {
		plugin.log.Infof("Old route %v unregistered", oldRouteIdentifier)
	} else {
		plugin.log.Warnf("Unregister failed, old route %v not found", oldRouteIdentifier)
	}

	return nil
}

func (plugin *RouteConfigurator) addNewRoute(newConfig *l3.StaticRoutes_Route, vrfFromKey string) error {
	// Validate new route data Vrf.
	if err := plugin.validateVrfFromKey(newConfig, vrfFromKey); err != nil {
		return err
	}

	swIdx, err := resolveInterfaceSwIndex(newConfig.OutgoingInterface, plugin.ifIndexes)
	if err != nil {
		return err
	}

	// Transform new route data.
	newRoute, err := TransformRoute(newConfig, swIdx, plugin.log)
	if err != nil {
		return err
	}
	// Create and register new route.
	if err = vppcalls.VppAddRoute(newRoute, plugin.vppChan, plugin.stopwatch); err != nil {
		return err
	}

	newRouteIdentifier := routeIdentifier(newConfig.VrfId, newConfig.DstIpAddr, newConfig.NextHopAddr)
	plugin.rtIndexes.RegisterName(newRouteIdentifier, plugin.rtIndexSeq, newConfig)
	plugin.rtIndexSeq++

	plugin.log.Infof("New route %v registered", newRouteIdentifier)
	return nil
}

// DeleteRoute processes the NB config and propagates it to bin api calls.
func (plugin *RouteConfigurator) DeleteRoute(config *l3.StaticRoutes_Route, vrfFromKey string) (wasError error) {
	plugin.log.Infof("Removing route %v -> %v", config.DstIpAddr, config.NextHopAddr)
	// Validate VRF index from key and it's value in data.
	if err := plugin.validateVrfFromKey(config, vrfFromKey); err != nil {
		return err
	}

	swIdx, err := resolveInterfaceSwIndex(config.OutgoingInterface, plugin.ifIndexes)
	if err != nil {
		return err
	}

	// Transform route data.
	route, err := TransformRoute(config, swIdx, plugin.log)
	if err != nil {
		return err
	}
	if route == nil {
		return nil
	}
	plugin.log.Debugf("deleting route: %+v", route)

	// Remove and unregister route.
	if err = vppcalls.VppDelRoute(route, plugin.vppChan, plugin.stopwatch); err != nil {
		return err
	}

	routeIdentifier := routeIdentifier(config.VrfId, config.DstIpAddr, config.NextHopAddr)
	_, _, found := plugin.rtIndexes.UnregisterName(routeIdentifier)
	if found {
		plugin.log.Infof("Route %v unregistered", routeIdentifier)
	} else {
		plugin.log.Warnf("Unregister failed, route %v not found", routeIdentifier)
	}

	plugin.log.Infof("Route %v -> %v removed", config.DstIpAddr, config.NextHopAddr)
	return nil
}

func (plugin *RouteConfigurator) validateVrfFromKey(config *l3.StaticRoutes_Route, vrfFromKey string) error {
	intVrfFromKey, err := strconv.Atoi(vrfFromKey)
	if intVrfFromKey != int(config.VrfId) {
		if err != nil {
			return err
		}
		plugin.log.Warnf("VRF index from key (%v) and from config (%v) does not match, using value from the key",
			intVrfFromKey, config.VrfId)
		config.VrfId = uint32(intVrfFromKey)
	}
	return nil
}

// ResolveCreatedInterface is responsible for reconfiguring cached routes and then from removing them from route cache
func (plugin *RouteConfigurator) ResolveCreatedInterface(ifName string, swIdx uint32) {
	routesWithIndex := plugin.rtCachedIndexes.LookupRouteAndIDByOutgoingIfc(ifName)
	if len(routesWithIndex) == 0 {
		return
	}
	plugin.log.Infof("Route configurator: resolving new interface %v for %d routes", ifName, len(routesWithIndex))
	for _, routeWithIndex := range routesWithIndex {
		route := routeWithIndex.Route
		plugin.log.WithFields(logging.Fields{
			"ifName":    ifName,
			"swIdx":     swIdx,
			"vrfID":     route.VrfId,
			"dstIPAddr": route.DstIpAddr,
		}).Debug("Remove routes from route cache - outgoing interface was added.")
		vrf := strconv.FormatUint(uint64(route.VrfId), 10)
		plugin.recreateRoute(route, vrf)
		plugin.rtCachedIndexes.UnregisterName(routeWithIndex.RouteID)
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

// ResolveDeletedInterface is responsible for moving routes of deleted interface to cache
func (plugin *RouteConfigurator) ResolveDeletedInterface(ifName string, swIdx uint32) {
	routesWithIndex := plugin.rtIndexes.LookupRouteAndIDByOutgoingIfc(ifName)
	if len(routesWithIndex) == 0 {
		return
	}
	plugin.log.Debugf("Route configurator: resolving deleted interface %v for %d routes", ifName, len(routesWithIndex))
	for _, routeWithIndex := range routesWithIndex {
		route := routeWithIndex.Route
		plugin.log.WithFields(logging.Fields{
			"ifName":    ifName,
			"swIdx":     swIdx,
			"vrfID":     route.VrfId,
			"dstIPAddr": route.DstIpAddr,
		}).Debug("Add routes to route cache - outgoing interface was deleted.")
		plugin.moveRouteToCache(route)
	}
}

func (plugin *RouteConfigurator) moveRouteToCache(config *l3.StaticRoutes_Route) (wasError error) {
	routeID := routeIdentifier(config.VrfId, config.DstIpAddr, config.NextHopAddr)
	_, _, found := plugin.rtIndexes.UnregisterName(routeID)
	if found {
		plugin.log.Infof("Route %v unregistered", routeID)
	} else {
		plugin.log.Warnf("Unregister failed, route %v not found", routeID)
	}

	plugin.log.Infof("Route %s registrated in cache", routeID)
	plugin.rtCachedIndexes.RegisterName(routeID, plugin.rtIndexSeq, config)
	plugin.rtIndexSeq++

	return nil
}
