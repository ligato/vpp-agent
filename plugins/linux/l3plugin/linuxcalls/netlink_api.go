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
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/vishvananda/netlink"
)

// NetlinkAPI interface covers all methods inside linux calls package needed to manage linux ARP entries and routes.
type NetlinkAPI interface {
	/* ARP */
	// AddArpEntry configures new linux ARP entry
	AddArpEntry(name string, arpEntry *netlink.Neigh) error
	// SetArpEntry modifies existing linux ARP entry
	SetArpEntry(name string, arpEntry *netlink.Neigh) error
	// DelArpEntry removes linux ARP entry
	DelArpEntry(name string, arpEntry *netlink.Neigh) error
	// GetArpEntries returns all configured ARP entries from current namespace
	GetArpEntries(interfaceIdx int, family int) ([]netlink.Neigh, error)
	/* Routes */
	// AddStaticRoute adds new linux static route
	AddStaticRoute(name string, route *netlink.Route) error
	// ReplaceStaticRoute changes existing linux static route
	ReplaceStaticRoute(name string, route *netlink.Route) error
	// DelStaticRoute removes linux static route
	DelStaticRoute(name string, route *netlink.Route) error
}

// netLinkHandler is accessor for netlink methods
type netLinkHandler struct {
	stopwatch *measure.Stopwatch
}

// NewNetLinkHandler creates new instance of netlink handler
func NewNetLinkHandler(stopwatch *measure.Stopwatch) *netLinkHandler {
	return &netLinkHandler{
		stopwatch: stopwatch,
	}
}
