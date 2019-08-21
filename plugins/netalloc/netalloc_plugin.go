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

//go:generate descriptor-adapter --descriptor-name AddrAlloc --value-type *netalloc.AddressAllocation --meta-type *netalloc.AddrAllocMetadata --import "github.com/ligato/vpp-agent/api/models/netalloc" --output-dir "descriptor"

package netalloc

import (
	"errors"
	"net"

	"github.com/ligato/cn-infra/infra"

	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/netalloc/descriptor"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/vpp-agent/api/models/netalloc"
)

// Plugin implements network allocation features.
// For more information, please refer to the netalloc proto model: api/models/netalloc/netalloc.proto
type Plugin struct {
	Deps

	// address allocation
	addrAllocDescriptor *kvs.KVDescriptor
	addrIndex           idxmap.NamedMapping
}

// Deps lists dependencies of the netalloc plugin.
type Deps struct {
	infra.PluginDeps
	KVScheduler  kvs.KVScheduler
}

// Init initializes netalloc descriptors.
func (p *Plugin) Init() error {
	// init & register descriptors
	p.addrAllocDescriptor = descriptor.NewAddrAllocDescriptor(p.Log)
	err := p.Deps.KVScheduler.RegisterKVDescriptor(p.addrAllocDescriptor)
	if err != nil {
		return err
	}

	// obtain map with metadata of allocated addresses
	p.addrIndex = p.KVScheduler.GetMetadataMap(descriptor.AddrAllocDescriptorName)
	if p.addrIndex == nil {
		return errors.New("missing index with metadata of allocated addresses")
	}
	return nil
}

// Close does nothing.
func (p *Plugin) Close() error {
	return nil
}

// GetAddressAllocDep reads what can be potentially a reference to an allocated
// address (of any type). If <allocRef> is indeed a reference, the function
// returns the corresponding dependency to be passed further into KVScheduler
// from the descriptor. Otherwise <hasAllocDep> is returned as false, and
// <allocRef> should be an actual address and not a reference.
func (p *Plugin) GetAddressAllocDep(addrOrAllocRef, ifaceName string) (
	dep kvs.Dependency, hasAllocDep bool) {

	// TODO
	return kvs.Dependency{}, false
}

// ValidateIPAddress checks validity of address reference or, if <addrOrAllocRef>
// already contains an actual IP address, it tries to parse it.
func (p *Plugin) ValidateIPAddress(addrOrAllocRef, ifaceName string) error {
	// TODO
	return nil
}

// GetOrParseIPAddress tries to get allocated IP address referenced by
// <addrOrAllocRef> in the requested form. But if the string contains
// an actual IP address instead of a reference, the address is parsed
// using methods from the net package and returned in the requested form.
// For ADDR_ONLY address form, the returned <addr> will have the mask unset
// and the IP address should be accessed as <addr>.IP
func (p *Plugin) GetOrParseIPAddress(addrOrAllocRef string, defaultIface string,
	addrForm netalloc.IPAddressForm) (addr *net.IPNet, err error) {

	// TODO
	return nil, nil
}

// CorrelateRetrievedIPs should be used in Retrieve to correlate one or more
// retrieved IP addresses with the expected configuration. The method will
// replace retrieved addresses with the corresponding allocation references
// from the expected configuration if there are any.
// The method returns one IP address or address-allocation reference for every
// address from <retrievedAddrs>.
func (p *Plugin) CorrelateRetrievedIPs(expAddrsOrRefs []string, retrievedAddrs []string,
	defaultIface string, addrForm netalloc.IPAddressForm) []string {

	// TODO
	return nil
}
