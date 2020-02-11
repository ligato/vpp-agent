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
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

const (
	// BDInterfaceDescriptorName is the name of the descriptor for bindings between
	// VPP bridge domains and interfaces.
	BDInterfaceDescriptorName = "vpp-bd-interface"

	// dependency labels
	interfaceDep = "interface-exists"
)

// BDInterfaceDescriptor teaches KVScheduler how to put interface into VPP bridge
// domain.
type BDInterfaceDescriptor struct {
	// dependencies
	log       logging.Logger
	bdIndex   idxvpp.NameToIndex
	bdHandler vppcalls.BridgeDomainVppAPI
}

// NewBDInterfaceDescriptor creates a new instance of the BDInterface descriptor.
func NewBDInterfaceDescriptor(bdIndex idxvpp.NameToIndex, bdHandler vppcalls.BridgeDomainVppAPI, log logging.PluginLogger) *BDInterfaceDescriptor {

	return &BDInterfaceDescriptor{
		bdIndex:   bdIndex,
		bdHandler: bdHandler,
		log:       log.NewLogger("bd-iface-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *BDInterfaceDescriptor) GetDescriptor() *adapter.BDInterfaceDescriptor {
	return &adapter.BDInterfaceDescriptor{
		Name:          BDInterfaceDescriptorName,
		KeySelector:   d.IsBDInterfaceKey,
		ValueTypeName: proto.MessageName(&l2.BridgeDomain_Interface{}),
		Create:        d.Create,
		Delete:        d.Delete,
		Dependencies:  d.Dependencies,
	}
}

// IsBDInterfaceKey returns true if the key is identifying binding between
// VPP bridge domain and interface.
func (d *BDInterfaceDescriptor) IsBDInterfaceKey(key string) bool {
	_, _, isBDIfaceKey := l2.ParseBDInterfaceKey(key)
	return isBDIfaceKey
}

// Create puts interface into bridge domain.
func (d *BDInterfaceDescriptor) Create(key string, bdIface *l2.BridgeDomain_Interface) (metadata interface{}, err error) {
	// get bridge domain index
	bdName, _, _ := l2.ParseBDInterfaceKey(key)
	bdMeta, found := d.bdIndex.LookupByName(bdName)
	if !found {
		err = errors.Errorf("failed to obtain metadata for bridge domain %s", bdName)
		d.log.Error(err)
		return nil, err
	}

	// put interface into the bridge domain
	err = d.bdHandler.AddInterfaceToBridgeDomain(bdMeta.GetIndex(), bdIface)
	if err != nil {
		d.log.Error(err)
		return nil, err

	}
	return nil, nil
}

// Delete removes interface from bridge domain.
func (d *BDInterfaceDescriptor) Delete(key string, bdIface *l2.BridgeDomain_Interface, metadata interface{}) error {
	// get bridge domain index
	bdName, _, _ := l2.ParseBDInterfaceKey(key)
	bdMeta, found := d.bdIndex.LookupByName(bdName)
	if !found {
		err := errors.Errorf("failed to obtain metadata for bridge domain %s", bdName)
		d.log.Error(err)
		return err
	}

	err := d.bdHandler.DeleteInterfaceFromBridgeDomain(bdMeta.GetIndex(), bdIface)
	if err != nil {
		d.log.Error(err)
		return err

	}
	return nil
}

// Dependencies lists the interface as the only dependency for the binding.
func (d *BDInterfaceDescriptor) Dependencies(key string, value *l2.BridgeDomain_Interface) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: interfaceDep,
			Key:   interfaces.InterfaceKey(value.Name),
		},
	}
}
