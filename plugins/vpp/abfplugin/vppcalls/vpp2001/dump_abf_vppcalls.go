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

package vpp2001

import (
	"net"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/abfplugin/vppcalls"
	vpp_abf "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/abf"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/fib_types"
	abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
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

func (h *ABFVppHandler) dumpABFInterfaces() (map[uint32][]*abf.ABF_AttachedInterface, error) {
	// ABF index <-> attached interfaces
	abfIfs := make(map[uint32][]*abf.ABF_AttachedInterface)

	req := &vpp_abf.AbfItfAttachDump{}
	reqCtx := h.callsChannel.SendMultiRequest(req)

	for {
		reply := &vpp_abf.AbfItfAttachDetails{}
		last, err := reqCtx.ReceiveReply(reply)
		if err != nil {
			return nil, err
		}
		if last {
			break
		}

		// interface name
		ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(uint32(reply.Attach.SwIfIndex))
		if !exists {
			ifName = unknownName
		}

		// attached interface entry
		attached := &abf.ABF_AttachedInterface{
			InputInterface: ifName,
			Priority:       reply.Attach.Priority,
			IsIpv6:         reply.Attach.IsIPv6,
		}

		_, ok := abfIfs[reply.Attach.PolicyID]
		if !ok {
			abfIfs[reply.Attach.PolicyID] = []*abf.ABF_AttachedInterface{}
		}
		abfIfs[reply.Attach.PolicyID] = append(abfIfs[reply.Attach.PolicyID], attached)
	}

	return abfIfs, nil
}

func (h *ABFVppHandler) dumpABFPolicy() ([]*vppcalls.ABFDetails, error) {
	var abfs []*vppcalls.ABFDetails
	req := &vpp_abf.AbfPolicyDump{}
	reqCtx := h.callsChannel.SendMultiRequest(req)

	for {
		reply := &vpp_abf.AbfPolicyDetails{}
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
		var fwdPaths []*abf.ABF_ForwardingPath
		for _, path := range reply.Policy.Paths {
			// interface name
			ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(path.SwIfIndex)
			if !exists {
				ifName = unknownName
			}

			// base fields
			fwdPath := &abf.ABF_ForwardingPath{
				NextHopIp:     parseNextHopToString(path.Nh, path.Proto),
				InterfaceName: ifName,
				Weight:        uint32(path.Weight),
				Preference:    uint32(path.Preference),
				Dvr:           path.Type == fib_types.FIB_API_PATH_TYPE_DVR,
			}
			fwdPaths = append(fwdPaths, fwdPath)
		}

		abfData := &vppcalls.ABFDetails{
			ABF: &abf.ABF{
				Index:           reply.Policy.PolicyID,
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

// returns next hop IP address
func parseNextHopToString(nh vpp_abf.FibPathNh, proto vpp_abf.FibPathNhProto) string {
	if proto == fib_types.FIB_API_PATH_NH_PROTO_IP4 {
		addr := nh.Address.GetIP4()
		return net.IP(addr[:]).To4().String()
	}
	if proto == fib_types.FIB_API_PATH_NH_PROTO_IP6 {
		addr := nh.Address.GetIP6()
		return net.IP(addr[:]).To16().String()
	}
	return ""
}
