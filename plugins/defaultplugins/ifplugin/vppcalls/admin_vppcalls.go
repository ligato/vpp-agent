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

package vppcalls

import (
	"fmt"
	"time"

	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
)

func interfaceSetFlags(ifIdx uint32, adminUp bool, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(interfaces.SwInterfaceSetFlagsReply{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &interfaces.SwInterfaceSetFlags{
		SwIfIndex: ifIdx,
	}
	if adminUp {
		req.AdminUpDown = 1
	} else {
		req.AdminUpDown = 0
	}

	reply := &interfaces.SwInterfaceSetFlagsReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// InterfaceAdminDown calls binary API SwInterfaceSetFlagsReply with AdminUpDown=0.
func InterfaceAdminDown(ifIdx uint32, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	return interfaceSetFlags(ifIdx, false, vppChan, stopwatch)
}

// InterfaceAdminUp calls binary API SwInterfaceSetFlagsReply with AdminUpDown=1.
func InterfaceAdminUp(ifIdx uint32, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	return interfaceSetFlags(ifIdx, true, vppChan, stopwatch)
}

func handleInterfaceTag(tag string, ifIdx uint32, isAdd bool, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(interfaces.SwInterfaceTagAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &interfaces.SwInterfaceTagAddDel{
		Tag:   []byte(tag),
		IsAdd: boolToUint(isAdd),
	}
	// For some reason, if deleting tag, the software interface index has to be 0 and only name should be set.
	// Otherwise reply returns with error core -2 (incorrect sw_if_idx)
	if isAdd {
		req.SwIfIndex = ifIdx
	}

	reply := &interfaces.SwInterfaceTagAddDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s %v (index %v) add/del returned %v", reply.GetMessageName(), tag, ifIdx, reply.Retval)
	}

	return nil
}

// SetInterfaceTag registers new interface index/tag pair
func SetInterfaceTag(tag string, ifIdx uint32, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	return handleInterfaceTag(tag, ifIdx, true, vppChan, stopwatch)
}

// RemoveInterfaceTag un-registers new interface index/tag pair
func RemoveInterfaceTag(tag string, ifIdx uint32, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	return handleInterfaceTag(tag, ifIdx, false, vppChan, stopwatch)
}
