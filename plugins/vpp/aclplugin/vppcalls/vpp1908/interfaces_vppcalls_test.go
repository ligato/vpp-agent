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

package vpp1908

import (
	"testing"

	acl_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/acl"
	. "github.com/onsi/gomega"
)

// Test assignment of IP acl rule to given interface
func TestRequestSetACLToInterfaces(t *testing.T) {
	ctx := setupACLTest(t)
	defer ctx.teardownACLTest()

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     1,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err := ctx.aclHandler.SetACLToInterfacesAsIngress(0, []uint32{0})
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     1,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = ctx.aclHandler.SetACLToInterfacesAsEgress(0, []uint32{0})
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = ctx.aclHandler.SetACLToInterfacesAsIngress(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     1,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = ctx.aclHandler.SetACLToInterfacesAsIngress(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     1,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{Retval: -1})
	err = ctx.aclHandler.SetACLToInterfacesAsIngress(0, []uint32{0})
	Expect(err).To(Not(BeNil()))
}

// Test deletion of IP acl rule from given interface
func TestRequestRemoveInterfacesFromACL(t *testing.T) {
	ctx := setupACLTest(t)
	defer ctx.teardownACLTest()

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     1,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err := ctx.aclHandler.RemoveACLFromInterfacesAsIngress(0, []uint32{0})
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     1,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = ctx.aclHandler.RemoveACLFromInterfacesAsEgress(0, []uint32{0})
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = ctx.aclHandler.RemoveACLFromInterfacesAsEgress(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     1,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = ctx.aclHandler.RemoveACLFromInterfacesAsEgress(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 0,
		Count:     1,
		NInput:    1,
		Acls:      []uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{Retval: -1})
	err = ctx.aclHandler.RemoveACLFromInterfacesAsEgress(0, []uint32{0})
	Expect(err).To(Not(BeNil()))
}

// Test assignment of MACIP acl rule to given interface
func TestSetMacIPAclToInterface(t *testing.T) {
	ctx := setupACLTest(t)
	defer ctx.teardownACLTest()

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{})
	err := ctx.aclHandler.SetMACIPACLToInterfaces(0, []uint32{0})
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = ctx.aclHandler.SetMACIPACLToInterfaces(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{Retval: -1})
	err = ctx.aclHandler.SetMACIPACLToInterfaces(0, []uint32{0})
	Expect(err).To(Not(BeNil()))
}

// Test deletion of MACIP acl rule from given interface
func TestRemoveMacIPIngressACLFromInterfaces(t *testing.T) {
	ctx := setupACLTest(t)
	defer ctx.teardownACLTest()

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{})
	err := ctx.aclHandler.RemoveMACIPACLFromInterfaces(1, []uint32{0})
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = ctx.aclHandler.RemoveMACIPACLFromInterfaces(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{Retval: -1})
	err = ctx.aclHandler.RemoveMACIPACLFromInterfaces(0, []uint32{0})
	Expect(err).To(Not(BeNil()))
}
