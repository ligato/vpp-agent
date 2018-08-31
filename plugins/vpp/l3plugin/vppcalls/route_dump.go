// Copyright (c) 2018 Cisco and/or its affiliates.
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

package vppcalls

import (
	"bytes"
	"fmt"
	"net"
	"time"

	l3binapi "github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
)

// RouteDetails is object returned as a VPP dump. It contains static route data in proto format, and VPP-specific
// metadata
type RouteDetails struct {
	Route *l3.StaticRoutes_Route
	Meta  *RouteMeta
}

// RouteMeta holds fields returned from the VPP as details which are not in the model
type RouteMeta struct {
	TableName         string
	OutgoingIfIdx     uint32
	IsIPv6            bool
	Afi               uint8
	IsLocal           bool
	IsUDPEncap        bool
	IsUnreach         bool
	IsProhibit        bool
	IsResolveHost     bool
	IsResolveAttached bool
	IsDvr             bool
	IsSourceLookup    bool
	NextHopID         uint32
	RpfID             uint32
	LabelStack        []l3binapi.FibMplsLabel
}

// DumpStaticRoutes implements route handler.
func (handler *RouteHandler) DumpStaticRoutes() ([]*RouteDetails, error) {
	// IPFibDump time measurement
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(l3binapi.IPFibDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	var routes []*RouteDetails

	// Dump IPv4 l3 FIB.
	reqCtx := handler.callsChannel.SendMultiRequest(&l3binapi.IPFibDump{})
	for {
		fibDetails := &l3binapi.IPFibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ipv4Route, err := handler.dumpStaticRouteIPv4Details(fibDetails)
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipv4Route...)
	}

	// Dump IPv6 l3 FIB.
	reqCtx = handler.callsChannel.SendMultiRequest(&l3binapi.IP6FibDump{})
	for {
		fibDetails := &l3binapi.IP6FibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ipv6Route, err := handler.dumpStaticRouteIPv6Details(fibDetails)
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipv6Route...)
	}

	return routes, nil
}

func (handler *RouteHandler) dumpStaticRouteIPv4Details(fibDetails *l3binapi.IPFibDetails) ([]*RouteDetails, error) {
	return handler.dumpStaticRouteIPDetails(fibDetails.TableID, fibDetails.TableName, fibDetails.Address, fibDetails.AddressLength, fibDetails.Path, false)
}

func (handler *RouteHandler) dumpStaticRouteIPv6Details(fibDetails *l3binapi.IP6FibDetails) ([]*RouteDetails, error) {
	return handler.dumpStaticRouteIPDetails(fibDetails.TableID, fibDetails.TableName, fibDetails.Address, fibDetails.AddressLength, fibDetails.Path, true)
}

// dumpStaticRouteIPDetails processes static route details and returns a route objects. Number of routes returned
// depends on size of path list.
func (handler *RouteHandler) dumpStaticRouteIPDetails(tableID uint32, tableName []byte, address []byte, prefixLen uint8, paths []l3binapi.FibPath, ipv6 bool) ([]*RouteDetails, error) {
	// Common fields for every route path (destination IP, VRF)
	var dstIP string
	if ipv6 {
		dstIP = fmt.Sprintf("%s/%d", net.IP(address).To16().String(), uint32(prefixLen))
	} else {
		dstIP = fmt.Sprintf("%s/%d", net.IP(address[:4]).To4().String(), uint32(prefixLen))
	}

	var routeDetails []*RouteDetails

	// Paths
	if len(paths) > 0 {
		for _, path := range paths {
			// Next hop IP address
			var nextHopIP string
			if ipv6 {
				nextHopIP = fmt.Sprintf("%s", net.IP(path.NextHop).To16().String())
			} else {
				nextHopIP = fmt.Sprintf("%s", net.IP(path.NextHop[:4]).To4().String())
			}

			// Route type (if via VRF is used)
			var routeType l3.StaticRoutes_Route_RouteType
			var viaVrfID uint32
			if path.SwIfIndex == NextHopOutgoingIfUnset && path.TableID != tableID {
				// outgoing interface not specified and path table id not equal to route table id = inter-VRF route
				routeType = l3.StaticRoutes_Route_INTER_VRF
				viaVrfID = path.TableID
			} else if uintToBool(path.IsDrop) {
				routeType = l3.StaticRoutes_Route_DROP
			} else {
				routeType = l3.StaticRoutes_Route_INTRA_VRF // default
			}

			// Outgoing interface
			var ifName string
			var ifIdx uint32
			if path.SwIfIndex != ^uint32(0) {
				var exists bool
				ifIdx = path.SwIfIndex
				if ifName, _, exists = handler.ifIndexes.LookupName(path.SwIfIndex); !exists {
					handler.log.Warnf("Static route dump: interface name for index %d not found", path.SwIfIndex)
				}
			}

			// Route configuration
			route := &l3.StaticRoutes_Route{
				Type:              routeType,
				VrfId:             tableID,
				DstIpAddr:         dstIP,
				NextHopAddr:       nextHopIP,
				OutgoingInterface: ifName,
				Weight:            uint32(path.Weight),
				Preference:        uint32(path.Preference),
				ViaVrfId:          viaVrfID,
			}

			// Route metadata
			meta := &RouteMeta{
				TableName:         string(bytes.SplitN(tableName, []byte{0x00}, 2)[0]),
				OutgoingIfIdx:     ifIdx,
				NextHopID:         path.NextHopID,
				IsIPv6:            ipv6,
				RpfID:             path.RpfID,
				Afi:               path.Afi,
				IsLocal:           uintToBool(path.IsLocal),
				IsUDPEncap:        uintToBool(path.IsUDPEncap),
				IsDvr:             uintToBool(path.IsDvr),
				IsProhibit:        uintToBool(path.IsProhibit),
				IsResolveAttached: uintToBool(path.IsResolveAttached),
				IsResolveHost:     uintToBool(path.IsResolveHost),
				IsSourceLookup:    uintToBool(path.IsSourceLookup),
				IsUnreach:         uintToBool(path.IsUnreach),
				LabelStack:        path.LabelStack,
			}

			routeDetails = append(routeDetails, &RouteDetails{
				Route: route,
				Meta:  meta,
			})
		}
	} else {
		// Return route without path fields, but this is not a valid configuration
		handler.log.Warnf("Route with destination IP %s (VRF %d) has no path specified", dstIP, tableID)
		route := &l3.StaticRoutes_Route{
			Type:      l3.StaticRoutes_Route_INTRA_VRF, // default
			VrfId:     tableID,
			DstIpAddr: dstIP,
		}
		meta := &RouteMeta{
			TableName: string(bytes.SplitN(tableName, []byte{0x00}, 2)[0]),
		}
		routeDetails = append(routeDetails, &RouteDetails{
			Route: route,
			Meta:  meta,
		})
	}

	return routeDetails, nil
}
