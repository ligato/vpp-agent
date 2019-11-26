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
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	vpe_vppcalls "go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls"
	vpe_vpp2001_379 "go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls/vpp2001"
	vpp_sr "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001/sr"
	vpp_vpe "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001/vpe"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/srplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_sr.AllMessages()...)
	msgs = append(msgs, vpp_vpe.AllMessages()...) // using also vpe -> need to have correct vpp version also for vpe

	vppcalls.Versions["vpp2001"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New: func(ch govppapi.Channel, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger) vppcalls.SRv6VppAPI {
			return NewSRv6VppHandler(ch, ifIndexes, log)
		},
	}
}

// SRv6VppHandler is accessor for SRv6-related vppcalls methods
type SRv6VppHandler struct {
	vpe_vppcalls.VppHandlerAPI

	log          logging.Logger
	ifIndexes    ifaceidx.IfaceMetadataIndex
	callsChannel govppapi.Channel
}

// NewSRv6VppHandler creates new instance of SRv6 vppcalls handler
func NewSRv6VppHandler(vppChan govppapi.Channel, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger) *SRv6VppHandler {
	return &SRv6VppHandler{
		callsChannel:  vppChan,
		ifIndexes:     ifIndexes,
		log:           log,
		VppHandlerAPI: vpe_vpp2001_379.NewVpeHandler(vppChan),
	}
}
