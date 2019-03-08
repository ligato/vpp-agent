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
	// BondEnslaveDescriptorName is the name of the descriptor for the bond-enslave
	// config-subsection of VPP interfaces.
	BondEnslaveDescriptorName = "bond-enslave-interface"

	// dependency labels
	bondInterfaceExists = "bond-interface-exists"
)

// BondEnslaveDescriptor sets/unsets VPP interfaces as a slave for a bond interface.
type BondEnslaveDescriptor struct {
	log       logging.Logger
	ifHandler vppcalls.InterfaceVppAPI
	ifIndex   ifaceidx.IfaceMetadataIndex
}

// NewBondEnslaveDescriptor creates a new instance of BondEnslaveDescriptor.
func NewBondEnslaveDescriptor(ifHandler vppcalls.InterfaceVppAPI, log logging.PluginLogger) *BondEnslaveDescriptor {
	return &BondEnslaveDescriptor{
		ifHandler: ifHandler,
		log:       log.NewLogger("bond-enslave-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter)
// with the KVScheduler.
func (d *BondEnslaveDescriptor) GetDescriptor() *adapter.BondEnslavementDescriptor {
	return &adapter.BondEnslavementDescriptor{
		Name:          BondEnslaveDescriptorName,
		KeySelector:   d.IsBondEnslaveKey,
		ValueTypeName: proto.MessageName(&interfaces.Interface_BondEnslavement{}),
		Create:        d.Create,
		Delete:        d.Delete,
		Dependencies:  d.Dependencies,
	}
}

// SetInterfaceIndex should be used to provide interface index immediately after
// the descriptor registration.
func (d *BondEnslaveDescriptor) SetInterfaceIndex(ifIndex ifaceidx.IfaceMetadataIndex) {
	d.ifIndex = ifIndex
}

// IsBondEnslaveKey returns true if the key is identifying bond VPP interface.
func (d *BondEnslaveDescriptor) IsBondEnslaveKey(key string) bool {
	_, isValid := interfaces.ParseNameFromBondEnslaveInterfaceKey(key)
	return isValid
}

// Create sets interface as bond slave.
func (d *BondEnslaveDescriptor) Create(key string, bondIf *interfaces.Interface_BondEnslavement) (metadata interface{}, err error) {
	ifName, _ := interfaces.ParseNameFromBondEnslaveInterfaceKey(key)

	ifMeta, found := d.ifIndex.LookupByName(ifName)
	if !found {
		err = errors.Errorf("failed to find bond-slave interface %s", ifName)
		d.log.Error(err)
		return nil, err
	}

	bondIfMeta, found := d.ifIndex.LookupByName(bondIf.BondInterface)
	if !found {
		err = errors.Errorf("failed to find bond interface %s referenced by %s",
			bondIf.BondInterface, ifName)
		d.log.Error(err)
		return nil, err
	}

	err = d.ifHandler.AttachInterfaceToBond(ifMeta.SwIfIndex, bondIfMeta.SwIfIndex, bondIf.IsPassive, bondIf.IsLongTimeout)
	if err != nil {
		d.log.Error(err)
	}
	return nil, err
}

// Delete detaches interface as a bond slave
func (d *BondEnslaveDescriptor) Delete(key string, bondIf *interfaces.Interface_BondEnslavement, metadata interface{}) error {
	ifName, _ := interfaces.ParseNameFromBondEnslaveInterfaceKey(key)

	ifMeta, found := d.ifIndex.LookupByName(ifName)
	if !found {
		err := errors.Errorf("failed to find bond-slave interface %s", ifName)
		d.log.Error(err)
		return err
	}

	err := d.ifHandler.DetachInterfaceFromBond(ifMeta.SwIfIndex)
	if err != nil {
		d.log.Error(err)
	}

	return err
}

// Dependencies lists dependencies for an bond slave VPP interface.
func (d *BondEnslaveDescriptor) Dependencies(key string, bondIf *interfaces.Interface_BondEnslavement) []kvs.Dependency {
	// link between slave interface and the referenced bond interface
	// - satisfied as along as the referenced interface is configured
	return []kvs.Dependency{{
		Label: bondInterfaceExists,
		Key:   interfaces.InterfaceKey(bondIf.BondInterface),
	}}
}
