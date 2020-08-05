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
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

const (
	// SPDescriptorName is the name of the descriptor for configuring VPP IPSec security policies.
	SPDescriptorName = "vpp-ipsec-sp"

	// dependency labels
	spdDep = "spd-exists"
	saDep  = "sa-exists"
)

// IPSecSPDescriptor teaches KVScheduler how to configure VPP IPSec Security Policies.
type IPSecSPDescriptor struct {
	// dependencies
	log          logging.Logger
	ipSecHandler vppcalls.IPSecVppAPI
}

// NewIPSecSPDescriptor creates a new instance of the SP descriptor.
func NewIPSecSPDescriptor(ipSecHandler vppcalls.IPSecVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &IPSecSPDescriptor{
		log:          log.NewLogger("ipsec-sp-descriptor"),
		ipSecHandler: ipSecHandler,
	}
	typedDescr := &adapter.SPDescriptor{
		Name:            SPDescriptorName,
		NBKeyPrefix:     ipsec.ModelSecurityPolicy.KeyPrefix(),
		ValueTypeName:   ipsec.ModelSecurityPolicy.ProtoName(),
		KeySelector:     ipsec.ModelSecurityPolicy.IsKeyValid,
		KeyLabel:        ipsec.ModelSecurityPolicy.StripKeyPrefix,
		ValueComparator: ctx.EquivalentSPs,
		Create:          ctx.Create,
		Delete:          ctx.Delete,
		Dependencies:    ctx.Dependencies,
		Retrieve:        ctx.Retrieve,
	}
	return adapter.NewSPDescriptor(typedDescr)
}

// EquivalentSPs compares two SPs for equivalency.
func (d *IPSecSPDescriptor) EquivalentSPs(key string, oldValue, newValue *ipsec.SecurityPolicy) bool {
	if oldValue.GetPriority() != newValue.GetPriority() ||
		oldValue.GetProtocol() != newValue.GetProtocol() ||
		oldValue.GetAction() != newValue.GetAction() {
		return false
	}

	normalizedPortRange := func(start, stop uint32) (uint32, uint32) {
		if start == 0 && stop == 0 {
			return 0, uint32(^uint16(0))
		}
		return start, stop
	}
	prevLPStart, prevLPStop := normalizedPortRange(oldValue.GetLocalPortStart(), oldValue.GetLocalPortStop())
	newLPStart, newLPStop := normalizedPortRange(newValue.GetLocalPortStart(), newValue.GetLocalPortStop())
	if prevLPStart != newLPStart || prevLPStop != newLPStop {
		return false
	}
	prevRPStart, prevRPStop := normalizedPortRange(oldValue.GetRemotePortStart(), oldValue.GetRemotePortStop())
	newRPStart, newRPStop := normalizedPortRange(newValue.GetRemotePortStart(), newValue.GetRemotePortStop())
	if prevRPStart != newRPStart || prevRPStop != newRPStop {
		return false
	}
	return true
}

// Create puts policy into security policy database.
func (d *IPSecSPDescriptor) Create(key string, policy *ipsec.SecurityPolicy) (metadata interface{}, err error) {
	err = d.ipSecHandler.AddSP(policy)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	return nil, nil
}

// Delete removes policy from security policy database.
func (d *IPSecSPDescriptor) Delete(key string, policy *ipsec.SecurityPolicy, metadata interface{}) (err error) {
	err = d.ipSecHandler.DeleteSP(policy)
	if err != nil {
		d.log.Error(err)
		return err
	}
	return nil
}

// Dependencies lists the associated security association and SPD as the dependencies of the policy.
func (d *IPSecSPDescriptor) Dependencies(key string, value *ipsec.SecurityPolicy) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: spdDep,
			Key:   ipsec.SPDKey(value.SpdIndex),
		},
		{
			Label: saDep,
			Key:   ipsec.SAKey(value.SaIndex),
		},
	}
}

// Retrieve returns all configured VPP IPSec Security Policies.
func (d *IPSecSPDescriptor) Retrieve(correlate []adapter.SPKVWithMetadata) (dump []adapter.SPKVWithMetadata, err error) {
	sps, err := d.ipSecHandler.DumpIPSecSP()
	if err != nil {
		d.log.Error(err)
		return dump, err
	}
	for _, sp := range sps {
		dump = append(dump, adapter.SPKVWithMetadata{
			Key:    models.Key(sp),
			Value:  sp,
			Origin: kvs.FromNB,
		})
	}
	return dump, nil
}
