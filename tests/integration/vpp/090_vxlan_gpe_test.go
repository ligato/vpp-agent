package vpp

import (
	"fmt"
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	ifplugin_vppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

func TestVxlanGpe(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppBinapi, logrus.NewLogger("test"))

	tests := []struct {
		name           string
		vxLan          *interfaces.VxlanLink
		mcastSwIfIndex uint32
		encapVrfID     uint32
		isFail         bool
	}{
		{
			name: "Create VxLAN-GPE tunnel",
			vxLan: &interfaces.VxlanLink{
				SrcAddress: "20.30.40.50",
				DstAddress: "50.40.30.20",
				Gpe: &interfaces.VxlanLink_Gpe{
					Protocol: interfaces.VxlanLink_Gpe_IP4,
				},
			},
			isFail: false,
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
			isFail: true,
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
			t.Logf("VxLAN-GPE interface created with index %d", ifIdx)

			ifaces, err := h.DumpInterfaces()
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}
			t.Logf("Interfaces:")
			for _, i := range ifaces {
				t.Logf("\t%+v\n", i)
			}

			vxlanGpes, err := h.DumpVxLanGpe(^uint32(0))
			if err != nil {
				t.Fatalf("dumping VxLAN-GPE failed: %v", err)
			}
			t.Logf("VxLAN-GPE tunnels:")
			for _, i := range vxlanGpes {
				t.Logf("\t%+v\n", i)
			}

			err = h.DelVxLanGpeTunnel(ifName, test.vxLan)
			if err != nil {
				t.Fatalf("delete VxLAN-GPE tunnel failed: %v\n", err)
			}

			ifaces, err = h.DumpInterfaces()
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}
			t.Logf("Interfaces:")
			for _, i := range ifaces {
				t.Logf("\t%+v\n", i)
			}

			vxlanGpes, err = h.DumpVxLanGpe(^uint32(0))
			if err != nil {
				t.Fatalf("dumping VxLAN-GPE failed: %v", err)
			}
			t.Logf("VxLAN-GPE tunnels:")
			for _, i := range vxlanGpes {
				t.Logf("\t%+v\n", i)
			}
		})
	}
}
