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

package linuxcalls

import (
	"github.com/ligato/vpp-agent/plugins/linux/model/interfaces"
)

// LinuxInterfaceDetails is the wrapper structure for the linux interface northbound API structure.
type LinuxInterfaceDetails struct {
	Interface *interfaces.LinuxInterfaces_Interface `json:"linux_interface"`
	Meta      *LinuxInterfaceMeta                   `json:"linux_interface_meta"`
}

// LinuxInterfaceMeta is combination of proto-modelled Interface data and linux provided metadata
type LinuxInterfaceMeta struct {
	Index uint32 `json:"index"`
}

// DumpInterfaces is an implementation of linux interface handler
func (h *NetLinkHandler) DumpInterfaces() ([]*LinuxInterfaceDetails, error) {
	var ifs []*LinuxInterfaceDetails

	// todo implement

	return ifs, nil
}
