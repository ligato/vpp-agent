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
	"net"
	"strings"

	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/mockcalls"
	interfaces "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/model"
	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

const (
	// InterfaceDescriptorName is the name of the descriptor for mock interfaces.
	InterfaceDescriptorName = "mock-interface"

	// how many characters interface name is allowed to have
	// (defined just to showcase Validation)
	nameLengthLimit = 15
)

// Example of some validation errors:
var (
	// ErrInterfaceWithoutName is returned when interface configuration has
	// undefined Name attribute.
	ErrInterfaceWithoutName = errors.New("mock interface defined without logical name")

	// ErrInterfaceNameTooLong is returned when mock interface name exceeds the length limit.
	ErrInterfaceNameTooLong = errors.New("mock interface logical name exceeds the length limit (15 characters)")

	// ErrInterfaceWithoutType is returned when mock interface configuration has undefined
	// Type attribute.
	ErrInterfaceWithoutType = errors.New("mock interface defined without type")
)

// InterfaceDescriptor teaches KVScheduler how to configure mock interfaces.
type InterfaceDescriptor struct {

	// dependencies
	log          logging.Logger
	ifaceHandler mockcalls.MockIfaceAPI
}

// NewInterfaceDescriptor creates a new instance of the Interface descriptor.
func NewInterfaceDescriptor(ifaceHandler mockcalls.MockIfaceAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	// descriptors are supposed to be stateless and this principle is not broken
	// here - we only need to keep context consisting of references to logger
	// and the interface handler for mock SB, to be used inside the CRUD methods.
	descrCtx := &InterfaceDescriptor{
		ifaceHandler: ifaceHandler,
		log:          log.NewLogger("mock-iface-descriptor"),
	}

	// use adapter to convert typed descriptor into generic descriptor API
	typedDescr := &adapter.InterfaceDescriptor{
		Name:               InterfaceDescriptorName,
		NBKeyPrefix:        interfaces.ModelInterface.KeyPrefix(),
		ValueTypeName:      interfaces.ModelInterface.ProtoName(),
		KeySelector:        interfaces.ModelInterface.IsKeyValid,
		KeyLabel:           interfaces.ModelInterface.StripKeyPrefix,
		ValueComparator:    descrCtx.EquivalentInterfaces,
		WithMetadata:       true,
		MetadataMapFactory: descrCtx.MetadataFactory,
		Validate:           descrCtx.Validate,
		Create:             descrCtx.Create,
		Delete:             descrCtx.Delete,
		Update:             descrCtx.Update,
		UpdateWithRecreate: descrCtx.UpdateWithRecreate,
		Retrieve:           descrCtx.Retrieve,
	}
	return adapter.NewInterfaceDescriptor(typedDescr)
}

// EquivalentInterfaces is case-insensitive comparison function for
// interfaces.Interface, also ignoring the order of assigned IP addresses.
func (d *InterfaceDescriptor) EquivalentInterfaces(key string, oldIntf, newIntf *interfaces.Interface) bool {
	// attributes compared as usually:
	if oldIntf.Name != newIntf.Name ||
		oldIntf.Type != newIntf.Type ||
		oldIntf.Enabled != newIntf.Enabled {
		return false
	}

	// compare MAC addresses case-insensitively (also handle unspecified MAC address)
	if newIntf.PhysAddress != "" &&
		strings.ToLower(oldIntf.PhysAddress) != strings.ToLower(newIntf.PhysAddress) {
		return false
	}

	return true
}

// MetadataFactory is a factory for index-map customized for VPP interfaces.
func (d *InterfaceDescriptor) MetadataFactory() idxmap.NamedMappingRW {
	return idxvpp.NewNameToIndex(d.log, "mock-iface-index", nil)
}

// Validate validates VPP interface configuration.
func (d *InterfaceDescriptor) Validate(key string, intf *interfaces.Interface) error {
	// validate name
	if name := intf.GetName(); name == "" {
		return kvs.NewInvalidValueError(ErrInterfaceWithoutName, "name")
	} else if len(name) > nameLengthLimit {
		return kvs.NewInvalidValueError(ErrInterfaceNameTooLong, "name")
	}

	// validate type
	if intf.Type == interfaces.Interface_UNDEFINED_TYPE {
		return kvs.NewInvalidValueError(ErrInterfaceWithoutType, "type")
	}

	// validate MAC address
	hwAddr := intf.GetPhysAddress()
	if hwAddr != "" {
		_, err := net.ParseMAC(hwAddr)
		if err != nil {
			return kvs.NewInvalidValueError(err, "phys_address")
		}
	}

	return nil
}

// Create creates mock interface.
func (d *InterfaceDescriptor) Create(key string, value *interfaces.Interface) (metadata *idxvpp.OnlyIndex, err error) {
	// create interface of the given type and with the given name
	var sbIfaceHandle uint32
	if value.Type == interfaces.Interface_LOOPBACK {
		sbIfaceHandle, err = d.ifaceHandler.CreateLoopbackInterface(value.Name)
	} else {
		sbIfaceHandle, err = d.ifaceHandler.CreateTapInterface(value.Name)
	}
	if err != nil {
		return nil, err
	}
	metadata = &idxvpp.OnlyIndex{Index: sbIfaceHandle}

	// set interface UP if requested
	if value.Enabled {
		err = d.ifaceHandler.InterfaceAdminUp(sbIfaceHandle)
		if err != nil {
			return nil, err
		}
	}

	// set interface MAC address if requested
	if value.PhysAddress != "" {
		err = d.ifaceHandler.SetInterfaceMac(sbIfaceHandle, value.PhysAddress)
		if err != nil {
			return nil, err
		}
	}

	return metadata, nil
}

// Delete removes a mock interface.
func (d *InterfaceDescriptor) Delete(key string, value *interfaces.Interface, metadata *idxvpp.OnlyIndex) error {
	var err error
	if value.Type == interfaces.Interface_LOOPBACK {
		err = d.ifaceHandler.DeleteLoopbackInterface(metadata.Index)
	} else {
		err = d.ifaceHandler.DeleteTapInterface(metadata.Index)
	}
	return err
}

// Update updates parameters of a mock interface.
func (d *InterfaceDescriptor) Update(key string, oldValue, newValue *interfaces.Interface, oldMetadata *idxvpp.OnlyIndex) (newMetadata *idxvpp.OnlyIndex, err error) {
	// no need to handle change of the Name or Type:
	//  - different Name implies different key, i.e. completely different interface
	//  - UpdateWithRecreate specifies that change of Type requires full interface re-creation

	// update of the admin status
	if oldValue.Enabled != newValue.Enabled {
		if newValue.Enabled {
			err = d.ifaceHandler.InterfaceAdminUp(oldMetadata.Index)
		} else {
			err = d.ifaceHandler.InterfaceAdminDown(oldMetadata.Index)
		}
		if err != nil {
			return nil, err
		}
	}

	// update of the MAC address
	if oldValue.PhysAddress != newValue.PhysAddress && newValue.PhysAddress != "" {
		err = d.ifaceHandler.SetInterfaceMac(oldMetadata.Index, newValue.PhysAddress)
		if err != nil {
			return nil, err
		}
	}

	return oldMetadata, nil // metadata (sbIfaceIndex) has not changed
}

// UpdateWithRecreate returns true if Type is requested to be changed.
func (d *InterfaceDescriptor) UpdateWithRecreate(key string, oldIntf, newIntf *interfaces.Interface, metadata *idxvpp.OnlyIndex) bool {
	if oldIntf.Type != newIntf.Type {
		return true
	}
	return false
}

// Retrieve returns all interfaces configured in the mock SB.
func (d *InterfaceDescriptor) Retrieve(correlate []adapter.InterfaceKVWithMetadata) (retrieved []adapter.InterfaceKVWithMetadata, err error) {
	ifaces, err := d.ifaceHandler.DumpInterfaces()
	if err != nil {
		return nil, err
	}

	for sbIfaceHandle, iface := range ifaces {
		retrieved = append(retrieved, adapter.InterfaceKVWithMetadata{
			Key:      models.Key(iface),
			Value:    iface,
			Metadata: &idxvpp.OnlyIndex{Index: sbIfaceHandle},
			Origin:   kvs.FromNB, // not considering OBTAINED interfaces in our simplified example
		})
	}
	return retrieved, nil
}
