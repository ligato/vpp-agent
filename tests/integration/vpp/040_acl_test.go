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

package vpp

import (
	"strings"
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"

	aclplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

func rulePerm(permit bool) acl.ACL_Rule_Action {
	var aclRulePerm acl.ACL_Rule_Action

	if permit {
		aclRulePerm = acl.ACL_Rule_PERMIT
	} else {
		aclRulePerm = acl.ACL_Rule_DENY
	}

	return aclRulePerm
}

// helper function which returns ACL_Rule (to avoid of declarating variable)
func newACLIPRule(permit bool, src, dst string) *acl.ACL_Rule {
	return &acl.ACL_Rule{
		Action: rulePerm(permit),
		IpRule: &acl.ACL_Rule_IpRule{
			Ip: &acl.ACL_Rule_IpRule_Ip{
				SourceNetwork:      src,
				DestinationNetwork: dst,
			},
		},
	}
}

// helper function which returns ACL_Rule (to avoid of declarating variable)
func newACLIPRuleIcmp(permit, isicmpv6 bool, codefrom, codeto, typefrom, typeto uint32) *acl.ACL_Rule {
	return &acl.ACL_Rule{
		Action: rulePerm(permit),
		IpRule: &acl.ACL_Rule_IpRule{
			Icmp: &acl.ACL_Rule_IpRule_Icmp{
				Icmpv6: isicmpv6,
				IcmpCodeRange: &acl.ACL_Rule_IpRule_Icmp_Range{
					First: codefrom,
					Last:  codeto,
				},
				IcmpTypeRange: &acl.ACL_Rule_IpRule_Icmp_Range{
					First: typefrom,
					Last:  typeto,
				},
			},
		},
	}
}

// helper function which returns ACL_Rule (to avoid of declarating variable)
func newACLIPRuleTCP(permit bool, flagsval, flagsmask, srcportfrom, srcportto, destportfrom, destportto uint32) *acl.ACL_Rule {
	return &acl.ACL_Rule{
		Action: rulePerm(permit),
		IpRule: &acl.ACL_Rule_IpRule{
			Tcp: &acl.ACL_Rule_IpRule_Tcp{
				TcpFlagsMask:  flagsmask,
				TcpFlagsValue: flagsval,
				SourcePortRange: &acl.ACL_Rule_IpRule_PortRange{
					LowerPort: srcportfrom,
					UpperPort: srcportto,
				},
				DestinationPortRange: &acl.ACL_Rule_IpRule_PortRange{
					LowerPort: destportfrom,
					UpperPort: destportto,
				},
			},
		},
	}
}

// helper function which returns ACL_Rule (to avoid of declarating variable)
func newACLIPRuleUDP(permit bool, srcportfrom, srcportto, destportfrom, destportto uint32) *acl.ACL_Rule {
	return &acl.ACL_Rule{
		Action: rulePerm(permit),
		IpRule: &acl.ACL_Rule_IpRule{
			Udp: &acl.ACL_Rule_IpRule_Udp{
				SourcePortRange: &acl.ACL_Rule_IpRule_PortRange{
					LowerPort: srcportfrom,
					UpperPort: srcportto,
				},
				DestinationPortRange: &acl.ACL_Rule_IpRule_PortRange{
					LowerPort: destportfrom,
					UpperPort: destportto,
				},
			},
		},
	}
}

// helper function which returns ACL_Rule (to avoid of declarating variable)
func newACLMacIPRule(permit bool, src string, srcPrefix uint32, srcMacAddr string, srcMacMask string) *acl.ACL_Rule {
	return &acl.ACL_Rule{
		Action: rulePerm(permit),
		MacipRule: &acl.ACL_Rule_MacIpRule{
			SourceAddress:        src,
			SourceAddressPrefix:  srcPrefix,
			SourceMacAddress:     srcMacAddr,
			SourceMacAddressMask: srcMacMask,
		},
	}
}

func TestCRUDIPAcl(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))
	Expect(ih).To(Not(BeNil()), "Handler should be created.")

	const ifName = "loop1"
	ifIdx, errI := ih.AddLoopbackInterface(ifName)
	Expect(errI).To(BeNil())
	t.Logf("Prerequsite: interface %v created - its index %v", ifName, ifIdx)

	const ifName2 = "loop2"
	ifIdx2, errI2 := ih.AddLoopbackInterface(ifName2)
	Expect(errI2).To(BeNil())
	t.Logf("Prerequsite: interface %v created - its index %v", ifName2, ifIdx2)

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-iface1"), "test-iface1")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{
		SwIfIndex: ifIdx,
	})
	ifIndexes.Put(ifName2, &ifaceidx.IfaceMetadata{
		SwIfIndex: ifIdx2,
	})

	h := aclplugin_vppcalls.CompatibleACLHandler(ctx.vppClient, ifIndexes)
	Expect(h).To(Not(BeNil()), "Handler should be created.")

	acls, errx := h.DumpACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(BeEmpty())
	t.Log("no acls dumped")

	const aclname = "test0"
	aclIdx, err := h.AddACL([]*acl.ACL_Rule{
		//RuleName:  "permitIPv4",
		newACLIPRule(true, "192.168.1.1/32", "10.20.0.0/24"),
		//RuleName:  "permitIPv6",
		newACLIPRule(true, "dead::1/64", "dead::2/64"),
		//RuleName:  "denyICMP",
		newACLIPRuleIcmp(false, false, 10, 20, 30, 40),
		//RuleName:  "denyICMPv6",
		newACLIPRuleIcmp(false, true, 10, 20, 30, 40),
		//RuleName:  "permitTCP",
		newACLIPRuleTCP(true, 10, 20, 150, 250, 1150, 1250),
		//RuleName:  "denyUDP",
		newACLIPRuleUDP(false, 150, 250, 1150, 1250),
	}, aclname)
	Expect(err).To(BeNil())
	Expect(aclIdx).To(BeEquivalentTo(0))
	t.Logf("acl \"%v\" added - its index %d", aclname, aclIdx)

	err = h.SetACLToInterfacesAsIngress(aclIdx, []uint32{ifIdx})
	Expect(err).To(BeNil())
	t.Logf("acl with index %d was assigned to interface %v ingress", aclIdx, ifName)

	acls, errx = h.DumpACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	var rules []*acl.ACL_Rule
	var isIPRulePresent, isICMPRulePresent, isICMP6RulePresent, isForInterface bool
	for _, item := range acls {
		rules = item.ACL.Rules
		if (item.Meta.Index == aclIdx) && (aclname == item.Meta.Tag) {
			t.Logf("found ACL \"%v\"", item.Meta.Tag)
			for _, rule := range rules {
				t.Logf("%+v", rule)
				//here maybe all rules should be checked
				if (rule.IpRule.GetIp().SourceNetwork == "192.168.1.1/32") &&
					(rule.IpRule.GetIp().DestinationNetwork == "10.20.0.0/24") {
					isIPRulePresent = true
				}
				if (rule.IpRule.GetIcmp().GetIcmpCodeRange().GetFirst() == 10) &&
					(rule.IpRule.GetIcmp().GetIcmpCodeRange().GetLast() == 20) &&
					(rule.IpRule.GetIcmp().GetIcmpTypeRange().GetFirst() == 30) &&
					(rule.IpRule.GetIcmp().GetIcmpTypeRange().GetLast() == 40) {
					if rule.IpRule.GetIcmp().GetIcmpv6() {
						isICMP6RulePresent = true
					} else {
						isICMPRulePresent = true
					}
				}
			}

			// check assignation to interface
			for _, intf := range item.ACL.Interfaces.Ingress {
				if intf == ifName {
					isForInterface = true
					break
				}
			}
		}
	}
	Expect(isIPRulePresent).To(BeTrue(), "Configured IP rule should be present")
	Expect(isICMPRulePresent).To(BeTrue(), "Configured ICMP rule should be present")
	Expect(isICMP6RulePresent).To(BeTrue(), "Configured ICMPv6 rule should be present")
	Expect(isForInterface).To(BeTrue(), "acl should be assigned to interface")

	indexes := []uint32{ifIdx, ifIdx2}
	ifaces, errI3 := h.DumpACLInterfaces(indexes)
	Expect(errI3).To(Succeed())
	Expect(ifaces).To(HaveLen(2))
	t.Logf("%v", ifaces)
	t.Logf("%v", ifaces[1])
	t.Logf("%v", ifaces[2])
	//this does not work for VPP 19.04 and maybe also other version
	//Expect(ifaces[0].Ingress).To(Equal([]string{ifName}))
	//Expect(ifaces[2].Egress).To(Equal([]string{ifName2}))

	//negative tests - it is expected failure
	t.Logf("Let us test some negative cases....")
	_, err = h.AddACL([]*acl.ACL_Rule{}, "test1")
	Expect(err).To(Not(BeNil()))
	t.Logf("adding acls failed: %v", err)

	_, err = h.AddACL([]*acl.ACL_Rule{newACLIPRule(true, ".0.", "10.20.0.0/24")}, "test2")
	Expect(err).To(Not(BeNil()))
	t.Logf("adding acls failed: %v", err)

	_, err = h.AddACL([]*acl.ACL_Rule{newACLIPRule(true, "192.168.1.1/32", ".0.")}, "test3")
	Expect(err).To(Not(BeNil()))
	t.Logf("adding acls failed: %v", err)

	_, err = h.AddACL([]*acl.ACL_Rule{newACLIPRule(true, "192.168.1.1/32", "dead::1/64")}, "test4")
	Expect(err).To(Not(BeNil()))
	t.Logf("adding acls failed: %v", err)

	_, err = h.AddACL([]*acl.ACL_Rule{newACLIPRule(true, "", "")}, "test4")
	Expect(err).To(Not(BeNil()))
	t.Logf("adding acls failed: %v", err)

	//add the same acls again but it will be assigned to the second interface
	t.Log("Now let us add the second acl to the second interface")
	const aclname2 = "test5"
	aclIdx, err = h.AddACL([]*acl.ACL_Rule{
		//RuleName:  "permitIPv4",
		newACLIPRule(true, "192.168.1.1/32", "10.20.0.0/24"),
		//RuleName:  "permitIPv6",
		newACLIPRule(true, "dead::1/64", "dead::2/64"),
		//RuleName:  "denyICMP",
		newACLIPRuleIcmp(false, false, 11, 21, 31, 41),
		//RuleName:  "denyICMPv6",
		newACLIPRuleIcmp(false, true, 11, 21, 31, 41),
		//RuleName:  "permitTCP",
		newACLIPRuleTCP(true, 10, 20, 150, 250, 1150, 1250),
		//RuleName:  "denyUDP",
		newACLIPRuleUDP(false, 150, 250, 1150, 1250),
	}, aclname2)

	Expect(err).To(BeNil())
	Expect(aclIdx).To(BeEquivalentTo(1))
	t.Logf("acl \"%v\" added - its index %d", aclname2, aclIdx)

	err = h.SetACLToInterfacesAsEgress(aclIdx, []uint32{ifIdx2})
	Expect(err).To(BeNil())
	t.Logf("acl with index %d was assigned to interface %v egress", aclIdx, ifName2)

	acls, errx = h.DumpACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(2))
	t.Log("amount of acls dumped: 2")

	isIPRulePresent = false
	isICMPRulePresent = false
	isICMP6RulePresent = false
	isForInterface = false
	for _, item := range acls {
		rules = item.ACL.Rules
		if (item.Meta.Index == aclIdx) && (aclname2 == item.Meta.Tag) {
			t.Logf("found ACL \"%v\"", item.Meta.Tag)
			for _, rule := range rules {
				//t.Logf("%+v", rule)
				if (rule.IpRule.GetIp().SourceNetwork == "192.168.1.1/32") &&
					(rule.IpRule.GetIp().DestinationNetwork == "10.20.0.0/24") {
					isIPRulePresent = true
				}
				if (rule.IpRule.GetIcmp().GetIcmpCodeRange().GetFirst() == 11) &&
					(rule.IpRule.GetIcmp().GetIcmpCodeRange().GetLast() == 21) &&
					(rule.IpRule.GetIcmp().GetIcmpTypeRange().GetFirst() == 31) &&
					(rule.IpRule.GetIcmp().GetIcmpTypeRange().GetLast() == 41) {
					if rule.IpRule.GetIcmp().GetIcmpv6() {
						isICMP6RulePresent = true
					} else {
						isICMPRulePresent = true
					}
				}
			}
			// check assignation to interface
			for _, intf := range item.ACL.Interfaces.Egress {
				if intf == ifName2 {
					isForInterface = true
					break
				}
			}
		}
	}
	Expect(isIPRulePresent).To(BeTrue(), "Configured IP should be present")
	Expect(isICMPRulePresent).To(BeTrue(), "Configured ICMP rule should be present")
	Expect(isICMP6RulePresent).To(BeTrue(), "Configured ICMPv6 rule should be present")
	Expect(isForInterface).To(BeTrue(), "acl should be assigned to interface")

	//negative tests
	err = h.DeleteACL(5)
	Expect(err).To(Not(BeNil()))
	t.Logf("deleting acls failed: %v", err)

	// find the acl with aclname test0
	var foundaclidx uint32
	for _, item := range acls {
		rules = item.ACL.Rules
		if aclname == item.Meta.Tag {
			foundaclidx = item.Meta.Index
			break
		}
	}
	err = h.DeleteACL(foundaclidx)
	Expect(err).To(Not(BeNil()))
	t.Logf("deleting of acl \"%v\" failed: %v", aclname, err)

	// DELETE ACL
	err = h.RemoveACLFromInterfacesAsIngress(foundaclidx, []uint32{ifIdx})
	Expect(err).To(BeNil())
	t.Logf("removing acl \"%v\" from interface with index %d succeed", aclname, ifIdx)
	err = h.DeleteACL(foundaclidx)
	Expect(err).To(BeNil())
	t.Logf("deleting of acl \"%v\" succeed", aclname)

	acls, errx = h.DumpACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	for _, aclrecord := range acls {
		if aclrecord.Meta.Index == foundaclidx {
			t.Fatalf("This acll should be deleted : %v", errx)
		}
	}

	// MODIFY ACL
	rule2modify := []*acl.ACL_Rule{
		newACLIPRule(true, "10.20.30.1/32", "10.20.0.0/24"),
		newACLIPRule(true, "dead:dead::3/64", "dead:dead::4/64"),
		newACLIPRuleIcmp(true, false, 15, 25, 35, 45),
	}

	const aclname4 = "test_modify0"
	err = h.ModifyACL(1, rule2modify, aclname4)
	Expect(err).To(BeNil())
	t.Logf("modifying of acl with index 1 succeed - the new name of acl is  \"%v\"", aclname4)

	acls, errx = h.DumpACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	isIPRulePresent = false
	isICMPRulePresent = false
	isForInterface = false
	var modifiedacl aclplugin_vppcalls.ACLDetails
	for _, item := range acls {
		modifiedacl = *item
		rules = item.ACL.Rules
		if item.Meta.Index == aclIdx && (aclname4 == item.Meta.Tag) {
			t.Logf("Found modified ACL \"%v\"", item.Meta.Tag)
			for _, rule := range rules {
				//t.Logf("%+v", rule)
				if (rule.IpRule.GetIp().SourceNetwork == "192.168.1.1/32") &&
					(rule.IpRule.GetIp().DestinationNetwork == "10.20.0.0/24") {
					t.Fatal("Old rules should not be present")
				}
				//here maybe should be checked all rules
				if (rule.IpRule.GetIp().SourceNetwork == "10.20.30.1/32") &&
					(rule.IpRule.GetIp().DestinationNetwork == "10.20.0.0/24") {
					isIPRulePresent = true
				}
				if (rule.IpRule.GetIcmp().GetIcmpCodeRange().GetFirst() == 15) &&
					(rule.IpRule.GetIcmp().GetIcmpCodeRange().GetLast() == 25) &&
					(rule.IpRule.GetIcmp().GetIcmpTypeRange().GetFirst() == 35) &&
					(rule.IpRule.GetIcmp().GetIcmpTypeRange().GetLast() == 45) {
					if rule.IpRule.GetIcmp().GetIcmpv6() {
						isICMP6RulePresent = true
					} else {
						isICMPRulePresent = true
					}
				}
			}
			// check assignation to interface
			for _, intf := range item.ACL.Interfaces.Egress {
				if intf == ifName2 {
					isForInterface = true
					break
				}
			}
		}
	}
	Expect(isIPRulePresent).To(BeTrue(), "Configured IP should be present")
	Expect(isICMPRulePresent).To(BeTrue(), "Configured ICMP rule should be present")
	Expect(isForInterface).To(BeTrue(), "acl should be assigned to interface")

	// negative test
	err = h.ModifyACL(1, []*acl.ACL_Rule{newACLIPRule(true, ".0.", "10.20.0.0/24")}, "test_modify1")
	Expect(err).To(Not(BeNil()))
	t.Logf("modifying of acl failed: %v", err)

	const aclname3 = "test_modify2"
	err = h.ModifyACL(1, []*acl.ACL_Rule{}, aclname3)
	Expect(err).To(BeNil())
	t.Logf("acl with index 1 was modified by empty ACL definition")

	acls, errx = h.DumpACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	isIPRulePresent = false
	isForInterface = false
	for _, item := range acls {
		if item.Meta.Index == aclIdx && (aclname4 == item.Meta.Tag) {
			t.Logf("Found modified ACL \"%v\"", item.Meta.Tag)
			Expect(item.ACL.String()).To(Equal(modifiedacl.ACL.String()), "Last update should not cause any change in acl definition.")
			break
		}
	}
	t.Logf("found no change in definition of acl \"%v\" after trying to modify by empty acl definition", aclname4)

	// DELETE ACL
	err = h.RemoveACLFromInterfacesAsEgress(aclIdx, []uint32{ifIdx2})
	Expect(err).To(BeNil())
	t.Log("removing acl from interface succeed")

	acls, errx = h.DumpACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	isIPRulePresent = false
	isForInterface = false
	for _, item := range acls {
		if item.Meta.Index == aclIdx { //&& (aclname2 == item.Meta.Tag) {
			t.Logf("Found modified ACL \"%v\"", item.Meta.Tag)
			Expect(item.ACL.Interfaces.String()).Should(BeEmpty(), "Interface assignment should be removed")
			break
		}
	}

	err = h.DeleteACL(aclIdx)
	Expect(err).To(BeNil())
	t.Logf("deleting acl succeed")

	acls, errx = h.DumpACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(BeEmpty())
	t.Log("no acls dumped")
}

// Test add MACIP acl rules
func TestCRUDMacIPAcl(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))
	Expect(ih).To(Not(BeNil()), "Handler should be created.")

	const ifName = "loop1"
	ifIdx, errI := ih.AddLoopbackInterface(ifName)
	Expect(errI).To(BeNil())
	t.Logf("Prerequsite: interface %v created - its index %v", ifName, ifIdx)

	const ifName2 = "loop2"
	ifIdx2, errI2 := ih.AddLoopbackInterface(ifName2)
	Expect(errI2).To(BeNil())
	t.Logf("Prerequsite: interface %v created - its index %v", ifName2, ifIdx2)

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-iface1"), "test-iface1")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{
		SwIfIndex: ifIdx,
	})
	ifIndexes.Put(ifName2, &ifaceidx.IfaceMetadata{
		SwIfIndex: ifIdx2,
	})

	h := aclplugin_vppcalls.CompatibleACLHandler(ctx.vppClient, ifIndexes)
	Expect(h).To(Not(BeNil()), "Handler should be created.")
	if h == nil {
		t.Fatalf("handler was not created")
	}

	acls, errx := h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(BeEmpty())
	t.Log("no acls dumped")

	const aclname = "test6"
	aclIdx, err := h.AddMACIPACL([]*acl.ACL_Rule{
		//RuleName:  "denyIPv4",
		newACLMacIPRule(false, "192.168.0.1", 16, "11:44:0A:B8:4A:35", "ff:ff:ff:ff:00:00"),
		//RuleName:  "denyIPv6",
		newACLMacIPRule(false, "dead::1", 64, "11:44:0A:B8:4A:35", "ff:ff:ff:ff:00:00"),
	}, aclname)
	Expect(err).To(BeNil())
	Expect(aclIdx).To(BeEquivalentTo(0))
	t.Logf("acl \"%v\" added - its index %d", aclname, aclIdx)

	err = h.SetMACIPACLToInterfaces(aclIdx, []uint32{ifIdx})
	Expect(err).To(BeNil())
	t.Logf("acl with index %d was assigned to interface %v", aclIdx, ifName)

	acls, errx = h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	var rules []*acl.ACL_Rule
	var isPresent bool
	var isForInterface bool
	for _, item := range acls {
		rules = item.ACL.Rules
		if (item.Meta.Index == aclIdx) && (aclname == item.Meta.Tag) {
			t.Logf("found ACL \"%v\"", item.Meta.Tag)
			for _, rule := range rules {
				if (rule.MacipRule.SourceAddress == "192.168.0.1") &&
					(rule.MacipRule.SourceAddressPrefix == 16) &&
					(strings.ToLower(rule.MacipRule.SourceMacAddress) == strings.ToLower("11:44:0A:B8:4A:35")) &&
					(rule.MacipRule.SourceMacAddressMask == "ff:ff:ff:ff:00:00") {
					isPresent = true
					break
				}
			}
			// check assignation to interface
			t.Logf("%v", item)
			t.Logf("%v", item.ACL.Interfaces)
			for _, intf := range item.ACL.Interfaces.Ingress {
				if intf == ifName {
					isForInterface = true
					break
				}
			}
		}
	}
	Expect(isPresent).To(BeTrue(), "Configured IP should be present")
	Expect(isForInterface).To(BeTrue(), "acl should be assigned to interface")

	indexes := []uint32{ifIdx, ifIdx2}
	ifaces, errI3 := h.DumpACLInterfaces(indexes)
	Expect(errI3).To(Succeed())
	Expect(ifaces).To(HaveLen(2))
	t.Logf("%v", ifaces)
	t.Logf("%v", ifaces[1])
	t.Logf("%v", ifaces[2])
	//this does not work for VPP 19.04 and maybe also other version
	//Expect(ifaces[0].Ingress).To(Equal([]string{ifName}))
	//Expect(ifaces[2].Egress).To(Equal([]string{ifName2}))

	//negative tests - it is expected failure
	t.Logf("Let us test some negative cases....")
	_, err = h.AddMACIPACL([]*acl.ACL_Rule{}, "test7")
	Expect(err).To(Not(BeNil()))
	t.Logf("adding acls failed: %v", err)

	_, err = h.AddMACIPACL([]*acl.ACL_Rule{newACLMacIPRule(true, "192.168.0.1", 16, "", "ff:ff:ff:ff:00:00")}, "test8")
	Expect(err).To(Not(BeNil()))
	t.Logf("adding acls failed: %v", err)

	_, err = h.AddMACIPACL([]*acl.ACL_Rule{newACLMacIPRule(true, "192.168.0.1", 16, "11:44:0A:B8:4A:36", "")}, "test9")
	Expect(err).To(Not(BeNil()))
	t.Logf("adding acls failed: %v", err)

	_, err = h.AddMACIPACL([]*acl.ACL_Rule{newACLMacIPRule(true, "", 16, "11:44:0A:B8:4A:36", "ff:ff:ff:ff:00:00")}, "test10")
	Expect(err).To(Not(BeNil()))
	Expect(err.Error()).To(BeEquivalentTo("invalid IP address "))
	t.Logf("adding acls failed: %v", err)

	// now let us add the same aclMACIPrules again
	//add the same acls again but it will be assigned to the second interface
	t.Log("Now let us add the second acl to the second interface")
	const aclname2 = "test11"
	aclIdx, err = h.AddMACIPACL([]*acl.ACL_Rule{
		//RuleName:  "denyIPv4",
		newACLMacIPRule(false, "192.168.0.1", 16, "11:44:0A:B8:4A:35", "ff:ff:ff:ff:00:00"),
		//RuleName:  "denyIPv6",
		newACLMacIPRule(false, "dead::1", 64, "11:44:0A:B8:4A:35", "ff:ff:ff:ff:00:00"),
	}, aclname2)
	Expect(err).To(BeNil())
	Expect(aclIdx).To(BeEquivalentTo(1))
	t.Logf("acl \"%v\" added - its index %d", aclname2, aclIdx)

	err = h.AddMACIPACLToInterface(aclIdx, ifName2)
	Expect(err).To(BeNil())
	t.Logf("acl with index %d was assigned to interface %v ", aclIdx, ifName2)

	acls, errx = h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(2))
	t.Log("amount of acls dumped: 2")

	isPresent = false
	isForInterface = false
	for _, item := range acls {
		rules = item.ACL.Rules
		if (item.Meta.Index == aclIdx) && (aclname2 == item.Meta.Tag) {
			t.Logf("found ACL \"%v\"", item.Meta.Tag)
			for _, rule := range rules {
				if (rule.MacipRule.SourceAddress == "192.168.0.1") &&
					(rule.MacipRule.SourceAddressPrefix == 16) &&
					(strings.ToLower(rule.MacipRule.SourceMacAddress) == strings.ToLower("11:44:0A:B8:4A:35")) &&
					(rule.MacipRule.SourceMacAddressMask == "ff:ff:ff:ff:00:00") {
					isPresent = true
					break
				}
			}
			// check assignation to interface
			t.Logf("%v", item)
			t.Logf("%v", item.ACL.Interfaces)
			for _, intf := range item.ACL.Interfaces.Ingress {
				if intf == ifName2 {
					isForInterface = true
					break
				}
			}
		}
	}
	Expect(isPresent).To(BeTrue(), "Configured IP should be present")
	Expect(isForInterface).To(BeTrue(), "acl should be assigned to interface")

	//negative tests
	err = h.DeleteMACIPACL(5)
	Expect(err).To(Not(BeNil()))
	t.Logf("deleting acls failed: %v", err)

	// find the acl with aclname test6
	var foundaclidx uint32
	for _, item := range acls {
		rules = item.ACL.Rules
		if aclname == item.Meta.Tag {
			foundaclidx = item.Meta.Index
			break
		}
	}
	err = h.DeleteMACIPACL(foundaclidx)
	Expect(err).To(BeNil())
	t.Logf("deleting of acl \"%v\" succeed", aclname)

	acls, errx = h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	for _, aclrecord := range acls {
		if aclrecord.Meta.Index == foundaclidx {
			t.Fatalf("This acll should be deleted : %v", errx)
		}
	}

	// MODIFY ACL
	rule2modify := []*acl.ACL_Rule{
		// acl.ACL_Rule_DENY
		newACLMacIPRule(false, "192.168.10.1", 24, "11:44:0A:B8:4A:37", "ff:ff:ff:ff:00:00"),
		newACLMacIPRule(false, "dead::2", 64, "11:44:0A:B8:4A:38", "ff:ff:ff:ff:00:00"),
	}

	const aclname4 = "test_modify0"
	err = h.ModifyMACIPACL(1, rule2modify, aclname4)
	Expect(err).To(BeNil())
	t.Logf("modifying of acl with index 1 succeed - the new name of acl is  \"%v\"", aclname4)

	acls, errx = h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	isPresent = false
	isForInterface = false
	var modifiedacl aclplugin_vppcalls.ACLDetails
	for _, item := range acls {
		modifiedacl = *item
		rules = item.ACL.Rules
		if item.Meta.Index == aclIdx && (aclname4 == item.Meta.Tag) {
			t.Logf("Found modified ACL \"%v\"", item.Meta.Tag)
			for _, rule := range rules {
				if (rule.MacipRule.SourceAddress == "192.168.0.1") &&
					(rule.MacipRule.SourceAddressPrefix == 16) &&
					(strings.ToLower(rule.MacipRule.SourceMacAddress) == strings.ToLower("11:44:0A:B8:4A:35")) &&
					(rule.MacipRule.SourceMacAddressMask == "ff:ff:ff:ff:00:00") {
					t.Fatal("Old rules should not be present")
				}
				if (rule.MacipRule.SourceAddress == "192.168.10.1") &&
					(rule.MacipRule.SourceAddressPrefix == 24) &&
					(strings.ToLower(rule.MacipRule.SourceMacAddress) == strings.ToLower("11:44:0A:B8:4A:37")) &&
					(rule.MacipRule.SourceMacAddressMask == "ff:ff:ff:ff:00:00") {
					isPresent = true
					break
				}
				//t.Logf("%+v", rule)
			}
			// check assignation to interface
			for _, intf := range item.ACL.Interfaces.Ingress {
				if intf == ifName2 {
					isForInterface = true
					break
				}
			}
		}
	}
	Expect(isPresent).To(BeTrue(), "Configured IP should be present")
	Expect(isForInterface).To(BeTrue(), "acl should be assigned to interface")

	t.Logf("%v", modifiedacl)

	// negative test
	err = h.ModifyMACIPACL(1, []*acl.ACL_Rule{newACLIPRule(true, ".0.", "10.20.0.0/24")}, "test_modify1")
	Expect(err).To(Not(BeNil()))
	t.Logf("modifying of acl failed: %v", err)

	err = h.ModifyMACIPACL(1, []*acl.ACL_Rule{
		//RuleName:  "permitIPv4",
		newACLIPRule(true, "192.168.1.1/32", "10.20.0.0/24"),
		//RuleName:  "permitIPv6",
		newACLIPRule(true, "dead::1/64", "dead::2/64"),
		//RuleName:  "permitIP",
		newACLIPRule(true, "", ""),
		//RuleName:  "denyICMP",
		newACLIPRuleIcmp(false, false, 150, 250, 1150, 1250),
		//RuleName:  "denyICMPv6",
		newACLIPRuleIcmp(false, true, 150, 250, 1150, 1250),
		//RuleName:  "permitTCP",
		newACLIPRuleTCP(true, 10, 20, 150, 250, 1150, 1250),
		//RuleName:  "denyUDP",
		newACLIPRuleUDP(false, 150, 250, 1150, 1250),
	}, "test_modify5")
	Expect(err).To(Not(BeNil()))
	t.Logf("modifying of acl failed: %v", err)

	err = h.SetMACIPACLToInterfaces(aclIdx, []uint32{ifIdx})
	Expect(err).To(BeNil())
	t.Logf("acl with index %d was assigned to interface %v", aclIdx, ifName)

	acls, errx = h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	isPresent = false
	isForInterface = false
	for _, item := range acls {
		rules = item.ACL.Rules
		if item.Meta.Index == aclIdx {
			t.Logf("found modified ACL \"%v\"", item.Meta.Tag)
			// check assignation to interface
			t.Logf("%v", item)
			t.Logf("%v", item.ACL.Interfaces)
			for _, intf := range item.ACL.Interfaces.Ingress {
				if intf == ifName {
					isForInterface = true
					break
				}
			}
		}
	}
	Expect(isForInterface).To(BeTrue(), "acl should be assigned to interface")

	err = h.SetMACIPACLToInterfaces(aclIdx, []uint32{ifIdx})
	Expect(err).To(BeNil())
	t.Logf("acl with index %d was assigned to interface %v", aclIdx, ifName)

	acls, errx = h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	isPresent = false
	isForInterface = false
	for _, item := range acls {
		rules = item.ACL.Rules
		if item.Meta.Index == aclIdx {
			t.Logf("found modified ACL \"%v\"", item.Meta.Tag)
			// check assignation to interface
			t.Logf("%v", item)
			t.Logf("%v", item.ACL.Interfaces)
			for _, intf := range item.ACL.Interfaces.Ingress {
				if intf == ifName {
					isForInterface = true
					break
				}
			}
		}
	}
	Expect(isForInterface).To(BeTrue(), "acl should be assigned to interface")

	err = h.DeleteMACIPACLFromInterface(aclIdx, ifName)
	Expect(err).To(BeNil())
	t.Logf("for acl with index %d was deleted the relation to interface %v", aclIdx, ifName)

	acls, errx = h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	isPresent = false
	isForInterface = false
	for _, item := range acls {
		rules = item.ACL.Rules
		if item.Meta.Index == aclIdx {
			t.Logf("Found modified ACL \"%v\"", item.Meta.Tag)
			// check assignation to interface
			t.Logf("%v", item)
			t.Logf("%v", item.ACL.Interfaces)
			for _, intf := range item.ACL.Interfaces.Ingress {
				if intf == ifName {
					t.Fatalf("acl should not be assigned to the interface %v", ifName)
				}
			}
		}
	}
	t.Logf("for acl was correctly deleted relation to interface %v", ifName)

	err = h.DeleteMACIPACLFromInterface(aclIdx, ifName2)
	Expect(err).To(BeNil())
	t.Logf("for acl with index %d was deleted the relation to interface %v", aclIdx, ifName2)

	acls, errx = h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	isPresent = false
	isForInterface = false
	for _, item := range acls {
		rules = item.ACL.Rules
		if item.Meta.Index == aclIdx {
			t.Logf("Found modified ACL \"%v\"", item.Meta.Tag)
			// check assignation to interface
			t.Logf("%v", item)
			t.Logf("%v", item.ACL.Interfaces)
			Expect(item.ACL.Interfaces).To(BeNil())
		}
	}
	t.Logf("for acl was correctly deleted relation to interface %v", ifName2)

	acls, errx = h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(HaveLen(1))
	t.Log("amount of acls dumped: 1")

	for _, aclrecord := range acls {
		foundaclidx = aclrecord.Meta.Index
	}

	err = h.DeleteMACIPACL(foundaclidx)
	Expect(err).To(BeNil())
	t.Logf("deleting acl succeed")

	acls, errx = h.DumpMACIPACL()
	Expect(errx).To(BeNil())
	Expect(acls).Should(BeEmpty())
	t.Log("no acls dumped")
}
