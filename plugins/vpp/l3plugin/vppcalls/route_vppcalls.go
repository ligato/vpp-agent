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
	"net"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

var RouteMessages = []govppapi.Message{
	&ip.IPAddDelRoute{},
	&ip.IPAddDelRouteReply{},
	&ip.IPFibDump{},
	&ip.IPFibDetails{},
	&ip.IP6FibDump{},
	&ip.IP6FibDetails{},
}

// Route represents a forward IP route entry with the parameters of gateway
// to which packets should be forwarded when a given routing table entry is applied.
type Route struct {
	VrfID       uint32    `json:"vrf_id"`
	TableName   string    `json:"table_name"`
	DstAddr     net.IPNet `json:"dst_addr"`
	NextHopAddr net.IP    `json:"next_hop_addr"`
	OutIface    uint32    `json:"out_iface"`
	Weight      uint32    `json:"weight"`
	Preference  uint32    `json:"preference"`
}

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
func vppAddDelRoute(route *Route, vppChan govppapi.Channel, delete bool, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ip.IPAddDelRoute{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ip.IPAddDelRoute{}
	if delete {
		req.IsAdd = 0
	} else {
		req.IsAdd = 1
	}

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

	// Next hop address and parameters
	req.NextHopAddress = []byte(route.NextHopAddr)
	req.NextHopSwIfIndex = route.OutIface
	req.NextHopWeight = uint8(route.Weight)
	req.NextHopPreference = uint8(route.Preference)
	req.NextHopTableID = route.VrfID
	req.NextHopViaLabel = NextHopViaLabelUnset
	req.ClassifyTableIndex = ClassifyTableIndexUnset
	req.IsDrop = 0

	// VRF
	req.TableID = route.VrfID

	// Multi path is always true
	req.IsMultipath = 1

	// Send message
	reply := &ip.IPAddDelRouteReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// VppAddRoute adds new route, according to provided input. Every route has to contain VRF ID (default is 0).
func VppAddRoute(route *Route, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	if err := ifvppcalls.CreateVrfIfNeeded(route.VrfID, vppChan); err != nil {
		return err
	}
	return vppAddDelRoute(route, vppChan, false, stopwatch)
}

// VppDelRoute removes old route, according to provided input. Every route has to contain VRF ID (default is 0).
func VppDelRoute(route *Route, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return vppAddDelRoute(route, vppChan, true, stopwatch)
}
