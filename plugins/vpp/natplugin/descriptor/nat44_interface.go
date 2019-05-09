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
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	nat "github.com/ligato/vpp-agent/api/models/vpp/nat"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/natplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/natplugin/vppcalls"
)

const (
	// NAT44InterfaceDescriptorName is the name of the descriptor for VPP NAT44
	// features applied to interfaces.
	NAT44InterfaceDescriptorName = "vpp-nat44-interface"

	// dependency labels
	natInterfaceDep = "interface-exists"
)

// NAT44InterfaceDescriptor teaches KVScheduler how to configure VPP NAT interface
// features.
type NAT44InterfaceDescriptor struct {
	log        logging.Logger
	natHandler vppcalls.NatVppAPI
}

// NewNAT44InterfaceDescriptor creates a new instance of the NAT44Interface descriptor.
func NewNAT44InterfaceDescriptor(natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &NAT44InterfaceDescriptor{
		natHandler: natHandler,
		log:        log.NewLogger("nat44-iface-descriptor"),
	}

	typedDescr := &adapter.NAT44InterfaceDescriptor{
		Name:          NAT44InterfaceDescriptorName,
		KeySelector:   ctx.IsNAT44InterfaceKey,
		ValueTypeName: proto.MessageName(&nat.Nat44Global_Interface{}),
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Dependencies:  ctx.Dependencies,
	}
	return adapter.NewNAT44InterfaceDescriptor(typedDescr)
}

// IsNAT44InterfaceKey returns true if the key is identifying NAT-44 configuration
// for interface.
func (d *NAT44InterfaceDescriptor) IsNAT44InterfaceKey(key string) bool {
	_, _, isNATIfaceKey := nat.ParseInterfaceNAT44Key(key)
	return isNATIfaceKey
}

// Create enables NAT44 for an interface.
func (d *NAT44InterfaceDescriptor) Create(key string, natIface *nat.Nat44Global_Interface) (metadata interface{}, err error) {
	err = d.natHandler.EnableNat44Interface(natIface.Name, natIface.IsInside, natIface.OutputFeature)
	if err != nil {
		d.log.Error(err)
		return nil, err

	}
	return nil, nil
}

// Delete disables NAT44 for an interface.
func (d *NAT44InterfaceDescriptor) Delete(key string, natIface *nat.Nat44Global_Interface, metadata interface{}) error {
	err := d.natHandler.DisableNat44Interface(natIface.Name, natIface.IsInside, natIface.OutputFeature)
	if err != nil {
		d.log.Error(err)
		return err

	}
	return nil
}

// Dependencies lists the interface as the only dependency.
func (d *NAT44InterfaceDescriptor) Dependencies(key string, natIface *nat.Nat44Global_Interface) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: natInterfaceDep,
			Key:   interfaces.InterfaceKey(natIface.Name),
		},
	}
}
