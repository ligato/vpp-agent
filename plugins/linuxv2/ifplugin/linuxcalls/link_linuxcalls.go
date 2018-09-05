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
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// GetLinkByName calls netlink API to get Link type from interface name
func (handler *NetLinkHandler) GetLinkByName(ifName string) (netlink.Link, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("get-link-from-interface").LogTimeEntry(time.Since(t))
	}(time.Now())

	return netlink.LinkByName(ifName)
}

// GetLinkList calls netlink API to get all Links in namespace
func (handler *NetLinkHandler) GetLinkList() ([]netlink.Link, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("get-link-list").LogTimeEntry(time.Since(t))
	}(time.Now())

	return netlink.LinkList()
}

// SetLinkNamespace puts link into a network namespace.
func (handler *NetLinkHandler) SetLinkNamespace(link netlink.Link, ns netns.NsHandle) (err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("set-link-namespace").LogTimeEntry(time.Since(t))
	}(time.Now())

	return netlink.LinkSetNsFd(link, int(ns))
}

// LinkSubscribe takes a channel to which notifications will be sent
// when links change. Close the 'done' chan to stop subscription.
func (handler *NetLinkHandler) LinkSubscribe(ch chan<- netlink.LinkUpdate, done <-chan struct{}) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("link-subscribe").LogTimeEntry(time.Since(t))
	}(time.Now())

	return netlink.LinkSubscribe(ch, done)
}

// GetInterfaceType returns the type (string representation) of a given interface.
func (handler *NetLinkHandler) GetInterfaceType(ifName string) (string, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("get-interface-type").LogTimeEntry(time.Since(t))
	}(time.Now())

	link, err := handler.GetLinkByName(ifName)
	if err != nil {
		return "", err
	}
	return link.Type(), nil
}

// InterfaceExists checks if interface with a given name exists.
func (handler *NetLinkHandler) InterfaceExists(ifName string) (bool, error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("interface-exists").LogTimeEntry(time.Since(t))
	}(time.Now())

	_, err := handler.GetLinkByName(ifName)
	if err == nil {
		return true, nil
	}
	if _, notFound := err.(netlink.LinkNotFoundError); notFound {
		return false, nil
	}
	return false, err
}

// DeleteInterface removes the given interface.
func (handler *NetLinkHandler) DeleteInterface(ifName string) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("delete-interface").LogTimeEntry(time.Since(t))
	}(time.Now())

	link, err := handler.GetLinkByName(ifName)
	if err != nil {
		return err
	}

	return netlink.LinkDel(link)
}

// RenameInterface changes the name of the interface <ifName> to <newName>.
func (handler *NetLinkHandler) RenameInterface(ifName string, newName string) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("rename-interface").LogTimeEntry(time.Since(t))
	}(time.Now())

	link, err := handler.GetLinkByName(ifName)
	if err != nil {
		return err
	}
	err = handler.SetInterfaceDown(ifName)
	if err != nil {
		return err
	}
	err = netlink.LinkSetName(link, newName)
	if err != nil {
		return err
	}
	err = handler.SetInterfaceUp(newName)
	if err != nil {
		return err
	}
	return nil
}

// SetInterfaceAlias sets the alias of the given interface.
// Equivalent to: `ip link set dev $ifName alias $alias`
func (handler *NetLinkHandler) SetInterfaceAlias(ifName, alias string) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog("set-link-alias").LogTimeEntry(time.Since(t))
	}(time.Now())

	link, err := handler.GetLinkByName(ifName)
	if err != nil {
		return err
	}

	return netlink.LinkSetAlias(link, alias)
}
