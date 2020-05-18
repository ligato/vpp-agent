package vpp

import (
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	ipfixplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls"
	vpp_ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin"
)

func TestFlowprobe(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	// Prepare an interface to test Flowprobe feature enable/disable actions.
	ih := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))
	Expect(ih).ToNot(BeNil(), "Interface VPP handler is not available")
	const ifName = "loop1"
	ifIdx, err := ih.AddLoopbackInterface(ifName)
	Expect(err).ToNot(HaveOccurred(), "failed to create an interface")
	t.Logf("interface created with SwIfIndex=%v", ifIdx)

	// Create interface index and add there previously created interface.
	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{SwIfIndex: ifIdx})

	// Finally, create and start using IPFIX VPP handler.
	h := ipfixplugin_vppcalls.CompatibleIpfixVppHandler(ctx.vppClient, ifIndexes, logrus.NewLogger("test"))
	Expect(h).ToNot(BeNil(), "IPFIX VPP handler is not available")

	// Try to set empty Flowprobe params. This must fail.
	Expect(h.SetFPParams(&vpp_ipfix.FlowProbeParams{})).ToNot(Succeed())

	// But if at least one of params is set, everything should be ok.
	Expect(h.SetFPParams(&vpp_ipfix.FlowProbeParams{RecordL2: true})).To(Succeed(),
		"setting new params for Flowprobe failed",
	)

	// Try to add Flowprobe feature to the interface.
	Expect(h.AddFPFeature(&vpp_ipfix.FlowProbeFeature{Interface: ifName})).To(Succeed(),
		"enabling Flowprobe feature for interface failed",
	)

	// Try to update Flowprobe feature for the interface.
	Expect(h.AddFPFeature(&vpp_ipfix.FlowProbeFeature{
		Interface: ifName,
		L2:        true,
	})).ToNot(Succeed(), "updating Flowprobe feature on interface was not expected to work")

	// So to update, first it needs to be deleted...
	Expect(h.DelFPFeature(&vpp_ipfix.FlowProbeFeature{Interface: ifName})).To(Succeed(),
		"removing Flowprobe feature for interface failed",
	)

	// ... and afterwards created.
	Expect(h.AddFPFeature(&vpp_ipfix.FlowProbeFeature{
		Interface: ifName,
		L2:        true,
	})).To(Succeed(), "enabling Flowprobe feature for interface failed")
}
