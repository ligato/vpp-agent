package vppcalls

import (
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	acl_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
	"testing"
)

// Test assignment of IP acl rule to given interface
func TestRequestSetACLToInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", nil))
	interfaces := NewACLInterfacesVppCalls(logrus.DefaultLogger(), ctx.MockChannel, ifIndexes, nil)

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err := interfaces.SetACLToInterfacesAsIngress(0, []uint32{0})
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = interfaces.SetACLToInterfacesAsEgress(0, []uint32{0})
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = interfaces.SetACLToInterfacesAsIngress(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = interfaces.SetACLToInterfacesAsIngress(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{Retval: -1})
	err = interfaces.SetACLToInterfacesAsIngress(0, []uint32{0})
	Expect(err).To(Not(BeNil()))
}

// Test deletion of IP acl rule from given interface
func TestRequestRemoveInterfacesFromACL(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", nil))
	interfaces := NewACLInterfacesVppCalls(logrus.DefaultLogger(), ctx.MockChannel, ifIndexes, nil)

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err := interfaces.RemoveIPIngressACLFromInterfaces(0, []uint32{0})
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = interfaces.RemoveIPEgressACLFromInterfaces(0, []uint32{0})
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = interfaces.RemoveIPEgressACLFromInterfaces(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = interfaces.RemoveIPEgressACLFromInterfaces(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{Retval: -1})
	err = interfaces.RemoveIPEgressACLFromInterfaces(0, []uint32{0})
	Expect(err).To(Not(BeNil()))
}

// Test assignment of MACIP acl rule to given interface
func TestSetMacIPAclToInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", nil))
	interfaces := NewACLInterfacesVppCalls(logrus.DefaultLogger(), ctx.MockChannel, ifIndexes, nil)

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{})
	err := interfaces.SetMacIPAclToInterface(0, []uint32{0})
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = interfaces.SetMacIPAclToInterface(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{Retval: -1})
	err = interfaces.SetMacIPAclToInterface(0, []uint32{0})
	Expect(err).To(Not(BeNil()))
}

// Test deletion of MACIP acl rule from given interface
func TestRemoveMacIPIngressACLFromInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", nil))
	interfaces := ACLInterfacesVppCalls{
		logrus.DefaultLogger(),
		ctx.MockChannel,
		ifIndexes,
		nil,
		nil,
	}

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{})
	err := interfaces.RemoveMacIPIngressACLFromInterfaces(1, []uint32{0})
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = interfaces.RemoveMacIPIngressACLFromInterfaces(0, []uint32{0})
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{Retval: -1})
	err = interfaces.RemoveMacIPIngressACLFromInterfaces(0, []uint32{0})
	Expect(err).To(Not(BeNil()))
}
