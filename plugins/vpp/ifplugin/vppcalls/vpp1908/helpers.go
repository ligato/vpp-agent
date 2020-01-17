package vpp1908

import (
	"bytes"
	"fmt"
	"net"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ip"
)

func IPToAddress(ipstr string) (addr ip.Address, err error) {
	netIP := net.ParseIP(ipstr)
	if netIP == nil {
		return ip.Address{}, fmt.Errorf("invalid IP: %q", ipstr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = ip.ADDRESS_IP6
		var ip6addr ip.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = ip.ADDRESS_IP4
		var ip4addr ip.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}

func uintToBool(value uint8) bool {
	return value != 0
}

func cleanString(b []byte) string {
	return string(bytes.SplitN(b, []byte{0x00}, 2)[0])
}
