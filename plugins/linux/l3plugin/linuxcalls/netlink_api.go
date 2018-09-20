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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linux/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	"github.com/vishvananda/netlink"
)

// NetlinkAPI interface covers all methods inside linux calls package needed to manage linux ARP entries and routes.
type NetlinkAPI interface {
	NetlinkAPIWrite
	NetlinkAPIRead
}

// NetlinkAPIWrite interface covers write methods inside linux calls package needed to manage linux ARP entries and routes.
type NetlinkAPIWrite interface {
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
	// GetStaticRoutes reads linux routes. Possible to filter by interface and IP family.
	GetStaticRoutes(link netlink.Link, family int) ([]netlink.Route, error)
}

// NetlinkAPIRead interface covers read methods inside linux calls package needed to manage linux ARP entries and routes.
type NetlinkAPIRead interface {
	// DumpArpEntries returns configured linux ARPs
	DumpArpEntries() ([]*LinuxArpDetails, error)
	// DumpRoutes returns configured linux Routes
	DumpRoutes() ([]*LinuxRouteDetails, error)
}

// NetLinkHandler is accessor for netlink methods
type NetLinkHandler struct {
	nsHandler    nsplugin.NamespaceAPI
	ifIndexes    ifaceidx.LinuxIfIndex
	arpIndexes   l3idx.LinuxARPIndex
	routeIndexes l3idx.LinuxRouteIndex
	stopwatch    *measure.Stopwatch
	log          logging.Logger
}

// NewNetLinkHandler creates new instance of netlink handler
func NewNetLinkHandler(nsHandler nsplugin.NamespaceAPI, ifIndexes ifaceidx.LinuxIfIndex, arpIndexes l3idx.LinuxARPIndex, routeIndexes l3idx.LinuxRouteIndex,
	log logging.Logger, stopwatch *measure.Stopwatch) *NetLinkHandler {
	return &NetLinkHandler{
		nsHandler:    nsHandler,
		ifIndexes:    ifIndexes,
		arpIndexes:   arpIndexes,
		routeIndexes: routeIndexes,
		stopwatch:    stopwatch,
		log:          log,
	}
}
