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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
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
		bytes.Equal(a.NextHopAddr, b.NextHopAddr) &&
		a.OutIface == b.OutIface &&
		a.Weight == b.Weight &&
		a.Preference == b.Preference
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
	if !bytes.Equal(a.NextHopAddr, b.NextHopAddr) {
		return bytes.Compare(a.NextHopAddr, b.NextHopAddr) < 0
	}
	if a.OutIface != b.OutIface {
		return a.OutIface < b.OutIface
	}
	if a.Preference != b.Preference {
		return a.Preference < b.Preference
	}
	return a.Weight < b.Weight

}

// TransformRoute converts raw route data to Route object
func TransformRoute(routeInput *l3.StaticRoutes_Route, index ifaceidx.SwIfIndex, log logging.Logger) (*vppcalls.Route, error) {
	if routeInput == nil {
		log.Infof("Route input is empty")
		return nil, nil
	}
	if routeInput.DstIpAddr == "" {
		log.Infof("Route does not contain destination address")
		return nil, nil
	}
	parsedDestIP, isIpv6, err := addrs.ParseIPWithPrefix(routeInput.DstIpAddr)
	if err != nil {
		return nil, err
	}
	vrfID := routeInput.VrfId

	ifName := routeInput.OutgoingInterface

	ifIndex := vppcalls.NextHopOutgoingIfUnset
	if ifName != "" {
		var exists bool
		ifIndex, _, exists = index.LookupIdx(ifName)
		if !exists {
			return nil, fmt.Errorf("route outgoing interface %v not found", ifName)
		}
	}

	nextHopIP := net.ParseIP(routeInput.NextHopAddr)
	if isIpv6 {
		nextHopIP = nextHopIP.To16()
	} else {
		nextHopIP = nextHopIP.To4()
	}
	route := &vppcalls.Route{
		VrfID:       vrfID,
		DstAddr:     *parsedDestIP,
		NextHopAddr: nextHopIP,
		OutIface:    ifIndex,
		Weight:      routeInput.Weight,
		Preference:  routeInput.Preference,
	}
	return route, nil
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
