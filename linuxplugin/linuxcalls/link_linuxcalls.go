// +build !windows,!darwin

package linuxcalls

import (
	"github.com/vishvananda/netlink"
)

// GetInterfaceType returns the type (string representation) of a given interface.
func GetInterfaceType(ifName string) (string, error) {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return "", err
	}
	return link.Type(), nil
}

// InterfaceExists checks if interface with a given name exists.
func InterfaceExists(ifName string) (bool, error) {
	_, err := netlink.LinkByName(ifName)
	if err == nil {
		return true, nil
	}
	if _, notFound := err.(netlink.LinkNotFoundError); notFound {
		return false, nil
	}
	return false, err
}
