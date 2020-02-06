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
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	netalloc_descr "go.ligato.io/vpp-agent/v3/plugins/netalloc/descriptor"
	ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	netalloc_api "go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	// RouteDescriptorName is the name of the descriptor for static routes.
	RouteDescriptorName = "vpp-route"

	// dependency labels
	routeOutInterfaceDep = "interface-exists"
	vrfTableDep          = "vrf-table-exists"
	viaVrfTableDep       = "via-vrf-table-exists"

	// static route weight by default
	defaultWeight = 1
)

// RouteDescriptor teaches KVScheduler how to configure VPP routes.
type RouteDescriptor struct {
	log          logging.Logger
	routeHandler vppcalls.RouteVppAPI
	addrAlloc    netalloc.AddressAllocator
}

// NewRouteDescriptor creates a new instance of the Route descriptor.
func NewRouteDescriptor(
	routeHandler vppcalls.RouteVppAPI, addrAlloc netalloc.AddressAllocator,
	log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &RouteDescriptor{
		routeHandler: routeHandler,
		addrAlloc:    addrAlloc,
		log:          log.NewLogger("static-route-descriptor"),
	}

	typedDescr := &adapter.RouteDescriptor{
		Name:            RouteDescriptorName,
		NBKeyPrefix:     l3.ModelRoute.KeyPrefix(),
		ValueTypeName:   l3.ModelRoute.ProtoName(),
		KeySelector:     l3.ModelRoute.IsKeyValid,
		ValueComparator: ctx.EquivalentRoutes,
		Validate:        ctx.Validate,
		Create:          ctx.Create,
		Delete:          ctx.Delete,
		Retrieve:        ctx.Retrieve,
		Dependencies:    ctx.Dependencies,
		RetrieveDependencies: []string{
			netalloc_descr.IPAllocDescriptorName,
			ifdescriptor.InterfaceDescriptorName,
			VrfTableDescriptorName},
	}
	return adapter.NewRouteDescriptor(typedDescr)
}

// EquivalentRoutes is case-insensitive comparison function for l3.Route.
func (d *RouteDescriptor) EquivalentRoutes(key string, oldRoute, newRoute *l3.Route) bool {
	if oldRoute.GetType() != newRoute.GetType() ||
		oldRoute.GetVrfId() != newRoute.GetVrfId() ||
		oldRoute.GetViaVrfId() != newRoute.GetViaVrfId() ||
		oldRoute.GetOutgoingInterface() != newRoute.GetOutgoingInterface() ||
		getWeight(oldRoute) != getWeight(newRoute) ||
		oldRoute.GetPreference() != newRoute.GetPreference() {
		return false
	}

	// compare dst networks
	if !equalNetworks(oldRoute.DstNetwork, newRoute.DstNetwork) {
		return false
	}

	// compare gw addresses (next hop)
	if !equalAddrs(getGwAddr(oldRoute), getGwAddr(newRoute)) {
		return false
	}

	return true
}

// Validate validates VPP static route configuration.
func (d *RouteDescriptor) Validate(key string, route *l3.Route) (err error) {
	// validate destination network
	err = d.addrAlloc.ValidateIPAddress(route.DstNetwork, "", "dst_network",
		netalloc.GWRefAllowed)
	if err != nil {
		return err
	}

	// validate next hop address (GW)
	err = d.addrAlloc.ValidateIPAddress(getGwAddr(route), route.OutgoingInterface,
		"gw_addr", netalloc.GWRefRequired)
	if err != nil {
		return err
	}

	// validate IP network implied by the IP and prefix length
	if !strings.HasPrefix(route.DstNetwork, netalloc_api.AllocRefPrefix) {
		_, ipNet, _ := net.ParseCIDR(route.DstNetwork)
		if strings.ToLower(ipNet.String()) != strings.ToLower(route.DstNetwork) {
			e := fmt.Errorf("DstNetwork (%s) must represent IP network (%s)",
				route.DstNetwork, ipNet.String())
			return kvs.NewInvalidValueError(e, "dst_network")
		}
	}

	// TODO: validate mix of IP versions?

	return nil
}

// Create adds VPP static route.
func (d *RouteDescriptor) Create(key string, route *l3.Route) (metadata interface{}, err error) {
	err = d.routeHandler.VppAddRoute(context.TODO(), route)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Delete removes VPP static route.
func (d *RouteDescriptor) Delete(key string, route *l3.Route, metadata interface{}) error {
	err := d.routeHandler.VppDelRoute(context.TODO(), route)
	if err != nil {
		return err
	}

	return nil
}

// Retrieve returns all routes associated with interfaces managed by this agent.
func (d *RouteDescriptor) Retrieve(correlate []adapter.RouteKVWithMetadata) (
	retrieved []adapter.RouteKVWithMetadata, err error,
) {
	// prepare expected configuration with de-referenced netalloc links
	nbCfg := make(map[string]*l3.Route)
	expCfg := make(map[string]*l3.Route)
	for _, kv := range correlate {
		dstNetwork := kv.Value.DstNetwork
		parsed, err := d.addrAlloc.GetOrParseIPAddress(kv.Value.DstNetwork,
			"", netalloc_api.IPAddressForm_ADDR_NET)
		if err == nil {
			dstNetwork = parsed.String()
		}
		nextHop := kv.Value.NextHopAddr
		parsed, err = d.addrAlloc.GetOrParseIPAddress(getGwAddr(kv.Value),
			kv.Value.OutgoingInterface, netalloc_api.IPAddressForm_ADDR_ONLY)
		if err == nil {
			nextHop = parsed.IP.String()
		}
		route := proto.Clone(kv.Value).(*l3.Route)
		route.DstNetwork = dstNetwork
		route.NextHopAddr = nextHop
		key := models.Key(route)
		expCfg[key] = route
		nbCfg[key] = kv.Value
	}

	// Retrieve VPP route configuration
	routes, err := d.routeHandler.DumpRoutes()
	if err != nil {
		return nil, errors.Errorf("failed to dump VPP routes: %v", err)
	}

	for _, route := range routes {
		key := models.Key(route.Route)
		value := route.Route
		origin := kvs.UnknownOrigin

		// correlate with the expected configuration
		if expCfg, hasExpCfg := expCfg[key]; hasExpCfg {
			if d.EquivalentRoutes(key, value, expCfg) {
				value = nbCfg[key]
				// recreate the key in case the dest. IP or GW IP were replaced with netalloc link
				key = models.Key(value)
				origin = kvs.FromNB
			}
		}

		retrieved = append(retrieved, adapter.RouteKVWithMetadata{
			Key:    key,
			Value:  value,
			Origin: origin,
		})
	}

	return retrieved, nil
}

// Dependencies lists dependencies for a VPP route.
func (d *RouteDescriptor) Dependencies(key string, route *l3.Route) []kvs.Dependency {
	var dependencies []kvs.Dependency
	// the outgoing interface must exist and be UP
	if route.OutgoingInterface != "" {
		dependencies = append(dependencies, kvs.Dependency{
			Label: routeOutInterfaceDep,
			Key:   interfaces.InterfaceKey(route.OutgoingInterface),
		})
	}

	// non-zero VRFs
	var protocol l3.VrfTable_Protocol
	_, isIPv6, _ := addrs.ParseIPWithPrefix(route.DstNetwork)
	if isIPv6 {
		protocol = l3.VrfTable_IPV6
	}
	if route.VrfId != 0 {
		dependencies = append(dependencies, kvs.Dependency{
			Label: vrfTableDep,
			Key:   l3.VrfTableKey(route.VrfId, protocol),
		})
	}
	if route.Type == l3.Route_INTER_VRF && route.ViaVrfId != 0 {
		dependencies = append(dependencies, kvs.Dependency{
			Label: viaVrfTableDep,
			Key:   l3.VrfTableKey(route.ViaVrfId, protocol),
		})
	}

	// if destination network is netalloc reference, then the address must be allocated first
	allocDep, hasAllocDep := d.addrAlloc.GetAddressAllocDep(route.DstNetwork,
		"", "dst_network-")
	if hasAllocDep {
		dependencies = append(dependencies, allocDep)
	}
	// if GW is netalloc reference, then the address must be allocated first
	allocDep, hasAllocDep = d.addrAlloc.GetAddressAllocDep(route.NextHopAddr,
		route.OutgoingInterface, "gw_addr-")
	if hasAllocDep {
		dependencies = append(dependencies, allocDep)
	}

	// TODO: perhaps check GW routability
	return dependencies
}

// equalAddrs compares two IP addresses for equality.
func equalAddrs(addr1, addr2 string) bool {
	if strings.HasPrefix(addr1, netalloc_api.AllocRefPrefix) ||
		strings.HasPrefix(addr1, netalloc_api.AllocRefPrefix) {
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

// getGwAddr returns the GW address chosen in the given route, handling the cases
// when it is left undefined.
func getGwAddr(route *l3.Route) string {
	if route.GetNextHopAddr() != "" {
		return route.GetNextHopAddr()
	}
	// return zero address
	// - with netalloc'd destination network, just assume it is for IPv4
	if !strings.HasPrefix(route.GetDstNetwork(), netalloc_api.AllocRefPrefix) {
		_, dstIPNet, err := net.ParseCIDR(route.GetDstNetwork())
		if err != nil {
			return ""
		}
		if dstIPNet.IP.To4() == nil {
			return net.IPv6zero.String()
		}
	}
	return net.IPv4zero.String()
}

// getWeight returns static route weight, handling the cases when it is left undefined.
func getWeight(route *l3.Route) uint32 {
	if route.Weight == 0 {
		return defaultWeight
	}
	return route.Weight
}

// equalNetworks compares two IP networks for equality.
func equalNetworks(net1, net2 string) bool {
	if strings.HasPrefix(net1, netalloc_api.AllocRefPrefix) ||
		strings.HasPrefix(net2, netalloc_api.AllocRefPrefix) {
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
