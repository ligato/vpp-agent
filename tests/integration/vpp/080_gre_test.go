package vpp

import (
	"fmt"
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"

	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

func TestGre(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))

	tests := []struct {
		name    string
		greLink *interfaces.GreLink
		isFail  bool
	}{
		{
			name: "create UNKNOWN GRE tunnel with IPv4",
			greLink: &interfaces.GreLink{
				TunnelType: interfaces.GreLink_UNKNOWN,
				SrcAddr:    "2000::8:13",
				DstAddr:    "2019::8:13",
			},
			isFail: true,
		},
		{
			name: "create L3 GRE tunnel with IPv4",
			greLink: &interfaces.GreLink{
				TunnelType: interfaces.GreLink_L3,
				SrcAddr:    "2000::8:23",
				DstAddr:    "2019::8:23",
			},
			isFail: false,
		},
		{
			name: "create TEB GRE tunnel with IPv4",
			greLink: &interfaces.GreLink{
				TunnelType: interfaces.GreLink_TEB,
				SrcAddr:    "2000::8:33",
				DstAddr:    "2019::8:33",
			},
			isFail: false,
		},
		{
			name: "create ERSPAN GRE tunnel with IPv4",
			greLink: &interfaces.GreLink{
				TunnelType: interfaces.GreLink_ERSPAN,
				SrcAddr:    "2000::8:43",
				DstAddr:    "2019::8:43",
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
			ifName := fmt.Sprintf("test%d", i)
			ifIdx, err := h.AddGreTunnel(ifName, test.greLink)

			if err != nil {
				if test.isFail {
					return
				}
				t.Fatalf("create GRE tunnel failed: %v\n", err)
			} else {
				if test.isFail {
					t.Fatal("create GRE tunnel must fail, but it's not")
				}
			}

			ifaces, err := h.DumpInterfaces(ctx.Context)
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}
			iface, ok := ifaces[ifIdx]
			if !ok {
				t.Fatalf("GRE interface not found in dump")
			}

			gre := iface.Interface.GetGre()

			if test.greLink.TunnelType != gre.TunnelType {
				t.Fatalf("expected tunnel type <%s>, got: <%s>", test.greLink.TunnelType, gre.TunnelType)
			}
			if test.greLink.SrcAddr != gre.SrcAddr {
				t.Fatalf("expected source address <%s>, got: <%s>", test.greLink.SrcAddr, gre.SrcAddr)
			}
			if test.greLink.DstAddr != gre.DstAddr {
				t.Fatalf("expected destination address <%s>, got: <%s>", test.greLink.DstAddr, gre.DstAddr)
			}
			if test.greLink.OuterFibId != gre.OuterFibId {
				t.Fatalf("expected outer FIB id <%d>, got: <%d>", test.greLink.OuterFibId, gre.OuterFibId)
			}
			if test.greLink.SessionId != gre.SessionId {
				t.Fatalf("expected session id <%d>, got: <%d>", test.greLink.SessionId, gre.SessionId)
			}

			ifIdx, err = h.DelGreTunnel(ifName, test.greLink)
			if err != nil {
				t.Fatalf("delete GRE tunnel failed: %v\n", err)
			}

			ifaces, err = h.DumpInterfaces(ctx.Context)
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}
			iface, ok = ifaces[ifIdx]
			if ok {
				t.Fatalf("GRE interface was found in dump")
			}
		})
	}
}
