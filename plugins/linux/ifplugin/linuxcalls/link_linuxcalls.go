//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// +build !windows,!darwin

package linuxcalls

import (
	"net"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// GetLinkByName calls netlink API to get Link type from interface name
func (h *NetLinkHandler) GetLinkByName(ifName string) (netlink.Link, error) {
	link, err := h.LinkByName(ifName)
	if err != nil {
		return nil, errors.Wrapf(err, "LinkByName %q", ifName)
	}
	return link, nil
}

// GetLinkByIndex calls netlink API to get Link type from interface index
func (h *NetLinkHandler) GetLinkByIndex(ifIdx int) (netlink.Link, error) {
	link, err := h.LinkByIndex(ifIdx)
	if err != nil {
		return nil, errors.Wrapf(err, "LinkByIndex %q", ifIdx)
	}
	return link, nil
}

// GetLinkList calls netlink API to get all Links in namespace
func (h *NetLinkHandler) GetLinkList() ([]netlink.Link, error) {
	return h.LinkList()
}

// SetLinkNamespace puts link into a network namespace.
func (h *NetLinkHandler) SetLinkNamespace(link netlink.Link, ns netns.NsHandle) (err error) {
	if err := h.LinkSetNsFd(link, int(ns)); err != nil {
		return errors.Wrapf(err, "LinkSetNsFd %v", ns)
	}
	return nil
}

// LinkSubscribe takes a channel to which notifications will be sent
// when links change. Close the 'done' chan to stop subscription.
func (h *NetLinkHandler) LinkSubscribe(ch chan<- netlink.LinkUpdate, done <-chan struct{}) error {
	return netlink.LinkSubscribeAt(h.nsHandle, ch, done)
}

// AddrSubscribe takes a channel to which notifications will be sent
// when addresses change. Close the 'done' chan to stop subscription.
func (h *NetLinkHandler) AddrSubscribe(ch chan<- netlink.AddrUpdate, done <-chan struct{}) error {
	return netlink.AddrSubscribeAt(h.nsHandle, ch, done)
}

// GetInterfaceType returns the type (string representation) of a given interface.
func (h *NetLinkHandler) GetInterfaceType(ifName string) (string, error) {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return "", err
	}
	return link.Type(), nil
}

// InterfaceExists checks if interface with a given name exists.
func (h *NetLinkHandler) InterfaceExists(ifName string) (bool, error) {
	_, err := h.GetLinkByName(ifName)
	if err == nil {
		return true, nil
	}
	if _, notFound := err.(netlink.LinkNotFoundError); notFound {
		return false, nil
	}
	return false, err
}

// IsInterfaceUp checks if the interface is UP.
func (h *NetLinkHandler) IsInterfaceUp(ifName string) (bool, error) {
	link, err := h.LinkByName(ifName)
	if err != nil {
		return false, errors.Wrapf(err, "LinkByName %s", ifName)
	}
	return isLinkUp(link), nil
}

// DeleteInterface removes the given interface.
func (h *NetLinkHandler) DeleteInterface(ifName string) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	if err := h.LinkDel(link); err != nil {
		return errors.Wrapf(err, "LinkDel %s", link)
	}
	return nil
}

// RenameInterface changes the name of the interface <ifName> to <newName>.
func (h *NetLinkHandler) RenameInterface(ifName string, newName string) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	var wasUp bool
	if isLinkUp(link) {
		wasUp = true
		if err = h.LinkSetDown(link); err != nil {
			return errors.Wrapf(err, "LinkSetDown %v", link)
		}
	}
	if err = h.LinkSetName(link, newName); err != nil {
		return errors.Wrapf(err, "LinkSetName %s", newName)
	}
	if wasUp {
		if err = h.LinkSetUp(link); err != nil {
			return errors.Wrapf(err, "LinkSetUp %v", link)
		}
	}
	return nil
}

// SetInterfaceAlias sets the alias of the given interface.
// Equivalent to: `ip link set dev $ifName alias $alias`
func (h *NetLinkHandler) SetInterfaceAlias(ifName, alias string) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	if err := h.LinkSetAlias(link, alias); err != nil {
		return errors.Wrapf(err, "LinkSetAlias %s", alias)
	}
	return nil
}

// SetInterfaceDown calls Netlink API LinkSetDown.
func (h *NetLinkHandler) SetInterfaceDown(ifName string) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	if err := h.LinkSetDown(link); err != nil {
		return errors.Wrapf(err, "LinkSetDown %v", link)
	}
	return nil
}

// SetInterfaceUp calls Netlink API LinkSetUp.
func (h *NetLinkHandler) SetInterfaceUp(ifName string) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	if err := h.LinkSetUp(link); err != nil {
		return errors.Wrapf(err, "LinkSetUp %v", link)
	}
	return nil
}

// GetAddressList calls AddrList netlink API
func (h *NetLinkHandler) GetAddressList(ifName string) ([]netlink.Addr, error) {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return nil, err
	}
	return h.AddrList(link, netlink.FAMILY_ALL)
}

// AddInterfaceIP calls AddrAdd Netlink API.
func (h *NetLinkHandler) AddInterfaceIP(ifName string, ip *net.IPNet) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{IPNet: ip}
	if err := h.AddrAdd(link, addr); err != nil {
		return errors.Wrapf(err, "AddrAdd %v", addr)
	}
	return nil
}

// DelInterfaceIP calls AddrDel Netlink API.
func (h *NetLinkHandler) DelInterfaceIP(ifName string, ip *net.IPNet) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{IPNet: ip}
	if err := h.AddrDel(link, addr); err != nil {
		return errors.Wrapf(err, "AddrDel %v", addr)
	}
	return nil
}

// SetInterfaceMTU calls LinkSetMTU Netlink API.
func (h *NetLinkHandler) SetInterfaceMTU(ifName string, mtu int) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	if err := h.LinkSetMTU(link, mtu); err != nil {
		return errors.Wrapf(err, "LinkSetMTU %v", mtu)
	}
	return nil
}

// SetInterfaceMac calls LinkSetHardwareAddr netlink API.
func (h *NetLinkHandler) SetInterfaceMac(ifName string, macAddress string) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	hwAddr, err := net.ParseMAC(macAddress)
	if err != nil {
		return err
	}
	if err := h.LinkSetHardwareAddr(link, hwAddr); err != nil {
		return errors.Wrapf(err, "LinkSetHardwareAddr %v", hwAddr)
	}
	return nil
}

// AddVethInterfacePair calls LinkAdd Netlink API for the Netlink.Veth interface type.
func (h *NetLinkHandler) AddVethInterfacePair(ifName, peerIfName string) error {
	link := &netlink.Veth{
		LinkAttrs: newLinkAttrs(ifName),
		PeerName:  peerIfName,
	}
	if err := h.LinkAdd(link); err != nil {
		return errors.Wrapf(err, "LinkAdd %v", link)
	}
	return nil
}

// AddDummyInterface configures dummy interface (effectively additional loopback).
func (h *NetLinkHandler) AddDummyInterface(ifName string) error {
	link := &netlink.Dummy{
		LinkAttrs: newLinkAttrs(ifName),
	}
	if err := h.LinkAdd(link); err != nil {
		return errors.Wrapf(err, "LinkAdd (dummy-ifName=%s)", ifName)
	}
	return nil
}

// AddVRFDevice configures new VRF network device.
func (h *NetLinkHandler) AddVRFDevice(vrfDevName string, routingTable uint32) error {
	link := &netlink.Vrf{
		LinkAttrs: newLinkAttrs(vrfDevName),
		Table:     routingTable,
	}
	if err := h.LinkAdd(link); err != nil {
		return errors.Wrapf(err, "LinkAdd (vrf=%s, rt=%d)",
			vrfDevName, routingTable)
	}
	return nil
}

// PutInterfaceIntoVRF assigns Linux interface into a given VRF.
func (h *NetLinkHandler) PutInterfaceIntoVRF(ifName, vrfDevName string) error {
	ifLink, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	vrfLink, err := h.GetLinkByName(vrfDevName)
	if err != nil {
		return err
	}
	if err := h.LinkSetMasterByIndex(ifLink, vrfLink.Attrs().Index); err != nil {
		return errors.Wrapf(err, "LinkSetMasterByIndex (interface=%s, vrf=%s, vrf-index=%d)",
			ifName, vrfDevName, vrfLink.Attrs().Index)
	}
	return nil
}

// RemoveInterfaceFromVRF un-assigns Linux interface from a given VRF.
func (h *NetLinkHandler) RemoveInterfaceFromVRF(ifName, vrfDevName string) error {
	ifLink, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	if err := h.LinkSetNoMaster(ifLink); err != nil {
		return errors.Wrapf(err, "LinkSetNoMaster (interface=%s, vrf=%s)",
			ifName, vrfDevName)
	}
	return nil
}

func isLinkUp(link netlink.Link) bool {
	return (link.Attrs().Flags & net.FlagUp) == net.FlagUp
}

func newLinkAttrs(name string) netlink.LinkAttrs {
	attrs := netlink.NewLinkAttrs()
	attrs.Name = name
	return attrs
}
