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

package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"

	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
)

// L3VppAPI groups L3 Vpp APIs.
type L3VppAPI interface {
	ArpVppAPI
	ProxyArpVppAPI
	RouteVppAPI
	IPNeighVppAPI
	VrfTableVppAPI
}

// ArpDetails holds info about ARP entry as a proto model
type ArpDetails struct {
	Arp  *l3.ARPEntry
	Meta *ArpMeta
}

// ArpMeta contains interface index of the ARP interface
type ArpMeta struct {
	SwIfIndex uint32
}

// ArpVppAPI provides methods for managing ARP entries
type ArpVppAPI interface {
	ArpVppRead

	// VppAddArp adds ARP entry according to provided input
	VppAddArp(entry *l3.ARPEntry) error
	// VppDelArp removes old ARP entry according to provided input
	VppDelArp(entry *l3.ARPEntry) error
}

// ArpVppRead provides read methods for ARPs
type ArpVppRead interface {
	// DumpArpEntries dumps ARPs from VPP and fills them into the provided static route map.
	DumpArpEntries() ([]*ArpDetails, error)
}

// ProxyArpRangesDetails holds info about proxy ARP range as a proto modeled data
type ProxyArpRangesDetails struct {
	Range *l3.ProxyARP_Range
}

// ProxyArpInterfaceDetails holds info about proxy ARP interfaces as a proto modeled data
type ProxyArpInterfaceDetails struct {
	Interface *l3.ProxyARP_Interface
	Meta      *ProxyArpInterfaceMeta
}

// ProxyArpInterfaceMeta contains interface vpp index
type ProxyArpInterfaceMeta struct {
	SwIfIndex uint32
}

// ProxyArpVppAPI provides methods for managing proxy ARP entries
type ProxyArpVppAPI interface {
	ProxyArpVppRead

	// EnableProxyArpInterface enables interface for proxy ARP
	EnableProxyArpInterface(ifName string) error
	// DisableProxyArpInterface disables interface for proxy ARP
	DisableProxyArpInterface(ifName string) error
	// AddProxyArpRange adds new IP range for proxy ARP
	AddProxyArpRange(firstIP, lastIP []byte) error
	// DeleteProxyArpRange removes proxy ARP IP range
	DeleteProxyArpRange(firstIP, lastIP []byte) error
}

// ProxyArpVppRead provides read methods for proxy ARPs
type ProxyArpVppRead interface {
	// DumpProxyArpRanges returns configured proxy ARP ranges
	DumpProxyArpRanges() ([]*ProxyArpRangesDetails, error)
	// DumpProxyArpRanges returns configured proxy ARP interfaces
	DumpProxyArpInterfaces() ([]*ProxyArpInterfaceDetails, error)
}

// RouteDetails is object returned as a VPP dump. It contains static route data in proto format, and VPP-specific
// metadata
type RouteDetails struct {
	Route *l3.Route
	Meta  *RouteMeta
}

// FibMplsLabel is object returned with route dump.
type FibMplsLabel struct {
	IsUniform bool
	Label     uint32
	TTL       uint8
	Exp       uint8
}

// RouteMeta holds fields returned from the VPP as details which are not in the model
type RouteMeta struct {
	TableName         string
	OutgoingIfIdx     uint32
	IsIPv6            bool
	Afi               uint8
	IsLocal           bool
	IsUDPEncap        bool
	IsUnreach         bool
	IsProhibit        bool
	IsResolveHost     bool
	IsResolveAttached bool
	IsDvr             bool
	IsSourceLookup    bool
	NextHopID         uint32
	RpfID             uint32
	LabelStack        []FibMplsLabel
}

// RouteVppAPI provides methods for managing routes
type RouteVppAPI interface {
	RouteVppRead

	// VppAddRoute adds new route, according to provided input.
	// Every route has to contain VRF ID (default is 0).
	VppAddRoute(route *l3.Route) error
	// VppDelRoute removes old route, according to provided input.
	// Every route has to contain VRF ID (default is 0).
	VppDelRoute(route *l3.Route) error
}

// RouteVppRead provides read methods for routes
type RouteVppRead interface {
	// DumpRoutes dumps l3 routes from VPP and fills them
	// into the provided static route map.
	DumpRoutes() ([]*RouteDetails, error)
}

// VrfTableVppAPI provides methods for managing VRF tables.
type VrfTableVppAPI interface {
	VrfTableVppRead

	// AddVrfTable adds new VRF table.
	AddVrfTable(table *l3.VrfTable) error
	// DelVrfTable deletes existing VRF table.
	DelVrfTable(table *l3.VrfTable) error
}

// VrfTableVppRead provides read methods for VRF tables.
type VrfTableVppRead interface {
	// DumpVrfTables dumps all configured VRF tables.
	DumpVrfTables() ([]*l3.VrfTable, error)
}

// IPNeighVppAPI provides methods for managing IP scan neighbor configuration
type IPNeighVppAPI interface {
	// SetIPScanNeighbor configures IP scan neighbor to the VPP
	SetIPScanNeighbor(data *l3.IPScanNeighbor) error
	// GetIPScanNeighbor returns IP scan neighbor configuration from the VPP
	GetIPScanNeighbor() (*l3.IPScanNeighbor, error)
}

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel, ifaceidx.IfaceMetadataIndex, logging.Logger) L3VppAPI
}

func CompatibleL3VppHandler(
	ch govppapi.Channel,
	ifIdx ifaceidx.IfaceMetadataIndex,
	log logging.Logger,
) L3VppAPI {
	for ver, h := range Versions {
		if err := ch.CheckCompatiblity(h.Msgs...); err != nil {
			log.Debugf("version %s not compatible", ver)
			continue
		}
		log.Debug("found compatible version:", ver)
		return h.New(ch, ifIdx, log)
	}
	panic("no compatible version available")
}
