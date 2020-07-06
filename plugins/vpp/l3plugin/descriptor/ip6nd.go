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
	"context"
	"fmt"

	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	// IP6ndDescriptorName is the name of the descriptor.
	IP6ndDescriptorName = "vpp-ip6nd"

	// dependency labels
	ip6ndEntryInterfaceDep = "interface-exists"
)

// IP6ndDescriptor instructs KVScheduler how to configure VPP TEIB entries.
type IP6ndDescriptor struct {
	log       logging.Logger
	handler   vppcalls.IP6ndVppAPI
	scheduler kvs.KVScheduler
}

// NewIP6ndDescriptor creates a new instance of the IP6ndDescriptor.
func NewIP6ndDescriptor(scheduler kvs.KVScheduler,
	handler vppcalls.IP6ndVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &IP6ndDescriptor{
		scheduler: scheduler,
		handler:   handler,
		log:       log.NewLogger("ip6nd-descriptor"),
	}

	typedDescr := &adapter.IP6NDDescriptor{
		Name:                 IP6ndDescriptorName,
		NBKeyPrefix:          l3.ModelIP6ND.KeyPrefix(),
		ValueTypeName:        l3.ModelIP6ND.ProtoName(),
		KeySelector:          l3.ModelIP6ND.IsKeyValid,
		KeyLabel:             l3.ModelIP6ND.StripKeyPrefix,
		Validate:             ctx.Validate,
		Create:               ctx.Create,
		Delete:               ctx.Delete,
		Retrieve:             ctx.Retrieve,
		Dependencies:         ctx.Dependencies,
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName},
	}
	return adapter.NewIP6NDDescriptor(typedDescr)
}

// Validate validates VPP IP6ND entry configuration.
func (d *IP6ndDescriptor) Validate(key string, entry *l3.IP6ND) (err error) {
	if entry.Interface == "" {
		return kvs.NewInvalidValueError(fmt.Errorf("no interface defined"), "interface")
	}
	return nil
}

// Create adds a VPP IP6ND entry.
func (d *IP6ndDescriptor) Create(key string, entry *l3.IP6ND) (interface{}, error) {
	return nil, d.handler.SetIP6ndAutoconfig(context.Background(), entry.Interface, entry.Autoconfig, entry.InstallDefaultRoutes)
}

// Delete removes a VPP IP6ND entry.
func (d *IP6ndDescriptor) Delete(key string, entry *l3.IP6ND, metadata interface{}) error {
	return d.handler.SetIP6ndAutoconfig(context.Background(), entry.Interface, false, false)
}

// Retrieve returns all IP6ND entries.
func (d *IP6ndDescriptor) Retrieve(correlate []adapter.IP6NDKVWithMetadata) (
	retrieved []adapter.IP6NDKVWithMetadata, err error,
) {
	// TODO: implement retrieve
	/*entries, err := d.handler.DumpIP6ND()
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		retrieved = append(retrieved, adapter.IP6NDKVWithMetadata{
			Key:    models.Key(entry),
			Value:  entry,
			Origin: kvs.UnknownOrigin,
		})
	}*/
	return
}

// Dependencies lists dependencies for a VPP IP6ND entry.
func (d *IP6ndDescriptor) Dependencies(key string, entry *l3.IP6ND) (deps []kvs.Dependency) {

	// the referenced interface must exist
	deps = append(deps, kvs.Dependency{
		Label: ip6ndEntryInterfaceDep,
		Key:   interfaces.InterfaceKey(entry.Interface),
	})

	return deps
}
