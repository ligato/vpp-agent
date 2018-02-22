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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
)

// AddLoopbackInterface calls CreateLoopback bin API.
func AddLoopbackInterface(ifName string, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (swIndex uint32, err error) {
	// CreateLoopback time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &interfaces.CreateLoopback{}

	reply := &interfaces.CreateLoopbackReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	if reply.Retval != 0 {
		return 0, fmt.Errorf("add loopback interface returned %d", reply.Retval)
	}

	return reply.SwIfIndex, SetInterfaceTag(ifName, reply.SwIfIndex, vppChan, timeLog)
}

// DeleteLoopbackInterface calls DeleteLoopback bin API.
func DeleteLoopbackInterface(ifName string, idx uint32, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// DeleteLoopback time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message.
	req := &interfaces.DeleteLoopback{}
	req.SwIfIndex = idx

	reply := &interfaces.DeleteLoopbackReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("deleting of loopback interface returned %d", reply.Retval)
	}

	return RemoveInterfaceTag(ifName, idx, vppChan, timeLog)
}
