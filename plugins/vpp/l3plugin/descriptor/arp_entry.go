//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package descriptor

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	// ArpDescriptorName is the name of the descriptor.
	ArpDescriptorName = "vpp-arp"

	// dependency labels
	arpEntryInterfaceDep = "interface-exists"
)

// ArpDescriptor teaches KVScheduler how to configure VPP ARPs.
type ArpDescriptor struct {
	log        logging.Logger
	arpHandler vppcalls.ArpVppAPI
	scheduler  kvs.KVScheduler
}

// NewArpDescriptor creates a new instance of the ArpDescriptor.
func NewArpDescriptor(scheduler kvs.KVScheduler,
	arpHandler vppcalls.ArpVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &ArpDescriptor{
		scheduler:  scheduler,
		arpHandler: arpHandler,
		log:        log.NewLogger("arp-descriptor"),
	}

	typedDescr := &adapter.ARPEntryDescriptor{
		Name:                 ArpDescriptorName,
		NBKeyPrefix:          l3.ModelARPEntry.KeyPrefix(),
		ValueTypeName:        l3.ModelARPEntry.ProtoName(),
		KeySelector:          l3.ModelARPEntry.IsKeyValid,
		KeyLabel:             l3.ModelARPEntry.StripKeyPrefix,
		ValueComparator:      ctx.EquivalentArps,
		Create:               ctx.Create,
		Delete:               ctx.Delete,
		Retrieve:             ctx.Retrieve,
		Dependencies:         ctx.Dependencies,
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName},
	}
	return adapter.NewARPEntryDescriptor(typedDescr)
}

// EquivalentArps is comparison function for ARP entries.
func (d *ArpDescriptor) EquivalentArps(key string, oldArp, newArp *l3.ARPEntry) bool {
	return proto.Equal(oldArp, newArp)
}

// Create adds VPP ARP entry.
func (d *ArpDescriptor) Create(key string, arp *l3.ARPEntry) (interface{}, error) {
	if err := d.arpHandler.VppAddArp(arp); err != nil {
		return nil, err
	}
	return nil, nil
}

// Delete removes VPP ARP entry.
func (d *ArpDescriptor) Delete(key string, arp *l3.ARPEntry, metadata interface{}) error {
	if err := d.arpHandler.VppDelArp(arp); err != nil {
		return err
	}
	return nil
}

// Retrieve returns all ARP entries associated with interfaces managed by this agent.
func (d *ArpDescriptor) Retrieve(correlate []adapter.ARPEntryKVWithMetadata) (
	retrieved []adapter.ARPEntryKVWithMetadata, err error,
) {
	// Retrieve VPP ARP entries.
	arpEntries, err := d.arpHandler.DumpArpEntries()
	if err != nil {
		return nil, errors.Errorf("failed to dump VPP arps: %v", err)
	}

	for _, arp := range arpEntries {
		retrieved = append(retrieved, adapter.ARPEntryKVWithMetadata{
			Key:    l3.ArpEntryKey(arp.Arp.Interface, arp.Arp.IpAddress),
			Value:  arp.Arp,
			Origin: kvs.UnknownOrigin,
		})
	}

	return retrieved, nil
}

// Dependencies lists dependencies for a VPP ARP entry.
func (d *ArpDescriptor) Dependencies(key string, arp *l3.ARPEntry) (deps []kvs.Dependency) {
	// the outgoing interface must exist
	if arp.Interface != "" {
		deps = append(deps, kvs.Dependency{
			Label: arpEntryInterfaceDep,
			Key:   interfaces.InterfaceKey(arp.Interface),
		})
	}
	return deps
}
