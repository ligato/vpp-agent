package linuxplugin

import (
	"net"
)

// GetLinuxInterfaceIndex returns the index of a Linux interface identified by its name.
// In Linux, interface index is a positive integer that starts at one, zero is never used.
// Function returns negative number in case of a failure, such as when the interface doesn't exist.
// TODO: move to the package with network utilities
func GetLinuxInterfaceIndex(ifName string) int {
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return -1
	}
	return iface.Index
}
