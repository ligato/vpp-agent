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

package vpp1810

import (
	"net"
	"strconv"

	"github.com/ligato/vpp-agent/api/models/vpp/abf"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/abf"
)

// placeholder for unknown names
const unknownName = "<unknown>"

// DumpABFPolicy retrieves VPP ABF configuration.
func (h *ABFVppHandler) DumpABFPolicy() ([]*vppcalls.ABFDetails, error) {
	// retrieve ABF interfaces
	attachedIfs, err := h.dumpABFInterfaces()
	if err != nil {
		return nil, err
	}

	// retrieve ABF policy
	abfPolicy, err := h.dumpABFPolicy()
	if err != nil {
		return nil, err
	}

	// merge attached interfaces data to policy
	for _, policy := range abfPolicy {
		ifData, ok := attachedIfs[policy.Meta.PolicyID]
		if ok {
			policy.ABF.AttachedInterfaces = ifData
		}
	}

	return abfPolicy, nil
}

func (h *ABFVppHandler) dumpABFInterfaces() (map[uint32][]*vpp_abf.ABF_AttachedInterface, error) {
	// ABF index <-> attached interfaces
	abfIfs := make(map[uint32][]*vpp_abf.ABF_AttachedInterface)

	req := &abf.AbfItfAttachDump{}
	reqCtx := h.dumpChannel.SendMultiRequest(req)

	for {
		reply := &abf.AbfItfAttachDetails{}
		last, err := reqCtx.ReceiveReply(reply)
		if err != nil {
			return nil, err
		}
		if last {
			break
		}

		// interface name
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(reply.Attach.SwIfIndex)
		if !exists {
			ifName = unknownName
		}

		// attached interface entry
		attached := &vpp_abf.ABF_AttachedInterface{
			InputInterface: ifName,
			Priority:       reply.Attach.Priority,
			IsIpv6:         uintToBool(reply.Attach.IsIPv6),
		}

		_, ok := abfIfs[reply.Attach.PolicyID]
		if !ok {
			abfIfs[reply.Attach.PolicyID] = []*vpp_abf.ABF_AttachedInterface{}
		}
		abfIfs[reply.Attach.PolicyID] = append(abfIfs[reply.Attach.PolicyID], attached)
	}

	return abfIfs, nil
}

func (h *ABFVppHandler) dumpABFPolicy() ([]*vppcalls.ABFDetails, error) {
	var abfs []*vppcalls.ABFDetails
	req := &abf.AbfPolicyDump{}
	reqCtx := h.dumpChannel.SendMultiRequest(req)

	for {
		reply := &abf.AbfPolicyDetails{}
		last, err := reqCtx.ReceiveReply(reply)
		if err != nil {
			return nil, err
		}
		if last {
			break
		}

		// ACL name
		aclName, _, exists := h.aclIndexes.LookupByIndex(reply.Policy.ACLIndex)
		if !exists {
			aclName = unknownName
		}

		// paths
		var fwdPaths []*vpp_abf.ABF_ForwardingPath
		for _, path := range reply.Policy.Paths {
			// interface name
			ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(path.SwIfIndex)
			if !exists {
				ifName = unknownName
			}

			// base fields
			fwdPath := &vpp_abf.ABF_ForwardingPath{
				NextHopIp:     parseNextHopToString(path.NextHop),
				InterfaceName: ifName,
				Weight:        uint32(path.Weight),
				Preference:    uint32(path.Preference),
				Dvr:           uintToBool(path.IsDvr),
			}

			fwdPaths = append(fwdPaths, fwdPath)
		}

		abfData := &vppcalls.ABFDetails{
			ABF: &vpp_abf.ABF{
				Index:           strconv.Itoa(int(reply.Policy.PolicyID)),
				AclName:         aclName,
				ForwardingPaths: fwdPaths,
			},
			Meta: &vppcalls.ABFMeta{
				PolicyID: reply.Policy.PolicyID,
			},
		}

		abfs = append(abfs, abfData)
	}

	return abfs, nil
}

// Parses IP address to string. The IP address is received in format where leading byte means IP version
// (1==IPv4, 2==IPv6) and rest is the IP address.
// TODO IPv6 not supported since there are only 15 bytes for address itself so it is not returned whole (see VPP-1641)
func parseNextHopToString(nh []byte) string {
	if len(nh) != net.IPv6len {
		return ""
	}
	// the first byte determines the IP version
	if nh[0] == 1 {
		// IPv4
		var nhShifted net.IP = append(nh[1:5])
		return nhShifted.To4().String()
	}

	return ""
}

func uintToBool(value uint8) bool {
	if value == 0 {
		return false
	}
	return true
}
