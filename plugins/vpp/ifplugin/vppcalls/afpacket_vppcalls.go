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

	"net"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/af_packet"
	intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
)

func (handler *ifVppHandler) AddAfPacketInterface(ifName string, hwAddr string, afPacketIntf *intf.Interfaces_Interface_Afpacket) (swIndex uint32, err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(af_packet.AfPacketCreate{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &af_packet.AfPacketCreate{
		HostIfName: []byte(afPacketIntf.HostIfName),
	}

	if hwAddr == "" {
		req.UseRandomHwAddr = 1
	} else {
		mac, err := net.ParseMAC(hwAddr)

		if err != nil {
			return 0, err
		}

		req.HwAddr = mac
	}

	reply := &af_packet.AfPacketCreateReply{}
	if err = handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	if reply.Retval != 0 {
		return 0, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return reply.SwIfIndex, handler.SetInterfaceTag(ifName, reply.SwIfIndex)
}

func (handler *ifVppHandler) DeleteAfPacketInterface(ifName string, idx uint32, afPacketIntf *intf.Interfaces_Interface_Afpacket) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(af_packet.AfPacketDelete{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &af_packet.AfPacketDelete{
		HostIfName: []byte(afPacketIntf.HostIfName),
	}

	reply := &af_packet.AfPacketDeleteReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return handler.RemoveInterfaceTag(ifName, idx)
}
