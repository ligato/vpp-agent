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
	"context"
	"errors"
	"net"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

var (
	// ErrIPNeighborNotImplemented is used for IPScanNeighAPI handlers that are missing implementation.
	ErrIPNeighborNotImplemented = errors.New("ip neighbor config not implemented")

	// ErrTeibUnsupported error is returned if TEIB is not supported on given VPP version.
	ErrTeibUnsupported = errors.New("TEIB is not supported")

	// ErrVRRPUnsupported error is returned if VRRP is not supported on given VPP version.
	ErrVRRPUnsupported = errors.New("VRRP is not supported")
)

// L3VppAPI groups L3 Vpp APIs.
type L3VppAPI interface {
	ArpVppAPI
	ProxyArpVppAPI
	RouteVppAPI
	IPNeighVppAPI
	VrfTableVppAPI
	DHCPProxyAPI
	L3XCVppAPI
	TeibVppAPI
	VrrpVppAPI
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

// DHCPProxyAPI provides methods for managing ARP entries
type DHCPProxyAPI interface {
	DHCPProxyRead

	// CreateDHCPProxy creates dhcp proxy according to provided input
	CreateDHCPProxy(entry *l3.DHCPProxy) error
	// DeleteDHCPProxy deletes created dhcp proxy
	DeleteDHCPProxy(entry *l3.DHCPProxy) error
}

// DHCPProxyRead provides read methods for routes
type DHCPProxyRead interface {
	// DumpDHCPProxy returns configured DHCP proxy
	DumpDHCPProxy() ([]*DHCPProxyDetails, error)
}

// DHCPProxyDetails holds info about DHCP proxy entry as a proto model
type DHCPProxyDetails struct {
	DHCPProxy *l3.DHCPProxy
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
	AddProxyArpRange(firstIP, lastIP []byte, vrfID uint32) error
	// DeleteProxyArpRange removes proxy ARP IP range
	DeleteProxyArpRange(firstIP, lastIP []byte, vrfID uint32) error
}

// ProxyArpVppRead provides read methods for proxy ARPs
type ProxyArpVppRead interface {
	// DumpProxyArpRanges returns configured proxy ARP ranges
	DumpProxyArpRanges() ([]*ProxyArpRangesDetails, error)
	// DumpProxyArpInterfaces returns configured proxy ARP interfaces
	DumpProxyArpInterfaces() ([]*ProxyArpInterfaceDetails, error)
}

// RouteDetails is object returned as a VPP dump. It contains static route data in proto format, and VPP-specific
// metadata
type RouteDetails struct {
	Route *l3.Route
	Meta  *RouteMeta
}

// VrrpDetails is object returned as a VRRP dump.
type VrrpDetails struct {
	Vrrp *l3.VRRPEntry
	Meta *VrrpMeta
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

// VrrpMeta holds fields returned from the VPP as details which are not in the model
type VrrpMeta struct{}

// RouteVppAPI provides methods for managing routes
type RouteVppAPI interface {
	RouteVppRead

	// VppAddRoute adds new route, according to provided input.
	// Every route has to contain VRF ID (default is 0).
	VppAddRoute(ctx context.Context, route *l3.Route) error
	// VppDelRoute removes old route, according to provided input.
	// Every route has to contain VRF ID (default is 0).
	VppDelRoute(ctx context.Context, route *l3.Route) error
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
	// SetVrfFlowHashSettings sets IP flow hash settings for a VRF table.
	SetVrfFlowHashSettings(vrfID uint32, isIPv6 bool, hashFields *l3.VrfTable_FlowHashSettings) error
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
	// DefaultIPScanNeighbor returns default IP scan neighbor configuration
	DefaultIPScanNeighbor() *l3.IPScanNeighbor
}

// TeibVppAPI provides methods for managing VPP tunnel information base.
type TeibVppAPI interface {
	TeibVppRead

	// VppAddTeibEntry adds a new TEIB entry.
	VppAddTeibEntry(ctx context.Context, entry *l3.TeibEntry) error
	// VppDelTeibEntry removes an existing TEIB entry.
	VppDelTeibEntry(ctx context.Context, entry *l3.TeibEntry) error
}

// TeibVppRead provides read methods VPP tunnel information base.
type TeibVppRead interface {
	// DumpTeib dumps TEIB entries from VPP and fills them into the provided TEIB entry map.
	DumpTeib() ([]*l3.TeibEntry, error)
}

// VrrpVppAPI provides methods for managing VPP VRRP.
type VrrpVppAPI interface {
	VppAddVrrp(entry *l3.VRRPEntry) error
	VppDelVrrp(entry *l3.VRRPEntry) error
	VppStartVrrp(entry *l3.VRRPEntry) error
	VppStopVrrp(entry *l3.VRRPEntry) error
	DumpVrrpEntries() ([]*VrrpDetails, error)
}

// Path represents FIB path entry.
type Path struct {
	SwIfIndex  uint32
	NextHop    net.IP
	Weight     uint8
	Preference uint8
}

// L3XC represents configuration for L3XC.
type L3XC struct {
	SwIfIndex uint32
	IsIPv6    bool
	Paths     []Path
}

// L3XCVppRead provides read methods for L3XC configuration.
type L3XCVppRead interface {
	DumpAllL3XC(ctx context.Context) ([]L3XC, error)
	DumpL3XC(ctx context.Context, index uint32) ([]L3XC, error)
}

// L3XCVppAPI provides methods for managing L3XC configuration.
type L3XCVppAPI interface {
	L3XCVppRead

	UpdateL3XC(ctx context.Context, l3xc *L3XC) error
	DeleteL3XC(ctx context.Context, index uint32, ipv6 bool) error
}

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "l3",
	HandlerAPI: (*L3VppAPI)(nil),
})

type NewHandlerFunc func(c vpp.Client, idx ifaceidx.IfaceMetadataIndex, vrfIdx vrfidx.VRFMetadataIndex, addrAlloc netalloc.AddressAllocator, log logging.Logger) L3VppAPI

func AddHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return ch.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			var vrfIdx vrfidx.VRFMetadataIndex
			if a[1] != nil {
				vrfIdx = a[1].(vrfidx.VRFMetadataIndex)
			}
			return h(c, a[0].(ifaceidx.IfaceMetadataIndex), vrfIdx, a[2].(netalloc.AddressAllocator), a[3].(logging.Logger))
		},
	})
}

func CompatibleL3VppHandler(
	c vpp.Client,
	ifIdx ifaceidx.IfaceMetadataIndex,
	vrfIdx vrfidx.VRFMetadataIndex,
	addrAlloc netalloc.AddressAllocator,
	log logging.Logger,
) L3VppAPI {
	if v := Handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, ifIdx, vrfIdx, addrAlloc, log).(L3VppAPI)
	}
	return nil
}
