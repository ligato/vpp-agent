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

package descriptor

import (
	"fmt"
	"net"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/pkg/models"
	"github.com/pkg/errors"

	nat "github.com/ligato/vpp-agent/api/models/vpp/nat"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	vpp_ifdescriptor "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vpp/natplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/natplugin/vppcalls"
)

const (
	// NAT44GlobalDescriptorName is the name of the descriptor for VPP NAT44 global
	// configuration.
	NAT44GlobalDescriptorName = "vpp-nat44-global"

	// default virtual reassembly configuration
	natReassTimeoutDefault = 2 // seconds
	natMaxReassDefault     = 1024
	natMaxFragDefault      = 5
	natDropFragDefault     = false
)

// A list of non-retriable errors:
var (
	// ErrNATInterfaceFeatureCollision is returned when VPP NAT features assigned
	// to a single interface collide.
	ErrNATInterfaceFeatureCollision = errors.New("VPP NAT interface feature collision")

	// ErrDuplicateNATAddress is returned when VPP NAT address pool contains duplicate
	// IP addresses.
	ErrDuplicateNATAddress = errors.New("Duplicate VPP NAT address")
)

// defaultGlobalCfg is the default NAT44 global configuration.
var defaultGlobalCfg = &nat.Nat44Global{
	VirtualReassembly: &nat.VirtualReassembly{
		Timeout:         natReassTimeoutDefault,
		MaxReassemblies: natMaxReassDefault,
		MaxFragments:    natMaxFragDefault,
		DropFragments:   natDropFragDefault,
	},
}

// NAT44GlobalDescriptor teaches KVScheduler how to configure global options for
// VPP NAT44.
type NAT44GlobalDescriptor struct {
	log        logging.Logger
	natHandler vppcalls.NatVppAPI
}

// NewNAT44GlobalDescriptor creates a new instance of the NAT44Global descriptor.
func NewNAT44GlobalDescriptor(natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &NAT44GlobalDescriptor{
		natHandler: natHandler,
		log:        log.NewLogger("nat44-global-descriptor"),
	}

	typedDescr := &adapter.NAT44GlobalDescriptor{
		Name:                 NAT44GlobalDescriptorName,
		NBKeyPrefix:          nat.ModelNat44Global.KeyPrefix(),
		ValueTypeName:        nat.ModelNat44Global.ProtoName(),
		KeySelector:          nat.ModelNat44Global.IsKeyValid,
		ValueComparator:      ctx.EquivalentNAT44Global,
		Validate:             ctx.Validate,
		Create:               ctx.Create,
		Delete:               ctx.Delete,
		Update:               ctx.Update,
		Retrieve:             ctx.Retrieve,
		DerivedValues:        ctx.DerivedValues,
		RetrieveDependencies: []string{vpp_ifdescriptor.InterfaceDescriptorName},
	}
	return adapter.NewNAT44GlobalDescriptor(typedDescr)
}

// EquivalentNAT44Global compares two NAT44 global configs for equality.
func (d *NAT44GlobalDescriptor) EquivalentNAT44Global(key string, oldGlobalCfg, newGlobalCfg *nat.Nat44Global) bool {
	if oldGlobalCfg.Forwarding != newGlobalCfg.Forwarding {
		return false
	}
	if !proto.Equal(getVirtualReassembly(oldGlobalCfg), getVirtualReassembly(newGlobalCfg)) {
		return false
	}

	// Note: interfaces & addresses are not compared here as they are represented
	//       via derived kv-pairs
	return true
}

// Validate validates VPP NAT44 global configuration.
func (d *NAT44GlobalDescriptor) Validate(key string, globalCfg *nat.Nat44Global) error {
	// check NAT interface features for collisions
	natIfaceMap := make(map[string]*natIface)
	for _, iface := range globalCfg.NatInterfaces {
		if _, hasEntry := natIfaceMap[iface.Name]; !hasEntry {
			natIfaceMap[iface.Name] = &natIface{}
		}
		ifaceCfg := natIfaceMap[iface.Name]
		if iface.IsInside {
			ifaceCfg.in++
		} else {
			ifaceCfg.out++
		}
		if iface.OutputFeature {
			ifaceCfg.output++
		}
	}
	natIfaceCollisionErr := kvs.NewInvalidValueError(ErrNATInterfaceFeatureCollision, "nat_interfaces")
	for _, ifaceCfg := range natIfaceMap {
		if ifaceCfg.in > 1 {
			// duplicate IN
			return natIfaceCollisionErr
		}
		if ifaceCfg.out > 1 {
			// duplicate OUT
			return natIfaceCollisionErr
		}
		if ifaceCfg.output == 1 && (ifaceCfg.in+ifaceCfg.out > 1) {
			// OUTPUT interface cannot be both IN and OUT
			return natIfaceCollisionErr
		}
	}

	// check NAT address pool for duplicities
	var snPool, tnPool []net.IP
	for _, addr := range globalCfg.AddressPool {
		ipAddr := net.ParseIP(addr.Address)
		if ipAddr == nil {
			// validated by NAT44Address descriptor
			continue
		}
		var pool *[]net.IP
		if addr.TwiceNat {
			pool = &tnPool
		} else {
			pool = &snPool
		}
		for _, ipAddr2 := range *pool {
			if ipAddr.Equal(ipAddr2) {
				return kvs.NewInvalidValueError(ErrDuplicateNATAddress,
					fmt.Sprintf("address_pool.address=%s", addr.Address))
			}
		}
		*pool = append(*pool, ipAddr)
	}
	return nil
}

// Create applies NAT44 global options.
func (d *NAT44GlobalDescriptor) Create(key string, globalCfg *nat.Nat44Global) (metadata interface{}, err error) {
	return d.Update(key, defaultGlobalCfg, globalCfg, nil)
}

// Delete sets NAT44 global options back to the defaults.
func (d *NAT44GlobalDescriptor) Delete(key string, globalCfg *nat.Nat44Global, metadata interface{}) error {
	_, err := d.Update(key, globalCfg, defaultGlobalCfg, metadata)
	return err
}

// Update updates NAT44 global options.
func (d *NAT44GlobalDescriptor) Update(key string, oldGlobalCfg, newGlobalCfg *nat.Nat44Global, oldMetadata interface{}) (newMetadata interface{}, err error) {
	// update forwarding
	if oldGlobalCfg.Forwarding != newGlobalCfg.Forwarding {
		if err = d.natHandler.SetNat44Forwarding(newGlobalCfg.Forwarding); err != nil {
			err = errors.Errorf("failed to set NAT44 forwarding to %t: %v", newGlobalCfg.Forwarding, err)
			d.log.Error(err)
			return nil, err
		}
	}

	// update virtual reassembly for IPv4
	if !proto.Equal(getVirtualReassembly(oldGlobalCfg), getVirtualReassembly(newGlobalCfg)) {
		if err = d.natHandler.SetVirtualReassemblyIPv4(getVirtualReassembly(newGlobalCfg)); err != nil {
			err = errors.Errorf("failed to set NAT virtual reassembly for IPv4: %v", err)
			d.log.Error(err)
			return nil, err
		}
	}

	return nil, nil
}

// Retrieve returns the current NAT44 global configuration.
func (d *NAT44GlobalDescriptor) Retrieve(correlate []adapter.NAT44GlobalKVWithMetadata) ([]adapter.NAT44GlobalKVWithMetadata, error) {
	globalCfg, err := d.natHandler.Nat44GlobalConfigDump()
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	origin := kvs.FromNB
	if proto.Equal(globalCfg, defaultGlobalCfg) {
		origin = kvs.FromSB
	}

	retrieved := []adapter.NAT44GlobalKVWithMetadata{{
		Key:    models.Key(globalCfg),
		Value:  globalCfg,
		Origin: origin,
	}}

	return retrieved, nil
}

// DerivedValues derives:
//   - nat.NatAddress for every IP address to be added into the NAT address pool,
//   - nat.NatInterface for every interface with assigned NAT configuration.
func (d *NAT44GlobalDescriptor) DerivedValues(key string, globalCfg *nat.Nat44Global) (derValues []kvs.KeyValuePair) {
	// NAT addresses
	for _, natAddr := range globalCfg.AddressPool {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   nat.AddressNAT44Key(natAddr.Address, natAddr.TwiceNat),
			Value: natAddr,
		})
	}
	// NAT interfaces
	for _, natIface := range globalCfg.NatInterfaces {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   nat.InterfaceNAT44Key(natIface.Name, natIface.IsInside),
			Value: natIface,
		})
	}
	return derValues
}

// natIface accumulates NAT interface configuration for validation purposes.
type natIface struct {
	// feature assignment counters
	in     int
	out    int
	output int
}

func getVirtualReassembly(globalCfg *nat.Nat44Global) *nat.VirtualReassembly {
	if globalCfg.VirtualReassembly == nil {
		return defaultGlobalCfg.VirtualReassembly
	}
	return globalCfg.VirtualReassembly
}