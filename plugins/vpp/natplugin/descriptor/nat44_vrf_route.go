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
	NAT44VrfRouteDescriptorName = "vpp-nat44-vrf-route"
)

// NAT44VrfRouteDescriptor teaches KVScheduler how to configure routes for
// VPP NAT44.
type NAT44VrfRouteDescriptor struct {
	log        logging.Logger
	natHandler vppcalls.NatVppAPI
}

// NewNAT44VrfRouteDescriptor creates a new instance of the Nat44VrfRoute descriptor.
func NewNAT44VrfRouteDescriptor(natHandler vppcalls.NatVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &NAT44VrfRouteDescriptor{
		natHandler: natHandler,
		log:        log.NewLogger("nat44-vrf-route-descriptor"),
	}
	typedDescr := &adapter.NAT44VrfRouteDescriptor{
		Name:                 NAT44VrfRouteDescriptorName,
		NBKeyPrefix:          nat.ModelNat44VrfRoute.KeyPrefix(),
		ValueTypeName:        nat.ModelNat44VrfRoute.ProtoName(),
		KeySelector:          nat.ModelNat44VrfRoute.IsKeyValid,
		KeyLabel:             nat.ModelNat44VrfRoute.StripKeyPrefix,
		Create:               ctx.Create,
		Delete:               ctx.Delete,
		Dependencies:         ctx.Dependencies,
		RetrieveDependencies: []string{NAT44VrfDescriptorName},
	}
	return adapter.NewNAT44VrfRouteDescriptor(typedDescr)
}

// Create sets new route to NAT44 vrf table
func (d *NAT44VrfRouteDescriptor) Create(key string, vrfRoute *nat.Nat44VrfRoute) (metadata interface{}, err error) {
	if !d.natHandler.WithLegacyStartupConf() {
		err = d.natHandler.AddNat44VrfRoute(vrfRoute.SrcVrfId, vrfRoute.DestVrfId)
		if err != nil {
			d.log.Error(err)
			return nil, err
		}
	}
	return
}

// Delete removes route from NAT44 vrf table
func (d *NAT44VrfRouteDescriptor) Delete(key string, vrfRoute *nat.Nat44VrfRoute, metadata interface{}) error {
	err := d.natHandler.DelNat44VrfRoute(vrfRoute.SrcVrfId, vrfRoute.DestVrfId)
	if err != nil {
		d.log.Error(err)
		return err
	}
	return nil
}

func (d *NAT44VrfRouteDescriptor) Dependencies(key string, natVrfTable *nat.Nat44VrfRoute) (deps []kvs.Dependency) {

	deps = append(deps, kvs.Dependency{
		Label: nat44VrfTableDep,
		Key:   nat.Nat44VrfTableKey(natVrfTable.SrcVrfId),
	})
	return deps
}
