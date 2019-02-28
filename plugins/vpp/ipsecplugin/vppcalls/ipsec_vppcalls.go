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
	"github.com/ligato/cn-infra/logging"
	ipsec "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
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
	// DelSPD deletes SPD from VPP via binary API
	DeleteSPD(spdID uint32) error
	// InterfaceAddSPD adds SPD interface assignment to VPP via binary API
	AddSPDInterface(spdID uint32, iface *ipsec.SecurityPolicyDatabase_Interface) error
	// InterfaceDelSPD deletes SPD interface assignment from VPP via binary API
	DeleteSPDInterface(spdID uint32, iface *ipsec.SecurityPolicyDatabase_Interface) error
	// AddSPDEntry adds SPD policy entry to VPP via binary API
	AddSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabase_PolicyEntry) error
	// DelSPDEntry deletes SPD policy entry from VPP via binary API
	DeleteSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabase_PolicyEntry) error
	// AddSAEntry adds SA to VPP via binary API
	AddSA(sa *ipsec.SecurityAssociation) error
	// DelSAEntry deletes SA from VPP via binary API
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

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel, ifaceidx.IfaceMetadataIndex, logging.Logger) IPSecVppAPI
}

func CompatibleIPSecVppHandler(
	ch govppapi.Channel, idx ifaceidx.IfaceMetadataIndex, log logging.Logger,
) IPSecVppAPI {
	if len(Versions) == 0 {
		// ipsecplugin is not loaded
		return nil
	}
	for ver, h := range Versions {
		log.Debugf("checking compatibility with %s", ver)
		if err := ch.CheckCompatiblity(h.Msgs...); err != nil {
			continue
		}
		log.Debug("found compatible version:", ver)
		return h.New(ch, idx, log)
	}
	panic("no compatible version available")
}
