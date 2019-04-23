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
	"errors"
	"net"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"

	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	nat "github.com/ligato/vpp-agent/api/models/vpp/nat"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/natplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/natplugin/vppcalls"
)

const (
	// NAT44AddressDescriptorName is the name of the descriptor for IP addresses
	// from the VPP NAT44 address pool.
	NAT44AddressDescriptorName = "vpp-nat44-address"

	// dependency labels
	addressVrfDep = "vrf-table-exists"
)

// A list of non-retriable errors:
var (
	// ErrInvalidNATAddress is returned when IP address from VPP NAT address pool
	// cannot be parsed.
	ErrInvalidNATAddress = errors.New("Invalid VPP NAT address")
)

// NAT44AddressDescriptor teaches KVScheduler how to add/remove IP addresses
// to/from the VPP NAT44 address pool.
type NAT44AddressDescriptor struct {
	log        logging.Logger
	natHandler vppcalls.NatVppAPI
}

// NewNAT44AddressDescriptor creates a new instance of the NAT44Address descriptor.
func NewNAT44AddressDescriptor(natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &NAT44AddressDescriptor{
		natHandler: natHandler,
		log:        log.NewLogger("nat44-address-descriptor"),
	}

	typedDescr := &adapter.NAT44AddressDescriptor{
		Name:          NAT44AddressDescriptorName,
		KeySelector:   ctx.IsNat44AddressKey,
		ValueTypeName: proto.MessageName(&nat.Nat44Global_Address{}),
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Dependencies:  ctx.Dependencies,
	}
	return adapter.NewNAT44AddressDescriptor(typedDescr)
}

// IsNat44AddressKey returns true if the key is identifying configuration
// for a single address from the NAT44 address pool.
func (d *NAT44AddressDescriptor) IsNat44AddressKey(key string) bool {
	_, _, isNATAddrKey := nat.ParseAddressNAT44Key(key)
	return isNATAddrKey
}

// Validates validate configuration for NAT44 address.
func (d *NAT44AddressDescriptor) Validate(key string, natAddr *nat.Nat44Global_Address) error {
	ipAddr := net.ParseIP(natAddr.Address)
	if ipAddr == nil {
		return kvs.NewInvalidValueError(ErrInvalidNATAddress,"address")
	}
	return nil
}

// Create adds IP address into the NAT44 address pool.
func (d *NAT44AddressDescriptor) Create(key string, natAddr *nat.Nat44Global_Address) (metadata interface{}, err error) {
	err = d.natHandler.AddNat44Address(natAddr.Address, natAddr.VrfId, natAddr.TwiceNat)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	return nil, nil
}

// Delete removes IP address from the NAT44 address pool.
func (d *NAT44AddressDescriptor) Delete(key string, natAddr *nat.Nat44Global_Address, metadata interface{}) error {
	err := d.natHandler.DelNat44Address(natAddr.Address, natAddr.VrfId, natAddr.TwiceNat)
	if err != nil {
		d.log.Error(err)
		return err
	}
	return nil
}

// Dependencies lists non-zero VRF as the only dependency.
func (d *NAT44AddressDescriptor) Dependencies(key string, natAddr *nat.Nat44Global_Address) []kvs.Dependency {
	if natAddr.VrfId == 0 || natAddr.VrfId == ^uint32(0) {
		return nil
	}
	return []kvs.Dependency{
		{
			Label: addressVrfDep,
			Key:   l3.VrfTableKey(natAddr.VrfId, l3.VrfTable_IPV4),
		},
	}
}
