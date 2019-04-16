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
	"fmt"
	"net"

	"github.com/go-errors/errors"
	"github.com/ligato/vpp-agent/api/models/vpp/abf"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/abf"
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

		// next hop IP
		nextHop := net.ParseIP(abfPath.NextHopIp)
		if nextHop.To4() == nil {
			nextHop = nextHop.To16()
		} else {
			nextHop = nextHop.To4()
		}

		fibPath := abf.FibPath{
			SwIfIndex:  ifData.SwIfIndex,
			Weight:     uint8(abfPath.Weight),
			Preference: uint8(abfPath.Preference),
			IsDvr:      boolToUint(abfPath.Dvr),
			NextHop:    parseNextHopToByte(abfPath.NextHopIp),
		}

		fibPaths = append(fibPaths, fibPath)
	}

	return fibPaths
}

// Parses IP address to abf-specific format, where leading zero means IPv4 and other bytes define IP address.
// TODO since there are only 15 bytes available for IP address, only IPv4 is supported (see VPP-1641)
func parseNextHopToByte(nh string) (nhIP []byte) {
	// Support IPv4 only
	if net.ParseIP(nh) == nil || net.ParseIP(nh).To4() == nil {
		return nhIP
	}
	nhIP = net.ParseIP(nh).To4()
	var nhShifted []byte
	for i := 0; i < net.IPv6len; i++ {
		if i == 0 {
			// first element determines IP version (0==IPv4 in this case)
			nhShifted = append(nhShifted, 0)
			continue
		}
		if i <= len(nhIP) {
			// indexes 1-5
			nhShifted = append(nhShifted, nhIP[i-1])
			continue
		}
		// trailing zeros
		nhShifted = append(nhShifted, 0)
	}

	return nhShifted
}

func boolToUint(input bool) uint8 {
	if input {
		return 1
	}
	return 0
}
