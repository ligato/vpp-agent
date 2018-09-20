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

package descriptor

import (
	"net"
	"strings"

	"github.com/go-errors/errors"
	"github.com/vishvananda/netlink"

	scheduler "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/cn-infra/kvscheduler/value/emptyval"
	"github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	ifdescriptor "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin/descriptor/adapter"
	l3linuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin/linuxcalls"
	ifmodel "github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linuxv2/model/l3"
	"github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin"
	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin/linuxcalls"
)

const (
	// RouteDescriptorName is the name of the descriptor for Linux routes.
	RouteDescriptorName = "linux-route"

	// IP addresses matching any destination.
	ipv4AddrAny = "0.0.0.0"
	ipv6AddrAny = "::"

	// dependency labels
	routeOutInterfaceDep   = "interface"
	routeGwReachabilityDep = "gw-reachability"
)

// A list of non-retriable errors:
var (
	// ErrRouteWithoutInterface is returned when Linux Route configuration is missing
	// outgoing interface reference.
	ErrRouteWithoutInterface = errors.New("Linux Route defined without outgoing interface reference")

	// ErrRouteWithoutDestination is returned when Linux Route configuration is missing destination network.
	ErrRouteWithoutDestination = errors.New("Linux Route defined without destination network")

	// ErrRouteWithUnsupportedScope is returned when Linux Route is configured with unrecognized scope.
	ErrRouteWithUnsupportedScope = errors.New("Linux Route defined with unsupported scope")

	// ErrRouteWithInvalidDst is returned when Linux Route configuration contains destination
	// network that cannot be parsed.
	ErrRouteWithInvalidDst = errors.New("Linux Route defined with invalid destination network")

	// ErrRouteWithInvalidGW is returned when Linux Route configuration contains gateway
	// address that cannot be parsed.
	ErrRouteWithInvalidGw = errors.New("Linux Route defined with invalid GW address")

	// ErrRouteLinkWithGw is returned when link-local Linux route has gateway address
	// specified - it shouldn't be since destination is already neighbour by definition.
	ErrRouteLinkWithGw = errors.New("Link-local Linux Route was defined with non-empty GW address")
)

// RouteDescriptor teaches KVScheduler how to configure Linux routes.
type RouteDescriptor struct {
	adapter.RouteDescriptorBase

	log       logging.Logger
	l3Handler l3linuxcalls.NetlinkAPI
	ifPlugin  ifplugin.API
	nsPlugin  nsplugin.API
	scheduler scheduler.KVScheduler
}

// NewRouteDescriptor creates a new instance of the Route descriptor.
func NewRouteDescriptor(
	scheduler scheduler.KVScheduler, ifPlugin ifplugin.API, nsPlugin nsplugin.API,
	l3Handler l3linuxcalls.NetlinkAPI, log logging.PluginLogger) *RouteDescriptor {

	return &RouteDescriptor{
		scheduler: scheduler,
		l3Handler: l3Handler,
		ifPlugin:  ifPlugin,
		nsPlugin:  nsPlugin,
		log:       log.NewLogger("-route-descriptor"),
	}
}

// GetName returns name of the descriptor for Linux Routes.
func (rd *RouteDescriptor) GetName() string {
	return RouteDescriptorName
}

// KeySelector selects values with the configuration for Linux routes.
func (rd *RouteDescriptor) KeySelector(key string) bool {
	return strings.HasPrefix(key, l3.StaticRouteKeyPrefix())
}

// NBKeyPrefixes returns NB-config key prefix for Linux routes.
func (rd *RouteDescriptor) NBKeyPrefixes() []string {
	return []string{l3.StaticRouteKeyPrefix()}
}

// Add add Linux route.
func (rd *RouteDescriptor) Add(key string, route *l3.LinuxStaticRoute) (metadata interface{}, err error) {
	err = rd.updateRoute(route, "add", rd.l3Handler.AddStaticRoute)
	return nil, err
}

// Delete removes Linux route.
func (rd *RouteDescriptor) Delete(key string, route *l3.LinuxStaticRoute, metadata interface{}) error {
	return rd.updateRoute(route, "delete", rd.l3Handler.DelStaticRoute)
}

// Modify is able to change route scope, metric and GW address.
func (rd *RouteDescriptor) Modify(key string, oldRoute, newRoute *l3.LinuxStaticRoute, oldMetadata interface{}) (newMetadata interface{}, err error) {
	err = rd.updateRoute(newRoute, "modify", rd.l3Handler.ReplaceStaticRoute)
	return nil, err
}

// updateRoute adds, modifies or deletes a Linux route.
func (rd *RouteDescriptor) updateRoute(route *l3.LinuxStaticRoute, actionName string, actionClb func(route *netlink.Route) error) error {
	var err error

	// validate the configuration first
	if route.OutgoingInterface == "" {
		err = ErrRouteWithoutInterface
		rd.log.Error(err)
		return err
	}
	if route.DstNetwork == "" {
		err = ErrRouteWithoutDestination
		rd.log.Error(err)
		return err
	}
	if route.Scope == l3.LinuxStaticRoute_LINK && route.GwAddr != "" {
		err = ErrRouteLinkWithGw
		rd.log.Error(err)
		return err
	}

	// Prepare Netlink Route object
	netlinkRoute := &netlink.Route{}

	// Get interface metadata
	ifMeta, found := rd.ifPlugin.GetInterfaceIndex().LookupByName(route.OutgoingInterface)
	if !found || ifMeta == nil {
		err = errors.Errorf("failed to obtain metadata for interface %s", route.OutgoingInterface)
		rd.log.Error(err)
		return err
	}

	// set link index
	netlinkRoute.LinkIndex = ifMeta.LinuxIfIndex

	// set destination network
	_, dstNet, err := net.ParseCIDR(route.DstNetwork)
	if err != nil {
		err = ErrRouteWithInvalidDst
		rd.log.Error(err)
		return err
	}
	netlinkRoute.Dst = dstNet

	// set gateway address
	if route.GwAddr != "" {
		gwAddr := net.ParseIP(route.GwAddr)
		if gwAddr == nil {
			err = ErrRouteWithInvalidGw
			rd.log.Error(err)
			return err
		}
		netlinkRoute.Gw = gwAddr
	}

	// set route scope
	scope, err := rtScopeFromNBToNetlink(route.Scope)
	if err != nil {
		err = ErrRouteWithUnsupportedScope
		rd.log.Error(err)
		return err
	}
	netlinkRoute.Scope = scope

	// set route metric
	netlinkRoute.Priority = int(route.Metric)

	// move to the namespace of the associated interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revertNs, err := rd.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
	if err != nil {
		err = errors.Errorf("failed to switch namespace: %v", err)
		rd.log.Error(err)
		return err
	}
	defer revertNs()

	// update route in the interface namespace
	err = actionClb(netlinkRoute)
	if err != nil {
		err = errors.Errorf("failed to %s linux route: %v", actionName, err)
		rd.log.Error(err)
		return err
	}

	return nil
}

// ModifyHasToRecreate returns true if the outgoing interfaces or destination IP has changed.
func (rd *RouteDescriptor) ModifyHasToRecreate(key string, oldRoute, newRoute *l3.LinuxStaticRoute, oldMetadata interface{}) bool {
	return oldRoute.OutgoingInterface != newRoute.OutgoingInterface || !equalNetworks(oldRoute.DstNetwork, newRoute.DstNetwork)
}

// Dependencies lists dependencies for a Linux route.
func (rd *RouteDescriptor) Dependencies(key string, route *l3.LinuxStaticRoute) []scheduler.Dependency {
	var dependencies []scheduler.Dependency
	// the outgoing interface must exist and be UP
	if route.OutgoingInterface != "" {
		dependencies = append(dependencies, scheduler.Dependency{
			Label: routeOutInterfaceDep,
			Key:   ifmodel.InterfaceStateKey(route.OutgoingInterface, true),
		})
	}
	// GW must be routable
	gwAddr := net.ParseIP(getGwAddr(route))
	if gwAddr != nil && !gwAddr.IsUnspecified() {
		dependencies = append(dependencies, scheduler.Dependency{
			Label: routeGwReachabilityDep,
			AnyOf: func(key string) bool {
				dstAddr, ifName, err := l3.ParseStaticLinkLocalRouteKey(key)
				if err == nil && ifName == route.OutgoingInterface && dstAddr.Contains(gwAddr) {
					// GW address is neighbour as told by another link-local route
					return true
				}
				ifName, addr, err := ifmodel.ParseInterfaceAddressKey(key)
				if err == nil && ifName == route.OutgoingInterface && addr.Contains(gwAddr) {
					// GW address is inside the local network of the outgoing interface
					// as given by the assigned IP address
					return true
				}
				return false
			},
		})
	}
	return dependencies
}

// DerivedValues derives empty value under StaticLinkLocalRouteKey if route is link-local.
// It is used in dependencies for network reachability of a route gateway (see above).
func (rd *RouteDescriptor) DerivedValues(key string, route *l3.LinuxStaticRoute) (derValues []scheduler.KeyValuePair) {
	if route.Scope == l3.LinuxStaticRoute_LINK {
		derValues = append(derValues, scheduler.KeyValuePair{
			Key:   l3.StaticLinkLocalRouteKey(route.DstNetwork, route.OutgoingInterface),
			Value: emptyval.NewEmptyValue(),
		})
	}
	return derValues
}

// Dump returns all routes associated with interfaces managed by this agent.
func (rd *RouteDescriptor) Dump(correlate []adapter.RouteKVWithMetadata) ([]adapter.RouteKVWithMetadata, error) {
	var err error
	var dump []adapter.RouteKVWithMetadata
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	ifMetaIdx := rd.ifPlugin.GetInterfaceIndex()

	// dump only routes with outgoing interfaces managed by this agent.
	for _, ifName := range ifMetaIdx.ListAllInterfaces() {
		// get interface metadata
		ifMeta, found := ifMetaIdx.LookupByName(ifName)
		if !found || ifMeta == nil {
			err = errors.Errorf("failed to obtain metadata for interface %s", ifName)
			rd.log.Error(err)
			return dump, err
		}

		// switch to the namespace of the interface
		revertNs, err := rd.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
		if err != nil {
			err = errors.Errorf("failed to switch namespace: %v", err)
			rd.log.Error(err)
			return dump, err
		}

		// get routes assigned to this interface
		v4Routes, v6Routes, err := rd.l3Handler.GetStaticRoutes(ifMeta.LinuxIfIndex)
		revertNs()
		if err != nil {
			rd.log.Error(err)
			return dump, err
		}

		// convert each route from Netlink representation to the NB representation
		for idx, route := range append(v4Routes, v6Routes...) {
			var dstNet, gwAddr string
			if route.Dst == nil {
				if idx < len(v4Routes) {
					dstNet = ipv4AddrAny + "/0"
				} else {
					dstNet = ipv6AddrAny + "/0"
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
			scope, err := rtScopeFromNetlinkToNB(route.Scope)
			if err != nil {
				// route not configured by the agent
				continue
			}
			dump = append(dump, adapter.RouteKVWithMetadata{
				Key: l3.StaticRouteKey(dstNet, ifName),
				Value: &l3.LinuxStaticRoute{
					OutgoingInterface: ifName,
					Scope:             scope,
					DstNetwork:        dstNet,
					GwAddr:            gwAddr,
					Metric:            uint32(route.Priority),
				},
				Origin: scheduler.UnknownOrigin, // let the scheduler to determine the origin
			})
		}
	}
	rd.log.WithField("dump", dump).Debug("Dumping Linux Routes")
	return dump, nil
}

// DumpDependencies tells scheduler to dump configured interfaces first.
func (rd *RouteDescriptor) DumpDependencies() []string {
	return []string{ifdescriptor.InterfaceDescriptorName}
}

// rtScopeFromNBToNetlink convert Route scope from NB configuration
// to the corresponding Netlink constant.
func rtScopeFromNBToNetlink(scope l3.LinuxStaticRoute_Scope) (netlink.Scope, error) {
	switch scope {
	case l3.LinuxStaticRoute_GLOBAL:
		return netlink.SCOPE_UNIVERSE, nil
	case l3.LinuxStaticRoute_HOST:
		return netlink.SCOPE_HOST, nil
	case l3.LinuxStaticRoute_LINK:
		return netlink.SCOPE_LINK, nil
	case l3.LinuxStaticRoute_SITE:
		return netlink.SCOPE_SITE, nil
	}
	return 0, ErrRouteWithUnsupportedScope
}

// rtScopeFromNetlinkToNB converts Route scope from Netlink constant
// to the corresponding NB constant.
func rtScopeFromNetlinkToNB(scope netlink.Scope) (l3.LinuxStaticRoute_Scope, error) {
	switch scope {
	case netlink.SCOPE_UNIVERSE:
		return l3.LinuxStaticRoute_GLOBAL, nil
	case netlink.SCOPE_HOST:
		return l3.LinuxStaticRoute_HOST, nil
	case netlink.SCOPE_LINK:
		return l3.LinuxStaticRoute_LINK, nil
	case netlink.SCOPE_SITE:
		return l3.LinuxStaticRoute_SITE, nil
	}
	return 0, ErrRouteWithUnsupportedScope
}
