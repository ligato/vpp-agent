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

	l2ba "github.com/ligato/vpp-agent/plugins/vpp/binapi/l2"
)


// AddL2FIB creates L2 FIB table entry.
func (h *FIBVppHandler) AddL2FIB(mac string, bdID uint32, ifaceIdx uint32, bvi bool, static bool) error {
	return h.l2fibAddDel(mac, bdID, ifaceIdx, bvi, static, true)
}

// DeleteL2FIB removes existing L2 FIB table entry.
func (h *FIBVppHandler) DeleteL2FIB(mac string, bdID uint32, ifaceIdx uint32) error {
	return h.l2fibAddDel(mac, bdID, ifaceIdx, false, false, false)
}

func (h *FIBVppHandler) l2fibAddDel(macstr string, bdID, ifaceIdx uint32, bvi, static, isAdd bool) (err error) {
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
		BdID:      bdID,
		SwIfIndex: ifaceIdx,
		BviMac:    boolToUint(bvi),
		StaticMac: boolToUint(static),
	}
	reply := &l2ba.L2fibAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	} else if reply.Retval != 0 {
		return fmt.Errorf("%s returned: %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}
