package vppdump

import (
	"testing"

	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppcalls"
	acl_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/acl"
)

func TestGetIPRuleMatch(t *testing.T) {
	ipRule := getIPRuleMatches(acl_api.ACLRule{
		SrcIPAddr:      []byte{10, 0, 0, 1},
		SrcIPPrefixLen: 24,
		DstIPAddr:      []byte{20, 0, 0, 1},
		DstIPPrefixLen: 24,
		Proto:          vppcalls.ICMPv4Proto,
	})
	t.Logf("ip rule: %+v", ipRule)

	if ipRule.GetIcmp() == nil {
		t.Fatal("should have icmp match")
	}
}
