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
	"net"
	"time"

	"github.com/ligato/cn-infra/logging"
	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/l2"
)

// FibLogicalReq groups multiple fields so that all of them do
// not enumerate in one function call (request, reply/callback).
type FibLogicalReq struct {
	IsAdd    bool
	MAC      string
	BDIdx    uint32
	SwIfIdx  uint32
	BVI      bool
	Static   bool
	callback func(error)
}

func (handler *fibVppHandler) Add(mac string, bdID uint32, ifIdx uint32, bvi bool, static bool, callback func(error)) error {
	handler.log.Debug("Adding L2 FIB table entry, mac: ", mac)

	handler.requestChan <- &FibLogicalReq{
		IsAdd:    true,
		MAC:      mac,
		BDIdx:    bdID,
		SwIfIdx:  ifIdx,
		BVI:      bvi,
		Static:   static,
		callback: callback,
	}
	return nil
}

func (handler *fibVppHandler) Delete(mac string, bdID uint32, ifIdx uint32, callback func(error)) error {
	handler.log.Debug("Removing L2 fib table entry, mac: ", mac)

	handler.requestChan <- &FibLogicalReq{
		IsAdd:    false,
		MAC:      mac,
		BDIdx:    bdID,
		SwIfIdx:  ifIdx,
		callback: callback,
	}
	return nil
}

func (handler *fibVppHandler) WatchFIBReplies() {
	for {
		select {
		case r := <-handler.requestChan:
			handler.log.Debug("VPP L2FIB request: ", r)
			err := handler.l2fibAddDel(r.MAC, r.BDIdx, r.SwIfIdx, r.BVI, r.Static, r.IsAdd)
			if err != nil {
				handler.log.WithFields(logging.Fields{"mac": r.MAC, "bdIdx": r.BDIdx}).
					Error("Static fib entry add/delete failed:", err)
			} else {
				handler.log.WithFields(logging.Fields{"mac": r.MAC, "bdIdx": r.BDIdx}).
					Debug("Static fib entry added/deleted.")
			}
			r.callback(err)
		}
	}
}

func (handler *fibVppHandler) l2fibAddDel(macstr string, bdIdx, swIfIdx uint32, bvi, static, isAdd bool) (err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(l2ba.L2fibAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	var mac []byte
	if macstr != "" {
		mac, err = net.ParseMAC(macstr)
		if err != nil {
			return err
		}
	}

	req := &l2ba.L2fibAddDel{
		IsAdd:     boolToUint(isAdd),
		Mac:       mac,
		BdID:      bdIdx,
		SwIfIndex: swIfIdx,
		BviMac:    boolToUint(bvi),
		StaticMac: boolToUint(static),
	}

	reply := &l2ba.L2fibAddDelReply{}
	if err := handler.asyncCallsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}
