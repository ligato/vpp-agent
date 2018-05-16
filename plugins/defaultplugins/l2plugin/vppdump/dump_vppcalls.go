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

package vppdump

import (
	"bytes"
	"net"
	"time"

	"github.com/ligato/cn-infra/logging/measure"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
	l2nb "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls"
)

// DumpBridgeDomainIDs lists all configured bridge domains. Auxiliary method for LookupFIBEntries.
// returns list of bridge domain IDs (BD IDs). First element of returned slice is 0. It is default BD to which all
// interfaces belong
func DumpBridgeDomainIDs(vppChannel vppcalls.VPPChannel, stopwatch *measure.Stopwatch) ([]uint32, error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.BridgeDomainDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &l2ba.BridgeDomainDump{BdID: ^uint32(0)}
	activeDomains := make([]uint32, 1)
	reqCtx := vppChannel.SendMultiRequest(req)
	for {
		msg := &l2ba.BridgeDomainDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if err != nil {
			return nil, err
		}
		if stop {
			break
		}
		activeDomains = append(activeDomains, msg.BdID)
	}

	return activeDomains, nil
}

// BridgeDomain is the wrapper structure for the bridge domain northbound API structure.
// NOTE: Interfaces in BridgeDomains_BridgeDomain is overridden by the local Interfaces member.
type BridgeDomain struct {
	Interfaces []*BridgeDomainInterface `json:"interfaces"`
	l2nb.BridgeDomains_BridgeDomain
}

// BridgeDomainInterface is the wrapper structure for the bridge domain interface northbound API structure.
type BridgeDomainInterface struct {
	SwIfIndex uint32 `json:"sw_if_index"`
	l2nb.BridgeDomains_BridgeDomain_Interfaces
}

// DumpBridgeDomains dumps VPP bridge domain data into the northbound API data structure
// map indexed by bridge domain ID.
//
// LIMITATIONS:
// - not able to dump ArpTerminationTable - missing binary API
//
func DumpBridgeDomains(vppChan vppcalls.VPPChannel, stopwatch *measure.Stopwatch) (map[uint32]*BridgeDomain, error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.BridgeDomainDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// map for the resulting BDs
	bds := make(map[uint32]*BridgeDomain)

	// First, dump all interfaces to create initial data.
	reqCtx := vppChan.SendMultiRequest(&l2ba.BridgeDomainDump{BdID: ^uint32(0)})

	for {
		bdDetails := &l2ba.BridgeDomainDetails{}
		stop, err := reqCtx.ReceiveReply(bdDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return nil, err
		}

		// bridge domain details
		bds[bdDetails.BdID] = &BridgeDomain{
			Interfaces: []*BridgeDomainInterface{},
			BridgeDomains_BridgeDomain: l2nb.BridgeDomains_BridgeDomain{
				Name:                string(bytes.Replace(bdDetails.BdTag, []byte{0x00}, []byte{}, -1)),
				Flood:               bdDetails.Flood > 0,
				UnknownUnicastFlood: bdDetails.UuFlood > 0,
				Forward:             bdDetails.Forward > 0,
				Learn:               bdDetails.Learn > 0,
				ArpTermination:      bdDetails.ArpTerm > 0,
				MacAge:              uint32(bdDetails.MacAge),
			},
		}

		// bridge domain interfaces
		for _, iface := range bdDetails.SwIfDetails {
			bds[bdDetails.BdID].Interfaces = append(bds[bdDetails.BdID].Interfaces, &BridgeDomainInterface{
				SwIfIndex: iface.SwIfIndex,
			})
		}

	}

	return bds, nil
}

// FIBTableEntry is the wrapper structure for the FIB table entry northbound API structure.
type FIBTableEntry struct {
	BridgeDomainIdx          uint32 `json:"bridge_domain_idx"`
	OutgoingInterfaceSwIfIdx uint32 `json:"outgoing_interface_sw_if_idx"`
	l2nb.FibTable_FibEntry
}

// DumpFIBTableEntries dumps VPP FIB table entries into the northbound API data structure
// map indexed by destination MAC address.
func DumpFIBTableEntries(vppChan vppcalls.VPPChannel, stopwatch *measure.Stopwatch) (map[string]*FIBTableEntry, error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.L2FibTableDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// map for the resulting FIBs
	fibs := make(map[string]*FIBTableEntry)

	reqCtx := vppChan.SendMultiRequest(&l2ba.L2FibTableDump{BdID: ^uint32(0)})
	for {
		fibDetails := &l2ba.L2FibTableDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return nil, err
		}

		mac := net.HardwareAddr(fibDetails.Mac).String()
		var action l2nb.FibTable_FibEntry_Action
		if fibDetails.FilterMac > 0 {
			action = l2nb.FibTable_FibEntry_DROP
		} else {
			action = l2nb.FibTable_FibEntry_FORWARD
		}

		fibs[mac] = &FIBTableEntry{
			BridgeDomainIdx:          uint32(fibDetails.BdID),
			OutgoingInterfaceSwIfIdx: fibDetails.SwIfIndex,
			FibTable_FibEntry: l2nb.FibTable_FibEntry{
				PhysAddress:             mac,
				Action:                  action,
				StaticConfig:            fibDetails.StaticMac > 0,
				BridgedVirtualInterface: fibDetails.BviMac > 0,
			},
		}
	}

	return fibs, nil
}

// XConnectPairs is the wrapper structure for the l2 xconnect northbound API structure.
type XConnectPairs struct {
	ReceiveInterfaceSwIfIdx  uint32 `json:"receive_interface_sw_if_idx"`
	TransmitInterfaceSwIfIdx uint32 `json:"transmit_interface_sw_if_idx"`
}

// DumpXConnectPairs dumps VPP xconnect pair data into the northbound API data structure
// map indexed by rx interface index.
func DumpXConnectPairs(vppChan vppcalls.VPPChannel, stopwatch *measure.Stopwatch) (map[uint32]*XConnectPairs, error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.L2XconnectDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// map for the resulting xconnect pairs
	xpairs := make(map[uint32]*XConnectPairs)

	reqCtx := vppChan.SendMultiRequest(&l2ba.L2XconnectDump{})
	for {
		pairs := &l2ba.L2XconnectDetails{}
		stop, err := reqCtx.ReceiveReply(pairs)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return nil, err
		}

		xpairs[pairs.RxSwIfIndex] = &XConnectPairs{
			ReceiveInterfaceSwIfIdx:  pairs.RxSwIfIndex,
			TransmitInterfaceSwIfIdx: pairs.TxSwIfIndex,
		}
	}

	return xpairs, nil
}
