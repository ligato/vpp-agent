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
	"fmt"
	"net"

	"github.com/go-errors/errors"

	vpp_abf "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/abf"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/fib_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interface_types"
	abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
)

const (
	// NextHopViaLabelUnset constant has to be assigned into the field next hop  via label
	// in abf_policy_add_del binary message if next hop via label is not defined.
	NextHopViaLabelUnset uint32 = 0xfffff + 1

	// ClassifyTableIndexUnset is a default value for field classify_table_index
	// in abf_policy_add_del binary message.
	ClassifyTableIndexUnset = ^uint32(0)
)

// GetAbfVersion retrieves version of the VPP ABF plugin
func (h *ABFVppHandler) GetAbfVersion() (ver string, err error) {
	req := &vpp_abf.AbfPluginGetVersion{}
	reply := &vpp_abf.AbfPluginGetVersionReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d.%d", reply.Major, reply.Minor), nil
}

// AddAbfPolicy creates new ABF entry together with a list of forwarding paths
func (h *ABFVppHandler) AddAbfPolicy(policyID, aclID uint32, abfPaths []*abf.ABF_ForwardingPath) error {
	if err := h.abfAddDelPolicy(policyID, aclID, abfPaths, true); err != nil {
		return errors.Errorf("failed to add ABF policy %d (ACL: %v): %v", policyID, aclID, err)
	}
	return nil
}

// DeleteAbfPolicy removes existing ABF entry
func (h *ABFVppHandler) DeleteAbfPolicy(policyID uint32, abfPaths []*abf.ABF_ForwardingPath) error {
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
	req := &vpp_abf.AbfItfAttachAddDel{
		IsAdd: isAdd,
		Attach: vpp_abf.AbfItfAttach{
			PolicyID:  policyID,
			SwIfIndex: interface_types.InterfaceIndex(ifIdx),
			Priority:  priority,
			IsIPv6:    isIPv6,
		},
	}
	reply := &vpp_abf.AbfItfAttachAddDelReply{}

	return h.callsChannel.SendRequest(req).ReceiveReply(reply)
}

func (h *ABFVppHandler) abfAddDelPolicy(policyID, aclID uint32, abfPaths []*abf.ABF_ForwardingPath, isAdd bool) error {
	req := &vpp_abf.AbfPolicyAddDel{
		IsAdd: isAdd,
		Policy: vpp_abf.AbfPolicy{
			PolicyID: policyID,
			ACLIndex: aclID,
			Paths:    h.toFibPaths(abfPaths),
			NPaths:   uint8(len(abfPaths)),
		},
	}
	reply := &vpp_abf.AbfPolicyAddDelReply{}

	return h.callsChannel.SendRequest(req).ReceiveReply(reply)
}

func (h *ABFVppHandler) toFibPaths(abfPaths []*abf.ABF_ForwardingPath) (fibPaths []vpp_abf.FibPath) {
	var err error
	for _, abfPath := range abfPaths {
		// fib path interface
		ifData, exists := h.ifIndexes.LookupByName(abfPath.InterfaceName)
		if !exists {
			continue
		}

		fibPath := vpp_abf.FibPath{
			SwIfIndex:  ifData.SwIfIndex,
			Weight:     uint8(abfPath.Weight),
			Preference: uint8(abfPath.Preference),
			Type:       setFibPathType(abfPath.Dvr),
		}
		if fibPath.Nh, fibPath.Proto, err = setFibPathNhAndProto(abfPath.NextHopIp); err != nil {
			h.log.Errorf("ABF path next hop error: %v", err)
		}
		fibPaths = append(fibPaths, fibPath)
	}

	return fibPaths
}

// supported cases are DVR and normal
func setFibPathType(isDvr bool) vpp_abf.FibPathType {
	if isDvr {
		return fib_types.FIB_API_PATH_TYPE_DVR
	}
	return fib_types.FIB_API_PATH_TYPE_NORMAL
}

// resolve IP address and return FIB path next hop (IP address) and IPv4/IPv6 version
func setFibPathNhAndProto(ipStr string) (nh vpp_abf.FibPathNh, proto vpp_abf.FibPathNhProto, err error) {
	netIP := net.ParseIP(ipStr)
	if netIP == nil {
		return nh, proto, errors.Errorf("failed to parse next hop IP address %s", ipStr)
	}
	var au fib_types.AddressUnion
	if ipv4 := netIP.To4(); ipv4 == nil {
		var address fib_types.IP6Address
		proto = fib_types.FIB_API_PATH_NH_PROTO_IP6
		copy(address[:], netIP[:])
		au.SetIP6(address)
	} else {
		var address fib_types.IP4Address
		proto = fib_types.FIB_API_PATH_NH_PROTO_IP4
		copy(address[:], netIP[12:])
		au.SetIP4(address)
	}
	return vpp_abf.FibPathNh{
		Address:            au,
		ViaLabel:           NextHopViaLabelUnset,
		ClassifyTableIndex: ClassifyTableIndexUnset,
	}, proto, nil
}
