// Copyright (c) 2019 Cisco and/or its affiliates.
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

package netalloc

import (
	"net"
	"strings"

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

const (
	// ModuleName is the module name used for models of the netalloc plugin.
	ModuleName = "netalloc"

	// AllocRefPrefix is a prefix added in front of references to allocated objects.
	AllocRefPrefix = "alloc:"

	// AllocRefGWSuffix is a suffix added at the back of the reference when address
	// of the default gateway is requested (instead of interface IP address).
	AllocRefGWSuffix = "/GW"
)

var (
	ModelIPAllocation = models.Register(&IPAllocation{}, models.Spec{
		Module:  ModuleName,
		Version: "v1",
		Type:    "ip",
	}, models.WithNameTemplate(
		"network/{{.NetworkName}}/interface/{{.InterfaceName}}",
	))
)

const (
	/* neighbour gateway (derived) */

	// neighGwKeyTemplate is a template for keys derived from IP allocations
	// where GW is a neighbour of the interface (addresses are from the same
	// IP network).
	neighGwKeyTemplate = "netalloc/neigh-gw/network/{network}/interface/{iface}"
)

// NeighGwKey returns a derived key used to represent IP allocation where
// GW is a neighbour of the interface (addresses are from the same IP network).
func NeighGwKey(network, iface string) string {
	key := strings.Replace(neighGwKeyTemplate, "{network}", network, 1)
	key = strings.Replace(key, "{iface}", iface, 1)
	return key
}

// IPAllocMetadata stores allocated IP address already parsed from string.
type IPAllocMetadata struct {
	IfaceAddr *net.IPNet
	GwAddr    *net.IPNet
}
