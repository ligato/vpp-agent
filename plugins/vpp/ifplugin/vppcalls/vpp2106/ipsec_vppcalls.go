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
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/tunnel_types"

	vpp_ipsec "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ipsec"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

// AddIPSecTunnelInterface adds a new IPSec tunnel interface.
func (h *InterfaceVppHandler) AddIPSecTunnelInterface(ctx context.Context, ifName string, ipSecLink *ifs.IPSecLink) (uint32, error) {
	reply, err := h.ipsec.IpsecItfCreate(ctx, &vpp_ipsec.IpsecItfCreate{
		Itf: vpp_ipsec.IpsecItf{
			Mode: tunnel_types.TunnelMode(ipSecLink.TunnelMode),
		},
	})
	if err != nil {
		return 0, err
	}

	return uint32(reply.SwIfIndex), nil
}

// DeleteIPSecTunnelInterface removes existing IPSec tunnel interface.
func (h *InterfaceVppHandler) DeleteIPSecTunnelInterface(ctx context.Context, ifName string, idx uint32, ipSecLink *ifs.IPSecLink) error {
	_, err := h.ipsec.IpsecItfDelete(ctx, &vpp_ipsec.IpsecItfDelete{
		SwIfIndex: interface_types.InterfaceIndex(idx),
	})
	if err != nil {
		return err
	}

	return nil
}
