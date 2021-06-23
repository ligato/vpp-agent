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

package vpp2106

import (
	"fmt"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	vpp_vmxnet3 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vmxnet3"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func (h *InterfaceVppHandler) AddVmxNet3(ifName string, vmxNet3 *ifs.VmxNet3Link) (swIdx uint32, err error) {
	if h.vmxnet3 == nil {
		return 0, errors.WithMessage(vpp.ErrPluginDisabled, "wmxnet")
	}

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

	return uint32(reply.SwIfIndex), h.SetInterfaceTag(ifName, uint32(reply.SwIfIndex))
}

func (h *InterfaceVppHandler) DeleteVmxNet3(ifName string, ifIdx uint32) error {
	if h.vmxnet3 == nil {
		return errors.WithMessage(vpp.ErrPluginDisabled, "wmxnet")
	}

	req := &vpp_vmxnet3.Vmxnet3Delete{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
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

// dumpVmxNet3Details dumps VmxNet3 interface details from VPP and fills them into the provided interface map.
func (h *InterfaceVppHandler) dumpVmxNet3Details(interfaces map[uint32]*vppcalls.InterfaceDetails) error {
	if h.vmxnet3 == nil {
		// no-op when disabled
		return nil
	}

	reqCtx := h.callsChannel.SendMultiRequest(&vpp_vmxnet3.Vmxnet3Dump{})
	for {
		vmxnet3Details := &vpp_vmxnet3.Vmxnet3Details{}
		stop, err := reqCtx.ReceiveReply(vmxnet3Details)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return fmt.Errorf("failed to dump VmxNet3 tunnel interface details: %v", err)
		}
		_, ifIdxExists := interfaces[uint32(vmxnet3Details.SwIfIndex)]
		if !ifIdxExists {
			continue
		}
		interfaces[uint32(vmxnet3Details.SwIfIndex)].Interface.Link = &ifs.Interface_VmxNet3{
			VmxNet3: &ifs.VmxNet3Link{
				RxqSize: uint32(vmxnet3Details.RxCount),
				TxqSize: uint32(vmxnet3Details.TxCount),
			},
		}
		interfaces[uint32(vmxnet3Details.SwIfIndex)].Interface.Type = ifs.Interface_VMXNET3_INTERFACE
		interfaces[uint32(vmxnet3Details.SwIfIndex)].Meta.Pci = vmxnet3Details.PciAddr
	}
	return nil
}
