// Copyright (c) 2018 Cisco and/or its affiliates.
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

package nsplugin

import (
	"os"
	"syscall"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// Defines all methods required for managing operating system, system calls and namespaces on system level
type SystemAPI interface {
	OperatingSystem
	Syscall
	NetNsNamespace
	NetlinkNamespace
}

// Operating system defines all methods calling os package
type OperatingSystem interface {
	// Open file
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	// MkDirAll creates a directory with all parent directories
	MkDirAll(path string, perm os.FileMode) error
	// Remove removes named file or directory
	Remove(name string) error
}

// Syscall defines methods using low-level operating system primitives
type Syscall interface {
	// Mount makes resources available
	Mount(source string, target string, fsType string, flags uintptr, data string) error
	// Unmount resources
	Unmount(target string, flags int) (err error)
}

// NetNsNamespace defines method for namespace handling from netns package
type NetNsNamespace interface {
	// NewNetworkNamespace crates new namespace and returns handle to manage it further
	NewNetworkNamespace() (ns netns.NsHandle, err error)
	// GetNamespaceFromName returns namespace handle from its name
	GetNamespaceFromName(name string) (ns netns.NsHandle, err error)
	// SetNamespace sets the current namespace to the namespace represented by the handle
	SetNamespace(ns netns.NsHandle) (err error)
}

// NetlinkNamespace defines method for namespace handling from netlink package
type NetlinkNamespace interface {
	// LinkSetNsFd puts the device into a new network namespace.
	LinkSetNsFd(link netlink.Link, fd int) (err error)
}

type systemHandler struct{}

func NewSystemHandler() *systemHandler {
	return &systemHandler{}
}

/* Operating system */

func (osh *systemHandler) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (osh *systemHandler) MkDirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (osh *systemHandler) Remove(name string) error {
	return os.Remove(name)
}

/* Syscall */

func (osh *systemHandler) Mount(source string, target string, fsType string, flags uintptr, data string) error {
	return syscall.Mount(source, target, fsType, flags, data)
}

func (osh *systemHandler) Unmount(target string, flags int) error {
	return syscall.Unmount(target, flags)
}

/* Netns namespace */

func (osh *systemHandler) NewNetworkNamespace() (ns netns.NsHandle, err error) {
	return netns.New()
}

func (osh *systemHandler) GetNamespaceFromName(name string) (ns netns.NsHandle, err error) {
	return netns.GetFromName(name)
}

func (osh *systemHandler) SetNamespace(ns netns.NsHandle) (err error) {
	return netns.Set(ns)
}

/* Netlink namespace */

func (osh *systemHandler) LinkSetNsFd(link netlink.Link, fd int) (err error) {
	return netlink.LinkSetNsFd(link, fd)
}
