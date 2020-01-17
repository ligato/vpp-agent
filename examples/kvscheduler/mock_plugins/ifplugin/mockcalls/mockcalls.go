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

package mockcalls

import (
	"fmt"

	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/model"
)

// SimulateFailedTapCreation allows to simulate failure of the next Create
// operation for a TAP interface.
var SimulateFailedTapCreation bool

// CreateLoopbackInterface creates loopback in the mock SB.
func (h *MockIfaceHandler) CreateLoopbackInterface(ifaceName string) (sbIfaceHandle uint32, err error) {
	sbIfaceHandle = h.nextIfaceHandle
	h.nextIfaceHandle++
	h.mockIfaces[sbIfaceHandle] = &mock_interfaces.Interface{
		Name: ifaceName,
		Type: mock_interfaces.Interface_LOOPBACK,
	}
	h.log.Debugf("Created Loopback interface: %s", ifaceName)
	return sbIfaceHandle, nil
}

// DeleteLoopbackInterface deletes loopback in the mock SB.
func (h *MockIfaceHandler) DeleteLoopbackInterface(sbIfaceHandle uint32) error {
	iface, err := h.getInterface(sbIfaceHandle)
	if err != nil {
		return err
	}
	delete(h.mockIfaces, sbIfaceHandle)
	h.log.Debugf("Deleted Loopback interface: %s", iface.Name)
	return nil
}

// CreateTapInterface creates TAP interface in the mock SB.
func (h *MockIfaceHandler) CreateTapInterface(ifaceName string) (sbIfaceHandle uint32, err error) {
	if SimulateFailedTapCreation {
		SimulateFailedTapCreation = false // next attempt will succeed
		return 0, fmt.Errorf("mock error")
	}
	sbIfaceHandle = h.nextIfaceHandle
	h.nextIfaceHandle++
	h.mockIfaces[sbIfaceHandle] = &mock_interfaces.Interface{
		Name: ifaceName,
		Type: mock_interfaces.Interface_TAP,
	}
	h.log.Debugf("Created TAP interface: %s", ifaceName)
	return sbIfaceHandle, nil
}

// CreateTapInterface deletes TAP interface in the mock SB.
func (h *MockIfaceHandler) DeleteTapInterface(sbIfaceHandle uint32) error {
	iface, err := h.getInterface(sbIfaceHandle)
	if err != nil {
		return err
	}
	delete(h.mockIfaces, sbIfaceHandle)
	h.log.Debugf("Deleted TAP interface: %s", iface.Name)
	return nil
}

// InterfaceAdminDown puts the given interface DOWN.
func (h *MockIfaceHandler) InterfaceAdminDown(sbIfaceHandle uint32) error {
	iface, err := h.getInterface(sbIfaceHandle)
	if err != nil {
		return err
	}
	iface.Enabled = false
	h.log.Debugf("Set interface '%s' DOWN", iface.Name)
	return nil
}

// InterfaceAdminUp puts the given interface UP.
func (h *MockIfaceHandler) InterfaceAdminUp(sbIfaceHandle uint32) error {
	iface, err := h.getInterface(sbIfaceHandle)
	if err != nil {
		return err
	}
	iface.Enabled = true
	h.log.Debugf("Set interface '%s' UP", iface.Name)
	return nil
}

// SetInterfaceMac changes MAC address of the given interface.
func (h *MockIfaceHandler) SetInterfaceMac(sbIfaceHandle uint32, macAddress string) error {
	iface, err := h.getInterface(sbIfaceHandle)
	if err != nil {
		return err
	}
	iface.PhysAddress = macAddress
	h.log.Debugf("Set interface '%s' MAC address: %s", iface.Name, macAddress)
	return nil
}

// DumpInterfaces returns interfaces "configured" in the mock SB.
func (h *MockIfaceHandler) DumpInterfaces() (mockIfaces, error) {
	h.log.Debugf("Dumped mock interfaces: %+v", h.mockIfaces)
	return h.mockIfaces, nil
}

// getInterface returns configuration of interface represented in the mock SB
// with the given integer handle.
func (h *MockIfaceHandler) getInterface(sbIfaceHandle uint32) (*mock_interfaces.Interface, error) {
	iface, exists := h.mockIfaces[sbIfaceHandle]
	if !exists {
		return nil, fmt.Errorf("cannot find interface with index: %d", sbIfaceHandle)
	}
	return iface, nil
}
