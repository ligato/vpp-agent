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
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	ifDescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	stn "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/stn"
)

const (
	// STNDescriptorName is the name of the descriptor for VPP STN rules
	STNDescriptorName = "vpp-stn-rules"

	// dependency labels
	stnInterfaceDep = "stn-interface-exists"
)

// A list of non-retriable errors:
var (
	// ErrSTNWithoutInterface is returned when VPP STN rule has undefined interface.
	ErrSTNWithoutInterface = errors.New("VPP STN rule defined without interface")

	// ErrSTNWithoutIPAddress is returned when VPP STN rule has undefined IP address.
	ErrSTNWithoutIPAddress = errors.New("VPP STN rule defined without IP address")
)

// STNDescriptor teaches KVScheduler how to configure VPP STN rules.
type STNDescriptor struct {
	// dependencies
	log        logging.Logger
	stnHandler vppcalls.StnVppAPI
}

// NewSTNDescriptor creates a new instance of the STN descriptor.
func NewSTNDescriptor(stnHandler vppcalls.StnVppAPI, log logging.PluginLogger) *STNDescriptor {
	return &STNDescriptor{
		log:        log.NewLogger("stn-descriptor"),
		stnHandler: stnHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *STNDescriptor) GetDescriptor() *adapter.STNDescriptor {
	return &adapter.STNDescriptor{
		Name:                 STNDescriptorName,
		NBKeyPrefix:          stn.ModelRule.KeyPrefix(),
		ValueTypeName:        stn.ModelRule.ProtoName(),
		KeySelector:          stn.ModelRule.IsKeyValid,
		KeyLabel:             stn.ModelRule.StripKeyPrefix,
		ValueComparator:      d.EquivalentSTNs,
		Validate:             d.Validate,
		Create:               d.Create,
		Delete:               d.Delete,
		Retrieve:             d.Retrieve,
		Dependencies:         d.Dependencies,
		RetrieveDependencies: []string{ifDescriptor.InterfaceDescriptorName},
	}
}

// EquivalentSTNs is case-insensitive comparison function for stn.Rule.
func (d *STNDescriptor) EquivalentSTNs(key string, oldSTN, newSTN *stn.Rule) bool {
	// parameters compared by proto equal
	if proto.Equal(oldSTN, newSTN) {
		return true
	}
	return false
}

// Validate validates VPP STN rule configuration.
func (d *STNDescriptor) Validate(key string, stn *stn.Rule) error {
	if stn.Interface == "" {
		return kvs.NewInvalidValueError(ErrSTNWithoutInterface, "interface")
	}
	if stn.IpAddress == "" {
		return kvs.NewInvalidValueError(ErrSTNWithoutIPAddress, "ip_address")
	}
	return nil
}

// Create adds new STN rule.
func (d *STNDescriptor) Create(key string, stn *stn.Rule) (metadata interface{}, err error) {
	// add STN rule
	err = d.stnHandler.AddSTNRule(stn)
	if err != nil {
		d.log.Error(err)
	}
	return nil, err
}

// Delete removes VPP STN rule.
func (d *STNDescriptor) Delete(key string, stn *stn.Rule, metadata interface{}) error {
	err := d.stnHandler.DeleteSTNRule(stn)
	if err != nil {
		d.log.Error(err)
	}
	return err
}

// Dependencies for STN rule are represented by interface
func (d *STNDescriptor) Dependencies(key string, stn *stn.Rule) (dependencies []kvs.Dependency) {
	dependencies = append(dependencies, kvs.Dependency{
		Label: stnInterfaceDep,
		Key:   interfaces.InterfaceKey(stn.Interface),
	})
	return dependencies
}

// Retrieve returns all configured VPP STN rules.
func (d *STNDescriptor) Retrieve(correlate []adapter.STNKVWithMetadata) (retrieved []adapter.STNKVWithMetadata, err error) {
	stnRules, err := d.stnHandler.DumpSTNRules()
	if err != nil {
		d.log.Error(err)
		return retrieved, err
	}
	for _, stnRule := range stnRules {
		retrieved = append(retrieved, adapter.STNKVWithMetadata{
			Key:    models.Key(stnRule.Rule), //stn.Key(stnRule.Rule.Interface, stnRule.Rule.IpAddress),
			Value:  stnRule.Rule,
			Origin: kvs.FromNB, // all STN rules are configured from NB
		})
	}

	return retrieved, nil
}
