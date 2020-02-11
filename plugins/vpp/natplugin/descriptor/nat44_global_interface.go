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
	"github.com/golang/protobuf/proto"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

const (
	// NAT44GlobalInterfaceDescriptorName is the name of the descriptor for VPP NAT44
	// features applied to interfaces.
	NAT44GlobalInterfaceDescriptorName = "vpp-nat44-global-interface"

	// dependency labels
	natInterfaceDep = "interface-exists"
)

// NAT44GlobalInterfaceDescriptor teaches KVScheduler how to configure VPP NAT interface
// features.
// Deprecated. Functionality moved to NAT44InterfaceDescriptor. Kept for backward compatibility.
type NAT44GlobalInterfaceDescriptor struct {
	log        logging.Logger
	natHandler vppcalls.NatVppAPI
}

// NewNAT44GlobalInterfaceDescriptor creates a new instance of the NAT44GlobalInterface descriptor.
// Deprecated. Functionality moved to NAT44InterfaceDescriptor. Kept for backward compatibility.
func NewNAT44GlobalInterfaceDescriptor(natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &NAT44GlobalInterfaceDescriptor{
		natHandler: natHandler,
		log:        log.NewLogger("nat44-global-iface-descriptor"),
	}

	typedDescr := &adapter.NAT44GlobalInterfaceDescriptor{
		Name:          NAT44GlobalInterfaceDescriptorName,
		KeySelector:   ctx.IsNAT44DerivedInterfaceKey,
		ValueTypeName: proto.MessageName(&nat.Nat44Global_Interface{}),
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Dependencies:  ctx.Dependencies,
	}
	return adapter.NewNAT44GlobalInterfaceDescriptor(typedDescr)
}

// IsNAT44DerivedInterfaceKey returns true if the key is identifying NAT-44 configuration
// for interface.
func (d *NAT44GlobalInterfaceDescriptor) IsNAT44DerivedInterfaceKey(key string) bool {
	_, _, isNATIfaceKey := nat.ParseDerivedInterfaceNAT44Key(key)
	return isNATIfaceKey
}

// Create enables NAT44 for an interface.
func (d *NAT44GlobalInterfaceDescriptor) Create(key string, natIface *nat.Nat44Global_Interface) (metadata interface{}, err error) {
	err = d.natHandler.EnableNat44Interface(natIface.Name, natIface.IsInside, natIface.OutputFeature)
	if err != nil {
		d.log.Error(err)
		return nil, err

	}
	return nil, nil
}

// Delete disables NAT44 for an interface.
func (d *NAT44GlobalInterfaceDescriptor) Delete(key string, natIface *nat.Nat44Global_Interface, metadata interface{}) error {
	err := d.natHandler.DisableNat44Interface(natIface.Name, natIface.IsInside, natIface.OutputFeature)
	if err != nil {
		d.log.Error(err)
		return err

	}
	return nil
}

// Dependencies lists the interface as the only dependency.
func (d *NAT44GlobalInterfaceDescriptor) Dependencies(key string, natIface *nat.Nat44Global_Interface) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: natInterfaceDep,
			Key:   interfaces.InterfaceKey(natIface.Name),
		},
	}
}
