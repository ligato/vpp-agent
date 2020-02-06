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

package vpp2001

import (
	"strings"

	vpp_ip "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

// DumpVrfTables dumps all configured VRF tables.
func (h *VrfTableHandler) DumpVrfTables() (tables []*l3.VrfTable, err error) {
	// dump IPv4 VRF tables
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_ip.IPTableDump{})
	for {
		fibDetails := &vpp_ip.IPTableDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		tables = append(tables, &l3.VrfTable{
			Id:       fibDetails.Table.TableID,
			Protocol: getTableProto(fibDetails.Table.IsIP6),
			Label:    strings.Trim(fibDetails.Table.Name, "\x00"),
		})
	}

	return tables, nil
}

func getTableProto(isIPv6 bool) l3.VrfTable_Protocol {
	if isIPv6 {
		return l3.VrfTable_IPV6
	}
	return l3.VrfTable_IPV4
}
