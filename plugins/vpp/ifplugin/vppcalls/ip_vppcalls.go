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

	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
)

const (
	addInterfaceIP uint8 = 1
	delInterfaceIP uint8 = 0
)

func (handler *ifVppHandler) addDelInterfaceIP(ifIdx uint32, addr *net.IPNet, isAdd uint8) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(interfaces.SwInterfaceAddDelAddress{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &interfaces.SwInterfaceAddDelAddress{
		SwIfIndex: ifIdx,
		IsAdd:     isAdd,
	}

	prefix, _ := addr.Mask.Size()
	req.AddressLength = byte(prefix)

	isIPv6, err := addrs.IsIPv6(addr.IP.String())
	if err != nil {
		return err
	}
	if isIPv6 {
		req.Address = []byte(addr.IP.To16())
		req.IsIpv6 = 1
	} else {
		req.Address = []byte(addr.IP.To4())
		req.IsIpv6 = 0
	}

	reply := &interfaces.SwInterfaceAddDelAddressReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

func (handler *ifVppHandler) AddInterfaceIP(ifIdx uint32, addr *net.IPNet) error {
	return handler.addDelInterfaceIP(ifIdx, addr, addInterfaceIP)
}

func (handler *ifVppHandler) DelInterfaceIP(ifIdx uint32, addr *net.IPNet) error {
	return handler.addDelInterfaceIP(ifIdx, addr, delInterfaceIP)
}

const (
	setUnnumberedIP   uint8 = 1
	unsetUnnumberedIP uint8 = 0
)

func (handler *ifVppHandler) setUnsetUnnumberedIP(uIfIdx uint32, ifIdxWithIP uint32, isAdd uint8) error {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(interfaces.SwInterfaceSetUnnumbered{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Prepare the message.
	req := &interfaces.SwInterfaceSetUnnumbered{
		SwIfIndex:           ifIdxWithIP,
		UnnumberedSwIfIndex: uIfIdx,
		IsAdd:               isAdd,
	}

	reply := &interfaces.SwInterfaceSetUnnumberedReply{}
	if err := handler.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

func (handler *ifVppHandler) SetUnnumberedIP(uIfIdx uint32, ifIdxWithIP uint32) error {
	return handler.setUnsetUnnumberedIP(uIfIdx, ifIdxWithIP, setUnnumberedIP)
}

func (handler *ifVppHandler) UnsetUnnumberedIP(uIfIdx uint32) error {
	return handler.setUnsetUnnumberedIP(uIfIdx, 0, unsetUnnumberedIP)
}
