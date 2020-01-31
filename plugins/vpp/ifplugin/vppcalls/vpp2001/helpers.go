package vpp2001

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
)

// IPToAddress converts string type IP address to VPP ip.api address representation
func IPToAddress(ipStr string) (addr ip.Address, err error) {
	netIP := net.ParseIP(ipStr)
	if netIP == nil {
		return ip.Address{}, fmt.Errorf("invalid IP: %q", ipStr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = ip_types.ADDRESS_IP6
		var ip6addr ip.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = ip_types.ADDRESS_IP4
		var ip4addr ip.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}

func uintToBool(value uint8) bool {
	return value != 0
}

func cleanString(s string) string {
	return strings.SplitN(s, "\x00", 2)[0]
}

func cleanBytes(b []byte) string {
	return string(bytes.SplitN(b, []byte{0x00}, 2)[0])
}
