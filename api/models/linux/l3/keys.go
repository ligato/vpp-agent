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

package linux_l3

import (
	"net"
	"strconv"
	"strings"

	"github.com/ligato/vpp-agent/api/models"
)

func init() {
	models.RegisterProto(&StaticARPEntry{}, models.Spec{
		Module:       "linux",
		Type:         "arp",
		Version:      "v2",
		NameTemplate: "{{.Interface}}/{{.IpAddress}}",
	})
	models.RegisterProto(&StaticRoute{}, models.Spec{
		Module:       "linux",
		Type:         "route",
		Version:      "v2",
		NameTemplate: `{{with ipnet .DstNetwork}}{{printf "%s/%d" .IP .MaskSize}}{{end}}/{{.OutgoingInterface}}`,
	})
}

// StaticArpKey returns the key used in ETCD to store configuration of a particular Linux ARP entry.
func StaticArpKey(iface, ipAddr string) string {
	return models.Key(&StaticARPEntry{
		Interface: iface,
		IpAddress: ipAddr,
	})
}

// StaticRouteKey returns the key used in ETCD to store configuration of a particular Linux route.
func StaticRouteKey(dstNetwork, outgoingInterface string) string {
	return models.Key(&StaticRoute{
		DstNetwork:        dstNetwork,
		OutgoingInterface: outgoingInterface,
	})
}

const (
	/* Link-local route (derived) */

	// StaticLinkLocalRouteKeyPrefix is a prefix for keys derived from link-local routes.
	StaticLinkLocalRouteKeyPrefix = "linux/link-local-route/"

	// staticLinkLocalRouteKeyTemplate is a template for key derived from link-local route.
	staticLinkLocalRouteKeyTemplate = StaticLinkLocalRouteKeyPrefix + "{dest-net}/{dest-mask}/{out-intf}"
)

/* Link-local Route (derived) */

// StaticLinkLocalRouteKey returns a derived key used to represent link-local route.
func StaticLinkLocalRouteKey(dstAddr, outgoingInterface string) string {
	return staticRouteKeyFromTemplate(staticLinkLocalRouteKeyTemplate, dstAddr, outgoingInterface)
}

// ParseStaticLinkLocalRouteKey parses route attributes from a key derived from link-local route.
func ParseStaticLinkLocalRouteKey(key string) (dstNetAddr *net.IPNet, outgoingInterface string, isRouteKey bool) {
	return parseStaticRouteFromKeySuffix(key, StaticLinkLocalRouteKeyPrefix, "invalid Linux link-local Route key: ")
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
func parseStaticRouteFromKeySuffix(key, prefix, errPrefix string) (dstNetAddr *net.IPNet, outgoingInterface string, isRouteKey bool) {
	var err error
	if strings.HasPrefix(key, prefix) {
		routeSuffix := strings.TrimPrefix(key, prefix)
		routeComps := strings.Split(routeSuffix, "/")
		if len(routeComps) != 3 {
			return nil, "", false
		}
		_, dstNetAddr, err = net.ParseCIDR(routeComps[0] + "/" + routeComps[1])
		if err != nil {
			return nil, "", false
		}
		outgoingInterface = routeComps[2]
		isRouteKey = true
		return
	}
	return nil, "", false
}
