// +build !windows,!darwin

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

package linuxcalls

import (
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/vishvananda/netlink"
	"time"
)

// AddStaticRoute creates the new static route
func AddStaticRoute(name string, route *netlink.Route, log logging.Logger, timeLog measure.StopWatchEntry) error {
	log.Debugf("Creating the new static route %v", name)
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	return netlink.RouteAdd(route)
}

// ModifyStaticRoute removes the static route
func ModifyStaticRoute(name string, route *netlink.Route, log logging.Logger, timeLog measure.StopWatchEntry) error {
	log.Debugf("Updating static route %v", name)
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	return netlink.RouteReplace(route)
}

// DeleteStaticRoute removes the static route
func DeleteStaticRoute(name string, route *netlink.Route, log logging.Logger, timeLog measure.StopWatchEntry) error {
	log.Debugf("Removing static route %v", name)
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	return netlink.RouteDel(route)
}

// ReadStaticRoutes reads static routes
// // <link> and <family> parameters works as filters, if they are set to nil/zero values, all routes are returned
func ReadStaticRoutes(link netlink.Link, family int, log logging.Logger, timeLog measure.StopWatchEntry) ([]netlink.Route, error) {
	log.Debug("Reading static routes")
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	return netlink.RouteList(link, family)
}
