// +build !windows,!darwin

package linuxcalls

import (
	"fmt"

	log "github.com/ligato/cn-infra/logging/logrus"

	"github.com/vishvananda/netlink"
)

// AddVethInterface calls LinkAdd Netlink API for the Netlink.Veth interface type.
func AddVethInterface(ifName, peerIfName string) error {
	log.WithFields(log.Fields{"ifName": ifName, "peerIfName": peerIfName}).Debug("Creating new Linux VETH pair")

	// Veth pair params
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:   ifName,
			TxQLen: 0,
		},
		PeerName: peerIfName,
	}

	// Create the veth pair
	if err := netlink.LinkAdd(veth); err != nil {
		return err
	}

	return nil
}

// DelVethInterface calls LinkDel Netlink API for the Netlink.Veth interface type.
func DelVethInterface(ifName, peerIfName string) error {
	log.WithFields(log.Fields{"ifName": ifName, "peerIfName": peerIfName}).Debug("Deleting Linux VETH pair")

	// Veth pair params
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:   ifName,
			TxQLen: 0,
		},
		PeerName: peerIfName,
	}

	// Create the veth pair
	if err := netlink.LinkDel(veth); err != nil {
		return err
	}

	return nil
}

// GetVethPeerName return the peer name for a given VETH interface.
func GetVethPeerName(ifName string) (string, error) {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return "", err
	}
	veth, isVeth := link.(*netlink.Veth)
	if !isVeth {
		return "", fmt.Errorf("Interface '%s' is not VETH", ifName)
	}
	return veth.PeerName, nil
}
