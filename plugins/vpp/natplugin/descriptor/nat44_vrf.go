// Copyright (c) 2023 Cisco and/or its affiliates.
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
	"go.ligato.io/cn-infra/v2/logging"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

const (
	NAT44VrfDescriptorName = "vpp-nat44-vrf"

	//Dependencies
	nat44VrfTableDep = "vrf-table-exists"
)

// NAT44VrfDescriptor teaches KVScheduler how to configure vrf tables for
// VPP NAT44.

type NAT44VrfDescriptor struct {
	log        logging.Logger
	natHandler vppcalls.NatVppAPI
}

// NewNAT44VrfDescriptor creates a new instance of the NAT44VrfDescriptor descriptor.
func NewNAT44VrfDescriptor(natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &NAT44VrfDescriptor{
		natHandler: natHandler,
		log:        log.NewLogger("nat44-vrf-descriptor"),
	}
	typedDescr := &adapter.NAT44VrfDescriptor{
		Name:          NAT44VrfDescriptorName,
		NBKeyPrefix:   nat.ModelNat44VrfTable.KeyPrefix(),
		ValueTypeName: nat.ModelNat44VrfTable.ProtoName(),
		KeySelector:   nat.ModelNat44VrfTable.IsKeyValid,
		KeyLabel:      nat.ModelNat44VrfTable.StripKeyPrefix,
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Retrieve:      ctx.Retrieve,
	}
	return adapter.NewNAT44VrfDescriptor(typedDescr)
}

// Create adds vrf table to NAT44 configuration.
func (d *NAT44VrfDescriptor) Create(key string, vrfTable *nat.Nat44VrfTable) (metadata interface{}, err error) {
	if !d.natHandler.WithLegacyStartupConf() {
		err = d.natHandler.AddNat44VrfTable(vrfTable.SrcVrfId)
		if err != nil {
			d.log.Error(err)
			return nil, err
		}
	}
	return
}

// Delete removes vrf table from NAT44 configuration.
func (d *NAT44VrfDescriptor) Delete(key string, vrfTable *nat.Nat44VrfTable, metadata interface{}) error {
	err := d.natHandler.DelNat44VrfTable(vrfTable.SrcVrfId)

	if err != nil {
		d.log.Error(err)
		return err
	}
	return nil
}

// Retrieve returns the whole list of NAT44VrfTables.
func (d *NAT44VrfDescriptor) Retrieve(correlate []adapter.NAT44VrfKVWithMetadata) (
	retrieved []adapter.NAT44VrfKVWithMetadata, err error) {

	natTables, err := d.natHandler.Nat44VrfTablesDump()
	if err != nil {
		return nil, err
	}
	for _, natEntry := range natTables {
		retrieved = append(retrieved, adapter.NAT44VrfKVWithMetadata{
			Key:    nat.Nat44VrfTableKey(natEntry.SrcVrfId),
			Value:  natEntry,
			Origin: kvs.FromSB,
		})
	}
	return
}
