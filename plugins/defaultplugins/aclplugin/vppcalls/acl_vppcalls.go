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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/logging/measure"
	acl_api "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/bin_api/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"time"
)

// AddIPAcl create new L3/4 ACL. Input index == 0xffffffff, VPP provides index in reply.
func AddIPAcl(rules []*acl.AccessLists_Acl_Rule, aclName string, log logging.Logger, vppChannel *api.Channel, timeLog measure.StopWatchEntry) (uint32, error) {
	// ACLAddReplace time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

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
			return 0, fmt.Errorf("failed to write ACL %v", aclName)
		}
		if 0 != reply.Retval {
			return 0, fmt.Errorf("error %v while writing ACL %v to VPP", reply.Retval, aclName)
		}
		log.Infof("%v Ip ACL rule(s) written for ACL %v with index %v", len(aclIPRules), aclName, reply.ACLIndex)

		return reply.ACLIndex, nil
	}

	return 0, fmt.Errorf("no rules found for ACL %v", aclName)
}

// AddMacIPAcl create new L2 MAC IP ACL. VPP provides index in reply.
func AddMacIPAcl(rules []*acl.AccessLists_Acl_Rule, aclName string, log logging.Logger, vppChannel *api.Channel, timeLog measure.StopWatchEntry) (uint32, error) {
	// MacipACLAdd time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

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
			return 0, fmt.Errorf("failed to write ACL %v", aclName)
		}
		if 0 != reply.Retval {
			return 0, fmt.Errorf("error %v while writing ACL %v to VPP", reply.Retval, aclName)
		}
		log.Infof("%v Mac Ip ACL rule(s) written for ACL %v with index %v", len(aclMacIPRules), aclName, reply.ACLIndex)

		return reply.ACLIndex, nil
	}
	log.Debugf("No Mac Ip ACL rules written for ACL %v", aclName)
	return 0, fmt.Errorf("no rules found for ACL %v", aclName)
}

// ModifyIPAcl uses index (provided by VPP) to identify ACL which is modified
func ModifyIPAcl(aclIndex uint32, rules []*acl.AccessLists_Acl_Rule, aclName string, log logging.Logger,
	vppChannel *api.Channel, timeLog measure.StopWatchEntry) error {
	// ACLAddReplace time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

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
			return fmt.Errorf("failed to write ACL %v", aclName)
		}
		if 0 != reply.Retval {
			return fmt.Errorf("error %v while writing ACL %v to VPP", reply.Retval, aclName)
		}
		log.Infof("%v Ip ACL rule(s) written for ACL %v with index %v", len(aclIPRules), aclName, aclIndex)
		return nil
	}
	log.Debugf("No Ip ACL rules written for ACL %v", aclName)
	return nil
}

// DeleteIPAcl removes L3/L4 ACL
func DeleteIPAcl(aclIndex uint32, log logging.Logger, vppChannel *api.Channel, timeLog measure.StopWatchEntry) error {
	// ACLDel time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	msg := &acl_api.ACLDel{}
	msg.ACLIndex = aclIndex

	reply := &acl_api.ACLDelReply{}
	err := vppChannel.SendRequest(msg).ReceiveReply(reply)
	if err != nil {
		return fmt.Errorf("failed to remove L3/L4 ACL %v", aclIndex)
	}
	if 0 != reply.Retval {
		return fmt.Errorf("error %v while removing L3/L4 ACL %v", reply.Retval, aclIndex)
	}
	log.Infof("L3/L4 ACL %v removed", aclIndex)

	return nil
}

// DeleteMacIPAcl removes L2 ACL
func DeleteMacIPAcl(aclIndex uint32, log logging.Logger, vppChannel *api.Channel, timeLog measure.StopWatchEntry) error {
	// MacipACLDel time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	msg := &acl_api.MacipACLDel{}
	msg.ACLIndex = aclIndex

	reply := &acl_api.MacipACLDelReply{}
	err := vppChannel.SendRequest(msg).ReceiveReply(reply)
	if err != nil {
		return fmt.Errorf("failed to remove L2 ACL %v", aclIndex)
	}
	if 0 != reply.Retval {
		return fmt.Errorf("error %v while removing L2 ACL %v", reply.Retval, aclIndex)
	}
	log.Infof("L2 ACL %v removed", aclIndex)

	return nil
}

// Method transforms provided set of IP proto ACL rules to binapi ACL rules.
func transformACLIpRules(rules []*acl.AccessLists_Acl_Rule) ([]acl_api.ACLRule, error) {
	var aclIPRules []acl_api.ACLRule
	for _, rule := range rules {
		aclRule := new(acl_api.ACLRule)
		var err error
		// Actions
		if rule.Actions != nil {
			aclRule.IsPermit = uint8(rule.Actions.AclAction)
		}
		// Matches
		if rule.Matches != nil && rule.Matches.IpRule != nil {
			// Concerned to IP rules only
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

// Sets an IP ACL rule fields into provided ACL Rule object. Source and destination address has to be the same IP
// version and contain a network mask.
func ipACL(ipRule *acl.AccessLists_Acl_Rule_Matches_IpRule_Ip, aclRule *acl_api.ACLRule) (*acl_api.ACLRule, error) {
	// Resolve source address
	srcIP, srcNetwork, err := net.ParseCIDR(ipRule.SourceNetwork)
	if err != nil {
		return nil, err
	}
	if srcNetwork == nil || srcNetwork.Mask == nil {
		return nil, fmt.Errorf("source address does not contain a mask")
	}
	maskSize, _ := srcNetwork.Mask.Size()
	srcMask := uint8(maskSize)
	if len(strings.TrimSpace(ipRule.SourceNetwork)) != 0 &&
		(srcIP.To4() == nil && srcIP.To16() == nil) {
		return aclRule, fmt.Errorf("source address %v is invalid", ipRule.SourceNetwork)
	}
	// Resolve destination address
	dstIP, dstNetwork, err := net.ParseCIDR(ipRule.DestinationNetwork)
	if err != nil {
		return nil, err
	}
	if dstNetwork == nil || srcNetwork.Mask == nil {
		return nil, fmt.Errorf("dest address does not contain a mask")
	}
	maskSize, _ = dstNetwork.Mask.Size()
	dstMask := uint8(maskSize)
	if len(strings.TrimSpace(ipRule.DestinationNetwork)) != 0 &&
		(dstIP.To4() == nil && dstIP.To16() == nil) {
		return aclRule, fmt.Errorf("destination address %v is invalid", ipRule.DestinationNetwork)
	}

	// Check IP version (they should be the same), beware: IPv4 address can be converted to IPv6
	if (srcIP.To4() != nil && dstIP.To4() == nil && dstIP.To16() != nil) ||
		(srcIP.To4() == nil && srcIP.To16() != nil && dstIP.To4() != nil) {
		return aclRule, fmt.Errorf("source address %v and destionation address %v have different IP versions",
			ipRule.SourceNetwork, ipRule.DestinationNetwork)
	}
	// Ipv4 case
	if srcIP.To4() != nil && dstIP.To4() != nil {
		aclRule.IsIpv6 = 0
		aclRule.SrcIPAddr = srcIP.To4()
		aclRule.SrcIPPrefixLen = uint8(srcMask)
		aclRule.DstIPAddr = dstIP.To4()
		aclRule.DstIPPrefixLen = uint8(dstMask)
		// Ipv6 case
	} else if srcIP.To16() != nil || dstIP.To16() != nil {
		aclRule.IsIpv6 = 1
		aclRule.SrcIPAddr = srcIP.To16()
		aclRule.SrcIPPrefixLen = uint8(srcMask)
		aclRule.DstIPAddr = dstIP.To16()
		aclRule.DstIPPrefixLen = uint8(dstMask)
		// Both empty
	} else {
		aclRule.IsIpv6 = 0
	}
	return aclRule, nil
}

// Sets an ICMP ACL rule fields into provided ACL Rule object. Ranges are exclusive, use first = 0 and last = 255/65535
// (icmpv4/icmpv6) to match "any".
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

// Sets an TCP ACL rule fields into provided ACL Rule object.
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

// Sets an UDP ACL rule fields into provided ACL Rule object.
func udpACL(udpRule *acl.AccessLists_Acl_Rule_Matches_IpRule_Udp, aclRule *acl_api.ACLRule) *acl_api.ACLRule {
	aclRule.Proto = 17 // IANA UDP
	aclRule.SrcportOrIcmptypeFirst = uint16(udpRule.SourcePortRange.LowerPort)
	aclRule.SrcportOrIcmptypeLast = uint16(udpRule.SourcePortRange.UpperPort)
	aclRule.DstportOrIcmpcodeFirst = uint16(udpRule.DestinationPortRange.LowerPort)
	aclRule.DstportOrIcmpcodeLast = uint16(udpRule.DestinationPortRange.UpperPort)
	return aclRule
}

func otherACL(otherRule *acl.AccessLists_Acl_Rule_Matches_IpRule_Other, aclRule *acl_api.ACLRule) *acl_api.ACLRule {
	logrus.DefaultLogger().Warn("unknown protocol: %v", otherRule.Protocol)
	return aclRule
}
