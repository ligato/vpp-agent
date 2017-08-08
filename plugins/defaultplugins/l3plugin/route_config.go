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

// Package l3plugin is the implementation of the L3 plugin that handles ip routes.
package l3plugin

import (
	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"

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
	GoVppmux      *govppmux.GOVPPPlugin
	RouteIndexes  idxvpp.NameToIdxRW
	RouteIndexSeq uint32
	SwIfIndexes   ifaceidx.SwIfIndex
	vppChan       *govppapi.Channel
}

// Init members (channels...) and start go routines
func (plugin *RouteConfigurator) Init() (err error) {
	log.Debug("Initializing L3 plugin")

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

// ConfigureRoutes process the NB config and propagates it to bin api calls
func (plugin *RouteConfigurator) ConfigureRoutes(config *l3.StaticRoutes_Route) error {
	log.Infof("Creating new route set with destination address %v", config.DestinationAddress)
	routes, err := plugin.transformRoute(config)
	if err != nil {
		return err
	}
	if len(routes) > 0 {
		for _, route := range routes {
			err := plugin.vppAddRoute(route)
			if err != nil {
				return err
			}
			routeIdentifier := routeIdentifier(route.DstAddr.String(), route.NextHop.Addr.String())
			plugin.RouteIndexes.RegisterName(routeIdentifier, plugin.RouteIndexSeq, nil)
			plugin.RouteIndexSeq++
			log.Infof("Route %v registered", routeIdentifier)
		}
	}

	return nil
}

// ModifyRoute process the NB config and propagates it to bin api calls
func (plugin *RouteConfigurator) ModifyRoute(newConfig *l3.StaticRoutes_Route, oldConfig *l3.StaticRoutes_Route) error {
	log.Infof("Modifying route set with destination address %v", oldConfig.DestinationAddress)
	newRoutes, err := plugin.transformRoute(newConfig)
	if err != nil {
		return err
	}
	oldRoutes, err := plugin.transformRoute(oldConfig)
	if err != nil {
		return err
	}

	for _, oldRoute := range oldRoutes {
		err := plugin.vppDelRoute(oldRoute)
		if err != nil {
			return err
		}
		oldRouteIdentifier := routeIdentifier(oldRoute.DstAddr.String(), oldRoute.NextHop.Addr.String())
		plugin.RouteIndexes.UnregisterName(oldRouteIdentifier)
		log.Infof("Old route %v unregistered", oldRouteIdentifier)
	}
	for _, newRoute := range newRoutes {
		err := plugin.vppAddRoute(newRoute)
		if err != nil {
			return err
		}
		newRouteIdentifier := routeIdentifier(newRoute.DstAddr.String(), newRoute.NextHop.Addr.String())
		plugin.RouteIndexes.RegisterName(newRouteIdentifier, plugin.RouteIndexSeq, nil)
		plugin.RouteIndexSeq++
		log.Infof("New route %v registered", newRouteIdentifier)
	}

	return nil
}

// DeleteRoute process the NB config and propagates it to bin api calls
func (plugin *RouteConfigurator) DeleteRoute(config *l3.StaticRoutes_Route) (wasError error) {
	log.Infof("Removing route set with destination address %v", config.DestinationAddress)
	routes, err := plugin.transformRoute(config)
	if err != nil {
		return err
	}
	for _, route := range routes {
		err := plugin.vppDelRoute(route)

		if err != nil {
			return err
		}
		routeIdentifier := routeIdentifier(route.DstAddr.String(), route.NextHop.Addr.String())
		plugin.RouteIndexes.UnregisterName(routeIdentifier)
		log.Infof("Route %v unregistered", routeIdentifier)
	}

	return nil
}
func (plugin *RouteConfigurator) vppAddRoute(route *vppcalls.Route) error {
	log.WithField("Route", *route).Debug("Adding")
	return vppcalls.VppAddRoute(route, plugin.vppChan)
}

func (plugin *RouteConfigurator) vppDelRoute(route *vppcalls.Route) error {
	log.WithField("Route", *route).Debug("Deleting")
	return vppcalls.VppDelRoute(route, plugin.vppChan)
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
		log.Error(err)
	}
	return err
}

// Close GOVPP channel
func (plugin *RouteConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// Creates unique identifier which serves as a name for index mapping
func routeIdentifier(destination string, nextHop string) string {
	return destination + "-" + nextHop
}
