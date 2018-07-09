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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
)

const (
	addInterfaceIP uint8 = 1
	delInterfaceIP uint8 = 0
)

func addDelInterfaceIP(ifIdx uint32, addr *net.IPNet, isAdd uint8, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(interfaces.SwInterfaceAddDelAddress{}).LogTimeEntry(time.Since(t))
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
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// AddInterfaceIP calls SwInterfaceAddDelAddress bin API with IsAdd=1.
func AddInterfaceIP(ifIdx uint32, addr *net.IPNet, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return addDelInterfaceIP(ifIdx, addr, addInterfaceIP, vppChan, stopwatch)
}

// DelInterfaceIP calls SwInterfaceAddDelAddress bin API with IsAdd=00.
func DelInterfaceIP(ifIdx uint32, addr *net.IPNet, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return addDelInterfaceIP(ifIdx, addr, delInterfaceIP, vppChan, stopwatch)
}

const (
	setUnnumberedIP   uint8 = 1
	unsetUnnumberedIP uint8 = 0
)

func setUnsetUnnumberedIP(uIfIdx uint32, ifIdxWithIP uint32, isAdd uint8, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(interfaces.SwInterfaceSetUnnumbered{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Prepare the message.
	req := &interfaces.SwInterfaceSetUnnumbered{
		SwIfIndex:           ifIdxWithIP,
		UnnumberedSwIfIndex: uIfIdx,
		IsAdd:               isAdd,
	}

	reply := &interfaces.SwInterfaceSetUnnumberedReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// SetUnnumberedIP sets interface as un-numbered, linking IP address of the another interface (ifIdxWithIP)
func SetUnnumberedIP(uIfIdx uint32, ifIdxWithIP uint32, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return setUnsetUnnumberedIP(uIfIdx, ifIdxWithIP, setUnnumberedIP, vppChan, stopwatch)
}

// UnsetUnnumberedIP unset provided interface as un-numbered. IP address of the linked interface is removed
func UnsetUnnumberedIP(uIfIdx uint32, vppChan govppapi.Channel, stopwatch *measure.Stopwatch) error {
	return setUnsetUnnumberedIP(uIfIdx, 0, unsetUnnumberedIP, vppChan, stopwatch)
}
