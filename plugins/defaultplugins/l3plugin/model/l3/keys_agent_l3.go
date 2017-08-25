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
	RoutesPrefix = "vpp/config/v1/vrf/{vrf}/fib/{net}/{mask}/{next-hop}"
)

// VrfKeyPrefix returns the prefix used in ETCD to store VRFs for vpp instance
func VrfKeyPrefix() string {
	return VrfPrefix
}

// RouteKey returns the key used in ETCD to store vpp route for vpp instance
func RouteKey(vrf uint32, dstAddr *net.IPNet, nextHopAddr string) string {
	dstNetAddr := dstAddr.IP.String()
	dstNetMask, _ := dstAddr.Mask.Size()
	key := strings.Replace(RoutesPrefix, "{vrf}", strconv.Itoa(int(vrf)), 1)
	key = strings.Replace(key, "{net}", dstNetAddr, 1)
	key = strings.Replace(key, "{mask}", strconv.Itoa(dstNetMask), 1)
	key = strings.Replace(key, "{next-hop}", nextHopAddr, 1)
	return key
}

// ParseRouteKey parses VRF label and route address from a route key.
func ParseRouteKey(key string) (isRouteKey bool, vrfIndex string, dstNetAddr string, dstNetMask int, nextHopAddr string) {
	if strings.HasPrefix(key, VrfKeyPrefix()) {
		vrfSuffix := strings.TrimPrefix(key, VrfKeyPrefix())
		routeComps := strings.Split(vrfSuffix, "/")
		if len(routeComps) >= 5 && routeComps[1] == "fib" {
			if mask, err := strconv.Atoi(routeComps[3]); err == nil {
				return true, routeComps[0], routeComps[2], mask, routeComps[4]
			}
		}
	}
	return false, "", "", 0, ""
}

// RouteKeyPrefix returns the prefix used in ETCD to store vpp routes for vpp instance
func RouteKeyPrefix() string {
	return RoutesPrefix
}
