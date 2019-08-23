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

package descriptor

import (
	"fmt"

	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/vpp-agent/api/models/netalloc"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/netalloc/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/netalloc/utils"
)

const (
	// AddrAllocDescriptorName is the name of the descriptor for allocating
	// IP/MAC/... addresses.
	AddrAllocDescriptorName = "netalloc-address"
)

// AddrAllocDescriptor just validates and parses allocated addresses.
type AddrAllocDescriptor struct {
	log logging.Logger
}

// NewAddrAllocDescriptor creates a new instance of AddrAllocDescriptor.
func NewAddrAllocDescriptor(log logging.PluginLogger) (descr *kvs.KVDescriptor) {
	ctx := &AddrAllocDescriptor{
		log: log.NewLogger("address-alloc-descriptor"),
	}
	typedDescr := &adapter.AddrAllocDescriptor{
		Name:          AddrAllocDescriptorName,
		NBKeyPrefix:   netalloc.ModelAddressAllocation.KeyPrefix(),
		ValueTypeName: netalloc.ModelAddressAllocation.ProtoName(),
		KeySelector:   netalloc.ModelAddressAllocation.IsKeyValid,
		KeyLabel:      netalloc.ModelAddressAllocation.StripKeyPrefix,
		WithMetadata:  true,
		Validate:      ctx.Validate,
		Create:        ctx.Create,
		Delete:        ctx.Delete,
	}
	descr = adapter.NewAddrAllocDescriptor(typedDescr)
	return
}

// Validate checks if the address can be parsed.
func (d *AddrAllocDescriptor) Validate(key string, addrAlloc *netalloc.AddressAllocation) (err error) {
	_, err = d.parseAddr(addrAlloc)
	return err
}

// Create parses the address and stores it into the metadata.
func (d *AddrAllocDescriptor) Create(key string, addrAlloc *netalloc.AddressAllocation) (metadata *netalloc.AddrAllocMetadata, err error) {
	return d.parseAddr(addrAlloc)
}

// Delete is NOOP.
func (d *AddrAllocDescriptor) Delete(key string, addrAlloc *netalloc.AddressAllocation, metadata *netalloc.AddrAllocMetadata) (err error) {
	return err
}

// parseAddr tries to parse the allocated address.
func (d *AddrAllocDescriptor) parseAddr(addrAlloc *netalloc.AddressAllocation) (parsed *netalloc.AddrAllocMetadata, err error) {
	switch addrAlloc.AddressType {
	case netalloc.AddressType_IPV4_ADDR:
		fallthrough
	case netalloc.AddressType_IPV4_GW:
		fallthrough
	case netalloc.AddressType_IPV6_ADDR:
		fallthrough
	case netalloc.AddressType_IPV6_GW:
		ipNet, err := utils.ParseIPAddr(addrAlloc.Address)
		if err != nil {
			return nil, err
		}
		return &netalloc.AddrAllocMetadata{IPAddr: ipNet}, nil
	}
	return nil, fmt.Errorf("address of undefined type: %s", addrAlloc.Address)
}
