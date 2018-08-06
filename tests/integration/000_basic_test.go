package integration

import (
	"testing"

	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
)

func TestVersion(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardown()

	channel, err := ctx.Conn.NewAPIChannel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()

	info, err := vppcalls.GetVersionInfo(channel)
	if err != nil {
		t.Fatalf("getting version info failed: %v", err)
		return
	}

	t.Logf("version info: %+v", info)
}
