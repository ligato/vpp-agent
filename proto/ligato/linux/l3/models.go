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
	"strings"

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

// ModuleName is the module name used for models.
const ModuleName = "linux.l3"

var (
	ModelARPEntry = models.Register(&ARPEntry{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "arp",
	}, models.WithNameTemplate("{{.Interface}}/{{.IpAddress}}"))

	ModelRoute = models.Register(&Route{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "route",
	}, models.WithNameTemplate(
		`{{with ipnet .DstNetwork}}{{printf "%s/%d" .IP .MaskSize}}`+
			`{{else}}{{.DstNetwork}}{{end}}/{{.OutgoingInterface}}`,
	))
)

// ArpKey returns the key used in ETCD to store configuration of a particular Linux ARP entry.
func ArpKey(iface, ipAddr string) string {
	return models.Key(&ARPEntry{
		Interface: iface,
		IpAddress: ipAddr,
	})
}

// RouteKey returns the key used in ETCD to store configuration of a particular Linux route.
func RouteKey(dstNetwork, outgoingInterface string) string {
	return models.Key(&Route{
		DstNetwork:        dstNetwork,
		OutgoingInterface: outgoingInterface,
	})
}

const (
	/* Link-local route (derived) */

	// StaticLinkLocalRouteKeyPrefix is a prefix for keys derived from link-local routes.
	LinkLocalRouteKeyPrefix = "linux/link-local-route/"

	// staticLinkLocalRouteKeyTemplate is a template for key derived from link-local route.
	linkLocalRouteKeyTemplate = LinkLocalRouteKeyPrefix + "{out-iface}/dest-address/{dest-address}"
)

/* Link-local Route (derived) */

// StaticLinkLocalRouteKey returns a derived key used to represent link-local route.
func StaticLinkLocalRouteKey(dstAddr, outgoingInterface string) string {
	key := strings.Replace(linkLocalRouteKeyTemplate, "{dest-address}", dstAddr, 1)
	key = strings.Replace(key, "{out-iface}", outgoingInterface, 1)
	return key
}

// StaticLinkLocalRoutePrefix returns longest-common prefix of keys representing
// link-local routes that have the given outgoing Linux interface.
func StaticLinkLocalRoutePrefix(outgoingInterface string) string {
	return LinkLocalRouteKeyPrefix + outgoingInterface + "/"
}

// ParseStaticLinkLocalRouteKey parses route attributes from a key derived from link-local route.
func ParseStaticLinkLocalRouteKey(key string) (dstAddr string, outgoingInterface string, isRouteKey bool) {
	if strings.HasPrefix(key, LinkLocalRouteKeyPrefix) {
		routeSuffix := strings.TrimPrefix(key, LinkLocalRouteKeyPrefix)
		parts := strings.Split(routeSuffix, "/dest-address/")

		if len(parts) != 2 {
			return "", "", false
		}
		outgoingInterface = parts[0]
		dstAddr = parts[1]
		isRouteKey = true
		return
	}
	return "", "", false
}
