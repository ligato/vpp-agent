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

// +build !windows,!darwin

package linuxcalls

import (
	"github.com/vishvananda/netlink"
)

// AddRoute creates the new static route
func (h *NetLinkHandler) AddRoute(route *netlink.Route) error {
	return netlink.RouteAdd(route)
}

// ReplaceRoute replaces the static route
func (h *NetLinkHandler) ReplaceRoute(route *netlink.Route) error {
	return netlink.RouteReplace(route)
}

// DelRoute removes the static route
func (h *NetLinkHandler) DelRoute(route *netlink.Route) error {
	return netlink.RouteDel(route)
}
