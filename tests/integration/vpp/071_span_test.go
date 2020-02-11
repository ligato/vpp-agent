package vpp

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"

	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
)

func TestSpan(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))

	tests := []struct {
		name string

		// SPAN params:
		swIfIndexFrom uint32
		swIfIndexTo   uint32
		direction     uint8
		isL2          uint8

		// If dump must return record (true) or empty slice (false):
		isDump bool
		// Action:
		isAdd bool
		// If action must fail:
		isFail bool
	}{
		{"enable Rx SPAN", 0, 1, uint8(vpp_interfaces.Span_RX), 0, true, true, false},
		{"enable Tx SPAN", 0, 1, uint8(vpp_interfaces.Span_TX), 0, true, true, false},
		{"enable Both SPAN", 0, 1, uint8(vpp_interfaces.Span_BOTH), 0, true, true, false},
		{"disable SPAN", 0, 1, uint8(vpp_interfaces.Span_BOTH), 0, false, false, false},
		{"enable SPAN with L2 set", 0, 1, uint8(vpp_interfaces.Span_RX), 1, true, true, false},
		{"disable SPAN with L2 set", 0, 1, uint8(vpp_interfaces.Span_RX), 1, false, false, false},
		{"enable bad SPAN", 0, 0, uint8(vpp_interfaces.Span_BOTH), 0, false, true, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err error

			if test.isAdd {
				err = h.AddSpan(test.swIfIndexFrom, test.swIfIndexTo, test.direction, test.isL2)
			} else {
				err = h.DelSpan(test.swIfIndexFrom, test.swIfIndexTo, test.isL2)
			}

			if test.isFail && err == nil {
				t.Fatal("must fail, but no error returned from action")
			} else if !test.isFail && err != nil {
				t.Fatalf("action failed: %v\n", err)
			}

			dumpResp, err := h.DumpSpan()
			if err != nil {
				t.Fatalf("dump span failed: %v\n", err)
			}

			if test.isDump {
				if len(dumpResp) != 1 {
					t.Fatalf("wrong number of SPANs in dump. Expected: 1. Got: %d\n", len(dumpResp))
				}

				if dumpResp[0].SwIfIndexFrom != test.swIfIndexFrom {
					t.Fatalf("wrong SwIfIndexFrom. Expected: %d. Got: %d\n", test.swIfIndexFrom, dumpResp[0].SwIfIndexFrom)
				}
				if dumpResp[0].SwIfIndexTo != test.swIfIndexTo {
					t.Fatalf("wrong SwIfIndexTo. Expected: %d. Got: %d\n", test.swIfIndexTo, dumpResp[0].SwIfIndexTo)
				}
				if dumpResp[0].Direction != test.direction {
					t.Fatalf("wrong Direction. Expected: %d. Got: %d\n", test.direction, dumpResp[0].Direction)
				}
				if dumpResp[0].IsL2 != test.isL2 {
					t.Fatalf("wrong IsL2. Expected: %d. Got: %d\n", test.isL2, dumpResp[0].IsL2)
				}
			} else {
				if len(dumpResp) != 0 {
					t.Fatalf("wrong number of SPANs in dump. Expected: 0. Got: %d\n", len(dumpResp))
				}
			}

		})
	}
}
