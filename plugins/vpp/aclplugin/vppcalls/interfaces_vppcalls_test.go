package vppcalls

import (
	"testing"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/cn-infra/logging/logrus"
	acl_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/acl"
	. "github.com/onsi/gomega"
)

func TestRequestSetACLToInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	interfaces := NewACLInterfacesVppCalls(ctx.MockChannel, ifIndexes, nil)

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err := interfaces.SetACLToInterfacesAsIngress(0, []uint32 {0},logrus.DefaultLogger())
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = interfaces.SetACLToInterfacesAsEgress(0, []uint32 {0},logrus.DefaultLogger())
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = interfaces.SetACLToInterfacesAsIngress(0, []uint32 {0},logrus.DefaultLogger())
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = interfaces.SetACLToInterfacesAsIngress(0, []uint32 {0},logrus.DefaultLogger())
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{Retval:-1})
	err = interfaces.SetACLToInterfacesAsIngress(0, []uint32 {0},logrus.DefaultLogger())
	Expect(err).To(Not(BeNil()))
}

func TestRequestRemoveInterfacesFromACL(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	interfaces := NewACLInterfacesVppCalls(ctx.MockChannel, ifIndexes, nil)

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err := interfaces.RemoveIPIngressACLFromInterfaces(0, []uint32 {0},logrus.DefaultLogger())
	Expect(err).To(BeNil())

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = interfaces.RemoveIPEgressACLFromInterfaces(0, []uint32 {0},logrus.DefaultLogger())
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{})
	err = interfaces.RemoveIPEgressACLFromInterfaces(0, []uint32 {0},logrus.DefaultLogger())
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = interfaces.RemoveIPEgressACLFromInterfaces(0, []uint32 {0},logrus.DefaultLogger())
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		0,
		1,
		1,
		[]uint32{0, 1},
	})
	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceSetACLListReply{Retval:-1})
	err = interfaces.RemoveIPEgressACLFromInterfaces(0, []uint32 {0},logrus.DefaultLogger())
	Expect(err).To(Not(BeNil()))
}

func TestSetMacIPAclToInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	interfaces := NewACLInterfacesVppCalls(ctx.MockChannel, ifIndexes, nil)

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{})
	err := interfaces.SetMacIPAclToInterface(0, []uint32 {0}, logrus.DefaultLogger())
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = interfaces.SetMacIPAclToInterface(0, []uint32 {0}, logrus.DefaultLogger())
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{Retval:-1})
	err = interfaces.SetMacIPAclToInterface(0, []uint32 {0}, logrus.DefaultLogger())
	Expect(err).To(Not(BeNil()))
}

func TestRemoveMacIPIngressACLFromInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ifIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-plugin", "if", nil))
	interfaces := ACLInterfacesVppCalls{
		ctx.MockChannel,
		ifIndexes,
		nil,
		nil,
	}

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{})
	err := interfaces.RemoveMacIPIngressACLFromInterfaces(1, []uint32 {0}, logrus.DefaultLogger())
	Expect(err).To(BeNil())

	// error cases

	ctx.MockVpp.MockReply(&acl_api.MacipACLAddReplaceReply{})
	err = interfaces.RemoveMacIPIngressACLFromInterfaces(0, []uint32 {0}, logrus.DefaultLogger())
	Expect(err).To(Not(BeNil()))

	ctx.MockVpp.MockReply(&acl_api.MacipACLInterfaceAddDelReply{Retval:-1})
	err = interfaces.RemoveMacIPIngressACLFromInterfaces(0, []uint32 {0}, logrus.DefaultLogger())
	Expect(err).To(Not(BeNil()))
}
