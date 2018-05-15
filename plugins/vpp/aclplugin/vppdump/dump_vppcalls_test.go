package vppdump

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	acl_api "github.com/ligato/vpp-agent/plugins/vpp/generated/bin_api/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/generated/bin_api/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/tests/vppcallmock"

	. "github.com/onsi/gomega"
)

func TestGetIPRuleMatch(t *testing.T) {
	ipRule := getIPRuleMatches(acl_api.ACLRule{
		SrcIPAddr:      []byte{10, 0, 0, 1},
		SrcIPPrefixLen: 24,
		DstIPAddr:      []byte{20, 0, 0, 1},
		DstIPPrefixLen: 24,
		Proto:          vppcalls.ICMPv4Proto,
	})
	t.Logf("ip rule: %+v", ipRule)

	if ipRule.GetIcmp() == nil {
		t.Fatal("should have icmp match")
	}
}

func TestDumpACLInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&acl_api.ACLInterfaceListDetails{
		SwIfIndex: 1,
		Count:     2,
		NInput:    1,
		Acls:      []uint32{11, 22},
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-sw_if_indexes", ifaceidx.IndexMetadata))
	swIfIndexes.RegisterName("if0", 1, nil)

	indexes := []uint32{11, 22}
	ifaces, err := DumpACLInterfaces(indexes, swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(ifaces).To(HaveLen(2))
	Expect(ifaces[11].Ingress).To(Equal([]string{"if0"}))
	Expect(ifaces[22].Egress).To(Equal([]string{"if0"}))
}
