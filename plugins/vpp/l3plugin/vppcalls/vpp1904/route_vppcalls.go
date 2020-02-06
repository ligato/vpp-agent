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

package vpp1904

import (
	context "context"
	"net"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/ip"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
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
func (h *RouteHandler) vppAddDelRoute(route *vpp_l3.Route, rtIfIdx uint32, delete bool) error {
	req := &ip.IPAddDelRoute{}
	if delete {
		req.IsAdd = 0
	} else {
		req.IsAdd = 1
	}

	// Destination address (route set identifier)
	parsedDstIP, err := h.addrAlloc.GetOrParseIPAddress(route.DstNetwork,
		"", netalloc.IPAddressForm_ADDR_NET)
	if err != nil {
		return err
	}
	isIpv6 := parsedDstIP.IP.To4() == nil
	prefix, _ := parsedDstIP.Mask.Size()

	// Next Hop IP address
	var parsedNextHopIP net.IP
	if route.NextHopAddr != "" {
		nextHopIPNet, err := h.addrAlloc.GetOrParseIPAddress(route.NextHopAddr,
			route.OutgoingInterface, netalloc.IPAddressForm_ADDR_ONLY)
		if err != nil {
			return err
		}
		parsedNextHopIP = nextHopIPNet.IP
	}

	if isIpv6 {
		req.IsIPv6 = 1
		req.DstAddress = parsedDstIP.IP.To16()
		req.NextHopAddress = parsedNextHopIP.To16()
	} else {
		req.IsIPv6 = 0
		req.DstAddress = parsedDstIP.IP.To4()
		req.NextHopAddress = parsedNextHopIP.To4()
	}
	req.DstAddressLength = byte(prefix)

	// Common route parameters
	req.NextHopWeight = uint8(route.Weight)
	req.NextHopPreference = uint8(route.Preference)
	req.NextHopViaLabel = NextHopViaLabelUnset
	req.ClassifyTableIndex = ClassifyTableIndexUnset

	// VRF/Other route parameters based on type
	req.TableID = route.VrfId
	if route.Type == vpp_l3.Route_INTER_VRF {
		req.NextHopSwIfIndex = rtIfIdx
		req.NextHopTableID = route.ViaVrfId
	} else if route.Type == vpp_l3.Route_DROP {
		req.IsDrop = 1
	} else {
		req.NextHopSwIfIndex = rtIfIdx
		req.NextHopTableID = route.VrfId
	}

	// Multi path is always true
	req.IsMultipath = 1

	// Send message
	reply := &ip.IPAddDelRouteReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// VppAddRoute implements route handler.
func (h *RouteHandler) VppAddRoute(ctx context.Context, route *vpp_l3.Route) error {
	swIfIdx, err := h.getRouteSwIfIndex(route.OutgoingInterface)
	if err != nil {
		return err
	}

	return h.vppAddDelRoute(route, swIfIdx, false)
}

// VppDelRoute implements route handler.
func (h *RouteHandler) VppDelRoute(ctx context.Context, route *vpp_l3.Route) error {
	swIfIdx, err := h.getRouteSwIfIndex(route.OutgoingInterface)
	if err != nil {
		return err
	}

	return h.vppAddDelRoute(route, swIfIdx, true)
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
