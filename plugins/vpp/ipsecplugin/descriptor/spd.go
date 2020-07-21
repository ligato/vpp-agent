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
	"errors"

	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	vppIfDescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

const (
	// IPSecSPDDescriptorName is the name of the descriptor for VPP IPSec SPD
	// configurations.
	IPSecSPDDescriptorName = "vpp-ipsec-spd"
)

// A list of non-retriable errors:
var (
	// ErrDeprecatedSPDPolicies is returned when the deprecated SecurityPolicyDatabase.PolicyEntries is used.
	ErrDeprecatedSPDPolicies = errors.New(
		"it is deprecated and no longer supported to define SPs inside SPD model (use SecurityPolicy model instead)")
)

// IPSecSPDDescriptor teaches KVScheduler how to configure IPSec SPD in VPP.
type IPSecSPDDescriptor struct {
	// dependencies
	log          logging.Logger
	ipSecHandler vppcalls.IPSecVppAPI
}

// NewIPSecSPDDescriptor creates a new instance of the IPSec SPD descriptor.
func NewIPSecSPDDescriptor(ipSecHandler vppcalls.IPSecVppAPI, log logging.PluginLogger) *IPSecSPDDescriptor {
	return &IPSecSPDDescriptor{
		ipSecHandler: ipSecHandler,
		log:          log.NewLogger("ipsec-spd-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *IPSecSPDDescriptor) GetDescriptor() *adapter.SPDDescriptor {
	return &adapter.SPDDescriptor{
		Name:                 IPSecSPDDescriptorName,
		NBKeyPrefix:          ipsec.ModelSecurityPolicyDatabase.KeyPrefix(),
		ValueTypeName:        ipsec.ModelSecurityPolicyDatabase.ProtoName(),
		KeySelector:          ipsec.ModelSecurityPolicyDatabase.IsKeyValid,
		KeyLabel:             ipsec.ModelSecurityPolicyDatabase.StripKeyPrefix,
		ValueComparator:      d.EquivalentIPSecSPDs,
		Validate:             d.Validate,
		Create:               d.Create,
		Delete:               d.Delete,
		Retrieve:             d.Retrieve,
		DerivedValues:        d.DerivedValues,
		RetrieveDependencies: []string{vppIfDescriptor.InterfaceDescriptorName},
	}
}

// EquivalentIPSecSPDs always returns true because all non-key attributes are derived out.
func (d *IPSecSPDDescriptor) EquivalentIPSecSPDs(key string, oldSPD, newSPD *ipsec.SecurityPolicyDatabase) bool {
	return true
}

// Validate validates IPSec SPD configuration.
func (d *IPSecSPDDescriptor) Validate(key string, spd *ipsec.SecurityPolicyDatabase) (err error) {
	if len(spd.GetPolicyEntries()) != 0 {
		return ErrDeprecatedSPDPolicies
	}
	return nil
}

// Create adds a new IPSec security policy database.
func (d *IPSecSPDDescriptor) Create(key string, spd *ipsec.SecurityPolicyDatabase) (metadata interface{}, err error) {
	// create a new SPD with index
	err = d.ipSecHandler.AddSPD(spd.GetIndex())
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	return nil, nil
}

// Delete removes VPP IPSec security policy database.
func (d *IPSecSPDDescriptor) Delete(key string, spd *ipsec.SecurityPolicyDatabase, metadata interface{}) error {
	err := d.ipSecHandler.DeleteSPD(spd.GetIndex())
	if err != nil {
		d.log.Error(err)
	}
	return err
}

// Retrieve returns all configured VPP security policy databases.
func (d *IPSecSPDDescriptor) Retrieve(correlate []adapter.SPDKVWithMetadata) (dump []adapter.SPDKVWithMetadata, err error) {
	nbCfg := map[uint32]*ipsec.SecurityPolicyDatabase{}
	for _, spd := range correlate {
		nbCfg[spd.Value.GetIndex()] = spd.Value
	}

	// dump security policy associations
	spds, err := d.ipSecHandler.DumpIPSecSPD()
	if err != nil {
		d.log.Error(err)
		return dump, err
	}
	for _, spd := range spds {
		// Correlate interface assignments which are not properly dumped (bug in ipsec_spd_interface_dump)
		spd.Interfaces = nbCfg[spd.GetIndex()].GetInterfaces()
		dump = append(dump, adapter.SPDKVWithMetadata{
			Key:      ipsec.SPDKey(spd.Index),
			Value:    spd,
			Origin:   kvs.FromNB,
		})
	}
	return dump, nil
}

// DerivedValues derives ipsec.SecurityPolicyDatabase_Interface for every interface assigned to the SPD.
func (d *IPSecSPDDescriptor) DerivedValues(key string, spd *ipsec.SecurityPolicyDatabase) (derValues []kvs.KeyValuePair) {
	// SPD interfaces
	for _, spdIface := range spd.Interfaces {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   ipsec.SPDInterfaceKey(spd.Index, spdIface.Name),
			Value: spdIface,
		})
	}

	return derValues
}