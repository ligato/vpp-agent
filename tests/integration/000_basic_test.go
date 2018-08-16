package integration

import (
	"testing"

	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
)

func TestVersion(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	channel, err := ctx.Conn.NewAPIChannel()
	if err != nil {
		t.Fatalf("creating channel failed: %v", err)
	}
	defer channel.Close()

	info, err := vppcalls.GetVersionInfo(channel)
	if err != nil {
		t.Fatalf("getting version info failed: %v", err)
	}

	t.Logf("version info: %+v", info)
}
