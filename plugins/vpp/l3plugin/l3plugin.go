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

//go:generate descriptor-adapter --descriptor-name Route --value-type *vpp_l3.Route --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name ARPEntry --value-type *vpp_l3.ARPEntry --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name ProxyARP --value-type *vpp_l3.ProxyARP --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name ProxyARPInterface --value-type *vpp_l3.ProxyARP_Interface --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name IPScanNeighbor --value-type *vpp_l3.IPScanNeighbor --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name VrfTable --value-type *vpp_l3.VrfTable --meta-type *vrfidx.VRFMetadata --import "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx" --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name DHCPProxy --value-type *vpp_l3.DHCPProxy --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name L3XC --value-type *vpp_l3.L3XConnect --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3" --output-dir "descriptor"

package l3plugin

import (
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/infra"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vrfidx"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls/vpp2001"
)

func init() {
	kvscheduler.AddNonRetryableError(vppcalls.ErrIPNeighborNotImplemented)
}

// L3Plugin configures Linux routes and ARP entries using Netlink API.
type L3Plugin struct {
	Deps

	// VPP handler
	l3Handler vppcalls.L3VppAPI

	// index maps
	vrfIndex vrfidx.VRFMetadataIndex
}

type Deps struct {
	infra.PluginDeps
	KVScheduler kvs.KVScheduler
	VPP         govppmux.API
	IfPlugin    ifplugin.API
	AddrAlloc   netalloc.AddressAllocator
	StatusCheck statuscheck.PluginStatusWriter // optional
}

// Init initializes and registers descriptors for Linux ARPs and Routes.
func (p *L3Plugin) Init() (err error) {
	// init handlers
	p.l3Handler = vppcalls.CompatibleL3VppHandler(p.VPP, p.IfPlugin.GetInterfaceIndex(),
		p.vrfIndex, p.AddrAlloc, p.Log)
	if p.l3Handler == nil {
		return errors.Errorf("could not find compatible L3VppHandler")
	}

	// init and register VRF descriptor
	vrfTableDescriptor := descriptor.NewVrfTableDescriptor(p.l3Handler, p.Log)
	if err = p.Deps.KVScheduler.RegisterKVDescriptor(vrfTableDescriptor); err != nil {
		return err
	}
	metadataMap := p.KVScheduler.GetMetadataMap(vrfTableDescriptor.Name)
	var withIndex bool
	p.vrfIndex, withIndex = metadataMap.(vrfidx.VRFMetadataIndex)
	if !withIndex {
		return errors.New("missing index with VRF metadata")
	}

	// set l3 handler again since VRF index was nil before
	p.l3Handler = vppcalls.CompatibleL3VppHandler(p.VPP, p.IfPlugin.GetInterfaceIndex(),
		p.vrfIndex, p.AddrAlloc, p.Log)

	// init & register descriptors
	routeDescriptor := descriptor.NewRouteDescriptor(p.l3Handler, p.AddrAlloc, p.Log)
	arpDescriptor := descriptor.NewArpDescriptor(p.KVScheduler, p.l3Handler, p.Log)
	proxyArpDescriptor := descriptor.NewProxyArpDescriptor(p.KVScheduler, p.l3Handler, p.Log)
	proxyArpIfaceDescriptor := descriptor.NewProxyArpInterfaceDescriptor(p.KVScheduler, p.l3Handler, p.Log)
	ipScanNeighborDescriptor := descriptor.NewIPScanNeighborDescriptor(p.KVScheduler, p.l3Handler, p.Log)
	dhcpProxyDescriptor := descriptor.NewDHCPProxyDescriptor(p.KVScheduler, p.l3Handler, p.Log)
	l3xcDescriptor := descriptor.NewL3XCDescriptor(p.l3Handler, p.IfPlugin.GetInterfaceIndex(), p.Log)

	err = p.Deps.KVScheduler.RegisterKVDescriptor(
		routeDescriptor,
		arpDescriptor,
		proxyArpDescriptor,
		proxyArpIfaceDescriptor,
		ipScanNeighborDescriptor,
		dhcpProxyDescriptor,
		l3xcDescriptor,
	)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *L3Plugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}

// GetVRFIndex gives read-only access to map with metadata of all configured VPP VRFs.
func (p *L3Plugin) GetVRFIndex() vrfidx.VRFMetadataIndex {
	return p.vrfIndex
}
