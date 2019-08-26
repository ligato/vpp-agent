package utils

import (
	"fmt"
	"net"
	"strings"

	"github.com/ligato/vpp-agent/api/models/netalloc"
)

// ParseIPAddr parses IP address from string.
func ParseIPAddr(addr string, expNet *net.IPNet) (ipNet *net.IPNet, err error) {
	if strings.Contains(addr, "/") {
		// IP with mask
		ip, ipNet, err := net.ParseCIDR(addr)
		if err != nil {
			return nil, err
		}
		ipNet.IP = ip
		return ipNet, nil
	}

	// IP without mask
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", addr)
	}
	if expNet != nil {
		expNet = &net.IPNet{IP: expNet.IP.Mask(expNet.Mask), Mask: expNet.Mask}
		if expNet.Contains(ip) {
			// IP address from the expected network
			return &net.IPNet{IP: ip.To4(), Mask: expNet.Mask}, nil
		}
	}

	// use all-ones mask
	defaultIpv4Mask := net.CIDRMask(32, 32)
	defaultIpv6Mask := net.CIDRMask(128, 128)

	if ip.To4() != nil {
		// IPv4 address
		return &net.IPNet{IP: ip.To4(), Mask: defaultIpv4Mask}, nil
	}

	// IPv6 address
	return &net.IPNet{IP: ip, Mask: defaultIpv6Mask}, nil
}

// ParseAddrAllocRef parses reference to allocated address.
func ParseAddrAllocRef(addrAllocRef, expIface string) (
	network, iface string, isRef bool, err error) {

	if !strings.HasPrefix(addrAllocRef, netalloc.AllocRefPrefix) {
		isRef = false
		return
	}

	isRef = true
	addrAllocRef = strings.TrimPrefix(addrAllocRef, netalloc.AllocRefPrefix)
	parts := strings.SplitN(addrAllocRef, "/", 2)

	// parse network name
	network = parts[0]
	if network == "" {
		err = fmt.Errorf("address allocation reference with empty network name: %s",
			addrAllocRef)
		return
	}

	if len(parts) == 2 {
		iface = parts[1]
		if expIface != "" && iface != expIface {
			err = fmt.Errorf("expected different interface name in the address allocation "+
				"reference: %s (expected=%s vs. actual=%s)", addrAllocRef, expIface, iface)
			return
		}
	} else {
		if expIface == "" {
			err = fmt.Errorf("missing interface name in the address allocation reference: %s",
				addrAllocRef)
			return
		} else {
			iface = expIface
		}
	}
	return
}

// GetIPAddrInGivenForm returns IP address in the requested form.
func GetIPAddrInGivenForm(addr *net.IPNet, form netalloc.IPAddressForm) *net.IPNet {
	switch form {
	case netalloc.IPAddressForm_UNDEFINED_FORM:
		return addr
	case netalloc.IPAddressForm_ADDR_ONLY:
		return &net.IPNet{IP: addr.IP}
	case netalloc.IPAddressForm_ADDR_WITH_MASK:
		return addr
	case netalloc.IPAddressForm_ADDR_NET:
		return &net.IPNet{IP: addr.IP.Mask(addr.Mask), Mask: addr.Mask}
	case netalloc.IPAddressForm_SINGLE_ADDR_NET:
		allOnesIpv4 := net.CIDRMask(32, 32)
		allOnesIpv6 := net.CIDRMask(128, 128)
		if addr.IP.To4() != nil {
			return &net.IPNet{IP: addr.IP, Mask: allOnesIpv4}
		}
		return &net.IPNet{IP: addr.IP, Mask: allOnesIpv6}
	}
	return addr
}
