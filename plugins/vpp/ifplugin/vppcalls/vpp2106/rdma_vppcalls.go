//  Copyright (c) 2021 Cisco and/or its affiliates.
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

package vpp2106

import (
	"context"

	"github.com/pkg/errors"
	"go.fd.io/govpp/api"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/rdma"
)

// AddRdmaInterface adds new interface with RDMA driver.
func (h *InterfaceVppHandler) AddRdmaInterface(ctx context.Context, ifName string, rdmaLink *interfaces.RDMALink) (swIdx uint32, err error) {
	if h.rdma == nil {
		return 0, errors.WithMessage(vpp.ErrPluginDisabled, "rdma")
	}

	req := &rdma.RdmaCreate{
		HostIf:  rdmaLink.GetHostIfName(),
		Name:    ifName,
		RxqNum:  uint16(rdmaLink.GetRxqNum()),
		RxqSize: uint16(rdmaLink.GetRxqSize()),
		TxqSize: uint16(rdmaLink.GetTxqSize()),
		Mode:    rdmaMode(rdmaLink.GetMode()),
	}

	reply, err := h.rdma.RdmaCreate(ctx, req)
	if err != nil {
		return 0, err
	} else if err = api.RetvalToVPPApiError(reply.Retval); err != nil {
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

	req := &rdma.RdmaDelete{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
	}
	if reply, err := h.rdma.RdmaDelete(ctx, req); err != nil {
		return err
	} else if err = api.RetvalToVPPApiError(reply.Retval); err != nil {
		return err
	}

	return h.RemoveInterfaceTag(ifName, ifIdx)
}

func rdmaMode(mode interfaces.RDMALink_Mode) rdma.RdmaMode {
	switch mode {
	case interfaces.RDMALink_DV:
		return rdma.RDMA_API_MODE_DV
	case interfaces.RDMALink_IBV:
		return rdma.RDMA_API_MODE_IBV
	default:
		return rdma.RDMA_API_MODE_AUTO
	}
}
