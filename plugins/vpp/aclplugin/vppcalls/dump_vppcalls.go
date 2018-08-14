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
	"fmt"
	"net"
	"time"

	"github.com/ligato/cn-infra/logging/logrus"
	acl_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
)

// Protocol types that can occur in ACLs
const (
	ICMPv4Proto = 1
	TCPProto    = 6
	UDPProto    = 17
	ICMPv6Proto = 58
)

// AclDetails is combination of proto-modelled ACL data and VPP provided metadata
type AclDetails struct {
	Acl  *acl.AccessLists_Acl `json:"acl"`
	Meta *AclMeta             `json:"acl_meta"`
}

// AclMeta holds VPP-specific metadata
type AclMeta struct {
	Index uint32 `json:"acl_index"`
	Tag   string `json:"acl_tag"`
}

// ACLToInterface is definition of interface and all ACLs which are bound to
// the interface either as ingress or egress
type ACLToInterface struct {
	SwIfIdx    uint32
	IngressACL []uint32
	EgressACL  []uint32
}

func (handler *aclVppHandler) DumpIPACL(swIfIndices ifaceidx.SwIfIndex) ([]*AclDetails, error) {
	ruleIPData := make(map[AclMeta][]*acl.AccessLists_Acl_Rule)

	// get all ACLs with IP ruleData
	IPRuleACLs, err := handler.DumpIPAcls()
	if len(IPRuleACLs) < 1 || err != nil {
		return nil, err
	}

	// resolve IP rules for every ACL
	// Note: currently ACL may have only IP ruleData or only MAC IP ruleData
	var wasErr error
	for identifier, IPRules := range IPRuleACLs {
		var rulesDetails []*acl.AccessLists_Acl_Rule

		if len(IPRules) > 0 {
			for _, IPRule := range IPRules {
				ruleDetails, err := handler.getIPRuleDetails(IPRule)
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
	interfaceData, err := handler.DumpIPACLInterfaces(indices, swIfIndices)
	if err != nil {
		return nil, err
	}

	var ACLs []*AclDetails
	// Build a list of ACL ruleData with ruleData, interfaces, index and tag (name)
	for identifier, rules := range ruleIPData {
		ACLs = append(ACLs, &AclDetails{
			Acl: &acl.AccessLists_Acl{
				AclName:    identifier.Tag,
				Rules:      rules,
				Interfaces: interfaceData[identifier.Index],
			},
			Meta: &AclMeta{
				Index: identifier.Index,
				Tag:   identifier.Tag,
			},
		})
	}

	return ACLs, wasErr
}

func (handler *aclVppHandler) DumpMACIPACL(swIfIndices ifaceidx.SwIfIndex) ([]*AclDetails, error) {

	ruleMACIPData := make(map[AclMeta][]*acl.AccessLists_Acl_Rule)

	// get all ACLs with MACIP ruleData
	MACIPRuleACLs, err := handler.DumpMacIPAcls()
	if len(MACIPRuleACLs) < 1 || err != nil {
		return nil, err
	}

	// resolve MACIP rules for every ACL
	var wasErr error
	for metadata, MACIPRules := range MACIPRuleACLs {
		var rulesDetails []*acl.AccessLists_Acl_Rule

		if len(MACIPRules) > 0 {
			for _, MACIPRule := range MACIPRules {
				ruleDetails, err := handler.getMACIPRuleDetails(MACIPRule)
				if err != nil {
					return nil, fmt.Errorf("failed to get MACIP Rule %v details: %v", MACIPRule, err)
				}
				rulesDetails = append(rulesDetails, ruleDetails)
			}
		}
		ruleMACIPData[metadata] = rulesDetails
	}

	// Prepare separate list of all active ACL indices on the VPP
	var indices []uint32
	for identifier := range ruleMACIPData {
		indices = append(indices, identifier.Index)
	}

	// Get all ACL indices with ingress and egress interfaces
	interfaceData, err := handler.DumpMACIPACLInterfaces(indices, swIfIndices)
	if err != nil {
		return nil, err
	}

	var ACLs []*AclDetails
	// Build a list of ACL ruleData with ruleData, interfaces, index and tag (name)
	for metadata, rules := range ruleMACIPData {
		ACLs = append(ACLs, &AclDetails{
			Acl: &acl.AccessLists_Acl{
				AclName:    metadata.Tag,
				Rules:      rules,
				Interfaces: interfaceData[metadata.Index],
			},
			Meta: &AclMeta{
				Index: metadata.Index,
				Tag:   metadata.Tag,
			},
		})
	}
	return ACLs, wasErr
}

func (handler *aclVppHandler) DumpIPACLInterfaces(indices []uint32, swIfIndices ifaceidx.SwIfIndex) (map[uint32]*acl.AccessLists_Acl_Interfaces, error) {
	defer func(start time.Time) {
		handler.stopwatch.TimeLog(&acl_api.ACLInterfaceListDump{}).LogTimeEntry(time.Since(start))
	}(time.Now())

	// list of ACL-to-interfaces
	aclsWithInterfaces := make(map[uint32]*acl.AccessLists_Acl_Interfaces)
	if swIfIndices == nil {
		return aclsWithInterfaces, nil
	}

	var interfaceData []*ACLToInterface
	var wasErr error

	msgIP := &acl_api.ACLInterfaceListDump{
		SwIfIndex: 0xffffffff, // dump all
	}
	reqIP := handler.dumpChannel.SendMultiRequest(msgIP)
	for {
		replyIP := &acl_api.ACLInterfaceListDetails{}
		stop, err := reqIP.ReceiveReply(replyIP)
		if stop {
			break
		}
		if err != nil {
			return aclsWithInterfaces, fmt.Errorf("ACL interface list dump reply error: %v", err)
		}

		if replyIP.Count > 0 {
			data := &ACLToInterface{
				SwIfIdx: replyIP.SwIfIndex,
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
					name, _, found := swIfIndices.LookupName(data.SwIfIdx)
					if !found {
						continue
					}
					ingress = append(ingress, name)
				}
			}
			// look for egress
			for _, egressACLIdx := range data.EgressACL {
				if egressACLIdx == aclIdx {
					name, _, found := swIfIndices.LookupName(data.SwIfIdx)
					if !found {
						continue
					}
					egress = append(egress, name)
				}
			}
		}

		aclsWithInterfaces[aclIdx] = &acl.AccessLists_Acl_Interfaces{
			Egress:  egress,
			Ingress: ingress,
		}
	}

	return aclsWithInterfaces, wasErr
}

func (handler *aclVppHandler) DumpMACIPACLInterfaces(indices []uint32, swIfIndices ifaceidx.SwIfIndex) (map[uint32]*acl.AccessLists_Acl_Interfaces, error) {
	defer func(start time.Time) {
		handler.stopwatch.TimeLog(acl_api.ACLInterfaceListDump{}).LogTimeEntry(time.Since(start))
	}(time.Now())

	// list of ACL-to-interfaces
	aclsWithInterfaces := make(map[uint32]*acl.AccessLists_Acl_Interfaces)
	if swIfIndices == nil {
		return aclsWithInterfaces, nil
	}

	var interfaceData []*ACLToInterface
	var wasErr error

	msgMACIP := &acl_api.MacipACLInterfaceListDump{
		SwIfIndex: 0xffffffff, // dump all
	}
	reqMACIP := handler.dumpChannel.SendMultiRequest(msgMACIP)
	for {
		replyMACIP := &acl_api.MacipACLInterfaceListDetails{}
		stop, err := reqMACIP.ReceiveReply(replyMACIP)
		if stop {
			break
		}
		if err != nil {
			return aclsWithInterfaces, fmt.Errorf("MACIP ACL interface list dump reply error: %v", err)
		}
		if replyMACIP.Count > 0 {
			data := &ACLToInterface{
				SwIfIdx: replyMACIP.SwIfIndex,
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
					name, _, found := swIfIndices.LookupName(data.SwIfIdx)
					if !found {
						continue
					}
					ingress = append(ingress, name)
				}
			}
		}
		aclsWithInterfaces[aclIdx] = &acl.AccessLists_Acl_Interfaces{
			Egress:  nil,
			Ingress: ingress,
		}
	}

	return aclsWithInterfaces, wasErr
}

func (handler *aclVppHandler) DumpIPAcls() (map[AclMeta][]acl_api.ACLRule, error) {
	defer func(start time.Time) {
		handler.stopwatch.TimeLog(acl_api.ACLDump{}).LogTimeEntry(time.Since(start))
	}(time.Now())

	aclIPRules := make(map[AclMeta][]acl_api.ACLRule)
	var wasErr error

	req := &acl_api.ACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := handler.dumpChannel.SendMultiRequest(req)
	for {
		msg := &acl_api.ACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return aclIPRules, fmt.Errorf("ACL dump reply error: %v", err)
		}
		if stop {
			break
		}

		metadata := AclMeta{
			Index: msg.ACLIndex,
			Tag:   string(bytes.SplitN(msg.Tag, []byte{0x00}, 2)[0]),
		}

		aclIPRules[metadata] = msg.R
	}

	return aclIPRules, wasErr
}

func (handler *aclVppHandler) DumpMacIPAcls() (map[AclMeta][]acl_api.MacipACLRule, error) {
	defer func(start time.Time) {
		handler.stopwatch.TimeLog(acl_api.MacipACLDump{}).LogTimeEntry(time.Since(start))
	}(time.Now())

	aclMACIPRules := make(map[AclMeta][]acl_api.MacipACLRule)
	var wasErr error

	req := &acl_api.MacipACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := handler.dumpChannel.SendMultiRequest(req)
	for {
		msg := &acl_api.MacipACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return aclMACIPRules, fmt.Errorf("ACL MACIP dump reply error: %v", err)
		}
		if stop {
			break
		}

		metadata := AclMeta{
			Index: msg.ACLIndex,
			Tag:   string(bytes.SplitN(msg.Tag, []byte{0x00}, 2)[0]),
		}

		aclMACIPRules[metadata] = msg.R
	}
	return aclMACIPRules, wasErr
}

func (handler *aclVppHandler) DumpInterfaceIPAcls(swIndex uint32) (acl.AccessLists, error) {
	allACLs := acl.AccessLists{
		Acls: []*acl.AccessLists_Acl{},
	}

	res, err := handler.DumpInterfaceIPACLs(swIndex)
	if err != nil {
		return allACLs, err
	}

	if res.SwIfIndex != swIndex {
		return allACLs, fmt.Errorf("returned interface index %d does not match request", res.SwIfIndex)
	}

	for aidx := range res.Acls {
		ipACL, err := handler.getIPACLDetails(uint32(aidx))
		if err != nil {
			return allACLs, err
		} else {
			allACLs.Acls = append(allACLs.Acls, ipACL)
		}
	}
	return allACLs, nil
}

func (handler *aclVppHandler) DumpInterfaceMACIPAcls(swIndex uint32) (acl.AccessLists, error) {
	allACLs := acl.AccessLists{
		Acls: []*acl.AccessLists_Acl{},
	}

	resMacIp, err := handler.DumpInterfaceMACIPACLs(swIndex)
	if err != nil {
		return allACLs, err
	}

	if resMacIp.SwIfIndex != swIndex {
		return allACLs, fmt.Errorf("returned interface index %d does not match request", resMacIp.SwIfIndex)
	}

	for aidx := range resMacIp.Acls {
		macipACL, err := handler.getMACIPACLDetails(uint32(aidx))
		if err != nil {
			return allACLs, err
		} else {
			allACLs.Acls = append(allACLs.Acls, macipACL)
		}
	}
	return allACLs, nil
}

func (handler *aclVppHandler) DumpInterfaceIPACLs(swIndex uint32) (*acl_api.ACLInterfaceListDetails, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(acl_api.ACLInterfaceListDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &acl_api.ACLInterfaceListDump{
		SwIfIndex: swIndex,
	}

	reply := &acl_api.ACLInterfaceListDetails{}
	if err := handler.dumpChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	return reply, nil
}

func (handler *aclVppHandler) DumpInterfaceMACIPACLs(swIndex uint32) (*acl_api.MacipACLInterfaceListDetails, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(acl_api.MacipACLInterfaceListDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &acl_api.MacipACLInterfaceListDump{
		SwIfIndex: swIndex,
	}

	reply := &acl_api.MacipACLInterfaceListDetails{}
	if err := handler.dumpChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	return reply, nil
}

func (handler *aclVppHandler) DumpInterfaces() ([]*acl_api.ACLInterfaceListDetails, []*acl_api.MacipACLInterfaceListDetails, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(acl_api.ACLInterfaceListDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	msgIPACL := &acl_api.ACLInterfaceListDump{
		SwIfIndex: 0xffffffff, // dump all
	}

	reqIPACL := handler.dumpChannel.SendMultiRequest(msgIPACL)

	var IPaclInterfaces []*acl_api.ACLInterfaceListDetails
	for {
		reply := &acl_api.ACLInterfaceListDetails{}
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

	msgMACIPACL := &acl_api.ACLInterfaceListDump{
		SwIfIndex: 0xffffffff, // dump all
	}

	reqMACIPACL := handler.dumpChannel.SendMultiRequest(msgMACIPACL)

	var MACIPaclInterfaces []*acl_api.MacipACLInterfaceListDetails
	for {
		reply := &acl_api.MacipACLInterfaceListDetails{}
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

func (handler *aclVppHandler) getIPRuleDetails(rule acl_api.ACLRule) (*acl.AccessLists_Acl_Rule, error) {
	// Resolve rule actions
	aclAction, err := handler.resolveRuleAction(rule.IsPermit)
	if err != nil {
		return nil, err
	}

	// Resolve rule matches
	match := &acl.AccessLists_Acl_Rule_Match{
		IpRule: handler.getIPRuleMatches(rule),
	}

	return &acl.AccessLists_Acl_Rule{
		AclAction: aclAction,
		Match:     match,
	}, nil
}

// getIPACLDetails gets details for a given IP ACL from VPP and translates
// them from the binary VPP API format into the ACL Plugin's NB format.
func (handler *aclVppHandler) getIPACLDetails(idx uint32) (aclRule *acl.AccessLists_Acl, err error) {
	req := &acl_api.ACLDump{
		ACLIndex: uint32(idx),
	}

	reply := &acl_api.ACLDetails{}
	if err := handler.dumpChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	var ruleData []*acl.AccessLists_Acl_Rule
	for _, r := range reply.R {
		rule := &acl.AccessLists_Acl_Rule{}

		ipRule, _ := handler.getIPRuleDetails(r)

		match := &acl.AccessLists_Acl_Rule_Match{
			IpRule: ipRule.GetMatch().GetIpRule(),
		}

		aclAction, err := handler.resolveRuleAction(r.IsPermit)
		if err != nil {
			return nil, err
		}

		rule.Match = match
		rule.AclAction = aclAction
		ruleData = append(ruleData, rule)
	}

	return &acl.AccessLists_Acl{Rules: ruleData, AclName: string(bytes.SplitN(reply.Tag, []byte{0x00}, 2)[0])}, nil
}

func (handler *aclVppHandler) getMACIPRuleDetails(rule acl_api.MacipACLRule) (*acl.AccessLists_Acl_Rule, error) {
	// Resolve rule actions
	aclAction, err := handler.resolveRuleAction(rule.IsPermit)
	if err != nil {
		return nil, err
	}

	// Resolve rule matches
	match := &acl.AccessLists_Acl_Rule_Match{
		MacipRule: handler.getMACIPRuleMatches(rule),
	}

	return &acl.AccessLists_Acl_Rule{
		AclAction: aclAction,
		Match:     match,
	}, nil
}

// getMACIPACLDetails gets details for a given MACIP ACL from VPP and translates
// them from the binary VPP API format into the ACL Plugin's NB format.
func (handler *aclVppHandler) getMACIPACLDetails(idx uint32) (aclRule *acl.AccessLists_Acl, err error) {
	req := &acl_api.MacipACLDump{
		ACLIndex: uint32(idx),
	}

	reply := &acl_api.MacipACLDetails{}
	if err := handler.dumpChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	var ruleData []*acl.AccessLists_Acl_Rule
	for _, r := range reply.R {
		rule := &acl.AccessLists_Acl_Rule{}

		ipRule, _ := handler.getMACIPRuleDetails(r)

		match := &acl.AccessLists_Acl_Rule_Match{
			IpRule: ipRule.GetMatch().GetIpRule(),
		}

		aclAction, err := handler.resolveRuleAction(r.IsPermit)
		if err != nil {
			return nil, err
		}

		rule.Match = match
		rule.AclAction = aclAction
		ruleData = append(ruleData, rule)
	}

	return &acl.AccessLists_Acl{Rules: ruleData, AclName: string(bytes.SplitN(reply.Tag, []byte{0x00}, 2)[0])}, nil
}

// getIPRuleMatches translates an IP rule from the binary VPP API format into the
// ACL Plugin's NB format
func (handler *aclVppHandler) getIPRuleMatches(r acl_api.ACLRule) *acl.AccessLists_Acl_Rule_Match_IpRule {
	var srcIP, dstIP string
	if r.IsIpv6 == 1 {
		srcIP = net.IP(r.SrcIPAddr).To16().String()
		dstIP = net.IP(r.DstIPAddr).To16().String()
	} else {
		srcIP = net.IP(r.SrcIPAddr[:4]).To4().String()
		dstIP = net.IP(r.DstIPAddr[:4]).To4().String()
	}

	ipRule := &acl.AccessLists_Acl_Rule_Match_IpRule{
		Ip: &acl.AccessLists_Acl_Rule_Match_IpRule_Ip{
			SourceNetwork:      fmt.Sprintf("%s/%d", srcIP, r.SrcIPPrefixLen),
			DestinationNetwork: fmt.Sprintf("%s/%d", dstIP, r.DstIPPrefixLen),
		},
	}

	switch r.Proto {
	case TCPProto:
		ipRule.Tcp = handler.getTCPMatchRule(r)
	case UDPProto:
		ipRule.Udp = handler.getUDPMatchRule(r)
	case ICMPv4Proto, ICMPv6Proto:
		ipRule.Icmp = handler.getIcmpMatchRule(r)
	}
	return ipRule
}

// getMACIPRuleMatches translates an MACIP rule from the binary VPP API format into the
// ACL Plugin's NB format
func (handler *aclVppHandler) getMACIPRuleMatches(rule acl_api.MacipACLRule) *acl.AccessLists_Acl_Rule_Match_MacIpRule {
	var srcAddr string
	if rule.IsIpv6 == 1 {
		srcAddr = net.IP(rule.SrcIPAddr).To16().String()
	} else {
		srcAddr = net.IP(rule.SrcIPAddr[:4]).To4().String()
	}
	return &acl.AccessLists_Acl_Rule_Match_MacIpRule{
		SourceAddress:        srcAddr,
		SourceAddressPrefix:  uint32(rule.SrcIPPrefixLen),
		SourceMacAddress:     string(rule.SrcMac),
		SourceMacAddressMask: string(rule.SrcMacMask),
	}
}

// getTCPMatchRule translates a TCP match rule from the binary VPP API format
// into the ACL Plugin's NB format
func (handler *aclVppHandler) getTCPMatchRule(r acl_api.ACLRule) *acl.AccessLists_Acl_Rule_Match_IpRule_Tcp {
	dstPortRange := &acl.AccessLists_Acl_Rule_Match_IpRule_PortRange{
		LowerPort: uint32(r.DstportOrIcmpcodeFirst),
		UpperPort: uint32(r.DstportOrIcmpcodeLast),
	}
	srcPortRange := &acl.AccessLists_Acl_Rule_Match_IpRule_PortRange{
		LowerPort: uint32(r.SrcportOrIcmptypeFirst),
		UpperPort: uint32(r.SrcportOrIcmptypeLast),
	}
	tcp := acl.AccessLists_Acl_Rule_Match_IpRule_Tcp{
		DestinationPortRange: dstPortRange,
		SourcePortRange:      srcPortRange,
		TcpFlagsMask:         uint32(r.TCPFlagsMask),
		TcpFlagsValue:        uint32(r.TCPFlagsValue),
	}
	return &tcp
}

// getUDPMatchRule translates a UDP match rule from the binary VPP API format
// into the ACL Plugin's NB format
func (handler *aclVppHandler) getUDPMatchRule(r acl_api.ACLRule) *acl.AccessLists_Acl_Rule_Match_IpRule_Udp {
	dstPortRange := &acl.AccessLists_Acl_Rule_Match_IpRule_PortRange{
		LowerPort: uint32(r.DstportOrIcmpcodeFirst),
		UpperPort: uint32(r.DstportOrIcmpcodeLast),
	}
	srcPortRange := &acl.AccessLists_Acl_Rule_Match_IpRule_PortRange{
		LowerPort: uint32(r.SrcportOrIcmptypeFirst),
		UpperPort: uint32(r.SrcportOrIcmptypeLast),
	}
	udp := acl.AccessLists_Acl_Rule_Match_IpRule_Udp{
		DestinationPortRange: dstPortRange,
		SourcePortRange:      srcPortRange,
	}
	return &udp
}

// getIcmpMatchRule translates an ICMP match rule from the binary VPP API
// format into the ACL Plugin's NB format
func (handler *aclVppHandler) getIcmpMatchRule(r acl_api.ACLRule) *acl.AccessLists_Acl_Rule_Match_IpRule_Icmp {
	icmp := &acl.AccessLists_Acl_Rule_Match_IpRule_Icmp{
		Icmpv6:        r.IsIpv6 > 0,
		IcmpCodeRange: &acl.AccessLists_Acl_Rule_Match_IpRule_Icmp_Range{},
		IcmpTypeRange: &acl.AccessLists_Acl_Rule_Match_IpRule_Icmp_Range{},
	}
	return icmp
}

// Returns rule action representation in model according to the vpp input
func (handler *aclVppHandler) resolveRuleAction(isPermit uint8) (acl.AclAction, error) {
	switch isPermit {
	case 0:
		return acl.AclAction_DENY, nil
	case 1:
		return acl.AclAction_PERMIT, nil
	case 2:
		return acl.AclAction_REFLECT, nil
	default:
		return acl.AclAction_DENY, fmt.Errorf("invalid match rule %d", isPermit)
	}
}
