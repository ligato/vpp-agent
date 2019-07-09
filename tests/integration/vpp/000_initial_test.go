package vpp

import (
	"testing"

	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
)

func TestVersion(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	ch, err := ctx.Conn.NewAPIChannel()
	if err != nil {
		t.Fatalf("creating channel failed: %v", err)
	}
	defer ch.Close()

	h := vppcalls.CompatibleVpeHandler(ch)

	if err := h.Ping(); err != nil {
		t.Fatalf("control ping failed: %v", err)
	}

	info, err := h.GetVersionInfo()
	if err != nil {
		t.Fatalf("getting version info failed: %v", err)
	}
	t.Logf("version info: %+v", info)
	if info.Version == "" {
		t.Error("invalid version info")
	}
}
