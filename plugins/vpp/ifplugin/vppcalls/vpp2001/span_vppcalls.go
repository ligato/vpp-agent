package vpp2001

import (
	"fmt"

	vpp_span "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/span"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
)

// SetSpan enables or disables SPAN on interface
func (h *InterfaceVppHandler) setSpan(ifIdxFrom, ifIdxTo uint32, state, isL2 uint8) error {
	req := &vpp_span.SwInterfaceSpanEnableDisable{
		SwIfIndexFrom: ifIdxFrom,
		SwIfIndexTo:   ifIdxTo,
		State:         state,
		IsL2:          isL2,
	}
	reply := &vpp_span.SwInterfaceSpanEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// AddSpan enables SPAN on interface
func (h *InterfaceVppHandler) AddSpan(ifIdxFrom, ifIdxTo uint32, direction, isL2 uint8) error {
	return h.setSpan(ifIdxFrom, ifIdxTo, direction, isL2)
}

// DelSpan disables SPAN on interface
func (h *InterfaceVppHandler) DelSpan(ifIdxFrom, ifIdxTo uint32, isL2 uint8) error {
	return h.setSpan(ifIdxFrom, ifIdxTo, 0, isL2)
}

// DumpSpan dumps all SPAN table
func (h *InterfaceVppHandler) DumpSpan() ([]*vppcalls.InterfaceSpanDetails, error) {
	var spans []*vppcalls.InterfaceSpanDetails

	isL2Spans, err := h.dumpSpan(&vpp_span.SwInterfaceSpanDump{IsL2: 1})
	if err != nil {
		return nil, err
	}
	spans = append(spans, isL2Spans...)

	isNotL2Spans, err := h.dumpSpan(&vpp_span.SwInterfaceSpanDump{IsL2: 0})
	if err != nil {
		return nil, err
	}
	spans = append(spans, isNotL2Spans...)

	return spans, nil
}

// dumpIsL2Span returns only SPANs with or without L2 set
func (h *InterfaceVppHandler) dumpSpan(msg *vpp_span.SwInterfaceSpanDump) ([]*vppcalls.InterfaceSpanDetails, error) {
	var spans []*vppcalls.InterfaceSpanDetails

	reqCtx := h.callsChannel.SendMultiRequest(msg)
	for {
		spanDetails := &vpp_span.SwInterfaceSpanDetails{}
		stop, err := reqCtx.ReceiveReply(spanDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to dump span: %v", err)
		}

		spanData := &vppcalls.InterfaceSpanDetails{
			SwIfIndexFrom: spanDetails.SwIfIndexFrom,
			SwIfIndexTo:   spanDetails.SwIfIndexTo,
			Direction:     spanDetails.State,
			IsL2:          spanDetails.IsL2,
		}
		spans = append(spans, spanData)
	}
	return spans, nil
}
