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
	acl_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
)

// ACLIdentifier contains fields for ACL index and Tag (used as a name in the configuration)
type ACLIdentifier struct {
	ACLIndex uint32 `json:"acl_index"`
	Tag      string `json:"acl_tag"`
}

// ACLEntry is cumulative object with ACL identification and details with all ruleData and
// interfaces belonging to the ACL
type ACLEntry struct {
	Identifier *ACLIdentifier
	ACLDetails *acl.AccessLists_Acl `json:"acl_details"`
}

// ACLToInterface is definition of interface and all ACLs which are bound to
// the interface either as ingress or egress
type ACLToInterface struct {
	SwIfIdx    uint32
	IngressACL []uint32
	EgressACL  []uint32
}

// DumpACLs return a list of all configured ACLs including ruleData and interfaces
func DumpACLs(log logging.Logger, swIfIndices ifaceidx.SwIfIndex, vppChannel *govppapi.Channel,
	timeLog measure.StopWatchEntry) ([]*ACLEntry, error) {
	log.Debugf("dumping ACLs")

	var ACLs []*ACLEntry

	// Get all IP and MACIP ruleData with particular ACL index
	ruleData, err := DumpACLRules(log, vppChannel, timeLog)
	if err != nil {
		return nil, err
	}
	// Prepare separate list of all active ACL indices on the VPP
	var indices []uint32
	for identifier := range ruleData {
		indices = append(indices, identifier.ACLIndex)
	}

	// Get all ACL indices with ingress and egress interfaces
	interfaceData, err := DumpACLInterfaces(indices, swIfIndices, log, vppChannel, timeLog)
	if err != nil {
		return nil, err
	}

	// Build a list of ACL ruleData with ruleData, interfaces, index and tag (name)
	for identifier, rules := range ruleData {
		ACLs = append(ACLs, &ACLEntry{
			Identifier: &identifier,
			ACLDetails: &acl.AccessLists_Acl{
				Rules:      rules,
				Interfaces: interfaceData[identifier.ACLIndex],
			},
		})
	}

	log.Debugf("dumped %d ACLs", len(ACLs))

	return ACLs, nil
}

// DumpACLRules returns all ruleData for every ACL
func DumpACLRules(log logging.Logger, vppChannel *govppapi.Channel,
	timeLog measure.StopWatchEntry) (map[ACLIdentifier][]*acl.AccessLists_Acl_Rule, error) {
	// rule map will be returned
	rules := make(map[ACLIdentifier][]*acl.AccessLists_Acl_Rule)

	// get all ACLs with IP ruleData
	IPRuleACLs, err := DumpIPAcls(log, vppChannel, timeLog)
	if err != nil {
		return nil, err
	}
	// get all ACLs with MACIP ruleData
	MACIPRuleACLs, err := DumpMacIPAcls(log, vppChannel, timeLog)
	if err != nil {
		return nil, err
	}

	// resolve IP rules for every ACL
	// Note: currently ACL may have only IP ruleData or only MAC IP ruleData
	var wasErr error
	for identifier, IPRules := range IPRuleACLs {
		var rulesDetails []*acl.AccessLists_Acl_Rule

		if len(IPRules) > 0 {
			for _, IPRule := range IPRules {
				ruleDetails, err := getIPRuleDetails(IPRule)
				if err != nil {
					log.Error(err)
					wasErr = err
				}
				rulesDetails = append(rulesDetails, ruleDetails)
			}
		}
		rules[identifier] = rulesDetails
	}

	// resolve MACIP rules for every ACL
	for identifier, MACIPRules := range MACIPRuleACLs {
		var rulesDetails []*acl.AccessLists_Acl_Rule

		if len(MACIPRules) > 0 {
			for _, MACIPRule := range MACIPRules {
				ruleDetails, err := getMACIPRuleDetails(MACIPRule)
				if err != nil {
					log.Error(err)
					wasErr = err
				}
				rulesDetails = append(rulesDetails, ruleDetails)
			}
		}
		rules[identifier] = rulesDetails
	}

	return rules, wasErr
}

// DumpACLInterfaces returns a map of ACL indices with interfaces
func DumpACLInterfaces(indices []uint32, swIfIndices ifaceidx.SwIfIndex, log logging.Logger, vppChannel *govppapi.Channel,
	timeLog measure.StopWatchEntry) (map[uint32]*acl.AccessLists_Acl_Interfaces, error) {
	// ACLInterfaceListDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// list of ACL-to-interfaces
	aclsWithInterfaces := make(map[uint32]*acl.AccessLists_Acl_Interfaces)
	if swIfIndices == nil {
		return aclsWithInterfaces, nil
	}

	var interfaceData []*ACLToInterface

	msg := &acl_api.ACLInterfaceListDump{
		SwIfIndex: 0xffffffff, // dump all
	}

	req := vppChannel.SendMultiRequest(msg)

	var wasErr error
	for {
		reply := &acl_api.ACLInterfaceListDetails{}
		stop, err := req.ReceiveReply(reply)
		if stop {
			break
		}
		if err != nil {
			log.Error(err)
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
						log.Warnf("ACL requires ingress interface with Idx %v which was not found in the mapping", data.SwIfIdx)
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

// DumpIPAcls returns a list of all configured ACLs with IP-type ruleData.
func DumpIPAcls(log logging.Logger, vch *govppapi.Channel,
	timeLog measure.StopWatchEntry) (map[ACLIdentifier][]acl_api.ACLRule, error) {
	// ACLDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	aclIPRules := make(map[ACLIdentifier][]acl_api.ACLRule)
	var wasErr error

	req := &acl_api.ACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := vch.SendMultiRequest(req)
	for {
		msg := &acl_api.ACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			log.Error(err)
			wasErr = err
		}
		if stop {
			break
		}

		identifier := ACLIdentifier{
			ACLIndex: msg.ACLIndex,
			Tag:      string(bytes.SplitN(msg.Tag, []byte{0x00}, 2)[0]),
		}

		aclIPRules[identifier] = msg.R
	}

	return aclIPRules, wasErr
}

// DumpMacIPAcls returns a list of all configured ACL with IPMAC-type ruleData.
func DumpMacIPAcls(log logging.Logger, vppChannel *govppapi.Channel,
	timeLog measure.StopWatchEntry) (map[ACLIdentifier][]acl_api.MacipACLRule, error) {
	// MacipACLDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	aclMACIPRules := make(map[ACLIdentifier][]acl_api.MacipACLRule)
	var wasErr error

	req := &acl_api.MacipACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := vppChannel.SendMultiRequest(req)
	for {
		msg := &acl_api.MacipACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			log.Error(err)
			wasErr = err
		}
		if stop {
			break
		}

		identifier := ACLIdentifier{
			ACLIndex: msg.ACLIndex,
			Tag:      string(bytes.SplitN(msg.Tag, []byte{0x00}, 2)[0]),
		}

		aclMACIPRules[identifier] = msg.R
	}
	return aclMACIPRules, wasErr
}

// DumpInterfaceAcls finds interface in VPP and returns its ACL configuration
func DumpInterfaceAcls(log logging.Logger, swIndex uint32, vppChannel *govppapi.Channel, stopwatch *measure.Stopwatch) (acl.AccessLists, uint8, error) {
	log.Info("DumpInterfaceAcls")
	alAcls := acl.AccessLists{
		Acl: []*acl.AccessLists_Acl{},
	}

	res, err := vppcalls.DumpInterface(swIndex, vppChannel, stopwatch)
	log.Infof("Res: %+v\n", res)
	if err != nil {
		return alAcls, 0, err
	}

	if res.SwIfIndex != swIndex {
		return alAcls, 0, fmt.Errorf("returned interface index %d does not match request", res.SwIfIndex)
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

func getIPRuleDetails(rule acl_api.ACLRule) (aclRule *acl.AccessLists_Acl_Rule, err error) {
	// Resolve rule actions
	actions := &acl.AccessLists_Acl_Rule_Actions{}
	if actions.AclAction, err = resolveRuleAction(rule.IsPermit); err != nil {
		return nil, err
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

func getMACIPRuleDetails(rule acl_api.MacipACLRule) (aclRule *acl.AccessLists_Acl_Rule, err error) {
	// Resolve rule actions
	actions := &acl.AccessLists_Acl_Rule_Actions{}
	if actions.AclAction, err = resolveRuleAction(rule.IsPermit); err != nil {
		return nil, err
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

// getIPACLDetails gets details for a given IP ACL from VPP and translates
// them from the binary VPP API format into the ACL Plugin's NB format.
func getIPACLDetails(vppChannel *govppapi.Channel, idx uint32) (aclRule *acl.AccessLists_Acl, err error) {
	req := &acl_api.ACLDump{}
	req.ACLIndex = uint32(idx)

	reply := &acl_api.ACLDetails{}
	err = vppChannel.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return nil, err
	}

	var ruleData []*acl.AccessLists_Acl_Rule
	for _, r := range reply.R {
		rule := acl.AccessLists_Acl_Rule{}

		ipRule, _ := getIPRuleDetails(r)

		matches := acl.AccessLists_Acl_Rule_Matches{
			IpRule: ipRule.Matches.GetIpRule(),
		}

		actions := acl.AccessLists_Acl_Rule_Actions{}
		if actions.AclAction, err = resolveRuleAction(r.IsPermit); err != nil {
			return nil, err
		}

		rule.Matches = &matches
		rule.Actions = &actions
		ruleData = append(ruleData, &rule)
	}

	return &acl.AccessLists_Acl{Rules: ruleData, AclName: string(bytes.SplitN(reply.Tag, []byte{0x00}, 2)[0])}, nil
}

// getIPRuleMatches translates an IP rule from the binary VPP API format into the
// ACL Plugin's NB format
func getIPRuleMatches(r acl_api.ACLRule) *acl.AccessLists_Acl_Rule_Matches_IpRule {
	ipRule := acl.AccessLists_Acl_Rule_Matches_IpRule{}

	var srcIP, dstIP string
	if r.IsIpv6 == 1 {
		srcIP = net.IP(r.SrcIPAddr).To16().String()
		dstIP = net.IP(r.DstIPAddr).To16().String()
	} else {
		srcIP = net.IP(r.SrcIPAddr[:4]).To4().String()
		dstIP = net.IP(r.DstIPAddr[:4]).To4().String()
	}

	ip := acl.AccessLists_Acl_Rule_Matches_IpRule_Ip{
		SourceNetwork:      fmt.Sprintf("%s/%d", srcIP, r.SrcIPPrefixLen),
		DestinationNetwork: fmt.Sprintf("%s/%d", dstIP, r.DstIPPrefixLen),
	}
	ipRule.Ip = &ip

	switch r.Proto {
	case vppcalls.TCPProto:
		ipRule.Tcp = getTCPMatchRule(r)
		break
	case vppcalls.UDPProto:
		ipRule.Udp = getUDPMatchRule(r)
		break
	case vppcalls.ICMPv4Proto:
	case vppcalls.ICMPv6Proto:
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
	var srcAddr string
	if rule.IsIpv6 == 1 {
		srcAddr = net.IP(rule.SrcIPAddr).To16().String()
	} else {
		srcAddr = net.IP(rule.SrcIPAddr[:4]).To4().String()
	}
	return &acl.AccessLists_Acl_Rule_Matches_MacIpRule{
		SourceAddress:        srcAddr,
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

// Returns rule action representation in model according to the vpp input
func resolveRuleAction(isPermit uint8) (acl.AclAction, error) {
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
