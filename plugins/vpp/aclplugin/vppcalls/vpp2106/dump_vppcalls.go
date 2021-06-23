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

package vpp2106

import (
	"fmt"
	"net"
	"strings"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/vppcalls"
	vpp_acl "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/acl"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/acl_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"
)

// DumpACL implements ACL handler.
func (h *ACLVppHandler) DumpACL() ([]*vppcalls.ACLDetails, error) {
	ruleIPData := make(map[vppcalls.ACLMeta][]*acl.ACL_Rule)

	// get all ACLs with IP ruleData
	IPRuleACLs, err := h.DumpIPAcls()
	if len(IPRuleACLs) < 1 || err != nil {
		return nil, err
	}

	// resolve IP rules for every ACL
	// Note: currently ACL may have only IP ruleData or only MAC IP ruleData
	var wasErr error
	for identifier, IPRules := range IPRuleACLs {
		var rulesDetails []*acl.ACL_Rule

		if len(IPRules) > 0 {
			for _, IPRule := range IPRules {
				ruleDetails, err := h.getIPRuleDetails(IPRule)
				if err != nil {
					return nil, fmt.Errorf("failed to get IP Rule %v details: %v", IPRule, err)
				}
				rulesDetails = append(rulesDetails, ruleDetails)
			}
		}
		ruleIPData[identifier] = rulesDetails
	}

	// Prepare separate list of all active ACL indices on the VPP
	var indices []uint32
	for identifier := range ruleIPData {
		indices = append(indices, identifier.Index)
	}

	// Get all ACL indices with ingress and egress interfaces
	interfaceData, err := h.DumpACLInterfaces(indices)
	if err != nil {
		return nil, err
	}

	var ACLs []*vppcalls.ACLDetails
	// Build a list of ACL ruleData with ruleData, interfaces, index and tag (name)
	for identifier, rules := range ruleIPData {
		ACLs = append(ACLs, &vppcalls.ACLDetails{
			ACL: &acl.ACL{
				Name:       identifier.Tag,
				Rules:      rules,
				Interfaces: interfaceData[identifier.Index],
			},
			Meta: &vppcalls.ACLMeta{
				Index: identifier.Index,
				Tag:   identifier.Tag,
			},
		})
	}

	return ACLs, wasErr
}

// DumpMACIPACL implements ACL handler.
func (h *ACLVppHandler) DumpMACIPACL() ([]*vppcalls.ACLDetails, error) {
	ruleMACIPData := make(map[vppcalls.ACLMeta][]*acl.ACL_Rule)

	// get all ACLs with MACIP ruleData
	MACIPRuleACLs, err := h.DumpMacIPAcls()
	if err != nil || len(MACIPRuleACLs) == 0 {
		return nil, err
	}

	// resolve MACIP rules for every ACL
	for metadata, MACIPRules := range MACIPRuleACLs {
		var rulesDetails []*acl.ACL_Rule

		for _, MACIPRule := range MACIPRules {
			ruleDetails, err := h.getMACIPRuleDetails(MACIPRule)
			if err != nil {
				return nil, fmt.Errorf("failed to get MACIP Rule %v details: %v", MACIPRule, err)
			}
			rulesDetails = append(rulesDetails, ruleDetails)
		}
		ruleMACIPData[metadata] = rulesDetails
	}

	// Prepare separate list of all active ACL indices on the VPP
	var indices []uint32
	for identifier := range ruleMACIPData {
		indices = append(indices, identifier.Index)
	}

	// Get all ACL indices with ingress and egress interfaces
	interfaceData, err := h.DumpMACIPACLInterfaces(indices)
	if err != nil {
		return nil, err
	}

	var ACLs []*vppcalls.ACLDetails
	// Build a list of ACL ruleData with ruleData, interfaces, index and tag (name)
	for metadata, rules := range ruleMACIPData {
		ACLs = append(ACLs, &vppcalls.ACLDetails{
			ACL: &acl.ACL{
				Name:       metadata.Tag,
				Rules:      rules,
				Interfaces: interfaceData[metadata.Index],
			},
			Meta: &vppcalls.ACLMeta{
				Index: metadata.Index,
				Tag:   metadata.Tag,
			},
		})
	}
	return ACLs, nil
}

// DumpACLInterfaces implements ACL handler.
func (h *ACLVppHandler) DumpACLInterfaces(indices []uint32) (map[uint32]*acl.ACL_Interfaces, error) {
	// list of ACL-to-interfaces
	aclsWithInterfaces := make(map[uint32]*acl.ACL_Interfaces)

	var interfaceData []*vppcalls.ACLToInterface
	var wasErr error

	msgIP := &vpp_acl.ACLInterfaceListDump{
		SwIfIndex: 0xffffffff, // dump all
	}
	reqIP := h.callsChannel.SendMultiRequest(msgIP)
	for {
		replyIP := &vpp_acl.ACLInterfaceListDetails{}
		stop, err := reqIP.ReceiveReply(replyIP)
		if stop {
			break
		}
		if err != nil {
			return aclsWithInterfaces, fmt.Errorf("ACL interface list dump reply error: %v", err)
		}

		if replyIP.Count > 0 {
			data := &vppcalls.ACLToInterface{
				SwIfIdx: uint32(replyIP.SwIfIndex),
			}
			for i, aclIdx := range replyIP.Acls {
				if i < int(replyIP.NInput) {
					data.IngressACL = append(data.IngressACL, aclIdx)
				} else {
					data.EgressACL = append(data.EgressACL, aclIdx)
				}
			}
			interfaceData = append(interfaceData, data)
		}
	}

	// sort interfaces for every ACL
	for _, aclIdx := range indices {
		var ingress []string
		var egress []string
		for _, data := range interfaceData {
			// look for ingress
			for _, ingressACLIdx := range data.IngressACL {
				if ingressACLIdx == aclIdx {
					name, _, found := h.ifIndexes.LookupBySwIfIndex(data.SwIfIdx)
					if !found {
						continue
					}
					ingress = append(ingress, name)
				}
			}
			// look for egress
			for _, egressACLIdx := range data.EgressACL {
				if egressACLIdx == aclIdx {
					name, _, found := h.ifIndexes.LookupBySwIfIndex(data.SwIfIdx)
					if !found {
						continue
					}
					egress = append(egress, name)
				}
			}
		}

		aclsWithInterfaces[aclIdx] = &acl.ACL_Interfaces{
			Egress:  egress,
			Ingress: ingress,
		}
	}

	return aclsWithInterfaces, wasErr
}

// DumpMACIPACLInterfaces implements ACL handler.
func (h *ACLVppHandler) DumpMACIPACLInterfaces(indices []uint32) (map[uint32]*acl.ACL_Interfaces, error) {
	// list of ACL-to-interfaces
	aclsWithInterfaces := make(map[uint32]*acl.ACL_Interfaces)

	var interfaceData []*vppcalls.ACLToInterface

	msgMACIP := &vpp_acl.MacipACLInterfaceListDump{
		SwIfIndex: 0xffffffff, // dump all
	}
	reqMACIP := h.callsChannel.SendMultiRequest(msgMACIP)
	for {
		replyMACIP := &vpp_acl.MacipACLInterfaceListDetails{}
		stop, err := reqMACIP.ReceiveReply(replyMACIP)
		if stop {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("MACIP ACL interface list dump reply error: %v", err)
		}
		if replyMACIP.Count > 0 {
			data := &vppcalls.ACLToInterface{
				SwIfIdx: uint32(replyMACIP.SwIfIndex),
			}
			for _, aclIdx := range replyMACIP.Acls {
				data.IngressACL = append(data.IngressACL, aclIdx)
			}
			interfaceData = append(interfaceData, data)
		}
	}

	for _, aclIdx := range indices {
		var ingress []string
		for _, data := range interfaceData {
			// look for ingress
			for _, ingressACLIdx := range data.IngressACL {
				if ingressACLIdx == aclIdx {
					name, _, found := h.ifIndexes.LookupBySwIfIndex(data.SwIfIdx)
					if !found {
						continue
					}
					ingress = append(ingress, name)
				}
			}
		}
		var ifaces *acl.ACL_Interfaces
		if len(ingress) > 0 {
			ifaces = &acl.ACL_Interfaces{
				Egress:  nil,
				Ingress: ingress,
			}
		}
		aclsWithInterfaces[aclIdx] = ifaces
	}

	return aclsWithInterfaces, nil
}

// DumpIPAcls implements ACL handler.
func (h *ACLVppHandler) DumpIPAcls() (map[vppcalls.ACLMeta][]acl_types.ACLRule, error) {
	aclIPRules := make(map[vppcalls.ACLMeta][]acl_types.ACLRule)
	var wasErr error

	req := &vpp_acl.ACLDump{
		ACLIndex: 0xffffffff,
	}
	reqContext := h.callsChannel.SendMultiRequest(req)
	for {
		msg := &vpp_acl.ACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return aclIPRules, fmt.Errorf("ACL dump reply error: %v", err)
		}
		if stop {
			break
		}

		metadata := vppcalls.ACLMeta{
			Index: msg.ACLIndex,
			Tag:   strings.Trim(msg.Tag, "\x00"),
		}

		aclIPRules[metadata] = msg.R
	}

	return aclIPRules, wasErr
}

// DumpMacIPAcls implements ACL handler.
func (h *ACLVppHandler) DumpMacIPAcls() (map[vppcalls.ACLMeta][]acl_types.MacipACLRule, error) {
	aclMACIPRules := make(map[vppcalls.ACLMeta][]acl_types.MacipACLRule)

	req := &vpp_acl.MacipACLDump{
		ACLIndex: 0xffffffff,
	}
	reqContext := h.callsChannel.SendMultiRequest(req)
	for {
		msg := &vpp_acl.MacipACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("ACL MACIP dump reply error: %v", err)
		}
		if stop {
			break
		}

		metadata := vppcalls.ACLMeta{
			Index: msg.ACLIndex,
			Tag:   strings.Trim(msg.Tag, "\x00"),
		}

		aclMACIPRules[metadata] = msg.R
	}
	return aclMACIPRules, nil
}

// DumpInterfaceACLs implements ACL handler.
func (h *ACLVppHandler) DumpInterfaceACLs(swIndex uint32) (acls []*acl.ACL, err error) {
	res, err := h.DumpInterfaceACLList(swIndex)
	if err != nil {
		return nil, err
	}

	if uint32(res.SwIfIndex) != swIndex {
		return nil, fmt.Errorf("returned interface index %d does not match request", res.SwIfIndex)
	}

	for aidx := range res.Acls {
		ipACL, err := h.getIPACLDetails(uint32(aidx))
		if err != nil {
			return nil, err
		}
		acls = append(acls, ipACL)
	}
	return acls, nil
}

// DumpInterfaceMACIPACLs implements ACL handler.
func (h *ACLVppHandler) DumpInterfaceMACIPACLs(swIndex uint32) (acls []*acl.ACL, err error) {
	resMacIP, err := h.DumpInterfaceMACIPACLList(swIndex)
	if err != nil {
		return nil, err
	}

	if uint32(resMacIP.SwIfIndex) != swIndex {
		return nil, fmt.Errorf("returned interface index %d does not match request", resMacIP.SwIfIndex)
	}

	for aidx := range resMacIP.Acls {
		macipACL, err := h.getMACIPACLDetails(uint32(aidx))
		if err != nil {
			return nil, err
		}
		acls = append(acls, macipACL)
	}
	return acls, nil
}

// DumpInterfaceACLList implements ACL handler.
func (h *ACLVppHandler) DumpInterfaceACLList(swIndex uint32) (*vpp_acl.ACLInterfaceListDetails, error) {
	req := &vpp_acl.ACLInterfaceListDump{
		SwIfIndex: interface_types.InterfaceIndex(swIndex),
	}
	reply := &vpp_acl.ACLInterfaceListDetails{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	return reply, nil
}

// DumpInterfaceMACIPACLList implements ACL handler.
func (h *ACLVppHandler) DumpInterfaceMACIPACLList(swIndex uint32) (*vpp_acl.MacipACLInterfaceListDetails, error) {
	req := &vpp_acl.MacipACLInterfaceListDump{
		SwIfIndex: interface_types.InterfaceIndex(swIndex),
	}
	reply := &vpp_acl.MacipACLInterfaceListDetails{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	return reply, nil
}

// DumpInterfacesLists implements ACL handler.
func (h *ACLVppHandler) DumpInterfacesLists() ([]*vpp_acl.ACLInterfaceListDetails, []*vpp_acl.MacipACLInterfaceListDetails, error) {
	msgIPACL := &vpp_acl.ACLInterfaceListDump{
		SwIfIndex: 0xffffffff, // dump all
	}

	reqIPACL := h.callsChannel.SendMultiRequest(msgIPACL)

	var IPaclInterfaces []*vpp_acl.ACLInterfaceListDetails
	for {
		reply := &vpp_acl.ACLInterfaceListDetails{}
		stop, err := reqIPACL.ReceiveReply(reply)
		if stop {
			break
		}
		if err != nil {
			logrus.DefaultLogger().Error(err)
			return nil, nil, err
		}
		IPaclInterfaces = append(IPaclInterfaces, reply)
	}

	msgMACIPACL := &vpp_acl.ACLInterfaceListDump{
		SwIfIndex: 0xffffffff, // dump all
	}

	reqMACIPACL := h.callsChannel.SendMultiRequest(msgMACIPACL)

	var MACIPaclInterfaces []*vpp_acl.MacipACLInterfaceListDetails
	for {
		reply := &vpp_acl.MacipACLInterfaceListDetails{}
		stop, err := reqMACIPACL.ReceiveReply(reply)
		if stop {
			break
		}
		if err != nil {
			logrus.DefaultLogger().Error(err)
			return nil, nil, err
		}
		MACIPaclInterfaces = append(MACIPaclInterfaces, reply)
	}

	return IPaclInterfaces, MACIPaclInterfaces, nil
}

func (h *ACLVppHandler) getIPRuleDetails(rule acl_types.ACLRule) (*acl.ACL_Rule, error) {
	// Resolve rule actions
	aclAction, err := h.resolveRuleAction(rule.IsPermit)
	if err != nil {
		return nil, err
	}

	return &acl.ACL_Rule{
		Action: aclAction,
		IpRule: h.getIPRuleMatches(rule),
	}, nil
}

// getIPACLDetails gets details for a given IP ACL from VPP and translates
// them from the binary VPP API format into the ACL Plugin's NB format.
func (h *ACLVppHandler) getIPACLDetails(idx uint32) (aclRule *acl.ACL, err error) {
	req := &vpp_acl.ACLDump{
		ACLIndex: uint32(idx),
	}

	reply := &vpp_acl.ACLDetails{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	var ruleData []*acl.ACL_Rule
	for _, r := range reply.R {
		rule := &acl.ACL_Rule{}

		ipRule, err := h.getIPRuleDetails(r)
		if err != nil {
			return nil, err
		}

		aclAction, err := h.resolveRuleAction(r.IsPermit)
		if err != nil {
			return nil, err
		}

		rule.IpRule = ipRule.GetIpRule()
		rule.Action = aclAction
		ruleData = append(ruleData, rule)
	}

	return &acl.ACL{Rules: ruleData, Name: strings.Trim(reply.Tag, "\x00")}, nil
}

func (h *ACLVppHandler) getMACIPRuleDetails(rule acl_types.MacipACLRule) (*acl.ACL_Rule, error) {
	// Resolve rule actions
	aclAction, err := h.resolveRuleAction(rule.IsPermit)
	if err != nil {
		return nil, err
	}

	return &acl.ACL_Rule{
		Action:    aclAction,
		MacipRule: h.getMACIPRuleMatches(rule),
	}, nil
}

// getMACIPACLDetails gets details for a given MACIP ACL from VPP and translates
// them from the binary VPP API format into the ACL Plugin's NB format.
func (h *ACLVppHandler) getMACIPACLDetails(idx uint32) (aclRule *acl.ACL, err error) {
	req := &vpp_acl.MacipACLDump{
		ACLIndex: uint32(idx),
	}

	reply := &vpp_acl.MacipACLDetails{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	var ruleData []*acl.ACL_Rule
	for _, r := range reply.R {
		rule := &acl.ACL_Rule{}

		ipRule, err := h.getMACIPRuleDetails(r)
		if err != nil {
			return nil, err
		}

		aclAction, err := h.resolveRuleAction(r.IsPermit)
		if err != nil {
			return nil, err
		}

		rule.IpRule = ipRule.GetIpRule()
		rule.Action = aclAction
		ruleData = append(ruleData, rule)
	}

	return &acl.ACL{Rules: ruleData, Name: strings.Trim(reply.Tag, "\x00")}, nil
}

// getIPRuleMatches translates an IP rule from the binary VPP API format into the
// ACL Plugin's NB format
func (h *ACLVppHandler) getIPRuleMatches(r acl_types.ACLRule) *acl.ACL_Rule_IpRule {
	srcNet := prefixToString(r.SrcPrefix)
	dstNet := prefixToString(r.DstPrefix)

	ipRule := &acl.ACL_Rule_IpRule{
		Ip: &acl.ACL_Rule_IpRule_Ip{
			SourceNetwork:      srcNet,
			DestinationNetwork: dstNet,
			Protocol:           uint32(r.Proto),
		},
	}

	switch r.Proto {
	case vppcalls.TCPProto:
		ipRule.Tcp = h.getTCPMatchRule(r)
	case vppcalls.UDPProto:
		ipRule.Udp = h.getUDPMatchRule(r)
	case vppcalls.ICMPv4Proto, vppcalls.ICMPv6Proto:
		ipRule.Icmp = h.getIcmpMatchRule(r)
	}
	return ipRule
}

// getMACIPRuleMatches translates an MACIP rule from the binary VPP API format into the
// ACL Plugin's NB format
func (h *ACLVppHandler) getMACIPRuleMatches(rule acl_types.MacipACLRule) *acl.ACL_Rule_MacIpRule {
	srcAddr := addressToIP(rule.SrcPrefix.Address)
	srcMacAddr := net.HardwareAddr(rule.SrcMac[:])
	srcMacAddrMask := net.HardwareAddr(rule.SrcMacMask[:])
	return &acl.ACL_Rule_MacIpRule{
		SourceAddress:        srcAddr.String(),
		SourceAddressPrefix:  uint32(rule.SrcPrefix.Len),
		SourceMacAddress:     srcMacAddr.String(),
		SourceMacAddressMask: srcMacAddrMask.String(),
	}
}

// getTCPMatchRule translates a TCP match rule from the binary VPP API format
// into the ACL Plugin's NB format
func (h *ACLVppHandler) getTCPMatchRule(r acl_types.ACLRule) *acl.ACL_Rule_IpRule_Tcp {
	dstPortRange := &acl.ACL_Rule_IpRule_PortRange{
		LowerPort: uint32(r.DstportOrIcmpcodeFirst),
		UpperPort: uint32(r.DstportOrIcmpcodeLast),
	}
	srcPortRange := &acl.ACL_Rule_IpRule_PortRange{
		LowerPort: uint32(r.SrcportOrIcmptypeFirst),
		UpperPort: uint32(r.SrcportOrIcmptypeLast),
	}
	tcp := acl.ACL_Rule_IpRule_Tcp{
		DestinationPortRange: dstPortRange,
		SourcePortRange:      srcPortRange,
		TcpFlagsMask:         uint32(r.TCPFlagsMask),
		TcpFlagsValue:        uint32(r.TCPFlagsValue),
	}
	return &tcp
}

// getUDPMatchRule translates a UDP match rule from the binary VPP API format
// into the ACL Plugin's NB format
func (h *ACLVppHandler) getUDPMatchRule(r acl_types.ACLRule) *acl.ACL_Rule_IpRule_Udp {
	dstPortRange := &acl.ACL_Rule_IpRule_PortRange{
		LowerPort: uint32(r.DstportOrIcmpcodeFirst),
		UpperPort: uint32(r.DstportOrIcmpcodeLast),
	}
	srcPortRange := &acl.ACL_Rule_IpRule_PortRange{
		LowerPort: uint32(r.SrcportOrIcmptypeFirst),
		UpperPort: uint32(r.SrcportOrIcmptypeLast),
	}
	udp := acl.ACL_Rule_IpRule_Udp{
		DestinationPortRange: dstPortRange,
		SourcePortRange:      srcPortRange,
	}
	return &udp
}

// getIcmpMatchRule translates an ICMP match rule from the binary VPP API
// format into the ACL Plugin's NB format
func (h *ACLVppHandler) getIcmpMatchRule(r acl_types.ACLRule) *acl.ACL_Rule_IpRule_Icmp {
	icmp := &acl.ACL_Rule_IpRule_Icmp{
		Icmpv6: r.Proto == ip_types.IP_API_PROTO_ICMP6,
		IcmpCodeRange: &acl.ACL_Rule_IpRule_Icmp_Range{
			First: uint32(r.DstportOrIcmpcodeFirst),
			Last:  uint32(r.DstportOrIcmpcodeLast),
		},
		IcmpTypeRange: &acl.ACL_Rule_IpRule_Icmp_Range{
			First: uint32(r.SrcportOrIcmptypeFirst),
			Last:  uint32(r.SrcportOrIcmptypeLast),
		},
	}
	return icmp
}

// Returns rule action representation in model according to the vpp input
func (h *ACLVppHandler) resolveRuleAction(isPermit acl_types.ACLAction) (acl.ACL_Rule_Action, error) {
	switch isPermit {
	case acl_types.ACL_ACTION_API_DENY:
		return acl.ACL_Rule_DENY, nil
	case acl_types.ACL_ACTION_API_PERMIT:
		return acl.ACL_Rule_PERMIT, nil
	case acl_types.ACL_ACTION_API_PERMIT_REFLECT:
		return acl.ACL_Rule_REFLECT, nil
	default:
		return acl.ACL_Rule_DENY, fmt.Errorf("invalid match rule %v", isPermit)
	}
}
