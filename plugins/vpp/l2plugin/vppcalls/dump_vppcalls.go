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
	"bytes"
	"net"
	"time"

	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/l2"
	l2nb "github.com/ligato/vpp-agent/plugins/vpp/model/l2"
)

// BridgeDomainDetails is the wrapper structure for the bridge domain northbound API structure.
// NOTE: Interfaces in BridgeDomains_BridgeDomain is overridden by the local Interfaces member.
type BridgeDomainDetails struct {
	Bd   *l2nb.BridgeDomains_BridgeDomain `json:"bridge_domain"`
	Meta *BridgeDomainMeta                `json:"bridge_domain_meta"`
}

// BridgeDomainMeta contains bridge domain interface name/index map
type BridgeDomainMeta struct {
	BdID          uint32            `json:"bridge_domain_id"`
	BdIfIdxToName map[uint32]string `json:"bridge_domain_id_to_name"`
}

// DumpBridgeDomains implements bridge domain handler.
func (handler *BridgeDomainVppHandler) DumpBridgeDomains() (map[uint32]*BridgeDomainDetails, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(l2ba.BridgeDomainDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// map for the resulting BDs
	bds := make(map[uint32]*BridgeDomainDetails)

	// First, dump all interfaces to create initial data.
	reqCtx := handler.callsChannel.SendMultiRequest(&l2ba.BridgeDomainDump{BdID: ^uint32(0)})

	for {
		bdDetails := &l2ba.BridgeDomainDetails{}
		stop, err := reqCtx.ReceiveReply(bdDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return nil, err
		}

		// base bridge domain details
		bds[bdDetails.BdID] = &BridgeDomainDetails{
			Bd: &l2nb.BridgeDomains_BridgeDomain{
				Name:                string(bytes.Replace(bdDetails.BdTag, []byte{0x00}, []byte{}, -1)),
				Flood:               bdDetails.Flood > 0,
				UnknownUnicastFlood: bdDetails.UuFlood > 0,
				Forward:             bdDetails.Forward > 0,
				Learn:               bdDetails.Learn > 0,
				ArpTermination:      bdDetails.ArpTerm > 0,
				MacAge:              uint32(bdDetails.MacAge),
			},
			Meta: &BridgeDomainMeta{
				BdID:          bdDetails.BdID,
				BdIfIdxToName: make(map[uint32]string),
			},
		}

		// bridge domain interfaces and metadata
		for _, iface := range bdDetails.SwIfDetails {
			ifName, _, exists := handler.ifIndexes.LookupName(iface.SwIfIndex)
			if !exists {
				handler.log.Warnf("Bridge domain dump: interface name for index %d not found", iface.SwIfIndex)
				continue
			}
			// Bvi
			var bvi bool
			if iface.SwIfIndex == bdDetails.BviSwIfIndex {
				bvi = true
			}
			// Add metadata entry
			bds[bdDetails.BdID].Meta.BdIfIdxToName[iface.SwIfIndex] = ifName
			// Add interface entry
			bds[bdDetails.BdID].Bd.Interfaces = append(bds[bdDetails.BdID].Bd.Interfaces, &l2nb.BridgeDomains_BridgeDomain_Interfaces{
				Name: ifName,
				BridgedVirtualInterface: bvi,
				SplitHorizonGroup:       uint32(iface.Shg),
			})
		}
	}

	return bds, nil
}

// DumpBridgeDomainIDs implements bridge domain handler.
func (handler *BridgeDomainVppHandler) DumpBridgeDomainIDs() ([]uint32, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(l2ba.BridgeDomainDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &l2ba.BridgeDomainDump{BdID: ^uint32(0)}
	var activeDomains []uint32
	reqCtx := handler.callsChannel.SendMultiRequest(req)
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

// FibTableDetails is the wrapper structure for the FIB table entry northbound API structure.
type FibTableDetails struct {
	Fib  *l2nb.FibTable_FibEntry `json:"fib"`
	Meta *FibMeta                `json:"fib_meta"`
}

// FibMeta contains FIB interface and bridge domain name/index map
type FibMeta struct {
	BdID  uint32 `json:"bridge_domain_id"`
	IfIdx uint32 `json:"outgoing_interface_sw_if_idx"`
}

// DumpFIBTableEntries implements fib handler.
func (handler *FibVppHandler) DumpFIBTableEntries() (map[string]*FibTableDetails, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(l2ba.L2FibTableDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// map for the resulting FIBs
	fibs := make(map[string]*FibTableDetails)

	reqCtx := handler.syncCallsChannel.SendMultiRequest(&l2ba.L2FibTableDump{BdID: ^uint32(0)})
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

		// Interface name
		ifName, _, exists := handler.ifIndexes.LookupName(fibDetails.SwIfIndex)
		if !exists {
			handler.log.Warnf("FIB dump: interface name for index %s not found", fibDetails.SwIfIndex)
		}
		// Bridge domain name
		bdName, _, exists := handler.bdIndexes.LookupName(fibDetails.BdID)
		if !exists {
			handler.log.Warnf("FIB dump: bridge domain name for index %s not found", fibDetails.BdID)
		}

		fibs[mac] = &FibTableDetails{
			Fib: &l2nb.FibTable_FibEntry{
				PhysAddress:             mac,
				BridgeDomain:            bdName,
				Action:                  action,
				OutgoingInterface:       ifName,
				StaticConfig:            fibDetails.StaticMac > 0,
				BridgedVirtualInterface: fibDetails.BviMac > 0,
			},
			Meta: &FibMeta{
				BdID:  fibDetails.BdID,
				IfIdx: fibDetails.SwIfIndex,
			},
		}
	}

	return fibs, nil
}

// XConnectDetails is the wrapper structure for the l2 xconnect northbound API structure.
type XConnectDetails struct {
	Xc   *l2nb.XConnectPairs_XConnectPair `json:"x_connect"`
	Meta *XcMeta                          `json:"x_connect_meta"`
}

// XcMeta contains cross connect rx/tx interface indexes
type XcMeta struct {
	ReceiveInterfaceSwIfIdx  uint32 `json:"receive_interface_sw_if_idx"`
	TransmitInterfaceSwIfIdx uint32 `json:"transmit_interface_sw_if_idx"`
}

// DumpXConnectPairs implements xconnect handler.
func (handler *XConnectVppHandler) DumpXConnectPairs() (map[uint32]*XConnectDetails, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(l2ba.L2XconnectDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// map for the resulting xconnect pairs
	xpairs := make(map[uint32]*XConnectDetails)

	reqCtx := handler.callsChannel.SendMultiRequest(&l2ba.L2XconnectDump{})
	for {
		pairs := &l2ba.L2XconnectDetails{}
		stop, err := reqCtx.ReceiveReply(pairs)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return nil, err
		}

		// Find interface names
		rxIfaceName, _, exists := handler.ifIndexes.LookupName(pairs.RxSwIfIndex)
		if !exists {
			handler.log.Warnf("XConnect dump: rx interface name for index %s not found", pairs.RxSwIfIndex)
		}
		txIfaceName, _, exists := handler.ifIndexes.LookupName(pairs.TxSwIfIndex)
		if !exists {
			handler.log.Warnf("XConnect dump: tx interface name for index %s not found", pairs.TxSwIfIndex)
		}

		xpairs[pairs.RxSwIfIndex] = &XConnectDetails{
			Xc: &l2nb.XConnectPairs_XConnectPair{
				ReceiveInterface:  rxIfaceName,
				TransmitInterface: txIfaceName,
			},
			Meta: &XcMeta{
				ReceiveInterfaceSwIfIdx:  pairs.RxSwIfIndex,
				TransmitInterfaceSwIfIdx: pairs.TxSwIfIndex,
			},
		}
	}

	return xpairs, nil
}
