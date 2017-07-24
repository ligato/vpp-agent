// +build !windows,!darwin

package linuxcalls

import (
	"net"

	"github.com/vishvananda/netlink"
)

// AddInterfaceIP calls AddrAdd Netlink API
func AddInterfaceIP(ifName string, addr *net.IPNet) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	address := &netlink.Addr{IPNet: addr}
	return netlink.AddrAdd(link, address)
}

// DelInterfaceIP calls AddrDel Netlink API
func DelInterfaceIP(ifName string, addr *net.IPNet) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	address := &netlink.Addr{IPNet: addr}
	return netlink.AddrDel(link, address)
}

// SetInterfaceMTU calls LinkSetMTU Netlink API
func SetInterfaceMTU(ifName string, mtu int) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	return netlink.LinkSetMTU(link, mtu)
}
