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

// Generate golang code from the protobuf model of our mock BDs and FIBs:
//go:generate protoc --proto_path=. --go_out=paths=source_relative:. model/bridge-domain.proto
//go:generate protoc --proto_path=. --go_out=paths=source_relative:. model/fib.proto

// Generate adapters for the descriptors of our mock BDs and FIBs:
//go:generate descriptor-adapter --descriptor-name BridgeDomain --value-type *mock_l2.BridgeDomain --meta-type *idxvpp.OnlyIndex --import "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model" --import "go.ligato.io/vpp-agent/v3/pkg/idxvpp" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name BDInterface --value-type *mock_l2.BridgeDomain_Interface --import "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name FIB  --value-type *mock_l2.FIBEntry --import "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model" --output-dir "descriptor"

package ifplugin

import (
	"github.com/pkg/errors"

	"go.ligato.io/cn-infra/v2/infra"

	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/descriptor"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/mockcalls"
	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

// L2Plugin configures mock bridge domains and L2 FIBs.
type L2Plugin struct {
	Deps

	// access to our mock SB
	// - it is recommended to put implementation of every SB call needed for
	//   your descriptor into a separate package `<southband-name>calls/` and
	//   expose them via interface. This will allow to replace access to SB with
	//   mocks and make unit testing easier (in our example these calls are
	//   already mocks)
	bdHandler  mockcalls.MockBDAPI
	fibHandler mockcalls.MockFIBAPI

	// descriptors for BDs and L2 FIBs
	bdDescriptor      *descriptor.BridgeDomainDescriptor
	bdIfaceDescriptor *descriptor.BDInterfaceDescriptor // for derived bindings between BD and interfaces
	fibDescriptor     *descriptor.FIBDescriptor

	// metadata index map (exposed read-only for other plugins)
	bdIndex idxvpp.NameToIndex
}

// Deps lists dependencies of the mock interface plugin.
type Deps struct {
	infra.PluginDeps

	// the plugin depends on KVScheduler because it needs to register the
	// descriptors for BDs and FIBs.
	KVScheduler kvs.KVScheduler

	// ifplugin is needed to convert interface name to the corresponding integer
	// handle used in the mock SB
	IfPlugin ifplugin.API
}

// Init of a real (not-mock) plugin usually:
//  - loads configuration from a file (if any)
//  - registers descriptors for all objects the plugin implements
//  - potentially starts go routine to watch for some asynchronous events
//    (from which usually sends notifications to KVScheduler via PushSBNotification)
//  - etc.
//
// In this mock ifplugin, we only create mock SB handlers and register the descriptors.
func (p *L2Plugin) Init() error {
	var err error

	// init BD handler
	p.bdHandler = mockcalls.NewMockBDHandler(p.IfPlugin.GetInterfaceIndex(), p.Log)

	// init & register BD descriptor
	bdDescriptor := descriptor.NewBridgeDomainDescriptor(p.bdHandler, p.Log)
	err = p.KVScheduler.RegisterKVDescriptor(bdDescriptor)
	if err != nil {
		return err
	}

	// obtain read-only reference to index map with bridge domain metadata
	var withIndex bool
	metadataMap := p.KVScheduler.GetMetadataMap(bdDescriptor.Name)
	p.bdIndex, withIndex = metadataMap.(idxvpp.NameToIndex)
	if !withIndex {
		return errors.New("missing index with bridge domain metadata")
	}

	// init & register descriptor for bindings between bridge domains and interfaces
	bdIfaceDescriptor := descriptor.NewBDInterfaceDescriptor(p.bdIndex, p.bdHandler, p.Log)
	err = p.KVScheduler.RegisterKVDescriptor(bdIfaceDescriptor)
	if err != nil {
		return err
	}

	// init FIB handler
	p.fibHandler = mockcalls.NewMockFIBHandler(p.IfPlugin.GetInterfaceIndex(), p.bdIndex, p.Log)

	// init & register descriptor for L2 FIBs
	fibDescriptor := descriptor.NewFIBDescriptor(p.fibHandler, p.Log)
	err = p.KVScheduler.RegisterKVDescriptor(fibDescriptor)
	if err != nil {
		return err
	}

	return nil
}

// Close of a real (not-mock) plugin usually:
//  - stops all the associated go routines (if any)
//  - closes channels, registrations, etc..
//
// In this example we do nothing (no need to un-register descriptor).
func (p *L2Plugin) Close() error {
	return nil
}

// GetBDIndex gives read-only access to map with metadata of all configured
// mock bridge domains.
func (p *L2Plugin) GetBDIndex() idxvpp.NameToIndex {
	return p.bdIndex
}
