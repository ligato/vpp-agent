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
	"container/list"
	"fmt"
	"net"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/l2"
)

// FibLogicalReq groups multiple fields so that all of them do not enumerate in one function call (request, reply/callback).
type FibLogicalReq struct {
	MAC      string
	BDIdx    uint32
	SwIfIdx  uint32
	BVI      bool
	Static   bool
	Delete   bool
	callback func(error)
}

// NewL2FibVppCalls is a constructor.
func NewL2FibVppCalls(vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) *L2FibVppCalls {
	return &L2FibVppCalls{vppChan, stopwatch, list.New()}
}

// L2FibVppCalls aggregates vpp calls related to l2 fib.
type L2FibVppCalls struct {
	vppChan         *govppapi.Channel
	stopwatch       *measure.Stopwatch
	waitingForReply *list.List
}

// Add creates L2 FIB table entry.
func (fib *L2FibVppCalls) Add(mac string, bdID uint32, ifIdx uint32, bvi bool, static bool, callback func(error), log logging.Logger) error {
	log.Debug("Adding L2 FIB table entry, mac: ", mac)

	defer func(t time.Time) {
		fib.stopwatch.TimeLog(l2ba.L2fibAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	return fib.request(&FibLogicalReq{
		MAC:      mac,
		BDIdx:    bdID,
		SwIfIdx:  ifIdx,
		BVI:      bvi,
		Static:   static,
		Delete:   false,
		callback: callback,
	}, log)
}

// Delete removes existing L2 FIB table entry.
func (fib *L2FibVppCalls) Delete(mac string, bdID uint32, ifIdx uint32, callback func(error), log logging.Logger) error {
	log.Debug("Removing L2 fib table entry, mac: ", mac)

	defer func(t time.Time) {
		fib.stopwatch.TimeLog(l2ba.L2fibAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	return fib.request(&FibLogicalReq{
		MAC:      mac,
		BDIdx:    bdID,
		SwIfIdx:  ifIdx,
		Delete:   true,
		callback: callback,
	}, log)
}

func (fib *L2FibVppCalls) request(logicalReq *FibLogicalReq, log logging.Logger) (err error) {
	// Convert MAC address.
	var mac []byte
	if logicalReq.MAC != "" {
		mac, err = net.ParseMAC(logicalReq.MAC)
	}
	if err != nil {
		return err
	}

	req := &l2ba.L2fibAddDel{}
	req.Mac = mac
	req.BdID = logicalReq.BDIdx
	req.SwIfIndex = logicalReq.SwIfIdx
	req.BviMac = boolToUint(logicalReq.BVI)
	req.StaticMac = boolToUint(logicalReq.Static)
	if logicalReq.Delete {
		req.IsAdd = 0
	} else {
		req.IsAdd = 1
	}

	fib.waitingForReply.PushFront(logicalReq)
	fib.vppChan.ReqChan <- &govppapi.VppRequest{
		Message: req,
	}

	log.WithFields(logging.Fields{"Mac": req.Mac, "BD index": req.BdID}).Debug("Static fib entry added/deleted.")
	return nil
}

// WatchFIBReplies is meant to be used in go routine.
func (fib *L2FibVppCalls) WatchFIBReplies(log logging.Logger) {
	for {
		vppReply := <-fib.vppChan.ReplyChan
		log.Debug("VPP FIB Reply ", vppReply)

		if vppReply.LastReplyReceived {
			log.Debug("Ping received")
			//TODO check with Rasto
			// ERRO[0001] no reply received within the timeout period 1s
			// loc="vppcalls/dump_vppcalls.go(70)" tag=00000000 D
			continue
		}

		if fib.waitingForReply.Len() == 0 {
			log.WithField("MessageID", vppReply.MessageID). //TODO WithField("err", vppReply.Error).
									Error("Unexpected message ", vppReply)
			continue
		}

		logicalReq := fib.waitingForReply.Remove(fib.waitingForReply.Front()).(*FibLogicalReq)
		log.WithField("Mac", logicalReq.MAC).Debug("VPP FIB Reply ", vppReply)

		if vppReply.Error != nil {
			logicalReq.callback(vppReply.Error)
		} else {
			reply := &l2ba.L2fibAddDelReply{}
			err := fib.vppChan.MsgDecoder.DecodeMsg(vppReply.Data, reply)
			if err != nil || 0 != reply.Retval {
				err = fmt.Errorf("adding/del Static fib entry returned %d", reply.Retval)
				logicalReq.callback(err)
			} else {
				logicalReq.callback(nil)
			}
		}
	}
}
