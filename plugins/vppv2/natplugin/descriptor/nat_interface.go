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
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"

	"github.com/ligato/vpp-agent/plugins/vppv2/natplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/natplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/nat"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
)

const (
	// NATInterfaceDescriptorName is the name of the descriptor for VPP NAT features
	// applied to interfaces.
	NATInterfaceDescriptorName = "vpp-nat-interface"

	// dependency labels
	interfaceDep = "interface-exists"
)

// NATInterfaceDescriptor teaches KVScheduler how to configure VPP NAT interface
// features.
type NATInterfaceDescriptor struct {
	log        logging.Logger
	natHandler vppcalls.NatVppAPI
}

// NewNATInterfaceDescriptor creates a new instance of the NATInterface descriptor.
func NewNATInterfaceDescriptor(natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *NATInterfaceDescriptor {

	return &NATInterfaceDescriptor{
		natHandler: natHandler,
		log:        log.NewLogger("nat-iface-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *NATInterfaceDescriptor) GetDescriptor() *adapter.NATInterfaceDescriptor {
	return &adapter.NATInterfaceDescriptor{
		Name:               NATInterfaceDescriptorName,
		KeySelector:        d.IsNATInterfaceKey,
		ValueTypeName:      proto.MessageName(&nat.Nat44Global_NatInterface{}),
		Add:                d.Add,
		Delete:             d.Delete,
		ModifyWithRecreate: d.ModifyWithRecreate,
		Dependencies:       d.Dependencies,
	}
}

// IsNATInterfaceKey returns true if the key is identifying NAT configuration
// for interface.
func (d *NATInterfaceDescriptor) IsNATInterfaceKey(key string) bool {
	_, _, isNATIfaceKey := nat.ParseInterfaceKey(key)
	return isNATIfaceKey
}

// Add puts interface into bridge domain.
func (d *NATInterfaceDescriptor) Add(key string, natIface *nat.Nat44Global_NatInterface) (metadata interface{}, err error) {
	err = d.natHandler.EnableNat44Interface(natIface.Name, natIface.IsInside, natIface.OutputFeature)
	if err != nil {
		d.log.Error(err)
		return nil, err

	}
	return nil, nil
}

// Delete removes interface from bridge domain.
func (d *NATInterfaceDescriptor) Delete(key string, natIface *nat.Nat44Global_NatInterface, metadata interface{}) error {
	err := d.natHandler.DisableNat44Interface(natIface.Name, natIface.IsInside, natIface.OutputFeature)
	if err != nil {
		d.log.Error(err)
		return err

	}
	return nil
}

// ModifyWithRecreate returns always true - a change in OUTPUT is always performed via Delete+Add.
func (d *NATInterfaceDescriptor) ModifyWithRecreate(key string, oldNATIface, newNATIface *nat.Nat44Global_NatInterface, metadata interface{}) bool {
	return true
}

// Dependencies lists the interface as the only dependency.
func (d *NATInterfaceDescriptor) Dependencies(key string, natIface *nat.Nat44Global_NatInterface) []scheduler.Dependency {
	return []scheduler.Dependency{
		{
			Label: interfaceDep,
			Key:   interfaces.InterfaceKey(natIface.Name),
		},
	}
}