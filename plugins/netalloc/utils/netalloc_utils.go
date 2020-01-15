package utils

import (
	"fmt"
	"net"
	"strings"

	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
)

// ParseIPAddr parses IP address from string.
func ParseIPAddr(addr string, expNet *net.IPNet) (ipNet *net.IPNet, fromExpNet bool, err error) {
	if expNet != nil {
		expNet = &net.IPNet{IP: expNet.IP.Mask(expNet.Mask), Mask: expNet.Mask}
	}

	if strings.Contains(addr, "/") {
		// IP with mask
		ip, ipNet, err := net.ParseCIDR(addr)
		if err != nil {
			return nil, false, err
		}
		if ip.To4() != nil {
			ip = ip.To4()
		}
		ipNet.IP = ip
		if expNet != nil {
			fromExpNet = expNet.Contains(ip)
		}
		return ipNet, fromExpNet, nil
	}

	// IP without mask
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, false, fmt.Errorf("invalid IP address: %s", addr)
	}
	if ip.To4() != nil {
		ip = ip.To4()
	}
	if expNet != nil {
		if expNet.Contains(ip) {
			// IP address from the expected network
			return &net.IPNet{IP: ip, Mask: expNet.Mask}, true, nil
		}
	}

	// use all-ones mask
	defaultIpv4Mask := net.CIDRMask(32, 32)
	defaultIpv6Mask := net.CIDRMask(128, 128)

	if ip.To4() != nil {
		// IPv4 address
		return &net.IPNet{IP: ip.To4(), Mask: defaultIpv4Mask}, false, nil
	}

	// IPv6 address
	return &net.IPNet{IP: ip, Mask: defaultIpv6Mask}, false, nil
}

// ParseAddrAllocRef parses reference to allocated address.
func ParseAddrAllocRef(addrAllocRef, expIface string) (
	network, iface string, isGW, isRef bool, err error) {

	if !strings.HasPrefix(addrAllocRef, netalloc.AllocRefPrefix) {
		isRef = false
		return
	}

	isRef = true
	addrAllocRef = strings.TrimPrefix(addrAllocRef, netalloc.AllocRefPrefix)
	if strings.HasSuffix(addrAllocRef, netalloc.AllocRefGWSuffix) {
		addrAllocRef = strings.TrimSuffix(addrAllocRef, netalloc.AllocRefGWSuffix)
		isGW = true
	}

	// parse network name
	parts := strings.SplitN(addrAllocRef, "/", 2)
	network = parts[0]
	if network == "" {
		err = fmt.Errorf("address allocation reference with empty network name: %s",
			addrAllocRef)
		return
	}

	// parse interface name
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
		zeroMaskIpv4 := net.CIDRMask(0, 32)
		zeroMaskIpv6 := net.CIDRMask(0, 128)
		if addr.IP.To4() != nil {
			return &net.IPNet{IP: addr.IP, Mask: zeroMaskIpv4}
		}
		return &net.IPNet{IP: addr.IP, Mask: zeroMaskIpv6}
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
