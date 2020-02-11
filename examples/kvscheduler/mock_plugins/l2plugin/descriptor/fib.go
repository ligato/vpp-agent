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

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	ifdescriptor "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/mockcalls"
	l2 "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

const (
	// FIBDescriptorName is the name of the descriptor for L2 FIBs in the mock SB.
	FIBDescriptorName = "mock-l2-fib"

	// dependency labels
	bridgedInterfaceDep = "bridged-interface"
	bridgeDomainDep     = "bridge-domain"
)

// Example of some validation errors:
var (
	// ErrFIBWithoutHwAddr is returned when mock L2 FIB has undefined hardware
	// address.
	ErrFIBWithoutHwAddr = errors.New("mock L2 FIB defined without hardware address")

	// ErrFIBWithoutBD is returned when mock L2 FIB has undefined bridge domain.
	ErrFIBWithoutBD = errors.New("mock L2 FIB defined without bridge domain")

	// ErrForwardFIBWithoutInterface is returned when mock L2 FORWARD FIB has undefined outgoing interface.
	ErrForwardFIBWithoutInterface = errors.New("mock L2 FORWARD FIB defined without outgoing interface")
)

// FIBDescriptor teaches KVScheduler how to configure L2 FIBs in the mock SB.
type FIBDescriptor struct {
	// dependencies
	log        logging.Logger
	fibHandler mockcalls.MockFIBAPI
}

// NewFIBDescriptor creates a new instance of the FIB descriptor.
func NewFIBDescriptor(fibHandler mockcalls.MockFIBAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	// descriptors are supposed to be stateless and this principle is not broken
	// here - we only need to keep context consisting of references to logger
	// and the FIB handler for mock SB, to be used inside the CRUD methods.
	descrCtx := &FIBDescriptor{
		fibHandler: fibHandler,
		log:        log.NewLogger("mock-l2-fib-descriptor"),
	}

	// use adapter to convert typed descriptor into generic descriptor API
	typedDescr := &adapter.FIBDescriptor{
		Name:            FIBDescriptorName,
		NBKeyPrefix:     l2.ModelFIBEntry.KeyPrefix(),
		ValueTypeName:   l2.ModelFIBEntry.ProtoName(),
		KeySelector:     l2.ModelFIBEntry.IsKeyValid,
		KeyLabel:        l2.ModelFIBEntry.StripKeyPrefix,
		ValueComparator: descrCtx.EquivalentFIBs,
		Validate:        descrCtx.Validate,
		Create:          descrCtx.Create,
		Delete:          descrCtx.Delete,
		Retrieve:        descrCtx.Retrieve,
		Dependencies:    descrCtx.Dependencies,

		// Note: Update operation is not defined, which will cause any change
		//       in the FIB configuration to be applied via full re-creation
		//       (Delete for the obsolete config, followed by Create for the new
		//        config).

		// Note: L2 FIBs do not need any metadata in our example with mock SB.

		// Retrieve interfaces and bridge domain first to have the indexes with
		// interface and BD metadata up-to-date when Retrieve for FIBs is called,
		// which then uses the index to translate interface and BD names to the
		// corresponding integer handles used in the mock SB.

		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName, BridgeDomainDescriptorName},
	}
	return adapter.NewFIBDescriptor(typedDescr)
}

// EquivalentFIBs is case-insensitive comparison function for l2.FIBEntry.
func (d *FIBDescriptor) EquivalentFIBs(key string, oldFIB, newFIB *l2.FIBEntry) bool {
	// parameters compared as usually
	if oldFIB.Action != newFIB.Action || oldFIB.BridgeDomain != newFIB.BridgeDomain {
		return false
	}

	// outgoing interface is relevant only for FORWARD FIBs
	if oldFIB.Action == l2.FIBEntry_FORWARD {
		if oldFIB.OutgoingInterface != newFIB.OutgoingInterface {
			return false
		}
	}

	// MAC addresses compared case-insensitively
	return strings.ToLower(oldFIB.PhysAddress) == strings.ToLower(newFIB.PhysAddress)
}

// Validate validates mock L2 FIB configuration.
func (d *FIBDescriptor) Validate(key string, fib *l2.FIBEntry) error {
	// validate MAC address
	if fib.PhysAddress == "" {
		return kvs.NewInvalidValueError(ErrFIBWithoutHwAddr, "phys_address")
	}
	_, err := net.ParseMAC(fib.PhysAddress)
	if err != nil {
		return kvs.NewInvalidValueError(err, "phys_address")
	}

	// validate outgoing interface reference
	if fib.Action == l2.FIBEntry_FORWARD && fib.OutgoingInterface == "" {
		return kvs.NewInvalidValueError(ErrForwardFIBWithoutInterface, "action", "outgoing_interface")
	}

	// validate bridge domain reference
	if fib.BridgeDomain == "" {
		return kvs.NewInvalidValueError(ErrFIBWithoutBD, "bridge_domain")
	}
	return nil
}

// Create adds new L2 FIB.
func (d *FIBDescriptor) Create(key string, fib *l2.FIBEntry) (metadata interface{}, err error) {
	// add L2 FIB
	err = d.fibHandler.CreateL2FIB(fib)
	if err != nil {
		d.log.Error(err)
	}
	return nil, err
}

// Delete removes VPP L2 FIB.
func (d *FIBDescriptor) Delete(key string, fib *l2.FIBEntry, metadata interface{}) error {
	err := d.fibHandler.DeleteL2FIB(fib)
	if err != nil {
		d.log.Error(err)
	}
	return err
}

// Retrieve returns all configured VPP L2 FIBs.
func (d *FIBDescriptor) Retrieve(correlate []adapter.FIBKVWithMetadata) (retrieved []adapter.FIBKVWithMetadata, err error) {
	fibs, err := d.fibHandler.DumpL2FIBs()
	if err != nil {
		return nil, err
	}

	for _, fib := range fibs {
		retrieved = append(retrieved, adapter.FIBKVWithMetadata{
			Key:    models.Key(fib),
			Value:  fib,
			Origin: kvs.FromNB, // not considering OBTAINED FIBs in our simplified example
		})
	}
	return retrieved, nil
}

// Dependencies for FIBs are:
//  * FORWARD FIB: bridge domain + outgoing interface already put into the bridge domain
//  * DROP FIB: bridge domain
func (d *FIBDescriptor) Dependencies(key string, fib *l2.FIBEntry) (dependencies []kvs.Dependency) {
	if fib.Action == l2.FIBEntry_FORWARD {
		// example of a dependency on a derived value
		dependencies = append(dependencies, kvs.Dependency{
			Label: bridgedInterfaceDep,
			Key:   l2.BDInterfaceKey(fib.BridgeDomain, fib.OutgoingInterface),
		})
	} else {
		dependencies = append(dependencies, kvs.Dependency{
			Label: bridgeDomainDep,
			Key:   l2.BridgeDomainKey(fib.BridgeDomain),
		})
	}
	return dependencies
}
