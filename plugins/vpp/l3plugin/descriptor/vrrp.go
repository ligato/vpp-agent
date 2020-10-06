//  Copyright (c) 2020 Cisco and/or its affiliates.
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
	// VrrpDescriptorName is the name of the descriptor.
	VrrpDescriptorName = "vrrp"

	// dependency labels
	vrrpEntryInterfaceDep = "interface-exists"
)

// A list of validation errors
var (
	ErrMissingInterface = errors.New("missing interface")
)

// VrrpDescriptor teaches KVScheduler how to configure VPP VRRPs.
type VrrpDescriptor struct {
	log         logging.Logger
	vrrpHandler vppcalls.VrrpVppAPI
}

// NewVrrpDescriptor creates a new instance of the VrrpDescriptor.
func NewVrrpDescriptor(vrrpHandler vppcalls.VrrpVppAPI,
	log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &VrrpDescriptor{
		log:         log,
		vrrpHandler: vrrpHandler,
	}

	typedDescr := &adapter.VRRPEntryDescriptor{
		Name:                 VrrpDescriptorName,
		NBKeyPrefix:          l3.ModelVRRPEntry.KeyPrefix(),
		ValueTypeName:        l3.ModelVRRPEntry.ProtoName(),
		KeySelector:          l3.ModelVRRPEntry.IsKeyValid,
		KeyLabel:             l3.ModelVRRPEntry.StripKeyPrefix,
		Create:               ctx.Create,
		Delete:               ctx.Delete,
		Validate:             ctx.Validate,
		Retrieve:             ctx.Retrieve,
		Dependencies:         ctx.Dependencies,
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName},
	}

	return adapter.NewVRRPEntryDescriptor(typedDescr)
}

// Validate returns error if given VRRP is not valid.
func (d *VrrpDescriptor) Validate(key string, vrrp *l3.VRRPEntry) error {
	if vrrp.Interface == "" {
		return kvs.NewInvalidValueError(ErrMissingInterface, "interface")
	}
	return nil
}

// Create adds VPP VRRP entry.
func (d *VrrpDescriptor) Create(key string, vrrp *l3.VRRPEntry) (interface{}, error) {
	if err := d.vrrpHandler.VppAddVrrp(vrrp); err != nil {
		return nil, err
	}
	if vrrp.Enabled {
		if err := d.vrrpHandler.VppStartVrrp(vrrp); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// Delete removes VPP VRRP entry.
func (d *VrrpDescriptor) Delete(key string, vrrp *l3.VRRPEntry, metadata interface{}) error {
	if err := d.vrrpHandler.VppDelVrrp(vrrp); err != nil {
		return err
	}
	return nil
}

// Dependencies lists dependencies for a VPP VRRP entry.
func (d *VrrpDescriptor) Dependencies(key string, vrrp *l3.VRRPEntry) (deps []kvs.Dependency) {
	// the outgoing interface must exist
	if vrrp.Interface != "" {
		deps = append(deps, kvs.Dependency{
			Label: vrrpEntryInterfaceDep,
			Key:   interfaces.InterfaceKey(vrrp.Interface),
		})
	}
	return deps
}

// Retrieve returns all configured VPP VRRP entries.
func (d *VrrpDescriptor) Retrieve(correlate []adapter.VRRPEntryKVWithMetadata) (
	retrieved []adapter.VRRPEntryKVWithMetadata, err error,
) {
	entries, err := d.vrrpHandler.DumpVrrpEntries()

	for _, entry := range entries {

		retrieved = append(retrieved, adapter.VRRPEntryKVWithMetadata{
			Key:    l3.VrrpEntryKey(entry.Vrrp.Interface, entry.Vrrp.Addrs),
			Value:  entry.Vrrp,
			Origin: kvs.UnknownOrigin,
		})
	}

	return retrieved, nil
}
