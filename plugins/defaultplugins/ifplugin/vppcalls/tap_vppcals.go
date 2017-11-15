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

	"errors"

	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/tap"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
)

// AddTapInterface calls TapConnect bin API.
func AddTapInterface(tapIf *interfaces.Interfaces_Interface_Tap, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (swIndex uint32, err error) {
	// TapConnect time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if tapIf == nil || tapIf.HostIfName == "" {
		return 0, errors.New("host interface name was not provided for the TAP interface")
	}

	// Prepare the message.
	req := &tap.TapConnect{}
	req.TapName = []byte(tapIf.HostIfName)

	req.UseRandomMac = 1

	reply := &tap.TapConnectReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		return 0, err
	}

	if 0 != reply.Retval {
		return 0, fmt.Errorf("add tap interface returned %d", reply.Retval)
	}

	return reply.SwIfIndex, nil
}

// DeleteTapInterface calls TapDelete bin API.
func DeleteTapInterface(idx uint32, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// TapDelete time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message.
	req := &tap.TapDelete{}
	req.SwIfIndex = idx

	reply := &tap.TapDeleteReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("deleting of interface returned %d", reply.Retval)
	}

	return nil
}
