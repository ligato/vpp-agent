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
	"fmt"
	"net"

	"strings"

	"git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	acl_api "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/bin_api/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
)

// AddIPAcl create new L3/4 ACL. Input index == 0xffffffff, VPP provides index in reply.
func AddIPAcl(rules []*acl.AccessLists_Acl_Rule, aclName string, vppChannel *api.Channel) (uint32, error) {
	// Prepare Ip rules
	aclIPRules, err := transformACLIpRules(rules)
	if err != nil {
		return 0, err
	}
	if len(aclIPRules) != 0 {
		msg := &acl_api.ACLAddReplace{}
		msg.ACLIndex = 0xffffffff // to make new Entry
		msg.Count = uint32(len(aclIPRules))
		msg.Tag = []byte(aclName)
		msg.R = aclIPRules

		reply := &acl_api.ACLAddReplaceReply{}

		err = vppChannel.SendRequest(msg).ReceiveReply(reply)
		if err != nil {
			return 0, fmt.Errorf("Failed to write ACL %v", aclName)
		}
		if 0 != reply.Retval {
			return 0, fmt.Errorf("Error %v while writing ACL %v to VPP", reply.Retval, aclName)
		}
		log.DefaultLogger().Infof("%v Ip ACL rule(s) written for ACL %v with index %v", len(aclIPRules), aclName, reply.ACLIndex)
		return reply.ACLIndex, nil
	}
	return 0, fmt.Errorf("No rules found for ACL %v", aclName)
}

// AddMacIPAcl create new L2 MAC IP ACL. VPP provides index in reply.
func AddMacIPAcl(rules []*acl.AccessLists_Acl_Rule, aclName string, vppChannel *api.Channel) (uint32, error) {
	// Prepare MAc Ip rules
	aclMacIPRules, err := transformACLMacIPRules(rules)
	if err != nil {
		return 0, err
	}
	if len(aclMacIPRules) != 0 {

		msg := &acl_api.MacipACLAdd{}
		msg.Count = uint32(len(aclMacIPRules))
		msg.Tag = []byte(aclName)
		msg.R = aclMacIPRules

		reply := &acl_api.MacipACLAddReply{}
		err = vppChannel.SendRequest(msg).ReceiveReply(reply)
		if err != nil {
			return 0, fmt.Errorf("Failed to write ACL %v", aclName)
		}
		if 0 != reply.Retval {
			return 0, fmt.Errorf("Error %v while writing ACL %v to VPP", reply.Retval, aclName)
		}
		log.DefaultLogger().Infof("%v Mac Ip ACL rule(s) written for ACL %v with index %v", len(aclMacIPRules), aclName, reply.ACLIndex)
		return reply.ACLIndex, nil
	}
	log.DefaultLogger().Debugf("No Mac Ip ACL rules written for ACL %v", aclName)
	return 0, fmt.Errorf("No rules found for ACL %v", aclName)
}

// ModifyIPAcl uses index (provided by VPP) to identify ACL which is modified
func ModifyIPAcl(aclIndex uint32, rules []*acl.AccessLists_Acl_Rule, aclName string, vppChannel *api.Channel) error {
	// Prepare Ip rules
	aclIPRules, err := transformACLIpRules(rules)
	if err != nil {
		return err
	}
	if len(aclIPRules) != 0 {
		msg := &acl_api.ACLAddReplace{}
		msg.ACLIndex = aclIndex
		msg.Count = uint32(len(aclIPRules))
		msg.Tag = []byte(aclName)
		msg.R = aclIPRules

		reply := &acl_api.ACLAddReplaceReply{}

		err = vppChannel.SendRequest(msg).ReceiveReply(reply)
		if err != nil {
			return fmt.Errorf("Failed to write ACL %v", aclName)
		}
		if 0 != reply.Retval {
			return fmt.Errorf("Error %v while writing ACL %v to VPP", reply.Retval, aclName)
		}
		log.DefaultLogger().Infof("%v Ip ACL rule(s) written for ACL %v with index %v", len(aclIPRules), aclName, aclIndex)
		return nil
	}
	log.DefaultLogger().Debugf("No Ip ACL rules written for ACL %v", aclName)
	return nil
}

// DeleteIPAcl removes L3/L4 ACL
func DeleteIPAcl(aclIndex uint32, vppChannel *api.Channel) error {
	msg := &acl_api.ACLDel{}
	msg.ACLIndex = aclIndex

	reply := &acl_api.ACLDelReply{}
	err := vppChannel.SendRequest(msg).ReceiveReply(reply)
	if err != nil {
		return fmt.Errorf("Failed to remove L3/L4 ACL %v", aclIndex)
	}
	if 0 != reply.Retval {
		return fmt.Errorf("Error %v while removing L3/L4 ACL %v", reply.Retval, aclIndex)
	}
	log.DefaultLogger().Infof("L3/L4 ACL %v removed", aclIndex)

	return nil
}

// DeleteMacIPAcl removes L2 ACL
func DeleteMacIPAcl(aclIndex uint32, vppChannel *api.Channel) error {
	msg := &acl_api.MacipACLDel{}
	msg.ACLIndex = aclIndex

	reply := &acl_api.MacipACLDelReply{}
	err := vppChannel.SendRequest(msg).ReceiveReply(reply)
	if err != nil {
		return fmt.Errorf("Failed to remove L2 ACL %v", aclIndex)
	}
	if 0 != reply.Retval {
		return fmt.Errorf("Error %v while removing L2 ACL %v", reply.Retval, aclIndex)
	}
	log.DefaultLogger().Infof("L2 ACL %v removed", aclIndex)

	return nil
}

// todo auxiliary methods can be moved to some util.go

func transformACLIpRules(rules []*acl.AccessLists_Acl_Rule) ([]acl_api.ACLRule, error) {
	var aclIPRules []acl_api.ACLRule
	for _, rule := range rules {
		aclRule := new(acl_api.ACLRule)
		var err error
		// Actions
		if rule.Actions != nil {
			ruleActions := rule.Actions
			aclRule.IsPermit = uint8(ruleActions.AclAction)
		}
		// Matches
		if rule.Matches != nil && rule.Matches.IpRule != nil {
			// Concerned to ip rules only
			ipRule := rule.Matches.IpRule
			// L3
			if ipRule.Ip != nil {
				aclRule, err = ipACL(ipRule.Ip, aclRule)
				if err != nil {
					return aclIPRules, err
				}
			}
			// ICMP/L4
			if ipRule.Icmp != nil {
				aclRule = icmpACL(ipRule.Icmp, aclRule)
			} else if ipRule.Tcp != nil {
				aclRule = tcpACL(ipRule.Tcp, aclRule)
			} else if ipRule.Udp != nil {
				aclRule = udpACL(ipRule.Udp, aclRule)
			} else if ipRule.Other != nil {
				aclRule = otherACL(ipRule.Other, aclRule)
			}
			aclIPRules = append(aclIPRules, *aclRule)
		}
	}
	return aclIPRules, nil
}

func transformACLMacIPRules(rules []*acl.AccessLists_Acl_Rule) ([]acl_api.MacipACLRule, error) {
	var aclMacIPRules []acl_api.MacipACLRule
	for _, rule := range rules {
		aclMacIPRule := new(acl_api.MacipACLRule)
		// Actions
		if rule.Actions != nil {
			ruleActions := rule.Actions
			aclMacIPRule.IsPermit = uint8(ruleActions.AclAction)
		}
		// Matches
		if rule.Matches != nil && rule.Matches.MacipRule != nil {
			// Concerned to MAC IP rules only
			macIPRule := rule.Matches.MacipRule
			if macIPRule == nil {
				continue
			}
			// Source IP Address + Prefix
			srcIPAddress := net.ParseIP(macIPRule.SourceAddress)
			if srcIPAddress.To4() != nil {
				aclMacIPRule.IsIpv6 = 0
				aclMacIPRule.SrcIPAddr = srcIPAddress.To4()
				aclMacIPRule.SrcIPPrefixLen = uint8(macIPRule.SourceAddressPrefix)
			} else if srcIPAddress.To16() != nil {
				aclMacIPRule.IsIpv6 = 1
				aclMacIPRule.SrcIPAddr = srcIPAddress.To16()
				aclMacIPRule.SrcIPPrefixLen = uint8(macIPRule.SourceAddressPrefix)
			} else {
				return aclMacIPRules, fmt.Errorf("IP address %v: unknown version", macIPRule.SourceAddress)
			}
			// MAC + mask
			srcMac, err := net.ParseMAC(macIPRule.SourceMacAddress)
			if err != nil {
				return aclMacIPRules, err
			}
			srcMacMask, err := net.ParseMAC(macIPRule.SourceMacAddressMask)
			if err != nil {
				return aclMacIPRules, err
			}
			aclMacIPRule.SrcMac = srcMac
			aclMacIPRule.SrcMacMask = srcMacMask
			aclMacIPRules = append(aclMacIPRules, *aclMacIPRule)
		}
	}
	return aclMacIPRules, nil
}

func ipACL(ipRule *acl.AccessLists_Acl_Rule_Matches_IpRule_Ip, aclRule *acl_api.ACLRule) (*acl_api.ACLRule, error) {
	sourceNetwork := net.ParseIP(ipRule.SourceNetwork)
	destinationNetwork := net.ParseIP(ipRule.DestinationNetwork)
	if len(strings.TrimSpace(ipRule.SourceNetwork)) != 0 &&
		(sourceNetwork.To4() == nil && sourceNetwork.To16() == nil) {
		return aclRule, fmt.Errorf("Source address %v is invalid", ipRule.SourceNetwork)
	}
	if len(strings.TrimSpace(ipRule.DestinationNetwork)) != 0 &&
		(destinationNetwork.To4() == nil && destinationNetwork.To16() == nil) {
		return aclRule, fmt.Errorf("Destination address %v is invalid", ipRule.DestinationNetwork)
	}

	// beware: IPv4 address can be converted to IPv6
	if (sourceNetwork.To4() != nil && destinationNetwork.To4() == nil && destinationNetwork.To16() != nil) ||
		(sourceNetwork.To4() == nil && sourceNetwork.To16() != nil && destinationNetwork.To4() != nil) {
		return aclRule, fmt.Errorf("Source address %v and destionation address %v have different IP versions",
			ipRule.SourceNetwork, ipRule.DestinationNetwork)
	}
	if sourceNetwork.To4() != nil || destinationNetwork.To4() != nil {
		aclRule.IsIpv6 = 0
		aclRule.SrcIPAddr = sourceNetwork.To4()
		aclRule.DstIPAddr = destinationNetwork.To4()
	} else if sourceNetwork.To16() != nil || destinationNetwork.To16() != nil {
		aclRule.IsIpv6 = 1
		aclRule.SrcIPAddr = sourceNetwork.To16()
		aclRule.DstIPAddr = destinationNetwork.To16()
	} else {
		// both empty
		aclRule.IsIpv6 = 0
	}
	return aclRule, nil
}

// Ranges are exclusive, use first = 0 and last = 255/65535 (icmpv4/icmpv6) to match "any"
func icmpACL(icmpRule *acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp, aclRule *acl_api.ACLRule) *acl_api.ACLRule {
	if icmpRule == nil {
		return aclRule
	}
	if icmpRule.Icmpv6 {
		aclRule.Proto = 58 // IANA ICMPv6
		aclRule.IsIpv6 = 1
		// ICMPv6 type range
		aclRule.SrcportOrIcmptypeFirst = uint16(icmpRule.IcmpTypeRange.First)
		aclRule.SrcportOrIcmptypeLast = uint16(icmpRule.IcmpTypeRange.Last)
		// ICMPv6 code range
		aclRule.DstportOrIcmpcodeFirst = uint16(icmpRule.IcmpCodeRange.First)
		aclRule.DstportOrIcmpcodeLast = uint16(icmpRule.IcmpCodeRange.First)
	} else {
		aclRule.Proto = 1 // IANA ICMPv4
		aclRule.IsIpv6 = 0
		// ICMPv4 type range
		aclRule.SrcportOrIcmptypeFirst = uint16(icmpRule.IcmpTypeRange.First)
		aclRule.SrcportOrIcmptypeLast = uint16(icmpRule.IcmpTypeRange.Last)
		// ICMPv4 code range
		aclRule.DstportOrIcmpcodeFirst = uint16(icmpRule.IcmpCodeRange.First)
		aclRule.DstportOrIcmpcodeLast = uint16(icmpRule.IcmpCodeRange.Last)
	}
	return aclRule
}

func tcpACL(tcpRule *acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp, aclRule *acl_api.ACLRule) *acl_api.ACLRule {
	aclRule.Proto = 6 // IANA TCP
	aclRule.SrcportOrIcmptypeFirst = uint16(tcpRule.SourcePortRange.LowerPort)
	aclRule.SrcportOrIcmptypeLast = uint16(tcpRule.SourcePortRange.UpperPort)
	aclRule.DstportOrIcmpcodeFirst = uint16(tcpRule.DestinationPortRange.LowerPort)
	aclRule.DstportOrIcmpcodeLast = uint16(tcpRule.DestinationPortRange.UpperPort)
	aclRule.TCPFlagsValue = uint8(tcpRule.TcpFlagsValue)
	aclRule.TCPFlagsMask = uint8(tcpRule.TcpFlagsMask)
	return aclRule
}

func udpACL(udpRule *acl.AccessLists_Acl_Rule_Matches_IpRule_Udp, aclRule *acl_api.ACLRule) *acl_api.ACLRule {
	aclRule.Proto = 17 // IANA UDP
	aclRule.SrcportOrIcmptypeFirst = uint16(udpRule.SourcePortRange.LowerPort)
	aclRule.SrcportOrIcmptypeLast = uint16(udpRule.SourcePortRange.UpperPort)
	aclRule.DstportOrIcmpcodeFirst = uint16(udpRule.DestinationPortRange.LowerPort)
	aclRule.DstportOrIcmpcodeLast = uint16(udpRule.DestinationPortRange.UpperPort)
	return aclRule
}

func otherACL(otherRule *acl.AccessLists_Acl_Rule_Matches_IpRule_Other, aclRule *acl_api.ACLRule) *acl_api.ACLRule {
	log.DefaultLogger().Warnf("Unknown protocol: %v", otherRule.Protocol)
	return aclRule
}
