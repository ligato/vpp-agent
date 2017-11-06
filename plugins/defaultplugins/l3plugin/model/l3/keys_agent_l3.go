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
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Prefixes
const (
	// VrfPrefix is the relative key prefix for VRFs.
	VrfPrefix = "vpp/config/v1/vrf/"
	// TablePrefix is the relative key prefix for tables
	TablePrefix = VrfPrefix + "{vrf}"
	// RoutesPrefix is the relative key prefix for routes.
	RoutesPrefix = VrfPrefix + "{vrf}/fib/{net}/{mask}/{next-hop}"
	// ArpPrefix is the relative key prefix for ARP.
	ArpPrefix = "vpp/config/v1/arp/"
	// ArpEntryPrefix is the relative key prefix for ARP table entries.
	ArpKey = ArpPrefix + "{if}/{ip}"
	// ProxyARPPrefix is the relative key prefix for proxy ARP configuration.
	ProxyARPPrefix = "vpp/config/v1/proxyarp/"
	// ProxyARPRangePrefix is the relative key prefix for proxy ARP ranges.
	ProxyARPRangePrefix = ProxyARPPrefix + "range/{lo_ip}/{hi_ip}"
	// ProxyARPInterfacePrefix is the relative key prefix for proxy ARP-enabled interfaces.
	ProxyARPInterfacePrefix = ProxyARPPrefix + "interface/{if}"
	// STNPrefix is the relative key prefix for STN entries.
	STNPrefix = "vpp/config/v1/stn/{ip}"
)

// VrfKeyPrefix returns the prefix used in ETCD to store VRFs for vpp instance
func VrfKeyPrefix() string {
	return VrfPrefix
}

// RouteKeyPrefix returns the prefix used in ETCD to store vpp routes for vpp instance
func RouteKeyPrefix() string {
	return RoutesPrefix
}

// TableKey returns the key used in ETCD to store vpp table for vpp instance
func TableKey(vrf uint32) string {
	key := TablePrefix
	key = strings.Replace(key, "{vrf}", strconv.Itoa(int(vrf)), 1)
	return key
}

// ArpKeyPrefix returns the prefix used in ETCD to store vpp APR tables for vpp instance
func ArpKeyPrefix() string {
	return ArpPrefix
}

// RouteKey returns the key used in ETCD to store vpp route for vpp instance
func RouteKey(vrf uint32, dstAddr *net.IPNet, nextHopAddr string) string {
	dstNetAddr := dstAddr.IP.String()
	dstNetMask, _ := dstAddr.Mask.Size()
	key := RoutesPrefix
	key = strings.Replace(key, "{vrf}", strconv.Itoa(int(vrf)), 1)
	key = strings.Replace(key, "{net}", dstNetAddr, 1)
	key = strings.Replace(key, "{mask}", strconv.Itoa(dstNetMask), 1)
	key = strings.Replace(key, "{next-hop}", nextHopAddr, 1)
	return key
}

// ParseVrfKey parses VRF index and route address from given key
func ParseVrfKey(key string) (isRouteKey bool, vrfIndex string, dstNetAddr string, dstNetMask int, nextHopAddr string, err error) {
	if !strings.HasPrefix(key, VrfPrefix) {
		err = fmt.Errorf("wrong prefix in key: %q", key)
		return
	}

	keySuffix := strings.TrimPrefix(key, VrfPrefix)
	routeComps := strings.Split(keySuffix, "/")

	vrfIndex = routeComps[0]
	if vrfIndex == "" {
		err = fmt.Errorf("invalid VRF index in key: %q", key)
		return
	}

	if len(routeComps) == 1 {
		isRouteKey = false
		return
	} else if len(routeComps) == 5 && routeComps[1] == "fib" {
		isRouteKey = true
		dstNetAddr = routeComps[2]
		nextHopAddr = routeComps[4]
		if dstNetMask, err = strconv.Atoi(routeComps[3]); err != nil {
			err = fmt.Errorf("invalid format in key: %q", key)
			return
		}
	} else {
		err = fmt.Errorf("invalid format in key: %q", key)
		return
	}
	return isRouteKey, vrfIndex, dstNetAddr, dstNetMask, nextHopAddr, nil
}

// ArpEntryKey returns the key to store ARP entry
func ArpEntryKey(iface, ipAddr string) string {
	key := ArpKey
	key = strings.Replace(key, "{if}", iface, 1)
	key = strings.Replace(key, "{ip}", ipAddr, 1)
	//key = strings.Replace(key, "{mac}", macAddr, 1)
	return key
}

// ParseArpKey parses ARP entry from a key
func ParseArpKey(key string) (iface string, ipAddr string, err error) {
	if strings.HasPrefix(key, ArpPrefix) {
		arpSuffix := strings.TrimPrefix(key, ArpPrefix)
		arpComps := strings.Split(arpSuffix, "/")
		if len(arpComps) == 2 {
			return arpComps[0], arpComps[1], nil
		}
	}
	return "", "", fmt.Errorf("invalid ARP key")
}
