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

package l3

import (
	"net"
	"strconv"
	"strings"
)

// Prefixes
const (
	// VrfPrefix is the relative key prefix for VRFs.
	VrfPrefix = "vpp/config/v1/vrf/"
	// RoutesPrefix is the relative key prefix for routes.
	RoutesPrefix = "vpp/config/v1/vrf/{vrf}/fib/"
)

// VrfKeyPrefix returns the prefix used in ETCD to store VRFs for vpp instance
func VrfKeyPrefix() string {
	return VrfPrefix
}

// RouteKey returns the key used in ETCD to store vpp route for vpp instance
func RouteKey(vrf uint32, dstAddr *net.IPNet, nextHopAddr string) string {
	dstNetAddr := dstAddr.IP.String()
	dstNetMask, _ := dstAddr.Mask.Size()
	identifier := dstNetAddr + "m" + strconv.Itoa(dstNetMask) + "-" + nextHopAddr
	return strings.Replace(RoutesPrefix, "{vrf}", strconv.Itoa(int(vrf)), 1) + identifier
}

// ParseRouteKey parses VRF label and route address from a route key.
func ParseRouteKey(key string) (isRouteKey bool, vrfIndex string, routeAddress string) {
	if strings.HasPrefix(key, VrfKeyPrefix()) {
		vrfSuffix := strings.TrimPrefix(key, VrfKeyPrefix())
		routeComps := strings.Split(vrfSuffix, "/")
		if len(routeComps) >= 3 && routeComps[1] == "fib" {
			return true, routeComps[0], routeComps[2]
		}
	}
	return false, "", ""
}

// RouteKeyPrefix returns the prefix used in ETCD to store vpp routes for vpp instance
func RouteKeyPrefix() string {
	return RoutesPrefix
}
