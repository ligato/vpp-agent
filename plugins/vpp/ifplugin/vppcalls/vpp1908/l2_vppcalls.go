//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp1908

import (
	"fmt"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/l2"
)

// TODO:  more suitable for the l2 plugin, but the tag-rewrite retrieve is a part of the vpp interface api

// SetInterfaceTagRewrite sets an interface tag rewrite
func (h *InterfaceVppHandler) SetVLanTagRewrite(ifIdx uint32, subIf *interfaces.SubInterface) error {
	req := &l2.L2InterfaceVlanTagRewrite{
		SwIfIndex: ifIdx,
		VtrOp:     getTagRewriteOption(subIf.TagRwOption),
		PushDot1q: uint32(boolToUint(subIf.PushDot1Q)),
		Tag1:      subIf.Tag1,
		Tag2:      subIf.Tag2,
	}
	reply := &l2.L2InterfaceVlanTagRewriteReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return fmt.Errorf("%s returned error: %v", reply.GetMessageName(), err)
	}

	return nil
}

func getTagRewriteOption(op interfaces.SubInterface_TagRewriteOptions) uint32 {
	switch op {
	case interfaces.SubInterface_PUSH1:
		return 1
	case interfaces.SubInterface_PUSH2:
		return 2
	case interfaces.SubInterface_POP1:
		return 3
	case interfaces.SubInterface_POP2:
		return 4
	case interfaces.SubInterface_TRANSLATE11:
		return 5
	case interfaces.SubInterface_TRANSLATE12:
		return 6
	case interfaces.SubInterface_TRANSLATE21:
		return 7
	case interfaces.SubInterface_TRANSLATE22:
		return 8
	default: // disabled
		return 0
	}
}
