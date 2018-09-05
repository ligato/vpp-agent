// Copyright (c) 2018 Cisco and/or its affiliates.
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

package descriptor

import (
	"strings"

	scheduler "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/kvscheduler/value/protoval"

	"github.com/ligato/vpp-agent/plugins/linuxv2/model/l3"
	"net"
	"bytes"
	"github.com/ligato/cn-infra/utils/addrs"
)

// RouteProtoValue overrides the default implementation of the Equivalent method.
type RouteProtoValue struct {
	protoval.ProtoValue
	route *l3.LinuxStaticRoute
}

// Equivalent is case-insensitive comparison function for l3.LinuxStaticRoute.
func (rpv *RouteProtoValue) Equivalent(v2 scheduler.Value) bool {
	rpv2, ok := v2.(*RouteProtoValue)
	if !ok {
		return false
	}
	route1 := rpv.route
	route2 := rpv2.route

	// attributes compared as usually:
	if route1.OutgoingInterface != route2.OutgoingInterface ||
		route1.Scope != route2.Scope ||
		route1.Metric != route2.Metric {
		return false
	}

	// compare IP addresses converted to net.IP(Net)
	if !equalNetworks(route1.DstNetwork, route2.DstNetwork) {
		return false
	}
	return equalAddrs(getGwAddr(route1), getGwAddr(route2))
}

// equalAddrs compares two IP addresses for equality.
func equalAddrs(addr1, addr2 string) bool {
	a1 := net.ParseIP(addr1)
	a2 := net.ParseIP(addr2)
	if a1 == nil || a2 == nil {
		// if parsing fails, compare as strings
		return strings.ToLower(addr1) == strings.ToLower(addr2)
	}
	return a1.Equal(a2)
}

// equalNetworks compares two IP networks for equality.
func equalNetworks(net1, net2 string) bool {
	_, n1, err1 := net.ParseCIDR(net1)
	_, n2, err2 := net.ParseCIDR(net2)
	if err1 != nil || err2 != nil {
		// if parsing fails, compare as strings
		return strings.ToLower(net1) == strings.ToLower(net2)
	}
	return n1.IP.Equal(n2.IP) && bytes.Equal(n1.Mask, n2.Mask)
}

// getGwAddr returns the GW address chosen in the given route, handling the cases
// when it is left undefined.
func getGwAddr(route *l3.LinuxStaticRoute) string {
	if route.GwAddr == "" {
		if ipv6, _ := addrs.IsIPv6(route.DstNetwork); ipv6 {
			return ipv6AddrAny
		}
		return ipv4AddrAny
	}
	return route.GwAddr
}