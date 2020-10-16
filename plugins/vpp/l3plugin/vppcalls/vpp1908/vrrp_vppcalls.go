//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package vpp1908

import (
	"fmt"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

// VppAddVrrp implements VRRP handler (not supported in VPP 19.08).
func (h *VrrpVppHandler) VppAddVrrp(entry *l3.VRRPEntry) error {
	return fmt.Errorf("%w in VPP 19.08", vppcalls.ErrVRRPUnsupported)
}

// VppDelVrrp implements VRRP handler (not supported in VPP 19.08).
func (h *VrrpVppHandler) VppDelVrrp(entry *l3.VRRPEntry) error {
	return fmt.Errorf("%w in VPP 19.08", vppcalls.ErrVRRPUnsupported)
}

// VppStartVrrp implements VRRP handler.
func (h *VrrpVppHandler) VppStartVrrp(entry *l3.VRRPEntry) error {
	return fmt.Errorf("%w in VPP 19.08", vppcalls.ErrVRRPUnsupported)
}

// VppStopVrrp implements VRRP handler.
func (h *VrrpVppHandler) VppStopVrrp(entry *l3.VRRPEntry) error {
	return fmt.Errorf("%w in VPP 19.08", vppcalls.ErrVRRPUnsupported)
}
