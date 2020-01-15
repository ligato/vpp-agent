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
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	interfaces "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/model"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/mockcalls"
	l2 "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model"
	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

const (
	// BDInterfaceDescriptorName is the name of the descriptor for bindings
	// between mock bridge domains and mock interfaces.
	BDInterfaceDescriptorName = "mock-bd-interface"

	// dependency labels
	interfaceDep = "interface-exists"
)

// BDInterfaceDescriptor teaches KVScheduler how to put interface into bridge
// domain in the mock SB.
// A binding between bridge domain and an interface is a value derived from the
// bridge domain. This descriptor is therefore only for derived values.
type BDInterfaceDescriptor struct {
	// dependencies
	log       logging.Logger
	bdIndex   idxvpp.NameToIndex
	bdHandler mockcalls.MockBDAPI
}

// NewBDInterfaceDescriptor creates a new instance of the BDInterface descriptor.
func NewBDInterfaceDescriptor(bdIndex idxvpp.NameToIndex, bdHandler mockcalls.MockBDAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	// descriptors are supposed to be stateless and this principle is not broken
	// here - we only need to keep context consisting of references to logger,
	// index with BD metadata and the BD handler, to be used inside the CRUD
	// methods
	descrCtx := &BDInterfaceDescriptor{
		bdIndex:   bdIndex,
		bdHandler: bdHandler,
		log:       log.NewLogger("mock-bd-iface-descriptor"),
	}

	// use adapter to convert typed descriptor into generic descriptor API
	typedDescr := &adapter.BDInterfaceDescriptor{
		Name:          BDInterfaceDescriptorName,
		KeySelector:   descrCtx.IsBDInterfaceKey,
		ValueTypeName: proto.MessageName(&l2.BridgeDomain_Interface{}),
		Create:        descrCtx.Create,
		Delete:        descrCtx.Delete,
		Dependencies:  descrCtx.Dependencies,

		// Note: this descriptor is only for derived values, therefore it doesn't
		//       need to define the Retrieve method - the values under its scope
		//       are derived from bridge domains, which are already Retrieved by
		//       BridgeDomainDescriptor.
	}
	return adapter.NewBDInterfaceDescriptor(typedDescr)
}

// IsBDInterfaceKey returns true if the key is identifying binding between
// bridge domain and interface in the mock SB.
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
	err = d.bdHandler.AddInterfaceToBridgeDomain(bdMeta.GetIndex(), bdIface.Name, bdIface.BridgedVirtualInterface)
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

	err := d.bdHandler.DeleteInterfaceFromBridgeDomain(bdMeta.GetIndex(), bdIface.Name)
	if err != nil {
		d.log.Error(err)
		return err

	}
	return nil
}

// Dependencies lists the interface as the only dependency for the binding.
//
// Note: some bindings derived from a given bridge domain may be pending (waiting
// for their interfaces), but others and the bridge domain itself will be unaffected
// and free to get configured. This is the power of derived values, which allows
// you to break complex items into multiple parts handled separately.
func (d *BDInterfaceDescriptor) Dependencies(key string, value *l2.BridgeDomain_Interface) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: interfaceDep,
			Key:   interfaces.InterfaceKey(value.Name),
		},
	}
}
