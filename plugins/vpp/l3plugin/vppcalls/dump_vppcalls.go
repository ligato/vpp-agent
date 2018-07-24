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
	"bytes"
	"fmt"
	"net"

	"time"

	"github.com/ligato/cn-infra/utils/addrs"
	l3binapi "github.com/ligato/vpp-agent/plugins/vpp/binapi/ip"
)

func (handler *routeHandler) DumpStaticRoutes() ([]*Route, error) {
	// IPFibDump time measurement
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(l3binapi.IPFibDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	var routes []*Route

	// Dump IPv4 l3 FIB.
	reqCtx := handler.callsChannel.SendMultiRequest(&l3binapi.IPFibDump{})
	for {
		fibDetails := &l3binapi.IPFibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return nil, err
		}
		if len(fibDetails.Path) > 0 && fibDetails.Path[0].IsDrop == 1 {
			// skip drop routes, not supported by vpp-agent
			continue
		}
		ipv4Route, err := handler.dumpStaticRouteIPv4Details(fibDetails)
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipv4Route)
	}

	// Dump IPv6 l3 FIB.
	reqCtx = handler.callsChannel.SendMultiRequest(&l3binapi.IP6FibDump{})
	for {
		fibDetails := &l3binapi.IP6FibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			return nil, err
		}
		if len(fibDetails.Path) > 0 && fibDetails.Path[0].IsDrop == 1 {
			// skip drop routes, not supported by vpp-agent
			continue
		}
		ipv6Route, err := handler.dumpStaticRouteIPv6Details(fibDetails)
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipv6Route)
	}

	return routes, nil
}

func (handler *routeHandler) dumpStaticRouteIPv4Details(fibDetails *l3binapi.IPFibDetails) (*Route, error) {
	return handler.dumpStaticRouteIPDetails(fibDetails.TableID, fibDetails.TableName, fibDetails.Address, fibDetails.AddressLength, fibDetails.Path, false)
}

func (handler *routeHandler) dumpStaticRouteIPv6Details(fibDetails *l3binapi.IP6FibDetails) (*Route, error) {
	return handler.dumpStaticRouteIPDetails(fibDetails.TableID, fibDetails.TableName, fibDetails.Address, fibDetails.AddressLength, fibDetails.Path, true)
}

// dumpStaticRouteIPDetails processes static route details and returns a route object
func (handler *routeHandler) dumpStaticRouteIPDetails(tableID uint32, tableName []byte, address []byte, prefixLen uint8, path []l3binapi.FibPath, ipv6 bool) (*Route, error) {
	// route details
	var ipAddr string
	if ipv6 {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(address).To16().String(), uint32(prefixLen))
	} else {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(address[:4]).To4().String(), uint32(prefixLen))
	}

	rt := &Route{
		Type: IntraVrf, // default
	}

	// IP net
	parsedIP, _, err := addrs.ParseIPWithPrefix(ipAddr)
	if err != nil {
		return nil, err
	}

	rt.TableName = string(bytes.SplitN(tableName, []byte{0x00}, 2)[0])
	rt.VrfID = tableID
	rt.DstAddr = *parsedIP

	if len(path) > 0 {
		// TODO: if len(path) > 1, it means multiple NB routes (load-balancing) - not implemented properly

		var nextHopAddr net.IP
		if ipv6 {
			nextHopAddr = net.IP(path[0].NextHop).To16()
		} else {
			nextHopAddr = net.IP(path[0].NextHop[:4]).To4()
		}

		rt.NextHopAddr = nextHopAddr

		if path[0].SwIfIndex == NextHopOutgoingIfUnset && path[0].TableID != tableID {
			// outgoing interface not specified and path table id not equal to route table id = inter-VRF route
			rt.Type = InterVrf
			rt.ViaVrfId = path[0].TableID
		}

		rt.OutIface = path[0].SwIfIndex
		rt.Preference = uint32(path[0].Preference)
		rt.Weight = uint32(path[0].Weight)
	}

	return rt, nil
}

func (handler *arpVppHandler) DumpArpEntries() ([]*ArpEntry, error) {
	// ArpDump time measurement
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(l3binapi.IPFibDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	var arps []*ArpEntry

	// Dump ARPs.
	reqCtx := handler.callsChannel.SendMultiRequest(&l3binapi.IPNeighborDump{
		SwIfIndex: 0xffffffff,
	})
	for {
		arpDetails := &l3binapi.IPNeighborDetails{}
		stop, err := reqCtx.ReceiveReply(arpDetails)
		if stop {
			break
		}
		if err != nil {
			handler.log.Error(err)
			return nil, err
		}

		var mac net.HardwareAddr = arpDetails.MacAddress
		arp := &ArpEntry{
			Interface:  arpDetails.SwIfIndex,
			MacAddress: mac.String(),
			Static:     uintToBool(arpDetails.IsStatic),
		}

		var address net.IP
		if arpDetails.IsIpv6 == 1 {
			address = net.IP(arpDetails.IPAddress).To16()
		} else {
			address = net.IP(arpDetails.IPAddress[:4]).To4()
		}
		arp.IPAddress = address

		arps = append(arps, arp)
	}

	return arps, nil
}

func uintToBool(value uint8) bool {
	if value == 0 {
		return false
	}
	return true
}
