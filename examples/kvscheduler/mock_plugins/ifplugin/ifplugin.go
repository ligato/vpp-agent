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

// Generate golang code from the protobuf model of our mock interfaces:
//go:generate protoc --proto_path=. --go_out=paths=source_relative:. model/interface.proto

// Generate adapter for the descriptor of our mock interfaces:
//go:generate descriptor-adapter --descriptor-name Interface  --value-type *mock_interfaces.Interface --meta-type *idxvpp.OnlyIndex --import "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/model" --import "go.ligato.io/vpp-agent/v3/pkg/idxvpp" --output-dir "descriptor"

package ifplugin

import (
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/infra"

	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/mockcalls"
	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

// IfPlugin configures mock interfaces.
type IfPlugin struct {
	Deps

	// access to our mock SB
	// - it is recommended to put implementation of every SB call needed for
	//   your descriptor into a separate package `<southband-name>calls/` and
	//   expose them via interface. This will allow to replace access to SB with
	//   mocks and make unit testing easier (in our example these calls are
	//   already mocks)
	ifaceHandler mockcalls.MockIfaceAPI

	// descriptor for interfaces
	ifaceDescriptor *descriptor.InterfaceDescriptor

	// metadata index map (exposed read-only for other plugins)
	intfIndex idxvpp.NameToIndex
}

// Deps lists dependencies of the mock interface plugin.
type Deps struct {
	infra.PluginDeps

	// the plugin depends on KVScheduler because it needs to register the
	// descriptor for interfaces.
	KVScheduler kvs.KVScheduler
}

// Init of a real (not-mock) plugin usually:
//  - loads configuration from a file (if any)
//  - registers descriptors for all objects the plugin implements
//  - potentially starts go routine to watch for some asynchronous events
//    (from which usually sends notifications to KVScheduler via PushSBNotification)
//  - etc.
//
// In this mock ifplugin, we only only create mock SB handlers and register the descriptor.
func (p *IfPlugin) Init() error {
	var err error

	// init handler
	p.ifaceHandler = mockcalls.NewMockIfaceHandler(p.Log)

	// init & register descriptors
	ifaceDescriptor := descriptor.NewInterfaceDescriptor(p.ifaceHandler, p.Log)
	err = p.KVScheduler.RegisterKVDescriptor(ifaceDescriptor)
	if err != nil {
		return err
	}

	// obtain read-only reference to index map with interface metadata
	var withIndex bool
	metadataMap := p.KVScheduler.GetMetadataMap(ifaceDescriptor.Name)
	p.intfIndex, withIndex = metadataMap.(idxvpp.NameToIndex)
	if !withIndex {
		return errors.New("missing index with interface metadata")
	}

	return nil
}

// Close of a real (not-mock) plugin usually:
//  - stops all the associated go routines (if any)
//  - closes channels, registrations, etc..
//
// In this example we do nothing (no need to un-register descriptor).
func (p *IfPlugin) Close() error {
	return nil
}

// GetInterfaceIndex gives read-only access to map with metadata of all configured
// mock interfaces.
func (p *IfPlugin) GetInterfaceIndex() idxvpp.NameToIndex {
	return p.intfIndex
}
