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

package vpp1901

import (
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/ip"
	"bytes"
)

// DumpVrfTables dumps all configured VRF tables.
func (h *VrfTableHandler) DumpVrfTables() (tables []*l3.VrfTable, err error) {
	// dump IPv4 VRF tables
	v4Tables := make(map[uint32]*l3.VrfTable)
	reqCtx := h.callsChannel.SendMultiRequest(&ip.IPFibDump{})
	for {
		fibDetails := &ip.IPFibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		if _, dumped := v4Tables[fibDetails.TableID]; !dumped {
			v4Tables[fibDetails.TableID] = &l3.VrfTable{
				Id:       fibDetails.TableID,
				Protocol: l3.VrfTable_IPV4,
				Label:    bytesToString(fibDetails.TableName),
			}
		}
	}

	// dump IPv6 VRF tables
	v6Tables := make(map[uint32]*l3.VrfTable)
	reqCtx = h.callsChannel.SendMultiRequest(&ip.IP6FibDump{})
	for {
		fibDetails := &ip.IP6FibDetails{}
		stop, err := reqCtx.ReceiveReply(fibDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		if _, dumped := v6Tables[fibDetails.TableID]; !dumped {
			v6Tables[fibDetails.TableID] = &l3.VrfTable{
				Id:       fibDetails.TableID,
				Protocol: l3.VrfTable_IPV6,
				Label:    bytesToString(fibDetails.TableName),
			}
		}
	}

	for _, table := range v4Tables {
		tables = append(tables, table)
	}
	for _, table := range v6Tables {
		tables = append(tables, table)
	}
	return tables, nil
}

func bytesToString(b []byte) string {
	return string(bytes.SplitN(b, []byte{0x00}, 2)[0])
}
