// +build !windows,!darwin

package linuxcalls

import (
	"github.com/vishvananda/netlink"
)

// InterfaceAdminDown calls Netlink API LinkSetDown
func InterfaceAdminDown(ifName string) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	return netlink.LinkSetDown(link)
}

// InterfaceAdminUp calls Netlink API LinkSetUp
func InterfaceAdminUp(ifName string) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	return netlink.LinkSetUp(link)
}
