// +build !windows,!darwin

package linuxcalls

import (
	"net"

	"github.com/vishvananda/netlink"
)

// SetInterfaceMac calls LinkSetHardwareAddr netlink API
func SetInterfaceMac(ifName string, macAddress string) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	hwAddr, err := net.ParseMAC(macAddress)
	if err != nil {
		return err
	}
	return netlink.LinkSetHardwareAddr(link, hwAddr)
}
