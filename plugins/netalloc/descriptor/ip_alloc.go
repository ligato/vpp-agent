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
	"net"

	prototypes "github.com/golang/protobuf/ptypes/empty"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc/utils"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
)

const (
	// IPAllocDescriptorName is the name of the descriptor for allocating
	// IP addresses.
	IPAllocDescriptorName = "netalloc-ip-address"
)

// IPAllocDescriptor just validates and parses allocated IP addresses.
type IPAllocDescriptor struct {
	log logging.Logger
}

// NewAddrAllocDescriptor creates a new instance of IPAllocDescriptor.
func NewAddrAllocDescriptor(log logging.PluginLogger) (descr *kvs.KVDescriptor) {
	ctx := &IPAllocDescriptor{
		log: log.NewLogger("ip-address-alloc-descriptor"),
	}
	typedDescr := &adapter.IPAllocDescriptor{
		Name:          IPAllocDescriptorName,
		NBKeyPrefix:   netalloc.ModelIPAllocation.KeyPrefix(),
		ValueTypeName: netalloc.ModelIPAllocation.ProtoName(),
		KeySelector:   netalloc.ModelIPAllocation.IsKeyValid,
		KeyLabel:      netalloc.ModelIPAllocation.StripKeyPrefix,
		WithMetadata:  true,
		Validate:      ctx.Validate,
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Retrieve:      ctx.Retrieve,
		DerivedValues: ctx.DerivedValues,
	}
	descr = adapter.NewIPAllocDescriptor(typedDescr)
	return
}

// Validate checks if the address can be parsed.
func (d *IPAllocDescriptor) Validate(key string, addrAlloc *netalloc.IPAllocation) (err error) {
	_, _, err = d.parseAddr(addrAlloc)
	return err
}

// Create parses the address and stores it into the metadata.
func (d *IPAllocDescriptor) Create(key string, addrAlloc *netalloc.IPAllocation) (metadata *netalloc.IPAllocMetadata, err error) {
	metadata, _, err = d.parseAddr(addrAlloc)
	return
}

// Delete is NOOP.
func (d *IPAllocDescriptor) Delete(key string, addrAlloc *netalloc.IPAllocation, metadata *netalloc.IPAllocMetadata) (err error) {
	return err
}

// DerivedValues derives "neighbour-gateway" key if GW is a neighbour of the interface
// (addresses are from the same IP network).
func (d *IPAllocDescriptor) DerivedValues(key string, addrAlloc *netalloc.IPAllocation) (derValues []kvs.KeyValuePair) {
	_, neighGw, _ := d.parseAddr(addrAlloc)
	if neighGw {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   netalloc.NeighGwKey(addrAlloc.NetworkName, addrAlloc.InterfaceName),
			Value: &prototypes.Empty{},
		})
	}
	return derValues
}

// Retrieve always returns what is expected to exists since Create doesn't really change
// anything in SB.
func (d *IPAllocDescriptor) Retrieve(correlate []adapter.IPAllocKVWithMetadata) (valid []adapter.IPAllocKVWithMetadata, err error) {
	for _, addrAlloc := range correlate {
		if meta, _, err := d.parseAddr(addrAlloc.Value); err == nil {
			valid = append(valid, adapter.IPAllocKVWithMetadata{
				Key:      addrAlloc.Key,
				Value:    addrAlloc.Value,
				Metadata: meta,
				Origin:   kvs.FromNB,
			})
		}
	}
	return valid, nil
}

// parseAddr tries to parse the allocated address.
func (d *IPAllocDescriptor) parseAddr(addrAlloc *netalloc.IPAllocation) (parsed *netalloc.IPAllocMetadata, neighGw bool, err error) {
	ifaceAddr, _, err := utils.ParseIPAddr(addrAlloc.Address, nil)
	if err != nil {
		return nil, false, err
	}
	var gwAddr *net.IPNet
	if addrAlloc.Gw != "" {
		gwAddr, neighGw, err = utils.ParseIPAddr(addrAlloc.Gw, ifaceAddr)
		if err != nil {
			return nil, false, err
		}
	}
	return &netalloc.IPAllocMetadata{IfaceAddr: ifaceAddr, GwAddr: gwAddr}, neighGw, nil
}
