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
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"go.fd.io/govpp/api"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/interface_types"
	vpp_memif "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/memif"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func (h *InterfaceVppHandler) AddMemifInterface(ctx context.Context, ifName string, memIface *ifs.MemifLink, socketID uint32) (swIdx uint32, err error) {
	if h.memif == nil {
		return 0, errors.WithMessage(vpp.ErrPluginDisabled, "memif")
	}

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
	} else if err = api.RetvalToVPPApiError(reply.Retval); err != nil {
		return 0, err
	}
	swIdx = uint32(reply.SwIfIndex)

	return swIdx, h.SetInterfaceTag(ifName, swIdx)
}

func (h *InterfaceVppHandler) DeleteMemifInterface(ctx context.Context, ifName string, idx uint32) error {
	if h.memif == nil {
		return errors.WithMessage(vpp.ErrPluginDisabled, "memif")
	}

	req := &vpp_memif.MemifDelete{
		SwIfIndex: interface_types.InterfaceIndex(idx),
	}
	if reply, err := h.memif.MemifDelete(ctx, req); err != nil {
		return err
	} else if err = api.RetvalToVPPApiError(reply.Retval); err != nil {
		return err
	}

	return h.RemoveInterfaceTag(ifName, idx)
}

func (h *InterfaceVppHandler) RegisterMemifSocketFilename(ctx context.Context, filename string, id uint32) error {
	if h.memif == nil {
		return errors.WithMessage(vpp.ErrPluginDisabled, "memif")
	}

	req := &vpp_memif.MemifSocketFilenameAddDel{
		SocketFilename: filename,
		SocketID:       id,
		IsAdd:          true, // sockets can be added only
	}
	if reply, err := h.memif.MemifSocketFilenameAddDel(ctx, req); err != nil {
		return err
	} else if err = api.RetvalToVPPApiError(reply.Retval); err != nil {
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

func (h *InterfaceVppHandler) DumpMemifSocketDetails(ctx context.Context) (map[string]uint32, error) {
	if h.memif == nil {
		return nil, errors.WithMessage(vpp.ErrPluginDisabled, "memif")
	}

	dump, err := h.memif.MemifSocketFilenameDump(ctx, &vpp_memif.MemifSocketFilenameDump{})
	if err != nil {
		return nil, err
	}
	memifSocketMap := make(map[string]uint32)
	for {
		socketDetails, err := dump.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		filename := strings.SplitN(socketDetails.SocketFilename, "\x00", 2)[0]
		memifSocketMap[filename] = socketDetails.SocketID
	}

	h.log.Debugf("Memif socket dump completed, found %d entries: %v", len(memifSocketMap), memifSocketMap)

	return memifSocketMap, nil
}

// dumpMemifDetails dumps memif interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpMemifDetails(ctx context.Context, interfaces map[uint32]*vppcalls.InterfaceDetails) error {
	if h.memif == nil {
		// no-op when disabled
		return nil
	}

	memifSocketMap, err := h.DumpMemifSocketDetails(ctx)
	if err != nil {
		return fmt.Errorf("dumping memif socket details failed: %v", err)
	}

	dump, err := h.memif.MemifDump(ctx, &vpp_memif.MemifDump{})
	if err != nil {
		return err
	}
	for {
		memifDetails, err := dump.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		_, ifIdxExists := interfaces[uint32(memifDetails.SwIfIndex)]
		if !ifIdxExists {
			continue
		}
		interfaces[uint32(memifDetails.SwIfIndex)].Interface.Link = &ifs.Interface_Memif{
			Memif: &ifs.MemifLink{
				Master: memifDetails.Role == 0,
				Mode:   memifModetoNB(memifDetails.Mode),
				Id:     memifDetails.ID,
				// Secret: // TODO: Secret - not available in the binary API
				SocketFilename: func(socketMap map[string]uint32) (filename string) {
					for filename, id := range socketMap {
						if memifDetails.SocketID == id {
							return filename
						}
					}
					// Socket for configured memif should exist
					h.log.Warnf("Socket ID not found for memif %v", memifDetails.SwIfIndex)
					return
				}(memifSocketMap),
				RingSize:   memifDetails.RingSize,
				BufferSize: uint32(memifDetails.BufferSize),
				// TODO: RxQueues, TxQueues - not available in the binary API
				// RxQueues:
				// TxQueues:
			},
		}
		interfaces[uint32(memifDetails.SwIfIndex)].Interface.Type = ifs.Interface_MEMIF
	}

	return nil
}
