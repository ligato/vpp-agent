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
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/vpe"
	"time"
)

// AddLoopbackInterface calls CreateLoopback bin API
func AddLoopbackInterface(vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (swIndex uint32, err error) {
	// CreateLoopback time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &vpe.CreateLoopback{}
	reply := &vpe.CreateLoopbackReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		return 0, err
	}

	if 0 != reply.Retval {
		return 0, fmt.Errorf("add loopback interface returned %d", reply.Retval)
	}

	return reply.SwIfIndex, nil
}

// DeleteLoopbackInterface calls DeleteLoopback bin API
func DeleteLoopbackInterface(idx uint32, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// DeleteLoopback time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// prepare the message
	req := &vpe.DeleteLoopback{}
	req.SwIfIndex = idx

	reply := &vpe.DeleteLoopbackReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("deleting of loopback interface returned %d", reply.Retval)
	}

	return nil
}
