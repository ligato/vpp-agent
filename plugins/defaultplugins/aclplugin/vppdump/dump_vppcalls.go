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
)

// IPACLList defines the list of IP ACLs returned by the DumpIPAcl function
type IPACLList struct {
	Index      uint32               `json:"acl_index"`
	ACLDetails *acl.AccessLists_Acl `json:"acl_details"`
}

// DumpInterfaceAcls finds interface in VPP and returns its ACL configuration
func DumpInterfaceAcls(log logging.Logger, swIndex uint32, vppChannel *govppapi.Channel,
	timeLog measure.StopWatchEntry) (acl.AccessLists, error) {

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
		return alAcls, err
	}

	if res.SwIfIndex != swIndex {
		return alAcls, fmt.Errorf(fmt.Sprintf("Returned interface index %d does not match request",
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
	return alAcls, nil
}

// DumpIPAcl returns a list of all configured IP ACLs.
func DumpIPAcl(log logging.Logger, vch *govppapi.Channel, tl measure.StopWatchEntry) (*[]IPACLList, error) {

	// ACLDump time measurement
	start := time.Now()
	defer func() {
		if tl != nil {
			tl.LogTimeEntry(time.Since(start))
		}
	}()

	acli := make([]uint32, 0)
	req := &acl_api.ACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := vch.SendMultiRequest(req)
	for {
		msg := &acl_api.ACLDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, err
		}
		if stop {
			break
		}
		acli = append(acli, msg.ACLIndex)

	}
	acll := make([]IPACLList, 0)
	for _, idx := range acli {
		aclp, err := getIPACLDetails(vch, idx)
		if err != nil {
			return nil, err
		}
		acll = append(acll, IPACLList{idx, aclp})
	}

	return &acll, nil
}

// DumpMacIPAcl test function
func DumpMacIPAcl(log logging.Logger, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// MacipACLDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &acl_api.MacipACLDump{}
	req.ACLIndex = 0xffffffff
	reqContext := vppChannel.SendMultiRequest(req)
	for {
		msg := &acl_api.MacipACLDump{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return err
		}
		if stop {
			break
		}
		log.Info(msg.ACLIndex)
	}
	return nil
}

// getIPACLDetails gets details for a given IP ACL from VPP and translates
// them from the binary VPP API format into the ACL Plugin's NB format.
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

		matches := acl.AccessLists_Acl_Rule_Matches{
			IpRule: getIPRule(r),
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

// getIPRule translates an IP rule from the binary VPP API format into the
// ACL Plugin's NB format
func getIPRule(r acl_api.ACLRule) *acl.AccessLists_Acl_Rule_Matches_IpRule {
	ipRule := acl.AccessLists_Acl_Rule_Matches_IpRule{}

	saddr := net.IPNet{IP: r.SrcIPAddr, Mask: []byte{}}
	daddr := net.IPNet{IP: r.DstIPAddr, Mask: []byte{}}
	ip := acl.AccessLists_Acl_Rule_Matches_IpRule_Ip{
		SourceNetwork:      fmt.Sprintf("%s/%d", saddr.String(), r.SrcIPPrefixLen),
		DestinationNetwork: fmt.Sprintf("%s/%d", daddr.String(), r.DstIPPrefixLen),
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
