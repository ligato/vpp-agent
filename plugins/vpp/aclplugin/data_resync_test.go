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

package aclplugin_test

import (
	acl_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
	. "github.com/onsi/gomega"
	"testing"
)

var acls = []*acl.AccessLists_Acl{
	{AclName: "acl1",
		Rules: []*acl.AccessLists_Acl_Rule{
			{
				AclAction: acl.AclAction_PERMIT,
				Match: &acl.AccessLists_Acl_Rule_Match{
					IpRule: &acl.AccessLists_Acl_Rule_Match_IpRule{
						Ip: &acl.AccessLists_Acl_Rule_Match_IpRule_Ip{
							SourceNetwork:      "192.168.1.1/32",
							DestinationNetwork: "10.20.0.1/24",
						},
					},
				},
			},
		},
		Interfaces: &acl.AccessLists_Acl_Interfaces{
			Ingress: []string{"if1"},
			Egress:  []string{"if2"},
		},
	},
	{AclName: "acl2",
		Rules: []*acl.AccessLists_Acl_Rule{
			{
				AclAction: acl.AclAction_PERMIT,
				Match: &acl.AccessLists_Acl_Rule_Match{
					MacipRule: &acl.AccessLists_Acl_Rule_Match_MacIpRule{
						SourceAddress:        "192.168.0.1",
						SourceAddressPrefix:  uint32(16),
						SourceMacAddress:     "11:44:0A:B8:4A:35",
						SourceMacAddressMask: "ff:ff:ff:ff:00:00",
					},
				},
			},
		},
		Interfaces: &acl.AccessLists_Acl_Interfaces{
			Ingress: []string{"if3"},
			Egress:  nil,
		},
	},
}

// Test synchronisation - writes ACLs to the empty VPP
func TestResyncEmpty(t *testing.T) {
	// Setup
	ctx, connection, plugin, log := aclTestSetup(t, false)
	defer aclTestTeardown(connection, plugin)

	ctx.MockVpp.MockReply(&acl_api.ACLDetails{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ctx.MockVpp.MockReply(&acl_api.ACLAddReplaceReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReply{})

	err := plugin.Resync(acls, log)
	Expect(err).To(BeNil())
}

// Test synchronisation - writes ACLs to the already configured VPP
func TestResyncConfigured(t *testing.T) {
	// Setup
	ctx, connection, plugin, log := aclTestSetup(t, false)
	defer aclTestTeardown(connection, plugin)

	ctx.MockVpp.MockReply(&acl_api.ACLDetails{
		ACLIndex: 0,
		Tag:      []byte{'a', 'c', 'l', '1'},
		Count:    1,
		R:        []acl_api.ACLRule{{IsPermit: 1}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{0, 2},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{
		ACLIndex: 0,
		Tag:      []byte{'a', 'c', 'l', '2'},
		Count:    2,
		R:        []acl_api.MacipACLRule{{IsPermit: 0}, {IsPermit: 2}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     1,
		Acls:      []uint32{1},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ctx.MockVpp.MockReply(&acl_api.ACLDelReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLDelReply{})

	ctx.MockVpp.MockReply(&acl_api.ACLAddReplaceReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReply{})

	err := plugin.Resync(acls, log)
	Expect(err).To(BeNil())
}

// Test Resync with error when removig existing IP ACL
func TestResyncErr1(t *testing.T) {
	// Setup
	ctx, connection, plugin, log := aclTestSetup(t, false)
	defer aclTestTeardown(connection, plugin)

	ctx.MockVpp.MockReply(&acl_api.ACLDetails{
		ACLIndex: 0,
		Tag:      []byte{'a', 'c', 'l', '1'},
		Count:    1,
		R:        []acl_api.ACLRule{{IsPermit: 1}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{0, 2},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{
		ACLIndex: 0,
		Tag:      []byte{'a', 'c', 'l', '2'},
		Count:    2,
		R:        []acl_api.MacipACLRule{{IsPermit: 0}, {IsPermit: 2}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     1,
		Acls:      []uint32{1},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ctx.MockVpp.MockReply(&acl_api.MacipACLDelReply{})
	//ctx.MockVpp.MockReply(&acl_api.MacipACLDelReply{})

	err := plugin.Resync(acls, log)
	Expect(err).To(Not(BeNil()))
}

// Test Resync with error when removig existing IP ACL
func TestResyncErr2(t *testing.T) {
	// Setup
	ctx, connection, plugin, log := aclTestSetup(t, false)
	defer aclTestTeardown(connection, plugin)

	ctx.MockVpp.MockReply(&acl_api.ACLDetails{
		ACLIndex: 0,
		Tag:      []byte{'a', 'c', 'l', '1'},
		Count:    1,
		R:        []acl_api.ACLRule{{IsPermit: 1}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{0, 2},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{
		ACLIndex: 0,
		Tag:      []byte{'a', 'c', 'l', '2'},
		Count:    2,
		R:        []acl_api.MacipACLRule{{IsPermit: 0}, {IsPermit: 2}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     1,
		Acls:      []uint32{1},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ctx.MockVpp.MockReply(&acl_api.ACLDelReply{})
	ctx.MockVpp.MockReply(&acl_api.ACLDelReply{})

	err := plugin.Resync(acls, log)
	Expect(err).To(Not(BeNil()))
}

// Test Resync with error when configuring new ALCs
func TestResyncErr3(t *testing.T) {
	// Setup
	ctx, connection, plugin, log := aclTestSetup(t, false)
	defer aclTestTeardown(connection, plugin)

	ctx.MockVpp.MockReply(&acl_api.ACLDetails{
		ACLIndex: 0,
		Tag:      []byte{'a', 'c', 'l', '1'},
		Count:    1,
		R:        []acl_api.ACLRule{{IsPermit: 1}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{0, 2},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ctx.MockVpp.MockReply(&acl_api.MacipACLDetails{
		ACLIndex: 0,
		Tag:      []byte{'a', 'c', 'l', '2'},
		Count:    2,
		R:        []acl_api.MacipACLRule{{IsPermit: 0}, {IsPermit: 2}},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     1,
		Acls:      []uint32{1},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	ctx.MockVpp.MockReply(&acl_api.ACLDelReply{})
	ctx.MockVpp.MockReply(&acl_api.MacipACLDelReply{})

	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	err := plugin.Resync(acls, log)
	Expect(err).To(Not(BeNil()))

}
