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

package vpp2106

import (
	"fmt"
	"net"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/fib_types"
	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

// DumpRoutes implements route handler.
func (h *RouteHandler) DumpRoutes() (routes []*vppcalls.RouteDetails, err error) {
	// dump routes for every VRF and for both IP versions
	for _, vrfMeta := range h.vrfIndexes.ListAllVrfMetadata() {
		ipRoutes, err := h.dumpRoutesForVrfAndIP(vrfMeta.GetIndex(), vrfMeta.GetProtocol())
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipRoutes...)
	}
	return routes, nil
}

// dumpRoutesForVrf returns routes for given VRF and IP versiob
func (h *RouteHandler) dumpRoutesForVrfAndIP(vrfID uint32, proto l3.VrfTable_Protocol) (routes []*vppcalls.RouteDetails, err error) {
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_ip.IPRouteDump{
		Table: vpp_ip.IPTable{
			TableID: vrfID,
			IsIP6:   protoToUint(proto),
		},
	})
	for {
		fibDetails := &vpp_ip.IPRouteDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ipRoute, err := h.dumpRouteIPDetails(fibDetails.Route)
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipRoute...)
	}

	return routes, nil
}

// dumpRouteIPDetails processes static route details and returns a route objects. Number of routes returned
// depends on size of path list.
func (h *RouteHandler) dumpRouteIPDetails(ipRoute vpp_ip.IPRoute) ([]*vppcalls.RouteDetails, error) {
	// Common fields for every route path (destination IP, VRF)
	var dstIP string
	if ipRoute.Prefix.Address.Af == ip_types.ADDRESS_IP6 {
		ip6Addr := ipRoute.Prefix.Address.Un.GetIP6()
		dstIP = fmt.Sprintf("%s/%d", net.IP(ip6Addr[:]).To16().String(), uint32(ipRoute.Prefix.Len))
	} else {
		ip4Addr := ipRoute.Prefix.Address.Un.GetIP4()
		dstIP = fmt.Sprintf("%s/%d", net.IP(ip4Addr[:4]).To4().String(), uint32(ipRoute.Prefix.Len))
	}

	var routeDetails []*vppcalls.RouteDetails

	// Paths
	if ipRoute.NPaths > 0 {
		for _, path := range ipRoute.Paths {
			// Next hop IP address
			var nextHopIP string
			netIP := make([]byte, 16)
			copy(netIP[:], path.Nh.Address.XXX_UnionData[:])
			if path.Proto == fib_types.FIB_API_PATH_NH_PROTO_IP6 {
				nextHopIP = fmt.Sprintf("%s", net.IP(netIP).To16().String())
			} else {
				nextHopIP = fmt.Sprintf("%s", net.IP(netIP[:4]).To4().String())
			}

			// Route type (if via VRF is used)
			var routeType l3.Route_RouteType
			var viaVrfID uint32
			if path.Type == fib_types.FIB_API_PATH_TYPE_DROP {
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
				IsIPv6:        path.Proto == fib_types.FIB_API_PATH_NH_PROTO_IP6,
				NextHopID:     path.Nh.ObjID,
				RpfID:         path.RpfID,
				LabelStack:    labelStack,
			}
			resolvePathType(meta, path.Type)
			resolvePathFlags(meta, path.Flags)
			// Note: VPP does not return table name as in older versions, the field
			// is filled using index map
			vrfName, _, exists := h.vrfIndexes.LookupByVRFIndex(ipRoute.TableID)
			if exists {
				meta.TableName = vrfName
			}

			routeDetails = append(routeDetails, &vppcalls.RouteDetails{
				Route: route,
				Meta:  meta,
			})
		}
	} else {
		// Return route without path fields, but this is not a valid configuration
		h.log.Warnf("Route with destination IP %s (VRF %d) has no path specified", dstIP, ipRoute.TableID)
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

func resolvePathType(meta *vppcalls.RouteMeta, pathType fib_types.FibPathType) {
	switch pathType {
	case fib_types.FIB_API_PATH_TYPE_LOCAL:
		meta.IsLocal = true
	case fib_types.FIB_API_PATH_TYPE_UDP_ENCAP:
		meta.IsUDPEncap = true
	case fib_types.FIB_API_PATH_TYPE_ICMP_UNREACH:
		meta.IsUnreach = true
	case fib_types.FIB_API_PATH_TYPE_ICMP_PROHIBIT:
		meta.IsProhibit = true
	case fib_types.FIB_API_PATH_TYPE_DVR:
		meta.IsDvr = true
	case fib_types.FIB_API_PATH_TYPE_SOURCE_LOOKUP:
		meta.IsSourceLookup = true
	}
}

func resolvePathFlags(meta *vppcalls.RouteMeta, pathFlags fib_types.FibPathFlags) {
	switch pathFlags {
	case fib_types.FIB_API_PATH_FLAG_RESOLVE_VIA_HOST:
		meta.IsResolveHost = true
	case fib_types.FIB_API_PATH_FLAG_RESOLVE_VIA_ATTACHED:
		meta.IsResolveAttached = true
	}
}

func protoToUint(proto l3.VrfTable_Protocol) bool {
	if proto == l3.VrfTable_IPV6 {
		return true
	}
	return false
}
