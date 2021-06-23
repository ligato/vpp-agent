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
	"fmt"
	"net"

	vpp_ipfix "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ipfix_export"
	ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"
)

// DumpExporters returns configured IPFIX.
// Since it always only one IPFIX configuration, this method signature
// defined as it is to keep consistensy between different vppcalls packages.
//
// Caution: VPP 20.01 does not support IPv6 addresses for IPFIX configuration,
// but this may change in future versions. Be careful porting this method to
// another version of VPP.
//
func (h *IpfixVppHandler) DumpExporters() ([]*ipfix.IPFIX, error) {
	var ipfixes []*ipfix.IPFIX
	reqCtx := h.callsChannel.SendMultiRequest(&vpp_ipfix.IpfixExporterDump{})
	for {
		details := &vpp_ipfix.IpfixExporterDetails{}
		stop, err := reqCtx.ReceiveReply(details)
		if stop {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to dump IPFIX: %v", err)
		}

		collectorIPAddr := details.CollectorAddress.Un.GetIP4()
		sourceIPAddr := details.SrcAddress.Un.GetIP4()

		ipfixes = append(ipfixes,
			&ipfix.IPFIX{
				Collector: &ipfix.IPFIX_Collector{
					Address: net.IP(collectorIPAddr[:]).To4().String(),
					Port:    uint32(details.CollectorPort),
				},
				SourceAddress:    net.IP(sourceIPAddr[:]).To4().String(),
				VrfId:            details.VrfID,
				PathMtu:          details.PathMtu,
				TemplateInterval: details.TemplateInterval,
			},
		)

	}

	return ipfixes, nil
}
