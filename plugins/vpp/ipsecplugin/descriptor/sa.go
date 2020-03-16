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
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

const (
	// SADescriptorName is the name of the descriptor for VPP security associations.
	SADescriptorName = "vpp-ipsec-sa"
)

// A list of non-retriable errors:
var (
	// ErrSAWithoutIndex is returned when VPP security association was defined
	// without index.
	ErrSAWithoutIndex = errors.New("VPP security association defined without index")

	// ErrSAInvalidIndex is returned when VPP security association was defined
	// with non-numerical index.
	ErrSAInvalidIndex = errors.New("VPP security association defined with invalid index")
)

// IPSecSADescriptor teaches KVScheduler how to configure VPP IPSec security associations.
type IPSecSADescriptor struct {
	// dependencies
	log          logging.Logger
	ipSecHandler vppcalls.IPSecVppAPI
}

// NewIPSecSADescriptor creates a new instance of the IPSec SA descriptor.
func NewIPSecSADescriptor(ipSecHandler vppcalls.IPSecVppAPI, log logging.PluginLogger) *IPSecSADescriptor {
	return &IPSecSADescriptor{
		ipSecHandler: ipSecHandler,
		log:          log.NewLogger("ipsec-sa-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *IPSecSADescriptor) GetDescriptor() *adapter.SADescriptor {
	return &adapter.SADescriptor{
		Name:            SADescriptorName,
		NBKeyPrefix:     ipsec.ModelSecurityAssociation.KeyPrefix(),
		ValueTypeName:   ipsec.ModelSecurityAssociation.ProtoName(),
		KeySelector:     ipsec.ModelSecurityAssociation.IsKeyValid,
		KeyLabel:        ipsec.ModelSecurityAssociation.StripKeyPrefix,
		ValueComparator: d.EquivalentIPSecSAs,
		Create:          d.Create,
		Delete:          d.Delete,
		Retrieve:        d.Retrieve,
	}
}

// EquivalentIPSecSAs is case-insensitive comparison function for
// ipsec.SecurityAssociation
func (d *IPSecSADescriptor) EquivalentIPSecSAs(key string, oldSA, newSA *ipsec.SecurityAssociation) bool {
	// compare base fields
	return oldSA.Spi == newSA.Spi &&
		oldSA.Protocol == newSA.Protocol &&
		oldSA.CryptoAlg == newSA.CryptoAlg &&
		oldSA.CryptoKey == newSA.CryptoKey &&
		oldSA.IntegAlg == newSA.IntegAlg &&
		oldSA.IntegKey == newSA.IntegKey &&
		oldSA.UseEsn == newSA.UseEsn &&
		oldSA.UseAntiReplay == newSA.UseAntiReplay &&
		oldSA.TunnelSrcAddr == newSA.TunnelSrcAddr &&
		oldSA.TunnelDstAddr == newSA.TunnelDstAddr &&
		oldSA.EnableUdpEncap == newSA.EnableUdpEncap
}

// Create adds a new security association pair.
func (d *IPSecSADescriptor) Create(key string, sa *ipsec.SecurityAssociation) (metadata interface{}, err error) {
	// add security association
	err = d.ipSecHandler.AddSA(sa)
	if err != nil {
		d.log.Error(err)
	}

	return nil, err
}

// Delete removes VPP security association.
func (d *IPSecSADescriptor) Delete(key string, sa *ipsec.SecurityAssociation, metadata interface{}) error {
	err := d.ipSecHandler.DeleteSA(sa)
	if err != nil {
		d.log.Error(err)
	}
	return err
}

// Retrieve returns all configured VPP security associations.
func (d *IPSecSADescriptor) Retrieve(correlate []adapter.SAKVWithMetadata) (dump []adapter.SAKVWithMetadata, err error) {
	// dump security associations
	sas, err := d.ipSecHandler.DumpIPSecSA()
	if err != nil {
		d.log.Error(err)
		return dump, err
	}
	for _, sa := range sas {
		dump = append(dump, adapter.SAKVWithMetadata{
			Key:      ipsec.SAKey(sa.Sa.Index),
			Value:    sa.Sa,
			Metadata: sa.Meta,
			Origin:   kvs.FromNB,
		})
	}

	return dump, nil
}
