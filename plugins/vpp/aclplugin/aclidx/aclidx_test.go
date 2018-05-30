package aclidx_test

import (
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin/aclidx"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
	. "github.com/onsi/gomega"
	"testing"
)

func aclIndexTestInitialization(t *testing.T) (idxvpp.NameToIdxRW, aclidx.AclIndexRW) {
	RegisterTestingT(t)

	// initialize index
	nameToIdx := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "index_test", nil)
	index := aclidx.NewAclIndex(nameToIdx)
	names := nameToIdx.ListNames()

	// check if names were empty
	Expect(names).To(BeEmpty())

	return index.GetMapping(), index
}

var acldata = acl.AccessLists_Acl{
	AclName:    "acl1",
	Rules:      []*acl.AccessLists_Acl_Rule{{AclAction: acl.AclAction_PERMIT}},
	Interfaces: &acl.AccessLists_Acl_Interfaces{},
}

// Tests registering and unregistering name to index
func TestRegisterAndUnregisterName(t *testing.T) {
	mapping, index := aclIndexTestInitialization(t)

	// Register entry
	index.RegisterName("acl1", 0, &acldata)
	names := mapping.ListNames()
	Expect(names).To(HaveLen(1))
	Expect(names).To(ContainElement("acl1"))

	// Unregister entry
	index.UnregisterName("acl1")
	names = mapping.ListNames()
	Expect(names).To(BeEmpty())
}

func TestLookupIndex(t *testing.T) {
	RegisterTestingT(t)

	_, aclIndex := aclIndexTestInitialization(t)

	aclIndex.RegisterName("acl", 0, &acldata)

	foundName, acl, exist := aclIndex.LookupName(0)
	Expect(exist).To(BeTrue())
	Expect(foundName).To(Equal("acl"))
	Expect(acl.AclName).To(Equal("acl1"))
}

func TestLookupName(t *testing.T) {
	RegisterTestingT(t)

	_, aclIndex := aclIndexTestInitialization(t)

	aclIndex.RegisterName("acl", 0, &acldata)

	foundName, acl, exist := aclIndex.LookupIdx("acl")
	Expect(exist).To(BeTrue())
	Expect(foundName).To(Equal(uint32(0)))
	Expect(acl.AclName).To(Equal("acl1"))
}

func TestWatchNameToIdx(t *testing.T) {
	RegisterTestingT(t)

	_, aclIndex := aclIndexTestInitialization(t)

	c := make(chan aclidx.AclIdxDto)
	aclIndex.WatchNameToIdx("testName", c)

	aclIndex.RegisterName("aclX", 0, &acldata)

	var dto aclidx.AclIdxDto
	Eventually(c).Should(Receive(&dto))
	Expect(dto.Name).To(Equal("aclX"))
	Expect(dto.NameToIdxDtoWithoutMeta.Idx).To(Equal(uint32(0)))
}
