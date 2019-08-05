package vpp

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"

	_ "github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	ifplugin_vppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

func TestSpan(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppBinapi, logrus.NewLogger("test"))

	// put to API
	const (
		stateDisable = iota
		stateRx
		stateTx
		stateBoth
	)

	// dump interfaces and log them
	ifaces, err := h.DumpInterfaces()
	if err != nil {
		t.Fatalf("dump interfaces failed: %v\n", err)
	}

	for idx, iface := range ifaces {
		t.Logf("%d. %+v\n", idx, iface)
	}

	// test invalid states

	// -----------------------------------------
	t.Log("enable SPAN from if0 to if1 Rx only")
	err = h.AddSpan(0, 1, stateRx, 0)
	if err != nil {
		t.Fatalf("enable span failed: %v\n", err)
	}

	dumpResp, err := h.DumpSpan()
	if err != nil {
		t.Fatalf("dump span failed: %v\n", err)
	}

	if len(dumpResp) != 1 {
		t.Fatalf("wrong number of SPANs in dump. Expected: 1. Got: %d\n", len(dumpResp))
	}

	// test other fields in dumpResp[0] too

	if dumpResp[0].State != stateRx {
		t.Fatalf("SPAN was created with wrong state. Expected: %d. Got: %d\n", stateRx, dumpResp[0].State)
	}

	// -----------------------------------------
	t.Log("update SPAN to forward Tx packets only")
	err = h.AddSpan(0, 1, stateTx, 0)
	if err != nil {
		t.Fatalf("update span failed: %v\n", err)
	}

	dumpResp, err = h.DumpSpan()
	if err != nil {
		t.Fatalf("dump span failed: %v\n", err)
	}

	if len(dumpResp) != 1 {
		t.Fatalf("wrong number of SPANs in dump. Expected: 1. Got: %d\n", len(dumpResp))
	}

	if dumpResp[0].State != stateTx {
		t.Fatalf("SPAN was not updated. Expected: %d. Got: %d\n", stateTx, dumpResp[0].State)
	}

	// -----------------------------------------
	t.Log("update SPAN to forward both Rx and Tx packets")
	err = h.AddSpan(0, 1, stateBoth, 0)
	if err != nil {
		t.Fatalf("update span failed: %v\n", err)
	}

	dumpResp, err = h.DumpSpan()
	if err != nil {
		t.Fatalf("dump span failed: %v\n", err)
	}

	if len(dumpResp) != 1 {
		t.Fatalf("wrong number of SPANs in dump. Expected: 1. Got: %d\n", len(dumpResp))
	}

	if dumpResp[0].State != stateBoth {
		t.Fatalf("SPAN was not updated. Expected: %d. Got: %d\n", stateBoth, dumpResp[0].State)
	}

	// -----------------------------------------
	t.Log("disable SPAN")
	err = h.DelSpan(0, 1, 0)
	if err != nil {
		t.Fatalf("delete span failed: %v\n", err)
	}

	dumpResp, err = h.DumpSpan()
	if err != nil {
		t.Fatalf("dump span failed: %v\n", err)
	}

	if len(dumpResp) != 0 {
		t.Fatalf("wrong number of SPANs in dump. Expected: 0. Got: %d\n", len(dumpResp))
	}
}
