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

//go:generate protoc --proto_path=model/l3 --gogo_out=model/l3 model/l3/l3.proto
//go:generate binapi-generator --input-file=/usr/share/vpp/api/ip.api.json --output-dir=bin_api

// Package l3plugin implements the L3 plugin that handles L3 FIBs.
package l3plugin

import (
	"fmt"
	"strconv"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// RouteConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of L3 routes as modelled by the proto file "../model/l3/l3.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1routes". Updates received from the northbound API
// are compared with the VPP run-time configuration and differences are applied through the VPP binary API.
type RouteConfigurator struct {
	Log           logging.Logger
	GoVppmux      govppmux.API
	RouteIndexes  idxvpp.NameToIdxRW
	RouteIndexSeq uint32
	SwIfIndexes   ifaceidx.SwIfIndex
	vppChan       *govppapi.Channel
	Stopwatch     *measure.Stopwatch // timer used to measure and store time
}

// Init members (channels...) and start go routines
func (plugin *RouteConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing L3 plugin")

	// Init VPP API channel
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

// ConfigureRoute process the NB config and propagates it to bin api calls
func (plugin *RouteConfigurator) ConfigureRoute(config *l3.StaticRoutes_Route, vrfFromKey string) error {
	plugin.Log.Infof("Creating new route %v -> %v", config.DstIpAddr, config.NextHopAddr)
	// Validate VRF index from key and it's value in data
	if err := plugin.validateVrfFromKey(config, vrfFromKey); err != nil {
		return err
	}
	// Transform route data
	route, err := TransformRoute(config, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	plugin.Log.Debugf("adding route: %+v", route)
	// Create and register new route
	if route != nil {
		err := vppcalls.VppAddRoute(route, plugin.vppChan, measure.GetTimeLog(ip.IPAddDelRoute{}, plugin.Stopwatch))
		if err != nil {
			return err
		}
		routeIdentifier := routeIdentifier(route.VrfID, route.DstAddr.String(), route.NextHopAddr.String())
		plugin.RouteIndexes.RegisterName(routeIdentifier, plugin.RouteIndexSeq, nil)
		plugin.RouteIndexSeq++
		plugin.Log.Infof("Route %v registered", routeIdentifier)
	}

	return nil
}

// ModifyRoute process the NB config and propagates it to bin api calls
func (plugin *RouteConfigurator) ModifyRoute(newConfig *l3.StaticRoutes_Route, oldConfig *l3.StaticRoutes_Route, vrfFromKey string) error {
	plugin.Log.Infof("Modifying route %v -> %v", oldConfig.DstIpAddr, oldConfig.NextHopAddr)
	// Validate old route data Vrf
	if err := plugin.validateVrfFromKey(oldConfig, vrfFromKey); err != nil {
		return err
	}
	// Transform old route data
	oldRoute, err := TransformRoute(oldConfig, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	// Remove and unregister old route
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

	// Validate new route data Vrf
	if err := plugin.validateVrfFromKey(newConfig, vrfFromKey); err != nil {
		return err
	}
	// Transform new route data
	newRoute, err := TransformRoute(newConfig, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	// Create and register new route
	err = vppcalls.VppAddRoute(newRoute, plugin.vppChan, measure.GetTimeLog(ip.IPAddDelRoute{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	newRouteIdentifier := routeIdentifier(newRoute.VrfID, newRoute.DstAddr.String(), newRoute.NextHopAddr.String())
	plugin.RouteIndexes.RegisterName(newRouteIdentifier, plugin.RouteIndexSeq, nil)
	plugin.RouteIndexSeq++
	plugin.Log.Infof("New route %v registered", newRouteIdentifier)

	return nil
}

// DeleteRoute process the NB config and propagates it to bin api calls
func (plugin *RouteConfigurator) DeleteRoute(config *l3.StaticRoutes_Route, vrfFromKey string) (wasError error) {
	plugin.Log.Infof("Removing route %v -> %v", config.DstIpAddr, config.NextHopAddr)
	// Validate VRF index from key and it's value in data
	if err := plugin.validateVrfFromKey(config, vrfFromKey); err != nil {
		return err
	}
	// Transform route data
	route, err := TransformRoute(config, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	if route == nil {
		return nil
	}
	plugin.Log.Debugf("deleting route: %+v", route)
	// Remove and unregister route
	err = vppcalls.VppDelRoute(route, plugin.vppChan, measure.GetTimeLog(ip.IPAddDelRoute{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	routeIdentifier := routeIdentifier(route.VrfID, route.DstAddr.String(), route.NextHopAddr.String())
	_, _, found := plugin.RouteIndexes.UnregisterName(routeIdentifier)
	if found {
		plugin.Log.Infof("Route %v unregistered", routeIdentifier)
	} else {
		plugin.Log.Warnf("Unregister failed, route %v not found", routeIdentifier)
	}

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

// Close GOVPP channel
func (plugin *RouteConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// Creates unique identifier which serves as a name in name to index mapping
func routeIdentifier(vrf uint32, destination string, nextHop string) string {
	return fmt.Sprintf("vrf%v-%v-%v", vrf, destination, nextHop)
}
