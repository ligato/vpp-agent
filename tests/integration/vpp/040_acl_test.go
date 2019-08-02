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
	"github.com/ligato/cn-infra/logging/logrus"
	acl "github.com/ligato/vpp-agent/api/models/vpp/acl"
	. "github.com/onsi/gomega"
	"testing"

	_ "github.com/ligato/vpp-agent/plugins/vpp/aclplugin"
	aclplugin_vppcalls "github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	_ "github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	//	"encoding/json"
)

var aclNoRules []*acl.ACL_Rule

var aclErr1Rules = []*acl.ACL_Rule{
	{
		Action: acl.ACL_Rule_PERMIT,
		IpRule: &acl.ACL_Rule_IpRule{
			Ip: &acl.ACL_Rule_IpRule_Ip{
				SourceNetwork:      ".0.",
				DestinationNetwork: "10.20.0.0/24",
			},
		},
	},
}

var aclErr2Rules = []*acl.ACL_Rule{
	{
		Action: acl.ACL_Rule_PERMIT,
		IpRule: &acl.ACL_Rule_IpRule{
			Ip: &acl.ACL_Rule_IpRule_Ip{
				SourceNetwork:      "192.168.1.1/32",
				DestinationNetwork: ".0.",
			},
		},
	},
}

var aclErr3Rules = []*acl.ACL_Rule{
	{
		Action: acl.ACL_Rule_PERMIT,
		IpRule: &acl.ACL_Rule_IpRule{
			Ip: &acl.ACL_Rule_IpRule_Ip{
				SourceNetwork:      "192.168.1.1/32",
				DestinationNetwork: "dead::1/64",
			},
		},
	},
}

var aclErr4Rules = []*acl.ACL_Rule{
	{
		Action: acl.ACL_Rule_PERMIT,
		MacipRule: &acl.ACL_Rule_MacIpRule{
			SourceAddress:        "192.168.0.1",
			SourceAddressPrefix:  uint32(16),
			SourceMacAddress:     "",
			SourceMacAddressMask: "ff:ff:ff:ff:00:00",
		},
	},
}

var aclErr5Rules = []*acl.ACL_Rule{
	{
		Action: acl.ACL_Rule_PERMIT,
		MacipRule: &acl.ACL_Rule_MacIpRule{
			SourceAddress:        "192.168.0.1",
			SourceAddressPrefix:  uint32(16),
			SourceMacAddress:     "11:44:0A:B8:4A:36",
			SourceMacAddressMask: "",
		},
	},
}

var aclErr6Rules = []*acl.ACL_Rule{
	{
		Action: acl.ACL_Rule_PERMIT,
		MacipRule: &acl.ACL_Rule_MacIpRule{
			SourceAddress:        "",
			SourceAddressPrefix:  uint32(16),
			SourceMacAddress:     "11:44:0A:B8:4A:36",
			SourceMacAddressMask: "ff:ff:ff:ff:00:00",
		},
	},
}

var aclIPrules = []*acl.ACL_Rule{
	{
		//RuleName:  "permitIPv4",
		Action: acl.ACL_Rule_PERMIT,
		IpRule: &acl.ACL_Rule_IpRule{
			Ip: &acl.ACL_Rule_IpRule_Ip{
				SourceNetwork:      "192.168.1.1/32",
				DestinationNetwork: "10.20.0.0/24",
			},
		},
	},
	{
		//RuleName:  "permitIPv6",
		Action: acl.ACL_Rule_PERMIT,
		IpRule: &acl.ACL_Rule_IpRule{
			Ip: &acl.ACL_Rule_IpRule_Ip{
				SourceNetwork:      "dead::1/64",
				DestinationNetwork: "dead::2/64",
			},
		},
	},
	{
		//RuleName:  "permitIP",
		Action: acl.ACL_Rule_PERMIT,
		IpRule: &acl.ACL_Rule_IpRule{
			Ip: &acl.ACL_Rule_IpRule_Ip{
				SourceNetwork:      "",
				DestinationNetwork: "",
			},
		},
	},
	{
		//RuleName:  "denyICMP",
		Action: acl.ACL_Rule_DENY,
		IpRule: &acl.ACL_Rule_IpRule{
			Icmp: &acl.ACL_Rule_IpRule_Icmp{
				Icmpv6: false,
				IcmpCodeRange: &acl.ACL_Rule_IpRule_Icmp_Range{
					First: 150,
					Last:  250,
				},
				IcmpTypeRange: &acl.ACL_Rule_IpRule_Icmp_Range{
					First: 1150,
					Last:  1250,
				},
			},
		},
	},
	{
		//RuleName:  "denyICMPv6",
		Action: acl.ACL_Rule_DENY,
		IpRule: &acl.ACL_Rule_IpRule{
			Icmp: &acl.ACL_Rule_IpRule_Icmp{
				Icmpv6: true,
				IcmpCodeRange: &acl.ACL_Rule_IpRule_Icmp_Range{
					First: 150,
					Last:  250,
				},
				IcmpTypeRange: &acl.ACL_Rule_IpRule_Icmp_Range{
					First: 1150,
					Last:  1250,
				},
			},
		},
	},
	{
		//RuleName:  "permitTCP",
		Action: acl.ACL_Rule_PERMIT,
		IpRule: &acl.ACL_Rule_IpRule{
			Tcp: &acl.ACL_Rule_IpRule_Tcp{
				TcpFlagsMask:  20,
				TcpFlagsValue: 10,
				SourcePortRange: &acl.ACL_Rule_IpRule_PortRange{
					LowerPort: 150,
					UpperPort: 250,
				},
				DestinationPortRange: &acl.ACL_Rule_IpRule_PortRange{
					LowerPort: 1150,
					UpperPort: 1250,
				},
			},
		},
	},
	{
		//RuleName:  "denyUDP",
		Action: acl.ACL_Rule_DENY,
		IpRule: &acl.ACL_Rule_IpRule{
			Udp: &acl.ACL_Rule_IpRule_Udp{
				SourcePortRange: &acl.ACL_Rule_IpRule_PortRange{
					LowerPort: 150,
					UpperPort: 250,
				},
				DestinationPortRange: &acl.ACL_Rule_IpRule_PortRange{
					LowerPort: 1150,
					UpperPort: 1250,
				},
			},
		},
	},
}

var aclMACIPrules = []*acl.ACL_Rule{
	{
		//RuleName:  "denyIPv4",
		Action: acl.ACL_Rule_DENY,
		MacipRule: &acl.ACL_Rule_MacIpRule{
			SourceAddress:        "192.168.0.1",
			SourceAddressPrefix:  uint32(16),
			SourceMacAddress:     "11:44:0A:B8:4A:35",
			SourceMacAddressMask: "ff:ff:ff:ff:00:00",
		},
	},
	{
		//RuleName:  "denyIPv6",
		Action: acl.ACL_Rule_DENY,
		MacipRule: &acl.ACL_Rule_MacIpRule{
			SourceAddress:        "dead::1",
			SourceAddressPrefix:  uint32(64),
			SourceMacAddress:     "11:44:0A:B8:4A:35",
			SourceMacAddressMask: "ff:ff:ff:ff:00:00",
		},
	},
}

func TestCRUDIPAcl(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppBinapi, logrus.NewLogger("test"))

	const ifName = "loop1"
	ifIdx, errI := ih.AddLoopbackInterface(ifName)
	if errI != nil {
		t.Fatalf("creating interface failed: %v", errI)
	}
	t.Logf("interface created %v", ifIdx)

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-iface1"), "test-iface1")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{
		SwIfIndex: ifIdx,
	})

	h := aclplugin_vppcalls.CompatibleACLVppHandler(ctx.vppBinapi, ifIndexes, logrus.NewLogger("test"))
	if h == nil {
		t.Fatalf("handler was not created")
	}

	acls, errx := h.DumpACL()
	if errx != nil {
		t.Fatalf("dumping acls failed: %v", errx)
	}
	t.Logf("%d acls dumped", len(acls))
	t.Logf("acls dumped %v", acls)

	aclIdx, err := h.AddACL(aclIPrules, "test0")
	Expect(err).To(BeNil())
	Expect(aclIdx).To(BeEquivalentTo(0))
	if err != nil {
		t.Fatalf("adding acls failed: %v", err)
	}
	t.Logf("acl added - with index %d", aclIdx)

	acls, errx = h.DumpACL()
	if errx != nil {
		t.Fatalf("dumping acls failed: %v", errx)
	}
	t.Logf("%d acls dumped", len(acls))
	t.Logf("acls dumped %v", acls)

	var rules []*acl.ACL_Rule
	for _, item := range acls {
		rules = item.ACL.Rules
		for _, rule := range rules {
			t.Logf("acls action dumped %v", rule.Action)

		}
		for _, rule := range rules {
			t.Logf("acls rule dumped %v", rule.IpRule.GetIp().DestinationNetwork)
			t.Logf("acls rule dumped %v", rule.IpRule.GetIp().SourceNetwork)

		}

		for _, rule := range rules {
			t.Logf("acls rule dumped %v", rule.IpRule.GetIcmp())

		}
		for _, rule := range rules {
			t.Logf("acls rule dumped %v", rule.IpRule.GetTcp())

		}
	}

	//json, err := json.MarshalIndent(acls[aclIdx].ACL, "", "  ")
	if err != nil {
		//	t.Fatalf("json processing failed: %v", err)
	}
	//t.Logf("%v", string(json))

	//t.Logf("%+v", acls[aclIdx].Meta.Index)
	//t.Logf("%+v", acls[aclIdx].Meta.Tag)
	//t.Logf("%+v", acls[aclIdx].ACL.GetInterfaces())
	//t.Logf("%+v", acls[aclIdx].ACL.GetName())
	//t.Logf("%+v", acls[aclIdx].ACL.GetRules())
	//t.Logf("%+v", acls[aclIdx].ACL.String())
	//t.Logf("%+v", acls[aclIdx].ACL.Name)
	//t.Logf("%+v", acls[aclIdx].ACL.Rules)
	//t.Logf("%+v", acls[aclIdx].ACL.Interfaces)
	//t.Logf("%+v", acls[aclIdx].ACL.XXX_MessageName())

	//negative tests - it is expected failure
	_, err = h.AddACL(aclNoRules, "test1")
	Expect(err).To(Not(BeNil()))
	if err != nil {
		t.Logf("adding acls failed: %v", err)
	}

	_, err = h.AddACL(aclErr1Rules, "test2")
	Expect(err).To(Not(BeNil()))
	if err != nil {
		t.Logf("adding acls failed: %v", err)
	}

	_, err = h.AddACL(aclErr2Rules, "test3")
	Expect(err).To(Not(BeNil()))
	if err != nil {
		t.Logf("adding acls failed: %v", err)
	}

	_, err = h.AddACL(aclErr3Rules, "test4")
	Expect(err).To(Not(BeNil()))
	if err != nil {
		t.Logf("adding acls failed: %v", err)
	}

	//add the same acls again
	aclIdx, err = h.AddACL(aclIPrules, "test5")
	if err != nil {
		t.Fatalf("adding acls failed: %v", err)
	}
	t.Logf("acl added with index %d", aclIdx)

	acls, errx = h.DumpACL()
	if errx != nil {
		t.Fatalf("dumping acls failed: %v", errx)
	}
	aclCnt := len(acls)
	t.Logf("%d acls dumped", aclCnt)
	if aclCnt != 2 {
		t.Fatalf("THe count of acls should be 2 not %d", aclCnt)
	}

	//t.Logf("%+v", acls[aclIdx].Meta.Index)
	//t.Logf("%+v", acls[aclIdx].Meta.Tag)

	//t.Logf("%+v", acls[aclIdx].ACL.GetInterfaces())
	//t.Logf("%+v", acls[aclIdx].ACL.GetName())
	//t.Logf("%+v", acls[aclIdx].ACL.GetRules())
	//t.Logf("%+v", acls[aclIdx].ACL.String())
	//t.Logf("%+v", acls[aclIdx].ACL.Name)
	//t.Logf("%+v", acls[aclIdx].ACL.Rules)
	//t.Logf("%+v", acls[aclIdx].ACL.Interfaces)
	//t.Logf("%+v", acls[aclIdx].ACL.XXX_MessageName())

	err = h.DeleteACL(5)
	Expect(err).To(Not(BeNil()))

	//ctx.MockVpp.MockReply(&acl_api.ACLDelReply{Retval: -1})
	//err = h.DeleteACL(5)
	//Expect(err).To(Not(BeNil()))

	err = h.DeleteACL(0)
	Expect(err).To(BeNil())

	acls, errx = h.DumpACL()

	if errx != nil {
		t.Fatalf("dumping acls failed: %v", errx)
	}
	t.Logf("%d acls dumped", len(acls))
	t.Logf("acls dumped %v", acls)

	for _, aclrecord := range acls {
		t.Logf("%+v", aclrecord.ACL)
		if aclrecord.Meta.Index == 0 {
			t.Fatalf("This acll should be deleted : %v", errx)
		}
	}

	rule2modify := []*acl.ACL_Rule{
		{
			Action: acl.ACL_Rule_PERMIT,
			IpRule: &acl.ACL_Rule_IpRule{
				Ip: &acl.ACL_Rule_IpRule_Ip{
					SourceNetwork:      "10.20.30.1/32",
					DestinationNetwork: "10.20.0.0/24",
				},
			},
		},
		{
			Action: acl.ACL_Rule_PERMIT,
			IpRule: &acl.ACL_Rule_IpRule{
				Ip: &acl.ACL_Rule_IpRule_Ip{
					SourceNetwork:      "dead:dead::3/64",
					DestinationNetwork: "dead:dead::4/64",
				},
			},
		},
	}

	err = h.ModifyACL(1, rule2modify, "test_modify0")
	Expect(err).To(BeNil())
	if err != nil {
		t.Logf("modifying of acl failed: %v", err)
	} else {
		t.Logf("acl was modified")
	}

	err = h.ModifyACL(1, aclErr1Rules, "test_modify1")
	Expect(err).To(Not(BeNil()))
	if err != nil {
		t.Logf("modifying of acl failed: %v", err)
	} else {
		t.Logf("acl was modified")
	}

	err = h.ModifyACL(1, aclNoRules, "test_modify2")
	Expect(err).To(BeNil())
	if err != nil {
		t.Logf("modifying of acl failed: %v", err)
	} else {
		t.Logf("acl was modified")
	}

	//ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	//err = h.ModifyACL(1, aclIPrules, "test_modify3")
	//Expect(err).To(Not(BeNil()))

	//ctx.MockVpp.MockReply(&acl_api.ACLAddReplaceReply{Retval: -1})
	//err = h.ModifyACL(1, aclIPrules, "test_modify4")
	//Expect(err).To(Not(BeNil()))

}

// Test add MACIP acl rules
func TestAddMacIPAcl(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppBinapi, logrus.NewLogger("test"))

	const ifName = "loop1"
	ifIdx, errI := ih.AddLoopbackInterface(ifName)
	if errI != nil {
		t.Fatalf("creating interface failed: %v", errI)
	}
	t.Logf("interface created %v", ifIdx)

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-iface1"), "test-iface1")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{
		SwIfIndex: ifIdx,
	})

	h := aclplugin_vppcalls.CompatibleACLVppHandler(ctx.vppBinapi, ifIndexes, logrus.NewLogger("test"))
	if h == nil {
		t.Fatalf("handler was not created")
	}

	aclIndex, err := h.AddMACIPACL(aclMACIPrules, "test6")
	Expect(err).To(BeNil())
	Expect(aclIndex).To(BeEquivalentTo(0))

	_, err = h.AddMACIPACL(aclNoRules, "test7")
	Expect(err).To(Not(BeNil()))

	_, err = h.AddMACIPACL(aclErr4Rules, "test8")
	Expect(err).To(Not(BeNil()))

	_, err = h.AddMACIPACL(aclErr5Rules, "test9")
	Expect(err).To(Not(BeNil()))

	_, err = h.AddMACIPACL(aclErr6Rules, "test10")
	Expect(err).To(Not(BeNil()))
	Expect(err.Error()).To(BeEquivalentTo("invalid IP address "))

	_, err = h.AddMACIPACL(aclMACIPrules, "test11")
	Expect(err).To(BeNil())

	_, err = h.AddMACIPACL(aclMACIPrules, "test12")
	Expect(err).To(BeNil())

	err = h.DeleteMACIPACL(5)
	Expect(err).To(Not(BeNil()))

	err = h.DeleteMACIPACL(1)
	Expect(err).To(BeNil())

	rule2modify := []*acl.ACL_Rule{
		{
			Action: acl.ACL_Rule_DENY,
			MacipRule: &acl.ACL_Rule_MacIpRule{
				SourceAddress:        "192.168.10.1",
				SourceAddressPrefix:  uint32(24),
				SourceMacAddress:     "11:44:0A:B8:4A:37",
				SourceMacAddressMask: "ff:ff:ff:ff:00:00",
			},
		},
		{
			Action: acl.ACL_Rule_DENY,
			MacipRule: &acl.ACL_Rule_MacIpRule{
				SourceAddress:        "dead::2",
				SourceAddressPrefix:  uint32(64),
				SourceMacAddress:     "11:44:0A:B8:4A:38",
				SourceMacAddressMask: "ff:ff:ff:ff:00:00",
			},
		},
	}

	err = h.ModifyMACIPACL(0, rule2modify, "test_modify0")
	Expect(err).To(BeNil())

	err = h.ModifyMACIPACL(0, aclErr1Rules, "test_modify1")
	Expect(err).To(Not(BeNil()))

	err = h.ModifyMACIPACL(0, aclMACIPrules, "test_modify4")
	Expect(err).To(BeNil())

	err = h.ModifyMACIPACL(0, aclIPrules, "test_modify5")
	Expect(err).To(Not(BeNil()))
}

func TestAcl(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppBinapi, logrus.NewLogger("test"))

	const ifName = "loop1"
	ifIdx, err := ih.AddLoopbackInterface(ifName)
	if err != nil {
		t.Fatalf("creating interface failed: %v", err)
	}
	t.Logf("interface created %v", ifIdx)

	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-iface1"), "test-iface1")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{
		SwIfIndex: ifIdx,
	})

	h := aclplugin_vppcalls.CompatibleACLVppHandler(ctx.vppBinapi, ifIndexes, logrus.NewLogger("test"))
	if h == nil {
		t.Fatalf("handler was not created")
	}

	acls, errx := h.DumpACL()
	if errx != nil {
		t.Fatalf("dumping acls failed: %v", errx)
	}
	t.Logf("%d acls dumped", len(acls))
	t.Logf("acls dumped %v", acls)

	aclName := "SpecialMakovica"
	var aclErr5Rules = []*acl.ACL_Rule{
		{
			Action: acl.ACL_Rule_PERMIT,
			MacipRule: &acl.ACL_Rule_MacIpRule{
				SourceAddress:        "192.168.0.1",
				SourceAddressPrefix:  uint32(16),
				SourceMacAddress:     "11:44:0A:B8:4A:36",
				SourceMacAddressMask: "ff:ff:ff:ff:00:00",
			},
		},
	}

	aclIdx, erry := h.AddMACIPACL(aclErr5Rules, aclName)
	if erry != nil {
		t.Fatalf("adding acls failed: %v", erry)
	}
	t.Logf("%d acl added", aclIdx)

	acls, errx = h.DumpMACIPACL()

	if errx != nil {
		t.Fatalf("dumping acls failed: %v", errx)
	}
	t.Logf("%d acls dumped", len(acls))
	t.Logf("acls dumped %v", acls)

	aclErr5Rules = []*acl.ACL_Rule{
		{
			Action: acl.ACL_Rule_PERMIT,
			MacipRule: &acl.ACL_Rule_MacIpRule{
				SourceAddress:        "192.168.0.2",
				SourceAddressPrefix:  uint32(16),
				SourceMacAddress:     "11:44:0A:B8:4A:38",
				SourceMacAddressMask: "ff:ff:ff:ff:00:00",
			},
		},
	}

	erry = h.ModifyACL(aclIdx, aclErr5Rules, aclName)
	if erry != nil {
		t.Fatalf("modifying acls failed: %v", erry)
	}
	t.Logf("%d acl modified", aclIdx)
	if errx != nil {
		t.Fatalf("dumping acls failed: %v", errx)
	}
	t.Logf("%d acls dumped", len(acls))
	t.Logf("acls dumped %v", acls)
}
