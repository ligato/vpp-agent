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

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/memif"
	intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
)

func (handler *ifVppHandler) AddMemifInterface(ifName string, memIface *intf.Interfaces_Interface_Memif, socketID uint32) (swIdx uint32, err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(memif.MemifCreate{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &memif.MemifCreate{
		ID:         memIface.Id,
		Mode:       uint8(memIface.Mode),
		Secret:     []byte(memIface.Secret),
		SocketID:   socketID,
		BufferSize: uint16(memIface.BufferSize),
		RingSize:   memIface.RingSize,
		RxQueues:   uint8(memIface.RxQueues),
		TxQueues:   uint8(memIface.TxQueues),
	}
	if memIface.Master {
		req.Role = 0
	} else {
		req.Role = 1
	}
	// TODO: temporary fix, waiting for https://gerrit.fd.io/r/#/c/7266/
	if req.RxQueues == 0 {
		req.RxQueues = 1
	}
	if req.TxQueues == 0 {
		req.TxQueues = 1
	}

	reply := &memif.MemifCreateReply{}
	if err = handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	if reply.Retval != 0 {
		return 0, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return reply.SwIfIndex, handler.SetInterfaceTag(ifName, reply.SwIfIndex)
}

func (handler *ifVppHandler) DeleteMemifInterface(ifName string, idx uint32) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(memif.MemifDelete{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &memif.MemifDelete{
		SwIfIndex: idx,
	}

	reply := &memif.MemifDeleteReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return handler.RemoveInterfaceTag(ifName, idx)
}

func (handler *ifVppHandler) RegisterMemifSocketFilename(filename []byte, id uint32) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(memif.MemifSocketFilenameAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &memif.MemifSocketFilenameAddDel{
		SocketFilename: filename,
		SocketID:       id,
		IsAdd:          1, // sockets can be added only
	}

	reply := &memif.MemifSocketFilenameAddDelReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}
