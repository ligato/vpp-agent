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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/ipsec"
)

func spdAddDel(spdID uint32, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ipsec.IpsecSpdAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ipsec.IpsecSpdAddDel{
		IsAdd: boolToUint(isAdd),
		SpdID: spdID,
	}

	reply := &ipsec.IpsecSpdAddDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

func spdAddDelEntry(spdID uint32, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ipsec.IpsecSpdAddDelEntry{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ipsec.IpsecSpdAddDelEntry{
		IsAdd: boolToUint(isAdd),
		SpdID: spdID,
	}

	reply := &ipsec.IpsecSpdAddDelEntryReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

func interfaceAddDelSpd(spdID uint32, swIfIdx uint32, isAdd bool, vppChan *govppapi.Channel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(ipsec.IpsecInterfaceAddDelSpd{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &ipsec.IpsecInterfaceAddDelSpd{
		IsAdd:     boolToUint(isAdd),
		SwIfIndex: swIfIdx,
		SpdID:     spdID,
	}

	reply := &ipsec.IpsecInterfaceAddDelSpdReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// CheckMsgCompatibilityForIPSec verifies compatibility of used binary API calls
func CheckMsgCompatibilityForIPSec(vppChan *govppapi.Channel) error {
	msgs := []govppapi.Message{
		&ipsec.IpsecSpdAddDel{},
		&ipsec.IpsecSpdAddDelReply{},
		&ipsec.IpsecInterfaceAddDelSpd{},
		&ipsec.IpsecInterfaceAddDelSpdReply{},
		&ipsec.IpsecSpdAddDelEntry{},
		&ipsec.IpsecSpdAddDelEntryReply{},
		&ipsec.IpsecSadAddDelEntry{},
		&ipsec.IpsecSadAddDelEntryReply{},
		&ipsec.IpsecSpdDump{},
		&ipsec.IpsecSpdDetails{},
		&ipsec.IpsecTunnelIfAddDel{},
		&ipsec.IpsecTunnelIfAddDelReply{},
		&ipsec.IpsecSaDump{},
		&ipsec.IpsecSaDetails{},
		&ipsec.IpsecTunnelIfSetKey{},
		&ipsec.IpsecTunnelIfSetKeyReply{},
		&ipsec.IpsecTunnelIfSetSa{},
		&ipsec.IpsecTunnelIfSetSaReply{},
	}
	return vppChan.CheckMessageCompatibility(msgs...)
}

func boolToUint(value bool) uint8 {
	if value {
		return 1
	}
	return 0
}
