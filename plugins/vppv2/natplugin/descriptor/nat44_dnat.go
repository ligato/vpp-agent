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

	"github.com/gogo/protobuf/proto"
	"github.com/go-errors/errors"

	"github.com/ligato/cn-infra/logging"

	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vppv2/natplugin/descriptor/adapter"
	vpp_ifdescriptor "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vppv2/natplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/nat"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
)

const (
	// DNAT44DescriptorName is the name of the descriptor for VPP NAT44
	// Destination-NAT configurations.
	DNAT44DescriptorName = "vpp-nat44-dnat"

	// untaggedDNAT is used as a label for DNAT grouping all untagged static
	// and identity mappings.
	untaggedDNAT = "UNTAGGED-DNAT"

	// dependency labels
	mappingInterfaceDep = "interface-exists"
)

// A list of non-retriable errors:
var (
	// ErrDNAT44WithEmptyLabel is returned when NAT44 DNAT configuration is defined
	// with empty label
	ErrDNAT44WithEmptyLabel = errors.New("NAT44 DNAT configuration defined with empty label")
)

// DNAT44Descriptor teaches KVScheduler how to configure Destination NAT44 in VPP.
type DNAT44Descriptor struct {
	log        logging.Logger
	natHandler vppcalls.NatVppAPI
}

// NewDNAT44Descriptor creates a new instance of the DNAT44 descriptor.
func NewDNAT44Descriptor(natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *DNAT44Descriptor {

	return &DNAT44Descriptor{
		natHandler: natHandler,
		log:        log.NewLogger("nat44-global-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *DNAT44Descriptor) GetDescriptor() *adapter.DNAT44Descriptor {
	return &adapter.DNAT44Descriptor{
		Name:               DNAT44DescriptorName,
		KeySelector:        d.IsDNAT44Key,
		ValueTypeName:      proto.MessageName(&nat.DNat44{}),
		NBKeyPrefix:        nat.PrefixNAT44,
		Add:                d.Add,
		Delete:             d.Delete,
		Modify:             d.Modify,
		IsRetriableFailure: d.IsRetriableFailure,
		Dependencies:       d.Dependencies,
		Dump:               d.Dump,
		DumpDependencies:   []string{vpp_ifdescriptor.InterfaceDescriptorName},
	}
}

// IsDNAT44Key returns true if the key is identifying VPP destination-NAT44.
func (d *DNAT44Descriptor) IsDNAT44Key(key string) bool {
	return strings.HasPrefix(key, nat.DNAT44Prefix)
}

// IsRetriableFailure returns <false> for errors related to invalid configuration.
func (d *DNAT44Descriptor) IsRetriableFailure(err error) bool {
	return err != ErrDNAT44WithEmptyLabel
}

// Add adds new destination-NAT44 configuration.
func (d *DNAT44Descriptor) Add(key string, dnat *nat.DNat44) (metadata interface{}, err error) {
	// Add = Modify from empty DNAT
	return d.Modify(key, &nat.DNat44{Label: dnat.Label}, dnat, nil)
}

// Delete removes existing destination-NAT44 configuration.
func (d *DNAT44Descriptor) Delete(key string, dnat *nat.DNat44, metadata interface{}) error {
	// Delete = Modify into empty DNAT
	_, err := d.Modify(key, dnat,  &nat.DNat44{Label: dnat.Label}, metadata)
	return err
}

// Modify updates destination-NAT44 configuration.
func (d *DNAT44Descriptor) Modify(key string, oldDNAT, newDNAT *nat.DNat44, oldMetadata interface{}) (newMetadata interface{}, err error) {
	// validate configuration first
	err = d.validateDNAT44Config(newDNAT)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	// remove obsolete identity mappings
	for _, oldMapping := range oldDNAT.IdMappings {
		found := false
		for _, newMapping := range newDNAT.IdMappings {
			if proto.Equal(oldMapping, newMapping) {
				found = true
				break
			}
		}
		if !found {
			if err = d.natHandler.DelNat44IdentityMapping(oldMapping, oldDNAT.Label); err != nil {
				err = errors.Errorf("failed to remove identity mapping from DNAT %s: %v", oldDNAT.Label, err)
				d.log.Error(err)
				return nil, err
			}
		}
	}

	// remove obsolete static mappings
	for _, oldMapping := range oldDNAT.StMappings {
		found := false
		for _, newMapping := range newDNAT.StMappings {
			if proto.Equal(oldMapping, newMapping) {
				found = true
				break
			}
		}
		if !found {
			if err = d.natHandler.DelNat44StaticMapping(oldMapping, oldDNAT.Label); err != nil {
				err = errors.Errorf("failed to remove static mapping from DNAT %s: %v", oldDNAT.Label, err)
				d.log.Error(err)
				return nil, err
			}
		}
	}

	// add new identity mappings
	for _, newMapping := range newDNAT.IdMappings {
		found := false
		for _, oldMapping := range oldDNAT.IdMappings {
			if proto.Equal(oldMapping, newMapping) {
				found = true
				break
			}
		}
		if !found {
			if err = d.natHandler.AddNat44IdentityMapping(newMapping, newDNAT.Label); err != nil {
				err = errors.Errorf("failed to add identity mapping for DNAT %s: %v", newDNAT.Label, err)
				d.log.Error(err)
				return nil, err
			}
		}
	}

	// add new static mappings
	for _, newMapping := range newDNAT.StMappings {
		found := false
		for _, oldMapping := range oldDNAT.StMappings {
			if proto.Equal(oldMapping, newMapping) {
				found = true
				break
			}
		}
		if !found {
			if err = d.natHandler.AddNat44StaticMapping(newMapping, newDNAT.Label); err != nil {
				err = errors.Errorf("failed to add static mapping for DNAT %s: %v", newDNAT.Label, err)
				d.log.Error(err)
				return nil, err
			}
		}
	}

	return nil, nil
}

// Dependencies lists external interfaces from mappings as dependencies.
func (d *DNAT44Descriptor) Dependencies(key string, dnat *nat.DNat44) (dependencies []scheduler.Dependency) {
	for _, mapping := range dnat.StMappings {
		if mapping.ExternalInterface != "" {
			dependencies = append(dependencies, scheduler.Dependency{
				Label: mappingInterfaceDep,
				Key:   interfaces.InterfaceKey(mapping.ExternalInterface),
			})
		}
	}
	for _, mapping := range dnat.IdMappings {
		if mapping.Interface != "" {
			dependencies = append(dependencies, scheduler.Dependency{
				Label: mappingInterfaceDep,
				Key:   interfaces.InterfaceKey(mapping.Interface),
			})
		}
	}
	return dependencies
}

// Dump returns the current NAT44 global configuration.
func (d *DNAT44Descriptor) Dump(correlate []adapter.DNAT44KVWithMetadata) (dump []adapter.DNAT44KVWithMetadata, err error) {
	dnatDump, err := d.natHandler.DNat44Dump()
	if err != nil {
		d.log.Error(err)
		return dump, err
	}

	for _, dnat := range dnatDump {
		if dnat.Label == "" {
			dnat.Label = untaggedDNAT
		}
		dump = append(dump, adapter.DNAT44KVWithMetadata{
			Key:    nat.DNAT44Key(dnat.Label),
			Value:  dnat,
			Origin: scheduler.FromNB,
		})
	}


	return nil, nil
}

// validateDNAT44Config validates VPP destination-NAT44 configuration.
func (d *DNAT44Descriptor) validateDNAT44Config(dnat *nat.DNat44) error {
	if dnat.Label == "" {
		return ErrDNAT44WithEmptyLabel
	}
	return nil
}