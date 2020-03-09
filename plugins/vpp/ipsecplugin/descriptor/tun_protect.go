// Copyright (c) 2020 Cisco and/or its affiliates.
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
	"github.com/pkg/errors"

	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

const (
	// TunProtectDescriptorName is the name of the descriptor for VPP tunnel protection.
	TunProtectDescriptorName = "vpp-tun-protect"

	// dependency labels
	ipsecSADep = "ipsec-sa-exists"
)

// A list of non-retriable errors:
var (
	// ErrTunProtectNoInterface is returned when VPP tunnel protection was defined without an interface.
	ErrTunProtectNoInterface = errors.New("VPP tunnel protection defined without interface")
	// ErrTunProtectNoSaOut is returned when VPP tunnel protection was defined without outbound SAs.
	ErrTunProtectNoSaOut = errors.New("VPP tunnel protection defined without outbound SAs")
	// ErrTunProtectNoSaIn is returned when VPP tunnel protection was defined without inbound SAs.
	ErrTunProtectNoSaIn = errors.New("VPP tunnel protection defined without inbound SAs")
	// ErrTunProtectUpdateIfMismatch is returned if old and new tunnel interface names are not matching by update operation.
	ErrTunProtectUpdateIfMismatch = errors.New("old/new tunnel interface mismatch")
)

// TunnelProtectDescriptor teaches KVScheduler how to configure VPP IPSec tunnel protections.
type TunnelProtectDescriptor struct {
	// dependencies
	log          logging.Logger
	ipSecHandler vppcalls.IPSecVppAPI
}

// NewTunnelProtectDescriptor creates a new instance of the IPSec tunnel protect descriptor.
func NewTunnelProtectDescriptor(ipSecHandler vppcalls.IPSecVppAPI, log logging.PluginLogger) *TunnelProtectDescriptor {
	return &TunnelProtectDescriptor{
		ipSecHandler: ipSecHandler,
		log:          log.NewLogger("tun-protect-descriptor"),
	}
}

// GetDescriptor returns a new tunnel protect descriptor suitable for registration with the KVScheduler.
func (d *TunnelProtectDescriptor) GetDescriptor() *adapter.TunProtectDescriptor {
	return &adapter.TunProtectDescriptor{
		Name:          TunProtectDescriptorName,
		NBKeyPrefix:   ipsec.ModelTunnelProtection.KeyPrefix(),
		ValueTypeName: ipsec.ModelTunnelProtection.ProtoName(),
		KeySelector:   ipsec.ModelTunnelProtection.IsKeyValid,
		KeyLabel:      ipsec.ModelTunnelProtection.StripKeyPrefix,
		Validate:      d.Validate,
		Create:        d.Create,
		Update:        d.Update,
		Delete:        d.Delete,
		Retrieve:      d.Retrieve,
		Dependencies:  d.Dependencies,
	}
}

// Validate validates VPP tunnel protect configuration.
func (d *TunnelProtectDescriptor) Validate(key string, tp *ipsec.TunnelProtection) error {
	if tp.Interface == "" {
		return kvs.NewInvalidValueError(ErrTunProtectNoInterface, "interface")
	}
	if len(tp.SaOut) == 0 {
		return kvs.NewInvalidValueError(ErrTunProtectNoSaOut, "sa_out")
	}
	if len(tp.SaIn) == 0 {
		return kvs.NewInvalidValueError(ErrTunProtectNoSaIn, "sa_in")
	}
	return nil
}

// Create adds a new IPSec tunnel protection.
func (d *TunnelProtectDescriptor) Create(key string, tp *ipsec.TunnelProtection) (metadata interface{}, err error) {
	return nil, d.ipSecHandler.AddTunnelProtection(tp)
}

// Update updates an existing IPSec tunnel protection.
func (d *TunnelProtectDescriptor) Update(key string, oldTp, newTp *ipsec.TunnelProtection, oldMeta interface{}) (
	metadata interface{}, err error) {
	if oldTp.Interface != newTp.Interface {
		return nil, ErrTunProtectUpdateIfMismatch
	}
	return nil, d.ipSecHandler.AddTunnelProtection(newTp)
}

// Delete removes an IPSec tunnel protection.
func (d *TunnelProtectDescriptor) Delete(key string, tp *ipsec.TunnelProtection, metadata interface{}) error {
	return d.ipSecHandler.DeleteTunnelProtection(tp)
}

// Retrieve returns all configured IPSec tunnel protections.
func (d *TunnelProtectDescriptor) Retrieve(correlate []adapter.TunProtectKVWithMetadata) (dump []adapter.TunProtectKVWithMetadata, err error) {
	tps, err := d.ipSecHandler.DumpTunnelProtections()
	for _, tp := range tps {
		dump = append(dump, adapter.TunProtectKVWithMetadata{
			Key:    models.Key(tp),
			Value:  tp,
			Origin: kvs.FromNB,
		})
	}
	return
}

// Dependencies lists the interface and SAs as the dependencies for the binding.
func (d *TunnelProtectDescriptor) Dependencies(key string, value *ipsec.TunnelProtection) []kvs.Dependency {
	deps := []kvs.Dependency{
		{
			Label: interfaceDep,
			Key:   interfaces.InterfaceKey(value.Interface),
		},
	}
	for _, sa := range value.SaOut {
		deps = append(deps, kvs.Dependency{
			Label: ipsecSADep,
			Key:   ipsec.SAKey(sa),
		})
	}
	for _, sa := range value.SaIn {
		deps = append(deps, kvs.Dependency{
			Label: ipsecSADep,
			Key:   ipsec.SAKey(sa),
		})
	}
	return deps
}
