package vpp

import (
	"testing"

	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
)

func TestPing(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := vppcalls.CompatibleVpeHandler(ctx.Chan)

	if err := h.Ping(); err != nil {
		t.Fatalf("control ping failed: %v", err)
	}
}

func TestVersion(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := vppcalls.CompatibleVpeHandler(ctx.Chan)

	info, err := h.GetVersionInfo()
	if err != nil {
		t.Fatalf("getting version info failed: %v", err)
	}
	t.Logf("version info: %+v", info)
	if info.Version == "" {
		t.Error("invalid version info")
	}
}
