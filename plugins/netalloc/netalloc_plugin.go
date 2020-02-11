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

//go:generate descriptor-adapter --descriptor-name IPAlloc --value-type *netalloc.IPAllocation --meta-type *netalloc.IPAllocMetadata --import "go.ligato.io/vpp-agent/v3/proto/ligato/netalloc" --output-dir "descriptor"

package netalloc

import (
	"bytes"
	"errors"
	"fmt"
	"net"

	"go.ligato.io/cn-infra/v2/infra"

	"go.ligato.io/cn-infra/v2/idxmap"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc/utils"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
)

// Plugin implements network allocation features.
// For more information, please refer to the netalloc proto model: proto/ligato/netalloc/netalloc.proto
type Plugin struct {
	Deps

	// IP address allocation
	ipAllocDescriptor *kvs.KVDescriptor
	ipIndex           idxmap.NamedMapping
}

// Deps lists dependencies of the netalloc plugin.
type Deps struct {
	infra.PluginDeps
	KVScheduler kvs.KVScheduler
}

// Init initializes netalloc descriptors.
func (p *Plugin) Init() error {
	// init & register descriptors
	p.ipAllocDescriptor = descriptor.NewAddrAllocDescriptor(p.Log)
	err := p.Deps.KVScheduler.RegisterKVDescriptor(p.ipAllocDescriptor)
	if err != nil {
		return err
	}

	// obtain map with metadata of allocated addresses
	p.ipIndex = p.KVScheduler.GetMetadataMap(descriptor.IPAllocDescriptorName)
	if p.ipIndex == nil {
		return errors.New("missing index with metadata of allocated addresses")
	}
	return nil
}

// Close does nothing.
func (p *Plugin) Close() error {
	return nil
}

// CreateAddressAllocRef creates reference to an allocated IP address.
func (p *Plugin) CreateAddressAllocRef(network, iface string, getGW bool) string {
	ref := netalloc.AllocRefPrefix + network
	if iface != "" {
		ref += "/" + iface
	}
	if getGW {
		ref += netalloc.AllocRefGWSuffix
	}
	return ref
}

// ParseAddressAllocRef parses reference to an allocated IP address.
func (p *Plugin) ParseAddressAllocRef(addrAllocRef, expIface string) (
	network, iface string, isGW, isRef bool, err error) {
	return utils.ParseAddrAllocRef(addrAllocRef, expIface)
}

// GetAddressAllocDep reads what can be potentially a reference to an allocated
// IP address. If <allocRef> is indeed a reference, the function returns
// the corresponding dependency to be passed further into KVScheduler
// from the descriptor. Otherwise <hasAllocDep> is returned as false, and
// <allocRef> should be an actual address and not a reference.
func (p *Plugin) GetAddressAllocDep(addrOrAllocRef, ifaceName, depLabelPrefix string) (
	dep kvs.Dependency, hasAllocDep bool) {

	network, iface, _, isRef, err := utils.ParseAddrAllocRef(addrOrAllocRef, ifaceName)
	if !isRef || err != nil {
		return kvs.Dependency{}, false
	}

	return kvs.Dependency{
		Label: depLabelPrefix + addrOrAllocRef,
		Key: models.Key(&netalloc.IPAllocation{
			NetworkName:   network,
			InterfaceName: iface,
		}),
	}, true
}

// ValidateIPAddress checks validity of address reference or, if <addrOrAllocRef>
// already contains an actual IP address, it tries to parse it.
func (p *Plugin) ValidateIPAddress(addrOrAllocRef, ifaceName, fieldName string, gwCheck GwValidityCheck) error {
	_, _, isGW, isRef, err := utils.ParseAddrAllocRef(addrOrAllocRef, ifaceName)
	if !isRef {
		_, _, err = utils.ParseIPAddr(addrOrAllocRef, nil)
	} else if err == nil {
		switch gwCheck {
		case GWRefRequired:
			if !isGW {
				err = errors.New("expected GW address reference")
			}
		case GwRefUnexpected:
			if isGW {
				err = errors.New("expected non-GW address reference")
			}
		}
	}
	if err != nil {
		if fieldName != "" {
			return kvs.NewInvalidValueError(err, fieldName)
		} else {
			return kvs.NewInvalidValueError(err)
		}
	}
	return nil
}

// GetOrParseIPAddress tries to get allocated interface (or GW) IP address
// referenced by <addrOrAllocRef> in the requested form. But if the string
// contains/ an actual IP address instead of a reference, the address is parsed
// using methods from the net package and returned in the requested form.
// For ADDR_ONLY address form, the returned <addr> will have the mask unset
// and the IP address should be accessed as <addr>.IP
func (p *Plugin) GetOrParseIPAddress(addrOrAllocRef string, ifaceName string,
	addrForm netalloc.IPAddressForm) (addr *net.IPNet, err error) {

	network, iface, getGW, isRef, err := utils.ParseAddrAllocRef(addrOrAllocRef, ifaceName)
	if isRef && err != nil {
		return nil, err
	}

	if isRef {
		// reference to allocated IP address
		allocName := models.Name(&netalloc.IPAllocation{
			NetworkName:   network,
			InterfaceName: iface,
		})
		allocVal, found := p.ipIndex.GetValue(allocName)
		if !found {
			return nil, fmt.Errorf("failed to find metadata for IP address allocation '%s'",
				allocName)
		}
		allocMeta, ok := allocVal.(*netalloc.IPAllocMetadata)
		if !ok {
			return nil, fmt.Errorf("invalid type of metadata stored for IP address allocation '%s'",
				allocName)
		}
		if getGW {
			if allocMeta.GwAddr == nil {
				return nil, fmt.Errorf("gw address is not defined for IP address allocation '%s'",
					allocName)
			}
			return utils.GetIPAddrInGivenForm(allocMeta.GwAddr, addrForm), nil
		}
		return utils.GetIPAddrInGivenForm(allocMeta.IfaceAddr, addrForm), nil
	}

	// not a reference - try to parse the address
	ipAddr, _, err := utils.ParseIPAddr(addrOrAllocRef, nil)
	if err != nil {
		return nil, err
	}
	return utils.GetIPAddrInGivenForm(ipAddr, addrForm), nil
}

// CorrelateRetrievedIPs should be used in Retrieve to correlate one or group
// of (model-wise indistinguishable) retrieved interface or GW IP addresses
// with the expected configuration. The method will replace retrieved addresses
// with the corresponding allocation references from the expected configuration
// if there are any.
// The method returns one IP address or address-allocation reference for every
// address from <retrievedAddrs>.
func (p *Plugin) CorrelateRetrievedIPs(expAddrsOrRefs []string, retrievedAddrs []string,
	ifaceName string, addrForm netalloc.IPAddressForm) (correlated []string) {

	expParsed := make([]*net.IPNet, len(expAddrsOrRefs))
	for i, addr := range expAddrsOrRefs {
		expParsed[i], _ = p.GetOrParseIPAddress(addr, ifaceName, addrForm)
	}

	for _, addr := range retrievedAddrs {
		ipAddr, _, err := utils.ParseIPAddr(addr, nil)
		if err != nil {
			// invalid - do not try to correlate, just return as is
			correlated = append(correlated, addr)
			continue
		}
		var addrCorrelated bool
		for i, expAddr := range expParsed {
			if expAddr == nil {
				continue
			}
			if bytes.Equal(ipAddr.IP, expAddr.IP) {
				if addrForm == netalloc.IPAddressForm_ADDR_ONLY ||
					bytes.Equal(ipAddr.Mask, expAddr.Mask) {
					// found match in the expected configuration
					correlated = append(correlated, expAddrsOrRefs[i])
					addrCorrelated = true
					break
				}
			}
		}
		if !addrCorrelated {
			// couldn't find match in the expected configuration, just return as is
			correlated = append(correlated, addr)
		}
	}
	return correlated
}
