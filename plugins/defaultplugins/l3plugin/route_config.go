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
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
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

const (
	// The constant that has to be assigned into the field next hop via label in ip_add_del_route binary message
	// if next hop via label is not defined.
	// equals to MPLS_LABEL_INVALID defined in VPP
	nextHopViaLabelUnset uint32 = 0xfffff + 1

	// Default value for field classify_table_index in ip_add_del_route binary message
	classifyTableIndexUnset uint32 = ^uint32(0)

	// The constant that has to be assigned into the field next_hop_outgoing_interface in ip_add_del_route binary message
	// if outgoing interface for next hop is not defined.
	nextHopOutgoingIfUnset uint32 = ^uint32(0)
)

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
	log.Infof("Creating new routes with destination address %v", config)
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
			routeIdentifier := routeIdentifier(route.destAddr.String(), route.nexthop.addr.String())
			plugin.RouteIndexes.RegisterName(routeIdentifier, plugin.RouteIndexSeq, nil)
			plugin.RouteIndexSeq++
			log.Infof("Route %v registered", routeIdentifier)
		}
	}

	return nil
}

// ModifyRoute process the NB config and propagates it to bin api calls
func (plugin *RouteConfigurator) ModifyRoute(newConfig *l3.StaticRoutes_Route, oldConfig *l3.StaticRoutes_Route) error {
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
		oldRouteIdentifier := routeIdentifier(oldRoute.destAddr.String(), oldRoute.nexthop.addr.String())
		plugin.RouteIndexes.UnregisterName(oldRouteIdentifier)
		log.Infof("Old route %v unregistered", oldRouteIdentifier)
	}
	for _, newRoute := range newRoutes {
		err := plugin.vppAddRoute(newRoute)
		if err != nil {
			return err
		}
		newRouteIdentifier := routeIdentifier(newRoute.destAddr.String(), newRoute.nexthop.addr.String())
		plugin.RouteIndexes.RegisterName(newRouteIdentifier, plugin.RouteIndexSeq, nil)
		plugin.RouteIndexSeq++
		log.Infof("New route %v registered", newRouteIdentifier)
	}

	return nil
}

// DeleteRoute process the NB config and propagates it to bin api calls
func (plugin *RouteConfigurator) DeleteRoute(config *l3.StaticRoutes_Route) (wasError error) {
	routes, err := plugin.transformRoute(config)
	if err != nil {
		return err
	}
	for _, route := range routes {
		err := plugin.vppDelRoute(route)

		if err != nil {
			return err
		}
		routeIdentifier := routeIdentifier(route.destAddr.String(), route.nexthop.addr.String())
		plugin.RouteIndexes.UnregisterName(routeIdentifier)
		log.Infof("Route %v unregistered", routeIdentifier)
	}

	return nil
}
func (plugin *RouteConfigurator) vppAddRoute(route *Route) error {
	log.WithField("Route", *route).Debug("Adding")
	return plugin.vppAddDelRoute(route, true)
}

func (plugin *RouteConfigurator) vppDelRoute(route *Route) error {
	log.WithField("Route", *route).Debug("Deleting")
	return plugin.vppAddDelRoute(route, false)
}

func (plugin *RouteConfigurator) vppAddDelRoute(route *Route, isAdd bool) error {
	// prepare the message
	req := &ip.IPAddDelRoute{}

	ipAddr := route.destAddr.IP
	isIpv6, err := addrs.IsIPv6(ipAddr.String())
	if err != nil {
		return err
	}
	prefix, _ := route.destAddr.Mask.Size()

	nextHopAddr := route.nexthop.addr
	req.NextHopAddress = []byte(nextHopAddr)
	if isAdd {
		req.IsAdd = 1
	} else {
		req.IsAdd = 0
	}
	req.ClassifyTableIndex = classifyTableIndexUnset
	req.DstAddressLength = byte(prefix)
	req.IsDrop = 0
	if isIpv6 {
		req.IsIpv6 = 1
		req.DstAddress = []byte(ipAddr.To16())
	} else {
		req.IsIpv6 = 0
		req.DstAddress = []byte(ipAddr.To4())
	}
	if route.multipath {
		req.IsMultipath = 1
	}
	req.NextHopSwIfIndex = route.nexthop.intf
	req.NextHopTableID = route.vrfID
	req.NextHopViaLabel = nextHopViaLabelUnset
	req.NextHopWeight = uint8(route.nexthop.weight)
	req.TableID = route.vrfID
	req.CreateVrfIfNeeded = 1
	reply := &ip.IPAddDelRouteReply{}
	err = plugin.vppChan.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("IPAddDelRoute returned %d", reply.Retval)
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
