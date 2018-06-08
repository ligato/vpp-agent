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
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
)

// AddLoopbackInterface calls CreateLoopback bin API.
func AddLoopbackInterface(ifName string, vppChan VPPChannel, stopwatch *measure.Stopwatch) (swIndex uint32, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(interfaces.CreateLoopback{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &interfaces.CreateLoopback{}

	reply := &interfaces.CreateLoopbackReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	if reply.Retval != 0 {
		return 0, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return reply.SwIfIndex, SetInterfaceTag(ifName, reply.SwIfIndex, vppChan, stopwatch)
}

// DeleteLoopbackInterface calls DeleteLoopback bin API.
func DeleteLoopbackInterface(ifName string, idx uint32, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(interfaces.DeleteLoopback{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Prepare the message.
	req := &interfaces.DeleteLoopback{
		SwIfIndex: idx,
	}

	reply := &interfaces.DeleteLoopbackReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return RemoveInterfaceTag(ifName, idx, vppChan, stopwatch)
}
