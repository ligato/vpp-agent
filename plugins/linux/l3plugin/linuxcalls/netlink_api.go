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
	// ModifyArpEntry modifies existing linux ARP entry
	ModifyArpEntry(name string, arpEntry *netlink.Neigh) error
	// DeleteArpEntry removes linux ARP entry
	DeleteArpEntry(name string, arpEntry *netlink.Neigh) error
	// ReadArpEntries returns all configured ARP entries from current namespace
	ReadArpEntries(interfaceIdx int, family int) ([]netlink.Neigh, error)
	/* Routes */
	// AddStaticRoute adds new linux static route
	AddStaticRoute(name string, route *netlink.Route) error
	// ModifyStaticRoute changes existing linux static route
	ModifyStaticRoute(name string, route *netlink.Route) error
	// DeleteStaticRoute removes linux static route
	DeleteStaticRoute(name string, route *netlink.Route) error

	// NetlinkHandlerSetup is post-init handler setup
	NetlinkHandlerSetup
}

// NetlinkHandlerSetup is post-init handler setup
type NetlinkHandlerSetup interface {
	// SetTimeLog sets time log instance to the handler
	SetStopwatch(stopwatch *measure.Stopwatch)
}

// netLinkHandler is accessor for netlink methods
type netLinkHandler struct {
	stopwatch *measure.Stopwatch
}

// SetTimeLog sets time log instance to the handler
func (handler *netLinkHandler) SetStopwatch(stopwatch *measure.Stopwatch) {
	handler.stopwatch = stopwatch
}

// NewNetLinkHandler creates new instance of netlink handler
func NewNetLinkHandler() *netLinkHandler {
	return &netLinkHandler{}
}
