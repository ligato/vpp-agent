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

package vppdump

import (
	"bytes"
	"fmt"
	"net"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
)

// ACLIdentifier contains fields for ACL index and Tag (used as a name in the configuration)
type ACLIdentifier struct {
	ACLIndex uint32 `json:"acl_index"`
	Tag      string `json:"acl_tag"`
}

// ACLEntry is cumulative object with ACL identification and details with all rules and
// interfaces belonging the the ACL
type ACLEntry struct {
	ACLIdentifier
	ACLDetails *acl.AccessLists_Acl `json:"acl_details"`
}

// ACLRules is partial ACL definition with all rules
type ACLRules struct {
	ACLIdentifier
	Rules []*acl.AccessLists_Acl_Rule
}

// ACLToIPRules is a partial ACL definition with IP rules
type ACLToIPRules struct {
	ACLIdentifier
	Rules []acl_api.ACLRule
	Count uint32
}

// ACLToMACIPRules is a partial ACL definition with MAC IP rules
type ACLToMACIPRules struct {
	ACLIdentifier
	Rules []acl_api.MacipACLRule
	Count uint32
}

// ACLToInterface is definition of interface and all ACLs which are bound to
// the interface either as ingress or egress
type ACLToInterface struct {
	SwIfIdx    uint32
	IngressACL []uint32
	EgressACL  []uint32
}

//
func DumpACLs(log logging.Logger, swIfIndices ifaceidx.SwIfIndex, vppChannel *govppapi.Channel,
	timeLog measure.StopWatchEntry) ([]*ACLEntry, error) {

	var ACLs []*ACLEntry

	aclRuleData, err := DumpACLRules(log, vppChannel, timeLog)
	if err != nil {
		return nil, err
	}
	// Get list of all ACL indices
	var aclIndices []uint32
	for _, aclRules := range aclRuleData {
		aclIndices = append(aclIndices, aclRules.ACLIndex)
	}
	aclInterfaceData, err := DumpACLInterfaces(aclIndices, swIfIndices, log, vppChannel, timeLog)
	if err != nil {
		return nil, err
	}

	for _, aclRules := range aclRuleData {
		aclDetails := &acl.AccessLists_Acl{
			Rules:      aclRules.Rules,
			Interfaces: aclInterfaceData[aclRules.ACLIndex],
		}

		ACLs = append(ACLs, &ACLEntry{
			ACLIdentifier: ACLIdentifier{
				ACLIndex: aclRules.ACLIndex,
				Tag:      aclRules.Tag,
			},
			ACLDetails: aclDetails,
		})
	}

	return ACLs, nil
}

// DumpACLRules returns a map of ACL indices with appropriate list of rules
func DumpACLRules(log logging.Logger, vppChannel *govppapi.Channel,
	timeLog measure.StopWatchEntry) ([]*ACLRules, error) {
	var aclRules []*ACLRules

	// get all ACLs with IP rules
	aclsWithIPRules, err := DumpIPAcls(log, vppChannel, timeLog)
	if err != nil {
		return nil, err
	}
	// get all ACLs with MACIP rules
	aclsWithMACIPRules, err := DumpMacIPAcls(log, vppChannel, timeLog)
	if err != nil {
		return nil, err
	}

	// resolve IP rules for every ACL
	// Note: currently ACL may have only IP rules or only MAC IP rules
	var wasErr error
	for _, aclToIPRule := range aclsWithIPRules {
		var rules []*acl.AccessLists_Acl_Rule

		if aclToIPRule.Count > 0 {
			for _, IPRule := range aclToIPRule.Rules {
				ruleDetails, err := getIPRuleDetails(IPRule)
				if err != nil {
					wasErr = err
				}
				rules = append(rules, ruleDetails)
			}
		}
		aclRules = append(aclRules, &ACLRules{
			ACLIdentifier: ACLIdentifier{
				ACLIndex: aclToIPRule.ACLIndex,
				Tag:      aclToIPRule.Tag,
			},
			Rules: rules,
		})
	}

	// resolve MACIP rules for every ACL
	for _, aclToMACIPRule := range aclsWithMACIPRules {
		var rules []*acl.AccessLists_Acl_Rule

		if aclToMACIPRule.Count > 0 {
			for _, MACIPRule := range aclToMACIPRule.Rules {
				ruleDetails, err := getMACIPRuleDetails(MACIPRule)
				if err != nil {
					wasErr = err
				}
				rules = append(rules, ruleDetails)
			}
		}
		aclRules = append(aclRules, &ACLRules{
			ACLIdentifier: ACLIdentifier{
				ACLIndex: aclToMACIPRule.ACLIndex,
				Tag:      aclToMACIPRule.Tag,
			},
			Rules: rules,
		})
	}

	return aclRules, wasErr
}

func DumpACLInterfaces(aclIndices []uint32, swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChannel *govppapi.Channel,
	timeLog measure.StopWatchEntry) (map[uint32]*acl.AccessLists_Acl_Interfaces, error) {
	// ACLInterfaceListDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	aclsWithInterfaces := make(map[uint32]*acl.AccessLists_Acl_Interfaces)
	if swIfIndices == nil {
		return aclsWithInterfaces, nil
	}

	var interfaceData []*ACLToInterface

	msg := &acl_api.ACLInterfaceListDump{}
	msg.SwIfIndex = 0xffffffff // dump all

	req := vppChannel.SendMultiRequest(msg)

	var wasErr error
	for {
		reply := &acl_api.ACLInterfaceListDetails{}
		stop, err := req.ReceiveReply(reply)
		if stop {
			break
		}
		if err != nil {
			wasErr = err
		}

		data := &ACLToInterface{}
		data.SwIfIdx = reply.SwIfIndex
		for i, aclIdx := range reply.Acls {
			if i <= int(reply.NInput) {
				data.IngressACL = append(data.IngressACL, aclIdx)
			} else {
				data.EgressACL = append(data.EgressACL, aclIdx)
			}
		}

		interfaceData = append(interfaceData, data)
	}

	for _, aclIdx := range aclIndices {
		var ingress []string
		var egress []string
		// sort out interfaces for the ACL
		for _, data := range interfaceData {
			// look for ingress
			for _, ingressACL := range data.IngressACL {
				if ingressACL == aclIdx {
					name, _, found := swIfIndices.LookupName(data.SwIfIdx)
					if !found {
						log.Warnf("ACL requires ingress interface with Idx %v which was not found in the mapping", data.SwIfIdx)
						continue
					}
					ingress = append(ingress, name)
				}
			}
			// look for egress
			for _, egressACL := range data.EgressACL {
				if egressACL == aclIdx {
					name, _, found := swIfIndices.LookupName(data.SwIfIdx)
					if !found {
						log.Warnf("ACL requires egress interface with Idx %v which was not found in the mapping", data.SwIfIdx)
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

// DumpIPAcls returns a list of all configured ACL with IP-type rules.
func DumpIPAcls(log logging.Logger, vch *govppapi.Channel, timeLog measure.StopWatchEntry) ([]*ACLToIPRules, error) {
	// ACLDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	var aclsWithRules []*ACLToIPRules
	var wasErr error

	req := &acl_api.ACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := vch.SendMultiRequest(req)
	for {
		msg := &acl_api.ACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			wasErr = err
		}
		if stop {
			break
		}

		entry := &ACLToIPRules{
			ACLIdentifier: ACLIdentifier{
				ACLIndex: msg.ACLIndex,
				Tag:      string(bytes.Trim(msg.Tag, "\x00")),
			},
			Rules: msg.R,
			Count: msg.Count,
		}
		aclsWithRules = append(aclsWithRules, entry)
	}

	return aclsWithRules, wasErr
}

// DumpMacIPAcls returns a list of all configured ACL with IPMAC-type rules.
func DumpMacIPAcls(log logging.Logger, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) ([]*ACLToMACIPRules, error) {
	// MacipACLDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	var aclsWithRules []*ACLToMACIPRules
	var wasErr error

	req := &acl_api.MacipACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := vppChannel.SendMultiRequest(req)
	for {
		msg := &acl_api.MacipACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			wasErr = err
		}
		if stop {
			break
		}

		entry := &ACLToMACIPRules{
			ACLIdentifier: ACLIdentifier{
				ACLIndex: msg.ACLIndex,
				Tag:      string(bytes.Trim(msg.Tag, "\x00")),
			},
			Rules: msg.R,
			Count: msg.Count,
		}

		aclsWithRules = append(aclsWithRules, entry)
	}
	return aclsWithRules, wasErr
}

// DumpInterfaceAcls finds interface in VPP and returns its ACL configuration
func DumpInterfaceAcls(log logging.Logger, swIndex uint32, vppChannel *govppapi.Channel,
	timeLog measure.StopWatchEntry) (acl.AccessLists, uint8, error) {

	log.Info("DumpInterfaceAcls")
	// ACLInterfaceListDump time measurement
	alAcls := acl.AccessLists{
		Acl: []*acl.AccessLists_Acl{},
	}

	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	res, err := vppcalls.DumpInterface(swIndex, vppChannel, timeLog)
	log.Infof("Res: %+v\n", res)
	if err != nil {
		return alAcls, 0, err
	}

	if res.SwIfIndex != swIndex {
		return alAcls, 0, fmt.Errorf(fmt.Sprintf("Returned interface index %d does not match request",
			res.SwIfIndex))
	}

	for aidx := range res.Acls {
		ipACL, err := getIPACLDetails(vppChannel, uint32(aidx))
		if err != nil {
			log.Error(err)
		} else {
			alAcls.Acl = append(alAcls.Acl, ipACL)
		}
	}
	return alAcls, res.NInput, nil
}

func getIPRuleDetails(rule acl_api.ACLRule) (*acl.AccessLists_Acl_Rule, error) {
	// Resolve rule actions
	actions := &acl.AccessLists_Acl_Rule_Actions{}
	switch rule.IsPermit {
	case 0:
		actions.AclAction = acl.AclAction_DENY
	case 1:
		actions.AclAction = acl.AclAction_PERMIT
	case 2:
		actions.AclAction = acl.AclAction_REFLECT
	default:
		return nil, fmt.Errorf("invalid match rule %d", rule.IsPermit)
	}

	// Resolve rule matches
	matches := &acl.AccessLists_Acl_Rule_Matches{
		IpRule: getIPRuleMatches(rule),
	}

	return &acl.AccessLists_Acl_Rule{
		Actions: actions,
		Matches: matches,
	}, nil
}

func getMACIPRuleDetails(rule acl_api.MacipACLRule) (*acl.AccessLists_Acl_Rule, error) {
	// Resolve rule actions
	actions := &acl.AccessLists_Acl_Rule_Actions{}
	switch rule.IsPermit {
	case 0:
		actions.AclAction = acl.AclAction_DENY
	case 1:
		actions.AclAction = acl.AclAction_PERMIT
	case 2:
		actions.AclAction = acl.AclAction_REFLECT
	default:
		return nil, fmt.Errorf("invalid match rule %d", rule.IsPermit)
	}

	// Resolve rule matches
	matches := &acl.AccessLists_Acl_Rule_Matches{
		MacipRule: getMACIPRuleMatches(rule),
	}

	return &acl.AccessLists_Acl_Rule{
		Actions: actions,
		Matches: matches,
	}, nil
}

//getIPACLDetails gets details for a given IP ACL from VPP and translates
//them from the binary VPP API format into the ACL Plugin's NB format.
func getIPACLDetails(vppChannel *govppapi.Channel, idx uint32) (*acl.AccessLists_Acl, error) {
	req := &acl_api.ACLDump{}
	req.ACLIndex = uint32(idx)

	reply := &acl_api.ACLDetails{}
	err := vppChannel.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return nil, err
	}

	rules := make([]*acl.AccessLists_Acl_Rule, 0)
	for _, r := range reply.R {
		rule := acl.AccessLists_Acl_Rule{}

		ipRule, _ := getIPRuleDetails(r)

		matches := acl.AccessLists_Acl_Rule_Matches{
			IpRule: ipRule.Matches.GetIpRule(),
		}

		actions := acl.AccessLists_Acl_Rule_Actions{}
		switch r.IsPermit {
		case 0:
			actions.AclAction = acl.AclAction_DENY
		case 1:
			actions.AclAction = acl.AclAction_PERMIT
		case 2:
			actions.AclAction = acl.AclAction_REFLECT
		default:
			return nil, fmt.Errorf("invalid match rule %d", r.IsPermit)
		}

		rule.Matches = &matches
		rule.Actions = &actions
		rules = append(rules, &rule)
	}

	return &acl.AccessLists_Acl{Rules: rules, AclName: string(bytes.Trim(reply.Tag, "\x00"))}, nil
}

// getIPRuleMatches translates an IP rule from the binary VPP API format into the
// ACL Plugin's NB format
func getIPRuleMatches(r acl_api.ACLRule) *acl.AccessLists_Acl_Rule_Matches_IpRule {
	ipRule := acl.AccessLists_Acl_Rule_Matches_IpRule{}

	ip := acl.AccessLists_Acl_Rule_Matches_IpRule_Ip{
		SourceNetwork:      fmt.Sprintf("%v/%d", decodeIPv4Address(r.SrcIPAddr), r.SrcIPPrefixLen),
		DestinationNetwork: fmt.Sprintf("%s/%d", decodeIPv4Address(r.DstIPAddr), r.DstIPPrefixLen),
	}
	ipRule.Ip = &ip

	switch r.Proto {
	case vppcalls.TCPProto:
		ipRule.Tcp = getTCPMatchRule(r)
		break
	case vppcalls.UDPProto:
		ipRule.Udp = getUDPMatchRule(r)
		break
	case vppcalls.Icmpv4Proto:
	case vppcalls.Icmpv6Proto:
		ipRule.Icmp = getIcmpMatchRule(r)
		break
	default:
		ipRule.Other = &acl.AccessLists_Acl_Rule_Matches_IpRule_Other{
			Protocol: uint32(r.Proto),
		}
	}
	return &ipRule
}

func getMACIPRuleMatches(rule acl_api.MacipACLRule) *acl.AccessLists_Acl_Rule_Matches_MacIpRule {
	return &acl.AccessLists_Acl_Rule_Matches_MacIpRule{
		SourceAddress:        decodeIPv4Address(rule.SrcIPAddr),
		SourceAddressPrefix:  uint32(rule.SrcIPPrefixLen),
		SourceMacAddress:     string(rule.SrcMac),
		SourceMacAddressMask: string(rule.SrcMacMask),
	}
}

// getTCPMatchRule translates a TCP match rule from the binary VPP API format
// into the ACL Plugin's NB format
func getTCPMatchRule(r acl_api.ACLRule) *acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp {
	dstPortRange := acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp_DestinationPortRange{
		LowerPort: uint32(r.DstportOrIcmpcodeFirst),
		UpperPort: uint32(r.DstportOrIcmpcodeLast),
	}
	srcPortRange := acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp_SourcePortRange{
		LowerPort: uint32(r.SrcportOrIcmptypeFirst),
		UpperPort: uint32(r.SrcportOrIcmptypeLast),
	}
	tcp := acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp{
		DestinationPortRange: &dstPortRange,
		SourcePortRange:      &srcPortRange,
		TcpFlagsMask:         uint32(r.TCPFlagsMask),
		TcpFlagsValue:        uint32(r.TCPFlagsValue),
	}
	return &tcp
}

// getUDPMatchRule translates a UDP match rule from the binary VPP API format
// into the ACL Plugin's NB format
func getUDPMatchRule(r acl_api.ACLRule) *acl.AccessLists_Acl_Rule_Matches_IpRule_Udp {
	dstPortRange := acl.AccessLists_Acl_Rule_Matches_IpRule_Udp_DestinationPortRange{
		LowerPort: uint32(r.DstportOrIcmpcodeFirst),
		UpperPort: uint32(r.DstportOrIcmpcodeLast),
	}
	srcPortRange := acl.AccessLists_Acl_Rule_Matches_IpRule_Udp_SourcePortRange{
		LowerPort: uint32(r.SrcportOrIcmptypeFirst),
		UpperPort: uint32(r.SrcportOrIcmptypeLast),
	}
	udp := acl.AccessLists_Acl_Rule_Matches_IpRule_Udp{
		DestinationPortRange: &dstPortRange,
		SourcePortRange:      &srcPortRange,
	}
	return &udp
}

// getIcmpMatchRule translates an ICMP match rule from the binary VPP API
// format into the ACL Plugin's NB format
func getIcmpMatchRule(r acl_api.ACLRule) *acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp {
	codeRange := acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp_IcmpCodeRange{}
	typeRange := acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp_IcmpTypeRange{}

	icmp := acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp{
		Icmpv6:        r.IsIpv6 > 0,
		IcmpCodeRange: &codeRange,
		IcmpTypeRange: &typeRange,
	}
	return &icmp
}

// decodeIPv4Address converts first four elements of provided byte array to IPv4 address
// as a string.
func decodeIPv4Address(addr []byte) string {
	var ipv4 []byte
	for i, octet := range addr {
		ipv4 = append(ipv4, octet)
		if i >= 3 {
			break
		}
	}
	var IPv4Addr net.IP = ipv4

	return IPv4Addr.To4().String()
}
