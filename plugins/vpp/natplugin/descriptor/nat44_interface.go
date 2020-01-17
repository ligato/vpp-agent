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
	"github.com/ligato/cn-infra/logging"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

const (
	// NAT44InterfaceDescriptorName is the name of the descriptor for VPP NAT44 features applied to interfaces.
	NAT44InterfaceDescriptorName = "vpp-nat44-interface"
)

// NAT44InterfaceDescriptor teaches KVScheduler how to configure VPP NAT interface features.
type NAT44InterfaceDescriptor struct {
	log             logging.Logger
	natHandler      vppcalls.NatVppAPI
	nat44GlobalDesc *NAT44GlobalDescriptor
}

// NewNAT44InterfaceDescriptor creates a new instance of the NAT44Interface descriptor.
func NewNAT44InterfaceDescriptor(nat44GlobalDesc *NAT44GlobalDescriptor,
	natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &NAT44InterfaceDescriptor{
		nat44GlobalDesc: nat44GlobalDesc,
		natHandler:      natHandler,
		log:             log.NewLogger("nat44-iface-descriptor"),
	}

	typedDescr := &adapter.NAT44InterfaceDescriptor{
		Name:          NAT44InterfaceDescriptorName,
		NBKeyPrefix:   nat.ModelNat44Interface.KeyPrefix(),
		ValueTypeName: nat.ModelNat44Interface.ProtoName(),
		KeySelector:   nat.ModelNat44Interface.IsKeyValid,
		KeyLabel:      nat.ModelNat44Interface.StripKeyPrefix,
		Validate:      ctx.Validate,
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Retrieve:      ctx.Retrieve,
		Dependencies:  ctx.Dependencies,
		// retrieve global NAT config first (required for deprecated global NAT interface & address API)
		RetrieveDependencies: []string{NAT44GlobalDescriptorName},
	}
	return adapter.NewNAT44InterfaceDescriptor(typedDescr)
}

// Validate validates VPP NAT44 interface configuration.
func (d *NAT44InterfaceDescriptor) Validate(key string, natIface *nat.Nat44Interface) error {
	if natIface.NatOutside && natIface.NatInside && natIface.OutputFeature {
		// output feature cannot be enabled on interface with both inside & outside NAT enabled
		return kvs.NewInvalidValueError(ErrNATInterfaceFeatureCollision, "output_feature")
	}
	return nil
}

// Create enables NAT44 on an interface.
func (d *NAT44InterfaceDescriptor) Create(key string, natIface *nat.Nat44Interface) (metadata interface{}, err error) {
	if natIface.NatInside {
		err = d.natHandler.EnableNat44Interface(natIface.Name, true, natIface.OutputFeature)
		if err != nil {
			return
		}
	}
	if natIface.NatOutside {
		err = d.natHandler.EnableNat44Interface(natIface.Name, false, natIface.OutputFeature)
		if err != nil {
			return
		}
	}
	return
}

// Delete disables NAT44 on an interface.
func (d *NAT44InterfaceDescriptor) Delete(key string, natIface *nat.Nat44Interface, metadata interface{}) (err error) {
	if natIface.NatInside {
		err = d.natHandler.DisableNat44Interface(natIface.Name, true, natIface.OutputFeature)
		if err != nil {
			return
		}
	}
	if natIface.NatOutside {
		err = d.natHandler.DisableNat44Interface(natIface.Name, false, natIface.OutputFeature)
		if err != nil {
			return
		}
	}
	return
}

// Retrieve returns the current NAT44 interface configuration.
func (d *NAT44InterfaceDescriptor) Retrieve(correlate []adapter.NAT44InterfaceKVWithMetadata) (
	retrieved []adapter.NAT44InterfaceKVWithMetadata, err error) {
	if d.nat44GlobalDesc.UseDeprecatedAPI {
		return nil, nil // NAT interfaces already dumped by global descriptor (deprecated API is in use)
	}
	natIfs, err := d.natHandler.Nat44InterfacesDump()
	if err != nil {
		return nil, err
	}
	for _, natIf := range natIfs {
		retrieved = append(retrieved, adapter.NAT44InterfaceKVWithMetadata{
			Key:    nat.Nat44InterfaceKey(natIf.Name),
			Value:  natIf,
			Origin: kvs.FromNB,
		})
	}
	return
}

// Dependencies lists the interface as the only dependency.
func (d *NAT44InterfaceDescriptor) Dependencies(key string, natIface *nat.Nat44Interface) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: natInterfaceDep,
			Key:   interfaces.InterfaceKey(natIface.Name),
		},
	}
}
