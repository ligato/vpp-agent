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

package vppcalls

import (
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/bin_api/ip"
	"net"
)

// Route represents a forward IP route entry.
type Route struct {
	VrfID     uint32
	DstAddr   net.IPNet
	MultiPath bool
	NextHop   NextHopList
}

// NextHopList defines the parameters of gateway to which packets should be forwarded
// when a given routing table entry is applied.
type NextHopList struct {
	Addr   net.IP
	Iface  uint32
	Weight uint32
}

const (
	// NextHopViaLabelUnset constant has to be assigned into the field next hop via label in ip_add_del_route binary message
	// if next hop via label is not defined.
	// equals to MPLS_LABEL_INVALID defined in VPP
	NextHopViaLabelUnset uint32 = 0xfffff + 1

	// ClassifyTableIndexUnset is a default value for field classify_table_index in ip_add_del_route binary message
	ClassifyTableIndexUnset uint32 = ^uint32(0)

	// NextHopOutgoingIfUnset constant has to be assigned into the field next_hop_outgoing_interface in ip_add_del_route binary message
	// if outgoing interface for next hop is not defined.
	NextHopOutgoingIfUnset uint32 = ^uint32(0)
)

// VppAddRoute adds new route according to provided input. Every route has to contain VRF ID (default is 0)
func VppAddRoute(route *Route, vppChan *govppapi.Channel) error {
	req := &ip.IPAddDelRoute{}
	req.IsAdd = 1

	// Destination address (route set identifier)
	ipAddr := route.DstAddr.IP
	prefix, _ := route.DstAddr.Mask.Size()
	isIpv6, err := addrs.IsIPv6(ipAddr.String())
	if err != nil {
		return err
	}
	if isIpv6 {
		req.IsIpv6 = 1
		req.DstAddress = []byte(ipAddr.To16())
	} else {
		req.IsIpv6 = 0
		req.DstAddress = []byte(ipAddr.To4())
	}
	req.DstAddressLength = byte(prefix)

	// Enable multipath if desired
	if route.MultiPath {
		req.IsMultipath = 1
	}

	// Next hop address and parameters
	nextHopAddr := route.NextHop.Addr
	req.NextHopAddress = []byte(nextHopAddr)
	req.NextHopSwIfIndex = route.NextHop.Iface
	req.NextHopWeight = uint8(route.NextHop.Weight)
	req.NextHopTableID = route.VrfID
	req.NextHopViaLabel = NextHopViaLabelUnset
	req.ClassifyTableIndex = ClassifyTableIndexUnset
	req.IsDrop = 0

	// VRF
	req.CreateVrfIfNeeded = 1
	req.TableID = route.VrfID

	// Send message
	reply := &ip.IPAddDelRouteReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("IPAddDelRoute returned %d", reply.Retval)
	}

	return nil
}

// VppDelRoute removes route from config
func VppDelRoute(route *Route, vppChan *govppapi.Channel) error {
	req := &ip.IPAddDelRoute{}
	req.IsAdd = 0

	// Destination address (route set identifier)
	ipAddr := route.DstAddr.IP
	isIpv6, err := addrs.IsIPv6(ipAddr.String())
	if err != nil {
		return err
	}
	if isIpv6 {
		req.IsIpv6 = 1
		req.DstAddress = []byte(ipAddr.To16())
	} else {
		req.IsIpv6 = 0
		req.DstAddress = []byte(ipAddr.To4())
	}

	// Send message
	reply := &ip.IPAddDelRouteReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("IPAddDelRoute returned %d", reply.Retval)
	}

	return nil
}
