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

package vpp1901

import (
	"net"
	"strconv"

	vpp_abf "github.com/ligato/vpp-agent/api/models/vpp/abf"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/abf"
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
			// ip address
			var nextHopIP net.IP = path.NextHop

			// interface name
			ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(path.SwIfIndex)
			if !exists {
				ifName = unknownName
			}

			// base fields
			fwdPath := &vpp_abf.ABF_ForwardingPath{
				NextHopIp:       nextHopIP.String(),
				InterfaceName:   ifName,
				Vrf:             path.TableID,
				Weight:          uint32(path.Weight),
				Preference:      uint32(path.Preference),
				Afi:             uint32(path.Afi),
				NextHopId:       path.NextHopID,
				RpfId:           path.RpfID,
				ViaLabel:        path.ViaLabel,
				Local:           uintToBool(path.IsLocal),
				Drop:            uintToBool(path.IsDrop),
				UdpEncap:        uintToBool(path.IsUDPEncap),
				Unreachable:     uintToBool(path.IsUnreach),
				Prohibit:        uintToBool(path.IsProhibit),
				ResolveHost:     uintToBool(path.IsResolveHost),
				ResolveAttached: uintToBool(path.IsResolveAttached),
				Dvr:             uintToBool(path.IsDvr),
				SourceLookup:    uintToBool(path.IsSourceLookup),
				InterfaceRx:     uintToBool(path.IsInterfaceRx),
			}

			// label stack
			var labelStack []*vpp_abf.ABF_ForwardingPath_Label
			for _, label := range path.LabelStack {
				labelEntry := &vpp_abf.ABF_ForwardingPath_Label{
					IsUniform: uintToBool(label.IsUniform),
					Label:     label.Label,
					TTL:       uint32(label.TTL),
					Exp:       uint32(label.Exp),
				}

				labelStack = append(labelStack, labelEntry)
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

func uintToBool(value uint8) bool {
	if value == 0 {
		return false
	}
	return true
}
