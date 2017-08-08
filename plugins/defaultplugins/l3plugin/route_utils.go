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

package l3plugin

import (
	"bytes"
	"fmt"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppcalls"
	"net"
	"sort"
)

// SortedRoutes type is used to implement sort interface for slice of Route
type SortedRoutes []*vppcalls.Route

// Returns length of slice
// Implements sort.Interface
func (arr SortedRoutes) Len() int {
	return len(arr)
}

// Swap swaps two items in slice identified by indexes
// Implements sort.Interface
func (arr SortedRoutes) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

// Less returns true if the item in slice at index i in slice
// should be sorted before the element with index j
// Implements sort.Interface
func (arr SortedRoutes) Less(i, j int) bool {
	return lessRoute(arr[i], arr[j])
}

func eqRoutes(a *vppcalls.Route, b *vppcalls.Route) bool {
	return a.VrfID == b.VrfID &&
		bytes.Equal(a.DstAddr.IP, b.DstAddr.IP) &&
		bytes.Equal(a.DstAddr.Mask, b.DstAddr.Mask) &&
		bytes.Equal(a.NextHop.Addr, b.NextHop.Addr) &&
		a.NextHop.Iface == b.NextHop.Iface &&
		a.NextHop.Weight == b.NextHop.Weight
}

func lessRoute(a *vppcalls.Route, b *vppcalls.Route) bool {
	if a.VrfID != b.VrfID {
		return a.VrfID < b.VrfID
	}
	if !bytes.Equal(a.DstAddr.IP, b.DstAddr.IP) {
		return bytes.Compare(a.DstAddr.IP, b.DstAddr.IP) < 0
	}
	if !bytes.Equal(a.DstAddr.Mask, b.DstAddr.Mask) {
		return bytes.Compare(a.DstAddr.Mask, b.DstAddr.Mask) < 0
	}
	if !bytes.Equal(a.NextHop.Addr, b.NextHop.Addr) {
		return bytes.Compare(a.NextHop.Addr, b.NextHop.Addr) < 0
	}
	if a.NextHop.Iface != b.NextHop.Iface {
		return a.NextHop.Iface < b.NextHop.Iface
	}
	return a.NextHop.Weight < b.NextHop.Weight

}

// Transform raw routes data to list of Route objects.
func (plugin *RouteConfigurator) transformRoute(routeInput *l3.StaticRoutes_Route) ([]*vppcalls.Route, error) {
	var routes []*vppcalls.Route
	if routeInput != nil {
		var (
			ifIndex uint32
			exists  bool
		)
		if routeInput.DestinationAddress == "" {
			return nil, fmt.Errorf("Route does not contain destination address")
		}
		parsedDestIP, isIpv6, err := addrs.ParseIPWithPrefix(routeInput.DestinationAddress)
		if err != nil {
			return nil, err
		}
		vrfID := routeInput.VrfId

		for _, nextHop := range routeInput.NextHops {

			ifName := nextHop.OutgoingInterface
			if ifName == "" {
				log.Infof("Outgoing interface not set for next hop %v, route skipped", nextHop.Address)
				continue
			}
			ifIndex, _, exists = plugin.SwIfIndexes.LookupIdx(ifName)
			if !exists {
				log.Infof("Interface %v not found, route skipped", ifName)
			}
			if !exists {
				ifIndex = vppcalls.NextHopOutgoingIfUnset
			}
			nextHopIP := net.ParseIP(nextHop.Address)
			if isIpv6 {
				nextHopIP = nextHopIP.To16()
			} else {
				nextHopIP = nextHopIP.To4()
			}
			route := &vppcalls.Route{
				VrfID:     vrfID,
				DstAddr:   *parsedDestIP,
				MultiPath: routeInput.Multipath,
				NextHop: vppcalls.NextHopList{
					Addr:   nextHopIP,
					Iface:  ifIndex,
					Weight: nextHop.Weight,
				},
			}
			routes = append(routes, route)
		}
	}

	return routes, nil
}

func (plugin *RouteConfigurator) diffRoutes(new []*vppcalls.Route, old []*vppcalls.Route) (toBeDeleted []*vppcalls.Route, toBeAdded []*vppcalls.Route) {
	newSorted := SortedRoutes(new)
	oldSorted := SortedRoutes(old)
	sort.Sort(newSorted)
	sort.Sort(oldSorted)

	//compare
	i := 0
	j := 0
	for i < len(newSorted) && j < len(oldSorted) {
		if eqRoutes(newSorted[i], oldSorted[j]) {
			i++
			j++
		} else {
			if lessRoute(newSorted[i], oldSorted[j]) {
				toBeAdded = append(toBeAdded, newSorted[i])
				i++
			} else {
				toBeDeleted = append(toBeDeleted, oldSorted[j])
				j++
			}
		}
	}

	for ; i < len(newSorted); i++ {
		toBeAdded = append(toBeAdded, newSorted[i])
	}

	for ; j < len(oldSorted); j++ {
		toBeDeleted = append(toBeDeleted, oldSorted[j])
	}
	return
}
