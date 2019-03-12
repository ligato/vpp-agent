package vpp1901

import (
	"fmt"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/l2"
)

// TODO:  more suitable for the l2 plugin, but since the tag-rewrite retrieve is a part of the vpp interface api ...

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
