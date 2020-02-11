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
	"strings"

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	vpp_ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

const (
	// FIBDescriptorName is the name of the descriptor for VPP L2 FIBs.
	FIBDescriptorName = "vpp-l2-fib"

	// dependency labels
	bridgedInterfaceDep = "bridged-interface"
	bridgeDomainDep     = "bridge-domain"
)

// A list of non-retriable errors:
var (
	// ErrFIBWithoutHwAddr is returned when VPP L2 FIB has undefined hardware
	// address.
	ErrFIBWithoutHwAddr = errors.New("VPP L2 FIB defined without hardware address")

	// ErrFIBWithoutBD is returned when VPP L2 FIB has undefined bridge domain.
	ErrFIBWithoutBD = errors.New("VPP L2 FIB defined without bridge domain")

	// ErrForwardFIBWithoutInterface is returned when VPP L2 FORWARD FIB has undefined outgoing interface.
	ErrForwardFIBWithoutInterface = errors.New("VPP L2 FORWARD FIB defined without outgoing interface")
)

// FIBDescriptor teaches KVScheduler how to configure VPP L2 FIBs.
type FIBDescriptor struct {
	// dependencies
	log        logging.Logger
	fibHandler vppcalls.FIBVppAPI
}

// NewFIBDescriptor creates a new instance of the FIB descriptor.
func NewFIBDescriptor(fibHandler vppcalls.FIBVppAPI, log logging.PluginLogger) *FIBDescriptor {

	return &FIBDescriptor{
		fibHandler: fibHandler,
		log:        log.NewLogger("l2-fib-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *FIBDescriptor) GetDescriptor() *adapter.FIBDescriptor {
	return &adapter.FIBDescriptor{
		Name:                 FIBDescriptorName,
		NBKeyPrefix:          l2.ModelFIBEntry.KeyPrefix(),
		ValueTypeName:        l2.ModelFIBEntry.ProtoName(),
		KeySelector:          l2.ModelFIBEntry.IsKeyValid,
		KeyLabel:             l2.ModelFIBEntry.StripKeyPrefix,
		ValueComparator:      d.EquivalentFIBs,
		Validate:             d.Validate,
		Create:               d.Create,
		Delete:               d.Delete,
		Retrieve:             d.Retrieve,
		Dependencies:         d.Dependencies,
		RetrieveDependencies: []string{vpp_ifdescriptor.InterfaceDescriptorName, BridgeDomainDescriptorName},
	}
}

// EquivalentFIBs is case-insensitive comparison function for l2.FIBEntry.
func (d *FIBDescriptor) EquivalentFIBs(key string, oldFIB, newFIB *l2.FIBEntry) bool {
	// parameters compared as usually
	if oldFIB.Action != newFIB.Action || oldFIB.BridgeDomain != newFIB.BridgeDomain {
		return false
	}

	// parameters relevant only for FORWARD FIBs
	if oldFIB.Action == l2.FIBEntry_FORWARD {
		if oldFIB.OutgoingInterface != newFIB.OutgoingInterface ||
			oldFIB.BridgedVirtualInterface != newFIB.BridgedVirtualInterface ||
			oldFIB.StaticConfig != newFIB.StaticConfig {
			return false
		}
	}

	// MAC addresses compared case-insensitively
	return strings.ToLower(oldFIB.PhysAddress) == strings.ToLower(newFIB.PhysAddress)
}

// Validate validates VPP L2 FIB configuration.
func (d *FIBDescriptor) Validate(key string, fib *l2.FIBEntry) error {
	if fib.PhysAddress == "" {
		return kvs.NewInvalidValueError(ErrFIBWithoutHwAddr, "phys_address")
	}
	if fib.Action == l2.FIBEntry_FORWARD && fib.OutgoingInterface == "" {
		return kvs.NewInvalidValueError(ErrForwardFIBWithoutInterface, "action", "outgoing_interface")
	}
	if fib.BridgeDomain == "" {
		return kvs.NewInvalidValueError(ErrFIBWithoutBD, "bridge_domain")
	}
	return nil
}

// Create adds new L2 FIB.
func (d *FIBDescriptor) Create(key string, fib *l2.FIBEntry) (metadata interface{}, err error) {
	// add L2 FIB
	err = d.fibHandler.AddL2FIB(fib)
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
		d.log.Error(err)
		return retrieved, err
	}
	for _, fib := range fibs {
		retrieved = append(retrieved, adapter.FIBKVWithMetadata{
			Key:    l2.FIBKey(fib.Fib.BridgeDomain, fib.Fib.PhysAddress),
			Value:  fib.Fib,
			Origin: kvs.UnknownOrigin, // there can be automatically created FIBs
		})
	}

	return retrieved, nil
}

// Dependencies for FIBs are:
//  * FORWARD FIB: bridge domain + outgoing interface already put into the bridge domain
//  * DROP FIB: bridge domain
func (d *FIBDescriptor) Dependencies(key string, fib *l2.FIBEntry) (dependencies []kvs.Dependency) {
	if fib.Action == l2.FIBEntry_FORWARD {
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
