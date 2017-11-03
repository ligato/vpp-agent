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
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	l3ba "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppcalls"
	"time"
)

// DumpStaticRoutes dumps l3 routes from VPP and fills them into the provided static route map.
func DumpStaticRoutes(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) ([]*vppcalls.Route, error) {
	// IPFibDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	var routes []*vppcalls.Route

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

		ipv4Route, err := dumpStaticRouteIPv4Details(fibDetails)
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipv4Route)
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
		ipv6Route, err := dumpStaticRouteIPv6Details(fibDetails)
		if err != nil {
			return nil, err
		}
		routes = append(routes, ipv6Route)
	}

	return routes, nil
}

func dumpStaticRouteIPv4Details(fibDetails *l3ba.IPFibDetails) (*vppcalls.Route, error) {
	return dumpStaticRouteIPDetails(fibDetails.TableID, fibDetails.Address, fibDetails.AddressLength, fibDetails.Path, false)

}

func dumpStaticRouteIPv6Details(fibDetails *l3ba.IP6FibDetails) (*vppcalls.Route, error) {
	return dumpStaticRouteIPDetails(fibDetails.TableID, fibDetails.Address, fibDetails.AddressLength, fibDetails.Path, true)
}

// dumpStaticRouteIPDetails processes static route details and returns a route object
func dumpStaticRouteIPDetails(tableID uint32, address []byte, prefixLen uint8, path []l3ba.FibPath, ipv6 bool) (*vppcalls.Route, error) {
	// route details
	var ipAddr string
	if ipv6 {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(address).To16().String(), uint32(prefixLen))
	} else {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(address[:4]).To4().String(), uint32(prefixLen))
	}

	rt := &vppcalls.Route{}

	// IP net
	parsedIP, _, err := addrs.ParseIPWithPrefix(ipAddr)
	if err != nil {
		return nil, err
	}

	rt.VrfID = tableID
	rt.DstAddr = *parsedIP

	if len(path) > 0 {
		var nextHopAddr net.IP
		if ipv6 {
			nextHopAddr = net.IP(path[0].NextHop).To16()
		} else {
			nextHopAddr = net.IP(path[0].NextHop[:4]).To4()
		}

		rt.NextHopAddr = nextHopAddr
		rt.OutIface = path[0].SwIfIndex
		rt.Preference = uint32(path[0].Preference)
		rt.Weight = uint32(path[0].Weight)
	}

	return rt, nil
}
