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

package vpp2001_324

import (
	"context"

	vpp_memif "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/memif"
	ifs "go.ligato.io/vpp-agent/v2/proto/ligato/vpp/interfaces"
)

func (h *InterfaceVppHandler) AddMemifInterface(ifName string, memIface *ifs.MemifLink, socketID uint32) (swIdx uint32, err error) {
	ctx := context.TODO()

	req := &vpp_memif.MemifCreate{
		ID:         memIface.Id,
		Mode:       memifMode(memIface.Mode),
		Secret:     memIface.Secret,
		SocketID:   socketID,
		BufferSize: uint16(memIface.BufferSize),
		RingSize:   memIface.RingSize,
		RxQueues:   uint8(memIface.RxQueues),
		TxQueues:   uint8(memIface.TxQueues),
	}
	if memIface.Master {
		req.Role = 0
	} else {
		req.Role = 1
	}
	// TODO: temporary fix, waiting for https://gerrit.fd.io/r/#/c/7266/
	if req.RxQueues == 0 {
		req.RxQueues = 1
	}
	if req.TxQueues == 0 {
		req.TxQueues = 1
	}

	reply, err := h.memif.MemifCreate(ctx, req)
	if err != nil {
		return 0, err
	}

	swIdx = uint32(reply.SwIfIndex)
	return swIdx, h.SetInterfaceTag(ifName, swIdx)
}

func (h *InterfaceVppHandler) DeleteMemifInterface(ifName string, idx uint32) error {
	ctx := context.TODO()

	req := &vpp_memif.MemifDelete{
		SwIfIndex: vpp_memif.InterfaceIndex(idx),
	}
	if _, err := h.memif.MemifDelete(ctx, req); err != nil {
		return err
	}

	return h.RemoveInterfaceTag(ifName, idx)
}

func (h *InterfaceVppHandler) RegisterMemifSocketFilename(filename string, id uint32) error {
	ctx := context.TODO()

	req := &vpp_memif.MemifSocketFilenameAddDel{
		SocketFilename: filename,
		SocketID:       id,
		IsAdd:          true, // sockets can be added only
	}
	if _, err := h.memif.MemifSocketFilenameAddDel(ctx, req); err != nil {
		return err
	}
	return nil
}

func memifMode(mode ifs.MemifLink_MemifMode) vpp_memif.MemifMode {
	switch mode {
	case ifs.MemifLink_IP:
		return vpp_memif.MEMIF_MODE_API_IP
	case ifs.MemifLink_PUNT_INJECT:
		return vpp_memif.MEMIF_MODE_API_PUNT_INJECT
	default:
		return vpp_memif.MEMIF_MODE_API_ETHERNET
	}
}
