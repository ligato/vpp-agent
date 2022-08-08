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
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	vpp2202 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202"
	vpp_fp "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/flowprobe"
	vpp_ipfix "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ipfix_export"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_ipfix.AllMessages()...)
	msgs = append(msgs, vpp_fp.AllMessages()...)

	vppcalls.AddIpfixHandlerVersion(vpp2202.Version, msgs, NewIpfixVppHandler)
}

// IpfixVppHandler is accessor for IPFIX-related vppcalls methods.
type IpfixVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// NewIpfixVppHandler creates new instance of IPFIX vppcalls handler.
func NewIpfixVppHandler(callsChan govppapi.Channel,
	ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger,
) vppcalls.IpfixVppAPI {
	return &IpfixVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}
