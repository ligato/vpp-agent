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

	"github.com/golang/protobuf/proto"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

const (
	// NAT44GlobalAddressDescriptorName is the name of the descriptor for IP addresses
	// from the VPP NAT44 address pool.
	NAT44GlobalAddressDescriptorName = "vpp-nat44-global-address"

	// dependency labels
	addressVrfDep = "vrf-table-exists"
)

// A list of non-retriable errors:
var (
	// ErrInvalidNATAddress is returned when IP address from VPP NAT address pool
	// cannot be parsed.
	ErrInvalidNATAddress = errors.New("Invalid VPP NAT address")
)

// NAT44GlobalAddressDescriptor teaches KVScheduler how to add/remove IP addresses
// to/from the VPP NAT44 address pool.
// Deprecated. Functionality moved to NAT44AddressPoolDescriptor. Kept for backward compatibility.
type NAT44GlobalAddressDescriptor struct {
	log        logging.Logger
	natHandler vppcalls.NatVppAPI
}

// NewNAT44GlobalAddressDescriptor creates a new instance of the NAT44Address descriptor.
// Deprecated. Functionality moved to NAT44AddressPoolDescriptor. Kept for backward compatibility.
func NewNAT44GlobalAddressDescriptor(natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &NAT44GlobalAddressDescriptor{
		natHandler: natHandler,
		log:        log.NewLogger("nat44-global-address-descriptor"),
	}

	typedDescr := &adapter.NAT44GlobalAddressDescriptor{
		Name:          NAT44GlobalAddressDescriptorName,
		KeySelector:   ctx.IsNat44DerivedAddressKey,
		ValueTypeName: proto.MessageName(&nat.Nat44Global_Address{}),
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Dependencies:  ctx.Dependencies,
	}
	return adapter.NewNAT44GlobalAddressDescriptor(typedDescr)
}

// IsNat44DerivedAddressKey returns true if the key is identifying configuration
// for a single address from the NAT44 address pool.
func (d *NAT44GlobalAddressDescriptor) IsNat44DerivedAddressKey(key string) bool {
	_, _, isNATAddrKey := nat.ParseDerivedAddressNAT44Key(key)
	return isNATAddrKey
}

// Validate validates configuration for NAT44 address.
func (d *NAT44GlobalAddressDescriptor) Validate(key string, natAddr *nat.Nat44Global_Address) error {
	ipAddr := net.ParseIP(natAddr.Address)
	if ipAddr == nil {
		return kvs.NewInvalidValueError(ErrInvalidNATAddress, "address")
	}
	return nil
}

// Create adds IP address into the NAT44 address pool.
func (d *NAT44GlobalAddressDescriptor) Create(key string, natAddr *nat.Nat44Global_Address) (metadata interface{}, err error) {
	err = d.natHandler.AddNat44AddressPool(natAddr.VrfId, natAddr.Address, "", natAddr.TwiceNat)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	return nil, nil
}

// Delete removes IP address from the NAT44 address pool.
func (d *NAT44GlobalAddressDescriptor) Delete(key string, natAddr *nat.Nat44Global_Address, metadata interface{}) error {
	err := d.natHandler.DelNat44AddressPool(natAddr.VrfId, natAddr.Address, "", natAddr.TwiceNat)
	if err != nil {
		d.log.Error(err)
		return err
	}
	return nil
}

// Dependencies lists non-zero VRF as the only dependency.
func (d *NAT44GlobalAddressDescriptor) Dependencies(key string, natAddr *nat.Nat44Global_Address) []kvs.Dependency {
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
