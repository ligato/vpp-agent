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
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/memif"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"time"
)

// AddMemifInterface calls MemifCreate bin API
func AddMemifInterface(memIntf *intf.Interfaces_Interface_Memif, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (swIndex uint32, err error) {
	// MemifCreate time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// prepare the message
	req := &memif.MemifCreate{}

	req.ID = memIntf.Id
	if memIntf.Master {
		req.Role = 0
	} else {
		req.Role = 1
	}
	req.Mode = uint8(memIntf.Mode)
	req.Secret = []byte(memIntf.Secret)
	req.SocketFilename = []byte(memIntf.SocketFilename)
	req.BufferSize = uint16(memIntf.BufferSize)
	req.RingSize = memIntf.RingSize
	req.RxQueues = uint8(memIntf.RxQueues)
	req.TxQueues = uint8(memIntf.TxQueues)

	/* TODO: temporary fix, waiting for https://gerrit.fd.io/r/#/c/7266/ */
	if req.RxQueues == 0 {
		req.RxQueues = 1
	}
	if req.TxQueues == 0 {
		req.TxQueues = 1
	}

	reply := &memif.MemifCreateReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		return 0, err
	}

	if 0 != reply.Retval {
		return 0, fmt.Errorf("add memif interface returned %d", reply.Retval)
	}

	return reply.SwIfIndex, nil
}

// DeleteMemifInterface calls MemifDelete bin API
func DeleteMemifInterface(idx uint32, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// MemifDelete time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// prepare the message
	req := &memif.MemifDelete{}
	req.SwIfIndex = idx

	reply := &memif.MemifDeleteReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("deleting of interface returned %d", reply.Retval)
	}

	return nil

}
