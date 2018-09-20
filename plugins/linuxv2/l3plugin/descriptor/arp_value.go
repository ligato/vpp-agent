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

package descriptor

import (
	"strings"

	scheduler "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/kvscheduler/value/protoval"

	"github.com/ligato/vpp-agent/plugins/linuxv2/model/l3"
)

// ARPProtoValue overrides the default implementation of the Equivalent method.
type ARPProtoValue struct {
	protoval.ProtoValue
	arp *l3.LinuxStaticARPEntry
}

// Equivalent is case-insensitive comparison function for l3.LinuxStaticARPEntry.
func (apv *ARPProtoValue) Equivalent(v2 scheduler.Value) bool {
	apv2, ok := v2.(*ARPProtoValue)
	if !ok {
		return false
	}
	arp1 := apv.arp
	arp2 := apv2.arp

	// interfaces compared as usually:
	if arp1.Interface != arp2.Interface {
		return false
	}

	// compare MAC addresses case-insensitively
	if strings.ToLower(arp1.HwAddress) != strings.ToLower(arp2.HwAddress) {
		return false
	}

	// compare IP addresses converted to net.IPNet
	return equalAddrs(arp1.IpAddr, arp2.IpAddr)
}
