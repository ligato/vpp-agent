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

package vpp1904

import (
	"fmt"
	"net"

	vpp_abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"

	"github.com/go-errors/errors"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/abf"
)

// GetAbfVersion retrieves version of the VPP ABF plugin
func (h *ABFVppHandler) GetAbfVersion() (ver string, err error) {
	req := &abf.AbfPluginGetVersion{}
	reply := &abf.AbfPluginGetVersionReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d.%d", reply.Major, reply.Minor), nil
}

// AddAbfPolicy creates new ABF entry together with a list of forwarding paths
func (h *ABFVppHandler) AddAbfPolicy(policyID, aclID uint32, abfPaths []*vpp_abf.ABF_ForwardingPath) error {
	if err := h.abfAddDelPolicy(policyID, aclID, abfPaths, true); err != nil {
		return errors.Errorf("failed to add ABF policy %d (ACL: %v): %v", policyID, aclID, err)
	}
	return nil
}

// DeleteAbfPolicy removes existing ABF entry
func (h *ABFVppHandler) DeleteAbfPolicy(policyID uint32, abfPaths []*vpp_abf.ABF_ForwardingPath) error {
	if err := h.abfAddDelPolicy(policyID, 0, abfPaths, false); err != nil {
		return errors.Errorf("failed to delete ABF policy %d: %v", policyID, err)
	}
	return nil
}

// AbfAttachInterfaceIPv4 attaches IPv4 interface to the ABF
func (h *ABFVppHandler) AbfAttachInterfaceIPv4(policyID, ifIdx, priority uint32) error {
	if err := h.abfAttachDetachInterface(policyID, ifIdx, priority, true, false); err != nil {
		return errors.Errorf("failed to attach IPv4 interface %d to ABF policy %d: %v", ifIdx, policyID, err)
	}
	return nil
}

// AbfDetachInterfaceIPv4 detaches IPV4 interface from the ABF
func (h *ABFVppHandler) AbfDetachInterfaceIPv4(policyID, ifIdx, priority uint32) error {
	if err := h.abfAttachDetachInterface(policyID, ifIdx, priority, false, false); err != nil {
		return errors.Errorf("failed to detach IPv4 interface %d from ABF policy %d: %v", ifIdx, policyID, err)
	}
	return nil
}

// AbfAttachInterfaceIPv6 attaches IPv6 interface to the ABF
func (h *ABFVppHandler) AbfAttachInterfaceIPv6(policyID, ifIdx, priority uint32) error {
	if err := h.abfAttachDetachInterface(policyID, ifIdx, priority, true, true); err != nil {
		return errors.Errorf("failed to attach IPv6 interface %d to ABF policy %d: %v", ifIdx, policyID, err)
	}
	return nil
}

// AbfDetachInterfaceIPv6 detaches IPv6 interface from the ABF
func (h *ABFVppHandler) AbfDetachInterfaceIPv6(policyID, ifIdx, priority uint32) error {
	if err := h.abfAttachDetachInterface(policyID, ifIdx, priority, false, true); err != nil {
		return errors.Errorf("failed to detach IPv6 interface %d from ABF policy %d: %v", ifIdx, policyID, err)
	}
	return nil
}

func (h *ABFVppHandler) abfAttachDetachInterface(policyID, ifIdx, priority uint32, isAdd, isIPv6 bool) error {
	req := &abf.AbfItfAttachAddDel{
		IsAdd: boolToUint(isAdd),
		Attach: abf.AbfItfAttach{
			PolicyID:  policyID,
			SwIfIndex: ifIdx,
			Priority:  priority,
			IsIPv6:    boolToUint(isIPv6),
		},
	}
	reply := &abf.AbfItfAttachAddDelReply{}

	return h.callsChannel.SendRequest(req).ReceiveReply(reply)
}

func (h *ABFVppHandler) abfAddDelPolicy(policyID, aclID uint32, abfPaths []*vpp_abf.ABF_ForwardingPath, isAdd bool) error {
	req := &abf.AbfPolicyAddDel{
		IsAdd: boolToUint(isAdd),
		Policy: abf.AbfPolicy{
			PolicyID: policyID,
			ACLIndex: aclID,
			Paths:    h.toFibPaths(abfPaths),
			NPaths:   uint8(len(abfPaths)),
		},
	}
	reply := &abf.AbfPolicyAddDelReply{}

	return h.callsChannel.SendRequest(req).ReceiveReply(reply)
}

func (h *ABFVppHandler) toFibPaths(abfPaths []*vpp_abf.ABF_ForwardingPath) (fibPaths []abf.FibPath) {
	for _, abfPath := range abfPaths {
		// fib path interface
		ifData, exists := h.ifIndexes.LookupByName(abfPath.InterfaceName)
		if !exists {
			continue
		}

		fibPath := abf.FibPath{
			SwIfIndex:  ifData.SwIfIndex,
			Weight:     uint8(abfPath.Weight),
			Preference: uint8(abfPath.Preference),
			IsDvr:      boolToUint(abfPath.Dvr),
		}

		// next hop IP
		nextHop := net.ParseIP(abfPath.NextHopIp)
		if nextHop.To4() == nil {
			nextHop = nextHop.To16()
			fibPath.Afi = 1
		} else {
			nextHop = nextHop.To4()
			fibPath.Afi = 0
		}
		fibPath.NextHop = nextHop

		fibPaths = append(fibPaths, fibPath)
	}

	return fibPaths
}

func boolToUint(input bool) uint8 {
	if input {
		return 1
	}
	return 0
}
