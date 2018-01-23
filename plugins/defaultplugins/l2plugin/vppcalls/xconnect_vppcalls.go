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

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
)

// VppSetL2XConnect creates xConnect between two existing interfaces.
func VppSetL2XConnect(receiveIfaceIndex uint32, transmitIfaceIndex uint32, log logging.Logger, vppChan VPPChannel, timeLog measure.StopWatchEntry) error {
	log.Debug("Setting up L2 xConnect pair for ", transmitIfaceIndex, receiveIfaceIndex)
	// SwInterfaceSetL2Xconnect time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &l2ba.SwInterfaceSetL2Xconnect{}
	req.TxSwIfIndex = transmitIfaceIndex
	req.RxSwIfIndex = receiveIfaceIndex
	req.Enable = 1

	reply := &l2ba.SwInterfaceSetL2XconnectReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("creating xConnect returned %d", reply.Retval)
	}

	log.WithFields(logging.Fields{"RxIface": receiveIfaceIndex, "TxIface": transmitIfaceIndex}).Debug("L2xConnect created.")
	return nil
}

// VppUnsetL2XConnect removes xConnect between two interfaces.
func VppUnsetL2XConnect(receiveIfaceIndex uint32, transmitIfaceIndex uint32, log logging.Logger, vppChan VPPChannel, timeLog measure.StopWatchEntry) error {
	log.Debug("Setting up L2 xConnect pair for ", transmitIfaceIndex, receiveIfaceIndex)
	// SwInterfaceSetL2Xconnect time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	req := &l2ba.SwInterfaceSetL2Xconnect{}
	req.RxSwIfIndex = receiveIfaceIndex
	req.TxSwIfIndex = transmitIfaceIndex
	req.Enable = 0

	reply := &l2ba.SwInterfaceSetL2XconnectReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("removing xConnect returned %d", reply.Retval)
	}

	log.WithFields(logging.Fields{"RxIface": receiveIfaceIndex, "TxIface": transmitIfaceIndex}).Debug("L2xConnect removed.")
	return nil
}
