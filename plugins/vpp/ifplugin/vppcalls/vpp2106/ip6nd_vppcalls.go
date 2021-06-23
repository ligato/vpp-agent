//  Copyright (c) 2020 Cisco and/or its affiliates.
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
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/rd_cp"
)

func (h *InterfaceVppHandler) SetIP6ndAutoconfig(ctx context.Context, ifIdx uint32, enable, installDefaultRoutes bool) error {
	_, err := h.rpcRdCp.IP6NdAddressAutoconfig(ctx, &rd_cp.IP6NdAddressAutoconfig{
		SwIfIndex:            interface_types.InterfaceIndex(ifIdx),
		Enable:               enable,
		InstallDefaultRoutes: installDefaultRoutes,
	})
	if err != nil {
		return err
	}
	return nil
}
