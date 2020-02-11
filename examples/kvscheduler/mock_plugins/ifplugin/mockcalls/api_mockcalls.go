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
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/model"
)

// map of interfaces "configured" in the mock SB.
type mockIfaces map[uint32]*mock_interfaces.Interface

// MockIfaceAPI provides methods for creating and managing interfaces
// in the mock SB
type MockIfaceAPI interface {
	MockIfaceWrite
	MockIfaceRead
}

// MockIfaceWrite provides write methods for interface plugin
type MockIfaceWrite interface {
	// CreateLoopbackInterface creates loopback in the mock SB.
	CreateLoopbackInterface(ifaceName string) (sbIfaceHandle uint32, err error)
	// DeleteLoopbackInterface deletes loopback in the mock SB.
	DeleteLoopbackInterface(sbIfaceHandle uint32) error

	// CreateTapInterface creates TAP interface in the mock SB.
	CreateTapInterface(ifaceName string) (sbIfaceHandle uint32, err error)
	// CreateTapInterface deletes TAP interface in the mock SB.
	DeleteTapInterface(sbIfaceHandle uint32) error

	// InterfaceAdminDown puts the given mock interface DOWN.
	InterfaceAdminDown(sbIfaceHandle uint32) error
	// InterfaceAdminUp puts the given interface UP.
	InterfaceAdminUp(sbIfaceHandle uint32) error

	// SetInterfaceMac changes MAC address of the given interface.
	SetInterfaceMac(sbIfaceHandle uint32, macAddress string) error
}

// MockIfaceRead provides read methods for interface plugin
type MockIfaceRead interface {
	// DumpInterfaces returns interfaces configured in the mock SB.
	DumpInterfaces() (mockIfaces, error)
}

// MockIfaceHandler is accessor for calls into mock SB.
type MockIfaceHandler struct {
	log logging.Logger

	// mock SB
	nextIfaceHandle uint32
	mockIfaces      mockIfaces
}

// NewMockIfHandler creates new instance of interface handler for mock SB.
func NewMockIfaceHandler(log logging.Logger) MockIfaceAPI {
	return &MockIfaceHandler{
		log:        log,
		mockIfaces: make(mockIfaces),
	}
}
