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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/session"
)

// EnableL4Features sets L4 feature flag on VPP to true
func EnableL4Features(vppChan *govppapi.Channel) error {
	req := &session.SessionEnableDisable{
		IsEnable: 1,
	}

	reply := &session.SessionEnableDisableReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %v", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// DisableL4Features sets L4 feature flag on VPP to false
func DisableL4Features(vppChan *govppapi.Channel) error {
	req := &session.SessionEnableDisable{
		IsEnable: 0,
	}

	reply := &session.SessionEnableDisableReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %v", reply.GetMessageName(), reply.Retval)
	}

	return nil
}
