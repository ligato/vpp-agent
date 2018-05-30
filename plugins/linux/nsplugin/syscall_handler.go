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
	"github.com/vishvananda/netns"
	"os"
	"syscall"
)

type SyscallAPI interface {
	MakeDirectoryAll(path string, perm os.FileMode) error
	Mount(source string, target string, fsType string, flags uintptr, data string) error
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	NewNetworkNamespace() (ns netns.NsHandle, err error)
	GetNsHandleFromName(name string) (ns netns.NsHandle, err error)
	SetNamespace(ns netns.NsHandle) (err error)
}

type syscallHandler struct {
}

func NewSyscallHandler() *syscallHandler {
	return &syscallHandler{}
}

func (osh *syscallHandler) MakeDirectoryAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (osh *syscallHandler) Mount(source string, target string, fsType string, flags uintptr, data string) error {
	return syscall.Mount(source, target, fsType, flags, data)
}

func (osh *syscallHandler) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (osh *syscallHandler) NewNetworkNamespace() (ns netns.NsHandle, err error) {
	return netns.New()
}

func (osh *syscallHandler) GetNsHandleFromName(name string) (ns netns.NsHandle, err error) {
	return netns.GetFromName(name)
}

func (osh *syscallHandler) SetNamespace(ns netns.NsHandle) (err error) {
	return netns.Set(ns)
}
