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
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/interfaces"
)

// InterfaceAdminDown calls binary API SwInterfaceSetFlagsReply with AdminUpDown=0.
func InterfaceAdminDown(ifIdx uint32, vppChan VPPChannel, timeLog measure.StopWatchEntry) error {
	// SwInterfaceSetFlags time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	return interfaceSetFlags(ifIdx, false, vppChan)
}

// InterfaceAdminUp calls binary API SwInterfaceSetFlagsReply with AdminUpDown=1.
func InterfaceAdminUp(ifIdx uint32, vppChan VPPChannel, timeLog measure.StopWatchEntry) error {
	// SwInterfaceSetFlags time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	return interfaceSetFlags(ifIdx, true, vppChan)
}

func interfaceSetFlags(ifIdx uint32, adminUp bool, vppChan VPPChannel) error {
	// Prepare the message.
	req := &interfaces.SwInterfaceSetFlags{
		SwIfIndex: ifIdx,
	}
	if adminUp {
		req.AdminUpDown = 1
	} else {
		req.AdminUpDown = 0
	}

	reply := &interfaces.SwInterfaceSetFlagsReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("setting of interface flags (%+v) returned %d", req, reply.Retval)
	}

	return nil
}
