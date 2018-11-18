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
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	vpp_ifdescriptor "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vppv2/l2plugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/l2"
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
		Name:            FIBDescriptorName,
		KeySelector:     d.IsFIBKey,
		ValueTypeName:   proto.MessageName(&l2.FIBEntry{}),
		ValueComparator: d.EquivalentFIBs,
		// NB keys already covered by the prefix for bridge domains
		Add:                d.Add,
		Delete:             d.Delete,
		ModifyWithRecreate: d.ModifyWithRecreate,
		IsRetriableFailure: d.IsRetriableFailure,
		Dependencies:       d.Dependencies,
		Dump:               d.Dump,
		DumpDependencies:   []string{vpp_ifdescriptor.InterfaceDescriptorName, BridgeDomainDescriptorName},
	}
}

// IsFIBKey returns true if the key is identifying VPP L2 FIB configuration.
func (d *FIBDescriptor) IsFIBKey(key string) bool {
	_, _, isFIBKey := l2.ParseFIBKey(key)
	return isFIBKey
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

// IsRetriableFailure returns <false> for errors related to invalid configuration.
func (d *FIBDescriptor) IsRetriableFailure(err error) bool {
	nonRetriable := []error{
		ErrFIBWithoutBD,
		ErrFIBWithoutHwAddr,
		ErrForwardFIBWithoutInterface,
	}
	for _, nonRetriableErr := range nonRetriable {
		if err == nonRetriableErr {
			return false
		}
	}
	return true
}

// Add adds new L2 FIB.
func (d *FIBDescriptor) Add(key string, fib *l2.FIBEntry) (metadata interface{}, err error) {
	// validate the configuration first
	err = d.validateFIBConfig(fib)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

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

// ModifyWithRecreate always returns true - L2 FIBs are modified via re-creation.
func (d *FIBDescriptor) ModifyWithRecreate(key string, oldFIB, newFIB *l2.FIBEntry, metadata interface{}) bool {
	return true
}

// Dependencies for FIBs are:
//  * FORWARD FIB: bridge domain + outgoing interface already put into the bridge domain
//  * DROP FIB: bridge domain
func (d *FIBDescriptor) Dependencies(key string, fib *l2.FIBEntry) (dependencies []scheduler.Dependency) {
	if fib.Action == l2.FIBEntry_FORWARD {
		dependencies = append(dependencies, scheduler.Dependency{
			Label: bridgedInterfaceDep,
			Key:   l2.BDInterfaceKey(fib.BridgeDomain, fib.OutgoingInterface),
		})
	} else {
		dependencies = append(dependencies, scheduler.Dependency{
			Label: bridgeDomainDep,
			Key:   l2.BridgeDomainKey(fib.BridgeDomain),
		})
	}
	return dependencies
}

// Dump returns all configured VPP L2 FIBs.
func (d *FIBDescriptor) Dump(correlate []adapter.FIBKVWithMetadata) (dump []adapter.FIBKVWithMetadata, err error) {
	fibs, err := d.fibHandler.DumpL2FIBs()
	if err != nil {
		d.log.Error(err)
		return dump, err
	}
	for _, fib := range fibs {
		dump = append(dump, adapter.FIBKVWithMetadata{
			Key:    l2.FIBKey(fib.Fib.BridgeDomain, fib.Fib.PhysAddress),
			Value:  fib.Fib,
			Origin: scheduler.UnknownOrigin, // there can be automatically created FIBs
		})
	}

	var dumpList string
	for _, d := range dump {
		dumpList += fmt.Sprintf("\n - %+v", d)
	}
	d.log.Debugf("Dumping %d L2 FIBs: %v", len(dump), dumpList)

	return dump, nil
}

// validateFIBConfig validates VPP L2 FIB configuration.
func (d *FIBDescriptor) validateFIBConfig(fib *l2.FIBEntry) error {
	if fib.PhysAddress == "" {
		return ErrFIBWithoutHwAddr
	}
	if fib.Action == l2.FIBEntry_FORWARD && fib.OutgoingInterface == "" {
		return ErrForwardFIBWithoutInterface
	}
	if fib.BridgeDomain == "" {
		return ErrFIBWithoutBD
	}
	return nil
}
