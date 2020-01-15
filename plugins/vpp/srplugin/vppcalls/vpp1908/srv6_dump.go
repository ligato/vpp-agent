// Copyright (c) 2019 Pantheon.tech
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

package vpp1908

import (
	"net"

	"github.com/go-errors/errors"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/sr"
	srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"
)

// DumpLocalSids retrieves all localsids
func (h *SRv6VppHandler) DumpLocalSids() (localsids []*srv6.LocalSID, err error) {
	h.log.Debug("Dumping LocalSIDs")

	reqCtx := h.callsChannel.SendMultiRequest(&sr.SrLocalsidsDump{})
	for {
		// retrieve dump for another localsid
		dumpReply := &sr.SrLocalsidsDetails{}
		stop, err := reqCtx.ReceiveReply(dumpReply)
		if stop {
			break
		}
		if err != nil {
			return nil, errors.Errorf("error while retrieving localsid details:%v", err)
		}

		// convert dumped localsid details into modeled Localsid struct
		sid, error := h.convertDumpedSid(&dumpReply.Addr)
		if error != nil {
			return localsids, errors.Errorf("can't properly handle sid address "+
				"of dumped localsid %v due to: %v", dumpReply, err)
		}
		localsid := &srv6.LocalSID{
			InstallationVrfId: dumpReply.FibTable,
			Sid:               sid.String(),
		}
		if err := h.fillEndFunction(localsid, dumpReply); err != nil {
			return localsids, errors.Errorf("can't properly handle end function "+
				"of dumped localsid %v due to: %v", dumpReply, err)
		}

		// collect all dumped localsids
		localsids = append(localsids, localsid)
	}

	return localsids, nil
}

// convertDumpedSid extract from dumped structure SID value and converts it to IPv6 (net.IP)
func (h *SRv6VppHandler) convertDumpedSid(srv6Sid *sr.Srv6Sid) (net.IP, error) {
	if srv6Sid == nil || srv6Sid.Addr == nil {
		return nil, errors.New("can't convert sid from nil dumped address (or nil srv6sid)")
	}
	sid := net.IP(srv6Sid.Addr).To16()
	if sid == nil {
		return nil, errors.Errorf("can't convert dumped SID bytes(%v) to net.IP", srv6Sid.Addr)
	}
	return sid, nil
}

// fillEndFunction create end function part of NB-modeled localsid from SB-dumped structure
func (h *SRv6VppHandler) fillEndFunction(localSID *srv6.LocalSID, dumpReply *sr.SrLocalsidsDetails) error {
	switch uint8(dumpReply.Behavior) {
	case BehaviorEnd:
		localSID.EndFunction = &srv6.LocalSID_BaseEndFunction{
			BaseEndFunction: &srv6.LocalSID_End{
				Psp: uintToBool(dumpReply.EndPsp),
			},
		}
	case BehaviorX:
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(dumpReply.XconnectIfaceOrVrfTable)
		if !exists {
			return errors.Errorf("there is no interface with sw index %v", dumpReply.XconnectIfaceOrVrfTable)
		}
		localSID.EndFunction = &srv6.LocalSID_EndFunctionX{
			EndFunctionX: &srv6.LocalSID_EndX{
				Psp:               uintToBool(dumpReply.EndPsp),
				OutgoingInterface: ifName,
				NextHop:           h.nextHop(dumpReply),
			},
		}
	case BehaviorT:
		localSID.EndFunction = &srv6.LocalSID_EndFunctionT{
			EndFunctionT: &srv6.LocalSID_EndT{
				Psp:   uintToBool(dumpReply.EndPsp),
				VrfId: dumpReply.XconnectIfaceOrVrfTable,
			},
		}
	case BehaviorDX2:
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(dumpReply.XconnectIfaceOrVrfTable)
		if !exists {
			return errors.Errorf("there is no interface with sw index %v", dumpReply.XconnectIfaceOrVrfTable)
		}
		localSID.EndFunction = &srv6.LocalSID_EndFunctionDx2{
			EndFunctionDx2: &srv6.LocalSID_EndDX2{
				VlanTag:           dumpReply.VlanIndex,
				OutgoingInterface: ifName,
			},
		}
	case BehaviorDX4:
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(dumpReply.XconnectIfaceOrVrfTable)
		if !exists {
			return errors.Errorf("there is no interface with sw index %v", dumpReply.XconnectIfaceOrVrfTable)
		}
		localSID.EndFunction = &srv6.LocalSID_EndFunctionDx4{
			EndFunctionDx4: &srv6.LocalSID_EndDX4{
				OutgoingInterface: ifName,
				NextHop:           h.nextHop(dumpReply),
			},
		}
	case BehaviorDX6:
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(dumpReply.XconnectIfaceOrVrfTable)
		if !exists {
			return errors.Errorf("there is no interface with sw index %v", dumpReply.XconnectIfaceOrVrfTable)
		}
		localSID.EndFunction = &srv6.LocalSID_EndFunctionDx6{
			EndFunctionDx6: &srv6.LocalSID_EndDX6{
				OutgoingInterface: ifName,
				NextHop:           h.nextHop(dumpReply),
			},
		}
	case BehaviorDT4:
		localSID.EndFunction = &srv6.LocalSID_EndFunctionDt4{
			EndFunctionDt4: &srv6.LocalSID_EndDT4{
				VrfId: dumpReply.XconnectIfaceOrVrfTable,
			},
		}
	case BehaviorDT6:
		localSID.EndFunction = &srv6.LocalSID_EndFunctionDt6{
			EndFunctionDt6: &srv6.LocalSID_EndDT6{
				VrfId: dumpReply.XconnectIfaceOrVrfTable,
			},
		}
	default:
		return errors.Errorf("localsid with unknown or unsupported behavior (%v)", dumpReply.Behavior)
	}

	return nil
}

// nextHop transforms SB dump data about next hop to NB modeled structure
func (h *SRv6VppHandler) nextHop(dumpReply *sr.SrLocalsidsDetails) string {
	nh4 := net.IP(dumpReply.XconnectNhAddr4)
	nh6 := net.IP(dumpReply.XconnectNhAddr6)
	nhStr := ""
	// default is no next hop address (i.e. L2 xconnect)
	if nh4 != nil && !nh4.Equal(net.IPv4zero) {
		nhStr = nh4.String()
	} else if nh6 != nil && !nh6.Equal(net.IPv6zero) {
		nhStr = nh6.String()
	}
	return nhStr
}

func uintToBool(input uint8) bool {
	return input != 0
}
