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
	"bytes"
	"errors"
	"net"

	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

const (
	// NAT44AddressPoolDescriptorName is the name of the descriptor for NAT44 IP address pools.
	NAT44AddressPoolDescriptorName = "vpp-nat44-address-pool"
)

// A list of non-retriable errors:
var (
	// errInvalidIPAddress is returned when IP address from NAT address pool cannot be parsed.
	errInvalidIPAddress = errors.New("invalid IP address")
	// errInvalidLastPoolAddress is returned when last IP of the pool is not higher than first IP of the pool, or empty.
	errInvalidLastPoolAddress = errors.New("last IP should be higher than first IP, or empty")
)

// NAT44AddressPoolDescriptor teaches KVScheduler how to add/remove VPP NAT44 IP addresses pools.
type NAT44AddressPoolDescriptor struct {
	log             logging.Logger
	natHandler      vppcalls.NatVppAPI
	nat44GlobalDesc *NAT44GlobalDescriptor
}

// NewNAT44AddressPoolDescriptor creates a new instance of the NAT44AddressPoolDescriptor.
func NewNAT44AddressPoolDescriptor(nat44GlobalDesc *NAT44GlobalDescriptor,
	natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &NAT44AddressPoolDescriptor{
		nat44GlobalDesc: nat44GlobalDesc,
		natHandler:      natHandler,
		log:             log.NewLogger("nat44-address-pool-descriptor"),
	}
	typedDescr := &adapter.NAT44AddressPoolDescriptor{
		Name:          NAT44AddressPoolDescriptorName,
		NBKeyPrefix:   nat.ModelNat44AddressPool.KeyPrefix(),
		ValueTypeName: nat.ModelNat44AddressPool.ProtoName(),
		KeySelector:   nat.ModelNat44AddressPool.IsKeyValid,
		KeyLabel:      nat.ModelNat44AddressPool.StripKeyPrefix,
		Validate:      ctx.Validate,
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Retrieve:      ctx.Retrieve,
		Dependencies:  ctx.Dependencies,
		// retrieve global NAT config first (required for deprecated global NAT interface & address API)
		RetrieveDependencies: []string{NAT44GlobalDescriptorName},
	}
	return adapter.NewNAT44AddressPoolDescriptor(typedDescr)
}

// Validate validates configuration for NAT44 IP addresses pool.
func (d *NAT44AddressPoolDescriptor) Validate(key string, natAddr *nat.Nat44AddressPool) error {
	firstIp := net.ParseIP(natAddr.FirstIp)
	if firstIp == nil {
		return kvs.NewInvalidValueError(errInvalidIPAddress, "first_ip")
	}
	if natAddr.LastIp != "" {
		lastIp := net.ParseIP(natAddr.LastIp)
		if lastIp == nil {
			return kvs.NewInvalidValueError(errInvalidIPAddress, "last_ip")
		}
		if bytes.Compare(firstIp, lastIp) > 0 {
			// last IP should be empty or higher than first IP
			return kvs.NewInvalidValueError(errInvalidLastPoolAddress, "last_ip")
		}
	}
	return nil
}

// Create adds IP address pool into VPP NAT44 address pools.
func (d *NAT44AddressPoolDescriptor) Create(key string, natAddr *nat.Nat44AddressPool) (metadata interface{}, err error) {
	return nil,
		d.natHandler.AddNat44AddressPool(natAddr.VrfId, natAddr.FirstIp, natAddr.LastIp, natAddr.TwiceNat)
}

// Delete removes IP address pool from VPP NAT44 address pools.
func (d *NAT44AddressPoolDescriptor) Delete(key string, natAddr *nat.Nat44AddressPool, metadata interface{}) error {
	return d.natHandler.DelNat44AddressPool(natAddr.VrfId, natAddr.FirstIp, natAddr.LastIp, natAddr.TwiceNat)
}

// Retrieve returns VPP IP address pools configured on VPP.
func (d *NAT44AddressPoolDescriptor) Retrieve(correlate []adapter.NAT44AddressPoolKVWithMetadata) (
	retrieved []adapter.NAT44AddressPoolKVWithMetadata, err error) {
	if d.nat44GlobalDesc.UseDeprecatedAPI {
		return nil, nil // NAT IP addresses already dumped by global descriptor (deprecated API is in use)
	}
	natPools, err := d.natHandler.Nat44AddressPoolsDump()
	if err != nil {
		return nil, err
	}
	for _, pool := range natPools {
		retrieved = append(retrieved, adapter.NAT44AddressPoolKVWithMetadata{
			Key:    nat.Nat44AddressPoolKey(pool.VrfId, pool.FirstIp, pool.LastIp),
			Value:  pool,
			Origin: kvs.FromNB,
		})
	}
	return
}

// Dependencies lists non-zero and non-all-ones VRF as the only dependency.
func (d *NAT44AddressPoolDescriptor) Dependencies(key string, natAddr *nat.Nat44AddressPool) []kvs.Dependency {
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
