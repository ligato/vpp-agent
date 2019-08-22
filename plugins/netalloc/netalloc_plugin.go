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
	"fmt"
	"net"
	"strings"

	"github.com/ligato/cn-infra/infra"

	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/vpp-agent/api/models/netalloc"
	"github.com/ligato/vpp-agent/pkg/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/netalloc/descriptor"
	"bytes"
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
	KVScheduler kvs.KVScheduler
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

	network, iface, addrType, isRef, err := p.parseAddrAllocRef(addrOrAllocRef, ifaceName)
	if !isRef || err != nil {
		return kvs.Dependency{}, false
	}

	return kvs.Dependency{
		Label: addrOrAllocRef,
		Key: models.Key(&netalloc.AddressAllocation{
			NetworkName:   network,
			InterfaceName: iface,
			AddressType:   addrType,
		}),
	}, true
}

// ValidateIPAddress checks validity of address reference or, if <addrOrAllocRef>
// already contains an actual IP address, it tries to parse it.
func (p *Plugin) ValidateIPAddress(addrOrAllocRef, ifaceName string) error {
	_, _, _, isRef, err := p.parseAddrAllocRef(addrOrAllocRef, ifaceName)
	if isRef {
		return err
	}

	_, err = descriptor.ParseIPAddr(addrOrAllocRef)
	return err
}

// GetOrParseIPAddress tries to get allocated IP address referenced by
// <addrOrAllocRef> in the requested form. But if the string contains
// an actual IP address instead of a reference, the address is parsed
// using methods from the net package and returned in the requested form.
// For ADDR_ONLY address form, the returned <addr> will have the mask unset
// and the IP address should be accessed as <addr>.IP
func (p *Plugin) GetOrParseIPAddress(addrOrAllocRef string, ifaceName string,
	addrForm netalloc.IPAddressForm) (addr *net.IPNet, err error) {

	network, iface, addrType, isRef, err := p.parseAddrAllocRef(addrOrAllocRef, ifaceName)
	if isRef && err != nil {
		return nil, err
	}

	if isRef {
		// reference to (to-be) allocated IP address
		allocName := models.Name(&netalloc.AddressAllocation{
			NetworkName:   network,
			InterfaceName: iface,
			AddressType:   addrType,
		})
		allocVal, found := p.addrIndex.GetValue(allocName)
		if !found {
			return nil, fmt.Errorf("failed to find metadata for address allocation '%s'",
				allocName)
		}
		allocMeta, ok := allocVal.(*netalloc.AddrAllocMetadata)
		if !ok {
			return nil, fmt.Errorf("invalid type of metadata stored for address allocation '%s'",
				allocName)
		}
		return p.getIPAddrInGivenForm(allocMeta.IPAddr, addrForm), nil
	}

	// not a reference - try to parse the address
	ipAddr, err := descriptor.ParseIPAddr(addrOrAllocRef)
	if err != nil {
		return nil, err
	}
	return p.getIPAddrInGivenForm(ipAddr, addrForm), nil
}

// CorrelateRetrievedIPs should be used in Retrieve to correlate one or group
// of (model-wise indistinguishable) retrieved IP addresses with the expected
// configuration. The method will replace retrieved addresses with the
// corresponding allocation references from the expected configuration if
// there are any.
// The method returns one IP address or address-allocation reference for every
// address from <retrievedAddrs>.
func (p *Plugin) CorrelateRetrievedIPs(expAddrsOrRefs []string, retrievedAddrs []string,
	ifaceName string, addrForm netalloc.IPAddressForm) (correlated []string) {

	expParsed := make([]*net.IPNet, len(expAddrsOrRefs))
	for i, addr := range expAddrsOrRefs {
		expParsed[i], _ = p.GetOrParseIPAddress(addr, ifaceName, addrForm)
	}

	for _, addr := range retrievedAddrs {
		ipAddr, err := descriptor.ParseIPAddr(addr)
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
	return nil
}

// parseAddrAllocRef parses reference to allocated address.
func (p *Plugin) parseAddrAllocRef(addrAllocRef, expIface string) (
	network, iface string, addrType netalloc.AddressType, isRef bool, err error) {

	if !strings.HasPrefix(addrAllocRef, netalloc.AllocRefPrefix) {
		isRef = false
		return
	}

	isRef = true
	addrAllocRef = strings.TrimPrefix(addrAllocRef, netalloc.AllocRefPrefix)
	parts := strings.Split(addrAllocRef, "/")

	// parse network name
	network = parts[0]
	parts = parts[1:]
	if network == "" {
		err = fmt.Errorf("address allocation reference with empty network name: %s",
			addrAllocRef)
		return
	}

	// parse address type
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		addrType = netalloc.AddressType(netalloc.AddressType_value[lastPart])
	}
	if addrType == netalloc.AddressType_UNDEFINED {
		addrType = netalloc.AddressType_IPV4_ADDR // default
	} else {
		parts = parts[:len(parts)-1]
	}

	if len(parts) > 0 {
		iface = strings.Join(parts, "/")
		if expIface != "" && iface != expIface {
			err = fmt.Errorf("expected different interface name in the address allocation "+
				"reference: %s (expected=%s vs. actual=%s)", addrAllocRef, expIface, iface)
			return
		}
	} else {
		if expIface == "" {
			err = fmt.Errorf("missing interface name in the address allocation reference: %s",
				addrAllocRef)
			return
		} else {
			iface = expIface
		}
	}
	return
}

// getAddrInGivenForm returns IP address in the requested form.
func (p *Plugin) getIPAddrInGivenForm(addr *net.IPNet, form netalloc.IPAddressForm) *net.IPNet {

	switch form {
	case netalloc.IPAddressForm_DEFAULT:
		fallthrough
	case netalloc.IPAddressForm_ADDR_ONLY:
		return &net.IPNet{IP: addr.IP}
	case netalloc.IPAddressForm_ADDR_WITH_MASK:
		return addr
	case netalloc.IPAddressForm_ADDR_NET:
		return &net.IPNet{IP: addr.IP.Mask(addr.Mask), Mask: addr.Mask}
	case netalloc.IPAddressForm_SINGLE_ADDR_NET:
		allOnesIpv4 := net.CIDRMask(32, 32)
		allOnesIpv6 := net.CIDRMask(128, 128)
		if addr.IP.To4() != nil {
			return &net.IPNet{IP: addr.IP, Mask: allOnesIpv4}
		}
		return &net.IPNet{IP: addr.IP, Mask: allOnesIpv6}
	}
	return addr
}