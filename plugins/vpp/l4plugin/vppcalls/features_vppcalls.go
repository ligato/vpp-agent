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

	"github.com/ligato/vpp-binapi/binapi/session"
)

// EnableL4Features enables L4 features.
func (h *L4VppHandler) EnableL4Features() error {
	req := &session.SessionEnableDisable{
		IsEnable: 1,
	}
	reply := &session.SessionEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	} else if reply.Retval != 0 {
		return fmt.Errorf("%s returned %v", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// DisableL4Features disables L4 features.
func (h *L4VppHandler) DisableL4Features() error {
	req := &session.SessionEnableDisable{
		IsEnable: 0,
	}
	reply := &session.SessionEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	} else if reply.Retval != 0 {
		return fmt.Errorf("%s returned %v", reply.GetMessageName(), reply.Retval)
	}

	return nil
}
