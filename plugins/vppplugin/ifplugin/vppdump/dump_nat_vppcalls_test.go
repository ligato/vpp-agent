package vppdump

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	bin_api "github.com/ligato/vpp-agent/plugins/vppplugin/generated/bin_api/nat"
	"github.com/ligato/vpp-agent/plugins/vppplugin/generated/bin_api/vpe"
	"github.com/ligato/vpp-agent/plugins/vppplugin/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/tests/vppcallmock"

	. "github.com/onsi/gomega"
)

func TestNat44InterfaceDump(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bin_api.Nat44InterfaceDetails{
		SwIfIndex: 1,
		IsInside:  0,
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-sw_if_indexes", ifaceidx.IndexMetadata))
	swIfIndexes.RegisterName("if0", 1, nil)

	ifaces, err := nat44InterfaceDump(swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(ifaces).To(HaveLen(1))
	Expect(ifaces[0].IsInside).To(BeFalse())
}

func TestNat44InterfaceDump2(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bin_api.Nat44InterfaceDetails{
		SwIfIndex: 1,
		IsInside:  1,
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-sw_if_indexes", ifaceidx.IndexMetadata))
	swIfIndexes.RegisterName("if0", 1, nil)

	ifaces, err := nat44InterfaceDump(swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(ifaces).To(HaveLen(1))
	Expect(ifaces[0].IsInside).To(BeTrue())
}

func TestNat44InterfaceDump3(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&bin_api.Nat44InterfaceDetails{
		SwIfIndex: 1,
		IsInside:  2,
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	swIfIndexes := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-sw_if_indexes", ifaceidx.IndexMetadata))
	swIfIndexes.RegisterName("if0", 1, nil)

	ifaces, err := nat44InterfaceDump(swIfIndexes, logrus.DefaultLogger(), ctx.MockChannel, nil)
	Expect(err).To(Succeed())
	Expect(ifaces).To(HaveLen(2))
	Expect(ifaces[0].IsInside).To(BeFalse())
	Expect(ifaces[1].IsInside).To(BeTrue())
}
