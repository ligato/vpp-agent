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
	"fmt"

	"github.com/pkg/errors"

	vpp_vmxnet3 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/vmxnet3"
	ifs "go.ligato.io/vpp-agent/v2/proto/ligato/vpp-agent/vpp/interfaces"
)

func (h *InterfaceVppHandler) AddVmxNet3(ifName string, vmxNet3 *ifs.VmxNet3Link) (swIdx uint32, err error) {
	var pci uint32
	pci, err = derivePCI(ifName)
	if err != nil {
		return 0, err
	}

	req := &vpp_vmxnet3.Vmxnet3Create{
		PciAddr: pci,
	}
	// Optional arguments
	if vmxNet3 != nil {
		req.EnableElog = int32(boolToUint(vmxNet3.EnableElog))
		req.RxqSize = uint16(vmxNet3.RxqSize)
		req.TxqSize = uint16(vmxNet3.TxqSize)
	}

	reply := &vpp_vmxnet3.Vmxnet3CreateReply{}
	if err = h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return 0, errors.Errorf(err.Error())
	}

	return reply.SwIfIndex, h.SetInterfaceTag(ifName, reply.SwIfIndex)
}

func (h *InterfaceVppHandler) DeleteVmxNet3(ifName string, ifIdx uint32) error {
	req := &vpp_vmxnet3.Vmxnet3Delete{
		SwIfIndex: ifIdx,
	}
	reply := &vpp_vmxnet3.Vmxnet3DeleteReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return errors.Errorf(err.Error())
	}

	return h.RemoveInterfaceTag(ifName, ifIdx)
}

func derivePCI(ifName string) (uint32, error) {
	var function, slot, bus, domain, pci uint32

	numLen, err := fmt.Sscanf(ifName, "vmxnet3-%x/%x/%x/%x", &domain, &bus, &slot, &function)
	if err != nil {
		err = errors.Errorf("cannot parse PCI address from the vmxnet3 interface name %s: %v", ifName, err)
		return 0, err
	}
	if numLen != 4 {
		err = errors.Errorf("cannot parse PCI address from the interface name %s: expected 4 address elements, received %d",
			ifName, numLen)
		return 0, err
	}

	pci |= function << 29
	pci |= slot << 24
	pci |= bus << 16
	pci |= domain

	return pci, nil
}
