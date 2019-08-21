// Copyright (c) 2019 Cisco and/or its affiliates.
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

package netalloc

import (
	"net"

	"github.com/ligato/vpp-agent/pkg/models"
)

// ModuleName is the module name used for models of the netalloc plugin.
const ModuleName = "netalloc"

var (
	ModelAddressAllocation = models.Register(&AddressAllocation{}, models.Spec{
		Module:  ModuleName,
		Version: "v1",
		Type:    "address",
	})
)

// AddrAllocMetadata stores allocated address already parsed from string.
type AddrAllocMetadata struct {
	IPAddr *net.IPNet
	HwAddr net.HardwareAddr
}