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

	"errors"

	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/tap"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/tapv2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
)

// AddTapInterface calls TapConnect bin API.
func AddTapInterface(tapIf *interfaces.Interfaces_Interface_Tap, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) (swIndex uint32, err error) {
	// TapConnect/TapCreateV2 time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	if tapIf == nil || tapIf.HostIfName == "" {
		return 0, errors.New("host interface name was not provided for the TAP interface")
	}

	var (
		retval    int32
		swIfIndex uint32
	)

	if tapIf.Version == 2 {
		// Configure fast virtio-based TAP interface
		req := &tapv2.TapCreateV2{}
		req.ID = ^uint32(0)
		req.HostIfName = []byte(tapIf.HostIfName)
		req.HostIfNameSet = 1
		req.UseRandomMac = 1
		if tapIf.Namespace != "" {
			req.HostNamespace = []byte(tapIf.Namespace)
			req.HostNamespaceSet = 1
		}
		req.RxRingSz = uint16(tapIf.RxRingSize)
		req.TxRingSz = uint16(tapIf.TxRingSize)

		reply := &tapv2.TapCreateV2Reply{}
		err = vppChan.SendRequest(req).ReceiveReply(reply)
		retval = reply.Retval
		swIfIndex = reply.SwIfIndex
	} else {
		// Configure the original TAP interface
		req := &tap.TapConnect{}
		req.TapName = []byte(tapIf.HostIfName)
		req.UseRandomMac = 1

		reply := &tap.TapConnectReply{}
		err = vppChan.SendRequest(req).ReceiveReply(reply)
		retval = reply.Retval
		swIfIndex = reply.SwIfIndex
	}

	if err != nil {
		return 0, err
	}
	if 0 != retval {
		return 0, fmt.Errorf("add tap interface returned %d", retval)
	}

	return swIfIndex, nil
}

// DeleteTapInterface calls TapDelete bin API.
func DeleteTapInterface(idx uint32, version uint32, vppChan *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// TapDelete time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	var (
		err    error
		retval int32
	)

	if version == 2 {
		req := &tapv2.TapDeleteV2{}
		req.SwIfIndex = idx

		reply := &tapv2.TapDeleteV2Reply{}
		err = vppChan.SendRequest(req).ReceiveReply(reply)
		retval = reply.Retval
	} else {
		req := &tap.TapDelete{}
		req.SwIfIndex = idx

		reply := &tap.TapDeleteReply{}
		err = vppChan.SendRequest(req).ReceiveReply(reply)
		retval = reply.Retval
	}

	if err != nil {
		return err
	}

	if 0 != retval {
		return fmt.Errorf("deleting of interface returned %d", retval)
	}

	return nil
}
