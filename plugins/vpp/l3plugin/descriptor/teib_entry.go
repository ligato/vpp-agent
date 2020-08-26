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

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/utils/addrs"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	// TeibDescriptorName is the name of the descriptor.
	TeibDescriptorName = "vpp-teib"

	// dependency labels
	teibEntryInterfaceDep = "interface-exists"
	teibEntryVrfTableDep  = "vrf-table-exists"
)

// TeibDescriptor instructs KVScheduler how to configure VPP TEIB entries.
type TeibDescriptor struct {
	log         logging.Logger
	teibHandler vppcalls.TeibVppAPI
	scheduler   kvs.KVScheduler
}

// NewTeibDescriptor creates a new instance of the TeibDescriptor.
func NewTeibDescriptor(scheduler kvs.KVScheduler,
	teibHandler vppcalls.TeibVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &TeibDescriptor{
		scheduler:   scheduler,
		teibHandler: teibHandler,
		log:         log.NewLogger("teib-descriptor"),
	}

	typedDescr := &adapter.TeibEntryDescriptor{
		Name:                 TeibDescriptorName,
		NBKeyPrefix:          l3.ModelTeib.KeyPrefix(),
		ValueTypeName:        l3.ModelTeib.ProtoName(),
		KeySelector:          l3.ModelTeib.IsKeyValid,
		KeyLabel:             l3.ModelTeib.StripKeyPrefix,
		Validate:             ctx.Validate,
		Create:               ctx.Create,
		Delete:               ctx.Delete,
		Retrieve:             ctx.Retrieve,
		Dependencies:         ctx.Dependencies,
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName},
	}
	return adapter.NewTeibEntryDescriptor(typedDescr)
}

// Validate validates VPP TEIB entry configuration.
func (d *TeibDescriptor) Validate(key string, entry *l3.TeibEntry) (err error) {
	if entry.Interface == "" {
		return kvs.NewInvalidValueError(fmt.Errorf("no interface defined"), "interface")
	}
	if entry.PeerAddr == "" {
		return kvs.NewInvalidValueError(fmt.Errorf("no peer address defined"), "peer_addr")
	}
	if entry.NextHopAddr == "" {
		return kvs.NewInvalidValueError(fmt.Errorf("no next hop address defined"), "next_hop_addr")
	}
	return nil
}

// Create adds a VPP TEIB entry.
func (d *TeibDescriptor) Create(key string, entry *l3.TeibEntry) (interface{}, error) {
	return nil, d.teibHandler.VppAddTeibEntry(context.Background(), entry)
}

// Delete removes a VPP TEIB entry.
func (d *TeibDescriptor) Delete(key string, entry *l3.TeibEntry, metadata interface{}) error {
	return d.teibHandler.VppDelTeibEntry(context.Background(), entry)
}

// Retrieve returns all TEIB entries.
func (d *TeibDescriptor) Retrieve(correlate []adapter.TeibEntryKVWithMetadata) (
	retrieved []adapter.TeibEntryKVWithMetadata, err error,
) {
	entries, err := d.teibHandler.DumpTeib()
	if errors.Is(err, vppcalls.ErrTeibUnsupported) {
		d.log.Debug("DumpTeib failed:", err)
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		retrieved = append(retrieved, adapter.TeibEntryKVWithMetadata{
			Key:    models.Key(entry),
			Value:  entry,
			Origin: kvs.UnknownOrigin,
		})
	}
	return
}

// Dependencies lists dependencies for a VPP TEIB entry.
func (d *TeibDescriptor) Dependencies(key string, entry *l3.TeibEntry) (deps []kvs.Dependency) {

	// the referenced interface must exist
	deps = append(deps, kvs.Dependency{
		Label: teibEntryInterfaceDep,
		Key:   interfaces.InterfaceKey(entry.Interface),
	})

	// non-zero VRF must exists
	if entry.VrfId != 0 {
		var protocol l3.VrfTable_Protocol
		_, isIPv6, _ := addrs.ParseIPWithPrefix(entry.NextHopAddr)
		if isIPv6 {
			protocol = l3.VrfTable_IPV6
		}
		deps = append(deps, kvs.Dependency{
			Label: teibEntryVrfTableDep,
			Key:   l3.VrfTableKey(entry.VrfId, protocol),
		})
	}
	return deps
}
