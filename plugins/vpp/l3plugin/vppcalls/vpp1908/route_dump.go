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

package vpp1908

import (
	"fmt"
	"net"

	"github.com/ligato/cn-infra/logging"

	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	l3binapi "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
)

// DumpRoutes implements route handler.
func (h *RouteHandler) DumpRoutes() ([]*vppcalls.RouteDetails, error) {
	var routes []*vppcalls.RouteDetails
	// Dump IPv4 l3 FIB.
	reqCtx := h.callsChannel.SendMultiRequest(&l3binapi.IPRouteDump{})
	for {
		fibDetails := &l3binapi.IPRouteDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ipv4Route, err := h.dumpRouteIPDetails(fibDetails.Route)
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipv4Route...)
	}

	return routes, nil
}

// dumpRouteIPDetails processes static route details and returns a route objects. Number of routes returned
// depends on size of path list.
func (h *RouteHandler) dumpRouteIPDetails(ipRoute l3binapi.IPRoute) ([]*vppcalls.RouteDetails, error) {
	// Common fields for every route path (destination IP, VRF)
	var dstIP string
	netIP := make([]byte, 16)
	copy(netIP[:], ipRoute.Prefix.Address.Un.XXX_UnionData[:])
	if ipRoute.Prefix.Address.Af == l3binapi.ADDRESS_IP6 {
		dstIP = fmt.Sprintf("%s/%d", net.IP(netIP).To16().String(), uint32(ipRoute.Prefix.Len))
	} else {
		dstIP = fmt.Sprintf("%s/%d", net.IP(netIP[:4]).To4().String(), uint32(ipRoute.Prefix.Len))
	}

	var routeDetails []*vppcalls.RouteDetails

	// Paths
	if ipRoute.NPaths > 0 {
		for _, path := range ipRoute.Paths {
			// Next hop IP address
			var nextHopIP string
			netIP := make([]byte, 16)
			copy(netIP[:], path.Nh.Address.XXX_UnionData[:])
			logging.DefaultLogger.Warnf("netip: %v, proto %v", path.Nh.Address.XXX_UnionData, path.Proto)
			if path.Proto == l3binapi.FIB_API_PATH_NH_PROTO_IP6 {
				nextHopIP = fmt.Sprintf("%s", net.IP(netIP).To16().String())
			} else {
				nextHopIP = fmt.Sprintf("%s", net.IP(netIP[:4]).To4().String())
			}

			// Route type (if via VRF is used)
			var routeType l3.Route_RouteType
			var viaVrfID uint32
			if path.Type == l3binapi.FIB_API_PATH_TYPE_DROP {
				routeType = l3.Route_DROP
			} else if path.SwIfIndex == NextHopOutgoingIfUnset && path.TableID != ipRoute.TableID {
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
				VrfId:             ipRoute.TableID,
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
				OutgoingIfIdx: ifIdx,
				RpfID:         path.RpfID,
				LabelStack:    labelStack,
			}

			routeDetails = append(routeDetails, &vppcalls.RouteDetails{
				Route: route,
				Meta:  meta,
			})
		}
	} else {
		// Return route without path fields, but this is not a valid configuration
		routeDetails = append(routeDetails, &vppcalls.RouteDetails{
			Route: &l3.Route{
				Type:       l3.Route_INTRA_VRF, // default
				VrfId:      ipRoute.TableID,
				DstNetwork: dstIP,
			},
		})
	}

	return routeDetails, nil
}
