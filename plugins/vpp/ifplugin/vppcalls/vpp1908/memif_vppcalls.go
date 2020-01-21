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
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	vpp_memif "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/memif"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func (h *InterfaceVppHandler) AddMemifInterface(ctx context.Context, ifName string, memIface *interfaces.MemifLink, socketID uint32) (swIdx uint32, err error) {
	if h.memif == nil {
		return 0, errors.WithMessage(vpp.ErrPluginDisabled, "memif")
	}

	req := &vpp_memif.MemifCreate{
		ID:         memIface.Id,
		Mode:       uint8(memIface.Mode),
		Secret:     []byte(memIface.Secret),
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

func (h *InterfaceVppHandler) DeleteMemifInterface(ctx context.Context, ifName string, idx uint32) error {
	if h.memif == nil {
		return errors.WithMessage(vpp.ErrPluginDisabled, "memif")
	}

	req := &vpp_memif.MemifDelete{
		SwIfIndex: idx,
	}
	if _, err := h.memif.MemifDelete(ctx, req); err != nil {
		return err
	}

	return h.RemoveInterfaceTag(ifName, idx)
}

func (h *InterfaceVppHandler) RegisterMemifSocketFilename(ctx context.Context, filename string, id uint32) error {
	if h.memif == nil {
		return errors.WithMessage(vpp.ErrPluginDisabled, "memif")
	}

	req := &vpp_memif.MemifSocketFilenameAddDel{
		SocketFilename: []byte(filename),
		SocketID:       id,
		IsAdd:          1, // sockets can be added only
	}
	if _, err := h.memif.MemifSocketFilenameAddDel(ctx, req); err != nil {
		return err
	}
	return nil
}

// DumpMemifSocketDetails implements interface handler.
func (h *InterfaceVppHandler) DumpMemifSocketDetails(ctx context.Context) (map[string]uint32, error) {
	if h.memif == nil {
		return nil, errors.WithMessage(vpp.ErrPluginDisabled, "memif")
	}

	dump, err := h.memif.DumpMemifSocketFilename(ctx, &vpp_memif.MemifSocketFilenameDump{})
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

		filename := string(bytes.SplitN(socketDetails.SocketFilename, []byte{0x00}, 2)[0])
		memifSocketMap[filename] = socketDetails.SocketID
	}

	h.log.Debugf("Memif socket dump completed, found %d entries: %v", len(memifSocketMap), memifSocketMap)

	return memifSocketMap, nil
}

// dumpMemifDetails dumps memif interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpMemifDetails(ctx context.Context, ifs map[uint32]*vppcalls.InterfaceDetails) error {
	if h.memif == nil {
		// no-op when disabled
		return nil
	}

	memifSocketMap, err := h.DumpMemifSocketDetails(ctx)
	if err != nil {
		return fmt.Errorf("dumping memif socket details failed: %v", err)
	}

	dump, err := h.memif.DumpMemif(ctx, &vpp_memif.MemifDump{})
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

		_, ifIdxExists := ifs[memifDetails.SwIfIndex]
		if !ifIdxExists {
			continue
		}
		ifs[memifDetails.SwIfIndex].Interface.Link = &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Master: memifDetails.Role == 0,
				Mode:   memifModetoNB(memifDetails.Mode),
				Id:     memifDetails.ID,
				//Secret: // TODO: Secret - not available in the binary API
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
				//RxQueues:
				//TxQueues:
			},
		}
		ifs[memifDetails.SwIfIndex].Interface.Type = interfaces.Interface_MEMIF
	}

	return nil
}
