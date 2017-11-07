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
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/l2"
	l2nb "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"time"
)

// DumpBridgeDomainIDs lists all configured bridge domains. Auxiliary method for LookupFIBEntries
func DumpBridgeDomainIDs(log logging.Logger, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) ([]uint32, error) {
	// BridgeDomainDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &l2ba.BridgeDomainDump{BdID: ^uint32(0)}
	activeDomains := make([]uint32, 1)
	reqContext := vppChannel.SendMultiRequest(req)
	for {
		msg := &l2ba.BridgeDomainDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			log.Error(err)
			return activeDomains, err
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
func DumpBridgeDomains(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (map[uint32]*BridgeDomain, error) {
	// BridgeDomainDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// map for the resulting BDs
	bds := make(map[uint32]*BridgeDomain)

	// first, dump all interfaces to create initial data
	reqCtx := vppChan.SendMultiRequest(&l2ba.BridgeDomainDump{BdID: ^uint32(0)})

	for {
		bdDetails := &l2ba.BridgeDomainDetails{}
		stop, err := reqCtx.ReceiveReply(bdDetails)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			log.Error(err)
			return nil, err
		}

		// bridge domain details
		bds[bdDetails.BdID] = &BridgeDomain{
			Interfaces: []*BridgeDomainInterface{},
			BridgeDomains_BridgeDomain: l2nb.BridgeDomains_BridgeDomain{
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
	l2nb.FibTableEntries_FibTableEntry
}

// DumpFIBTableEntries dumps VPP FIB table entries into the northbound API data structure
// map indexed by destination MAC address.
func DumpFIBTableEntries(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (map[string]*FIBTableEntry, error) {
	// L2FibTableDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// map for the resulting FIBs
	fibs := make(map[string]*FIBTableEntry)

	reqCtx := vppChan.SendMultiRequest(&l2ba.L2FibTableDump{BdID: ^uint32(0)})
	for {
		fibDetails := &l2ba.L2FibTableDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			log.Error(err)
			return nil, err
		}

		mac := net.HardwareAddr(fibDetails.Mac).String()
		var action l2nb.FibTableEntries_FibTableEntry_Action
		if fibDetails.FilterMac > 0 {
			action = l2nb.FibTableEntries_FibTableEntry_DROP
		} else {
			action = l2nb.FibTableEntries_FibTableEntry_FORWARD
		}

		fibs[mac] = &FIBTableEntry{
			BridgeDomainIdx:          uint32(fibDetails.BdID),
			OutgoingInterfaceSwIfIdx: fibDetails.SwIfIndex,
			FibTableEntries_FibTableEntry: l2nb.FibTableEntries_FibTableEntry{
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
func DumpXConnectPairs(log logging.Logger, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (map[uint32]*XConnectPairs, error) {
	// L2XconnectDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// map for the resulting xconnect pairs
	xpairs := make(map[uint32]*XConnectPairs)

	reqCtx := vppChan.SendMultiRequest(&l2ba.L2XconnectDump{})
	for {
		pairs := &l2ba.L2XconnectDetails{}
		stop, err := reqCtx.ReceiveReply(pairs)
		if stop {
			break // break out of the loop
		}
		if err != nil {
			log.Error(err)
			return nil, err
		}

		xpairs[pairs.RxSwIfIndex] = &XConnectPairs{
			ReceiveInterfaceSwIfIdx:  pairs.RxSwIfIndex,
			TransmitInterfaceSwIfIdx: pairs.TxSwIfIndex,
		}
	}

	return nil, nil
}
