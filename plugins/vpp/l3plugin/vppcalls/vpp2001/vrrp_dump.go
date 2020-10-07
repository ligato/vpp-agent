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

package vpp2001

import (
	"fmt"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

// DumpVrrpEntries implements VRRP handler (not supported in VPP 20.01).
func (h *VrrpVppHandler) DumpVrrpEntries() (entries []*l3.VRRPEntry, err error) {
	err = fmt.Errorf("%w in VPP 20.01", vppcalls.ErrVRRPUnsupported)
	return
}
