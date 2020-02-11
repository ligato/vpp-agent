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

package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

// IPSecSaDetails holds security association with VPP metadata
type IPSecSaDetails struct {
	Sa   *ipsec.SecurityAssociation
	Meta *IPSecSaMeta
}

// IPSecSaMeta contains all VPP-specific metadata
type IPSecSaMeta struct {
	SaID           uint32
	Interface      string
	IfIdx          uint32
	CryptoKeyLen   uint8
	IntegKeyLen    uint8
	Salt           uint32
	SeqOutbound    uint64
	LastSeqInbound uint64
	ReplayWindow   uint64
	TotalDataSize  uint64
}

// IPSecSpdDetails represents IPSec policy databases with particular metadata
type IPSecSpdDetails struct {
	Spd         *ipsec.SecurityPolicyDatabase
	PolicyMeta  map[string]*SpdMeta // SA index name is a key
	NumPolicies uint32
}

// SpdMeta hold VPP-specific data related to SPD
type SpdMeta struct {
	SaID    uint32
	Policy  uint8
	Bytes   uint64
	Packets uint64
}

// IPSecVppAPI provides methods for creating and managing of a IPsec configuration
type IPSecVppAPI interface {
	IPSecVPPRead

	// AddSPD adds SPD to VPP via binary API
	AddSPD(spdID uint32) error
	// DeleteSPD deletes SPD from VPP via binary API
	DeleteSPD(spdID uint32) error
	// AddSPDInterface adds SPD interface assignment to VPP via binary API
	AddSPDInterface(spdID uint32, iface *ipsec.SecurityPolicyDatabase_Interface) error
	// DeleteSPDInterface deletes SPD interface assignment from VPP via binary API
	DeleteSPDInterface(spdID uint32, iface *ipsec.SecurityPolicyDatabase_Interface) error
	// AddSPDEntry adds SPD policy entry to VPP via binary API
	AddSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabase_PolicyEntry) error
	// DeleteSPDEntry deletes SPD policy entry from VPP via binary API
	DeleteSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabase_PolicyEntry) error
	// AddSA adds SA to VPP via binary API
	AddSA(sa *ipsec.SecurityAssociation) error
	// DeleteSA deletes SA from VPP via binary API
	DeleteSA(sa *ipsec.SecurityAssociation) error
}

// IPSecVPPRead provides read methods for IPSec
type IPSecVPPRead interface {
	// DumpIPSecSPD returns a list of IPSec security policy databases
	DumpIPSecSPD() (spdList []*IPSecSpdDetails, err error)
	// DumpIPSecSA returns a list of configured security associations
	DumpIPSecSA() (saList []*IPSecSaDetails, err error)
	// DumpIPSecSAWithIndex returns a security association with provided index
	DumpIPSecSAWithIndex(saID uint32) (saList []*IPSecSaDetails, err error)
}

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "ipsec",
	HandlerAPI: (*IPSecVppAPI)(nil),
})

type NewHandlerFunc func(ch govppapi.Channel, ifDdx ifaceidx.IfaceMetadataIndex, log logging.Logger) IPSecVppAPI

func AddHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return ch.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return h(ch, a[0].(ifaceidx.IfaceMetadataIndex), a[1].(logging.Logger))
		},
	})
}

func CompatibleIPSecVppHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) IPSecVppAPI {
	if v := Handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, ifIdx, log).(IPSecVppAPI)
	}
	return nil
}
