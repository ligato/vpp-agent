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
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
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

// FibTableDetails is the wrapper structure for the FIB table entry northbound API structure.
type FibTableDetails struct {
	Fib  *l2.FIBEntry `json:"fib"`
	Meta *FibMeta     `json:"fib_meta"`
}

// FibMeta contains FIB interface and bridge domain name/index map
type FibMeta struct {
	BdID  uint32 `json:"bridge_domain_id"`
	IfIdx uint32 `json:"outgoing_interface_sw_if_idx"`
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

// XConnectDetails is the wrapper structure for the l2 xconnect northbound API structure.
type XConnectDetails struct {
	Xc   *l2.XConnectPair `json:"x_connect"`
	Meta *XcMeta          `json:"x_connect_meta"`
}

// XcMeta contains cross connect rx/tx interface indexes
type XcMeta struct {
	ReceiveInterfaceSwIfIdx  uint32 `json:"receive_interface_sw_if_idx"`
	TransmitInterfaceSwIfIdx uint32 `json:"transmit_interface_sw_if_idx"`
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

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "l2",
	HandlerAPI: (*L2VppAPI)(nil),
})

type NewHandlerFunc func(ch govppapi.Channel, ifDdx ifaceidx.IfaceMetadataIndex, bdIdx idxvpp.NameToIndex, log logging.Logger) L2VppAPI

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
			var bdIdx idxvpp.NameToIndex
			if a[1] != nil {
				bdIdx = a[1].(idxvpp.NameToIndex)
			}
			return h(ch, a[0].(ifaceidx.IfaceMetadataIndex), bdIdx, a[2].(logging.Logger))
		},
	})
}

func CompatibleL2VppHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, bdIdx idxvpp.NameToIndex, log logging.Logger) L2VppAPI {
	if v := Handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, ifIdx, bdIdx, log).(L2VppAPI)
	}
	return nil
}
