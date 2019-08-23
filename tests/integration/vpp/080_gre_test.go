package vpp

import (
	"fmt"
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	ifplugin_vppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

func TestGre(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppBinapi, logrus.NewLogger("test"))

	tests := []struct {
		name    string
		greLink *interfaces.GreLink
		isFail  bool
	}{
		{
			name: "create ERSPAN GRE tunnel with IPv4",
			greLink: &interfaces.GreLink{
				TunnelType: interfaces.GreLink_ERSPAN,
				SrcAddr:    "2000::8:23",
				DstAddr:    "2019::8:23",
			},
			isFail: false,
		},
		{
			name: "create ERSPAN GRE tunnel with out of range session id",
			greLink: &interfaces.GreLink{
				TunnelType: interfaces.GreLink_ERSPAN,
				SrcAddr:    "10.10.10.10",
				DstAddr:    "20.20.20.20",
				SessionId:  1024,
			},
			isFail: true,
		},
		{
			name: "create GRE tunnel with bad source address",
			greLink: &interfaces.GreLink{
				TunnelType: interfaces.GreLink_ERSPAN,
				SrcAddr:    "badip",
				DstAddr:    "20.20.20.20",
			},
			isFail: true,
		},
		{
			name: "create GRE tunnel with destination address not set",
			greLink: &interfaces.GreLink{
				TunnelType: interfaces.GreLink_ERSPAN,
				SrcAddr:    "10.10.10.10",
			},
			isFail: true,
		},

		{
			name: "create GRE tunnel with equal source and destination addresses",
			greLink: &interfaces.GreLink{
				TunnelType: interfaces.GreLink_ERSPAN,
				SrcAddr:    "10.10.10.10",
				DstAddr:    "10.10.10.10",
			},
			isFail: true,
		},
		{
			name: "create GRE tunnel with addresses in IPv4 and IPv6",
			greLink: &interfaces.GreLink{
				TunnelType: interfaces.GreLink_ERSPAN,
				SrcAddr:    "10.10.10.10",
				DstAddr:    "2019::8:23",
			},
			isFail: true,
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ifIdx, err := h.AddGreTunnel(fmt.Sprintf("test%d", i), test.greLink)

			if err != nil {
				if test.isFail {
					return
				}
				t.Fatalf("create GRE tunnel failed: %v\n", err)
			}

			gres, err := h.DumpGre(ifIdx)
			if err != nil {
				t.Fatalf("dump GRE tunnels failed: %v\n", err)
			}

			if len(gres) != 1 {
				t.Fatalf("expected 1 GRE tunnel, got: %d", len(gres))
			}

			gre := gres[0]

			if uint8(test.greLink.TunnelType) != gre.TunnelType {
				t.Fatalf("expected tunnel type address <%d>, got: <%d>", test.greLink.TunnelType, gre.TunnelType)
			}
			if test.greLink.SrcAddr != gre.SrcAddress.String() {
				t.Fatalf("expected source address <%s>, got: <%s>", test.greLink.SrcAddr, gre.SrcAddress)
			}
			if test.greLink.DstAddr != gre.DstAddress.String() {
				t.Fatalf("expected destination address <%s>, got: <%s>", test.greLink.DstAddr, gre.DstAddress)
			}
			if test.greLink.OuterFibId != gre.OuterFibID {
				t.Fatalf("expected outer FIB id <%d>, got: <%d>", test.greLink.OuterFibId, gre.OuterFibID)
			}
			if uint16(test.greLink.SessionId) != gre.SessionID {
				t.Fatalf("expected session id <%d>, got: <%d>", test.greLink.SessionId, gre.SessionID)
			}
		})
	}

	t.Log("All interfaces:")
	ifaces, err := h.DumpInterfaces()
	if err != nil {
		t.Fatalf("dump interfaces failed: %v\n", err)
	}
	for _, i := range ifaces {
		t.Logf("\t%+v\n", i)
	}
}
