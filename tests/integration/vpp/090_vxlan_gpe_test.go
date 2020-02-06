package vpp

import (
	"fmt"
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"

	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

func TestVxlanGpe(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))

	tests := []struct {
		name           string
		vxLan          *interfaces.VxlanLink
		mcastSwIfIndex uint32
		encapVrfID     uint32
		isFail         bool
	}{
		{
			name: "Create VxLAN-GPE tunnel (IP4)",
			vxLan: &interfaces.VxlanLink{
				SrcAddress: "20.30.40.50",
				DstAddress: "50.40.30.20",
				Gpe: &interfaces.VxlanLink_Gpe{
					Protocol: interfaces.VxlanLink_Gpe_IP4,
				},
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         false,
		},
		{
			name: "Create VxLAN-GPE tunnel (IP6)",
			vxLan: &interfaces.VxlanLink{
				SrcAddress: "20.30.40.50",
				DstAddress: "50.40.30.20",
				Gpe: &interfaces.VxlanLink_Gpe{
					Protocol: interfaces.VxlanLink_Gpe_IP6,
				},
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         false,
		},
		{
			name: "Create VxLAN-GPE tunnel (Ethernet)",
			vxLan: &interfaces.VxlanLink{
				SrcAddress: "20.30.40.50",
				DstAddress: "50.40.30.20",
				Gpe: &interfaces.VxlanLink_Gpe{
					Protocol: interfaces.VxlanLink_Gpe_ETHERNET,
				},
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         false,
		},
		{
			name: "Create VxLAN-GPE tunnel (NSH)",
			vxLan: &interfaces.VxlanLink{
				SrcAddress: "20.30.40.50",
				DstAddress: "50.40.30.20",
				Gpe: &interfaces.VxlanLink_Gpe{
					Protocol: interfaces.VxlanLink_Gpe_NSH,
				},
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         false,
		},
		{
			name: "Create VxLAN-GPE tunnel with same source and destination",
			vxLan: &interfaces.VxlanLink{
				SrcAddress: "20.30.40.50",
				DstAddress: "20.30.40.50",
				Gpe: &interfaces.VxlanLink_Gpe{
					Protocol: interfaces.VxlanLink_Gpe_IP4,
				},
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         true,
		},
		{
			name: "Create VxLAN-GPE tunnel with src and dst ip versions mismatch",
			vxLan: &interfaces.VxlanLink{
				SrcAddress: "20.30.40.50",
				DstAddress: "::1",
				Gpe: &interfaces.VxlanLink_Gpe{
					Protocol: interfaces.VxlanLink_Gpe_IP4,
				},
			},
			mcastSwIfIndex: 0xFFFFFFFF,
			isFail:         true,
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ifName := fmt.Sprintf("test%d", i)
			ifIdx, err := h.AddVxLanGpeTunnel(ifName, test.encapVrfID, test.mcastSwIfIndex, test.vxLan)

			if err != nil {
				if test.isFail {
					return
				}
				t.Fatalf("create VxLAN-GPE tunnel failed: %v\n", err)
			} else {
				if test.isFail {
					t.Fatal("create VxLAN-GPE tunnel must fail, but it's not")
				}
			}

			ifaces, err := h.DumpInterfaces(ctx.Context)
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}
			iface, ok := ifaces[ifIdx]
			if !ok {
				t.Fatalf("VxLAN-GPE interface was not found in dump")
			}

			vxLan := iface.Interface.GetVxlan()
			if test.vxLan.SrcAddress != vxLan.SrcAddress {
				t.Fatalf("expected source address <%s>, got: <%s>", test.vxLan.SrcAddress, vxLan.SrcAddress)
			}
			if test.vxLan.DstAddress != vxLan.DstAddress {
				t.Fatalf("expected destination address <%s>, got: <%s>", test.vxLan.DstAddress, vxLan.DstAddress)
			}
			if test.vxLan.Vni != vxLan.Vni {
				t.Fatalf("expected VNI <%d>, got: <%d>", test.vxLan.Vni, vxLan.Vni)
			}
			if test.vxLan.Multicast != vxLan.Multicast {
				t.Fatalf("expected multicast interface name <%s>, got: <%s>", test.vxLan.Multicast, vxLan.Multicast)
			}
			if test.vxLan.Gpe.Protocol != vxLan.Gpe.Protocol {
				t.Fatalf("expected VxLAN-GPE protocol <%d>, got: <%d>", test.vxLan.Gpe.Protocol, vxLan.Gpe.Protocol)
			}
			if test.vxLan.Gpe.DecapVrfId != vxLan.Gpe.DecapVrfId {
				t.Fatalf("expected VxLAN-GPE DecapVrfId <%d>, got: <%d>", test.vxLan.Gpe.DecapVrfId, vxLan.Gpe.DecapVrfId)
			}

			err = h.DeleteVxLanGpeTunnel(ifName, test.vxLan)
			if err != nil {
				t.Fatalf("delete VxLAN-GPE tunnel failed: %v\n", err)
			}

			ifaces, err = h.DumpInterfaces(ctx.Context)
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}

			if _, ok := ifaces[ifIdx]; ok {
				t.Fatalf("VxLAN-GPE interface was found in dump after removing")
			}
		})
	}
}
