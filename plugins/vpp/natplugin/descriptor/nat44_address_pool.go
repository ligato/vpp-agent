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
	"net"
	"strings"

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/vpp-agent/v3/pkg/models"
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
		DerivedValues: ctx.DerivedValues,
		// retrieve global NAT config first (required for deprecated global NAT interface & address API)
		RetrieveDependencies: []string{NAT44GlobalDescriptorName},
	}
	return adapter.NewNAT44AddressPoolDescriptor(typedDescr)
}

// Validate validates configuration for NAT44 IP addresses pool.
func (d *NAT44AddressPoolDescriptor) Validate(key string, natAddr *nat.Nat44AddressPool) error {
	firstIP := net.ParseIP(natAddr.FirstIp)
	if firstIP == nil {
		return kvs.NewInvalidValueError(errInvalidIPAddress, "first_ip")
	}
	if natAddr.LastIp != "" {
		lastIP := net.ParseIP(natAddr.LastIp)
		if lastIP == nil {
			return kvs.NewInvalidValueError(errInvalidIPAddress, "last_ip")
		}
		if bytes.Compare(firstIP, lastIP) > 0 {
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

	// dumping pools
	natPools, err := d.natHandler.Nat44AddressPoolsDump()
	if err != nil {
		return nil, err
	}

	// processing the pool dump
	for _, sbPool := range natPools {
		// try to find NB Pool corresponding to SB Pool (for named pools we link name to SB pools)
		pool := sbPool
		for _, nbPool := range correlate {
			if d.equalNamelessPool(nbPool.Value, sbPool) {
				pool = nbPool.Value // NB pool found
				break
			}
		}

		// creating SB view result
		retrieved = append(retrieved, adapter.NAT44AddressPoolKVWithMetadata{
			Key:    models.Key(pool),
			Value:  pool,
			Origin: kvs.FromNB,
		})
	}
	return
}

// Dependencies lists endpoint-dependent mode and non-zero/non-all-ones VRF as dependencies.
func (d *NAT44AddressPoolDescriptor) Dependencies(key string, natAddr *nat.Nat44AddressPool) (deps []kvs.Dependency) {
	if natAddr.VrfId != 0 && natAddr.VrfId != ^uint32(0) {
		deps = append(deps, kvs.Dependency{
			Label: addressVrfDep,
			Key:   l3.VrfTableKey(natAddr.VrfId, l3.VrfTable_IPV4),
		})
	}
	if !d.natHandler.WithLegacyStartupConf() {
		deps = append(deps, kvs.Dependency{
			Label: addressEpModeDep,
			Key:   nat.Nat44EndpointDepKey,
		})
	}
	return deps
}

// DerivedValues derives:
//   - for twiceNAT address pool the pool itself with exposed IP addresses and VRF in derived key
func (d *NAT44AddressPoolDescriptor) DerivedValues(key string, addrPool *nat.Nat44AddressPool) (derValues []kvs.KeyValuePair) {
	if addrPool.TwiceNat {
		// this derived value may seem as copy of nat44-pool, but nat44-pool key can have 2 forms and in form
		// where nat44-pool key is only pool name, there can't be made dependency based on IP address and
		// twiceNAT bool => this derived key is needed
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   nat.DerivedTwiceNATAddressPoolKey(addrPool.FirstIp, addrPool.LastIp, addrPool.VrfId),
			Value: addrPool,
		})
	}
	return derValues
}

// equalNamelessPool determine equality between 2 Nat44AddressPools ignoring Name field
func (d *NAT44AddressPoolDescriptor) equalNamelessPool(pool1, pool2 *nat.Nat44AddressPool) bool {
	return pool1.VrfId == pool2.VrfId &&
		pool1.TwiceNat == pool2.TwiceNat &&
		equivalentIPv4(pool1.FirstIp, pool2.FirstIp) &&
		equivalentIPv4(pool1.LastIp, pool2.LastIp)
}

func equivalentIPv4(ip1str, ip2str string) bool {
	ip1, err1 := ParseIPv4(ip1str)
	ip2, err2 := ParseIPv4(ip2str)
	if err1 != nil || err2 != nil { // one of values is invalid, but that will handle validator -> compare by strings
		return equivalentTrimmedLowered(ip1str, ip2str)
	}
	return ip1.Equal(ip2) // form doesn't matter, are they representing the same IP value ?
}

func equivalentTrimmedLowered(str1, str2 string) bool {
	return strings.TrimSpace(strings.ToLower(str1)) == strings.TrimSpace(strings.ToLower(str2))
}

// ParseIPv4 parses string <str> to IPv4 address
func ParseIPv4(str string) (net.IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return nil, errors.Errorf(" %q is not ip address", str)
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return nil, errors.Errorf(" %q is not ipv4 address", str)
	}
	return ipv4, nil
}
