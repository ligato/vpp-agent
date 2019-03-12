// Copyright (c) 2019 PANTHEON.tech
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
	"github.com/go-errors/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

const (
	// BondInterfaceDescriptorName is the name of the descriptor for the bond-interface
	// config-subsection of VPP interfaces.
	BondedInterfaceDescriptorName = "bond-interface"

	// dependency labels
	interfaceDep = "interface-exists"
)

// BondedInterfaceDescriptor sets/unsets VPP interfaces as a slave for the bond interface.
type BondedInterfaceDescriptor struct {
	log       logging.Logger
	ifHandler vppcalls.InterfaceVppAPI
	ifIndex   ifaceidx.IfaceMetadataIndex
}

// NewBondedInterfaceDescriptor creates a new instance of BondedInterfaceDescriptor.
func NewBondedInterfaceDescriptor(ifHandler vppcalls.InterfaceVppAPI, ifIndex ifaceidx.IfaceMetadataIndex,
	log logging.PluginLogger) (*kvs.KVDescriptor, *BondedInterfaceDescriptor) {
	descriptorCtx := &BondedInterfaceDescriptor{
		ifHandler: ifHandler,
		log:       log.NewLogger("bonded-interface-descriptor"),
		ifIndex:   ifIndex,
	}
	typedDescriptor := &adapter.BondedInterfaceDescriptor{
		Name:          BondedInterfaceDescriptorName,
		KeySelector:   descriptorCtx.IsBondEnslaveKey,
		ValueTypeName: proto.MessageName(&interfaces.BondLink_BondedInterface{}),
		Create:        descriptorCtx.Create,
		Delete:        descriptorCtx.Delete,
		Dependencies:  descriptorCtx.Dependencies,
	}
	return adapter.NewBondedInterfaceDescriptor(typedDescriptor), descriptorCtx
}

// SetInterfaceIndex should be used to provide interface index immediately after
// the descriptor registration.
func (d *BondedInterfaceDescriptor) SetInterfaceIndex(ifIndex ifaceidx.IfaceMetadataIndex) {
	d.ifIndex = ifIndex
}

// IsBondEnslaveKey returns true if the key is identifying bond VPP interface.
func (d *BondedInterfaceDescriptor) IsBondEnslaveKey(key string) bool {
	_, _, isValid := interfaces.ParseBondedInterfaceKey(key)
	return isValid
}

// Create sets an interface as a bond interface slave.
func (d *BondedInterfaceDescriptor) Create(key string, bondedIf *interfaces.BondLink_BondedInterface) (metadata interface{}, err error) {
	bondIf, _, _ := interfaces.ParseBondedInterfaceKey(key)

	bondIfMeta, found := d.ifIndex.LookupByName(bondIf)
	if !found {
		err = errors.Errorf("failed to find bond interface %s", bondIf)
		d.log.Error(err)
		return nil, err
	}

	slaveMeta, found := d.ifIndex.LookupByName(bondedIf.Name)
	if !found {
		err = errors.Errorf("failed to find bond %s slave interface %s", bondIf, bondedIf.Name)
		d.log.Error(err)
		return nil, err
	}

	err = d.ifHandler.AttachInterfaceToBond(slaveMeta.SwIfIndex, bondIfMeta.SwIfIndex, bondedIf.IsPassive, bondedIf.IsLongTimeout)
	if err != nil {
		d.log.Error(err)
	}
	return nil, err
}

// Delete detaches interface as a bond slave
func (d *BondedInterfaceDescriptor) Delete(key string, bondedIf *interfaces.BondLink_BondedInterface, metadata interface{}) error {
	bondIf, _, _ := interfaces.ParseBondedInterfaceKey(key)

	slaveIfMeta, found := d.ifIndex.LookupByName(bondedIf.Name)
	if !found {
		err := errors.Errorf("failed to find bond %s slave interface %s", bondIf, bondedIf.Name)
		d.log.Error(err)
		return err
	}

	err := d.ifHandler.DetachInterfaceFromBond(slaveIfMeta.SwIfIndex)
	if err != nil {
		d.log.Error(err)
	}

	return err
}

// Dependencies lists dependencies for an bond slave VPP interface.
func (d *BondedInterfaceDescriptor) Dependencies(key string, bondedIf *interfaces.BondLink_BondedInterface) []kvs.Dependency {
	// link between slave interface and the referenced bond interface
	// - satisfied as along as the referenced interface is configured
	return []kvs.Dependency{{
		Label: interfaceDep,
		Key:   interfaces.InterfaceKey(bondedIf.Name),
	}}
}
