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
	"context"
	"net"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/fib_types"
	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	// NextHopViaLabelUnset constant has to be assigned into the field next hop
	// via label in ip_add_del_route binary message if next hop via label is not defined.
	// Equals to MPLS_LABEL_INVALID defined in VPP
	NextHopViaLabelUnset uint32 = 0xfffff + 1

	// ClassifyTableIndexUnset is a default value for field classify_table_index in ip_add_del_route binary message.
	ClassifyTableIndexUnset = ^uint32(0)

	// NextHopOutgoingIfUnset constant has to be assigned into the field next_hop_outgoing_interface
	// in ip_add_del_route binary message if outgoing interface for next hop is not defined.
	NextHopOutgoingIfUnset = ^uint32(0)
)

// vppAddDelRoute adds or removes route, according to provided input. Every route has to contain VRF ID (default is 0).
func (h *RouteHandler) vppAddDelRoute(route *l3.Route, rtIfIdx uint32, delete bool) error {
	req := &vpp_ip.IPRouteAddDel{
		// Multi path is always true
		IsMultipath: true,
		IsAdd:       !delete,
	}

	// Common route parameters
	fibPath := fib_types.FibPath{
		Weight:     uint8(route.Weight),
		Preference: uint8(route.Preference),
	}
	if route.NextHopAddr != "" {
		nextHop, err := h.addrAlloc.GetOrParseIPAddress(route.NextHopAddr,
			route.OutgoingInterface, netalloc.IPAddressForm_ADDR_ONLY)
		if err != nil {
			return err
		}
		fibPath.Nh, fibPath.Proto = setFibPathNhAndProto(nextHop.IP)
	}

	// VRF/Other route parameters based on type
	if route.Type == l3.Route_INTER_VRF {
		fibPath.SwIfIndex = rtIfIdx
		fibPath.TableID = route.ViaVrfId
	} else if route.Type == l3.Route_DROP {
		fibPath.Type = fib_types.FIB_API_PATH_TYPE_DROP
	} else {
		fibPath.SwIfIndex = rtIfIdx
		fibPath.TableID = route.VrfId
	}
	// Destination address
	dstNet, err := h.addrAlloc.GetOrParseIPAddress(route.DstNetwork,
		"", netalloc.IPAddressForm_ADDR_NET)
	if err != nil {
		return err
	}
	prefix := networkToPrefix(dstNet)

	req.Route = vpp_ip.IPRoute{
		TableID: route.VrfId,
		Prefix:  prefix,
		NPaths:  1,
		Paths:   []fib_types.FibPath{fibPath},
	}

	reply := &vpp_ip.IPRouteAddDelReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// VppAddRoute implements route handler.
func (h *RouteHandler) VppAddRoute(ctx context.Context, route *l3.Route) error {
	swIfIdx, err := h.getRouteSwIfIndex(route.OutgoingInterface)
	if err != nil {
		return err
	}

	return h.vppAddDelRoute(route, swIfIdx, false)
}

// VppDelRoute implements route handler.
func (h *RouteHandler) VppDelRoute(ctx context.Context, route *l3.Route) error {
	swIfIdx, err := h.getRouteSwIfIndex(route.OutgoingInterface)
	if err != nil {
		return err
	}

	return h.vppAddDelRoute(route, swIfIdx, true)
}

func setFibPathNhAndProto(netIP net.IP) (nh fib_types.FibPathNh, proto fib_types.FibPathNhProto) {
	var addrUnion ip_types.AddressUnion
	if netIP.To4() == nil {
		proto = fib_types.FIB_API_PATH_NH_PROTO_IP6
		var ip6addr ip_types.IP6Address
		copy(ip6addr[:], netIP.To16())
		addrUnion.SetIP6(ip6addr)
	} else {
		proto = fib_types.FIB_API_PATH_NH_PROTO_IP4
		var ip4addr ip_types.IP4Address
		copy(ip4addr[:], netIP.To4())
		addrUnion.SetIP4(ip4addr)
	}

	return fib_types.FibPathNh{
		Address:            addrUnion,
		ViaLabel:           NextHopViaLabelUnset,
		ClassifyTableIndex: ClassifyTableIndexUnset,
	}, proto
}

func (h *RouteHandler) getRouteSwIfIndex(ifName string) (swIfIdx uint32, err error) {
	swIfIdx = NextHopOutgoingIfUnset
	if ifName != "" {
		meta, found := h.ifIndexes.LookupByName(ifName)
		if !found {
			return 0, errors.Errorf("interface %s not found", ifName)
		}
		swIfIdx = meta.SwIfIndex
	}
	return
}
