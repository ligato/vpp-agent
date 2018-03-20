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
	"errors"
	"fmt"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/tap"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/tapv2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
)

// AddTapInterface calls TapConnect bin API.
func AddTapInterface(ifName string, tapIf *interfaces.Interfaces_Interface_Tap, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) (uint32, error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(tap.TapConnect{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	if tapIf == nil || tapIf.HostIfName == "" {
		return 0, errors.New("host interface name was not provided for the TAP interface")
	}

	var (
		err       error
		retval    int32
		swIfIndex uint32
		msgName   string
	)
	if tapIf.Version == 2 {
		// Configure fast virtio-based TAP interface
		req := &tapv2.TapCreateV2{
			ID:            ^uint32(0),
			HostIfName:    []byte(tapIf.HostIfName),
			HostIfNameSet: 1,
			UseRandomMac:  1,
			RxRingSz:      uint16(tapIf.RxRingSize),
			TxRingSz:      uint16(tapIf.TxRingSize),
		}
		if tapIf.Namespace != "" {
			req.HostNamespace = []byte(tapIf.Namespace)
			req.HostNamespaceSet = 1
		}

		reply := &tapv2.TapCreateV2Reply{}
		err = vppChan.SendRequest(req).ReceiveReply(reply)
		retval = reply.Retval
		swIfIndex = reply.SwIfIndex
		msgName = reply.GetMessageName()
	} else {
		// Configure the original TAP interface
		req := &tap.TapConnect{
			TapName:      []byte(tapIf.HostIfName),
			UseRandomMac: 1,
		}

		reply := &tap.TapConnectReply{}
		err = vppChan.SendRequest(req).ReceiveReply(reply)
		retval = reply.Retval
		swIfIndex = reply.SwIfIndex
		msgName = reply.GetMessageName()
	}
	if err != nil {
		return 0, err
	}
	if retval != 0 {
		return 0, fmt.Errorf("%s returned %d", msgName, retval)
	}

	return swIfIndex, SetInterfaceTag(ifName, swIfIndex, vppChan, stopwatch)
}

// DeleteTapInterface calls TapDelete bin API.
func DeleteTapInterface(ifName string, idx uint32, version uint32, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(tap.TapDelete{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	var (
		err     error
		retval  int32
		msgName string
	)
	if version == 2 {
		req := &tapv2.TapDeleteV2{
			SwIfIndex: idx,
		}

		reply := &tapv2.TapDeleteV2Reply{}
		err = vppChan.SendRequest(req).ReceiveReply(reply)
		retval = reply.Retval
		msgName = reply.GetMessageName()
	} else {
		req := &tap.TapDelete{
			SwIfIndex: idx,
		}

		reply := &tap.TapDeleteReply{}
		err = vppChan.SendRequest(req).ReceiveReply(reply)
		retval = reply.Retval
		msgName = reply.GetMessageName()
	}
	if err != nil {
		return err
	}
	if retval != 0 {
		return fmt.Errorf("%s returned %d", msgName, retval)
	}

	return RemoveInterfaceTag(ifName, idx, vppChan, stopwatch)
}
