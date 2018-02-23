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
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/ipsec"
)

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
