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
	"reflect"
	"strings"

	"github.com/gogo/protobuf/proto"

	scheduler "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/kvscheduler/value/protoval"
	"github.com/ligato/cn-infra/utils/addrs"

	"github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"
)

// InterfaceProtoValue overrides the default implementation of the Equivalent method.
type InterfaceProtoValue struct {
	protoval.ProtoValue
	linuxIntf *interfaces.LinuxInterface
}

// Equivalent is case-insensitive comparison function for interfaces.LinuxInterface,
// also ignoring the order of assigned IP addresses.
func (ipv *InterfaceProtoValue) Equivalent(v2 scheduler.Value) bool {
	ipv2, ok := v2.(*InterfaceProtoValue)
	if !ok {
		return false
	}
	intf1 := ipv.linuxIntf
	intf2 := ipv2.linuxIntf

	// attributes compared as usually:
	if intf1.Name != intf2.Name || intf1.Type != intf2.Type || intf1.Enabled != intf2.Enabled ||
		getHostIfName(intf1) != getHostIfName(intf2) {
		return false
	}
	if !proto.Equal(intf1.Namespace, intf2.Namespace) || !proto.Equal(intf1.Veth, intf2.Veth) ||
		!proto.Equal(intf1.Tap, intf2.Tap) {
		return false
	}

	// handle default MTU
	if getInterfaceMTU(intf1) != getInterfaceMTU(intf2) {
		return false
	}

	// compare MAC addresses case-insensitively
	if strings.ToLower(intf1.PhysAddress) != strings.ToLower(intf2.PhysAddress) {
		return false
	}

	// order-irrelevant comparison of IP addresses
	intf1Addrs, err1 := addrs.StrAddrsToStruct(intf1.IpAddresses)
	intf2Addrs, err2 := addrs.StrAddrsToStruct(intf2.IpAddresses)
	if err1 != nil || err2 != nil {
		// one or both of the configurations are invalid, compare lazily
		return reflect.DeepEqual(intf1.IpAddresses, intf2.IpAddresses)
	}
	obsolete, new := addrs.DiffAddr(intf1Addrs, intf2Addrs)
	return len(obsolete) == 0 && len(new) == 0
}
