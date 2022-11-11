//  Copyright (c) 2022 Cisco and/or its affiliates.
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

package vpp2210

import (
	"fmt"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/interface_types"
	vpp_span "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/span"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
)

// SetSpan enables or disables SPAN on interface
func (h *InterfaceVppHandler) setSpan(ifIdxFrom, ifIdxTo uint32, state uint8, isL2 bool) error {
	req := &vpp_span.SwInterfaceSpanEnableDisable{
		SwIfIndexFrom: interface_types.InterfaceIndex(ifIdxFrom),
		SwIfIndexTo:   interface_types.InterfaceIndex(ifIdxTo),
		State:         vpp_span.SpanState(state),
		IsL2:          isL2,
	}
	reply := &vpp_span.SwInterfaceSpanEnableDisableReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// AddSpan enables SPAN on interface
func (h *InterfaceVppHandler) AddSpan(ifIdxFrom, ifIdxTo uint32, direction uint8, isL2 bool) error {
	return h.setSpan(ifIdxFrom, ifIdxTo, direction, isL2)
}

// DelSpan disables SPAN on interface
func (h *InterfaceVppHandler) DelSpan(ifIdxFrom, ifIdxTo uint32, isL2 bool) error {
	return h.setSpan(ifIdxFrom, ifIdxTo, 0, isL2)
}

// DumpSpan dumps all SPAN table
func (h *InterfaceVppHandler) DumpSpan() ([]*vppcalls.InterfaceSpanDetails, error) {
	var spans []*vppcalls.InterfaceSpanDetails

	isL2Spans, err := h.dumpSpan(&vpp_span.SwInterfaceSpanDump{IsL2: true})
	if err != nil {
		return nil, err
	}
	spans = append(spans, isL2Spans...)

	isNotL2Spans, err := h.dumpSpan(&vpp_span.SwInterfaceSpanDump{IsL2: false})
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
			SwIfIndexFrom: uint32(spanDetails.SwIfIndexFrom),
			SwIfIndexTo:   uint32(spanDetails.SwIfIndexTo),
			Direction:     uint8(spanDetails.State),
			IsL2:          spanDetails.IsL2,
		}
		spans = append(spans, spanData)
	}
	return spans, nil
}
