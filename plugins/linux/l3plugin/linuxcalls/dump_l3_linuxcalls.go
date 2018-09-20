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

package linuxcalls

import (
	"github.com/ligato/vpp-agent/plugins/linux/model/l3"
)

// LinuxArpDetails is the wrapper structure for the linux ARP northbound API structure.
type LinuxArpDetails struct {
	Interface *l3.LinuxStaticArpEntries_ArpEntry `json:"linux_arp"`
	Meta      *LinuxArpMeta                      `json:"linux_arp_meta"`
}

// LinuxArpMeta is combination of proto-modelled ARP data and linux provided metadata
type LinuxArpMeta struct {
}

// DumpArpEntries is an implementation of linux L3 handler
func (h *NetLinkHandler) DumpArpEntries() ([]*LinuxArpDetails, error) {
	var arps []*LinuxArpDetails

	// todo implement

	return arps, nil
}

// LinuxRouteDetails is the wrapper structure for the linux route northbound API structure.
type LinuxRouteDetails struct {
	Interface *l3.LinuxStaticRoutes_Route `json:"linux_route"`
	Meta      *LinuxArpMeta               `json:"linux_route_meta"`
}

// LinuxRouteMeta is combination of proto-modelled route data and linux provided metadata
type LinuxRouteMeta struct {
}

// DumpRoutes is an implementation of linux route handler
func (h *NetLinkHandler) DumpRoutes() ([]*LinuxRouteDetails, error) {
	var routes []*LinuxRouteDetails

	// todo implement

	return routes, nil
}
