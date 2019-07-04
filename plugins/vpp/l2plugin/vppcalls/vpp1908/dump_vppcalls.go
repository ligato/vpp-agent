//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp1908

import (
	"bytes"
	"net"

	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	"github.com/pkg/errors"

	l2nb "github.com/ligato/vpp-agent/api/models/vpp/l2"
	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/l2"
)

// DumpBridgeDomains implements bridge domain handler.
func (h *BridgeDomainVppHandler) DumpBridgeDomains() ([]*vppcalls.BridgeDomainDetails, error) {
	// At first prepare bridge domain ARP termination table which needs to be dumped separately.
	bdArpTab, err := h.dumpBridgeDomainMacTable()
	if err != nil {
		return nil, errors.Errorf("failed to dump arp termination table: %v", err)
	}

	// list of resulting BDs
	var bds []*vppcalls.BridgeDomainDetails

	// dump bridge domains
	reqCtx := h.callsChannel.SendMultiRequest(&l2ba.BridgeDomainDump{BdID: ^uint32(0)})

	for {
		bdDetails := &l2ba.BridgeDomainDetails{}
		stop, err := reqCtx.ReceiveReply(bdDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		// bridge domain metadata
		bdData := &vppcalls.BridgeDomainDetails{
			Bd: &l2nb.BridgeDomain{
				Name:                string(bytes.Replace(bdDetails.BdTag, []byte{0x00}, []byte{}, -1)),
				Flood:               bdDetails.Flood > 0,
				UnknownUnicastFlood: bdDetails.UuFlood > 0,
				Forward:             bdDetails.Forward > 0,
				Learn:               bdDetails.Learn > 0,
				ArpTermination:      bdDetails.ArpTerm > 0,
				MacAge:              uint32(bdDetails.MacAge),
			},
			Meta: &vppcalls.BridgeDomainMeta{
				BdID: bdDetails.BdID,
			},
		}

		// bridge domain interfaces
		for _, iface := range bdDetails.SwIfDetails {
			ifaceName, _, exists := h.ifIndexes.LookupBySwIfIndex(iface.SwIfIndex)
			if !exists {
				h.log.Warnf("Bridge domain dump: interface name for index %d not found", iface.SwIfIndex)
				continue
			}
			// Bvi
			var bvi bool
			if iface.SwIfIndex == bdDetails.BviSwIfIndex {
				bvi = true
			}
			// add interface entry
			bdData.Bd.Interfaces = append(bdData.Bd.Interfaces, &l2nb.BridgeDomain_Interface{
				Name:                    ifaceName,
				BridgedVirtualInterface: bvi,
				SplitHorizonGroup:       uint32(iface.Shg),
			})
		}

		// Add ARP termination entries.
		arpTable, ok := bdArpTab[bdDetails.BdID]
		if ok {
			bdData.Bd.ArpTerminationTable = arpTable
		}

		bds = append(bds, bdData)
	}

	return bds, nil
}

// Reads ARP termination table from all bridge domains. Result is then added to bridge domains.
func (h *BridgeDomainVppHandler) dumpBridgeDomainMacTable() (map[uint32][]*l2nb.BridgeDomain_ArpTerminationEntry, error) {
	bdArpTable := make(map[uint32][]*l2nb.BridgeDomain_ArpTerminationEntry)
	req := &l2ba.BdIPMacDump{BdID: ^uint32(0)}

	reqCtx := h.callsChannel.SendMultiRequest(req)
	for {
		msg := &l2ba.BdIPMacDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if err != nil {
			return nil, err
		}
		if stop {
			break
		}

		// Prepare ARP entry
		arpEntry := &l2nb.BridgeDomain_ArpTerminationEntry{}
		arpEntry.IpAddress = parseAddressToString(msg.Entry.IP)
		arpEntry.PhysAddress = net.HardwareAddr(msg.Entry.Mac[:]).String()

		// Add ARP entry to result map
		bdArpTable[msg.Entry.BdID] = append(bdArpTable[msg.Entry.BdID], arpEntry)
	}

	return bdArpTable, nil
}

// DumpL2FIBs dumps VPP L2 FIB table entries into the northbound API
// data structure map indexed by destination MAC address.
func (h *FIBVppHandler) DumpL2FIBs() (map[string]*vppcalls.FibTableDetails, error) {
	// map for the resulting FIBs
	fibs := make(map[string]*vppcalls.FibTableDetails)

	reqCtx := h.callsChannel.SendMultiRequest(&l2ba.L2FibTableDump{BdID: ^uint32(0)})
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
		var action l2nb.FIBEntry_Action
		if fibDetails.FilterMac > 0 {
			action = l2nb.FIBEntry_DROP
		} else {
			action = l2nb.FIBEntry_FORWARD
		}

		// Interface name (only for FORWARD entries)
		var ifName string
		if action == l2nb.FIBEntry_FORWARD {
			var exists bool
			ifName, _, exists = h.ifIndexes.LookupBySwIfIndex(fibDetails.SwIfIndex)
			if !exists {
				h.log.Warnf("FIB dump: interface name for index %d not found", fibDetails.SwIfIndex)
				continue
			}
		}
		// Bridge domain name
		bdName, _, exists := h.bdIndexes.LookupByIndex(fibDetails.BdID)
		if !exists {
			h.log.Warnf("FIB dump: bridge domain name for index %d not found", fibDetails.BdID)
			continue
		}

		fibs[mac] = &vppcalls.FibTableDetails{
			Fib: &l2nb.FIBEntry{
				PhysAddress:             mac,
				BridgeDomain:            bdName,
				Action:                  action,
				OutgoingInterface:       ifName,
				StaticConfig:            fibDetails.StaticMac > 0,
				BridgedVirtualInterface: fibDetails.BviMac > 0,
			},
			Meta: &vppcalls.FibMeta{
				BdID:  fibDetails.BdID,
				IfIdx: fibDetails.SwIfIndex,
			},
		}
	}

	return fibs, nil
}

// DumpXConnectPairs implements xconnect handler.
func (h *XConnectVppHandler) DumpXConnectPairs() (map[uint32]*vppcalls.XConnectDetails, error) {
	// map for the resulting xconnect pairs
	xpairs := make(map[uint32]*vppcalls.XConnectDetails)

	reqCtx := h.callsChannel.SendMultiRequest(&l2ba.L2XconnectDump{})
	for {
		pairs := &l2ba.L2XconnectDetails{}
		stop, err := reqCtx.ReceiveReply(pairs)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		// Find interface names
		rxIfaceName, _, exists := h.ifIndexes.LookupBySwIfIndex(pairs.RxSwIfIndex)
		if !exists {
			h.log.Warnf("XConnect dump: rx interface name for index %d not found", pairs.RxSwIfIndex)
			continue
		}
		txIfaceName, _, exists := h.ifIndexes.LookupBySwIfIndex(pairs.TxSwIfIndex)
		if !exists {
			h.log.Warnf("XConnect dump: tx interface name for index %d not found", pairs.TxSwIfIndex)
			continue
		}

		xpairs[pairs.RxSwIfIndex] = &vppcalls.XConnectDetails{
			Xc: &l2nb.XConnectPair{
				ReceiveInterface:  rxIfaceName,
				TransmitInterface: txIfaceName,
			},
			Meta: &vppcalls.XcMeta{
				ReceiveInterfaceSwIfIdx:  pairs.RxSwIfIndex,
				TransmitInterfaceSwIfIdx: pairs.TxSwIfIndex,
			},
		}
	}

	return xpairs, nil
}

func parseAddressToString(address l2ba.Address) string {
	var nhIP net.IP = make([]byte, 16)
	copy(nhIP[:], address.Un.XXX_UnionData[:])
	if address.Af == l2ba.ADDRESS_IP4 {
		return nhIP[:4].To4().String()
	}
	if address.Af == l2ba.ADDRESS_IP6 {
		return nhIP.To16().String()
	}

	return ""
}
