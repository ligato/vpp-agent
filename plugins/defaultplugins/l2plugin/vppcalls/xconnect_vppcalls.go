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
func VppSetL2XConnect(rxIfaceIdx uint32, txIfaceIdx uint32, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	log.Debug("Setting up L2 xConnect pair for ", txIfaceIdx, rxIfaceIdx)

	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.SwInterfaceSetL2Xconnect{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &l2ba.SwInterfaceSetL2Xconnect{
		Enable:      1,
		TxSwIfIndex: txIfaceIdx,
		RxSwIfIndex: rxIfaceIdx,
	}

	reply := &l2ba.SwInterfaceSetL2XconnectReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	log.WithFields(logging.Fields{"RxIface": rxIfaceIdx, "TxIface": txIfaceIdx}).Debug("L2xConnect created")
	return nil
}

// VppUnsetL2XConnect removes xConnect between two interfaces.
func VppUnsetL2XConnect(rxIfaceIdx uint32, txIfaceIdx uint32, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	log.Debug("Setting up L2 xConnect pair for ", txIfaceIdx, rxIfaceIdx)

	defer func(t time.Time) {
		stopwatch.TimeLog(l2ba.SwInterfaceSetL2Xconnect{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &l2ba.SwInterfaceSetL2Xconnect{
		Enable:      0,
		TxSwIfIndex: txIfaceIdx,
		RxSwIfIndex: rxIfaceIdx,
	}

	reply := &l2ba.SwInterfaceSetL2XconnectReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	log.WithFields(logging.Fields{"RxIface": rxIfaceIdx, "TxIface": txIfaceIdx}).Debug("L2xConnect removed")
	return nil
}
