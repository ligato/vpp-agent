package vpp1904

import (
	"fmt"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1904/span"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

// SetSpan enables or disables SPAN on interface
func (h *InterfaceVppHandler) setSpan(ifIdxFrom, ifIdxTo uint32, state, isL2 uint8) error {
	req := &span.SwInterfaceSpanEnableDisable{
		SwIfIndexFrom: ifIdxFrom,
		SwIfIndexTo:   ifIdxTo,
		State:         state,
		IsL2:          isL2,
	}
	reply := &span.SwInterfaceSpanEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// AddSpan enables SPAN on interface
func (h *InterfaceVppHandler) AddSpan(ifIdxFrom, ifIdxTo uint32, state, isL2 uint8) error {
	return h.setSpan(ifIdxFrom, ifIdxTo, state, isL2)
}

// DelSpan disables SPAN on interface
func (h *InterfaceVppHandler) DelSpan(ifIdxFrom, ifIdxTo uint32, isL2 uint8) error {
	return h.setSpan(ifIdxFrom, ifIdxTo, 0, isL2)
}

// DumpSpan dumps SPAN table
func (h *InterfaceVppHandler) DumpSpan() ([]*vppcalls.InterfaceSpanDetails, error) {
	var spans []*vppcalls.InterfaceSpanDetails

	reqCtx := h.callsChannel.SendMultiRequest(&span.SwInterfaceSpanDump{})
	for {
		spanDetails := &span.SwInterfaceSpanDetails{}
		stop, err := reqCtx.ReceiveReply(spanDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to dump span: %v", err)
		}

		span := &vppcalls.InterfaceSpanDetails{
			SwIfIndexFrom: spanDetails.SwIfIndexFrom,
			SwIfIndexTo:   spanDetails.SwIfIndexTo,
			State:         spanDetails.State,
			IsL2:          spanDetails.IsL2,
		}
		spans = append(spans, span)
	}

	return spans, nil
}
