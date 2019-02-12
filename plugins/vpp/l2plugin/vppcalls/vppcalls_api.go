// Copyright (c) 2017 Cisco and/or its affiliates.
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

	l2 "github.com/ligato/vpp-agent/api/models/vpp/l2"
	"github.com/ligato/vpp-agent/pkg/idxvpp"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
)

// BridgeDomainDetails is the wrapper structure for the bridge domain northbound API structure.
// NOTE: Interfaces in BridgeDomains_BridgeDomain is overridden by the local Interfaces member.
type BridgeDomainDetails struct {
	Bd   *l2.BridgeDomain  `json:"bridge_domain"`
	Meta *BridgeDomainMeta `json:"bridge_domain_meta"`
}

// BridgeDomainMeta contains bridge domain interface name/index map
type BridgeDomainMeta struct {
	BdID uint32 `json:"bridge_domain_id"`
}

// L2VppAPI groups L2 Vpp APIs.
type L2VppAPI interface {
	BridgeDomainVppAPI
	FIBVppAPI
	XConnectVppAPI
}

// BridgeDomainVppAPI provides methods for managing bridge domains.
type BridgeDomainVppAPI interface {
	BridgeDomainVppRead

	// AddBridgeDomain adds new bridge domain.
	AddBridgeDomain(bdIdx uint32, bd *l2.BridgeDomain) error
	// DeleteBridgeDomain removes existing bridge domain.
	DeleteBridgeDomain(bdIdx uint32) error
	// AddInterfaceToBridgeDomain puts interface into bridge domain.
	AddInterfaceToBridgeDomain(bdIdx uint32, ifaceCfg *l2.BridgeDomain_Interface) error
	// DeleteInterfaceFromBridgeDomain removes interface from bridge domain.
	DeleteInterfaceFromBridgeDomain(bdIdx uint32, ifaceCfg *l2.BridgeDomain_Interface) error
	// AddArpTerminationTableEntry creates ARP termination entry for bridge domain.
	AddArpTerminationTableEntry(bdID uint32, mac string, ip string) error
	// RemoveArpTerminationTableEntry removes ARP termination entry from bridge domain.
	RemoveArpTerminationTableEntry(bdID uint32, mac string, ip string) error
}

// BridgeDomainVppRead provides read methods for bridge domains.
type BridgeDomainVppRead interface {
	// DumpBridgeDomains dumps VPP bridge domain data into the northbound API data structure
	// map indexed by bridge domain ID.
	DumpBridgeDomains() ([]*BridgeDomainDetails, error)
}

// FIBVppAPI provides methods for managing FIBs.
type FIBVppAPI interface {
	FIBVppRead

	// AddL2FIB creates L2 FIB table entry.
	AddL2FIB(fib *l2.FIBEntry) error
	// DeleteL2FIB removes existing L2 FIB table entry.
	DeleteL2FIB(fib *l2.FIBEntry) error
}

// FIBVppRead provides read methods for FIBs.
type FIBVppRead interface {
	// DumpL2FIBs dumps VPP L2 FIB table entries into the northbound API
	// data structure map indexed by destination MAC address.
	DumpL2FIBs() (map[string]*FibTableDetails, error)
}

// XConnectVppAPI provides methods for managing cross connects.
type XConnectVppAPI interface {
	XConnectVppRead

	// AddL2XConnect creates xConnect between two existing interfaces.
	AddL2XConnect(rxIface, txIface string) error
	// DeleteL2XConnect removes xConnect between two interfaces.
	DeleteL2XConnect(rxIface, txIface string) error
}

// XConnectVppRead provides read methods for cross connects.
type XConnectVppRead interface {
	// DumpXConnectPairs dumps VPP xconnect pair data into the northbound API
	// data structure map indexed by rx interface index.
	DumpXConnectPairs() (map[uint32]*XConnectDetails, error)
}

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel, ifaceidx.IfaceMetadataIndex, idxvpp.NameToIndex, logging.Logger) L2VppAPI
}

func CompatibleL2VppHandler(
	ch govppapi.Channel,
	ifIdx ifaceidx.IfaceMetadataIndex,
	bdIdx idxvpp.NameToIndex,
	log logging.Logger,
) L2VppAPI {
	for ver, h := range Versions {
		if err := ch.CheckCompatiblity(h.Msgs...); err != nil {
			log.Debugf("version %s not compatible", ver)
			continue
		}
		log.Debug("found compatible version:", ver)
		return h.New(ch, ifIdx, bdIdx, log)
	}
	panic("no compatible version available")
}
