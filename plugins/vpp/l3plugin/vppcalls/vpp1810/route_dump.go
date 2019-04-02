//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp1810

import (
	"bytes"
	"fmt"
	"net"

	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	l3binapi "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
)

// DumpRoutes implements route handler.
func (h *RouteHandler) DumpRoutes() ([]*vppcalls.RouteDetails, error) {
	var routes []*vppcalls.RouteDetails
	// Dump IPv4 l3 FIB.
	reqCtx := h.callsChannel.SendMultiRequest(&l3binapi.IPFibDump{})
	for {
		fibDetails := &l3binapi.IPFibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ipv4Route, err := h.dumpRouteIPv4Details(fibDetails)
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipv4Route...)
	}

	// Dump IPv6 l3 FIB.
	reqCtx = h.callsChannel.SendMultiRequest(&l3binapi.IP6FibDump{})
	for {
		fibDetails := &l3binapi.IP6FibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ipv6Route, err := h.dumpRouteIPv6Details(fibDetails)
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipv6Route...)
	}

	return routes, nil
}

func (h *RouteHandler) dumpRouteIPv4Details(fibDetails *l3binapi.IPFibDetails) ([]*vppcalls.RouteDetails, error) {
	return h.dumpRouteIPDetails(fibDetails.TableID, fibDetails.TableName, fibDetails.Address, fibDetails.AddressLength, fibDetails.Path, false)
}

func (h *RouteHandler) dumpRouteIPv6Details(fibDetails *l3binapi.IP6FibDetails) ([]*vppcalls.RouteDetails, error) {
	return h.dumpRouteIPDetails(fibDetails.TableID, fibDetails.TableName, fibDetails.Address, fibDetails.AddressLength, fibDetails.Path, true)
}

// dumpRouteIPDetails processes static route details and returns a route objects. Number of routes returned
// depends on size of path list.
func (h *RouteHandler) dumpRouteIPDetails(tableID uint32, tableName []byte, address []byte, prefixLen uint8, paths []l3binapi.FibPath, ipv6 bool) ([]*vppcalls.RouteDetails, error) {
	// Common fields for every route path (destination IP, VRF)
	var dstIP string
	if ipv6 {
		dstIP = fmt.Sprintf("%s/%d", net.IP(address).To16().String(), uint32(prefixLen))
	} else {
		dstIP = fmt.Sprintf("%s/%d", net.IP(address[:4]).To4().String(), uint32(prefixLen))
	}

	var routeDetails []*vppcalls.RouteDetails

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
			var routeType l3.Route_RouteType
			var viaVrfID uint32
			if uintToBool(path.IsDrop) {
				routeType = l3.Route_DROP
			} else if path.SwIfIndex == NextHopOutgoingIfUnset && path.TableID != tableID {
				// outgoing interface not specified and path table is not equal to route table id = inter-VRF route
				routeType = l3.Route_INTER_VRF
				viaVrfID = path.TableID
			} else {
				routeType = l3.Route_INTRA_VRF // default
			}

			// Outgoing interface
			var ifName string
			var ifIdx uint32
			if path.SwIfIndex == NextHopOutgoingIfUnset {
				ifIdx = NextHopOutgoingIfUnset
			} else {
				var exists bool
				ifIdx = path.SwIfIndex
				if ifName, _, exists = h.ifIndexes.LookupBySwIfIndex(path.SwIfIndex); !exists {
					h.log.Warnf("Static route dump: interface name for index %d not found", path.SwIfIndex)
				}
			}

			// Route configuration
			route := &l3.Route{
				Type:              routeType,
				VrfId:             tableID,
				DstNetwork:        dstIP,
				NextHopAddr:       nextHopIP,
				OutgoingInterface: ifName,
				Weight:            uint32(path.Weight),
				Preference:        uint32(path.Preference),
				ViaVrfId:          viaVrfID,
			}

			labelStack := make([]vppcalls.FibMplsLabel, len(path.LabelStack))
			for i, l := range path.LabelStack {
				labelStack[i] = vppcalls.FibMplsLabel{
					IsUniform: uintToBool(l.IsUniform),
					Label:     l.Label,
					TTL:       l.TTL,
					Exp:       l.Exp,
				}
			}

			// Route metadata
			meta := &vppcalls.RouteMeta{
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
				LabelStack:        labelStack,
			}

			routeDetails = append(routeDetails, &vppcalls.RouteDetails{
				Route: route,
				Meta:  meta,
			})
		}
	} else {
		// Return route without path fields, but this is not a valid configuration
		h.log.Warnf("Route with destination IP %s (VRF %d) has no path specified", dstIP, tableID)
		route := &l3.Route{
			Type:       l3.Route_INTRA_VRF, // default
			VrfId:      tableID,
			DstNetwork: dstIP,
		}
		meta := &vppcalls.RouteMeta{
			TableName: string(bytes.SplitN(tableName, []byte{0x00}, 2)[0]),
		}
		routeDetails = append(routeDetails, &vppcalls.RouteDetails{
			Route: route,
			Meta:  meta,
		})
	}

	return routeDetails, nil
}
