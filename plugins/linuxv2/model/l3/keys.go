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
	"strings"
	"strconv"
)

const (
	/* ARP */

	// staticArpKeyPrefix is a prefix used in ETCD to store configuration for Linux static ARPs.
	staticArpKeyPrefix = "linux/config/v2/arp/"

	// staticArpKeyTemplate is a template for key representing Linux ARP entry configuration.
	staticArpKeyTemplate = staticArpKeyPrefix + "{if}/{ip}"

	/* Route Config */

	// staticRouteKeyPrefix is a prefix used in ETCD to store configuration for Linux static routes.
	staticRouteKeyPrefix = "linux/config/v2/route/"

	// staticRouteKeySuffix is a suffix common to all keys representing routes.
	staticRouteKeySuffix = "{dest-net}/{dest-mask}/{out-intf}"

	// staticRouteKeyTemplate is a template for key representing Linux Route configuration.
	staticRouteKeyTemplate = staticRouteKeyPrefix + staticRouteKeySuffix

	/* Link-local route (derived) */

	// staticLinkLocalRouteKeyPrefix is a prefix for keys derived from link-local routes.
	staticLinkLocalRouteKeyPrefix = "linux/link-local-route/"

	// staticLinkLocalRouteKeyTemplate is a template for key derived from link-local route.
	staticLinkLocalRouteKeyTemplate = staticLinkLocalRouteKeyPrefix + staticRouteKeySuffix
)

/* ARP */

// StaticArpKeyPrefix returns the prefix used in ETCD to store config for Linux static ARPs
func StaticArpKeyPrefix() string {
	return staticArpKeyPrefix
}

// StaticArpKey returns the key used in ETCD to store configuration of a particular Linux ARP entry.
func StaticArpKey(iface, ipAddr string) string {
	key := strings.Replace(staticArpKeyTemplate, "{if}", iface, 1)
	key = strings.Replace(key, "{ip}", ipAddr, 1)
	return key
}

// ParseStaticArpKey parses ARP entry from a key.
func ParseStaticArpKey(key string) (iface string, ipAddr net.IP, err error) {
	errPrefix := "invalid Linux ARP key: "
	if strings.HasPrefix(key, staticArpKeyPrefix) {
		arpSuffix := strings.TrimPrefix(key, staticArpKeyPrefix)
		arpComps := strings.Split(arpSuffix, "/")
		if len(arpComps) != 2 {
			return "", nil, fmt.Errorf(errPrefix + "invalid suffix")
		}
		ipAddr = net.ParseIP(arpComps[1])
		if ipAddr == nil {
			return "", nil, fmt.Errorf(errPrefix + "invalid IP address")
		}
		iface = arpComps[0]
		return
	}
	return "", nil, fmt.Errorf(errPrefix + "invalid prefix")
}

/* Route Config */

// StaticRouteKeyPrefix returns the prefix used in ETCD to store config for Linux static routes.
func StaticRouteKeyPrefix() string {
	return staticRouteKeyPrefix
}

// StaticRouteKey returns the key used in ETCD to store configuration of a particular Linux route.
func StaticRouteKey(dstAddr, outgoingInterface string) string {
	return staticRouteKeyFromTemplate(staticRouteKeyTemplate, dstAddr, outgoingInterface)
}

// ParseStaticRouteKey parses Linux route attributes from a key.
func ParseStaticRouteKey(key string) (dstNetAddr *net.IPNet, outgoingInterface string, err error) {
	return parseStaticRouteFromKeySuffix(key, staticRouteKeyPrefix, "invalid Linux Route key: ")
}

/* Link-local Route (derived) */

// StaticLinkLocalRouteKey returns a derived key used to represent link-local route.
func StaticLinkLocalRouteKey(dstAddr, outgoingInterface string) string {
	return staticRouteKeyFromTemplate(staticLinkLocalRouteKeyTemplate, dstAddr, outgoingInterface)
}

// ParseStaticLinkLocalRouteKey parses route attributes from a key derived from link-local route.
func ParseStaticLinkLocalRouteKey(key string) (dstNetAddr *net.IPNet, outgoingInterface string, err error) {
	return parseStaticRouteFromKeySuffix(key, staticLinkLocalRouteKeyPrefix, "invalid Linux link-local Route key: ")
}

/* Route helpers */

// staticRouteKeyFromTemplate fills key template with route attributes.
func staticRouteKeyFromTemplate(template, dstAddr, outgoingInterface string) string {
	_, dstNet, _ := net.ParseCIDR(dstAddr)
	dstNetAddr := dstNet.IP.String()
	dstNetMask, _ := dstNet.Mask.Size()
	key := strings.Replace(template, "{dest-net}", dstNetAddr, 1)
	key = strings.Replace(key, "{dest-mask}", strconv.Itoa(dstNetMask), 1)
	key = strings.Replace(key, "{out-intf}", outgoingInterface, 1)
	return key
}

// parseStaticRouteFromKeySuffix parses destination network and outgoing interface from a route key suffix.
func parseStaticRouteFromKeySuffix(key, prefix, errPrefix string) (dstNetAddr *net.IPNet, outgoingInterface string, err error) {
	if strings.HasPrefix(key, prefix) {
		routeSuffix := strings.TrimPrefix(key, prefix)
		routeComps := strings.Split(routeSuffix, "/")
		if len(routeComps) != 3 {
			return nil, "", fmt.Errorf(errPrefix + "invalid suffix")
		}
		_, dstNetAddr, err = net.ParseCIDR(routeComps[0] + "/" + routeComps[1])
		if err != nil {
			return nil, "", fmt.Errorf(errPrefix + "invalid destination address")
		}
		outgoingInterface = routeComps[2]
		return
	}
	return nil, "", fmt.Errorf(errPrefix + "invalid prefix")
}

