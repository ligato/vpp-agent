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

package vppdump

import (
	"fmt"
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	l3ba "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/bin_api/ip"
	l3nb "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
)

// StaticRoutes is the wrapper structure for the static routes API structure.
type StaticRoutes struct {
	IP []*StaticRouteIP
}

// StaticRouteIP is the wrapper structure for the static IP route API structure.
// NOTE: NextHops in StaticRoutes_Ip is overridden by the local NextHops member.
type StaticRouteIP struct {
	NextHops []*NextHop
	l3nb.StaticRoutes_Route
}

// NextHop is the wrapper structure for the bridge domain interface northbound API structure.
type NextHop struct {
	OutgoingInterfaceSwIfIdx    uint32
	OutgoingInterfaceConfigured bool
	l3nb.StaticRoutes_Route_NextHops
}

// DumpStaticRoutes dumps l3 routes from VPP and fills them into the provided static route map.
func DumpStaticRoutes(log logging.Logger, vppChan *govppapi.Channel) (map[uint32]*StaticRoutes, error) {

	// map for the resulting l3 FIBs
	routes := make(map[uint32]*StaticRoutes)

	// dump IPv4 l3 FIB
	reqCtx := vppChan.SendMultiRequest(&l3ba.IPFibDump{})
	for {
		fibDetails := &l3ba.IPFibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			log.Error(err)
			return nil, err
		}
		dumpStaticRouteDetails(routes, fibDetails.TableID, fibDetails.Address, fibDetails.AddressLength, fibDetails.Path, true)
	}

	// dump IPv6 l3 FIB
	reqCtx = vppChan.SendMultiRequest(&l3ba.IP6FibDump{})
	for {
		fibDetails := &l3ba.IP6FibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			log.Error(err)
			return nil, err
		}
		dumpStaticRouteDetails(routes, fibDetails.TableID, fibDetails.Address, fibDetails.AddressLength, fibDetails.Path, true)
	}

	return routes, nil
}

// dumpStaticRouteDetails processes static route details and fills them into the provided routes map.
func dumpStaticRouteDetails(routes map[uint32]*StaticRoutes, tableID uint32,
	address []byte, prefixLen uint8, paths []l3ba.FibPath, ipv6 bool) {

	// route details
	var ipAddr string
	if ipv6 {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(address).To16().String(), uint32(prefixLen))
	} else {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(address[:4]).To4().String(), uint32(prefixLen))
	}
	if _, ok := routes[tableID]; !ok {
		routes[tableID] = &StaticRoutes{
			IP: make([]*StaticRouteIP, 0),
		}
	}
	route := &StaticRouteIP{
		StaticRoutes_Route: l3nb.StaticRoutes_Route{
			VrfId:              tableID,
			DestinationAddress: ipAddr,
		},
		NextHops: []*NextHop{},
	}
	routes[tableID].IP = append(routes[tableID].IP, route)

	// next hops
	for _, path := range paths {
		var nextHopAddr string
		if ipv6 {
			nextHopAddr = net.IP(path.NextHop).To16().String()
		} else {
			nextHopAddr = net.IP(path.NextHop[:4]).To4().String()
		}
		route.NextHops = append(route.NextHops, &NextHop{
			OutgoingInterfaceSwIfIdx:    path.SwIfIndex,
			OutgoingInterfaceConfigured: path.SwIfIndex < ^uint32(0),
			StaticRoutes_Route_NextHops: l3nb.StaticRoutes_Route_NextHops{
				Address:    nextHopAddr,
				Weight:     path.Weight,
				Preference: uint32(path.Preference),
			},
		})
	}
}
