package vpp2005

import (
	"fmt"
	"net"
	"strings"

	"go.ligato.io/cn-infra/v2/utils/addrs"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005/ip_types"
)

// IPToAddress converts string type IP address to VPP ip.api address representation
func IPToAddress(ipStr string) (addr ip_types.Address, err error) {
	netIP := net.ParseIP(ipStr)
	if netIP == nil {
		return ip_types.Address{}, fmt.Errorf("invalid IP: %q", ipStr)
	}
	if ip4 := netIP.To4(); ip4 == nil {
		addr.Af = ip_types.ADDRESS_IP6
		var ip6addr ip_types.IP6Address
		copy(ip6addr[:], netIP.To16())
		addr.Un.SetIP6(ip6addr)
	} else {
		addr.Af = ip_types.ADDRESS_IP4
		var ip4addr ip_types.IP4Address
		copy(ip4addr[:], ip4)
		addr.Un.SetIP4(ip4addr)
	}
	return
}

func ipToAddress(address *net.IPNet, isIPv6 bool) (ipAddr interfaces.Address) {
	if isIPv6 {
		ipAddr.Af = ip_types.ADDRESS_IP6
		var ip6addr interfaces.IP6Address
		copy(ip6addr[:], address.IP.To16())
		ipAddr.Un.SetIP6(ip6addr)
	} else {
		ipAddr.Af = ip_types.ADDRESS_IP4
		var ip4addr interfaces.IP4Address
		copy(ip4addr[:], address.IP.To4())
		ipAddr.Un.SetIP4(ip4addr)
	}
	return
}

func IPtoPrefix(addr string) (ip_types.Prefix, error) {
	ipAddr, isIPv6, err := addrs.ParseIPWithPrefix(addr)
	if err != nil {
		return ip_types.Prefix{}, err
	}
	var prefix ip_types.Prefix
	maskSize, _ := ipAddr.Mask.Size()
	prefix.Len = byte(maskSize)
	if isIPv6 {
		prefix.Address.Af = ip_types.ADDRESS_IP6
		var ip6addr interfaces.IP6Address
		copy(ip6addr[:], ipAddr.IP.To16())
		prefix.Address.Un.SetIP6(ip6addr)
	} else {
		prefix.Address.Af = ip_types.ADDRESS_IP4
		var ip4addr interfaces.IP4Address
		copy(ip4addr[:], ipAddr.IP.To4())
		prefix.Address.Un.SetIP4(ip4addr)
	}
	return prefix, nil
}

func boolToUint(input bool) uint8 {
	if input {
		return 1
	}
	return 0
}

func uintToBool(value uint8) bool {
	return value != 0
}

func cleanString(s string) string {
	return strings.SplitN(s, "\x00", 2)[0]
}
