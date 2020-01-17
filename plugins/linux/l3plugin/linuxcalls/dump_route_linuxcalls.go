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

// +build !windows,!darwin

package linuxcalls

import (
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
)

const (
	// IP addresses matching any destination.
	IPv4AddrAny = "0.0.0.0"
	IPv6AddrAny = "::"

	// minimum number of interfaces to be given to a single Go routine for processing
	// in the Retrieve operation
	minWorkForGoRoutine = 3
)

// retrievedRoutes is used as the return value sent via channel by retrieveRoutes().
type retrievedRoutes struct {
	routes []*RouteDetails
	err    error
}

// GetRoutes reads all configured static routes with the given outgoing
// interface.
// <interfaceIdx> works as filter, if set to zero, all routes in the namespace
// are returned.
func (h *NetLinkHandler) GetRoutes(interfaceIdx int) (v4Routes, v6Routes []netlink.Route, err error) {
	var link netlink.Link
	if interfaceIdx != 0 {
		// netlink.RouteList reads only link index
		link = &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Index: interfaceIdx}}
	}

	v4Routes, err = netlink.RouteList(link, netlink.FAMILY_V4)
	if err != nil {
		return
	}
	v6Routes, err = netlink.RouteList(link, netlink.FAMILY_V6)
	return
}

// DumpRoutes reads all route entries and returns them as details
// with proto-modeled route data and additional metadata
func (h *NetLinkHandler) DumpRoutes() ([]*RouteDetails, error) {
	interfaces := h.ifIndexes.ListAllInterfaces()
	goRoutinesCnt := len(interfaces) / minWorkForGoRoutine
	if goRoutinesCnt == 0 {
		goRoutinesCnt = 1
	}
	if goRoutinesCnt > h.goRoutineCount {
		goRoutinesCnt = h.goRoutineCount
	}
	ch := make(chan retrievedRoutes, goRoutinesCnt)

	// invoke multiple go routines for more efficient parallel route retrieval
	for idx := 0; idx < goRoutinesCnt; idx++ {
		if goRoutinesCnt > 1 {
			go h.retrieveRoutes(interfaces, idx, goRoutinesCnt, ch)
		} else {
			h.retrieveRoutes(interfaces, idx, goRoutinesCnt, ch)
		}
	}

	// collect results from the go routines
	var routeDetails []*RouteDetails
	for idx := 0; idx < goRoutinesCnt; idx++ {
		retrieved := <-ch
		if retrieved.err != nil {
			return nil, retrieved.err
		}
		// correlate with the expected configuration
		routeDetails = append(routeDetails, retrieved.routes...)
	}

	return routeDetails, nil
}

// retrieveRoutes is run by a separate go routine to retrieve all routes entries
// associated with every <goRoutineIdx>-th interface.
func (h *NetLinkHandler) retrieveRoutes(interfaces []string, goRoutineIdx, goRoutinesCnt int, ch chan<- retrievedRoutes) {
	var retrieved retrievedRoutes
	nsCtx := linuxcalls.NewNamespaceMgmtCtx()

	for i := goRoutineIdx; i < len(interfaces); i += goRoutinesCnt {
		ifName := interfaces[i]
		// get interface metadata
		ifMeta, found := h.ifIndexes.LookupByName(ifName)
		if !found || ifMeta == nil {
			retrieved.err = errors.Errorf("failed to obtain metadata for interface %s", ifName)
			h.log.Error(retrieved.err)
			break
		}

		// switch to the namespace of the interface
		revertNs, err := h.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
		if err != nil {
			// namespace and all the routes it had contained no longer exist
			h.log.WithFields(logging.Fields{
				"err":       err,
				"namespace": ifMeta.Namespace,
			}).Warn("Failed to retrieve routes from the namespace")
			continue
		}

		// get routes assigned to this interface
		v4Routes, v6Routes, err := h.GetRoutes(ifMeta.LinuxIfIndex)
		revertNs()
		if err != nil {
			retrieved.err = err
			h.log.Error(retrieved.err)
			break
		}

		// convert each route from Netlink representation to the NB representation
		for idx, route := range append(v4Routes, v6Routes...) {
			var dstNet, gwAddr string
			if route.Dst == nil {
				if idx < len(v4Routes) {
					dstNet = IPv4AddrAny + "/0"
				} else {
					dstNet = IPv6AddrAny + "/0"
				}
			} else {
				if route.Dst.IP.To4() == nil && route.Dst.IP.IsLinkLocalUnicast() {
					// skip link-local IPv6 destinations until there is a requirement to support them
					continue
				}
				dstNet = route.Dst.String()
			}
			if len(route.Gw) != 0 {
				gwAddr = route.Gw.String()
			}
			retrieved.routes = append(retrieved.routes, &RouteDetails{
				Route: &linux_l3.Route{
					OutgoingInterface: ifName,
					DstNetwork:        dstNet,
					GwAddr:            gwAddr,
					Metric:            uint32(route.Priority),
				},
				Meta: &RouteMeta{
					InterfaceIndex: uint32(route.LinkIndex),
					NetlinkScope:   route.Scope,
					Protocol:       uint32(route.Protocol),
					MTU:            uint32(route.MTU),
				},
			})
		}
	}

	ch <- retrieved
}
