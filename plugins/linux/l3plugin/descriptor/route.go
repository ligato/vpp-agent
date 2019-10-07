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
	"bytes"
	"net"
	"strings"

	"github.com/gogo/protobuf/proto"
	prototypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/addrs"

	ifmodel "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/l3"
	netalloc_api "github.com/ligato/vpp-agent/api/models/netalloc"
	"github.com/ligato/vpp-agent/pkg/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin"
	ifdescriptor "github.com/ligato/vpp-agent/plugins/linux/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/linux/l3plugin/descriptor/adapter"
	l3linuxcalls "github.com/ligato/vpp-agent/plugins/linux/l3plugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linux/nsplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/netalloc"
	netalloc_descr "github.com/ligato/vpp-agent/plugins/netalloc/descriptor"
)

const (
	// RouteDescriptorName is the name of the descriptor for Linux routes.
	RouteDescriptorName = "linux-route"

	// IP addresses matching any destination.
	ipv4AddrAny = "0.0.0.0"
	ipv6AddrAny = "::"

	// dependency labels
	routeOutInterfaceDep       = "outgoing-interface-is-up"
	routeOutInterfaceIPAddrDep = "outgoing-interface-has-ip-address"
	routeGwReachabilityDep     = "gw-reachable"
	allocatedAddrAttached      = "allocated-addr-attached"
)

// A list of non-retriable errors:
var (
	// ErrRouteWithoutInterface is returned when Linux Route configuration is missing
	// outgoing interface reference.
	ErrRouteWithoutInterface = errors.New("Linux Route defined without outgoing interface reference")

	// ErrRouteWithUndefinedScope is returned when Linux Route is configured without scope.
	ErrRouteWithUndefinedScope = errors.New("Linux Route defined without scope")

	// ErrRouteLinkWithGw is returned when link-local Linux route has gateway address
	// specified - it shouldn't be since destination is already neighbour by definition.
	ErrRouteLinkWithGw = errors.New("Link-local Linux Route was defined with non-empty GW address")
)

// RouteDescriptor teaches KVScheduler how to configure Linux routes.
type RouteDescriptor struct {
	log       logging.Logger
	l3Handler l3linuxcalls.NetlinkAPI
	ifPlugin  ifplugin.API
	nsPlugin  nsplugin.API
	addrAlloc netalloc.AddressAllocator
	scheduler kvs.KVScheduler

	// parallelization of the Retrieve operation
	goRoutinesCnt int
}

// NewRouteDescriptor creates a new instance of the Route descriptor.
func NewRouteDescriptor(
	scheduler kvs.KVScheduler, ifPlugin ifplugin.API, nsPlugin nsplugin.API, addrAlloc netalloc.AddressAllocator,
	l3Handler l3linuxcalls.NetlinkAPI, log logging.PluginLogger, goRoutinesCnt int) *kvs.KVDescriptor {

	ctx := &RouteDescriptor{
		scheduler:     scheduler,
		l3Handler:     l3Handler,
		ifPlugin:      ifPlugin,
		nsPlugin:      nsPlugin,
		addrAlloc:     addrAlloc,
		goRoutinesCnt: goRoutinesCnt,
		log:           log.NewLogger("route-descriptor"),
	}
	typedDescr := &adapter.RouteDescriptor{
		Name:            RouteDescriptorName,
		NBKeyPrefix:     linux_l3.ModelRoute.KeyPrefix(),
		ValueTypeName:   linux_l3.ModelRoute.ProtoName(),
		KeySelector:     linux_l3.ModelRoute.IsKeyValid,
		KeyLabel:        linux_l3.ModelRoute.StripKeyPrefix,
		ValueComparator: ctx.EquivalentRoutes,
		Validate:        ctx.Validate,
		Create:          ctx.Create,
		Delete:          ctx.Delete,
		Update:          ctx.Update,
		Retrieve:        ctx.Retrieve,
		DerivedValues:   ctx.DerivedValues,
		Dependencies:    ctx.Dependencies,
		RetrieveDependencies: []string{
			netalloc_descr.IPAllocDescriptorName,
			ifdescriptor.InterfaceDescriptorName},
	}
	return adapter.NewRouteDescriptor(typedDescr)
}

// EquivalentRoutes is case-insensitive comparison function for l3.LinuxRoute.
func (d *RouteDescriptor) EquivalentRoutes(key string, oldRoute, newRoute *linux_l3.Route) bool {
	// attributes compared as usually:
	if oldRoute.OutgoingInterface != newRoute.OutgoingInterface ||
		oldRoute.Scope != newRoute.Scope ||
		oldRoute.Metric != newRoute.Metric {
		return false
	}

	// compare IP addresses converted to net.IP(Net)
	if !equalNetworks(oldRoute.DstNetwork, newRoute.DstNetwork) {
		return false
	}
	return equalAddrs(getGwAddr(oldRoute), getGwAddr(newRoute))
}

// Validate validates static route configuration.
func (d *RouteDescriptor) Validate(key string, route *linux_l3.Route) (err error) {
	if route.OutgoingInterface == "" {
		return kvs.NewInvalidValueError(ErrRouteWithoutInterface, "outgoing_interface")
	}
	if route.Scope == linux_l3.Route_LINK && route.GwAddr != "" {
		return kvs.NewInvalidValueError(ErrRouteLinkWithGw, "scope", "gw_addr")
	}
	err = d.addrAlloc.ValidateIPAddress(route.DstNetwork, "", "dst_network",
		netalloc.GWRefAllowed)
	if err != nil {
		return err
	}
	return d.addrAlloc.ValidateIPAddress(getGwAddr(route), route.OutgoingInterface,
		"gw_addr", netalloc.GWRefRequired)
}

// Create adds Linux route.
func (d *RouteDescriptor) Create(key string, route *linux_l3.Route) (metadata interface{}, err error) {
	err = d.updateRoute(route, "add", d.l3Handler.AddRoute)
	return nil, err
}

// Delete removes Linux route.
func (d *RouteDescriptor) Delete(key string, route *linux_l3.Route, metadata interface{}) error {
	return d.updateRoute(route, "delete", d.l3Handler.DelRoute)
}

// Update is able to change route scope, metric and GW address.
func (d *RouteDescriptor) Update(key string, oldRoute, newRoute *linux_l3.Route, oldMetadata interface{}) (newMetadata interface{}, err error) {
	err = d.updateRoute(newRoute, "modify", d.l3Handler.ReplaceRoute)
	return nil, err
}

// updateRoute adds, modifies or deletes a Linux route.
func (d *RouteDescriptor) updateRoute(route *linux_l3.Route, actionName string, actionClb func(route *netlink.Route) error) error {
	var err error

	// Prepare Netlink Route object
	netlinkRoute := &netlink.Route{}

	// Get interface metadata
	ifMeta, found := d.ifPlugin.GetInterfaceIndex().LookupByName(route.OutgoingInterface)
	if !found || ifMeta == nil {
		err = errors.Errorf("failed to obtain metadata for interface %s", route.OutgoingInterface)
		d.log.Error(err)
		return err
	}

	// set link index
	netlinkRoute.LinkIndex = ifMeta.LinuxIfIndex

	// set destination network
	dstNet, err := d.addrAlloc.GetOrParseIPAddress(route.DstNetwork, "",
		netalloc_api.IPAddressForm_ADDR_NET)
	if err != nil {
		d.log.Error(err)
		return err
	}
	netlinkRoute.Dst = dstNet

	// set gateway address
	if route.GwAddr != "" {
		gwAddr, err := d.addrAlloc.GetOrParseIPAddress(route.GwAddr, route.OutgoingInterface,
			netalloc_api.IPAddressForm_ADDR_ONLY)
		if err != nil {
			d.log.Error(err)
			return err
		}
		netlinkRoute.Gw = gwAddr.IP
	}

	// set route scope
	scope, err := rtScopeFromNBToNetlink(route.Scope)
	if err != nil {
		d.log.Error(err)
		return err
	}
	netlinkRoute.Scope = scope

	// set route metric
	netlinkRoute.Priority = int(route.Metric)

	// move to the namespace of the associated interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revertNs, err := d.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
	if err != nil {
		err = errors.Errorf("failed to switch namespace: %v", err)
		d.log.Error(err)
		return err
	}
	defer revertNs()

	// update route in the interface namespace
	err = actionClb(netlinkRoute)
	if err != nil {
		err = errors.Errorf("failed to %s linux route: %v", actionName, err)
		d.log.Error(err)
		return err
	}

	return nil
}

// Dependencies lists dependencies for a Linux route.
func (d *RouteDescriptor) Dependencies(key string, route *linux_l3.Route) []kvs.Dependency {
	var dependencies []kvs.Dependency
	// the outgoing interface must exist and be UP
	if route.OutgoingInterface != "" {
		dependencies = append(dependencies, kvs.Dependency{
			Label: routeOutInterfaceDep,
			Key:   ifmodel.InterfaceStateKey(route.OutgoingInterface, true),
		})
	}
	// if destination network is netalloc reference, then the address must be allocated first
	allocDep, hasAllocDep := d.addrAlloc.GetAddressAllocDep(route.DstNetwork, "",
		"dst_network-")
	if hasAllocDep {
		dependencies = append(dependencies, allocDep)
	}
	// if GW is netalloc reference, then the address must be allocated first
	allocDep, hasAllocDep = d.addrAlloc.GetAddressAllocDep(route.GwAddr, route.OutgoingInterface,
		"gw_addr-")
	if hasAllocDep {
		dependencies = append(dependencies, allocDep)
	}
	// GW must be routable
	network, iface, _, isRef, _ := d.addrAlloc.ParseAddressAllocRef(route.GwAddr, route.OutgoingInterface)
	if isRef {
		// GW is netalloc reference
		dependencies = append(dependencies, kvs.Dependency{
			Label: routeGwReachabilityDep,
			AnyOf: kvs.AnyOfDependency{
				KeyPrefixes: []string{
					netalloc_api.NeighGwKey(network, iface),
					linux_l3.StaticLinkLocalRouteKey(
						d.addrAlloc.CreateAddressAllocRef(network, iface, true),
						route.OutgoingInterface),
				},
			},
		})
		dependencies = append(dependencies, kvs.Dependency{
			Label: allocatedAddrAttached,
			Key: ifmodel.InterfaceAddressKey(
				route.OutgoingInterface, d.addrAlloc.CreateAddressAllocRef(network, "", false),
				netalloc_api.IPAddressSource_ALLOC_REF),
		})
	} else if gwAddr := net.ParseIP(getGwAddr(route)); gwAddr != nil && !gwAddr.IsUnspecified() {
		// GW is not netalloc reference but an actual IP
		dependencies = append(dependencies, kvs.Dependency{
			Label: routeGwReachabilityDep,
			AnyOf: kvs.AnyOfDependency{
				KeyPrefixes: []string{
					ifmodel.InterfaceAddressPrefix(route.OutgoingInterface),
					linux_l3.StaticLinkLocalRoutePrefix(route.OutgoingInterface),
				},
				KeySelector: func(key string) bool {
					dstAddr, ifName, isRouteKey := linux_l3.ParseStaticLinkLocalRouteKey(key)
					if isRouteKey && ifName == route.OutgoingInterface {
						if _, dstNet, err := net.ParseCIDR(dstAddr); err == nil && dstNet.Contains(gwAddr) {
							// GW address is neighbour as told by another link-local route
							return true
						}
						return false
					}
					ifName, address, source, _, isAddrKey := ifmodel.ParseInterfaceAddressKey(key)
					if isAddrKey && source != netalloc_api.IPAddressSource_ALLOC_REF {
						if _, network, err := net.ParseCIDR(address); err == nil && network.Contains(gwAddr) {
							// GW address is inside the local network of the outgoing interface
							// as given by the assigned IP address
							return true
						}
					}
					return false
				},
			},
		})
	} else if route.OutgoingInterface != "" {
		// route also requires the interface to be in the L3 mode (have at least one
		// IP address assigned) - we set this only for routes without GW and other
		// routes will inherit this dependency transitively through GW-reachability dep.
		dependencies = append(dependencies, kvs.Dependency{
			Label: routeOutInterfaceIPAddrDep,
			AnyOf: kvs.AnyOfDependency{
				KeyPrefixes: []string{
					ifmodel.InterfaceAddressPrefix(route.OutgoingInterface),
				},
			},
		})
	}
	return dependencies
}

// DerivedValues derives empty value under StaticLinkLocalRouteKey if route is link-local.
// It is used in dependencies for network reachability of a route gateway (see above).
func (d *RouteDescriptor) DerivedValues(key string, route *linux_l3.Route) (derValues []kvs.KeyValuePair) {
	if route.Scope == linux_l3.Route_LINK {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   linux_l3.StaticLinkLocalRouteKey(route.DstNetwork, route.OutgoingInterface),
			Value: &prototypes.Empty{},
		})
	}
	return derValues
}

// retrievedRoutes is used as the return value sent via channel by retrieveRoutes().
type retrievedRoutes struct {
	routes []adapter.RouteKVWithMetadata
	err    error
}

// Retrieve returns all routes associated with interfaces managed by this agent.
func (d *RouteDescriptor) Retrieve(correlate []adapter.RouteKVWithMetadata) ([]adapter.RouteKVWithMetadata, error) {
	var values []adapter.RouteKVWithMetadata

	// prepare expected configuration with de-referenced netalloc links
	nbCfg := make(map[string]*linux_l3.Route)
	expCfg := make(map[string]*linux_l3.Route)
	for _, kv := range correlate {
		dstNetwork := kv.Value.DstNetwork
		parsed, err := d.addrAlloc.GetOrParseIPAddress(kv.Value.DstNetwork,
			"", netalloc_api.IPAddressForm_ADDR_NET)
		if err == nil {
			dstNetwork = parsed.String()
		}
		gwAddr := kv.Value.GwAddr
		parsed, err = d.addrAlloc.GetOrParseIPAddress(getGwAddr(kv.Value),
			kv.Value.OutgoingInterface, netalloc_api.IPAddressForm_ADDR_ONLY)
		if err == nil {
			gwAddr = parsed.IP.String()
		}
		route := proto.Clone(kv.Value).(*linux_l3.Route)
		route.DstNetwork = dstNetwork
		route.GwAddr = gwAddr
		key := models.Key(route)
		expCfg[key] = route
		nbCfg[key] = kv.Value
	}

	interfaces := d.ifPlugin.GetInterfaceIndex().ListAllInterfaces()
	goRoutinesCnt := len(interfaces) / minWorkForGoRoutine
	if goRoutinesCnt == 0 {
		goRoutinesCnt = 1
	}
	if goRoutinesCnt > d.goRoutinesCnt {
		goRoutinesCnt = d.goRoutinesCnt
	}
	ch := make(chan retrievedRoutes, goRoutinesCnt)

	// invoke multiple go routines for more efficient parallel route retrieval
	for idx := 0; idx < goRoutinesCnt; idx++ {
		if goRoutinesCnt > 1 {
			go d.retrieveRoutes(interfaces, idx, goRoutinesCnt, ch)
		} else {
			d.retrieveRoutes(interfaces, idx, goRoutinesCnt, ch)
		}
	}

	// collect results from the go routines
	for idx := 0; idx < goRoutinesCnt; idx++ {
		retrieved := <-ch
		if retrieved.err != nil {
			return values, retrieved.err
		}
		// correlate with the expected configuration
		for _, route := range retrieved.routes {
			key := linux_l3.RouteKey(route.Value.DstNetwork, route.Value.OutgoingInterface)
			if expCfg, hasExpCfg := expCfg[key]; hasExpCfg {
				if d.EquivalentRoutes(key, route.Value, expCfg) {
					route.Value = nbCfg[key]
					// recreate the key in case the dest. IP was replaced with netalloc link
					route.Key = models.Key(route.Value)
				}
			}
			values = append(values, route)
		}
	}

	return values, nil
}

// retrieveRoutes is run by a separate go routine to retrieve all routes entries
// associated with every <goRoutineIdx>-th interface.
func (d *RouteDescriptor) retrieveRoutes(interfaces []string, goRoutineIdx, goRoutinesCnt int, ch chan<- retrievedRoutes) {
	var retrieved retrievedRoutes
	ifMetaIdx := d.ifPlugin.GetInterfaceIndex()
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()

	for i := goRoutineIdx; i < len(interfaces); i += goRoutinesCnt {
		ifName := interfaces[i]
		// get interface metadata
		ifMeta, found := ifMetaIdx.LookupByName(ifName)
		if !found || ifMeta == nil {
			retrieved.err = errors.Errorf("failed to obtain metadata for interface %s", ifName)
			d.log.Error(retrieved.err)
			break
		}

		// switch to the namespace of the interface
		revertNs, err := d.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
		if err != nil {
			// namespace and all the routes it had contained no longer exist
			d.log.WithFields(logging.Fields{
				"err":       err,
				"namespace": ifMeta.Namespace,
			}).Warn("Failed to retrieve routes from the namespace")
			continue
		}

		// get routes assigned to this interface
		v4Routes, v6Routes, err := d.l3Handler.GetRoutes(ifMeta.LinuxIfIndex)
		revertNs()
		if err != nil {
			retrieved.err = err
			d.log.Error(retrieved.err)
			break
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
			retrieved.routes = append(retrieved.routes, adapter.RouteKVWithMetadata{
				Key: linux_l3.RouteKey(dstNet, ifName),
				Value: &linux_l3.Route{
					OutgoingInterface: ifName,
					Scope:             scope,
					DstNetwork:        dstNet,
					GwAddr:            gwAddr,
					Metric:            uint32(route.Priority),
				},
				Origin: kvs.UnknownOrigin, // let the scheduler to determine the origin
			})
		}
	}

	ch <- retrieved
}

// rtScopeFromNBToNetlink convert Route scope from NB configuration
// to the corresponding Netlink constant.
func rtScopeFromNBToNetlink(scope linux_l3.Route_Scope) (netlink.Scope, error) {
	switch scope {
	case linux_l3.Route_GLOBAL:
		return netlink.SCOPE_UNIVERSE, nil
	case linux_l3.Route_HOST:
		return netlink.SCOPE_HOST, nil
	case linux_l3.Route_LINK:
		return netlink.SCOPE_LINK, nil
	case linux_l3.Route_SITE:
		return netlink.SCOPE_SITE, nil
	}
	return 0, ErrRouteWithUndefinedScope
}

// rtScopeFromNetlinkToNB converts Route scope from Netlink constant
// to the corresponding NB constant.
func rtScopeFromNetlinkToNB(scope netlink.Scope) (linux_l3.Route_Scope, error) {
	switch scope {
	case netlink.SCOPE_UNIVERSE:
		return linux_l3.Route_GLOBAL, nil
	case netlink.SCOPE_HOST:
		return linux_l3.Route_HOST, nil
	case netlink.SCOPE_LINK:
		return linux_l3.Route_LINK, nil
	case netlink.SCOPE_SITE:
		return linux_l3.Route_SITE, nil
	}
	return 0, ErrRouteWithUndefinedScope
}

// equalAddrs compares two IP addresses for equality.
func equalAddrs(addr1, addr2 string) bool {
	if strings.HasPrefix(addr1, netalloc_api.AllocRefPrefix) {
		return addr1 == addr2
	}
	a1 := net.ParseIP(addr1)
	a2 := net.ParseIP(addr2)
	if a1 == nil || a2 == nil {
		// if parsing fails, compare as strings
		return strings.ToLower(addr1) == strings.ToLower(addr2)
	}
	return a1.Equal(a2)
}

// equalNetworks compares two IP networks for equality.
func equalNetworks(net1, net2 string) bool {
	if strings.HasPrefix(net1, netalloc_api.AllocRefPrefix) {
		return net1 == net2
	}
	_, n1, err1 := net.ParseCIDR(net1)
	_, n2, err2 := net.ParseCIDR(net2)
	if err1 != nil || err2 != nil {
		// if parsing fails, compare as strings
		return strings.ToLower(net1) == strings.ToLower(net2)
	}
	return n1.IP.Equal(n2.IP) && bytes.Equal(n1.Mask, n2.Mask)
}

// getGwAddr returns the GW address chosen in the given route, handling the cases
// when it is left undefined.
func getGwAddr(route *linux_l3.Route) string {
	if route.GwAddr == "" {
		if ipv6, _ := addrs.IsIPv6(route.DstNetwork); ipv6 {
			return ipv6AddrAny
		}
		return ipv4AddrAny
	}
	return route.GwAddr
}
