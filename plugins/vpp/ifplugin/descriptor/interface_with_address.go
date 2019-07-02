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
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

const (
	// InterfaceWithAddressDescriptorName is the name of the descriptor for marking
	// interfaces with at least one IP address assigned.
	InterfaceWithAddressDescriptorName = "vpp-interface-has-address"

	// dependency labels
	interfaceHasIPDep = "interface-has-IP"
)

// InterfaceWithAddrDescriptor assigns property key-value pairs to interfaces
// with at least one IP address.
type InterfaceWithAddrDescriptor struct {
	log       logging.Logger
}

// NewInterfaceWithAddrDescriptor creates a new instance of InterfaceWithAddrDescriptor.
func NewInterfaceWithAddrDescriptor(log logging.PluginLogger) *kvs.KVDescriptor {

	descrCtx := &InterfaceWithAddrDescriptor{
		log:       log.NewLogger("interface-has-address-descriptor"),
	}
	return &kvs.KVDescriptor{
		Name:         InterfaceWithAddressDescriptorName,
		KeySelector:  descrCtx.IsInterfaceWithAddressKey,
		Create:       descrCtx.Create,
		Delete:       descrCtx.Delete,
		Dependencies: descrCtx.Dependencies,
	}
}

// IsInterfaceWithAddressKey returns true if the key is a property assigned to interface
// with at least one IP address.
func (d *InterfaceWithAddrDescriptor) IsInterfaceWithAddressKey(key string) bool {
	_, isIfaceWithIPKey := interfaces.ParseInterfaceWithIPKey(key)
	return isIfaceWithIPKey
}

// Create is NOOP (the key-value pair is a property).
func (d *InterfaceWithAddrDescriptor) Create(key string, emptyVal proto.Message) (metadata kvs.Metadata, err error) {
	return nil, nil
}

// Delete is NOOP (the key-value pair is a property)
func (d *InterfaceWithAddrDescriptor) Delete(key string, emptyVal proto.Message, metadata kvs.Metadata) (err error) {
	return nil
}

// Dependencies ensures that the property is created only after at least one IP
// address is successfully assigned to the interface.
func (d *InterfaceWithAddrDescriptor) Dependencies(key string, emptyVal proto.Message) (deps []kvs.Dependency) {
	ifaceName, _ := interfaces.ParseInterfaceWithIPKey(key)
	return []kvs.Dependency{
		{
			Label: interfaceHasIPDep,
			AnyOf: kvs.AnyOfDependency{
				KeyPrefixes: []string{interfaces.InterfaceAddressPrefix(ifaceName)},
			},
		},
	}
}

