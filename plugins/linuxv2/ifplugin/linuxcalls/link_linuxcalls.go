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

// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package linuxcalls

import (
	"net"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// GetLinkByName calls netlink API to get Link type from interface name
func (h *NetLinkHandler) GetLinkByName(ifName string) (netlink.Link, error) {
	return netlink.LinkByName(ifName)
}

// GetLinkList calls netlink API to get all Links in namespace
func (h *NetLinkHandler) GetLinkList() ([]netlink.Link, error) {
	return netlink.LinkList()
}

// SetLinkNamespace puts link into a network namespace.
func (h *NetLinkHandler) SetLinkNamespace(link netlink.Link, ns netns.NsHandle) (err error) {
	return netlink.LinkSetNsFd(link, int(ns))
}

// LinkSubscribe takes a channel to which notifications will be sent
// when links change. Close the 'done' chan to stop subscription.
func (h *NetLinkHandler) LinkSubscribe(ch chan<- netlink.LinkUpdate, done <-chan struct{}) error {
	return netlink.LinkSubscribe(ch, done)
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

// IsInterfaceEnabled checks if the interface is UP.
func (h *NetLinkHandler) IsInterfaceEnabled(ifName string) (bool, error) {
	intf, err := net.InterfaceByName(ifName)
	if err != nil {
		return false, err
	}
	return (intf.Flags & net.FlagUp) != 0, nil
}

// DeleteInterface removes the given interface.
func (h *NetLinkHandler) DeleteInterface(ifName string) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}

	return netlink.LinkDel(link)
}

// RenameInterface changes the name of the interface <ifName> to <newName>.
func (h *NetLinkHandler) RenameInterface(ifName string, newName string) error {
	link, err := h.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	err = h.SetInterfaceDown(ifName)
	if err != nil {
		return err
	}
	err = netlink.LinkSetName(link, newName)
	if err != nil {
		return err
	}
	err = h.SetInterfaceUp(newName)
	if err != nil {
		return err
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

	return netlink.LinkSetAlias(link, alias)
}
