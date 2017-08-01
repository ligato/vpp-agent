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
	vppChan     *govppapi.Channel
	SwIfIndexes ifaceidx.SwIfIndex
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
	plugin.vppChan, err = govppmux.NewAPIChannel()
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
func (plugin *RouteConfigurator) ConfigureRoutes(config *l3.StaticRoutes) (wasError error) {
	routes := plugin.protoRoutesToStruct(config)

	for i := range routes {
		err := plugin.vppAddRoute(routes[i])
		if err != nil {
			wasError = err
		}
	}

	return wasError
}

// ModifyRoutes process the NB config and propagates it to bin api calls
func (plugin *RouteConfigurator) ModifyRoutes(newConfig *l3.StaticRoutes, oldConfig *l3.StaticRoutes) (
	wasError error) {
	newRoutes := plugin.protoRoutesToStruct(newConfig)
	oldRoutes := plugin.protoRoutesToStruct(oldConfig)
	toBeDeleted, toBeAdded := plugin.diffRoutes(newRoutes, oldRoutes)
	for i := range toBeDeleted {
		err := plugin.vppDelRoute(toBeDeleted[i])
		if err != nil {
			wasError = err
		}
	}
	for i := range toBeAdded {
		err := plugin.vppAddRoute(toBeAdded[i])
		if err != nil {
			wasError = err
		}
	}

	return wasError
}

// DeleteRoutes process the NB config and propagates it to bin api calls
func (plugin *RouteConfigurator) DeleteRoutes(config *l3.StaticRoutes) (wasError error) {
	routes := plugin.protoRoutesToStruct(config)
	for i := range routes {
		err := plugin.vppDelRoute(routes[i])
		if err != nil {
			wasError = err
		}
	}

	return wasError
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
	if isAdd {
		req.IsAdd = 1
	} else {

		req.IsAdd = 0
	}
	isIpv6, err := addrs.IsIPv6(route.destAddr.IP.String())
	if err != nil {
		return err
	}
	if isIpv6 {
		req.IsIpv6 = 1
		req.DstAddress = []byte(route.destAddr.IP.To16())
	} else {
		req.IsIpv6 = 0
		req.DstAddress = []byte(route.destAddr.IP.To4())
	}
	prefix, _ := route.destAddr.Mask.Size()
	req.DstAddressLength = byte(prefix)
	req.TableID = route.vrfID
	req.ClassifyTableIndex = classifyTableIndexUnset

	req.NextHopAddress = []byte(route.nexthop.addr)
	req.NextHopSwIfIndex = route.nexthop.intf
	req.NextHopWeight = uint8(route.nexthop.weight)
	req.NextHopTableID = route.vrfID

	req.NextHopViaLabel = nextHopViaLabelUnset
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
