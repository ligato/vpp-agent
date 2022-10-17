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

package vpp2202

import (
	"context"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/interface_types"
	vpp_rdma "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/rdma"
)

// AddRdmaInterface adds new interface with RDMA driver.
func (h *InterfaceVppHandler) AddRdmaInterface(ctx context.Context, ifName string, rdmaLink *interfaces.RDMALink) (swIdx uint32, err error) {
	if h.rdma == nil {
		return 0, errors.WithMessage(vpp.ErrPluginDisabled, "rdma")
	}

	req := &vpp_rdma.RdmaCreateV3{
		HostIf:  rdmaLink.GetHostIfName(),
		Name:    ifName,
		RxqNum:  uint16(rdmaLink.GetRxqNum()),
		RxqSize: uint16(rdmaLink.GetRxqSize()),
		TxqSize: uint16(rdmaLink.GetTxqSize()),
		Mode:    rdmaMode(rdmaLink.GetMode()),
	}

	reply := &vpp_rdma.RdmaCreateV3Reply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, err
	}
	swIdx = uint32(reply.SwIfIndex)

	return swIdx, h.SetInterfaceTag(ifName, swIdx)
}

// DeleteRdmaInterface removes interface with RDMA driver.
func (h *InterfaceVppHandler) DeleteRdmaInterface(ctx context.Context, ifName string, ifIdx uint32) error {
	if h.rdma == nil {
		return errors.WithMessage(vpp.ErrPluginDisabled, "rdma")
	}

	req := &vpp_rdma.RdmaDelete{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
	}

	reply := &vpp_rdma.RdmaDeleteReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return h.RemoveInterfaceTag(ifName, ifIdx)
}

func rdmaMode(mode interfaces.RDMALink_Mode) vpp_rdma.RdmaMode {
	switch mode {
	case interfaces.RDMALink_DV:
		return vpp_rdma.RDMA_API_MODE_DV
	case interfaces.RDMALink_IBV:
		return vpp_rdma.RDMA_API_MODE_IBV
	default:
		return vpp_rdma.RDMA_API_MODE_AUTO
	}
}
