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

	"strings"

	"sort"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
)

// SortedRoutes type is used to implement sort interface for slice of Route.
type SortedRoutes []*l3.StaticRoutes_Route

// Return length of slice.
// Implements sort.Interface
func (arr SortedRoutes) Len() int {
	return len(arr)
}

// Swap swaps two items in slice identified by indices.
// Implements sort.Interface
func (arr SortedRoutes) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

// Less returns true if the item at index i in slice
// should be sorted before the element with index j.
// Implements sort.Interface
func (arr SortedRoutes) Less(i, j int) bool {
	return lessRoute(arr[i], arr[j])
}

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
	vppChan govppapi.Channel
	// VPP API handlers
	ifHandler ifvppcalls.IfVppWrite
	rtHandler vppcalls.RouteVppAPI

	// Timer used to measure and store time
	stopwatch *measure.Stopwatch
}

// Init members (channels...) and start go routines.
func (plugin *RouteConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, swIfIndexes ifaceidx.SwIfIndex,
	enableStopwatch bool) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-l3-route-conf")
	plugin.log.Debug("Initializing L3 Route configurator")

	// Configurator-wide stopwatch instance
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("Route-configurator", plugin.log)
	}

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

	// VPP API handlers
	plugin.ifHandler = ifvppcalls.NewIfVppHandler(plugin.vppChan, plugin.log, plugin.stopwatch)
	plugin.rtHandler = vppcalls.NewRouteVppHandler(plugin.vppChan, plugin.ifIndexes, plugin.log, plugin.stopwatch)

	return nil
}

// GetRouteIndexes exposes rtIndexes mapping
func (plugin *RouteConfigurator) GetRouteIndexes() l3idx.RouteIndex {
	return plugin.rtIndexes
}

// GetCachedRouteIndexes exposes rtCachedIndexes mapping
func (plugin *RouteConfigurator) GetCachedRouteIndexes() l3idx.RouteIndex {
	return plugin.rtCachedIndexes
}

// Close GOVPP channel.
func (plugin *RouteConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// clearMapping prepares all in-memory-mappings and other cache fields. All previous cached entries are removed.
func (plugin *RouteConfigurator) clearMapping() {
	plugin.rtIndexes.Clear()
	plugin.rtCachedIndexes.Clear()
}

// Create unique identifier which serves as a name in name-to-index mapping.
func routeIdentifier(vrf uint32, destination string, nextHop string) string {
	if nextHop == "<nil>" {
		nextHop = ""
	}
	return fmt.Sprintf("vrf%v-%v-%v", vrf, destination, nextHop)
}

// ConfigureRoute processes the NB config and propagates it to bin api calls.
func (plugin *RouteConfigurator) ConfigureRoute(route *l3.StaticRoutes_Route, vrfFromKey string) error {
	plugin.log.Infof("Configuring new route %v -> %v", route.DstIpAddr, route.NextHopAddr)
	// Validate VRF index from key and it's value in data.
	if err := plugin.validateVrfFromKey(route, vrfFromKey); err != nil {
		return err
	}

	routeID := routeIdentifier(route.VrfId, route.DstIpAddr, route.NextHopAddr)

	swIdx, err := resolveInterfaceSwIndex(route.OutgoingInterface, plugin.ifIndexes)
	if err != nil {
		plugin.rtCachedIndexes.RegisterName(routeID, plugin.rtIndexSeq, route)
		plugin.rtIndexSeq++
		plugin.log.Debugf("Route %v registered to cache", routeID)
		return nil
	}

	// Check mandatory destination address
	if route.DstIpAddr == "" {
		return fmt.Errorf("route %v does not contain destination address", routeID)
		return nil
	}

	// Create new route.
	err = plugin.rtHandler.VppAddRoute(plugin.ifHandler, route, swIdx)
	if err != nil {
		return err
	}

	// Register configured route
	_, _, routeExists := plugin.rtIndexes.LookupIdx(routeID)
	if !routeExists {
		plugin.rtIndexes.RegisterName(routeID, plugin.rtIndexSeq, route)
		plugin.rtIndexSeq++
		plugin.log.Infof("Route %v registered", routeID)
	}

	plugin.log.Infof("Route %v -> %v configured", route.DstIpAddr, route.NextHopAddr)
	return nil
}

// ModifyRoute processes the NB config and propagates it to bin api calls.
func (plugin *RouteConfigurator) ModifyRoute(newConfig *l3.StaticRoutes_Route, oldConfig *l3.StaticRoutes_Route, vrfFromKey string) error {
	plugin.log.Infof("Modifying route %v -> %v", oldConfig.DstIpAddr, oldConfig.NextHopAddr)

	routeID := routeIdentifier(oldConfig.VrfId, oldConfig.DstIpAddr, oldConfig.NextHopAddr)
	if newConfig.OutgoingInterface != "" {
		_, _, existsNewOutgoing := plugin.ifIndexes.LookupIdx(newConfig.OutgoingInterface)
		newrouteID := routeIdentifier(newConfig.VrfId, newConfig.DstIpAddr, newConfig.NextHopAddr)
		if existsNewOutgoing {
			plugin.log.Debugf("Route %s unregistered from cache", newrouteID)
			plugin.rtCachedIndexes.UnregisterName(newrouteID)
		} else {
			if routeIdx, _, isCached := plugin.rtCachedIndexes.LookupIdx(routeID); isCached {
				plugin.rtCachedIndexes.RegisterName(newrouteID, routeIdx, newConfig)
			} else {
				plugin.rtCachedIndexes.RegisterName(newrouteID, plugin.rtIndexSeq, newConfig)
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

func (plugin *RouteConfigurator) deleteOldRoute(route *l3.StaticRoutes_Route, vrfFromKey string) error {
	// Check if route entry is not just cached
	routeID := routeIdentifier(route.VrfId, route.DstIpAddr, route.NextHopAddr)
	_, _, found := plugin.rtCachedIndexes.LookupIdx(routeID)
	if found {
		plugin.log.Debugf("Route entry %v found in cache, removed", routeID)
		plugin.rtCachedIndexes.UnregisterName(routeID)
		// Cached route is not configured on the VPP, return
		return nil
	}

	swIdx, err := resolveInterfaceSwIndex(route.OutgoingInterface, plugin.ifIndexes)
	if err != nil {
		return err
	}

	// Validate old cachedRoute data Vrf.
	if err := plugin.validateVrfFromKey(route, vrfFromKey); err != nil {
		return err
	}
	// Remove and unregister old route.
	if err := plugin.rtHandler.VppDelRoute(route, swIdx); err != nil {
		return err
	}
	_, _, found = plugin.rtIndexes.UnregisterName(routeID)
	if found {
		plugin.log.Infof("Old route %v unregistered", routeID)
	} else {
		plugin.log.Warnf("Unregister failed, old route %v not found", routeID)
	}

	return nil
}

func (plugin *RouteConfigurator) addNewRoute(route *l3.StaticRoutes_Route, vrfFromKey string) error {
	// Validate new route data Vrf.
	if err := plugin.validateVrfFromKey(route, vrfFromKey); err != nil {
		return err
	}

	swIdx, err := resolveInterfaceSwIndex(route.OutgoingInterface, plugin.ifIndexes)
	if err != nil {
		return err
	}

	// Create and register new route.
	if err = plugin.rtHandler.VppAddRoute(plugin.ifHandler, route, swIdx); err != nil {
		return err
	}

	routeID := routeIdentifier(route.VrfId, route.DstIpAddr, route.NextHopAddr)
	plugin.rtIndexes.RegisterName(routeID, plugin.rtIndexSeq, route)
	plugin.rtIndexSeq++

	plugin.log.Infof("New route %v registered", routeID)
	return nil
}

// DeleteRoute processes the NB config and propagates it to bin api calls.
func (plugin *RouteConfigurator) DeleteRoute(route *l3.StaticRoutes_Route, vrfFromKey string) (wasError error) {
	plugin.log.Infof("Removing route %v -> %v", route.DstIpAddr, route.NextHopAddr)

	// Validate VRF index from key and it's value in data.
	if err := plugin.validateVrfFromKey(route, vrfFromKey); err != nil {
		return err
	}

	// Check if route entry is not just cached
	routeID := routeIdentifier(route.VrfId, route.DstIpAddr, route.NextHopAddr)
	_, _, found := plugin.rtCachedIndexes.LookupIdx(routeID)
	if found {
		plugin.log.Debugf("Route entry %v found in cache, removed", routeID)
		plugin.rtCachedIndexes.UnregisterName(routeID)
		// Cached route is not configured on the VPP, return
		return nil
	}

	swIdx, err := resolveInterfaceSwIndex(route.OutgoingInterface, plugin.ifIndexes)
	if err != nil {
		return err
	}

	// Remove and unregister route.
	if err := plugin.rtHandler.VppDelRoute(route, swIdx); err != nil {
		return err
	}

	routeIdentifier := routeIdentifier(route.VrfId, route.DstIpAddr, route.NextHopAddr)
	_, _, found = plugin.rtIndexes.UnregisterName(routeIdentifier)
	if found {
		plugin.log.Infof("Route %v unregistered", routeIdentifier)
	} else {
		plugin.log.Warnf("Unregister failed, route %v not found", routeIdentifier)
	}

	plugin.log.Infof("Route %v -> %v removed", route.DstIpAddr, route.NextHopAddr)
	return nil
}

// DiffRoutes calculates route diff from two sets of routes and returns routes to be added and removed
func (plugin *RouteConfigurator) DiffRoutes(new, old []*l3.StaticRoutes_Route) (toBeDeleted, toBeAdded []*l3.StaticRoutes_Route) {
	oldSorted, newSorted := SortedRoutes(old), SortedRoutes(new)
	sort.Sort(newSorted)
	sort.Sort(oldSorted)

	// Compare.
	i, j := 0, 0
	for i < len(newSorted) && j < len(oldSorted) {
		if *newSorted[i] == *oldSorted[j] {
			i++
			j++
		} else {
			if lessRoute(newSorted[i], oldSorted[j]) {
				toBeAdded = append(toBeAdded, newSorted[i])
				i++
			} else {
				toBeDeleted = append(toBeDeleted, oldSorted[j])
				j++
			}
		}
	}

	for ; i < len(newSorted); i++ {
		toBeAdded = append(toBeAdded, newSorted[i])
	}

	for ; j < len(oldSorted); j++ {
		toBeDeleted = append(toBeDeleted, oldSorted[j])
	}
	return
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
		if err := plugin.recreateRoute(route, vrf); err != nil {
			plugin.log.Errorf("Error recreating interface %s: %v", ifName, err)
		}
		plugin.rtCachedIndexes.UnregisterName(routeWithIndex.RouteID)
	}
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
func (plugin *RouteConfigurator) recreateRoute(route *l3.StaticRoutes_Route, vrf string) error {
	if err := plugin.DeleteRoute(route, vrf); err != nil {
		return nil
	}
	return plugin.ConfigureRoute(route, vrf)
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

func resolveInterfaceSwIndex(ifName string, index ifaceidx.SwIfIndex) (uint32, error) {
	ifIndex := vppcalls.NextHopOutgoingIfUnset
	if ifName != "" {
		var exists bool
		ifIndex, _, exists = index.LookupIdx(ifName)
		if !exists {
			return ifIndex, fmt.Errorf("route outgoing interface %v not found", ifName)
		}
	}
	return ifIndex, nil
}

func lessRoute(a, b *l3.StaticRoutes_Route) bool {
	if a.Type != b.Type {
		return a.Type < b.Type
	}
	if a.VrfId != b.VrfId {
		return a.VrfId < b.VrfId
	}
	if !strings.EqualFold(a.DstIpAddr, b.DstIpAddr) {
		return strings.Compare(a.DstIpAddr, b.DstIpAddr) < 0
	}
	if !strings.EqualFold(a.NextHopAddr, b.NextHopAddr) {
		return strings.Compare(a.NextHopAddr, b.NextHopAddr) < 0
	}
	if a.ViaVrfId != b.ViaVrfId {
		return a.ViaVrfId < b.ViaVrfId
	}
	if a.OutgoingInterface != b.OutgoingInterface {
		return a.OutgoingInterface < b.OutgoingInterface
	}
	if a.Preference != b.Preference {
		return a.Preference < b.Preference
	}
	return a.Weight < b.Weight

}
