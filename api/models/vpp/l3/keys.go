//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp_l3

import (
	"strings"

	"github.com/ligato/vpp-agent/api/models"
)

const (
	ProxyARP_GlobalID    = "GLOBAL"
	IPScanNeigh_GlobalID = "GLOBAL"
)

func init() {
	models.Register(&ARPEntry{}, models.Spec{
		Module:  "vpp",
		Class:   "config",
		Version: "v2",
		Kind:    "arp",
		TmplID:  "{{.Interface}}/{{.IpAddress}}",
	})
	models.Register(&StaticRoute{}, models.Spec{
		Module:  "vpp",
		Class:   "config",
		Version: "v2",
		Kind:    "route",
		TmplID:  `vrf/{{.VrfId}}/dst/{{with ipnet .DstNetwork}}{{printf "%s/%d" .IP .MaskSize}}{{end}}/gw/{{.NextHopAddr}}`,
	})
	models.Register(&ProxyARP{}, models.Spec{
		Module:  "vpp",
		Class:   "config",
		Version: "v2",
		Kind:    "proxyarp",
		TmplID:  ProxyARP_GlobalID,
	})
	models.Register(&IPScanNeighbor{}, models.Spec{
		Module:  "vpp",
		Class:   "config",
		Version: "v2",
		Kind:    "ipneigh",
		TmplID:  IPScanNeigh_GlobalID,
	})
}

// RouteKey returns the key used in ETCD to store vpp route for vpp instance.
func RouteKey(vrf uint32, dstNet string, nextHopAddr string) string {
	return models.Key(&StaticRoute{
		VrfId:       vrf,
		DstNetwork:  dstNet,
		NextHopAddr: nextHopAddr,
	})
}

// ArpEntryKey returns the key to store ARP entry
func ArpEntryKey(iface, ipAddr string) string {
	return models.Key(&ARPEntry{
		Interface: iface,
		IpAddress: ipAddr,
	})
}

const (
	proxyARPInterfacePrefix   = "vpp/proxyarp/interface/"
	proxyARPInterfaceTemplate = proxyARPInterfacePrefix + "{iface}"
)

const (
	// InvalidKeyPart is used in key for parts which are invalid
	InvalidKeyPart = "<invalid>"
)

// RouteKey returns the key used in ETCD to store vpp route for vpp instance.
/*func RouteKey(vrf uint32, dstNet string, nextHopAddr string) string {
	var key = routeKeyTemplate

	key = strings.Replace(key, "{vrf}", strconv.Itoa(int(vrf)), 1)

	var dstIP string
	var dstMask string
	_, dstIPNet, err := net.ParseCIDR(dstNet)
	if err == nil {
		dstIP = dstIPNet.IP.String()
		maskSize, _ := dstIPNet.Mask.Size()
		dstMask = strconv.Itoa(maskSize)
	} else {
		dstIP = InvalidKeyPart
		dstMask = InvalidKeyPart
	}
	key = strings.Replace(key, "{dst-ip}", dstIP, 1)
	key = strings.Replace(key, "{dst-mask}", dstMask, 1)

	if nextHopAddr == "" && dstIPNet != nil {
		if dstIPNet.IP.To4() == nil {
			nextHopAddr = net.IPv6zero.String()
		} else {
			nextHopAddr = net.IPv4zero.String()
		}
	} else if net.ParseIP(nextHopAddr) == nil {
		nextHopAddr = InvalidKeyPart
	}
	key = strings.Replace(key, "{next-hop}", nextHopAddr, 1)

	return key
}

// ParseRouteKey parses VRF label and route address from a route key.
func ParseRouteKey(key string) (vrfIndex string, dstNetAddr string, dstNetMask int, nextHopAddr string, isRouteKey bool) {
	if routeKey := strings.TrimPrefix(key, RoutePrefix); routeKey != key {
		keyParts := strings.Split(routeKey, "/")
		if len(keyParts) >= 7 &&
			keyParts[0] == "vrf" &&
			keyParts[2] == "dst" &&
			keyParts[5] == "gw" {
			if mask, err := strconv.Atoi(keyParts[4]); err == nil {
				return keyParts[1], keyParts[3], mask, keyParts[6], true
			}
		}
	}
	return "", "", 0, "", false
}*/

// ArpEntryKey returns the key to store ARP entry
/*func ArpEntryKey(iface, ipAddr string) string {
	key := arpKeyTemplate
	if iface == "" {
		iface = InvalidKeyPart
	}
	if net.ParseIP(ipAddr) == nil {
		ipAddr = InvalidKeyPart
	}
	key = strings.Replace(key, "{if}", iface, 1)
	key = strings.Replace(key, "{ip}", ipAddr, 1)
	return key
}

// ParseArpKey parses ARP entry from a key
func ParseArpKey(key string) (iface string, ipAddr string, isArpKey bool) {
	if arpSuffix := strings.TrimPrefix(key, ArpPrefix); arpSuffix != key {
		arpComps := strings.Split(arpSuffix, "/")
		if len(arpComps) == 2 {
			return arpComps[0], arpComps[1], true
		}
	}
	return "", "", false
}*/

// ProxyARPInterfaceKey returns the key used to represent binding for interface with enabled proxy ARP.
func ProxyARPInterfaceKey(iface string) string {
	key := proxyARPInterfaceTemplate
	key = strings.Replace(key, "{iface}", iface, 1)
	return key
}

// ParseProxyARPInterfaceKey parses key representing binding for interface with enabled proxy ARP.
func ParseProxyARPInterfaceKey(key string) (iface string, isProxyARPInterfaceKey bool) {
	suffix := strings.TrimPrefix(key, proxyARPInterfacePrefix)
	if suffix != key && suffix != "" {
		return suffix, true
	}
	return "", false
}
