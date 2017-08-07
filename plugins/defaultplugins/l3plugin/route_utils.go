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
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"net"
	"sort"
	"fmt"
)

// Route represents a forward IP route entry.
type Route struct {
	vrfID    uint32
	destAddr net.IPNet
	nexthop  NextHop
}

// NextHop defines the parameters of gateway to which packets should be forwarded
// when a given routing table entry is applied.
type NextHop struct {
	addr   net.IP
	intf   uint32
	multipath bool
	weight uint32
}

// SortedRoutes type is used to implement sort interface for slice of Route
type SortedRoutes []*Route

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

func eqRoutes(a *Route, b *Route) bool {
	return a.vrfID == b.vrfID &&
		bytes.Equal(a.destAddr.IP, b.destAddr.IP) &&
		bytes.Equal(a.destAddr.Mask, b.destAddr.Mask) &&
		bytes.Equal(a.nexthop.addr, b.nexthop.addr) &&
		a.nexthop.intf == b.nexthop.intf &&
		a.nexthop.weight == b.nexthop.weight
}

func lessRoute(a *Route, b *Route) bool {
	if a.vrfID != b.vrfID {
		return a.vrfID < b.vrfID
	}
	if !bytes.Equal(a.destAddr.IP, b.destAddr.IP) {
		return bytes.Compare(a.destAddr.IP, b.destAddr.IP) < 0
	}
	if !bytes.Equal(a.destAddr.Mask, b.destAddr.Mask) {
		return bytes.Compare(a.destAddr.Mask, b.destAddr.Mask) < 0
	}
	if !bytes.Equal(a.nexthop.addr, b.nexthop.addr) {
		return bytes.Compare(a.nexthop.addr, b.nexthop.addr) < 0
	}
	if a.nexthop.intf != b.nexthop.intf {
		return a.nexthop.intf < b.nexthop.intf
	}
	return a.nexthop.weight < b.nexthop.weight

}

func (plugin *RouteConfigurator) transformRoute(routeInput *l3.StaticRoutes_Route) (*Route, error) {
	var route *Route
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
		ifName := routeInput.OutgoingInterface
		ifIndex, _, exists = plugin.SwIfIndexes.LookupIdx(ifName)
		if ifName != "" && !exists {
			return nil, fmt.Errorf("Interface %v not found, route skipped", ifName)
		}
		if !exists {
			ifIndex = nextHopOutgoingIfUnset
		}
		nextHopIP := net.ParseIP(routeInput.NextHopAddress)
		if isIpv6 {
			nextHopIP = nextHopIP.To16()
		} else {
			nextHopIP = nextHopIP.To4()
		}
		route = &Route{
			vrfID:    vrfID,
			destAddr: *parsedDestIP,
			nexthop: NextHop{
				addr:   nextHopIP,
				intf:   ifIndex,
				multipath: routeInput.Multipath,
				weight: routeInput.Weight,
			},
		}
	}

	return route, nil
}

func (plugin *RouteConfigurator) diffRoutes(new []*Route, old []*Route) (toBeDeleted []*Route, toBeAdded []*Route) {
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
