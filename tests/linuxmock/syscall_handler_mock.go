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

package linuxmock

import (
	"github.com/vishvananda/netns"
	"os"
)

// todo

type syscallHandlerMock struct {
}

func NewMockOsHandler() *syscallHandlerMock {
	return &syscallHandlerMock{}
}

func (mock *syscallHandlerMock) MakeDirectoryAll(path string, perm os.FileMode) error {
	return nil
}

func (mock *syscallHandlerMock) Mount(source string, target string, fsType string, flags uintptr, data string) error {
	return nil
}

func (mock *syscallHandlerMock) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return &os.File{}, nil // todo better mock?
}

func (mock *syscallHandlerMock) NewNetworkNamespace() (ns netns.NsHandle, err error) {
	return 1, nil
}

func (mock *syscallHandlerMock) GetNsHandleFromName(name string) (ns netns.NsHandle, err error) {
	return 1, nil
}

func (mock *syscallHandlerMock) SetNamespace(ns netns.NsHandle) (err error) {
	return nil
}
