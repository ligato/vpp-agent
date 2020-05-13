package vpp

import (
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ipfixplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls"
	vpp_ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin"
)

func TestIPFIX(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ifIdx := ifaceidx.NewIfaceIndex(logrus.NewLogger("test-if"), "test-if")
	h := ipfixplugin_vppcalls.CompatibleIpfixVppHandler(ctx.vppClient, ifIdx, logrus.NewLogger("test"))
	Expect(h).ToNot(BeNil(), "IPFIX VPP handler is not available")

	// Check default IPFIX configuration.
	exporters, err := h.DumpExporters()
	Expect(err).ToNot(HaveOccurred(), "failed to dump IPFIX configuration")
	Expect(exporters).To(HaveLen(1), "dump must return only one record")
	Expect(exporters[0].GetCollector().GetAddress()).To(Equal("0.0.0.0"), "unexpected initial address of collector")

	tests := []struct {
		name       string
		ipfix      *vpp_ipfix.IPFIX
		shouldFail bool
	}{
		{
			name: "Simple test",
			ipfix: &vpp_ipfix.IPFIX{
				Collector: &vpp_ipfix.IPFIX_Collector{
					Address: "10.10.10.10",
				},
				SourceAddress: "20.20.20.20",
			},
			shouldFail: false,
		},
		{
			name: "Collector IP 0.0.0.0 fail",
			ipfix: &vpp_ipfix.IPFIX{
				Collector: &vpp_ipfix.IPFIX_Collector{
					Address: "0.0.0.0",
				},
				SourceAddress: "20.20.20.20",
			},
			shouldFail: true,
		},
		{
			name: "Source IP 0.0.0.0 fail",
			ipfix: &vpp_ipfix.IPFIX{
				Collector: &vpp_ipfix.IPFIX_Collector{
					Address: "20.20.20.20",
				},
				SourceAddress: "0.0.0.0",
			},
			shouldFail: true,
		},
		{
			name: "Collector IP6 fail",
			ipfix: &vpp_ipfix.IPFIX{
				Collector: &vpp_ipfix.IPFIX_Collector{
					Address: "2020::4:7",
				},
				SourceAddress: "20.20.20.20",
			},
			shouldFail: true,
		},
		{
			name: "Source IP6 fail",
			ipfix: &vpp_ipfix.IPFIX{
				Collector: &vpp_ipfix.IPFIX_Collector{
					Address: "20.20.20.20",
				},
				SourceAddress: "2020::4:7",
			},
			shouldFail: true,
		},
		{
			name: "MTU is in range",
			ipfix: &vpp_ipfix.IPFIX{
				Collector: &vpp_ipfix.IPFIX_Collector{
					Address: "10.10.10.10",
				},
				SourceAddress: "20.20.20.20",
				PathMtu:       256,
			},
			shouldFail: false,
		},
		{
			name: "Too small MTU fail",
			ipfix: &vpp_ipfix.IPFIX{
				Collector: &vpp_ipfix.IPFIX_Collector{
					Address: "10.10.10.10",
				},
				SourceAddress: "20.20.20.20",
				PathMtu:       1,
			},
			shouldFail: true,
		},
		{
			name: "Too big MTU fail",
			ipfix: &vpp_ipfix.IPFIX{
				Collector: &vpp_ipfix.IPFIX_Collector{
					Address: "10.10.10.10",
				},
				SourceAddress: "20.20.20.20",
				PathMtu:       9999,
			},
			shouldFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.shouldFail {
				Expect(h.SetExporter(test.ipfix)).ToNot(Succeed())
				return
			}

			Expect(h.SetExporter(test.ipfix)).To(Succeed())

			exporters, err := h.DumpExporters()
			Expect(err).ToNot(HaveOccurred(), "failed to dump IPFIX configuration")
			Expect(exporters).To(HaveLen(1), "dump must return only one record")
			e := exporters[0]
			Expect(e.GetCollector().GetAddress()).To(Equal(test.ipfix.GetCollector().GetAddress()))
			Expect(e.GetSourceAddress()).To(Equal(test.ipfix.GetSourceAddress()))
		})
	}
}
